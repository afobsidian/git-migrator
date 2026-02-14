.PHONY: all build test clean install lint help

# Variables
BINARY_NAME=git-migrator
GO=go
GOFLAGS=-v

# Default target
all: test build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o bin/$(BINARY_NAME) ./cmd/git-migrator

## test: Run all tests
test: test-unit test-integration
	@echo "All tests passed!"

## test-unit: Run unit tests
test-unit:
	@echo "Running unit tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GO) test -v -tags=integration ./test/integration/...

## test-coverage: Generate coverage report
test-coverage: test-unit
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## test-regression: Run full regression suite
test-regression:
	@echo "Running regression tests..."
	$(GO) test -v ./test/regression/...

## test-requirements: Validate requirements coverage
test-requirements:
	@echo "Checking requirements coverage..."
	$(GO) run ./scripts/check-requirements.go

## lint: Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

## install: Install binary to GOPATH/bin
install:
	@echo "Installing..."
	$(GO) install ./cmd/git-migrator

## install-hooks: Install pre-commit hooks
install-hooks:
	@echo "Installing pre-commit hooks..."
	cp scripts/pre-commit .git/hooks/
	chmod +x .git/hooks/pre-commit

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run --rm -it $(BINARY_NAME):latest

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	gofmt -s -w .

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

## help: Show this help message
help:
	@echo "Git-Migrator - Makefile Commands"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

# Development targets (not shown in help)
watch:
	@echo "Watching for changes..."
	reflex -r '\.go$$' -s -- sh -c 'make test && make build'
