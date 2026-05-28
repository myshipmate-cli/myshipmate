.PHONY: build test clean install release-all

BINARY=shipmate
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags="-s -w -X main.Version=$(VERSION)"
CLI_DIR=cli
WEB_DIR=web

# Default: build for current platform
build:
	cd $(CLI_DIR) && go build $(LDFLAGS) -o $(BINARY) .

# Run all tests
test:
	cd $(CLI_DIR) && go test -v -count=1 ./...

# Format and vet
check:
	cd $(CLI_DIR) && go fmt ./... && go vet ./...

# Build for all platforms
release-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	cd $(CLI_DIR) && \
		GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o ../dist/$(BINARY)_linux_amd64 . && \
		GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o ../dist/$(BINARY)_linux_arm64 . && \
		GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o ../dist/$(BINARY)_darwin_amd64 . && \
		GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o ../dist/$(BINARY)_darwin_arm64 . && \
		GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o ../dist/$(BINARY)_windows_amd64.exe . && \
		GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o ../dist/$(BINARY)_windows_arm64.exe .
	@echo ""
	@echo "Built binaries:"
	@ls -lh dist/

# Install locally
install: build
	cp $(CLI_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "Installed to /usr/local/bin/$(BINARY)"

# Clean build artifacts
clean:
	rm -f $(CLI_DIR)/$(BINARY)
	rm -rf dist/
	rm -rf $(WEB_DIR)/.next

# Frontend
web-build:
	cd $(WEB_DIR) && pnpm build

web-dev:
	cd $(WEB_DIR) && pnpm dev

# Deploy to Fly.io
deploy:
	cd $(WEB_DIR) && fly deploy --app myshipmate

# Create a new release tag
tag:
	@read -p "Version (e.g. v0.1.0): " VERSION; \
	git tag $$VERSION && git push origin $$VERSION && \
	echo "Pushed tag $$VERSION — GitHub Actions will build and release"

# Full check suite
verify: check test
	cd $(WEB_DIR) && pnpm lint && pnpm build
	@echo ""
	@echo "✓ All checks passed"
