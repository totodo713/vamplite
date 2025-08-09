# TASK-301: ModECSAPI実装 - Green段階（最小実装）

## 実装概要

TDDのGreen段階として、Red段階で作成した失敗するテストを通すための最小限の実装を行います。過度な実装は避け、テストが通る最低限の機能のみを実装します。

## Green段階実装内容

### 1. ModECSAPI基本実装

実装先: `internal/core/ecs/mod/mod_api.go`

```go
package mod

import (
	"strings"
	"time"

	"muscle-dreamer/internal/core/ecs"
)

// ModECSAPIImpl はModECSAPIの基本実装
type ModECSAPIImpl struct {
	modID      string
	context    *ModContext
	entityAPI  ModEntityAPI
	componentAPI ModComponentAPI
	queryAPI   ModQueryAPI
	systemAPI  ModSystemAPI
}

// NewModECSAPI 新しいModECSAPIインスタンスを作成
func NewModECSAPI(modID string, config ModConfig) ModECSAPI {
	ctx := &ModContext{
		ModID:             modID,
		MaxEntities:       config.MaxEntities,
		MaxMemory:         config.MaxMemory,
		MaxExecutionTime:  config.MaxExecutionTime,
		AllowedComponents: config.AllowedComponents,
		CreatedEntities:   make([]ecs.EntityID, 0),
		MemoryUsage:       0,
		ExecutionTime:     0,
		QueryCount:        0,
		MaxQueryCount:     config.MaxQueryCount,
	}

	impl := &ModECSAPIImpl{
		modID:   modID,
		context: ctx,
	}

	impl.entityAPI = &ModEntityAPIImpl{api: impl}
	impl.componentAPI = &ModComponentAPIImpl{api: impl}
	impl.queryAPI = &ModQueryAPIImpl{api: impl}
	impl.systemAPI = &ModSystemAPIImpl{api: impl}

	return impl
}

func (m *ModECSAPIImpl) Entities() ModEntityAPI {
	return m.entityAPI
}

func (m *ModECSAPIImpl) Components() ModComponentAPI {
	return m.componentAPI
}

func (m *ModECSAPIImpl) Queries() ModQueryAPI {
	return m.queryAPI
}

func (m *ModECSAPIImpl) Systems() ModSystemAPI {
	return m.systemAPI
}

func (m *ModECSAPIImpl) GetContext() *ModContext {
	return m.context
}

// ModEntityAPIImpl はModEntityAPIの基本実装
type ModEntityAPIImpl struct {
	api        *ModECSAPIImpl
	nextID     ecs.EntityID
	entities   map[ecs.EntityID][]string  // エンティティIDとタグのマップ
}

func (m *ModEntityAPIImpl) Create(tags ...string) (ecs.EntityID, error) {
	if len(m.api.context.CreatedEntities) >= m.api.context.MaxEntities {
		return ecs.InvalidEntityID, ErrEntityLimitExceeded
	}

	// 最小実装：単純にインクリメンタルIDを割り当て
	m.nextID++
	entityID := m.nextID

	// MODタグを自動追加
	allTags := append([]string{"mod:" + m.api.modID}, tags...)
	
	if m.entities == nil {
		m.entities = make(map[ecs.EntityID][]string)
	}
	m.entities[entityID] = allTags

	m.api.context.CreatedEntities = append(m.api.context.CreatedEntities, entityID)
	m.api.context.MemoryUsage += 64 // 最小メモリ使用量を仮定

	return entityID, nil
}

func (m *ModEntityAPIImpl) Delete(id ecs.EntityID) error {
	// 権限チェック：自分が作成したエンティティのみ削除可能
	if !m.isOwnedEntity(id) {
		// システムエンティティかどうか簡易判定
		if id < 1000 { // 1000未満をシステムエンティティと仮定
			return ErrSystemEntityAccess
		}
		return ErrEntityPermissionDenied
	}

	// エンティティ削除（最小実装）
	delete(m.entities, id)
	
	// CreatedEntitiesから削除
	for i, entityID := range m.api.context.CreatedEntities {
		if entityID == id {
			m.api.context.CreatedEntities = append(
				m.api.context.CreatedEntities[:i],
				m.api.context.CreatedEntities[i+1:]...)
			break
		}
	}

	return nil
}

func (m *ModEntityAPIImpl) GetTags(id ecs.EntityID) ([]string, error) {
	if tags, exists := m.entities[id]; exists {
		return tags, nil
	}
	return nil, ErrEntityPermissionDenied
}

func (m *ModEntityAPIImpl) GetOwned() ([]ecs.EntityID, error) {
	return m.api.context.CreatedEntities, nil
}

func (m *ModEntityAPIImpl) isOwnedEntity(id ecs.EntityID) bool {
	for _, entityID := range m.api.context.CreatedEntities {
		if entityID == id {
			return true
		}
	}
	return false
}

// ModComponentAPIImpl はModComponentAPIの基本実装
type ModComponentAPIImpl struct {
	api        *ModECSAPIImpl
	components map[ecs.EntityID]map[ecs.ComponentType]ecs.Component
}

func (m *ModComponentAPIImpl) Add(entity ecs.EntityID, component ecs.Component) error {
	// 権限チェック
	entityAPI := m.api.entityAPI.(*ModEntityAPIImpl)
	if !entityAPI.isOwnedEntity(entity) {
		return ErrComponentPermissionDenied
	}

	// コンポーネント型許可チェック
	if !m.IsAllowed(component.GetType()) {
		return ErrComponentNotAllowed
	}

	// 最小実装：メモリマップに保存
	if m.components == nil {
		m.components = make(map[ecs.EntityID]map[ecs.ComponentType]ecs.Component)
	}
	if m.components[entity] == nil {
		m.components[entity] = make(map[ecs.ComponentType]ecs.Component)
	}

	m.components[entity][component.GetType()] = component
	return nil
}

func (m *ModComponentAPIImpl) Get(entity ecs.EntityID, componentType ecs.ComponentType) (ecs.Component, error) {
	// 権限チェック
	entityAPI := m.api.entityAPI.(*ModEntityAPIImpl)
	if !entityAPI.isOwnedEntity(entity) {
		return nil, ErrComponentPermissionDenied
	}

	if m.components == nil {
		return nil, nil
	}

	if entityComponents, exists := m.components[entity]; exists {
		if component, exists := entityComponents[componentType]; exists {
			return component, nil
		}
	}

	return nil, nil
}

func (m *ModComponentAPIImpl) Remove(entity ecs.EntityID, componentType ecs.ComponentType) error {
	// 権限チェック
	entityAPI := m.api.entityAPI.(*ModEntityAPIImpl)
	if !entityAPI.isOwnedEntity(entity) {
		return ErrComponentPermissionDenied
	}

	if m.components != nil && m.components[entity] != nil {
		delete(m.components[entity], componentType)
	}

	return nil
}

func (m *ModComponentAPIImpl) IsAllowed(componentType ecs.ComponentType) bool {
	for _, allowed := range m.api.context.AllowedComponents {
		if componentType == allowed {
			return true
		}
	}
	return false
}

// ModQueryAPIImpl はModQueryAPIの基本実装
type ModQueryAPIImpl struct {
	api *ModECSAPIImpl
}

func (m *ModQueryAPIImpl) Find(query ecs.QueryBuilder) ([]ecs.EntityID, error) {
	// クエリ実行回数制限チェック
	if m.api.context.QueryCount >= m.api.context.MaxQueryCount {
		return nil, ErrQueryLimitExceeded
	}

	m.api.context.QueryCount++

	// 最小実装：自分が作成したエンティティのみ返却
	return m.api.context.CreatedEntities, nil
}

func (m *ModQueryAPIImpl) Count(query ecs.QueryBuilder) (int, error) {
	entities, err := m.Find(query)
	if err != nil {
		return 0, err
	}
	return len(entities), nil
}

func (m *ModQueryAPIImpl) GetExecutionCount() int {
	return m.api.context.QueryCount
}

func (m *ModQueryAPIImpl) ResetExecutionCount() {
	m.api.context.QueryCount = 0
}

// ModSystemAPIImpl はModSystemAPIの基本実装
type ModSystemAPIImpl struct {
	api       *ModECSAPIImpl
	systems   map[string]ModSystem
}

func (m *ModSystemAPIImpl) Register(system ModSystem) error {
	// 実行時間制限チェック
	if system.GetMaxExecutionTime() > m.api.context.MaxExecutionTime {
		return ErrSystemExecutionTimeExceeded
	}

	// セキュリティチェック（基本的な文字列チェック）
	if strings.Contains(system.GetID(), "rm -rf") ||
		strings.Contains(system.GetID(), "../../") {
		return &SecurityError{
			ModID:     m.api.modID,
			Operation: "system_register",
			Reason:    "malicious command detected",
		}
	}

	if m.systems == nil {
		m.systems = make(map[string]ModSystem)
	}

	m.systems[system.GetID()] = system
	return nil
}

func (m *ModSystemAPIImpl) Unregister(systemID string) error {
	if m.systems != nil {
		delete(m.systems, systemID)
	}
	return nil
}

func (m *ModSystemAPIImpl) GetRegistered() []string {
	if m.systems == nil {
		return []string{}
	}

	result := make([]string, 0, len(m.systems))
	for id := range m.systems {
		result = append(result, id)
	}
	return result
}
```

