# TASK-201: QueryEngine実装 - テストケース仕様書

## テスト戦略

QueryEngineの各機能について、単体テスト・統合テスト・パフォーマンステストを体系的に実施します。TDD原則に従い、テストを先に書いてから実装を進めます。

## テストケース分類

### A. ビットセット操作テスト
### B. QueryBuilder機能テスト  
### C. アーキタイプシステムテスト
### D. クエリキャッシュテスト
### E. パフォーマンステスト
### F. 並列実行テスト

---

## A. ビットセット操作テスト

### A-001: ComponentBitSet基本操作

**目的**: ビットセットの基本操作が正常に動作することを確認

```go
func TestComponentBitSet_BasicOperations(t *testing.T)
```

**テストケース**:
- [ ] **A-001-01**: 新しいビットセットは初期状態で全ビットが0
- [ ] **A-001-02**: Set操作で指定ビットが1になる
- [ ] **A-001-03**: Clear操作で指定ビットが0になる  
- [ ] **A-001-04**: Has操作で正しいビット状態を返す
- [ ] **A-001-05**: 同じビット位置に複数回Set/Clearしても正常動作

**テストデータ**:
```go
componentTypes := []ComponentType{
    TransformComponent, // bit 0
    SpriteComponent,    // bit 1
    PhysicsComponent,   // bit 2
    HealthComponent,    // bit 3
}
```

### A-002: ビットセット論理演算

**目的**: AND/OR演算の正確性を確認

```go
func TestComponentBitSet_LogicalOperations(t *testing.T)
```

**テストケース**:
- [ ] **A-002-01**: AND演算で共通ビットのみ1になる
- [ ] **A-002-02**: OR演算でいずれかが1なら1になる
- [ ] **A-002-03**: 空ビットセット同士のAND/ORは空
- [ ] **A-002-04**: フルビットセットとのANDは元のビットセット
- [ ] **A-002-05**: 複雑なビットパターンでの演算精度

**検証データ例**:
```go
testCases := []struct {
    name     string
    bitsetA  ComponentBitSet
    bitsetB  ComponentBitSet
    expectedAnd ComponentBitSet
    expectedOr  ComponentBitSet
}{
    {"Transform+Sprite AND Physics+Health", 0b0011, 0b1100, 0b0000, 0b1111},
    {"Transform+Physics AND Transform+Sprite", 0b0101, 0b0011, 0b0001, 0b0111},
}
```

### A-003: ビットセット境界値テスト

**目的**: 64ビット制限近辺での動作確認

```go
func TestComponentBitSet_BoundaryValues(t *testing.T)
```

**テストケース**:
- [ ] **A-003-01**: 0番目のビット操作
- [ ] **A-003-02**: 63番目のビット操作（最大値）
- [ ] **A-003-03**: 64番目のビット操作でエラー処理
- [ ] **A-003-04**: 全ビット設定時の動作
- [ ] **A-003-05**: オーバーフロー防止確認

---

## B. QueryBuilder機能テスト

### B-001: 基本クエリ構築

**目的**: シンプルなクエリが正常に構築できることを確認

```go
func TestQueryBuilder_BasicQueries(t *testing.T)
```

**テストケース**:
- [ ] **B-001-01**: With単一コンポーネント
- [ ] **B-001-02**: With複数コンポーネント  
- [ ] **B-001-03**: Without単一コンポーネント
- [ ] **B-001-04**: Without複数コンポーネント
- [ ] **B-001-05**: WithとWithoutの組み合わせ

**期待されるクエリ例**:
```go
// Transform + Sprite を持つエンティティ
query1 := builder.With(TransformComponent, SpriteComponent).Build()

// Physics を持たないエンティティ  
query2 := builder.Without(PhysicsComponent).Build()

// Transform + Sprite を持ち、Disabled を持たない
query3 := builder.
    With(TransformComponent, SpriteComponent).
    Without(DisabledComponent).
    Build()
```

### B-002: 複雑クエリ構築

**目的**: OR/AND論理演算子を含む複雑なクエリ構築

```go
func TestQueryBuilder_ComplexQueries(t *testing.T)
```

**テストケース**:
- [ ] **B-002-01**: OR条件の基本動作
- [ ] **B-002-02**: AND条件の基本動作
- [ ] **B-002-03**: OR/AND混合クエリ
- [ ] **B-002-04**: ネストした条件式
- [ ] **B-002-05**: 優先順位の確認

**複雑クエリ例**:
```go
// (Transform + Sprite) OR (Transform + Physics)
query := builder.
    With(TransformComponent, SpriteComponent).
    Or().
    With(TransformComponent, PhysicsComponent).
    Build()
```

