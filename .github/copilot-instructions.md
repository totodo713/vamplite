# GitHub Copilot Custom Instructions (English Translation)

# Guidelines

This document defines the project's rules, objectives, and progress management methods. Please proceed with the project according to the following content.

## Top-Level Rules

- To maximize efficiency, **if you need to execute multiple independent processes, invoke those tools concurrently, not sequentially**.
- **You must think exclusively in English**. However, you are required to **respond in Japanese**.
- To understand how to use a library, **always use the Contex7 MCP** to retrieve the latest information.

## Programming Rules

- Avoid hard-coding values unless absolutely necessary.
- Do not use `any` or `unknown` types in TypeScript.
- You must not use a TypeScript `class` unless it is absolutely necessary (e.g., extending the `Error` class for custom error handling that requires `instanceof` checks).


## Project Overview
- **Project Name**: Muscle Dreamer
- **Type**: 2D Survival Action Roguelike Game
- **Language**: Go 1.22
- **Game Engine**: Ebitengine v2.6.3
- **Architecture**: Entity Component System (ECS)
- **Development Environment**: GoLand IDE
- **Team Development**: Yes

## Coding Standards and Style

### Go Language Standards
- Follow Go de facto standards
- Use `go fmt` and `goimports` for formatting
- Pass code analysis with `golangci-lint`
- Use standard Go project layout (with `internal/` packages)

### Naming Conventions
- Package names: lowercase, short, descriptive
- Function/Method names: camelCase, exported start with uppercase
- Variable names: camelCase, length according to scope
- Constant names: camelCase or UPPER_SNAKE_CASE

### Comments
- Always add comments to exported functions/types
- Include package-level doc comments
- Add explanatory comments for complex logic

## Architecture Design Principles

### ECS (Entity Component System)
- Clearly separate responsibilities of entities, components, and systems
- Components contain only data, logic is implemented in systems
- Minimize dependencies between systems

### Directory Structure
```
cmd/game/          # Application entry point
internal/core/     # Core ECS framework
internal/theme/    # Theme system
internal/mod/      # MOD system (secure sandbox)
internal/platform/ # Cross-platform abstraction
internal/ui/       # UI system
themes/            # Theme content
mods/              # MOD content
assets/            # Core game assets
config/            # YAML config files
saves/             # Save data
web/               # WebAssembly deployment
```

### Layer Separation
- Separate game logic from rendering
- Abstract platform-specific features with interfaces
- Separate business logic from external dependencies

## Ebitengine Specific Considerations

### Basic Patterns
- Implement `Update()` and `Draw()` methods in the `Game` struct
- Start the game loop with `ebiten.RunGame()`
- Efficient processing to maintain 60 FPS

### Resource Management
- Manage image resources with `ebiten.NewImageFromImage()`
- Control audio playback with `audio.NewPlayer()`
- Load assets during initialization, avoid loading in the game loop

### Input Handling
- Use the `inpututil` package for input state management
- Support keyboard, mouse, and gamepad
- Track input state per frame

## Testing Strategy

### Unit Testing
- Create tests for all core systems
- Test files use `*_test.go` naming
- Use the `testing` package
- Use mocks to isolate external dependencies

### Integration Testing
- Test interactions between systems
- Verify game loop operation
- Test platform-specific features

### Test Coverage
- Critical path: 90%+
- General features: 80%+
- Include performance tests

## Security Considerations

### MOD System
- All MOD code runs in a sandbox environment
- Restrict file system access to limited directories
- Disable network access
- Require validation of user-provided content

### Data Validation
- Validate when loading config files
- Check integrity of save data
- Ensure safety of external assets

## Performance Requirements

### Targets
- Frame rate: maintain 60 FPS
- Memory usage: under 256MB
- Startup time: under 3 seconds
- Cross-platform support (including WebAssembly)

### Optimization Guidelines
- Reduce garbage collection load
- Optimize rendering
- Use memory pools
- Actively use profiling tools

## Configuration Management

### YAML Usage
- Use YAML format for all config files
- Manage structured, hierarchical settings
- Support environment-specific overrides

### Config File Examples
- `config/game.yaml`: Main game settings
- `themes/[name]/theme.yaml`: Theme definition
- `mods/[name]/mod.yaml`: MOD metadata

## Development Workflow

### Daily Commands
- `make dev`: Run local development
- `make test`: Run unit tests
- `make lint`: Code analysis
- `make format`: Code formatting

### Build & Deploy
- `make build`: Debug build
- `make build-release`: Release build
- `make build-all`: Build for all platforms
- `make build-web`: Build for WebAssembly

## Error Handling

### Go Language Patterns
- Explicit error handling (`if err != nil`)
- Use custom error types
- Wrap errors with context information
- Recover from panics and log appropriately

### Logging
- Use structured logging
- Set appropriate log levels
- Output detailed debug info (in development mode)

## Important Constraints

### Offline First
- Fully functional without internet connection
- Run game with only local resources
- Save data locally

### Backward Compatibility
- Maintain save file format compatibility
- Ensure theme format stability
- Non-destructive API changes

---

## Instructions for Copilot

Based on the above information, please prioritize the following when generating code:

1. Always use idiomatic Go
2. Design compatible with ECS architecture
3. Prioritize readability and include appropriate comments
4. Propose test code simultaneously
5. Implement proper error handling
6. Consider performance in implementation
7. Adhere to security requirements (especially MOD system)
8. Abstract for cross-platform support

When proposing code, briefly explain why you chose that implementation.
