package depman

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"

	"github.com/devnadeemashraf/depman/internal/downloader"
	"github.com/devnadeemashraf/depman/internal/environment"
	"github.com/devnadeemashraf/depman/internal/logger"
)

// NewManager creates a new dependency manager with optional configuration
func NewManager(configPath string, opts ...Option) (*Manager, error) {
	// Load dependency configuration
	config, err := LoadDependencyConfig(configPath)
	if err != nil {
		return nil, err
	}

	// Create a new manager with defaults
	manager := &Manager{
		Config:     config,
		ConfigPath: configPath,
		Platform:   runtime.GOOS, // "windows", "linux", or "darwin"
		logger:     logger.Default(),
		envManager: environment.NewManager(),
	}

	// Apply any provided options
	for _, opt := range opts {
		opt(manager)
	}

	return manager, nil
}

// GetPlatformConfig returns platform-specific configuration for a dependency
func (m *Manager) GetPlatformConfig(dep *Dependency) (*PlatformConfig, error) {
	// Check if we have configuration for current platform
	platform, ok := dep.Platforms[m.Platform]
	if !ok {
		return nil, fmt.Errorf("no configuration available for platform: %s", m.Platform)
	}

	return &platform, nil
}

// CheckDependency verifies if a dependency is installed and if it needs updating
func (m *Manager) CheckDependency(dep *Dependency) (*DependencyStatus, error) {
	// Use the more thorough verification
	return m.VerifyDependency(dep)
}

// validateDependencies checks if all dependencies are properly defined
func (m *Manager) validateDependencies() []error {
	var errors []error

	// Check if there are any dependencies defined
	if len(m.Config.Dependencies) == 0 {
		errors = append(errors, fmt.Errorf("no dependencies defined in configuration"))
		return errors
	}

	// Validate each dependency
	for _, dep := range m.Config.Dependencies {
		// Check if platform-specific config exists
		if _, ok := dep.Platforms[m.Platform]; !ok {
			errors = append(errors, fmt.Errorf("dependency '%s' has no configuration for platform '%s'",
				dep.Name, m.Platform))
			continue
		}

		// Validate version information
		if dep.Version.Required == "" {
			errors = append(errors, fmt.Errorf("dependency '%s' has no required version", dep.Name))
		}

		// If constraint is provided, make sure it's valid
		if dep.Version.Constraint != "" {
			if _, err := semver.NewConstraint(dep.Version.Constraint); err != nil {
				errors = append(errors, fmt.Errorf("dependency '%s' has invalid version constraint '%s': %w",
					dep.Name, dep.Version.Constraint, err))
			}
		}
	}

	return errors
}

// installDependency handles the actual installation of a dependency

