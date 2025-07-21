.PHONY: build clean test run install build-all

# Build all tools
build-all:
	@echo "Building all tools..."
	@cd issue-monitor && make build
	@echo "✅ All tools built successfully"

# Build specific tool
build-issue-monitor:
	@echo "Building issue-monitor..."
	@cd issue-monitor && make build

# Clean all build artifacts
clean:
	@echo "Cleaning all tools..."
	@cd issue-monitor && make clean
	@echo "✅ Clean complete"

# Run tests for all tools
test:
	@echo "Running tests for all tools..."
	@cd issue-monitor && make test
	@echo "✅ All tests passed"

# Install dependencies for all tools
install:
	@echo "Installing dependencies for all tools..."
	@cd issue-monitor && make install
	@echo "✅ Dependencies installed"

# Format code for all tools
fmt:
	@echo "Formatting code for all tools..."
	@cd issue-monitor && make fmt
	@echo "✅ Code formatted"

# Lint code for all tools
lint:
	@echo "Linting code for all tools..."
	@cd issue-monitor && make lint
	@echo "✅ Code linted"

# Development mode (requires air)
dev-issue-monitor:
	@cd issue-monitor && make dev

# Generate documentation
docs:
	@echo "Generating documentation..."
	@cd issue-monitor && make docs
	@echo "✅ Documentation generated"

# Help
help:
	@echo "Available commands:"
	@echo "  build-all          - Build all tools"
	@echo "  build-issue-monitor - Build issue-monitor tool"
	@echo "  clean              - Clean all build artifacts"
	@echo "  test               - Run tests for all tools"
	@echo "  install            - Install dependencies for all tools"
	@echo "  fmt                - Format code for all tools"
	@echo "  lint               - Lint code for all tools"
	@echo "  dev-issue-monitor  - Run issue-monitor in development mode"
	@echo "  docs               - Generate documentation"
	@echo "  help               - Show this help message" 