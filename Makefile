.PHONY: all build run clean test lint dev

# Variables
BINARY_NAME=omniroute
MAIN_PATH=./cmd/server
BUILD_DIR=./bin

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

# Build the application
build:
	@echo "Building..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Run the application
run:
	@echo "Running..."
	@go run $(MAIN_PATH)/main.go

# Run with hot reload (requires air)
dev:
	@echo "Running in development mode..."
	@air

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf ./storage/*.db
	@go clean

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Generate swagger docs (if using swaggo)
swagger:
	@echo "Generating swagger docs..."
	@swag init -g cmd/server/main.go -o docs

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Docker build
docker-build:
	@echo "Building Docker image..."
	@docker build -t omniroute-go:latest .

# Docker run
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 -v $(pwd)/storage:/app/storage omniroute-go:latest

# Initialize project (download deps, create dirs)
init:
	@echo "Initializing project..."
	@mkdir -p $(BUILD_DIR) storage
	@go mod download
	@go mod tidy
	@echo "Project initialized!"

# Help
help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make dev          - Run with hot reload (requires air)"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make test         - Run tests"
	@echo "  make test-coverage- Run tests with coverage"
	@echo "  make lint         - Run linter"
	@echo "  make deps         - Download dependencies"
	@echo "  make build-all    - Build for all platforms"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"
	@echo "  make init         - Initialize project"