### B-003: クエリバリデーション

**目的**: 不正なクエリ構築時のエラーハンドリング

```go
func TestQueryBuilder_Validation(t *testing.T)
```

**テストケース**:
- [ ] **B-003-01**: 空のクエリ構築時のエラー
- [ ] **B-003-02**: 存在しないComponentType指定時のエラー
- [ ] **B-003-03**: 矛盾する条件（WithとWithoutで同じコンポーネント）
- [ ] **B-003-04**: 不正な論理演算子の組み合わせ
- [ ] **B-003-05**: Build前のExecute呼び出しエラー

---

## C. アーキタイプシステムテスト

### C-001: アーキタイプ生成・管理

**目的**: アーキタイプの自動生成と管理機能

```go
func TestArchetypeManager_Creation(t *testing.T)
```

**テストケース**:
- [ ] **C-001-01**: 新しいシグネチャでアーキタイプ自動生成
- [ ] **C-001-02**: 既存アーキタイプの再利用
- [ ] **C-001-03**: アーキタイプの一意性保証
- [ ] **C-001-04**: 最大アーキタイプ数制限
- [ ] **C-001-05**: アーキタイプ削除（エンティティ0個時）

### C-002: エンティティアーキタイプ移動

**目的**: コンポーネント追加/削除時のアーキタイプ移動

```go
func TestArchetypeManager_EntityMovement(t *testing.T)
```

**テストケース**:
- [ ] **C-002-01**: コンポーネント追加時の移動
- [ ] **C-002-02**: コンポーネント削除時の移動
- [ ] **C-002-03**: 複数コンポーネント同時変更時の移動
- [ ] **C-002-04**: 移動後のデータ整合性確認
- [ ] **C-002-05**: 移動失敗時のロールバック

**テストシナリオ**:
```go
// Transform のみ → Transform + Sprite
entity := world.CreateEntity()
world.AddComponent(entity, &TransformComponent{})
// この時点でアーキタイプ移動が発生
world.AddComponent(entity, &SpriteComponent{})
```

### C-003: アーキタイプ検索

**目的**: クエリ条件にマッチするアーキタイプの効率的な検索

```go
func TestArchetypeManager_Matching(t *testing.T)
```

**テストケース**:
- [ ] **C-003-01**: 単一コンポーネント条件でのマッチング
- [ ] **C-003-02**: 複数コンポーネント条件でのマッチング
- [ ] **C-003-03**: Without条件を含むマッチング
- [ ] **C-003-04**: 複雑なOR/AND条件でのマッチング
- [ ] **C-003-05**: マッチしないケースの確認

---

## D. クエリキャッシュテスト

### D-001: キャッシュ基本動作

**目的**: クエリ結果の正確なキャッシュと取得

```go
func TestQueryCache_BasicOperations(t *testing.T)
```

**テストケース**:
- [ ] **D-001-01**: 初回クエリ実行時のキャッシュ作成
- [ ] **D-001-02**: 二回目実行時のキャッシュヒット
- [ ] **D-001-03**: 異なるクエリでのキャッシュ分離
- [ ] **D-001-04**: キャッシュミス時の実行・保存
- [ ] **D-001-05**: キャッシュヒット率の測定

### D-002: キャッシュ無効化

**目的**: エンティティ・コンポーネント変更時の適切なキャッシュ無効化

```go
func TestQueryCache_Invalidation(t *testing.T)
```

**テストケース**:
- [ ] **D-002-01**: エンティティ作成時のキャッシュ無効化
- [ ] **D-002-02**: エンティティ削除時のキャッシュ無効化
- [ ] **D-002-03**: コンポーネント追加時の部分無効化
- [ ] **D-002-04**: コンポーネント削除時の部分無効化
- [ ] **D-002-05**: 大量変更時のバッチ無効化

### D-003: キャッシュポリシー

**目的**: LRU/TTLキャッシュポリシーの動作確認

```go
func TestQueryCache_Policies(t *testing.T)
```

**テストケース**:
- [ ] **D-003-01**: LRU最古エントリの削除
- [ ] **D-003-02**: TTL期限切れエントリの削除
- [ ] **D-003-03**: メモリ制限時の削除優先順位
- [ ] **D-003-04**: キャッシュサイズ制限の動作
- [ ] **D-003-05**: 手動キャッシュクリアの動作

---

## E. パフォーマンステスト

### E-001: クエリ実行速度

**目的**: NFR-002のパフォーマンス要件達成確認

```go
func BenchmarkQueryEngine_ExecutionSpeed(b *testing.B)
```

