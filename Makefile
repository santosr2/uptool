.PHONY: help build test lint fmt clean install run-scan run-plan run-update \
        version version-show version-bump-patch version-bump-minor version-bump-major \
        complexity

# Default target
help:
	@echo "uptool - Universal DevOps tool updater"
	@echo ""
	@echo "Available targets:"
	@echo "  build        Build the uptool binary"
	@echo "  install      Install uptool to \$$GOPATH/bin"
	@echo "  test         Run all tests"
	@echo "  lint         Run golangci-lint"
	@echo "  complexity   Check cyclomatic complexity with gocyclo"
	@echo "  fmt          Format code with gofmt"
	@echo "  clean        Remove build artifacts"
	@echo "  run-scan     Run uptool scan on this repository"
	@echo "  run-plan     Run uptool plan on this repository"
	@echo "  run-update   Run uptool update --dry-run on this repository"
	@echo ""
	@echo "Version Management:"
	@echo "  version-show         Show current version"
	@echo "  version-bump-patch   Bump patch version (0.1.0 -> 0.1.1)"
	@echo "  version-bump-minor   Bump minor version (0.1.0 -> 0.2.0)"
	@echo "  version-bump-major   Bump major version (0.1.0 -> 1.0.0)"

# Build the binary
build:
	@echo "Building uptool..."
	@mkdir -p dist
	go build -o dist/uptool ./cmd/uptool
	@echo "✓ Built: ./dist/uptool"

# Install to GOPATH/bin
install:
	@echo "Installing uptool..."
	go install ./cmd/uptool
	@echo "✓ Installed to \$$(go env GOPATH)/bin/uptool"

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "✓ Tests passed"

# Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

# Run linter
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "✓ Linting passed"; \
	else \
		echo "⚠ golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -w -s .
	@echo "✓ Code formatted"

# Vet code
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "✓ Vet passed"

# Check cyclomatic complexity
complexity:
	@echo "Checking cyclomatic complexity..."
	@if command -v gocyclo >/dev/null 2>&1; then \
		gocyclo -over 15 .; \
		echo "✓ Complexity check passed"; \
	else \
		echo "⚠ gocyclo not installed. Install with:"; \
		echo "  mise install  # or go install github.com/fzipp/gocyclo/cmd/gocyclo@latest"; \
	fi

# Run all checks
check: fmt vet complexity lint test
	@echo "✓ All checks passed"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f uptool
	rm -f coverage.out coverage.html
	go clean
	@echo "✓ Cleaned"

# Development commands
run-scan: build
	@echo "Running scan on current repository..."
	./dist/uptool scan

run-plan: build
	@echo "Running plan on current repository..."
	./dist/uptool plan

run-update: build
	@echo "Running update --dry-run on current repository..."
	./dist/uptool update --dry-run --diff

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build -o dist/uptool-linux-amd64 ./cmd/uptool
	GOOS=linux GOARCH=arm64 go build -o dist/uptool-linux-arm64 ./cmd/uptool
	GOOS=darwin GOARCH=amd64 go build -o dist/uptool-darwin-amd64 ./cmd/uptool
	GOOS=darwin GOARCH=arm64 go build -o dist/uptool-darwin-arm64 ./cmd/uptool
	GOOS=windows GOARCH=amd64 go build -o dist/uptool-windows-amd64.exe ./cmd/uptool
	@echo "✓ Built all platforms in dist/"

# Setup development environment
setup:
	@echo "Setting up development environment..."
	go mod download
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@echo "✓ Development environment ready"

# Version management targets
version-show:
	@echo "Current version:"
	@cat internal/version/VERSION

version-bump-patch:
	@echo "Bumping patch version..."
	@if ! command -v bump-my-version >/dev/null 2>&1; then \
		echo "⚠ bump-my-version not installed. Install with:"; \
		echo "  mise install  # or pip install bump-my-version"; \
		exit 1; \
	fi
	@bump-my-version bump patch --verbose
	@echo "✓ Version bumped. Don't forget to commit the changes!"

version-bump-minor:
	@echo "Bumping minor version..."
	@if ! command -v bump-my-version >/dev/null 2>&1; then \
		echo "⚠ bump-my-version not installed. Install with:"; \
		echo "  mise install  # or pip install bump-my-version"; \
		exit 1; \
	fi
	@bump-my-version bump minor --verbose
	@echo "✓ Version bumped. Don't forget to commit the changes!"

version-bump-major:
	@echo "Bumping major version..."
	@if ! command -v bump-my-version >/dev/null 2>&1; then \
		echo "⚠ bump-my-version not installed. Install with:"; \
		echo "  mise install  # or pip install bump-my-version"; \
		exit 1; \
	fi
	@bump-my-version bump major --verbose
	@echo "✓ Version bumped. Don't forget to commit the changes!"

# Alias for convenience
version: version-show
