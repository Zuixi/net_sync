.PHONY: build run test clean install docker-build docker-run lint fmt help

# App name
APP_NAME := easy-sync
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Directories
BUILD_DIR := build
DIST_DIR := dist

# Targets
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	@mkdir -p $(DIST_DIR)

	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 ./cmd/server

	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-linux-arm64 ./cmd/server

	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd/server

	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 ./cmd/server

	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-darwin-arm64 ./cmd/server

run: ## Run the application
	$(GOCMD) run ./cmd/server

test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	@which golangci-lint > /dev/null 2>&1 || (echo "Please install golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

fmt: ## Format code
	$(GOFMT) ./...

vet: ## Run go vet
	$(GOVET) ./...

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html

install: build ## Install the application
	@echo "Installing $(APP_NAME)..."
	cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/

docker-build: ## Build Docker image
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

docker-run: ## Run Docker container
	docker run -p 3280:3280 -v $(PWD)/uploads:/app/uploads $(APP_NAME):latest

dev: ## Run in development mode with hot reload
	@which air > /dev/null 2>&1 || (echo "Please install air: go install github.com/air-verse/air@latest" && exit 1)
	air

install-dev-tools: ## Install development tools
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u github.com/air-verse/air

generate-docs: ## Generate documentation
	@echo "Generating documentation..."
	@which godoc > /dev/null 2>&1 || (echo "Please install godoc: go install golang.org/x/tools/cmd/godoc@latest" && exit 1)
	godoc -http=:6060

release: clean test lint build-all ## Create a release
	@echo "Creating release for version $(VERSION)..."
	@mkdir -p $(DIST_DIR)/release
	@cd $(DIST_DIR) && for binary in $(APP_NAME)-*; do \
		if [[ $$binary == *.exe ]]; then \
			zip release/$${binary%.exe}.zip $$binary; \
		else \
			tar -czf release/$$binary.tar.gz $$binary; \
		fi; \
	done
	@echo "Release files created in $(DIST_DIR)/release/"

# Default target
.DEFAULT_GOAL := help