package depman

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
)

// LoadDependencyConfig loads and parses the dependency configuration file
func LoadDependencyConfig(path string) (*DependencyConfig, error) {
	// Find the file if path is not provided
	if path == "" {
		var err error
		path, err = FindDependencyFile("")
		if err != nil {
			return nil, err
		}
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read dependency file: %w", err)
	}

	// Parse YAML
	var config DependencyConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse dependency file: %w", err)
	}

	return &config, nil
}

// FindDependencyFile looks for the app-dependencies.yml file in standard locations
func FindDependencyFile(customPath string) (string, error) {
	// If a custom path is provided, check it first
	if customPath != "" {
		if _, err := os.Stat(customPath); err == nil {
			return customPath, nil
		}
		// If custom path has no extension, try with .yml extension
		if !strings.HasSuffix(customPath, ".yml") && !strings.HasSuffix(customPath, ".yaml") {
			withExt := customPath + ".yml"
			if _, err := os.Stat(withExt); err == nil {
				return withExt, nil
			}
		}
	}

	// Standard locations to check
	searchPaths := []string{
		"app-dependencies.yml",           // Current directory
		"config/app-dependencies.yml",    // Config subdirectory
		"../app-dependencies.yml",        // Parent directory
		"../config/app-dependencies.yml", // Parent's config subdirectory
		filepath.Join(os.Getenv("HOME"), ".config/depman/app-dependencies.yml"), // User config directory
	}

	// On Windows, also check AppData
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			searchPaths = append(searchPaths, filepath.Join(appData, "depman", "app-dependencies.yml"))
		}
	}

	// Check each path
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("dependency configuration file not found")
}

// CheckVersionUpdate determines if and what type of update is needed
func CheckVersionUpdate(currentVersion, requiredVersion string) (UpdateType, error) {
	// Parse versions
	current, err := semver.NewVersion(currentVersion)
	if err != nil {
		return NoUpdate, fmt.Errorf("invalid current version '%s': %w", currentVersion, err)
	}

	required, err := semver.NewVersion(requiredVersion)
	if err != nil {
		return NoUpdate, fmt.Errorf("invalid required version '%s': %w", requiredVersion, err)
	}

	// Compare versions
	if current.Equal(required) {
		return NoUpdate, nil
	}

	// Determine update type
	if current.Major() < required.Major() {
		return MajorUpdate, nil
	} else if current.Minor() < required.Minor() {
		return MinorUpdate, nil
	} else if current.Patch() < required.Patch() {
		return PatchUpdate, nil
	}

	// Current version is newer than required
	return NoUpdate, nil
}

// IsVersionCompatible checks if the current version satisfies the constraint
func IsVersionCompatible(currentVersion, constraintStr string) (bool, error) {
	// Parse current version
	version, err := semver.NewVersion(currentVersion)
	if err != nil {
		return false, fmt.Errorf("invalid version '%s': %w", currentVersion, err)
	}

	// Parse constraint
	constraint, err := semver.NewConstraint(constraintStr)
	if err != nil {
		return false, fmt.Errorf("invalid constraint '%s': %w", constraintStr, err)
	}

	// Check if version satisfies constraint
	return constraint.Check(version), nil
}
