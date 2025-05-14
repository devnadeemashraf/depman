package depman

import (
	"fmt"
)

// EnsureDependencies checks and installs all dependencies if needed
// This is the main function that most applications should use
func (m *Manager) EnsureDependencies() (map[string]*DependencyStatus, error) {
	// First check if dependencies are properly configured
	if err := m.validateConfiguration(); err != nil {
		return nil, fmt.Errorf("invalid dependency configuration: %w", err)
	}

	// Check current status of all dependencies
	statuses, err := m.CheckAllDependencies()
	if err != nil {
		return statuses, err
	}

	// Install or update dependencies as needed
	for name, status := range statuses {
		// Skip if already installed and compatible
		if status.Installed && status.Compatible && status.RequiredUpdate == NoUpdate {
			continue
		}

		// Find the dependency definition
		var dep *Dependency
		for i := range m.Config.Dependencies {
			if m.Config.Dependencies[i].Name == name {
				dep = &m.Config.Dependencies[i]
				break
			}
		}

		if dep == nil {
			return statuses, fmt.Errorf("dependency '%s' not found in configuration", name)
		}

		// Install or update the dependency
		if err := m.installDependency(dep); err != nil {
			status.Error = err
			status.Installed = false
			return statuses, err
		}

		// Set up environment for the dependency
		if err := m.setupDependencyEnvironment(dep); err != nil {
			m.logger.Warnf("Failed to set up environment for dependency %s: %v", dep.Name, err)
		}

		// Verify the installation worked
		updatedStatus, err := m.CheckDependency(dep)
		if err != nil {
			return statuses, err
		}

		// Update the status in our results
		statuses[name] = updatedStatus
	}

	// Apply environment changes to the current process
	if err := m.envManager.ApplyToCurrentProcess(); err != nil {
		m.logger.Warnf("Failed to apply environment changes: %v", err)
	}

	return statuses, nil
}

// Add a method to get the updated environment
func (m *Manager) GetUpdatedEnvironment() []string {
	return m.envManager.GetUpdatedEnvironment()
}

// CheckAllDependencies checks the status of all dependencies without installing
// Use this to inspect what would be installed/updated
func (m *Manager) CheckAllDependencies() (map[string]*DependencyStatus, error) {
	results := make(map[string]*DependencyStatus)

	// Validate dependencies configuration
	errors := m.validateDependencies()
	if len(errors) > 0 {
		return nil, fmt.Errorf("dependency configuration errors: %v", errors)
	}

	// Check each dependency
	for _, dep := range m.Config.Dependencies {
		status, _ := m.CheckDependency(&dep) // We still want to return status even if there's an error
		results[dep.Name] = status
	}

	return results, nil
}

// validateConfiguration performs overall configuration validation
func (m *Manager) validateConfiguration() error {
	// Check if config is loaded
	if m.Config == nil {
		return fmt.Errorf("no dependency configuration loaded")
	}

	// Validate dependencies
	errors := m.validateDependencies()
	if len(errors) > 0 {
		return fmt.Errorf("dependency validation errors: %v", errors)
	}

	return nil
}

// WithLogger sets a custom logger for the dependency manager
func WithLogger(log Logger) Option {
	return func(m *Manager) {
		m.logger = log
	}
}
