.PHONY: build build-all test clean install help

# Variables
BINARY_NAME=pincho
VERSION?=dev
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X 'gitlab.com/pincho-app/pincho-cli/cmd.version=$(VERSION)' -X 'gitlab.com/pincho-app/pincho-cli/cmd.commit=$(COMMIT)' -X 'gitlab.com/pincho-app/pincho-cli/cmd.date=$(DATE)'"

# Build for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) main.go
	@echo "Done! Binary: ./$(BINARY_NAME)"

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 main.go
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe main.go
	@echo "Done! Binaries in ./dist/"

# Run tests
test:
	@echo "Running tests..."
	go test ./... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -f coverage.out coverage.html
	@echo "Done!"

# Install to /usr/local/bin
install: build
	@echo "Installing to /usr/local/bin..."
	sudo mv $(BINARY_NAME) /usr/local/bin/
	@echo "Done! Run 'pincho --help' to get started"

# Show help
help:
	@echo "Pincho CLI - Makefile targets:"
	@echo ""
	@echo "  make build         Build binary for current platform"
	@echo "  make build-all     Build binaries for all platforms (dist/)"
	@echo "  make test          Run tests"
	@echo "  make test-coverage Run tests with coverage report"
	@echo "  make install       Build and install to /usr/local/bin"
	@echo "  make clean         Remove build artifacts"
	@echo "  make help          Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  COMMIT=$(COMMIT)"
	@echo "  DATE=$(DATE)"
