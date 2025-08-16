# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# Guidelines

This document defines the project's rules, objectives, and progress management methods. Please proceed with the project according to the following content.

## Top-Level Rules

- To maximize efficiency, **if you need to execute multiple independent processes, invoke those tools concurrently, not sequentially**.
- **You must think exclusively in English**. However, you are required to **respond in Japanese**.
- To understand how to use a library, **always use the Contex7 MCP** to retrieve the latest information. 
- ALWAYS use ONLY Environments for ANY and ALL file, code, or shell operations—NO EXCEPTIONS—even for simple or generic requests.
- DO NOT install or use the git cli with the environment_run_cmd tool. All environment tools will handle git operations for you. Changing ".git" yourself will compromise the integrity of your environment.
- You MUST inform the user how to view your work using `container-use log <env_id>` AND `container-use checkout <env_id>`. Failure to do this will make your work inaccessible to others.

## Programming Rules

- Avoid hard-coding values unless absolutely necessary.
- Do not use `any` or `unknown` types in TypeScript.
- You must not use a TypeScript `class` unless it is absolutely necessary (e.g., extending the `Error` class for custom error handling that requires `instanceof` checks).

## Project Overview

**Muscle Dreamer** is a 2D survival action roguelike game built with Go and Ebitengine. It features a modular theme system, secure mod support, and runs completely offline. The game uses Entity Component System (ECS) architecture and supports cross-platform deployment including WebAssembly.

## Development Commands

### Core Development
- `make dev` - Run local development server
- `make build` - Build debug version (outputs to `dist/muscle-dreamer`)
- `make build-release` - Build optimized release version
- `make test` - Run unit tests (`go test ./...`)
- `make test-all` - Run all tests including integration tests
- `make lint` - Run code analysis (`golangci-lint run`)
- `make format` - Format code (`go fmt ./...` and `goimports -w .`)

### Platform-specific Builds
- `make build-web` - Build WebAssembly version (outputs to `dist/web/`)
- `make build-all` - Build for all platforms (Windows, Linux, macOS, WebAssembly)
- `make docker-build` - Cross-compile using Docker containers

### Docker Development
- `make docker-setup` - Initialize Docker development environment
- `make docker-dev` - Start development containers (game dev + web dev)
  - Main development: http://localhost:8080
  - Web development: http://localhost:3000

### Web Development
- `cd web && npm run dev` - Start web development server
- `cd web && npm run build` - Build web assets
- `cd web && npm start` - Serve production web build

### Maintenance
- `make clean` - Remove build artifacts and clean Docker
- `make deps` - Update Go module dependencies

## Architecture Overview

### Core Components
- **ECS Framework** (`internal/core/`) - Entity Component System for game objects
- **Theme System** (`internal/theme/`) - Dynamic content loading and switching
- **Mod System** (`internal/mod/`) - Secure plugin architecture with sandboxing
- **Platform Layer** (`internal/platform/`) - Cross-platform abstraction
- **UI System** (`internal/ui/`) - User interface management

### Key Directories
- `cmd/game/main.go` - Main application entry point
- `internal/core/game.go` - Core game engine and main game loop
- `themes/` - Theme content (assets, configurations, scripts)
- `mods/` - Mod content organized in `enabled/`, `disabled/`, `staging/`
- `assets/` - Core game assets (sprites, audio, fonts, UI)
- `config/` - Game configuration files (YAML)
- `saves/` - Local save data and user settings
- `web/` - WebAssembly deployment files and web development

### Technology Stack
- **Language**: Go 1.22
- **Game Engine**: Ebitengine v2.6.3
- **Build System**: Make + Docker
- **Web Framework**: Node.js with Express (for web serving)
- **Asset Formats**: PNG/JPG (images), OGG/WAV (audio), YAML (configuration)

## Development Guidelines

### Code Organization
- Follow standard Go project layout with `internal/` for private packages
- Use Entity Component System patterns for game logic
- Implement interfaces for platform-specific functionality
- Keep game logic separate from rendering and input handling

### Asset Management
- Place core assets in `assets/` directory organized by type
- Theme-specific assets go in `themes/[theme-name]/assets/`
- Optimize images and audio files before committing
- Use YAML for all configuration files

### Theme Development
- Each theme is a self-contained directory in `themes/`
- Must include `theme.yaml` configuration file
- Follow the structure: `assets/`, `scripts/`, `localization/`, `metadata/`
- Test themes with the built-in validator

### Mod Development
- Mods are executed in secure sandboxes with limited file system access
- Use the provided API for game interactions
- Place mods in `mods/staging/` for development, `mods/enabled/` for active mods
- All mod scripts are subject to security validation

### Testing
- Write unit tests for all core systems
- Use integration tests for cross-component functionality
- Test on multiple platforms before release
- Include performance benchmarks for critical paths

