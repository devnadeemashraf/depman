# Makefile for depman - Dependency Manager
#
# This Makefile provides commands for building, testing, and managing the depman project.
# Use 'make help' to see all available commands.

.PHONY: help build build-all build-windows build-linux build-macos test clean run install lint fmt

# Configuration variables
BINARY_NAME=depman
BUILD_DIR=build
MAIN_PACKAGE=./cmd/depman
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=${VERSION}"

# Display help information
help:
	@echo "depman Makefile commands:"
	@echo "  build         - Build depman for current platform"
	@echo "  build-all     - Build depman for Windows, Linux and macOS"
	@echo "  build-windows - Build depman for Windows"
	@echo "  build-linux   - Build depman for Linux"
	@echo "  build-macos   - Build depman for macOS"
	@echo "  test          - Run all tests"
	@echo "  clean         - Remove build artifacts"
	@echo "  run           - Run depman (use ARGS='command' to pass arguments)"
	@echo "  install       - Install depman locally"
	@echo "  lint          - Run linters"
	@echo "  fmt           - Format code"

# Build the depman binary for current platform
build:
	@echo "Building depman v${VERSION}..."
	@mkdir -p $(BUILD_DIR)
	go build ${LDFLAGS} -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# Build for multiple platforms
build-all: build-windows build-linux build-macos

# Build for Windows
build-windows:
	@echo "Building for Windows v${VERSION}..."
	@mkdir -p $(BUILD_DIR)/windows
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o $(BUILD_DIR)/windows/$(BINARY_NAME).exe $(MAIN_PACKAGE)

# Build for Linux
build-linux:
	@echo "Building for Linux v${VERSION}..."
	@mkdir -p $(BUILD_DIR)/linux
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o $(BUILD_DIR)/linux/$(BINARY_NAME) $(MAIN_PACKAGE)

# Build for macOS
build-macos:
	@echo "Building for macOS v${VERSION}..."
	@mkdir -p $(BUILD_DIR)/macos
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o $(BUILD_DIR)/macos/$(BINARY_NAME) $(MAIN_PACKAGE)

# Run tests with coverage
test:
	@echo "Running tests..."
	go test -v -cover ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

# Run depman directly (use make run ARGS="command --flag")
run:
	@go run ${LDFLAGS} $(MAIN_PACKAGE) $(ARGS)

# Install depman locally
install:
	@echo "Installing depman v${VERSION}..."
	go install ${LDFLAGS} $(MAIN_PACKAGE)

# Run linters
lint:
	@echo "Running linters..."
	go vet ./...
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping additional lint checks"; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
