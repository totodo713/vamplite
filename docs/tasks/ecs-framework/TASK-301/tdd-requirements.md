# TASK-301: ModECSAPI実装 - 要件定義

## 概要

MOD向けの制限されたECS APIを実装します。セキュリティを最優先として、MODが安全にECSシステムを操作できるサンドボックス化されたAPIを提供します。

## 機能要件

### FR-301-001: 制限されたEntity操作API
**要件**: MODはエンティティの作成・削除・属性変更を制限された範囲で実行できる
**実装詳細**:
- `ModEntityAPI` インターフェース
- エンティティ作成上限設定（デフォルト100個/MOD）
- MOD専用エンティティタグ付与（`mod:{mod_name}`）
- 他MODのエンティティへのアクセス禁止
- システムエンティティへのアクセス禁止

### FR-301-002: 制限されたComponent操作API  
**要件**: MODは許可されたコンポーネントのみ操作可能
**実装詳細**:
- `ModComponentAPI` インターフェース
- 許可コンポーネント型ホワイトリスト
- MODが作成したコンポーネントのみ削除可能
- 読み取り専用コンポーネント（Transform等）の保護
- 危険コンポーネント（FileIO等）の完全ブロック

### FR-301-003: 制限されたQuery操作API
**要件**: MODは安全なクエリでエンティティ検索可能
**実装詳細**:
- `ModQueryAPI` インターフェース  
- MOD作成エンティティのみ検索対象
- システムエンティティの検索結果除外
- クエリ実行回数制限（1000回/フレーム）
- 複雑クエリの時間制限（10ms）

### FR-301-004: サンドボックス化されたSystem実行
**要件**: MODはサンドボックス内でシステムロジックを実行
**実装詳細**:
- `ModSystemAPI` インターフェース
- システム実行時間制限（5ms/フレーム）
- メモリ使用量制限（10MB/MOD）
- ファイルシステムアクセス完全ブロック
- ネットワークアクセス完全ブロック

## 非機能要件

### NFR-301-001: セキュリティ
- **脅威モデル**: 悪意のあるMODからのシステム保護
- **防御対象**: 
  - パストラバーサル攻撃（`../../../etc/passwd`）
  - システムコマンド実行（`rm -rf /`）
  - メモリ破壊攻撃
  - DoS攻撃（無限ループ、メモリリーク）
- **セキュリティ境界**: MODとコアシステムの完全分離

### NFR-301-002: パフォーマンス
- **API呼び出しオーバーヘッド**: <100μs
- **サンドボックス実行オーバーヘッド**: <5%
- **メモリ使用量**: <10MB/MOD
- **システム実行時間**: <5ms/フレーム

### NFR-301-003: 可用性
- **MODエラー隔離**: MODエラーがコアシステムに影響しない
- **グレースフル劣化**: MOD失敗時もゲーム継続
- **エラー回復**: MOD再ロード・無効化機能

## 技術アーキテクチャ

### API構造
```go
// MOD向けメインAPI
type ModECSAPI interface {
    Entities() ModEntityAPI
    Components() ModComponentAPI  
    Queries() ModQueryAPI
    Systems() ModSystemAPI
}

// セキュリティ制約付きEntity操作
type ModEntityAPI interface {
    Create(tags ...string) (EntityID, error)
    Delete(id EntityID) error
    GetTags(id EntityID) ([]string, error)
    // SetParent等の階層操作は禁止
}

// 制限されたComponent操作
type ModComponentAPI interface {
    Add(entity EntityID, component Component) error
    Get(entity EntityID, componentType ComponentType) (Component, error)
    Remove(entity EntityID, componentType ComponentType) error
    // システムコンポーネントの操作は禁止
}

// 安全なQuery操作
type ModQueryAPI interface {
    Find(query Query) ([]EntityID, error)
    Count(query Query) (int, error)
    // システムエンティティは結果から除外
}

// サンドボックス化されたSystem実行
type ModSystemAPI interface {
    Register(system ModSystem) error
    Unregister(systemID string) error
    // 実行は自動的にサンドボックス内で制限付き実行
}
```

### サンドボックス実装
```go
// MOD実行コンテキスト
type ModContext struct {
    ModID           string
    MaxEntities     int           // デフォルト100
    MaxMemory       int64         // デフォルト10MB  
    MaxExecutionTime time.Duration // デフォルト5ms
    AllowedComponents []ComponentType
    CreatedEntities  []EntityID
}

// リソース制限監視
type ResourceLimiter struct {
    entityCount   int
    memoryUsage   int64
    executionTime time.Duration
    violations    []string
}
```

## セキュリティ仕様

### アクセス制御
1. **Entity制限**:
   - MOD作成エンティティのみアクセス可能
   - システムエンティティは完全に非表示
   - 他MODエンティティへのアクセス禁止

2. **Component制限**:
   - ホワイトリスト式の許可型のみ操作可能
   - 危険コンポーネント（FileIO, NetworkIO等）は完全ブロック
   - システムコンポーネントは読み取り専用

3. **実行制限**:
   - CPU時間制限: 5ms/フレーム
   - メモリ制限: 10MB/MOD
   - ファイルアクセス: 完全ブロック
   - ネットワークアクセス: 完全ブロック

### 監査・ログ
- API呼び出し履歴記録
- セキュリティ違反検出・記録
- リソース使用量監視
- 異常動作検出・アラート

## 受け入れ基準

### 機能受け入れ基準
- [ ] MODが制限範囲内でECS操作を実行できる
- [ ] 許可されていない操作が確実にブロックされる
- [ ] MOD間の分離が保証される
- [ ] システムの安定性が維持される

### セキュリティ受け入れ基準
- [ ] パストラバーサル攻撃が100%防御される
- [ ] システムコマンド実行が100%ブロックされる
- [ ] メモリ破壊攻撃が検出・防御される
- [ ] DoS攻撃が制限によって軽減される

### パフォーマンス受け入れ基準
- [ ] API呼び出しオーバーヘッド<100μs
- [ ] サンドボックス実行オーバーヘッド<5%
- [ ] MODメモリ使用量<10MB
- [ ] 実行時間制限遵守率100%

## 依存関係

### 前提条件
- ✅ TASK-204: MetricsCollector完了
- ✅ ECSコアシステム（EntityManager, ComponentStore, SystemManager）
- ✅ EventBusシステム

### 連携コンポーネント
- `internal/core/ecs/world.go` - World実装との統合
- `internal/core/ecs/metrics.go` - メトリクス収集との連携
- （予定）`internal/mod/` - MODシステムとの統合

## 実装戦略

### フェーズ分割
1. **Phase 1**: 基本API構造とインターフェース定義
2. **Phase 2**: セキュリティ制約の実装  
3. **Phase 3**: サンドボックス実行環境
4. **Phase 4**: 監査・ログ機能
5. **Phase 5**: パフォーマンス最適化

### リスク軽減
- **セキュリティリスク**: 段階的権限追加、包括的テスト
- **パフォーマンスリスク**: 継続的ベンチマーキング
- **複雑性リスク**: インターフェース設計の単純化

---

**作成日時**: 2025-08-08  
**担当**: Claude Code  
**レビュー状態**: 要レビュー