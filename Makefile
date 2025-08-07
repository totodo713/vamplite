# Muscle Dreamer - ECS Framework Makefile
# High-performance game development with comprehensive tooling

.PHONY: all build build-release test test-all test-coverage lint format benchmark docs clean
.PHONY: build-web build-all docker-build docker-dev docker-setup deps
.PHONY: test-unit test-integration test-performance test-security ecs-test ecs-benchmark

# Build configuration
BINARY_NAME := muscle-dreamer
BUILD_DIR := dist
MAIN_PACKAGE := ./cmd/game
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# ECS Framework specific flags
ECS_TAGS := -tags "ecs,performance"
DEBUG_TAGS := -tags "ecs,debug"
TEST_TIMEOUT := 30m
COVERAGE_OUT := coverage.out

#==============================================================================
# Build Targets
#==============================================================================

all: clean lint test build ## Run all checks and build

help: ## Show this help message
	@echo "Muscle Dreamer ECS Framework - Development Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "🎯 Quick Start:"
	@echo "  make deps    # Install dependencies"
	@echo "  make test    # Run tests"
	@echo "  make build   # Build debug version"
	@echo "  make dev     # Start development server"

build: ## Build debug version
	@echo "🔨 Building debug version..."
	@mkdir -p $(BUILD_DIR)
	go build $(DEBUG_TAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "✅ Debug build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-release: ## Build optimized release version  
	@echo "🚀 Building release version..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) $(ECS_TAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "✅ Release build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-web: ## Build WebAssembly version
	@echo "🌐 Building WebAssembly version..."
	@mkdir -p $(BUILD_DIR)/web
	GOOS=js GOARCH=wasm go build $(LDFLAGS) -o $(BUILD_DIR)/web/game.wasm $(MAIN_PACKAGE)
	@cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(BUILD_DIR)/web/
	@echo "✅ WebAssembly build complete: $(BUILD_DIR)/web/"

build-all: ## Build for all platforms
	@echo "🏗️ Building for all platforms..."
	@mkdir -p $(BUILD_DIR)/{windows,linux,macos,web}
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) $(ECS_TAGS) -o $(BUILD_DIR)/windows/$(BINARY_NAME).exe $(MAIN_PACKAGE)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) $(ECS_TAGS) -o $(BUILD_DIR)/linux/$(BINARY_NAME) $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) $(ECS_TAGS) -o $(BUILD_DIR)/macos/$(BINARY_NAME) $(MAIN_PACKAGE)
	GOOS=js GOARCH=wasm go build $(LDFLAGS) -o $(BUILD_DIR)/web/game.wasm $(MAIN_PACKAGE)
	@cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(BUILD_DIR)/web/
	@echo "✅ All platform builds complete"

#==============================================================================
# Test Targets
#==============================================================================

test: test-unit ## Run unit tests

test-unit: ## Run unit tests only
	@echo "🧪 Running unit tests..."
	go test $(DEBUG_TAGS) -timeout $(TEST_TIMEOUT) -race ./internal/core/...

test-integration: ## Run integration tests
	@echo "🔗 Running integration tests..."
	go test $(DEBUG_TAGS) -timeout $(TEST_TIMEOUT) -race -tags integration ./internal/core/ecs/tests/

test-all: test-unit test-integration ## Run all tests
	@echo "✅ All tests complete"

test-coverage: ## Generate test coverage report
	@echo "📊 Generating test coverage..."
	go test $(DEBUG_TAGS) -timeout $(TEST_TIMEOUT) -race -coverprofile=$(COVERAGE_OUT) ./internal/core/... ./internal/mod/...
	go tool cover -html=$(COVERAGE_OUT) -o coverage.html
	go tool cover -func=$(COVERAGE_OUT)
	@echo "📈 Coverage report: coverage.html"

#==============================================================================
# ECS Framework Specific Targets  
#==============================================================================

ecs-test: ## Run ECS framework specific tests
	@echo "⚙️ Running ECS framework tests..."
	go test $(DEBUG_TAGS) -timeout $(TEST_TIMEOUT) -race -v ./internal/core/ecs/...

ecs-benchmark: ## Run ECS performance benchmarks
	@echo "⚡ Running ECS benchmarks..."
	go test $(ECS_TAGS) -bench=. -benchmem -benchtime=5s ./internal/core/ecs/tests/
	@echo "Target: 10,000 entities @ 60FPS, <1ms queries, <100B/entity"

#==============================================================================
# Code Quality Targets
#==============================================================================

lint: ## Run code linting
	@echo "🔍 Running linter..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "⚠️ golangci-lint not installed. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run ./...
	@echo "✅ Linting complete"

format: ## Format code
	@echo "🎨 Formatting code..."
	go fmt $$(go list ./... | grep -v /docs/)
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w $$(find . -name "*.go" -not -path "./docs/*" -not -path "./vendor/*" -not -path "./.git/*"); \
	else \
		echo "⚠️ goimports not installed. Installing..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		goimports -w $$(find . -name "*.go" -not -path "./docs/*" -not -path "./vendor/*" -not -path "./.git/*"); \
	fi
	@echo "✅ Code formatting complete"

#==============================================================================
# Docker Targets
#==============================================================================

docker-setup: ## Initialize Docker development environment
	@echo "🐳 Setting up Docker environment..."
	docker compose build
	docker compose run --rm dev go mod tidy
	@echo "✅ Docker environment ready"

docker-dev: ## Start development containers
	@echo "🐳 Starting development containers..."
	docker compose up -d dev web-dev
	@echo "✅ Development containers running"
	@echo "🌐 Game dev server: http://localhost:8080"
	@echo "🌐 Web dev server: http://localhost:3000"

docker-build: ## Cross-compile using Docker
	@echo "🐳 Cross-compiling with Docker..."
	docker run --rm -v "$(PWD)":/usr/src/app -w /usr/src/app golang:1.24 make build-all
	@echo "✅ Docker cross-compilation complete"

#==============================================================================
# Development Targets
#==============================================================================

dev: ## Run local development server
	@echo "🚀 Starting development server..."
	go run $(DEBUG_TAGS) $(MAIN_PACKAGE)

deps: ## Update Go module dependencies
	@echo "📦 Updating dependencies..."
	go mod tidy
	go mod download
	@echo "✅ Dependencies updated"

clean: ## Remove build artifacts and clean Docker
	@echo "🧹 Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_OUT) coverage.html
	go clean -cache -testcache
	@if command -v docker >/dev/null 2>&1; then \
		docker compose down -v >/dev/null 2>&1 || true; \
		docker system prune -f >/dev/null 2>&1 || true; \
	fi
	@echo "✅ Clean complete"