**テスト条件**:
- **エンティティ数**: 1,000 / 10,000 / 100,000
- **コンポーネント数**: 5 / 15 / 64
- **アーキタイプ数**: 10 / 100 / 1,000

**成功基準**:
- [ ] **E-001-01**: 10,000エンティティクエリ < 1ms
- [ ] **E-001-02**: 100,000エンティティクエリ < 10ms
- [ ] **E-001-03**: 複雑クエリでも性能劣化 < 2倍
- [ ] **E-001-04**: キャッシュヒット時 < 0.1ms
- [ ] **E-001-05**: 並列クエリでのスケーリング効率

### E-002: メモリ使用量

**目的**: メモリ効率要件の確認

```go
func TestQueryEngine_MemoryUsage(t *testing.T)
```

**テストケース**:
- [ ] **E-002-01**: エンティティ1つあたり8バイト以内
- [ ] **E-002-02**: アーキタイプ作成時のメモリ増加量
- [ ] **E-002-03**: キャッシュメモリ使用量
- [ ] **E-002-04**: 24時間実行でのメモリリーク検出
- [ ] **E-002-05**: GCプレッシャーの測定

### E-003: スケーラビリティ

**目的**: 大規模データでの性能確認

```go
func TestQueryEngine_Scalability(t *testing.T)
```

**テストケース**:
- [ ] **E-003-01**: 100,000エンティティでの安定動作
- [ ] **E-003-02**: 1,000アーキタイプでの検索効率
- [ ] **E-003-03**: 100同時クエリでの処理能力
- [ ] **E-003-04**: データ増加に対する線形スケーリング
- [ ] **E-003-05**: メモリ不足時の graceful degradation

---

## F. 並列実行テスト

### F-001: スレッドセーフティ

**目的**: 並列アクセス時のデータ整合性確保

```go
func TestQueryEngine_ThreadSafety(t *testing.T)
```

**テストケース**:
- [ ] **F-001-01**: 同時クエリ実行での競合状態なし
- [ ] **F-001-02**: クエリ実行中のエンティティ変更安全性
- [ ] **F-001-03**: キャッシュアクセスでの競合状態なし
- [ ] **F-001-04**: アーキタイプ作成での競合状態なし
- [ ] **F-001-05**: デッドロック検出（10分間ストレステスト）

### F-002: 並列パフォーマンス

**目的**: 並列実行による性能向上確認

```go
func BenchmarkQueryEngine_ParallelPerformance(b *testing.B)
```

**テストケース**:
- [ ] **F-002-01**: 2スレッド並列で1.5倍以上の高速化
- [ ] **F-002-02**: 4スレッド並列で2.5倍以上の高速化
- [ ] **F-002-03**: CPUコア数以上での性能飽和確認
- [ ] **F-002-04**: 並列オーバーヘッド < 20%
- [ ] **F-002-05**: 適切なワーカープール管理

---

## テスト環境・設定

### テストデータセットアップ

```go
type QueryEngineTestSuite struct {
    world           World
    queryEngine     QueryEngine
    testEntities    []EntityID
    componentTypes  []ComponentType
}

func (s *QueryEngineTestSuite) SetupTest() {
    // 標準テストデータセット作成
    s.createTestEntities(1000)
    s.createVariousArchetypes(10)
    s.setupPerformanceData()
}
```

### テスト実行順序

1. **単体テスト** (A, B, C, D部分)
2. **統合テスト** (コンポーネント間連携)
3. **パフォーマンステスト** (E部分) 
4. **並列テスト** (F部分)

### モック・スタブ

```go
// テスト用の軽量World実装
type MockWorld struct {
    entities    map[EntityID]*MockEntity
    components  map[EntityID]map[ComponentType]Component
}

// パフォーマンステスト用の高速データ生成器
type TestDataGenerator struct {
    seed           int64
    entityCount    int
    archetypeCount int
}
```

---

## テスト品質管理

### カバレッジ目標
- **コードカバレッジ**: 100%
- **ブランチカバレッジ**: 95%以上
- **条件カバレッジ**: 90%以上

### 成功基準
- [ ] 全テストケースが成功（320個）
- [ ] パフォーマンステスト基準達成
- [ ] 並列テストでの競合状態ゼロ
- [ ] 24時間安定性テスト成功

### CI/CD統合
```yaml
# GitHub Actions example
- name: Run QueryEngine Tests
  run: |
    go test ./internal/core/ecs/query/... -v -race -coverprofile=coverage.out
    go test ./internal/core/ecs/query/... -bench=. -benchmem
    go test ./internal/core/ecs/query/... -run=TestQueryEngine_24HourStability
```

**次ステップ**: このテストケース仕様に基づき、失敗するテストから実装を開始します。