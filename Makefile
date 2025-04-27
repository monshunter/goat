.PHONY: all build clean test lint cover fmt help install uninstall clean-cache kill-build release package-release

# Set Go environment variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOLINT=golangci-lint
GOFMT=$(GOCMD) fmt
BINARY_NAME=goat
BUILD_DIR=bin
GOPATH ?= $(shell $(GOCMD) env GOPATH)

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date +%FT%T%z)
BUILD_FLAGS = -ldflags "-X main.Version=${VERSION} -X main.Commit=${GIT_COMMIT} -X main.BuildDate=${BUILD_DATE}"

# Default target
all: lint test build

# Build application
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/goat

# Quick build (without test and lint)
build-quick:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/goat

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests and generate coverage report
cover:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Run code check
lint:
	$(GOVET) ./...

# Format code
fmt:
	$(GOFMT) ./...

# Static check
check: lint test

# Install dependency tools
deps:
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install compiled binary
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# Uninstall binary from GOPATH
uninstall:
	rm -f $(GOPATH)/bin/$(BINARY_NAME)

# Clean build files
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out

# Clean Go cache
clean-cache:
	$(GOCMD) clean -cache -modcache -i -r

# Kill all Go processes (when build hangs)
kill-build:
	@echo "Killing all Go processes..."
	@pgrep go | xargs kill -9 2>/dev/null || echo "No Go processes found"

# Build release binaries for multiple platforms
release:
	@echo "Building release binaries..."
	mkdir -p $(BUILD_DIR)/release

	# Linux builds
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)_linux_amd64 ./cmd/goat
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)_linux_arm64 ./cmd/goat
	GOOS=linux GOARCH=386 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)_linux_386 ./cmd/goat

	# macOS builds
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)_darwin_amd64 ./cmd/goat
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)_darwin_arm64 ./cmd/goat

	# Windows builds
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)_windows_amd64.exe ./cmd/goat
	GOOS=windows GOARCH=386 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)_windows_386.exe ./cmd/goat

	@echo "Release binaries built in $(BUILD_DIR)/release/"

# Package release binaries into compressed archives
package-release: release
	@echo "Packaging release binaries..."
	mkdir -p $(BUILD_DIR)/packages

	# Create archives for Linux binaries
	tar -czf $(BUILD_DIR)/packages/$(BINARY_NAME)_$(VERSION)_linux_amd64.tar.gz -C $(BUILD_DIR)/release $(BINARY_NAME)_linux_amd64
	tar -czf $(BUILD_DIR)/packages/$(BINARY_NAME)_$(VERSION)_linux_arm64.tar.gz -C $(BUILD_DIR)/release $(BINARY_NAME)_linux_arm64
	tar -czf $(BUILD_DIR)/packages/$(BINARY_NAME)_$(VERSION)_linux_386.tar.gz -C $(BUILD_DIR)/release $(BINARY_NAME)_linux_386

	# Create archives for macOS binaries
	tar -czf $(BUILD_DIR)/packages/$(BINARY_NAME)_$(VERSION)_darwin_amd64.tar.gz -C $(BUILD_DIR)/release $(BINARY_NAME)_darwin_amd64
	tar -czf $(BUILD_DIR)/packages/$(BINARY_NAME)_$(VERSION)_darwin_arm64.tar.gz -C $(BUILD_DIR)/release $(BINARY_NAME)_darwin_arm64

	# Create archives for Windows binaries
	zip -j $(BUILD_DIR)/packages/$(BINARY_NAME)_$(VERSION)_windows_amd64.zip $(BUILD_DIR)/release/$(BINARY_NAME)_windows_amd64.exe
	zip -j $(BUILD_DIR)/packages/$(BINARY_NAME)_$(VERSION)_windows_386.zip $(BUILD_DIR)/release/$(BINARY_NAME)_windows_386.exe

	@echo "Release packages created in $(BUILD_DIR)/packages/"

# Help information
help:
	@echo "Available commands:"
	@echo "  make build       - Build application"
	@echo "  make build-quick - Quick build (skip tests and lint)"
	@echo "  make test        - Run tests"
	@echo "  make cover       - Generate test coverage report"
	@echo "  make lint        - Run code check"
	@echo "  make fmt         - Format code"
	@echo "  make check       - Run tests and code check"
	@echo "  make deps        - Install dependency tools"
	@echo "  make install     - Install compiled binary"
	@echo "  make uninstall   - Uninstall binary from GOPATH"
	@echo "  make clean       - Clean build files"
	@echo "  make clean-cache - Clean Go cache"
	@echo "  make kill-build  - Kill all Go processes (when build hangs)"
	@echo "  make release     - Build release binaries for multiple platforms"
	@echo "  make package-release - Build and package release binaries"
