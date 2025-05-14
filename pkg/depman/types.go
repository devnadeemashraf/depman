package depman

import (
	"fmt"

	"github.com/devnadeemashraf/depman/internal/environment"
	"github.com/devnadeemashraf/depman/internal/logger"
)

// Version represents dependency version information with semver support
type Version struct {
	Required   string `yaml:"required"`   // Exact version required
	Constraint string `yaml:"constraint"` // Semver constraint (e.g., "^1.2.3", ">=2.0.0", etc.)
}

// Installer contains information about how to install a dependency
type Installer struct {
	Type     string `yaml:"type"`     // Installation type (e.g., "msi", "pkg", "binary")
	URL      string `yaml:"url"`      // URL to download the dependency
	Checksum string `yaml:"checksum"` // Checksum for verification (format: "algorithm:hash")
}

// Commands for different operations on a dependency
type Commands struct {
	Install   []string `yaml:"install"`   // Command to install the dependency
	Verify    []string `yaml:"verify"`    // Command to verify the installation (should output version)
	Uninstall []string `yaml:"uninstall"` // Command to uninstall the dependency
}

// PlatformConfig holds platform-specific configuration
type PlatformConfig struct {
	Installer Installer `yaml:"installer"` // Installer information
	Commands  Commands  `yaml:"commands"`  // Platform-specific commands
}

// Environment variables and paths for a dependency
type Environment struct {
	Path      []string          `yaml:"path"`      // Paths to add to PATH
	Variables map[string]string `yaml:"variables"` // Environment variables to set
}

// Dependency represents a single dependency with all its properties
type Dependency struct {
	Name         string                    `yaml:"name"`         // Unique name of the dependency
	Description  string                    `yaml:"description"`  // Human-readable description
	Version      Version                   `yaml:"version"`      // Version requirements
	Platforms    map[string]PlatformConfig `yaml:"platforms"`    // Platform-specific configurations
	Environment  Environment               `yaml:"environment"`  // Environment configuration
	Dependencies []string                  `yaml:"dependencies"` // Dependencies of this dependency
}

// DependencyConfig represents the entire dependency configuration file
type DependencyConfig struct {
	Version      string       `yaml:"version"`      // Configuration format version
	Name         string       `yaml:"name"`         // Application name
	Description  string       `yaml:"description"`  // Application description
	Dependencies []Dependency `yaml:"dependencies"` // List of dependencies
}

// Manager handles dependency management operations
type Manager struct {
	Config     *DependencyConfig    // Dependency configuration
	ConfigPath string               // Path to configuration file
	Platform   string               // Current platform (windows, linux, darwin)
	logger     Logger               // Logger for operations
	envManager *environment.Manager // Environment manager
}

// UpdateType represents the type of update needed
type UpdateType int

const (
	NoUpdate UpdateType = iota
	PatchUpdate
	MinorUpdate
	MajorUpdate
)

func (u UpdateType) String() string {
	return [...]string{"No Update", "Patch Update", "Minor Update", "Major Update"}[u]
}

// DependencyStatus represents the installation status of a dependency
type DependencyStatus struct {
	Name           string     // Name of the dependency
	Installed      bool       // Whether the dependency is installed
	CurrentVersion string     // Current installed version
	RequiredUpdate UpdateType // Type of update required
	Compatible     bool       // Whether the current version is compatible with constraints
	Error          error      // Any error that occurred during checking
}

// Option represents a configuration option for the dependency manager
type Option func(*Manager)

// WithPlatform sets a specific platform to use instead of auto-detecting
func WithPlatform(platform string) Option {
	return func(m *Manager) {
		m.Platform = platform
	}
}

// WithLogLevel sets the log level for the dependency manager
func WithLogLevel(level logger.Level) Option {
	return func(m *Manager) {
		if l, ok := m.logger.(*logger.Logger); ok {
			m.logger = l.WithLevel(level)
		}
	}
}

// Logger interface for logging dependency operations
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// defaultLogger is a simple logger that prints to stdout
type defaultLogger struct{}

func (l *defaultLogger) Infof(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}

func (l *defaultLogger) Debugf(format string, args ...interface{}) {
	// By default, debug logs are disabled
}
