# Muscle Dreamer

2D Survival Action Roguelike Game built with Go and Ebitengine.

## Features

- Entity Component System (ECS) architecture
- Cross-platform support (Windows, macOS, Linux, WebAssembly)
- Theme system for customization
- Secure MOD system with sandbox
- Offline-first design

## Development

### Requirements
- Go 1.22+
- Ebitengine v2.6.3

### Development Setup
```bash
# 1. First-time setup: Configure Git hooks and commit template
./scripts/setup-git-hooks.sh

# 2. Create feature branch using new alias
git new-feature my-awesome-feature  # Creates developer/my-awesome-feature

# 3. Start development
make dev        # Run local development
make test       # Run unit tests
make build      # Build debug version
```

### Development Process
We follow **Extreme Programming (XP) + GitHub Flow**:

1. **Work in feature branches**: Use `developer/*` naming convention
2. **Commit frequently**: Small, atomic commits every 30 minutes max
3. **Quality first**: Pre-commit hooks run `make format && make lint && make test`
4. **TDD cycle**: Red → Green → Refactor with separate commits
5. **Boy Scout Rule**: Leave code cleaner than you found it

### Commit Message Format
```
<type>: <概要>

<なぜこの変更が必要か>

Refs: #issue-number
```

**Examples:**
- `feat: EntityManagerの基本実装を追加`
- `test: EntityManagerのユニットテストを作成`
- `fix: メモリリークを修正`

See [CLAUDE.md](CLAUDE.md) for complete development guidelines.

## License

**Creative Commons Attribution-NonCommercial 4.0 International (CC BY-NC 4.0)**

This project is licensed under CC BY-NC 4.0, which means:

✅ **You can:**
- Share — copy and redistribute the material
- Adapt — remix, transform, and build upon the material

📋 **Under these conditions:**
- **Attribution** — You must give appropriate credit and indicate if changes were made
- **NonCommercial** — You may not use the material for commercial purposes

For the full license text, see [LICENSE](LICENSE) or visit https://creativecommons.org/licenses/by-nc/4.0/

## Third-Party Licenses

This project uses:
- [Ebitengine](https://github.com/hajimehoshi/ebiten) - Apache License 2.0

## Contributing

Contributions are welcome! By contributing, you agree that your contributions will be licensed under the same CC BY-NC 4.0 terms.

---

© 2025 totodo713. All rights reserved.
