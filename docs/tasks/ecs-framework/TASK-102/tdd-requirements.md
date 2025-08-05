# TASK-102: ComponentStore実装 - 詳細要件定義

## タスク概要

- **タスクID**: TASK-102
- **タスクタイプ**: TDD実装
- **要件リンク**: REQ-002, REQ-005, REQ-104
- **依存タスク**: TASK-101 (EntityManager実装) ✅

## 機能要件

### F-102-001: スパースセット実装

**要件**: Entity → Component マッピングにスパースセットアルゴリズムを使用した高速アクセス実現

**詳細**:
- エンティティIDを直接インデックスとして使用可能なスパースセット構造
- O(1)でのコンポーネント存在確認
- O(1)でのコンポーネント追加・削除
- メモリ効率を考慮した動的サイズ調整

**受け入れ基準**:
- [ ] 10,000エンティティまでの線形スケーラビリティ
- [ ] GetComponent操作が1ms未満で完了
- [ ] AddComponent/RemoveComponent操作が0.1ms未満で完了
- [ ] メモリ使用量がエンティティ数に比例（O(n)）

### F-102-002: 型安全なコンポーネントストレージ

**要件**: コンパイル時・実行時の型安全性を保証したコンポーネント管理

**詳細**:
- 型情報による自動型検証
- 型不整合時の明確なエラーメッセージ
- ジェネリクスを使用した型安全なAPI設計
- リフレクションを使用しない高速な型検査

**受け入れ基準**:
- [ ] 間違った型でのコンポーネント追加時にコンパイルエラー
- [ ] 実行時型検証でのわかりやすいエラーメッセージ
- [ ] 型変換処理でのパフォーマンス劣化なし（<5%）
- [ ] 全コンポーネント型で型安全性テスト通過

### F-102-003: Structure of Arrays (SoA) メモリレイアウト

**要件**: CPU キャッシュ効率を最大化するメモリレイアウトの実装

**詳細**:
- 同一コンポーネント型のデータを連続メモリ領域に配置
- イテレーション時のキャッシュミス最小化
- SIMD命令での並列処理最適化準備
- メモリ断片化の防止

**受け入れ基準**:
- [ ] 大量コンポーネントイテレーション時のキャッシュ効率向上（20%以上）
- [ ] 連続メモリアクセスパターンの確認
- [ ] メモリ断片化率5%以下維持
- [ ] SIMD対応準備完了

### F-102-004: コンポーネント型動的登録

**要件**: 実行時にコンポーネント型を動的に登録・管理する機能

**詳細**:
- 新しいコンポーネント型の実行時登録
- 型情報の自動収集・管理
- 登録済み型の一覧・検索機能
- 型登録時の初期化・設定

**受け入れ基準**:
- [ ] 実行時に50種類以上のコンポーネント型登録可能
- [ ] 型登録処理が10ms未満で完了
- [ ] 登録済み型の検索がO(1)で実行
- [ ] 重複登録時の適切なエラーハンドリング

### F-102-005: バルク操作・メモリ効率最適化

**要件**: 大量データ処理に最適化されたバッチ操作機能

**詳細**:
- 複数エンティティの一括コンポーネント追加・削除
- バッチ処理時のメモリアロケーション最小化
- 処理済みエンティティの進捗追跡
- エラー時の部分的ロールバック機能

**受け入れ基準**:
- [ ] 1000エンティティの一括処理が100ms未満で完了
- [ ] バッチ処理中のメモリアロケーション数を50%削減
- [ ] エラー時の一貫性保証（ACID特性）
- [ ] 進捗監視・キャンセル機能の動作確認

## 非機能要件

### NFR-102-001: パフォーマンス要件

**要件**: 高負荷環境での安定したパフォーマンス維持

**詳細**:
- GetComponent: < 1ms (99th percentile)
- AddComponent: < 0.1ms (99th percentile)  
- RemoveComponent: < 0.1ms (99th percentile)
- 大量コンポーネント処理: 10,000個/秒以上

**測定条件**:
- エンティティ数: 10,000個
- コンポーネント型数: 50種類
- 同時実行: 4スレッド
- メモリ制限: 256MB

### NFR-102-002: メモリ効率要件

**要件**: メモリ使用量の最適化とメモリリーク防止

**詳細**:
- エンティティあたりメモリ使用量: < 100B
- メモリ断片化率: < 5%
- 24時間連続実行時のメモリリーク: < 50MB
- ガベージコレクション頻度の最小化

### NFR-102-003: 並行性要件

**要件**: マルチスレッド環境での安全性と性能

**詳細**:
- 読み取り専用操作の並列実行サポート
- 書き込み操作の排他制御
- デッドロックの防止
- レースコンディションの排除