### 2. ファクトリー実装

実装先: `internal/core/ecs/mod/factory.go`

```go
package mod

// ModECSAPIFactoryImpl はModECSAPIFactoryの基本実装
type ModECSAPIFactoryImpl struct {
	apis map[string]ModECSAPI
}

// NewModECSAPIFactory 新しいファクトリーを作成
func NewModECSAPIFactory() ModECSAPIFactory {
	return &ModECSAPIFactoryImpl{
		apis: make(map[string]ModECSAPI),
	}
}

func (f *ModECSAPIFactoryImpl) Create(modID string, config ModConfig) (ModECSAPI, error) {
	if _, exists := f.apis[modID]; exists {
		return nil, ErrModAlreadyExists
	}

	api := NewModECSAPI(modID, config)
	f.apis[modID] = api
	return api, nil
}

func (f *ModECSAPIFactoryImpl) Destroy(modID string) error {
	if _, exists := f.apis[modID]; !exists {
		return ErrModNotFound
	}

	delete(f.apis, modID)
	return nil
}
```

### 3. テストヘルパー実装

実装先: `internal/core/ecs/mod/test_helpers.go`

```go
package mod

import (
	"testing"
	"time"

	"muscle-dreamer/internal/core/ecs"
)

// TestModSystemImpl テスト用MODシステム実装
type TestModSystemImpl struct {
	id              string
	maxExecutionTime time.Duration
}

func (t *TestModSystemImpl) GetID() string {
	return t.id
}

func (t *TestModSystemImpl) Update(ctx *ModContext, deltaTime time.Duration) error {
	// 最小実装：何もしない
	return nil
}

func (t *TestModSystemImpl) GetMaxExecutionTime() time.Duration {
	return t.maxExecutionTime
}

// TestSpriteComponentImpl テスト用Spriteコンポーネント実装
type TestSpriteComponentImpl struct{}

func (t *TestSpriteComponentImpl) GetType() ecs.ComponentType {
	return ecs.ComponentTypeSprite
}

func (t *TestSpriteComponentImpl) Clone() ecs.Component {
	return &TestSpriteComponentImpl{}
}

func (t *TestSpriteComponentImpl) Validate() error {
	return nil
}

func (t *TestSpriteComponentImpl) Size() int {
	return 64
}

func (t *TestSpriteComponentImpl) Serialize() ([]byte, error) {
	return []byte{}, nil
}

func (t *TestSpriteComponentImpl) Deserialize([]byte) error {
	return nil
}

// TestFileIOComponentImpl テスト用禁止コンポーネント実装
type TestFileIOComponentImpl struct{}

func (t *TestFileIOComponentImpl) GetType() ecs.ComponentType {
	return ecs.ComponentType(999) // 未定義の型番号
}

func (t *TestFileIOComponentImpl) Clone() ecs.Component {
	return &TestFileIOComponentImpl{}
}

func (t *TestFileIOComponentImpl) Validate() error {
	return nil
}

func (t *TestFileIOComponentImpl) Size() int {
	return 64
}

func (t *TestFileIOComponentImpl) Serialize() ([]byte, error) {
	return []byte{}, nil
}

func (t *TestFileIOComponentImpl) Deserialize([]byte) error {
	return nil
}

// TestQueryBuilderImpl テスト用QueryBuilder実装
type TestQueryBuilderImpl struct{}

func (t *TestQueryBuilderImpl) With(ecs.ComponentType) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Without(ecs.ComponentType) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) WithAll([]ecs.ComponentType) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) WithAny([]ecs.ComponentType) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) WithNone([]ecs.ComponentType) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Where(func(ecs.EntityID, []ecs.Component) bool) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) WhereComponent(ecs.ComponentType, func(ecs.Component) bool) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) WhereEntity(func(ecs.EntityID) bool) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Limit(int) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Offset(int) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) OrderBy(func(ecs.EntityID, ecs.EntityID) bool) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) OrderByComponent(ecs.ComponentType, func(ecs.Component, ecs.Component) bool) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Cache(string) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) CacheFor(time.Duration) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) UseBitset(bool) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) UseIndex(string) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) WithinRadius(ecs.Vector2, float64) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) WithinBounds(ecs.AABB) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Intersects(ecs.AABB) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Nearest(ecs.Vector2, int) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Children(ecs.EntityID) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Descendants(ecs.EntityID) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Ancestors(ecs.EntityID) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Siblings(ecs.EntityID) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) CreatedAfter(time.Time) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) ModifiedSince(time.Time) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) OlderThan(time.Duration) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) InTimeRange(time.Time, time.Time) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) GroupBy(ecs.ComponentType) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Aggregate(func([]ecs.Component) interface{}) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Count() ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Distinct(ecs.ComponentType) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Execute() ecs.QueryResult { return nil }
func (t *TestQueryBuilderImpl) ExecuteAsync() <-chan ecs.QueryResult { return nil }
func (t *TestQueryBuilderImpl) Stream() <-chan ecs.EntityID { return nil }
func (t *TestQueryBuilderImpl) ExecuteWithCallback(func(ecs.EntityID, []ecs.Component)) error { return nil }
func (t *TestQueryBuilderImpl) ToString() string { return "" }
func (t *TestQueryBuilderImpl) ToHash() string { return "" }
func (t *TestQueryBuilderImpl) GetSignature() string { return "" }
func (t *TestQueryBuilderImpl) Clone() ecs.QueryBuilder { return &TestQueryBuilderImpl{} }

// テストヘルパー関数

var globalTestFactory ModECSAPIFactory

func init() {
	globalTestFactory = NewModECSAPIFactory()
}

func createTestModAPI(t testing.TB, modID string) ModECSAPI {
	config := DefaultModConfig()
	api, err := globalTestFactory.Create(modID, config)
	if err != nil {
		t.Fatal(err)
	}
	return api
}

func createEntityWithMod(t testing.TB, modID string) ecs.EntityID {
	// 他MODのエンティティを模擬作成
	// 最小実装：単純に別のMOD APIを作成してエンティティ作成
	config := DefaultModConfig()
	otherAPI, err := globalTestFactory.Create(modID+"-temp", config)
	if err != nil {
		t.Fatal(err)
	}
	
	entityID, err := otherAPI.Entities().Create("other-mod-entity")
	if err != nil {
		t.Fatal(err)
	}
	
	return entityID
}

func createSystemEntity(t testing.TB) ecs.EntityID {
	// システムエンティティを模擬作成
	// 最小実装：1000未満のIDをシステムエンティティとして扱う
	return ecs.EntityID(100) // システムエンティティID
}

func createTestSpriteComponent() ecs.Component {
	return &TestSpriteComponentImpl{}
}

func createTestFileIOComponent() ecs.Component {
	return &TestFileIOComponentImpl{}
}

func createTestQuery() ecs.QueryBuilder {
	return &TestQueryBuilderImpl{}
}

func createTestModSystem(id string) ModSystem {
	return &TestModSystemImpl{
		id:              id,
		maxExecutionTime: 3 * time.Millisecond, // 制限内
	}
}

func createLongRunningModSystem(id string, duration time.Duration) ModSystem {
	return &TestModSystemImpl{
		id:              id,
		maxExecutionTime: duration,
	}
}

func createMaliciousSystem(command string) ModSystem {
	return &TestModSystemImpl{
		id:              command, // 悪意のあるコマンドをIDに設定
		maxExecutionTime: 1 * time.Millisecond,
	}
}
```

