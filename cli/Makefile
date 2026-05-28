# Shipmate CLI Makefile
# Usage:
#   make build        - Build for current platform
#   make build-all    - Build for all platforms
#   make test         - Run tests
#   make install      - Install to /usr/local/bin
#   make clean        - Remove build artifacts
#   make release      - Create a new release (requires VERSION env)

BINARY_NAME=shipmate
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"
BUILD_DIR=build

.PHONY: build build-all test install clean release

# Build for current platform
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Build for all platforms
build-all: clean
	@mkdir -p $(BUILD_DIR)
	@echo "Building for Linux amd64..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 .
	@echo "Building for Linux arm64..."
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_linux_arm64 .
	@echo "Building for macOS amd64..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 .
	@echo "Building for macOS arm64..."
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_arm64 .
	@echo "Building for Windows amd64..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe .
	@echo ""
	@echo "All binaries built in $(BUILD_DIR)/"
	@ls -lh $(BUILD_DIR)/

# Generate checksums for all binaries
checksums: build-all
	cd $(BUILD_DIR) && sha256sum $(BINARY_NAME)_* > checksums.txt
	@echo ""
	@echo "Checksums:"
	@cat $(BUILD_DIR)/checksums.txt

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...
	go vet ./...
	@echo ""
	@echo "Coverage:"
	@go tool cover -func=coverage.out | tail -1

# Install to /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@if [ -w /usr/local/bin ]; then \
		mv $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME); \
	else \
		sudo mv $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME); \
	fi
	@echo "✓ Installed. Run 'shipmate --help' to get started."

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	rm -rf $(BUILD_DIR)
	rm -f coverage.out

# Create a new release (tags and pushes)
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make release VERSION=v0.2.0"; \
		exit 1; \
	fi
	@echo "Creating release $(VERSION)..."
	@echo ""
	@echo "This will:"
	@echo "  1. Create a git tag: $(VERSION)"
	@echo "  2. Push to origin (triggers GitHub Actions release workflow)"
	@echo ""
	@read -p "Continue? (y/n): " confirm && [ "$$confirm" = "y" ] || exit 1
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo ""
	@echo "✓ Tag pushed. GitHub Actions will build and publish the release."
	@echo "  Monitor at: https://github.com/emmogba/myshipmate/actions"

# Format code
fmt:
	go fmt ./...
	@echo "✓ Code formatted"

# Show help
help:
	@echo "Shipmate CLI Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build        Build for current platform"
	@echo "  make build-all    Build for all platforms (Linux/macOS/Windows)"
	@echo "  make checksums    Build all + generate SHA256 checksums"
	@echo "  make test         Run tests with coverage"
	@echo "  make install      Install to /usr/local/bin"
	@echo "  make clean        Remove build artifacts"
	@echo "  make fmt          Format code"
	@echo "  make release      Create a new release (VERSION=vX.Y.Z)"
	@echo "  make help         Show this help"
