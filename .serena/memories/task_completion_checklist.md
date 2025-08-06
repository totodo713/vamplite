# Task Completion Checklist

## MANDATORY Steps After Code Changes

### 1. Code Quality (REQUIRED)
```bash
make format  # Run go fmt and goimports
make lint    # Run golangci-lint
```
**These MUST pass before considering task complete**

### 2. Testing (REQUIRED)
```bash
make test           # Run unit tests
make test-integration  # Run integration tests (if applicable)
```
For ECS-specific changes:
```bash
make ecs-test       # Run ECS framework tests
make ecs-benchmark  # Verify performance targets
```

### 3. Build Verification (REQUIRED)
```bash
make build          # Verify debug build works
```
For release candidates:
```bash
make build-release  # Verify release build works
```

### 4. Performance Validation (For ECS changes)
Run benchmarks and verify targets:
- 10,000 entities @ 60FPS
- <1ms queries
- <100B/entity memory usage

### 5. Integration Testing (When applicable)
```bash
make test-all       # Run complete test suite
```

## Additional Checks

### For New Features
- Add appropriate unit tests
- Update integration tests if cross-component functionality
- Consider performance impact on benchmarks
- Verify cross-platform compatibility if needed

### For ECS Framework Changes
- Run `make ecs-benchmark` to ensure performance targets
- Test with different entity counts
- Verify memory usage patterns
- Check component query performance

### For Platform-Specific Code
- Test on target platforms
- Verify WebAssembly builds with `make build-web`
- Run full platform build suite with `make build-all`

## Error Handling
- If any step fails, DO NOT consider task complete
- Address all linting issues
- Ensure all tests pass
- Fix any build errors

## Documentation Updates (When Required)
- Update relevant .md files in docs/
- Maintain CLAUDE.md if development workflow changes
- Update API documentation for public interfaces