package depman

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindDependencyFile(t *testing.T) {
	// Create a temporary directory for our tests
	tempDir, err := os.MkdirTemp("", "depman-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test dependency file
	testFile := filepath.Join(tempDir, "app-dependencies.yml")
	if err := os.WriteFile(testFile, []byte("version: \"1.0\"\nname: \"Test App\""), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a nested config directory
	configDir := filepath.Join(tempDir, "config")
	if err := os.Mkdir(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create a test dependency file in the config directory
	nestedFile := filepath.Join(configDir, "app-dependencies.yml")
	if err := os.WriteFile(nestedFile, []byte("version: \"1.0\"\nname: \"Nested App\""), 0644); err != nil {
		t.Fatalf("Failed to create nested test file: %v", err)
	}

	// Change to the temp directory for the test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Test cases
	testCases := []struct {
		name         string
		customPath   string
		expectError  bool
		expectedPath string
	}{
		{
			name:         "Find in current directory",
			customPath:   "",
			expectError:  false,
			expectedPath: "app-dependencies.yml",
		},
		{
			name:         "Find with custom path",
			customPath:   testFile,
			expectError:  false,
			expectedPath: testFile,
		},
		{
			name:         "Find in config directory",
			customPath:   "config",
			expectError:  false,
			expectedPath: filepath.Join("config", "app-dependencies.yml"),
		},
		{
			name:         "Error on non-existent file",
			customPath:   "not-exists.yml",
			expectError:  true,
			expectedPath: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path, err := FindDependencyFile(tc.customPath)

			// Check error expectation
			if tc.expectError && err == nil {
				t.Errorf("Expected an error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Did not expect an error but got: %v", err)
			}

			// If we don't expect an error, check the path
			if !tc.expectError {
				if !filepath.IsAbs(path) {
					// Convert to absolute for comparison if it's relative
					absPath, _ := filepath.Abs(path)
					path = absPath
				}

				expectedAbsPath, _ := filepath.Abs(tc.expectedPath)
				if path != expectedAbsPath {
					t.Errorf("Expected path %s but got %s", expectedAbsPath, path)
				}
			}
		})
	}
}

func TestLoadDependencyConfig(t *testing.T) {
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
`
	validFile := filepath.Join(tempDir, "valid.yml")
	if err := os.WriteFile(validFile, []byte(validYAML), 0644); err != nil {
		t.Fatalf("Failed to create valid test file: %v", err)
	}

	// Create an invalid test dependency file
	invalidYAML := `
version: "1.0"
name: "Invalid App"
dependencies:
  - name: "invalid-dep"
    description: 123  # This should be a string
`
	invalidFile := filepath.Join(tempDir, "invalid.yml")
	if err := os.WriteFile(invalidFile, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to create invalid test file: %v", err)
	}

	// Test cases
	testCases := []struct {
		name        string
		path        string
		expectError bool
		appName     string
	}{
		{
			name:        "Load valid config",
			path:        validFile,
			expectError: false,
			appName:     "Test App",
		},
		{
			name:        "Error on invalid config",
			path:        invalidFile,
			expectError: true,
			appName:     "",
		},
		{
			name:        "Error on non-existent file",
			path:        filepath.Join(tempDir, "not-exists.yml"),
			expectError: true,
			appName:     "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := LoadDependencyConfig(tc.path)

			// Check error expectation
			if tc.expectError && err == nil {
				t.Errorf("Expected an error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Did not expect an error but got: %v", err)
			}

			// If we don't expect an error, check the config
			if !tc.expectError {
				if config == nil {
					t.Fatalf("Config is nil despite no error")
				}
				if config.Name != tc.appName {
					t.Errorf("Expected app name %s but got %s", tc.appName, config.Name)
				}
			}
		})
	}
}
