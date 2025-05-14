package manager

import (
	"errors"
	"sync"
)

type Dependency struct {
	Name    string
	Version string
}

type Manager struct {
	mu           sync.Mutex
	dependencies map[string]Dependency
}

func NewManager() *Manager {
	return &Manager{
		dependencies: make(map[string]Dependency),
	}
}

func (m *Manager) Add(dep Dependency) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.dependencies[dep.Name]; exists {
		return errors.New("dependency already exists")
	}

	m.dependencies[dep.Name] = dep
	return nil
}

func (m *Manager) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.dependencies[name]; !exists {
		return errors.New("dependency not found")
	}

	delete(m.dependencies, name)
	return nil
}

func (m *Manager) List() []Dependency {
	m.mu.Lock()
	defer m.mu.Unlock()

	deps := make([]Dependency, 0, len(m.dependencies))
	for _, dep := range m.dependencies {
		deps = append(deps, dep)
	}
	return deps
}