**受け入れ基準**:
- [ ] 4スレッド同時読み取りでパフォーマンス劣化なし
- [ ] 読み書き混在時の一貫性保証
- [ ] 1000回のコンカレンシーテスト全て成功
- [ ] データレース検出ツールでの検証通過

## 技術仕様

### アーキテクチャ設計

```go
type ComponentStore struct {
    // スパースセット: 高速エンティティ検索
    sparseSets map[ComponentType]*SparseSet
    
    // メモリプール: 効率的メモリ管理
    memoryPools map[ComponentType]*MemoryPool
    
    // コンポーネントデータ: SoAレイアウト
    components map[ComponentType]map[EntityID]Component
    
    // エンティティ追跡: 所有コンポーネント管理
    entities map[EntityID]map[ComponentType]bool
    
    // 並行性制御
    mutex sync.RWMutex
    
    // 型登録管理
    registeredTypes map[ComponentType]bool
}
```

### 主要インターフェース設計

```go
// コンポーネント管理
func (s *ComponentStore) RegisterComponentType(componentType ComponentType, poolSize int) error
func (s *ComponentStore) AddComponent(entity EntityID, component Component) error
func (s *ComponentStore) GetComponent(entity EntityID, componentType ComponentType) (Component, error)
func (s *ComponentStore) RemoveComponent(entity EntityID, componentType ComponentType) error
func (s *ComponentStore) HasComponent(entity EntityID, componentType ComponentType) bool

// バルク操作
func (s *ComponentStore) AddComponents(entities []EntityID, components []Component) error
func (s *ComponentStore) RemoveComponents(entities []EntityID, componentType ComponentType) error
func (s *ComponentStore) GetComponentsBatch(entities []EntityID, componentType ComponentType) ([]Component, error)

// クエリ操作
func (s *ComponentStore) GetEntitiesWithComponent(componentType ComponentType) []EntityID
func (s *ComponentStore) GetEntitiesWithComponents(componentTypes []ComponentType) []EntityID
func (s *ComponentStore) GetAllComponents(entity EntityID) []Component

// 統計・監視
func (s *ComponentStore) GetStorageStatistics() []*StorageStats
func (s *ComponentStore) GetComponentCount(componentType ComponentType) int
func (s *ComponentStore) GetEntityCount() int
```

## エラーハンドリング設計

### エラー種別

1. **ComponentTypeNotRegisteredError**: 未登録コンポーネント型アクセス
2. **DuplicateComponentError**: 重複コンポーネント追加
3. **ComponentNotFoundError**: 存在しないコンポーネント取得
4. **MemoryAllocationError**: メモリ不足時の処理
5. **ConcurrencyError**: 並行アクセス競合エラー

### エラーレスポンス

```go
type ComponentStoreError struct {
    Type      ErrorType
    Message   string
    EntityID  EntityID
    ComponentType ComponentType
    Cause     error
}

func (e *ComponentStoreError) Error() string
func (e *ComponentStoreError) Unwrap() error
func (e *ComponentStoreError) Is(target error) bool
```

## テスト戦略

### 単体テストカバレッジ

- **機能テスト**: 全API機能の正常系・異常系テスト
- **パフォーマンステスト**: レスポンス時間・メモリ使用量測定
- **ストレステスト**: 大量データ処理・長期実行テスト
- **並行性テスト**: マルチスレッド環境での安全性検証

### 統合テストシナリオ

- **EntityManager連携**: エンティティライフサイクル管理
- **SystemManager連携**: システム実行時のコンポーネント操作
- **QueryEngine連携**: 複雑クエリでのパフォーマンス検証
- **MemoryManager連携**: メモリプール統合動作

## 完了条件

### 実装完了条件

- [ ] 全機能要件の実装完了
- [ ] 全単体テストの通過（カバレッジ95%以上）
- [ ] 全統合テストの通過
- [ ] パフォーマンス要件の達成
- [ ] メモリ効率要件の達成
- [ ] 並行性テストの通過

### 品質保証条件

- [ ] コードレビューの完了
- [ ] 静的解析ツールでの検証通過
- [ ] ドキュメント整備完了
- [ ] ベンチマーク結果の記録
- [ ] 技術的負債の解消

## 実装優先度

### Phase 1: 基本機能実装
1. コンポーネント型登録機能
2. 基本CRUD操作（Add/Get/Remove）
3. エンティティ存在確認機能

### Phase 2: パフォーマンス最適化
1. スパースセット実装
2. SoAメモリレイアウト
3. メモリプール統合

### Phase 3: 高度機能実装
1. バルク操作機能
2. 統計・監視機能
3. エラーハンドリング強化

### Phase 4: 最終最適化
1. 並行性制御強化
2. パフォーマンス微調整
3. 統合テスト・品質保証

---

## 次のステップ

この要件定義を基に、次は **tdd-testcases.md** でテストケースの詳細設計を行います。