// installDependency handles the actual installation of a dependency
func (m *Manager) installDependency(dep *Dependency) error {
	// Get platform config
	platformConfig, err := m.GetPlatformConfig(dep)
	if err != nil {
		return err
	}

	// Create a temporary directory for downloads
	tempDir, err := os.MkdirTemp("", "depman-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up when done

	// Download dependency if URL is specified
	downloadPath := ""
	if platformConfig.Installer.URL != "" {
		m.logger.Infof("Downloading %s from %s", dep.Name, platformConfig.Installer.URL)

		// Set up download options
		opts := downloader.DownloadOptions{
			URL:          platformConfig.Installer.URL,
			DestDir:      tempDir,
			ShowProgress: true,
		}

		// Add checksum if provided
		if platformConfig.Installer.Checksum != "" {
			opts.Checksum = platformConfig.Installer.Checksum
		}

		// Download the file
		result, err := downloader.Download(opts)
		if err != nil {
			return fmt.Errorf("failed to download dependency: %w", err)
		}

		downloadPath = result.FilePath
		m.logger.Infof("Downloaded %s (%d bytes)", dep.Name, result.Size)
	}

	// Prepare install command with replacements
	installCmd := make([]string, len(platformConfig.Commands.Install))
	for i, arg := range platformConfig.Commands.Install {
		// Replace placeholders in command arguments
		arg = strings.ReplaceAll(arg, "{download_path}", downloadPath)

		// Add more replacements as needed:
		// - {install_dir} for installation directory
		// - {product_id} for product ID
		// - etc.

		installCmd[i] = arg
	}

	m.logger.Infof("Installing %s using command: %s", dep.Name, strings.Join(installCmd, " "))

	// Execute installation command
	cmd := exec.Command(installCmd[0], installCmd[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installation failed: %w, output: %s", err, output)
	}

	m.logger.Infof("Successfully installed %s", dep.Name)
	return nil
}

// VerifyDependency performs a thorough check of an installed dependency
func (m *Manager) VerifyDependency(dep *Dependency) (*DependencyStatus, error) {
	status := &DependencyStatus{
		Name:      dep.Name,
		Installed: false,
	}

	// Get platform-specific configuration
	platformConfig, err := m.GetPlatformConfig(dep)
	if err != nil {
		status.Error = err
		return status, err
	}

	// Check if verify command is provided
	if len(platformConfig.Commands.Verify) == 0 {
		status.Error = fmt.Errorf("no verification command provided for dependency: %s", dep.Name)
		return status, status.Error
	}

	// Log the verification attempt
	m.logger.Infof("Verifying dependency: %s", dep.Name)

	// Run verify command with timeout to avoid hanging
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create the command
	cmd := exec.CommandContext(ctx, platformConfig.Commands.Verify[0], platformConfig.Commands.Verify[1:]...)

	// Capture output
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	// Handle timeout separately
	if ctx.Err() == context.DeadlineExceeded {
		status.Error = fmt.Errorf("verification command timed out after 30 seconds")
		return status, status.Error
	}

	// Handle command errors
	if err != nil {
		status.Error = fmt.Errorf("dependency verification failed: %w, output: %s", err, outputStr)
		return status, status.Error
	}

	// Dependency is installed
	status.Installed = true
	m.logger.Infof("Dependency %s is installed", dep.Name)

	// Parse current version from command output
	status.CurrentVersion = outputStr

	// Check if we can extract a cleaner version
	version := extractVersion(outputStr)
	if version != "" {
		status.CurrentVersion = version
	}

	// Check if update is needed
	if dep.Version.Required != "" {
		updateType, err := CheckVersionUpdate(status.CurrentVersion, dep.Version.Required)
		if err != nil {
			status.Error = err
			m.logger.Errorf("Failed to check version update: %v", err)
		} else {
			status.RequiredUpdate = updateType
			if updateType != NoUpdate {
				m.logger.Infof("Dependency %s requires a %s (current: %s, required: %s)",
					dep.Name, updateType, status.CurrentVersion, dep.Version.Required)
			}
		}
	}

	// Check if current version is compatible with constraint
	if dep.Version.Constraint != "" {
		compatible, err := IsVersionCompatible(status.CurrentVersion, dep.Version.Constraint)
		if err != nil {
			status.Error = err
			m.logger.Errorf("Failed to check version compatibility: %v", err)
		} else {
			status.Compatible = compatible
			if !compatible {
				m.logger.Infof("Dependency %s version %s is not compatible with constraint %s",
					dep.Name, status.CurrentVersion, dep.Version.Constraint)
			}
		}
	} else {
		// If no constraint is specified, consider it compatible
		status.Compatible = true
	}

	return status, nil
}

// extractVersion tries to extract a clean semantic version from output text
// This helps with commands that return more than just a version number
func extractVersion(output string) string {
	// Common version patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`v?(\d+\.\d+\.\d+)`),                     // Matches: 1.2.3, v1.2.3
		regexp.MustCompile(`version\s+v?(\d+\.\d+\.\d+)`),           // Matches: version 1.2.3
		regexp.MustCompile(`v?(\d+\.\d+\.\d+)[\-+]([0-9A-Za-z-]+)`), // Matches: 1.2.3-alpha, v1.2.3+build
	}

	for _, pattern := range patterns {
		match := pattern.FindStringSubmatch(output)
		if len(match) >= 2 {
			return match[1] // Return the captured version
		}
	}

	return output // Return the original if no pattern matches
}

func (m *Manager) setupDependencyEnvironment(dep *Dependency) error {
	// Check if dependency has environment settings
	if dep.Environment.Path == nil && len(dep.Environment.Variables) == 0 {
		return nil // No environment to set up
	}

	// Add paths to PATH
	for _, path := range dep.Environment.Path {
		// Expand variables in path
		expandedPath := m.envManager.ExpandVariables(path)
		m.envManager.AddPath(expandedPath)
		m.logger.Debugf("Added %s to PATH for dependency %s", expandedPath, dep.Name)
	}

	// Add environment variables
	for key, value := range dep.Environment.Variables {
		// Expand variables in value
		expandedValue := m.envManager.ExpandVariables(value)
		m.envManager.AddVariable(key, expandedValue)
		m.logger.Debugf("Set environment variable %s=%s for dependency %s", key, expandedValue, dep.Name)
	}

	return nil
}
