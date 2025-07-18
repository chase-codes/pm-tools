.PHONY: build clean test run install

# Build the AKS monitor dashboard
build:
	go build -o bin/aks-monitor cmd/aks-monitor/main.go

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Run tests
test:
	go test ./...

# Run the application
run: build
	./bin/aks-monitor

# Install dependencies
install:
	go mod download
	go mod tidy

# Build for different platforms
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/aks-monitor-linux cmd/aks-monitor/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/aks-monitor.exe cmd/aks-monitor/main.go

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o bin/aks-monitor-darwin cmd/aks-monitor/main.go

# Build all platforms
build-all: build-linux build-windows build-darwin

# Development mode with hot reload (requires air)
dev:
	air

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate documentation
docs:
	godoc -http=:6060 