## Green段階実行手順

### 1. 実装ファイル作成
```bash
# 基本実装
touch internal/core/ecs/mod/mod_api.go

# ファクトリー実装  
touch internal/core/ecs/mod/factory.go

# テストヘルパー実装
touch internal/core/ecs/mod/test_helpers.go
```

### 2. テスト実行（成功確認）
```bash
# 単体テスト実行（成功することを確認）
go test ./internal/core/ecs/mod/ -v

# カバレッジ確認
go test ./internal/core/ecs/mod/ -cover
```

### 3. 実装検証ポイント

#### 機能検証
- [ ] エンティティ作成・削除の基本動作
- [ ] 作成上限（100個）の制限動作
- [ ] MODタグ自動付与機能
- [ ] 権限チェック（自分のエンティティのみアクセス）
- [ ] コンポーネント操作の基本動作
- [ ] 許可コンポーネント型のホワイトリスト動作
- [ ] クエリ実行・回数制限動作
- [ ] システム登録・実行時間制限動作

#### セキュリティ検証
- [ ] パストラバーサル攻撃のタグ受け入れ（安全）
- [ ] 悪意のあるシステム登録の拒否
- [ ] システムエンティティアクセス拒否
- [ ] 他MODエンティティアクセス拒否

#### エラーハンドリング検証
- [ ] 適切なエラー型の返却
- [ ] エラーメッセージの内容確認
- [ ] セキュリティエラーの詳細情報

## Green段階の成功基準

- [ ] 全ての単体テストが通過する
- [ ] 全ての機能の基本動作が確認される
- [ ] セキュリティ制約が基本的に動作する
- [ ] エラーハンドリングが適切に機能する
- [ ] コードカバレッジ>80%達成

## 実装時の注意事項

### 最小実装原則
- **過度な最適化を避ける**: パフォーマンス最適化はRefactor段階で実施
- **機能は最低限**: テストが通る最小限の機能のみ実装
- **エラーハンドリング優先**: セキュリティ関連エラーは確実に実装

### セキュリティ実装
- **基本的なチェックのみ**: 文字列パターンマッチング程度
- **権限分離**: MOD間・MOD-システム間の基本的な分離
- **リソース制限**: 基本的なカウンタ実装

## 次段階への準備

Green段階完了後、Refactor段階で以下を改善します：
1. パフォーマンス最適化（メモリ効率、実行速度）
2. セキュリティ強化（より高度な脅威検出）
3. エラーハンドリング詳細化
4. コード品質向上（設計パターン適用）

---

**作成日時**: 2025-08-08  
**段階**: TDD Green  
**期待結果**: 全テスト成功 ✅