.PHONY: run build test test-unit test-coverage test-short clean migrate seed help

# Default target
help:
	@echo "Alfafaa Blog Backend - Available Commands:"
	@echo ""
	@echo "  make run           - Run the server in development mode"
	@echo "  make build         - Build the binary"
	@echo "  make test          - Run all tests"
	@echo "  make test-unit     - Run unit tests only (utils and services)"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make test-short    - Run short tests only"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make migrate       - Run database migrations"
	@echo "  make seed          - Seed the database with initial data"
	@echo "  make deps          - Download dependencies"
	@echo "  make fmt           - Format code"
	@echo "  make lint          - Run linter (requires golangci-lint)"
	@echo ""

# Run the server
run:
	go run cmd/server/main.go

# Build the binary
build:
	go build -o bin/alfafaa-blog cmd/server/main.go

# Run all tests
test:
	@echo "Running all tests..."
	go test ./... -v

# Run unit tests only (utils and services)
test-unit:
	@echo "Running unit tests..."
	go test ./internal/utils/... ./internal/services/... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	go tool cover -func=coverage.out | grep total

# Run short tests only
test-short:
	@echo "Running short tests..."
	go test ./... -short -v

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Run migrations
migrate:
	go run cmd/server/main.go -migrate

# Seed the database
seed:
	go run cmd/server/main.go -seed

# Download dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run ./...
