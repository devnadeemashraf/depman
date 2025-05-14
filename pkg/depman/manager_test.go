package depman

import (
	"os"
	"path/filepath"
	"testing"
)

// mockLogger is a simple logger for testing
type mockLogger struct {
	infoLogs  []string
	errorLogs []string
	debugLogs []string
	warnLogs  []string
}

func (l *mockLogger) Infof(format string, args ...interface{}) {
	// No need to actually format for tests
	l.infoLogs = append(l.infoLogs, format)
}

func (l *mockLogger) Errorf(format string, args ...interface{}) {
	l.errorLogs = append(l.errorLogs, format)
}

func (l *mockLogger) Debugf(format string, args ...interface{}) {
	l.debugLogs = append(l.debugLogs, format)
}

func (l *mockLogger) Warnf(format string, args ...interface{}) {
	l.warnLogs = append(l.warnLogs, format)
}

// TestNewManager tests the creation of a new manager
func TestNewManager(t *testing.T) {
	// Create a temporary directory for our tests
	tempDir, err := os.MkdirTemp("", "depman-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a valid test dependency file
	validYAML := `
version: "1.0"
name: "Test App"
description: "Test description"
dependencies:
  - name: "test-dep"
    description: "Test dependency"
    version:
      required: "1.0.0"
      constraint: "^1.0.0"
    platforms:
      windows:
        installer:
          type: "msi"
          url: "https://example.com/test.msi"
        commands:
          install: ["msiexec", "/i", "{download_path}"]
          verify: ["test-dep", "--version"]
      linux:
        installer:
          type: "binary"
          url: "https://example.com/test.tar.gz"
        commands:
          install: ["tar", "-xzf", "{download_path}"]
          verify: ["test-dep", "--version"]
      darwin:
        installer:
          type: "pkg"
          url: "https://example.com/test.pkg"
        commands:
          install: ["installer", "-pkg", "{download_path}", "-target", "/"]
          verify: ["test-dep", "--version"]
`
	validFile := filepath.Join(tempDir, "valid.yml")
	if err := os.WriteFile(validFile, []byte(validYAML), 0644); err != nil {
		t.Fatalf("Failed to create valid test file: %v", err)
	}

	// Test creating a manager with a valid configuration
	t.Run("Create with valid config", func(t *testing.T) {
		manager, err := NewManager(validFile)
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		if manager == nil {
			t.Fatalf("Manager is nil despite no error")
		}

		if manager.Config == nil {
			t.Fatalf("Manager.Config is nil")
		}

		if manager.Config.Name != "Test App" {
			t.Errorf("Expected app name 'Test App' but got '%s'", manager.Config.Name)
		}

		if len(manager.Config.Dependencies) != 1 {
			t.Errorf("Expected 1 dependency but got %d", len(manager.Config.Dependencies))
		}

		if manager.Config.Dependencies[0].Name != "test-dep" {
			t.Errorf("Expected dependency name 'test-dep' but got '%s'", manager.Config.Dependencies[0].Name)
		}
	})

	// Test creating a manager with options
	t.Run("Create with options", func(t *testing.T) {
		mockLog := &mockLogger{}

		manager, err := NewManager(validFile,
			WithPlatform("linux"),
			WithLogger(mockLog))

		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		if manager.Platform != "linux" {
			t.Errorf("Expected platform 'linux' but got '%s'", manager.Platform)
		}

		if manager.logger != mockLog {
			t.Errorf("Logger not set correctly")
		}
	})

	// Test with a non-existent file
	t.Run("Error on non-existent file", func(t *testing.T) {
		_, err := NewManager(filepath.Join(tempDir, "does-not-exist.yml"))
		if err == nil {
			t.Errorf("Expected an error but got none")
		}
	})
}

// TestGetPlatformConfig tests retrieving platform-specific configuration
func TestGetPlatformConfig(t *testing.T) {
	// Create a dependency with platform configurations
	dep := &Dependency{
		Name: "test-dep",
		Platforms: map[string]PlatformConfig{
			"windows": {
				Installer: Installer{
					Type: "msi",
					URL:  "https://example.com/test.msi",
				},
				Commands: Commands{
					Install: []string{"msiexec", "/i", "{download_path}"},
					Verify:  []string{"test-dep", "--version"},
				},
			},
			"linux": {
				Installer: Installer{
					Type: "binary",
					URL:  "https://example.com/test.tar.gz",
				},
				Commands: Commands{
					Install: []string{"tar", "-xzf", "{download_path}"},
					Verify:  []string{"test-dep", "--version"},
				},
			},
		},
	}

	// Create a manager with Windows platform
	managerWindows := &Manager{
		Platform: "windows",
		logger:   &mockLogger{},
	}

	// Test retrieving Windows configuration
	t.Run("Get Windows config", func(t *testing.T) {
		config, err := managerWindows.GetPlatformConfig(dep)
		if err != nil {
			t.Fatalf("Failed to get platform config: %v", err)
		}

		if config.Installer.Type != "msi" {
			t.Errorf("Expected installer type 'msi' but got '%s'", config.Installer.Type)
		}

		if config.Installer.URL != "https://example.com/test.msi" {
			t.Errorf("Expected URL 'https://example.com/test.msi' but got '%s'", config.Installer.URL)
		}
	})

	// Create a manager with Linux platform
	managerLinux := &Manager{
		Platform: "linux",
		logger:   &mockLogger{},
	}

	// Test retrieving Linux configuration
	t.Run("Get Linux config", func(t *testing.T) {
		config, err := managerLinux.GetPlatformConfig(dep)
		if err != nil {
			t.Fatalf("Failed to get platform config: %v", err)
		}

		if config.Installer.Type != "binary" {
			t.Errorf("Expected installer type 'binary' but got '%s'", config.Installer.Type)
		}

		if config.Installer.URL != "https://example.com/test.tar.gz" {
			t.Errorf("Expected URL 'https://example.com/test.tar.gz' but got '%s'", config.Installer.URL)
		}
	})

	// Create a manager with an unsupported platform
	managerUnsupported := &Manager{
		Platform: "unsupported",
		logger:   &mockLogger{},
	}

	// Test retrieving an unsupported platform configuration
	t.Run("Error on unsupported platform", func(t *testing.T) {
		_, err := managerUnsupported.GetPlatformConfig(dep)
		if err == nil {
			t.Errorf("Expected an error but got none")
		}
	})
}

// TestValidateDependencies tests the dependency validation
func TestValidateDependencies(t *testing.T) {
	// Test with no dependencies
	t.Run("No dependencies", func(t *testing.T) {
		manager := &Manager{
			Config: &DependencyConfig{
				Name:         "Test App",
				Dependencies: []Dependency{},
			},
			Platform: "windows",
		}

		errors := manager.validateDependencies()
		if len(errors) == 0 {
			t.Errorf("Expected an error but got none")
		}
	})

	// Test with missing platform configuration
	t.Run("Missing platform config", func(t *testing.T) {
		manager := &Manager{
			Config: &DependencyConfig{
				Name: "Test App",
				Dependencies: []Dependency{
					{
						Name: "test-dep",
						Platforms: map[string]PlatformConfig{
							"linux": {}, // No windows config
						},
					},
				},
			},
			Platform: "windows",
		}

		errors := manager.validateDependencies()
		if len(errors) == 0 {
			t.Errorf("Expected an error but got none")
		}
	})

	// Test with valid configuration
	t.Run("Valid configuration", func(t *testing.T) {
		manager := &Manager{
			Config: &DependencyConfig{
				Name: "Test App",
				Dependencies: []Dependency{
					{
						Name: "test-dep",
						Version: Version{
							Required: "1.0.0",
						},
						Platforms: map[string]PlatformConfig{
							"windows": {},
						},
					},
				},
			},
			Platform: "windows",
		}

		errors := manager.validateDependencies()
		if len(errors) > 0 {
			t.Errorf("Expected no errors but got: %v", errors)
		}
	})
}
