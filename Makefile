.PHONY: build run test clean migrate seed docker help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
API_BINARY=jptaku-api
WS_BINARY=jptaku-ws
MIGRATE_BINARY=jptaku-migrate

# Build directories
BUILD_DIR=./build
CMD_DIR=./cmd

# Default target
all: build

## build: Build all binaries
build: build-api build-ws build-migrate

## build-api: Build API server
build-api:
	@echo "Building API server..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(API_BINARY) $(CMD_DIR)/api/main.go

## build-ws: Build WebSocket server
build-ws:
	@echo "Building WebSocket server..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(WS_BINARY) $(CMD_DIR)/ws/main.go 2>/dev/null || echo "WebSocket server not implemented yet"

## build-migrate: Build migration tool
build-migrate:
	@echo "Building migration tool..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(MIGRATE_BINARY) $(CMD_DIR)/migrate/main.go

## run: Run API server
run:
	@echo "Starting API server..."
	$(GOCMD) run $(CMD_DIR)/api/main.go

## run-dev: Run API server with hot reload (requires air)
run-dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air not installed. Run: go install github.com/cosmtrek/air@latest"; \
		$(GOCMD) run $(CMD_DIR)/api/main.go; \
	fi

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download

## tidy: Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

## migrate-up: Run database migrations
migrate-up: build-migrate
	@echo "Running migrations..."
	$(BUILD_DIR)/$(MIGRATE_BINARY) -action=up

## migrate-down: Rollback database migrations
migrate-down: build-migrate
	@echo "Rolling back migrations..."
	$(BUILD_DIR)/$(MIGRATE_BINARY) -action=down

## seed: Seed database with sample data
seed: build-migrate
	@echo "Seeding database..."
	$(BUILD_DIR)/$(MIGRATE_BINARY) -action=seed

## lint: Run linter
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t jptaku-api:latest .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 30001:30001 --env-file .env jptaku-api:latest

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

