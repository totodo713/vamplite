# TASK-001: プロジェクト構造・基本設定 - 直接実装

## 実装目標

ECSフレームワーク開発のための基盤環境を構築し、開発効率と品質を向上させる開発インフラを整備する。

## 実装ステップ

### Step 1: ECSディレクトリ構造作成

```
internal/core/ecs/
├── entity.go          # EntityManager実装
├── component.go       # ComponentStore実装  
├── system.go          # SystemManager実装
├── query.go           # QueryEngine実装
├── memory.go          # MemoryManager実装
├── world.go           # World統合管理
├── events.go          # EventBus実装
├── metrics.go         # MetricsCollector実装
├── types.go           # 基本型定義
├── errors.go          # エラー型定義
└── tests/             # テストファイル
    ├── entity_test.go
    ├── component_test.go
    ├── system_test.go
    ├── query_test.go
    ├── memory_test.go
    ├── world_test.go
    ├── events_test.go
    ├── metrics_test.go
    ├── integration_test.go
    └── performance_test.go

internal/core/components/
├── transform.go       # TransformComponent
├── sprite.go          # SpriteComponent
├── physics.go         # PhysicsComponent
├── health.go          # HealthComponent
├── ai.go              # AIComponent
└── tests/
    ├── transform_test.go
    ├── sprite_test.go
    ├── physics_test.go
    ├── health_test.go
    └── ai_test.go

internal/core/systems/
├── movement.go        # MovementSystem
├── rendering.go       # RenderingSystem
├── physics.go         # PhysicsSystem
├── collision.go       # CollisionSystem
├── audio.go           # AudioSystem
└── tests/
    ├── movement_test.go
    ├── rendering_test.go
    ├── physics_test.go
    ├── collision_test.go
    └── audio_test.go

internal/mod/ecs/
├── sandbox.go         # MODサンドボックス
├── validator.go       # MODバリデーター
├── bridge.go          # Go-Luaブリッジ
├── security.go        # セキュリティポリシー
└── tests/
    ├── sandbox_test.go
    ├── validator_test.go
    ├── bridge_test.go
    └── security_test.go
```

### Step 2: Go モジュール依存関係定義

ECSフレームワークに必要な依存関係を`go.mod`に追加：

- `github.com/hajimehoshi/ebiten/v2` - ゲームエンジン
- テストフレームワーク・ベンチマーク用ライブラリ
- Lua統合ライブラリ
- パフォーマンス監視ライブラリ

### Step 3: Makefile作成

開発効率を向上させるMakefileターゲット：

- `make build` - デバッグビルド
- `make build-release` - リリースビルド
- `make test` - 単体テスト実行
- `make test-all` - 全テスト実行（統合・パフォーマンス含む）
- `make test-coverage` - テストカバレッジ取得
- `make lint` - コード品質チェック
- `make format` - コードフォーマット
- `make benchmark` - パフォーマンステスト
- `make docs` - ドキュメント生成
- `make clean` - ビルド成果物削除

### Step 4: golangci-lint設定

高品質なコードを保証するlint設定：

- Go推奨設定
- パフォーマンス・セキュリティ関連チェック
- ECS特有のパターンチェック
- テストコード品質チェック

### Step 5: GitHub Actions CI/CD設定

自動化されたCI/CDパイプライン：

- 複数Goバージョンでのテスト
- lintチェック
- テストカバレッジ測定・レポート
- パフォーマンステスト実行
- セキュリティスキャン
- ドキュメント自動生成

## 実装開始

### 1. ディレクトリ構造作成