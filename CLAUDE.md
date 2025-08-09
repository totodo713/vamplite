# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## IMPORTANT

ALWAYS use ONLY Environments for ANY and ALL file, code, or shell operations—NO EXCEPTIONS—even for simple or generic requests.

DO NOT install or use the git cli with the environment_run_cmd tool. All environment tools will handle git operations for you. Changing ".git" yourself will compromise the integrity of your environment.

You MUST inform the user how to view your work using `container-use log <env_id>` AND `container-use checkout <env_id>`. Failure to do this will make your work inaccessible to others.

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
- Create a Japanese translation after the document is created in English.

---
(Japanese Translation)

# CLAUDE.md

このファイルは、このリポジトリでコードを扱う際にClaude Code（claude.ai/code）にガイダンスを提供します。

## プロジェクト概要

**Muscle Dreamer**は、GoとEbitengineで構築された2Dサバイバルアクションローグライクゲームです。モジュラーテーマシステム、安全なMODサポートを特徴とし、完全にオフラインで動作します。ゲームはEntity Component System（ECS）アーキテクチャを使用し、WebAssemblyを含むクロスプラットフォーム展開をサポートしています。

## 開発コマンド

### コア開発
- `make dev` - ローカル開発サーバーを実行
- `make build` - デバッグバージョンをビルド（`dist/muscle-dreamer`に出力）
- `make build-release` - 最適化されたリリースバージョンをビルド
- `make test` - 単体テストを実行（`go test ./...`）
- `make test-all` - 統合テストを含む全テストを実行
- `make lint` - コード解析を実行（`golangci-lint run`）
- `make format` - コードをフォーマット（`go fmt ./...`と`goimports -w .`）

### プラットフォーム固有のビルド
- `make build-web` - WebAssemblyバージョンをビルド（`dist/web/`に出力）
- `make build-all` - 全プラットフォーム用にビルド（Windows、Linux、macOS、WebAssembly）
- `make docker-build` - Dockerコンテナを使用してクロスコンパイル

### Docker開発
- `make docker-setup` - Docker開発環境を初期化
- `make docker-dev` - 開発コンテナを開始（ゲーム開発 + ウェブ開発）
  - メイン開発: http://localhost:8080
  - ウェブ開発: http://localhost:3000

### ウェブ開発
- `cd web && npm run dev` - ウェブ開発サーバーを開始
- `cd web && npm run build` - ウェブアセットをビルド
- `cd web && npm start` - プロダクションウェブビルドを提供

### メンテナンス
- `make clean` - ビルド成果物を削除し、Dockerをクリーン
- `make deps` - Goモジュール依存関係を更新

## アーキテクチャ概要

### コアコンポーネント
- **ECSフレームワーク**（`internal/core/`）- ゲームオブジェクト用のEntity Component System
- **テーマシステム**（`internal/theme/`）- 動的コンテンツローディングと切り替え
- **MODシステム**（`internal/mod/`）- サンドボックス化された安全なプラグインアーキテクチャ
- **プラットフォーム層**（`internal/platform/`）- クロスプラットフォーム抽象化
- **UIシステム**（`internal/ui/`）- ユーザーインターフェース管理

### 主要ディレクトリ
- `cmd/game/main.go` - メインアプリケーションエントリーポイント
- `internal/core/game.go` - コアゲームエンジンとメインゲームループ
- `themes/` - テーマコンテンツ（アセット、設定、スクリプト）
- `mods/` - `enabled/`、`disabled/`、`staging/`に整理されたMODコンテンツ
- `assets/` - コアゲームアセット（スプライト、オーディオ、フォント、UI）
- `config/` - ゲーム設定ファイル（YAML）
- `saves/` - ローカルセーブデータとユーザー設定
- `web/` - WebAssembly展開ファイルとウェブ開発

### 技術スタック
- **言語**: Go 1.22
- **ゲームエンジン**: Ebitengine v2.6.3
- **ビルドシステム**: Make + Docker
- **ウェブフレームワーク**: Express付きNode.js（ウェブ配信用）
- **アセット形式**: PNG/JPG（画像）、OGG/WAV（オーディオ）、YAML（設定）

