# TASK-201: QueryEngine実装 - 詳細要件定義書

## 概要

高速なエンティティクエリを実現するQueryEngineを実装します。ビットセット操作による効率的なコンポーネント組み合わせ検索、QueryBuilderによる複雑なクエリ対応、キャッシュシステムによる最適化を提供します。

## 要件リンク

- **REQ-004**: 高速クエリエンジン
- **NFR-002**: パフォーマンス要求
- **NFR-003**: スケーラビリティ要求

## 機能要件

### 1. ビットセットベースクエリ (REQ-004-001)

**目的**: 高速なコンポーネント組み合わせ検索

**要件**:
- エンティティのコンポーネント保有状況をビットセットで表現
- ビット演算による高速なクエリ実行
- コンポーネントタイプごとのビット位置管理
- 動的なコンポーネント登録対応

**技術要件**:
```go
type ComponentBitSet uint64
type ArchetypeSignature ComponentBitSet

// 基本操作
func (b ComponentBitSet) Has(componentType ComponentType) bool
func (b ComponentBitSet) Set(componentType ComponentType) ComponentBitSet
func (b ComponentBitSet) Clear(componentType ComponentType) ComponentBitSet
func (b ComponentBitSet) And(other ComponentBitSet) ComponentBitSet
func (b ComponentBitSet) Or(other ComponentBitSet) ComponentBitSet
```

### 2. QueryBuilder実装 (REQ-004-002)

**目的**: 複雑なクエリ条件の構築と実行

**要件**:
- Fluent API による直感的なクエリ記述
- With/Without/Or/Andオペレーター対応
- ネストしたクエリ条件対応
- クエリの最適化と実行計画生成

**API設計**:
```go
type QueryBuilder interface {
    With(componentTypes ...ComponentType) QueryBuilder
    Without(componentTypes ...ComponentType) QueryBuilder
    Or() QueryBuilder
    And() QueryBuilder
    Build() Query
    Execute(world World) EntityIterator
}

// 使用例
entities := world.Query().
    With(TransformComponent, SpriteComponent).
    Without(DisabledComponent).
    Or().
    With(TransformComponent, PhysicsComponent).
    Execute(world)
```

### 3. アーキタイプシステム (REQ-004-003)

**目的**: 同じコンポーネント構成のエンティティをグループ化

**要件**:
- アーキタイプ自動生成・管理
- エンティティのアーキタイプ間移動対応
- アーキタイプベースの効率的なイテレーション
- メモリ局所性の最適化

**設計**:
```go
type Archetype struct {
    signature    ComponentBitSet
    entities     []EntityID
    componentStorages map[ComponentType]interface{}
}

type ArchetypeManager interface {
    GetArchetype(signature ComponentBitSet) *Archetype
    GetOrCreateArchetype(signature ComponentBitSet) *Archetype
    MoveEntity(entityID EntityID, from, to *Archetype)
    GetMatchingArchetypes(querySignature ComponentBitSet) []*Archetype
}
```

### 4. クエリキャッシュシステム (REQ-004-004)

**目的**: 頻繁なクエリ結果のキャッシュによる高速化

**要件**:
- クエリ結果の自動キャッシュ
- エンティティ・コンポーネント変更時のキャッシュ無効化
- LRU/TTL キャッシュポリシー
- メモリ使用量制限

**設計**:
```go
type QueryCache interface {
    Get(query Query) (EntityIterator, bool)
    Set(query Query, result EntityIterator)
    Invalidate(affectedComponentTypes ...ComponentType)
    Clear()
    GetStats() CacheStats
}
```

### 5. 並列クエリ実行 (REQ-004-005)

**目的**: 大量エンティティに対する並列処理対応

**要件**:
- クエリ結果の並列処理対応
- チャンク分割による効率的な並列実行
- ワーカープール管理
- 同期プリミティブによる安全性確保

## 非機能要件

### パフォーマンス要件 (NFR-002)

- **クエリ実行時間**: 10,000エンティティに対して1ms以内
- **メモリオーバーヘッド**: エンティティ1つあたり8バイト以内
- **CPU使用率**: シングルスレッド実行で90%以上の効率
- **キャッシュヒット率**: 80%以上（頻繁なクエリに対して）

### スケーラビリティ要件 (NFR-003)

- **最大エンティティ数**: 100,000エンティティ
- **同時クエリ実行数**: 100クエリ
- **コンポーネントタイプ数**: 64タイプ（ビットセット制限）
- **アーキタイプ数**: 1,000アーキタイプ

### メモリ効率要件 (NFR-001)

- **メモリ断片化**: 最小限
- **ガベージコレクション**: 低頻度
- **メモリプール活用**: 頻繁な割り当て/解放の最適化

## システムアーキテクチャ

### コンポーネント構成

```
QueryEngine
├── ArchetypeManager    # アーキタイプ管理
├── QueryBuilder        # クエリ構築
├── QueryExecutor       # クエリ実行エンジン
├── QueryCache          # クエリキャッシュ
├── BitSetManager       # ビットセット管理
└── ParallelExecutor    # 並列実行管理
```

### データフロー

1. **クエリ構築フェーズ**
   - QueryBuilderでクエリ条件定義
   - ビットセット署名生成
   - クエリ最適化処理

2. **キャッシュチェックフェーズ**  
   - QueryCacheでキャッシュヒット確認
   - ヒット時は結果を即座に返却

3. **実行フェーズ**
   - ArchetypeManagerでマッチするアーキタイプ検索
   - 並列実行によるエンティティイテレーション
   - 結果キャッシュへの保存

### インターフェース設計

```go
type QueryEngine interface {
    // クエリ作成
    NewQuery() QueryBuilder
    
    // 直接実行
    FindEntitiesWith(componentTypes ...ComponentType) EntityIterator
    FindEntitiesWithout(componentTypes ...ComponentType) EntityIterator
    
    // 統計情報
    GetStats() QueryEngineStats
    
    // 設定管理
    SetCachePolicy(policy CachePolicy)
    SetParallelism(enabled bool, workerCount int)
}
```

## 品質基準

### テスト要件

1. **単体テスト**: 各コンポーネントの個別動作確認
2. **統合テスト**: コンポーネント間の連携動作確認  
3. **パフォーマンステスト**: NFR達成確認
4. **並行性テスト**: 並列実行時の安全性確認

### 成功基準

- [ ] 全単体テストが成功（100%カバレッジ）
- [ ] パフォーマンス要件を満たす
- [ ] メモリリークなし（24時間実行）
- [ ] 並行性エラーなし（競合状態検出ツール使用）

## 実装制約

### 技術的制約
- Go 1.22以上
- 標準ライブラリのみ使用（外部依存なし）
- ビットセット64bit制限（コンポーネント数≤64）

### 設計制約
- 既存ECSインターフェースとの互換性維持
- メモリアロケーション最小化
- 型安全性の確保

## リスク分析

### 高リスク
- **ビットセット制限**: 64コンポーネント超過時の拡張性
- **メモリ断片化**: 大量アーキタイプ生成時のメモリ効率

### 中リスク  
- **キャッシュ一貫性**: 頻繁な更新時のキャッシュ無効化オーバーヘッド
- **並列処理**: 複雑な同期処理による潜在的デッドロック

### 対策
- ビットセット拡張アルゴリズムの事前設計
- メモリプール活用による断片化防止
- 段階的並列化実装によるデバッグ性確保

---

**次ステップ**: この要件定義に基づき、詳細なテストケースを設計します。