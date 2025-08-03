# GitHub Copilot Custom Instructions (English Translation)

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
9. Always write comments and documentation in both English and Japanese

When proposing code, briefly explain why you chose that implementation.

---
(Japanese Translation)

# GitHub Copilot カスタムインストラクション

## プロジェクト概要
- **プロジェクト名**: Muscle Dreamer
- **種類**: 2Dサバイバルアクションローグライクゲーム
- **言語**: Go 1.22
- **ゲームエンジン**: Ebitengine v2.6.3
- **アーキテクチャ**: Entity Component System (ECS)
- **開発環境**: GoLand IDE
- **チーム開発**: あり

## コーディング規約とスタイル

### Go言語の規約
- Goのデファクトスタンダードに従う
- `go fmt`、`goimports`を使用したフォーマット
- `golangci-lint`によるコード解析を通す
- 標準Goプロジェクトレイアウト（`internal/`パッケージ使用）

### 命名規則
- パッケージ名: 小文字、短く、説明的
- 関数/メソッド名: キャメルケース、公開は大文字開始
- 変数名: キャメルケース、スコープに応じた長さ
- 定数名: キャメルケース、または大文字スネークケース

### コメント
- 公開関数/型には必ずコメントを記述
- パッケージレベルのdocコメントを含める
- 複雑なロジックには説明コメントを追加

## アーキテクチャ設計原則

### ECS (Entity Component System)
- エンティティ、コンポーネント、システムの責任を明確に分離
- コンポーネントはデータのみ、ロジックはシステムに実装
- システム間の依存関係を最小化

### ディレクトリ構造
```
cmd/game/          # アプリケーションエントリーポイント
internal/core/     # コアECSフレームワーク
internal/theme/    # テーマシステム
internal/mod/      # MODシステム（セキュアサンドボックス）
internal/platform/ # クロスプラットフォーム抽象化
internal/ui/       # UIシステム
themes/           # テーマコンテンツ
mods/             # MODコンテンツ
assets/           # コアゲームアセット
config/           # YAML設定ファイル
saves/            # セーブデータ
web/              # WebAssembly展開
```

### レイヤー分離
- ゲームロジックをレンダリングから分離
- プラットフォーム固有機能はインターフェースで抽象化
- ビジネスロジックと外部依存を分離

## Ebitengine特有の考慮事項

### 基本パターン
- `Game`構造体に`Update()`と`Draw()`メソッドを実装
- `ebiten.RunGame()`でゲームループを開始
- 60 FPS維持を目標とした効率的な処理

### リソース管理
- 画像リソースは`ebiten.NewImageFromImage()`で管理
- 音声は`audio.NewPlayer()`で再生制御
- アセット読み込みは初期化時に実行、ゲームループでは避ける

### 入力処理
- `inpututil`パッケージを活用した入力状態管理
- キーボード、マウス、ゲームパッド対応
- 入力のフレーム単位での状態追跡

## テスト戦略

### 単体テスト
- 全コアシステムに対してテストを作成
- テストファイルは`*_test.go`命名規則
- `testing`パッケージを使用
- モックを使用した外部依存の分離

### 統合テスト
- システム間の相互作用をテスト
- ゲームループの動作確認
- プラットフォーム固有機能の動作検証

### テストカバレッジ
- クリティカルパス: 90%以上
- 一般的な機能: 80%以上
- パフォーマンステストも含める

## セキュリティ考慮事項

### MODシステム
- 全MODコードはサンドボックス環境で実行
- ファイルシステムアクセスを制限ディレクトリに限定
- ネットワークアクセス無効
- ユーザー提供コンテンツの検証を必須とする

### データ検証
- 設定ファイル読み込み時の検証
- セーブデータの整合性チェック
- 外部アセットの安全性確認

## パフォーマンス要件

### 目標値
- フレームレート: 60 FPS維持
- メモリ使用量: 256MB未満
- 起動時間: 3秒未満
- クロスプラットフォーム対応（WebAssembly含む）

### 最適化指針
- ガベージコレクションの負荷軽減
- 描画処理の効率化
- メモリプールの活用
- プロファイリングツールの積極的活用

## 設定管理

### YAML使用
- 全設定ファイルにYAMLフォーマット採用
- 構造化された設定の階層管理
- 環境ごとのオーバーライド機能

### 設定ファイル例
- `config/game.yaml`: メインゲーム設定
- `themes/[name]/theme.yaml`: テーマ定義
- `mods/[name]/mod.yaml`: MODメタデータ

## 開発フロー

### 日常的なコマンド
- `make dev`: ローカル開発実行
- `make test`: 単体テスト実行
- `make lint`: コード解析
- `make format`: コードフォーマット

### ビルド・デプロイ
- `make build`: デバッグビルド
- `make build-release`: リリースビルド
- `make build-all`: 全プラットフォーム向けビルド
- `make build-web`: WebAssembly版ビルド

## エラーハンドリング

### Go言語パターン
- 明示的なエラー処理（`if err != nil`）
- カスタムエラー型の活用
- エラーのラップと文脈情報の追加
- パニックは回復可能にし、適切にログ出力

### ログ出力
- 構造化ログの使用
- ログレベルの適切な設定
- デバッグ情報の詳細な出力（開発モード）

## 重要な制約事項

### オフラインファースト
- インターネット接続なしで完全動作
- ローカルリソースのみでゲーム実行
- セーブデータのローカル保存

### 後方互換性
- セーブファイル形式の互換性維持
- テーマ形式の安定性確保
- APIの非破壊的変更

---

## Copilotへの指示

上記の情報に基づいて、以下を重視してコード生成してください：

1. **Go言語の慣用的な書き方**を常に採用
2. **ECSアーキテクチャ**に適合した設計
3. **可読性**を最優先し、適切なコメントを含める
4. **テストコード**も同時に提案
5. **エラーハンドリング**を適切に実装
6. **パフォーマンス**を考慮した実装
7. **セキュリティ**要件（特にMODシステム）を遵守
8. **クロスプラットフォーム**対応を意識した抽象化
9. コメントやドキュメントは英語と日本語の両方をかならず記述する

コード提案時は、なぜその実装を選択したかの理由も簡潔に説明してください。
