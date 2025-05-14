package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/devnadeemashraf/depman/internal/logger"
	"github.com/devnadeemashraf/depman/pkg/depman"
	"github.com/spf13/cobra"
)

var (
	// Versioning
	version = "dev"

	// Flags
	configPath   string
	platformFlag string
	logLevel     string
	verbose      bool
	outputFile   string
	force        bool

	// Root command
	rootCmd = &cobra.Command{
		Use:   "depman",
		Short: "Depman is a dependency manager for applications",
		Long: `Depman is a dependency manager that helps applications manage
external system dependencies like tools, runtimes, and libraries.

It can check for, install, and verify dependencies on various platforms.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set log level from flags
			if verbose {
				logLevel = "debug"
			}
		},
	}

	// Check command
	checkCmd = &cobra.Command{
		Use:   "check",
		Short: "Check dependencies without installing them",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck()
		},
	}

	// Ensure command
	ensureCmd = &cobra.Command{
		Use:   "ensure",
		Short: "Ensure all dependencies are installed and up to date",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnsure()
		},
	}

	// List command
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all dependencies in the configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList()
		},
	}

	// Version command
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show depman version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Depman version %s\n", version)
		},
	}

	// Generate command
	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate a template dependency configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate()
		},
	}
)

func main() {
	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add flags to root command
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to dependency configuration file")
	rootCmd.PersistentFlags().StringVarP(&platformFlag, "platform", "p", "", "Override platform detection (windows, linux, darwin)")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Add commands
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(ensureCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)

	// Add Generate Command
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "app-dependencies.yml", "Output file path")
	generateCmd.Flags().BoolVarP(&force, "force", "f", false, "Force overwrite existing file")
}

// createManager creates a new dependency manager with the specified options
func createManager() (*depman.Manager, error) {
	// Set up options
	var options []depman.Option

	// Set platform if specified
	if platformFlag != "" {
		options = append(options, depman.WithPlatform(platformFlag))
	}

	// Set log level
	loggerLevel := logger.LevelInfo
	switch strings.ToLower(logLevel) {
	case "debug":
		loggerLevel = logger.LevelDebug
	case "info":
		loggerLevel = logger.LevelInfo
	case "warn":
		loggerLevel = logger.LevelWarn
	case "error":
		loggerLevel = logger.LevelError
	}
	options = append(options, depman.WithLogLevel(loggerLevel))

	// Create manager
	return depman.NewManager(configPath, options...)
}

// runCheck checks dependencies without installing them
func runCheck() error {
	manager, err := createManager()
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// Check dependencies
	statuses, err := manager.CheckAllDependencies()
	if err != nil {
		return fmt.Errorf("failed to check dependencies: %w", err)
	}

	// Print results
	fmt.Println("Dependency Status:")
	fmt.Println("==================")

	allOk := true
	for name, status := range statuses {
		fmt.Printf("- %s: ", name)

		if status.Installed {
			fmt.Printf("Installed (v%s)", status.CurrentVersion)
			if status.RequiredUpdate != depman.NoUpdate {
				fmt.Printf(" [%s needed]", status.RequiredUpdate)
				allOk = false
			}
			if !status.Compatible {
				fmt.Printf(" [Incompatible]")
				allOk = false
			}
		} else {
			fmt.Printf("Not installed")
			allOk = false
		}

		if status.Error != nil {
			fmt.Printf(" [Error: %v]", status.Error)
			allOk = false
		}

		fmt.Println()
	}

	if !allOk {
		return fmt.Errorf("one or more dependencies need attention")
	}

	return nil
}

// runEnsure ensures all dependencies are installed and up to date
func runEnsure() error {
	manager, err := createManager()
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// Ensure dependencies
	statuses, err := manager.EnsureDependencies()
	if err != nil {
		return fmt.Errorf("failed to ensure dependencies: %w", err)
	}

	// Print results
	fmt.Println("Dependency Status:")
	fmt.Println("==================")

	for name, status := range statuses {
		fmt.Printf("- %s: ", name)

		if status.Installed {
			fmt.Printf("Installed (v%s)", status.CurrentVersion)
			if status.Compatible {
				fmt.Printf(" [Compatible]")
			} else {
				fmt.Printf(" [Incompatible]")
			}
		} else {
			fmt.Printf("Failed to install")
		}

		if status.Error != nil {
			fmt.Printf(" [Error: %v]", status.Error)
		}

		fmt.Println()
	}

	return nil
}

// runList lists all dependencies in the configuration
func runList() error {
	manager, err := createManager()
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// Get configuration
	config := manager.Config

	fmt.Printf("Application: %s\n", config.Name)
	if config.Description != "" {
		fmt.Printf("Description: %s\n", config.Description)
	}
	fmt.Printf("Configuration Version: %s\n", config.Version)
	fmt.Println()

	fmt.Println("Dependencies:")
	fmt.Println("=============")

	for _, dep := range config.Dependencies {
		fmt.Printf("- %s: %s\n", dep.Name, dep.Description)
		fmt.Printf("  Version: %s", dep.Version.Required)
		if dep.Version.Constraint != "" {
			fmt.Printf(" (Constraint: %s)", dep.Version.Constraint)
		}
		fmt.Println()

		// Show platforms
		platforms := make([]string, 0, len(dep.Platforms))
		for platform := range dep.Platforms {
			platforms = append(platforms, platform)
		}
		if len(platforms) > 0 {
			fmt.Printf("  Platforms: %s\n", strings.Join(platforms, ", "))
		}

		// Show dependencies if any
		if len(dep.Dependencies) > 0 {
			fmt.Printf("  Depends on: %s\n", strings.Join(dep.Dependencies, ", "))
		}

		fmt.Println()
	}

	return nil
}

// Add this function to handle the generate command
func runGenerate() error {
	// Check if file already exists
	if _, err := os.Stat(outputFile); err == nil {
		// File exists
		if !force {
			// Prompt user for confirmation
			fmt.Printf("File %s already exists. Overwrite? [y/N] ", outputFile)
			var response string
			fmt.Scanln(&response)

			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Operation cancelled.")
				return nil
			}
		}
	}

	// Template content
	template := `# Dependency configuration for depman
version: "1.0"
name: "My Application"
description: "Application dependencies configuration"

dependencies:
  - name: "example-tool"
    description: "Example tool dependency"
    version:
      required: "1.0.0"
      constraint: "^1.0.0"
    platforms:
      windows:
        installer:
          type: "msi"
          url: "https://example.com/tool-1.0.0-windows.msi"
          checksum: "sha256:0000000000000000000000000000000000000000000000000000000000000000"
        commands:
          install: ["msiexec", "/i", "{download_path}", "/quiet"]
          verify: ["example-tool", "--version"]
          uninstall: ["msiexec", "/x", "{download_path}", "/quiet"]
      linux:
        installer:
          type: "tarball"
          url: "https://example.com/tool-1.0.0-linux.tar.gz"
          checksum: "sha256:0000000000000000000000000000000000000000000000000000000000000000"
        commands:
          install: ["tar", "-xzf", "{download_path}", "-C", "/usr/local/bin"]
          verify: ["example-tool", "--version"]
      darwin:
        installer:
          type: "pkg"
          url: "https://example.com/tool-1.0.0-macos.pkg"
          checksum: "sha256:0000000000000000000000000000000000000000000000000000000000000000"
        commands:
          install: ["installer", "-pkg", "{download_path}", "-target", "/"]
          verify: ["example-tool", "--version"]
    environment:
      path: ["/usr/local/bin"]
      variables:
        EXAMPLE_HOME: "/usr/local/example"
`

	// Write the template to the file
	err := os.WriteFile(outputFile, []byte(template), 0644)
	if err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	fmt.Printf("Dependency configuration template created at %s\n", outputFile)
	fmt.Println("Customize it with your actual dependencies and requirements.")

	return nil
}
