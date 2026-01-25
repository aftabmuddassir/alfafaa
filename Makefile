.PHONY: run build test clean migrate seed help

# Default target
help:
	@echo "Alfafaa Blog Backend - Available Commands:"
	@echo ""
	@echo "  make run       - Run the server in development mode"
	@echo "  make build     - Build the binary"
	@echo "  make test      - Run all tests"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make migrate   - Run database migrations"
	@echo "  make seed      - Seed the database with initial data"
	@echo "  make deps      - Download dependencies"
	@echo "  make fmt       - Format code"
	@echo "  make lint      - Run linter (requires golangci-lint)"
	@echo ""

# Run the server
run:
	go run cmd/server/main.go

# Build the binary
build:
	go build -o bin/alfafaa-blog cmd/server/main.go

# Run tests
test:
	go test -v ./...

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
