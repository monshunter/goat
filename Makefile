.PHONY: all build clean test lint cover fmt help install uninstall clean-cache kill-build

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
