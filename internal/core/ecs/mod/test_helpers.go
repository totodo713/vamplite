package mod

import (
	"fmt"
	"testing"
	"time"

	"muscle-dreamer/internal/core/ecs"
)

// TestModSystemImpl テスト用MODシステム実装
type TestModSystemImpl struct {
	id               string
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
	return ecs.ComponentType("test-fileio") // 未定義の型番号
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

func (t *TestQueryBuilderImpl) With(ecs.ComponentType) ecs.QueryBuilder       { return t }
func (t *TestQueryBuilderImpl) Without(ecs.ComponentType) ecs.QueryBuilder    { return t }
func (t *TestQueryBuilderImpl) WithAll([]ecs.ComponentType) ecs.QueryBuilder  { return t }
func (t *TestQueryBuilderImpl) WithAny([]ecs.ComponentType) ecs.QueryBuilder  { return t }
func (t *TestQueryBuilderImpl) WithNone([]ecs.ComponentType) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Where(func(ecs.EntityID, []ecs.Component) bool) ecs.QueryBuilder {
	return t
}

func (t *TestQueryBuilderImpl) WhereComponent(ecs.ComponentType, func(ecs.Component) bool) ecs.QueryBuilder {
	return t
}
func (t *TestQueryBuilderImpl) WhereEntity(func(ecs.EntityID) bool) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Limit(int) ecs.QueryBuilder                           { return t }
func (t *TestQueryBuilderImpl) Offset(int) ecs.QueryBuilder                          { return t }
func (t *TestQueryBuilderImpl) OrderBy(func(ecs.EntityID, ecs.EntityID) bool) ecs.QueryBuilder {
	return t
}

func (t *TestQueryBuilderImpl) OrderByComponent(ecs.ComponentType, func(ecs.Component, ecs.Component) bool) ecs.QueryBuilder {
	return t
}
func (t *TestQueryBuilderImpl) Cache(string) ecs.QueryBuilder                      { return t }
func (t *TestQueryBuilderImpl) CacheFor(time.Duration) ecs.QueryBuilder            { return t }
func (t *TestQueryBuilderImpl) UseBitset(bool) ecs.QueryBuilder                    { return t }
func (t *TestQueryBuilderImpl) UseIndex(string) ecs.QueryBuilder                   { return t }
func (t *TestQueryBuilderImpl) WithinRadius(ecs.Vector2, float64) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) WithinBounds(ecs.AABB) ecs.QueryBuilder             { return t }
func (t *TestQueryBuilderImpl) Intersects(ecs.AABB) ecs.QueryBuilder               { return t }
func (t *TestQueryBuilderImpl) Nearest(ecs.Vector2, int) ecs.QueryBuilder          { return t }
func (t *TestQueryBuilderImpl) Children(ecs.EntityID) ecs.QueryBuilder             { return t }
func (t *TestQueryBuilderImpl) Descendants(ecs.EntityID) ecs.QueryBuilder          { return t }
func (t *TestQueryBuilderImpl) Ancestors(ecs.EntityID) ecs.QueryBuilder            { return t }
func (t *TestQueryBuilderImpl) Siblings(ecs.EntityID) ecs.QueryBuilder             { return t }
func (t *TestQueryBuilderImpl) CreatedAfter(time.Time) ecs.QueryBuilder            { return t }
func (t *TestQueryBuilderImpl) ModifiedSince(time.Time) ecs.QueryBuilder           { return t }
func (t *TestQueryBuilderImpl) OlderThan(time.Duration) ecs.QueryBuilder           { return t }
func (t *TestQueryBuilderImpl) InTimeRange(time.Time, time.Time) ecs.QueryBuilder  { return t }
func (t *TestQueryBuilderImpl) GroupBy(ecs.ComponentType) ecs.QueryBuilder         { return t }
func (t *TestQueryBuilderImpl) Aggregate(func([]ecs.Component) interface{}) ecs.QueryBuilder {
	return t
}
func (t *TestQueryBuilderImpl) Count() ecs.QueryBuilder                     { return t }
func (t *TestQueryBuilderImpl) Distinct(ecs.ComponentType) ecs.QueryBuilder { return t }
func (t *TestQueryBuilderImpl) Execute() ecs.QueryResult                    { return nil }
func (t *TestQueryBuilderImpl) ExecuteAsync() <-chan ecs.QueryResult        { return nil }
func (t *TestQueryBuilderImpl) Stream() <-chan ecs.EntityID                 { return nil }
func (t *TestQueryBuilderImpl) ExecuteWithCallback(func(ecs.EntityID, []ecs.Component)) error {
	return nil
}
func (t *TestQueryBuilderImpl) ToString() string        { return "" }
func (t *TestQueryBuilderImpl) ToHash() string          { return "" }
func (t *TestQueryBuilderImpl) GetSignature() string    { return "" }
func (t *TestQueryBuilderImpl) Clone() ecs.QueryBuilder { return &TestQueryBuilderImpl{} }

// テストヘルパー関数

var (
	globalTestFactory ModECSAPIFactory
	testCounter       int
)

func init() {
	globalTestFactory = NewModECSAPIFactory()
}

func createTestModAPI(t testing.TB, modID string) ModECSAPI {
	testCounter++
	uniqueModID := fmt.Sprintf("%s-%d", modID, testCounter)
	config := DefaultModConfig()
	api, err := globalTestFactory.Create(uniqueModID, config)
	if err != nil {
		t.Fatal(err)
	}
	return api
}

func createEntityWithMod(t testing.TB, modID string) ecs.EntityID {
	// 他MODのエンティティを模擬作成
	// 最小実装：大きなID（1000以上）を返してシステムエンティティと区別
	testCounter++
	return ecs.EntityID(1000 + testCounter) // 他MODのエンティティID
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
		id:               "safe-" + id,         // "system"パターンを避ける
		maxExecutionTime: 3 * time.Millisecond, // 制限内
	}
}

func createLongRunningModSystem(id string, duration time.Duration) ModSystem {
	return &TestModSystemImpl{
		id:               id,
		maxExecutionTime: duration,
	}
}

func createMaliciousSystem(command string) ModSystem {
	return &TestModSystemImpl{
		id:               command, // 悪意のあるコマンドをIDに設定
		maxExecutionTime: 1 * time.Millisecond,
	}
}