## 開発ガイドライン

### コード整理
- プライベートパッケージ用の`internal/`を含む標準Goプロジェクトレイアウトに従う
- ゲームロジックにはEntity Component Systemパターンを使用
- プラットフォーム固有機能にはインターフェースを実装
- ゲームロジックをレンダリングと入力処理から分離

### アセット管理
- コアアセットはタイプ別に整理された`assets/`ディレクトリに配置
- テーマ固有のアセットは`themes/[theme-name]/assets/`に配置
- コミット前に画像とオーディオファイルを最適化
- 全設定ファイルにYAMLを使用

### テーマ開発
- 各テーマは`themes/`内の自己完結型ディレクトリ
- `theme.yaml`設定ファイルを含む必要がある
- 構造に従う: `assets/`、`scripts/`、`localization/`、`metadata/`
- 内蔵バリデータでテーマをテスト

### MOD開発
- MODは制限されたファイルシステムアクセスを持つ安全なサンドボックス内で実行
- ゲームインタラクションには提供されたAPIを使用
- 開発用には`mods/staging/`に、アクティブなMODには`mods/enabled/`に配置
- 全MODスクリプトはセキュリティ検証の対象

### テスト
- 全コアシステムに単体テストを記述
- コンポーネント間機能には統合テストを使用
- リリース前に複数プラットフォームでテスト
- 重要なパスにはパフォーマンスベンチマークを含める

### セキュリティ考慮事項
- 全MODコードはサンドボックス環境で実行
- ファイルシステムアクセスは指定されたディレクトリに制限
- MODのネットワークアクセスは無効
- ロード前に全ユーザー提供コンテンツを検証

## プラットフォーム固有の注意事項

### WebAssembly
- WebAssemblyビルドはサイズと読み込み速度に最適化
- ネイティブビルドと比較して一部機能が制限される場合がある
- ブラウザ互換性テストにはウェブ開発サーバーを使用
- アセットは`web/`のExpress.jsサーバーを通じて提供

### デスクトッププラットフォーム
- ネイティブビルドは高度なオーディオとグラフィックスを含む全機能セットをサポート
- クロスコンパイルはDockerコンテナを通じて処理
- ビルド中にプラットフォーム固有の最適化が適用

## ファイル命名と整理

### 設定ファイル
- `config/game.yaml` - メインゲーム設定
- `themes/[name]/theme.yaml` - テーマ定義
- `mods/[name]/mod.yaml` - MODメタデータ

### アセット整理
- スプライト: `assets/sprites/`または`themes/[name]/assets/sprites/`
- オーディオ: `assets/audio/`または`themes/[name]/assets/audio/`
- フォント: `assets/fonts/`
- UI要素: `assets/ui/`

## 共通開発タスク

### 新しいゲーム機能の追加
1. ECSパターンに従ってコンポーネントとシステムを設計
2. クロスプラットフォームサポートが必要な場合は`internal/core/`にインターフェースを追加
3. `internal/platform/`にプラットフォーム固有のコードを実装
4. 適切なYAMLファイルに設定オプションを追加
5. テストを記述し、ドキュメントを更新

### 新しいテーマの作成
1. `themes/[theme-name]/`にディレクトリ構造を作成
2. theme.yaml設定を設計
3. 命名規則に従ってアセットを作成
4. テーマバリデータでテスト
5. 配布用にパッケージ化

### 問題のデバッグ
- デバッグシンボル付きの開発ビルドには`make dev`を使用
- 詳細なエラーメッセージについてはコンソール出力を確認
- WebAssemblyデバッグにはブラウザ開発者ツールを使用
- 開発モードで詳細ログを有効化

## 重要な注意事項

- 全開発はオフラインファースト機能を維持する必要がある
- セキュリティは最重要、特にMODシステムについて
- パフォーマンス目標: 60 FPS、<256MBメモリ使用量、<3秒起動時間
- ゲームはインターネット接続なしで完全に動作するよう設計
- セーブファイルとテーマ形式の後方互換性を維持
- 英語でドキュメントを作成した後、日本語翻訳を作成する
