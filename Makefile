.PHONY: build test install clean release

# Version info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS = -X 'github.com/Torres-Atlantic/multikey-cli/internal/cli.Version=$(VERSION)' \
          -X 'github.com/Torres-Atlantic/multikey-cli/internal/cli.GitCommit=$(GIT_COMMIT)' \
          -X 'github.com/Torres-Atlantic/multikey-cli/internal/cli.BuildDate=$(BUILD_DATE)'

# Build directory
BUILD_DIR = build

# Default target
all: build

# Build for current platform
build:
	@echo "Building multikey..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/multikey ./cmd/multikey
	@echo "Built: $(BUILD_DIR)/multikey"

# Build for multiple platforms
release:
	@echo "Building releases..."
	@mkdir -p $(BUILD_DIR)
	
	# macOS
	@echo "Building for darwin/amd64..."
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/multikey-darwin-amd64 ./cmd/multikey
	
	@echo "Building for darwin/arm64..."
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/multikey-darwin-arm64 ./cmd/multikey
	
	# Linux
	@echo "Building for linux/amd64..."
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/multikey-linux-amd64 ./cmd/multikey
	
	@echo "Building for linux/arm64..."
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/multikey-linux-arm64 ./cmd/multikey
	
	@echo "Releases built in $(BUILD_DIR)/"

# Create release archives for GitHub
release-archives: release
	@echo "Creating release archives..."
	@cd $(BUILD_DIR) && \
		tar -czf multikey-darwin-amd64.tar.gz multikey-darwin-amd64 && \
		tar -czf multikey-darwin-arm64.tar.gz multikey-darwin-arm64 && \
		tar -czf multikey-linux-amd64.tar.gz multikey-linux-amd64 && \
		tar -czf multikey-linux-arm64.tar.gz multikey-linux-arm64
	@echo "Release archives created in $(BUILD_DIR)/"

# Install locally
install: build
	@echo "Installing multikey..."
	go install -ldflags "$(LDFLAGS)" ./cmd/multikey
	@echo "Installed: $$(go env GOPATH)/bin/multikey"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	go clean

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

