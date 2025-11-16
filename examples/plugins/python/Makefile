.PHONY: build install install-system clean test

PLUGIN_NAME := python
PLUGIN_FILE := $(PLUGIN_NAME).so
USER_PLUGIN_DIR := $(HOME)/.uptool/plugins
SYSTEM_PLUGIN_DIR := /usr/local/lib/uptool/plugins

build:
	@echo "Building $(PLUGIN_NAME) plugin..."
	go build -buildmode=plugin -o $(PLUGIN_FILE) .
	@echo "✓ Built $(PLUGIN_FILE)"

install: build
	@echo "Installing to $(USER_PLUGIN_DIR)..."
	@mkdir -p $(USER_PLUGIN_DIR)
	@cp $(PLUGIN_FILE) $(USER_PLUGIN_DIR)/
	@echo "✓ Installed to $(USER_PLUGIN_DIR)/$(PLUGIN_FILE)"

install-system: build
	@echo "Installing to $(SYSTEM_PLUGIN_DIR) (requires sudo)..."
	@sudo mkdir -p $(SYSTEM_PLUGIN_DIR)
	@sudo cp $(PLUGIN_FILE) $(SYSTEM_PLUGIN_DIR)/
	@echo "✓ Installed to $(SYSTEM_PLUGIN_DIR)/$(PLUGIN_FILE)"

clean:
	@echo "Cleaning..."
	@rm -f $(PLUGIN_FILE)
	@echo "✓ Cleaned"

test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "✓ Tests passed"

test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover ./...

lint:
	@echo "Running linter..."
	@golangci-lint run || echo "Note: golangci-lint not installed"
