package environment

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Manager handles environment variable operations
type Manager struct {
	// Map of environment variables to set
	Variables map[string]string

	// Paths to add to the PATH variable
	Paths []string
}

// NewManager creates a new environment manager
func NewManager() *Manager {
	return &Manager{
		Variables: make(map[string]string),
		Paths:     []string{},
	}
}

// AddVariable adds or updates an environment variable
func (m *Manager) AddVariable(key, value string) {
	m.Variables[key] = value
}

// AddPath adds a path to the PATH variable
func (m *Manager) AddPath(path string) {
	// Normalize path for the current OS
	path = filepath.Clean(path)

	// Check if path already exists in our list
	for _, p := range m.Paths {
		if p == path {
			return // Already added
		}
	}

	m.Paths = append(m.Paths, path)
}

// GetUpdatedEnvironment returns a new environment with the applied changes
func (m *Manager) GetUpdatedEnvironment() []string {
	// Start with the current environment
	env := os.Environ()
	result := make([]string, 0, len(env))

	// Track which variables we've updated
	updated := make(map[string]bool)

	// Apply path changes
	if len(m.Paths) > 0 {
		pathVar := "PATH"
		if runtime.GOOS == "windows" {
			// Windows is case-insensitive for env vars, find the actual case
			for _, e := range env {
				if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
					pathVar = e[:strings.Index(e, "=")]
					break
				}
			}
		}

		// Get current PATH value
		currentPath := os.Getenv(pathVar)

		// Add our paths
		newPaths := strings.Join(m.Paths, string(os.PathListSeparator))
		if currentPath != "" {
			newPaths = newPaths + string(os.PathListSeparator) + currentPath
		}

		// Add updated PATH to result
		result = append(result, fmt.Sprintf("%s=%s", pathVar, newPaths))
		updated[pathVar] = true
	}

	// Apply variable changes
	for key, value := range m.Variables {
		result = append(result, fmt.Sprintf("%s=%s", key, value))
		updated[key] = true
	}

	// Add remaining unchanged variables
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			continue
		}

		// Skip variables we've already updated
		if updated[parts[0]] {
			continue
		}

		result = append(result, e)
	}

	return result
}

// ApplyToCurrentProcess applies the environment changes to the current process
func (m *Manager) ApplyToCurrentProcess() error {
	// Set variables
	for key, value := range m.Variables {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %w", key, err)
		}
	}

	// Update PATH
	if len(m.Paths) > 0 {
		pathVar := "PATH"
		if runtime.GOOS == "windows" {
			// Windows is case-insensitive, so we don't need to worry about the case
			pathVar = "PATH"
		}

		currentPath := os.Getenv(pathVar)
		pathsToAdd := strings.Join(m.Paths, string(os.PathListSeparator))

		newPath := ""
		if currentPath != "" {
			newPath = pathsToAdd + string(os.PathListSeparator) + currentPath
		} else {
			newPath = pathsToAdd
		}

		if err := os.Setenv(pathVar, newPath); err != nil {
			return fmt.Errorf("failed to update PATH: %w", err)
		}
	}

	return nil
}

// ExpandVariables expands placeholders in a string using the current variables
func (m *Manager) ExpandVariables(text string) string {
	result := text

	// Replace our variables
	for key, value := range m.Variables {
		result = strings.ReplaceAll(result, fmt.Sprintf("{%s}", key), value)
	}

	// Replace environment variables
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		result = strings.ReplaceAll(result, fmt.Sprintf("{%s}", parts[0]), parts[1])
	}

	return result
}
