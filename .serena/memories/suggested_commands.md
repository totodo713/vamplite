# Essential Development Commands

## Core Development
- `make dev` - Start development server with debug tags
- `make build` - Build debug version (outputs to `dist/muscle-dreamer`)
- `make build-release` - Build optimized release version with performance tags

## Testing
- `make test` or `make test-unit` - Run unit tests with race detection
- `make test-integration` - Run integration tests
- `make test-all` - Run all tests (unit + integration)
- `make test-coverage` - Generate coverage report (creates coverage.html)
- `make ecs-test` - Run ECS framework specific tests
- `make ecs-benchmark` - Run ECS performance benchmarks

## Code Quality (MUST RUN AFTER CODE CHANGES)
- `make lint` - Run golangci-lint (installs if missing)
- `make format` - Run go fmt and goimports (installs goimports if missing)

## Platform Builds
- `make build-web` - Build WebAssembly version
- `make build-all` - Build for all platforms (Windows, Linux, macOS, WebAssembly)

## Docker Development
- `make docker-setup` - Initialize Docker environment
- `make docker-dev` - Start dev containers (game:8080, web:3000)
- `make docker-build` - Cross-compile using Docker

## Maintenance
- `make clean` - Remove build artifacts, clean caches and Docker
- `make deps` - Update Go module dependencies

## Direct Go Commands
- `go run -tags "ecs,debug" ./cmd/game` - Run with debug tags
- `go test -tags "ecs,debug" -race ./internal/core/...` - Run tests with race detection
- `go test -bench=. -benchmem ./internal/core/ecs/tests/` - Run benchmarks

## Key Build Tags
- `ecs,performance` - Production ECS builds
- `ecs,debug` - Development ECS builds with debug info