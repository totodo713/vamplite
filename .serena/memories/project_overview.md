# Muscle Dreamer Project Overview

## Project Purpose
Muscle Dreamer is a 2D survival action roguelike game built with Go and Ebitengine. The project features:
- Modular theme system for dynamic content
- Secure mod support with sandboxing
- Complete offline functionality
- Entity Component System (ECS) architecture
- Cross-platform deployment including WebAssembly

## Current Focus
The project is currently developing an ECS (Entity Component System) framework as the core game engine architecture. This is being implemented in the `internal/core/ecs/` directory with comprehensive test coverage and performance optimization targets.

## Technology Stack
- **Language**: Go 1.24
- **Game Engine**: Ebitengine v2.6.3
- **Testing**: testify v1.10.0
- **Build System**: Make + Docker
- **Deployment**: Native builds + WebAssembly
- **Configuration**: YAML files

## Performance Goals
- 10,000+ entities @ 60FPS
- <1ms queries
- <100B/entity memory usage
- <256MB total memory usage
- <3s startup time

## Architecture
- `cmd/game/main.go` - Main application entry point
- `internal/core/` - ECS framework and game engine
- `internal/core/ecs/` - ECS implementation with components, systems, storage
- `internal/mod/` - Secure mod system
- `internal/theme/` - Theme management
- `internal/platform/` - Cross-platform abstraction
- `internal/ui/` - User interface system
- `themes/` - Theme content
- `mods/` - Mod content (enabled/disabled/staging)
- `assets/` - Core game assets
- `config/` - Game configuration
- `web/` - WebAssembly deployment files