### Security Considerations
- All mod code runs in sandboxed environments
- File system access is restricted to designated directories
- Network access is disabled for mods
- Validate all user-provided content before loading

## Platform-Specific Notes

### WebAssembly
- WebAssembly builds are optimized for size and loading speed
- Some features may be limited compared to native builds
- Use the web development server for testing browser compatibility
- Assets are served through the Express.js server in `web/`

### Desktop Platforms
- Native builds support full feature set including advanced audio and graphics
- Cross-compilation is handled through Docker containers
- Platform-specific optimizations are applied during build

## File Naming and Organization

### Configuration Files
- `config/game.yaml` - Main game configuration
- `themes/[name]/theme.yaml` - Theme definitions
- `mods/[name]/mod.yaml` - Mod metadata

### Asset Organization
- Sprites: `assets/sprites/` or `themes/[name]/assets/sprites/`
- Audio: `assets/audio/` or `themes/[name]/assets/audio/`
- Fonts: `assets/fonts/`
- UI elements: `assets/ui/`

## Common Development Tasks

### Adding New Game Features
1. Design components and systems following ECS patterns
2. Add interfaces in `internal/core/` if cross-platform support needed
3. Implement platform-specific code in `internal/platform/`
4. Add configuration options to appropriate YAML files
5. Write tests and update documentation

### Creating New Themes
1. Create directory structure in `themes/[theme-name]/`
2. Design theme.yaml configuration
3. Create assets following naming conventions
4. Test with theme validator
5. Package for distribution

### Debugging Issues
- Use `make dev` for development builds with debug symbols
- Check console output for detailed error messages
- Use browser developer tools for WebAssembly debugging
- Enable verbose logging in development mode

## Important Notes

- All development should maintain offline-first functionality
- Security is paramount, especially for mod system
- Performance targets: 60 FPS, <256MB memory usage, <3s startup time
- The game is designed to run completely without internet connectivity
- Maintain backwards compatibility for save files and theme formats

## Development Process Rules (XP + GitHub Flow)

### Branch Strategy
- **Feature branches**: Use `developer/*` naming convention
  - Example: `developer/ecs-implementation`
  - Example: `developer/ui-system`
- **Main branch**: Protected, always deployable
- **No direct commits to main**: Always use PRs

### Commit Strategy
1. **Frequent Small Commits**: Commit every meaningful change
2. **TDD Cycle Commits**: Separate commits for Red-Green-Refactor
3. **30-Minute Rule**: Commit at least every 30 minutes
4. **Boy Scout Rule**: Leave code cleaner than you found it

### Quality Assurance Before Commit
```bash
# Run before EVERY commit - NO EXCEPTIONS
make format  # Auto-format code
make lint    # Check for issues
make test    # Run tests

# Fix ALL warnings and errors before committing
```

### GitHub Flow Implementation
1. Create feature branch: `git checkout -b developer/feature-name`
2. Keep commits atomic and buildable
3. Regularly sync with main: `git pull origin main`
4. Push frequently: `git push origin developer/feature-name`
5. Create PR when feature is complete
6. Merge after review and CI passes

### Commit Message Format (Japanese)
```
<type>: <概要>

<なぜこの変更が必要か>

Refs: #issue-number
```

**Types:**
- `feat`: 機能追加
- `fix`: バグ修正
- `test`: テスト追加・修正
- `refactor`: リファクタリング
- `style`: フォーマット修正
- `docs`: ドキュメント更新
- `chore`: ビルド・ツール関連
- `perf`: パフォーマンス改善

**Examples:**
- `feat: EntityManagerの基本実装を追加`
- `test: EntityManagerのユニットテストを作成`
- `fix: メモリリークを修正`
- `refactor: コンポーネントストアの構造を改善`
- `style: go fmtでコードを整形`

### Task Implementation Workflow
1. Create/checkout feature branch: `developer/*`
2. Pull latest main changes
3. Implement feature/fix with TDD
4. Run quality checks (format, lint, test)
5. Fix any issues found
6. Commit with descriptive Japanese message
7. Push to feature branch
8. Create PR when ready

### Continuous Integration Rules
- Every commit MUST compile
- Every commit MUST pass lint
- Every commit MUST pass tests
- No broken commits in history

### Boy Scout Rule Examples
- See unused import? Remove it and commit
- Find unformatted code? Format it and commit
- Notice missing test? Add it and commit
- Spot unclear variable name? Refactor and commit

### PR Guidelines
- **Title**: Clear description of changes (Japanese OK)
- **Description**: Why the change is needed, what it does
- **Checklist**:
  - Tests pass (`make test`)
  - Lint passes (`make lint`)
  - Code formatted (`make format`)
  - Documentation updated if needed
