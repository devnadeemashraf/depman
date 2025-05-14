# depman - External Dependency Management Solution

[![Go Report Card](https://goreportcard.com/badge/github.com/devnadeemashraf/depman)](https://goreportcard.com/report/github.com/devnadeemashraf/depman)
[![GoDoc](https://godoc.org/github.com/devnadeemashraf/depman?status.svg)](https://godoc.org/github.com/devnadeemashraf/depman)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Overview

`depman` is a Go library for managing external system-level dependencies in your applications. It ensures that all required external tools and applications are properly installed before your main application runs.

Unlike traditional package managers that handle code libraries, `depman` focuses on system dependencies like Node.js, Python, or database servers that your application needs to function properly.

With a single YAML configuration file, `depman` validates, installs, and verifies external dependencies across Windows, Linux, and macOS, providing a seamless bootstrapping experience.

## Features

- **Cross-platform support** for Windows, Linux, and macOS
- **Semantic versioning** to specify exact versions or version constraints
- **Dependency discovery** across standard system locations
- **Installation verification** to ensure dependencies are correctly installed
- **Customizable installation commands** for platform-specific needs
- **Environment variable management** to configure dependencies properly

## Installation

To add `depman` to your Go project:

```bash
go get github.com/devnadeemashraf/depman
```

## Quick Start

### 1. Create a dependency configuration file

Create an `app-dependencies.yml` file in your project root:

```yaml
version: "1.0"
name: "My Application"
description: "Dependencies for my application"
dependencies:
  - name: "nodejs"
    description: "JavaScript runtime"
    version:
      required: "16.15.1"
      constraint: "^16.0.0"
    platforms:
      windows:
        installer:
          type: "msi"
          url: "https://nodejs.org/dist/v16.15.1/node-v16.15.1-x64.msi"
        commands:
          install: ["msiexec", "/i", "{download_path}", "/quiet"]
          verify: ["node", "--version"]
          uninstall: ["msiexec", "/x", "{product_id}", "/quiet"]
      linux:
        installer:
          type: "binary"
          url: "https://nodejs.org/dist/v16.15.1/node-v16.15.1-linux-x64.tar.xz"
        commands:
          install: ["tar", "-xf", "{download_path}", "-C", "/usr/local/"]
          verify: ["node", "--version"]
      darwin:
        installer:
          type: "pkg"
          url: "https://nodejs.org/dist/v16.15.1/node-v16.15.1.pkg"
        commands:
          install: ["installer", "-pkg", "{download_path}", "-target", "/"]
          verify: ["node", "--version"]
```

### 2. Use `depman` in your application

Import and use `depman` to ensure dependencies are installed at application startup:

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/devnadeemashraf/depman/pkg/depman"
)

func main() {
	// Initialize dependency manager
	manager, err := depman.NewManager("")
	if err != nil {
		log.Fatalf("Failed to initialize dependency manager: %v", err)
	}

	// Check and install dependencies
	results, err := manager.EnsureDependencies()
	if err != nil {
		log.Fatalf("Dependency error: %v", err)
	}

	// Check if all dependencies are satisfied
	for name, status := range results {
		if !status.Installed || !status.Compatible {
			log.Fatalf("Dependency %s is not properly installed or incompatible", name)
		}
	}

	// Continue with your application logic
	fmt.Println("All dependencies are installed and verified!")
	// ...
}
```

## API Reference

### Core Functions

#### NewManager

```go
func NewManager(configPath string) (*Manager, error)
```

Creates a new dependency manager using the specified configuration file path. If `configPath` is empty, it searches for the file in standard locations.

#### EnsureDependencies

```go
func (m *Manager) EnsureDependencies() (map[string]*DependencyStatus, error)
```

Checks and installs all dependencies if needed, returning their status.

#### CheckDependency

```go
func (m *Manager) CheckDependency(dep *Dependency) (*DependencyStatus, error)
```

Checks if a specific dependency is installed and compatible.

#### InstallDependency

```go
func (m *Manager) InstallDependency(dep *Dependency) error
```

Installs a specific dependency according to platform requirements.

### Configuration File Format

The `app-dependencies.yml` file defines all the dependencies your project needs:

```yaml
version: "1.0" # Configuration file version
name: "Application Name" # Your application name
description: "Description" # Brief description

dependencies:
  - name: "dependency-name" # Unique identifier for the dependency
    description: "..." # What this dependency is used for
    version:
      required: "1.2.3" # Exact version required
      constraint: "^1.2.0" # Semver constraint (flexible version range)
    platforms:
      windows: # Windows-specific configuration
        installer:
          type: "msi" # Installation type (msi, exe, zip, etc.)
          url: "https://..." # Download URL
          checksum: "sha256:..." # Verification checksum
        commands:
          install: ["...", "..."] # How to install the dependency
          verify: ["...", "..."] # How to verify the installation
          uninstall: ["...", "..."] # How to uninstall
      linux:# Linux-specific configuration
        # Similar structure as Windows
      darwin:# macOS-specific configuration
        # Similar structure as Windows
    environment:
      path: ["{install_dir}/bin"] # Paths to add to PATH
      variables: # Environment variables to set
        KEY: "value"
    dependencies: [] # Other dependencies this one requires
```

## Advanced Usage

### Accessing Dependency Status

```go
package main

import (
	"fmt"
	"github.com/devnadeemashraf/depman/pkg/depman"
)

func main() {
	// Create a new dependency manager
	manager, _ := depman.NewManager("")

	// Just check dependencies without installing
	statuses, _ := manager.CheckAllDependencies()

	// Print status report
	for name, status := range statuses {
		fmt.Printf("Dependency: %s\n", name)
		fmt.Printf("  Installed: %v\n", status.Installed)
		fmt.Printf("  Current Version: %s\n", status.CurrentVersion)
		fmt.Printf("  Update Required: %s\n", status.RequiredUpdate)
		fmt.Printf("  Compatible: %v\n", status.Compatible)

		if status.Error != nil {
			fmt.Printf("  Error: %v\n", status.Error)
		}
		fmt.Println()
	}
}
```

### Custom Dependency Path

```go
// Use a custom path for the dependency configuration
manager, err := depman.NewManager("./config/dependencies.yml")
```

## Development

### Requirements

- Go 1.19 or higher
- Make (for using the provided Makefile)
- golangci-lint (for linting)

### Useful Commands

The project includes a Makefile with several useful commands:

```bash
make help         # Show available commands
make build        # Build the library
make test         # Run tests
make test-verbose # Run tests with verbose output
make coverage     # Generate test coverage report
make lint         # Run linters
make clean        # Clean up build artifacts
```

_Note: You should have make installed on your system to use the Makefile._

### Project Structure

```
depman/
├── internal/       # Private application packages
├── pkg/
│   └── depman/     # Public API packages
├── .vscode/        # VS Code settings
├── coverage/       # Test coverage reports
├── Makefile        # Build automation
├── go.mod          # Go modules definition
└── README.md       # This file
```

## FAQ

**Q: How is this different from package managers like NPM, Pip, or Go Modules?**
A: `depman` focuses on system-level dependencies (like Node.js, Python, or Git), not code libraries used within your application.

**Q: Does depman handle permission issues for system-level installations?**
A: `depman` will attempt to install dependencies with the current user's permissions. For system-level installations that require elevated privileges, you may need to run your application with appropriate permissions.

**Q: What happens if a dependency can't be installed?**
A: `depman` returns detailed error information that your application can use to provide appropriate feedback to users.

**Q: Can I use this to bootstrap CI/CD environments?**
A: Yes! `depman` is perfect for ensuring CI/CD environments have the correct dependencies installed.

## Contributing

Contributions are welcome! Here's how you can contribute:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

Please read CONTRIBUTING.md for detailed contribution guidelines.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [Masterminds/semver](https://github.com/Masterminds/semver) for semantic versioning support
- [yaml.v3](https://gopkg.in/yaml.v3) for YAML parsing

---

_Built with ❤️ by [Nadeem Ashraf](https://github.com/devnadeemashraf)_
