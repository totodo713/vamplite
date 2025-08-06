# Code Style and Conventions

## Go Code Standards
Based on .golangci.yml configuration:

### Complexity Limits
- Maximum cyclomatic complexity: 15
- Duplicate code threshold: 100 lines
- Minimum constant length: 2 characters, 2 occurrences

### Import Organization
- Local packages prefixed with `muscle-dreamer`
- Use gci for import grouping
- Run `goimports` for automatic import management

### Code Quality Rules
- Enabled linter tags: diagnostic, experimental, opinionated, performance, style
- Error handling: Use errorlint for proper error formatting
- Exhaustive checking for switch statements
- gocritic for advanced static analysis

### Naming Conventions
- Follow standard Go naming conventions
- Package names should be lowercase, single words
- Use descriptive variable names
- Constants in UPPER_CASE for exported values

## ECS-Specific Conventions
- Entity IDs use custom EntityID type for type safety
- Components implement Component interface
- Systems implement System interface
- Use build tags: `ecs,performance` for release, `ecs,debug` for development

## Testing Standards
- Use testify for assertions
- Race detection enabled by default
- 30-minute test timeout for long-running tests
- Separate integration tests with `integration` build tag
- Benchmark tests target: 10,000 entities @ 60FPS

## Project Structure
- `internal/` for private packages
- Clear separation between ECS framework (`internal/core/ecs/`) and game logic
- Components in `internal/core/ecs/components/`
- Systems in `internal/core/systems/`
- Storage implementations in `internal/core/ecs/storage/`

## Performance Considerations
- Memory pool usage for efficient allocation
- Sparse set data structures for component storage
- Cache-friendly data layouts
- Avoid allocations in hot paths