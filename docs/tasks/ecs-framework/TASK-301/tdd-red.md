# TASK-301: ModECSAPI実装 - Red段階（失敗するテスト）

## 実装概要

TDDのRed段階として、ModECSAPIの機能要件を満たす失敗するテストを実装します。この段階では実装は存在せず、すべてのテストが失敗することを確認します。

## Red段階実装内容

### 1. ModECSAPIインターフェース定義

実装先: `internal/core/ecs/mod/interfaces.go`

```go
package mod

import (
    "time"
    "github.com/muscle-dreamer/internal/core/ecs"
)

// ModECSAPI はMOD向けの制限されたECS APIのメインインターフェース
type ModECSAPI interface {
    Entities() ModEntityAPI
    Components() ModComponentAPI
    Queries() ModQueryAPI
    Systems() ModSystemAPI
    GetContext() *ModContext
}

// ModEntityAPI はMOD向けの制限されたエンティティ操作API
type ModEntityAPI interface {
    Create(tags ...string) (ecs.EntityID, error)
    Delete(id ecs.EntityID) error
    GetTags(id ecs.EntityID) ([]string, error)
    GetOwned() ([]ecs.EntityID, error)
}

// ModComponentAPI はMOD向けの制限されたコンポーネント操作API
type ModComponentAPI interface {
    Add(entity ecs.EntityID, component ecs.Component) error
    Get(entity ecs.EntityID, componentType ecs.ComponentType) (ecs.Component, error)
    Remove(entity ecs.EntityID, componentType ecs.ComponentType) error
    IsAllowed(componentType ecs.ComponentType) bool
}

// ModQueryAPI はMOD向けの制限されたクエリ操作API
type ModQueryAPI interface {
    Find(query ecs.Query) ([]ecs.EntityID, error)
    Count(query ecs.Query) (int, error)
    GetExecutionCount() int
    ResetExecutionCount()
}

// ModSystemAPI はMOD向けの制限されたシステム操作API
type ModSystemAPI interface {
    Register(system ModSystem) error
    Unregister(systemID string) error
    GetRegistered() []string
}

// ModSystem はMOD向けシステムインターフェース
type ModSystem interface {
    GetID() string
    Update(ctx *ModContext, deltaTime time.Duration) error
    GetMaxExecutionTime() time.Duration
}

// ModContext はMOD実行コンテキスト
type ModContext struct {
    ModID             string
    MaxEntities       int
    MaxMemory         int64
    MaxExecutionTime  time.Duration
    AllowedComponents []ecs.ComponentType
    CreatedEntities   []ecs.EntityID
    MemoryUsage      int64
    ExecutionTime    time.Duration
    QueryCount       int
    MaxQueryCount    int
}

// ModECSAPIFactory はModECSAPIの作成ファクトリー
type ModECSAPIFactory interface {
    Create(modID string, config ModConfig) (ModECSAPI, error)
    Destroy(modID string) error
}

// ModConfig はMOD設定
type ModConfig struct {
    MaxEntities       int
    MaxMemory         int64
    MaxExecutionTime  time.Duration
    AllowedComponents []ecs.ComponentType
    MaxQueryCount     int
}

// DefaultModConfig デフォルトMOD設定
func DefaultModConfig() ModConfig {
    return ModConfig{
        MaxEntities:      100,
        MaxMemory:        10 * 1024 * 1024, // 10MB
        MaxExecutionTime: 5 * time.Millisecond,
        AllowedComponents: []ecs.ComponentType{
            ecs.ComponentTypeSprite,
            ecs.ComponentTypePhysics,
            ecs.ComponentTypeHealth,
            ecs.ComponentTypeAI,
            ecs.ComponentTypeInventory,
            ecs.ComponentTypeEnergy,
        },
        MaxQueryCount: 1000,
    }
}
```

### 2. 失敗するテスト実装

実装先: `internal/core/ecs/mod/mod_api_test.go`

```go
package mod

import (
    "testing"
    "time"
    "github.com/muscle-dreamer/internal/core/ecs"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestModEntityAPI_Create(t *testing.T) {
    // この段階では実装が存在しないため全てのテストが失敗する
    
    // セットアップ（まだ実装されていない）
    api := createTestModAPI(t, "test-mod")
    
    t.Run("正常なエンティティ作成", func(t *testing.T) {
        entityID, err := api.Entities().Create("test-tag")
        require.NoError(t, err)
        assert.NotEqual(t, ecs.InvalidEntityID, entityID)
        
        // MODタグが自動付与されることを確認
        tags, err := api.Entities().GetTags(entityID)
        require.NoError(t, err)
        assert.Contains(t, tags, "mod:test-mod")
        assert.Contains(t, tags, "test-tag")
    })
    
    t.Run("エンティティ作成上限テスト", func(t *testing.T) {
        // 100個まで作成成功
        for i := 0; i < 100; i++ {
            _, err := api.Entities().Create()
            require.NoError(t, err)
        }
        
        // 101個目は失敗
        _, err := api.Entities().Create()
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "entity limit exceeded")
    })
}

func TestModEntityAPI_Delete(t *testing.T) {
    api := createTestModAPI(t, "test-mod")
    
    t.Run("自分のエンティティ削除", func(t *testing.T) {
        entityID, err := api.Entities().Create("test-entity")
        require.NoError(t, err)
        
        err = api.Entities().Delete(entityID)
        assert.NoError(t, err)
    })
    
    t.Run("他MODエンティティ削除拒否", func(t *testing.T) {
        // 他MODのエンティティを模擬
        otherModEntity := createEntityWithMod(t, "other-mod")
        
        err := api.Entities().Delete(otherModEntity)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "permission denied")
    })
    
    t.Run("システムエンティティ削除拒否", func(t *testing.T) {
        // システムエンティティを模擬
        systemEntity := createSystemEntity(t)
        
        err := api.Entities().Delete(systemEntity)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "system entity access denied")
    })
}

func TestModComponentAPI_Add(t *testing.T) {
    api := createTestModAPI(t, "test-mod")
    
    t.Run("許可されたコンポーネント追加", func(t *testing.T) {
        entityID, err := api.Entities().Create()
        require.NoError(t, err)
        
        // 許可されたコンポーネント（Sprite）
        spriteComponent := createTestSpriteComponent()
        err = api.Components().Add(entityID, spriteComponent)
        assert.NoError(t, err)
    })
    
    t.Run("禁止コンポーネント追加拒否", func(t *testing.T) {
        entityID, err := api.Entities().Create()
        require.NoError(t, err)
        
        // 禁止されたコンポーネント（FileIO - 存在しないが模擬）
        fileIOComponent := createTestFileIOComponent()
        err = api.Components().Add(entityID, fileIOComponent)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "component not allowed")
    })
    
    t.Run("他MODエンティティへのコンポーネント追加拒否", func(t *testing.T) {
        otherModEntity := createEntityWithMod(t, "other-mod")
        spriteComponent := createTestSpriteComponent()
        
        err := api.Components().Add(otherModEntity, spriteComponent)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "permission denied")
    })
}

func TestModComponentAPI_Get(t *testing.T) {
    api := createTestModAPI(t, "test-mod")
    
    t.Run("自分のコンポーネント取得", func(t *testing.T) {
        entityID, err := api.Entities().Create()
        require.NoError(t, err)
        
        spriteComponent := createTestSpriteComponent()
        err = api.Components().Add(entityID, spriteComponent)
        require.NoError(t, err)
        
        retrieved, err := api.Components().Get(entityID, ecs.ComponentTypeSprite)
        assert.NoError(t, err)
        assert.NotNil(t, retrieved)
    })
    
    t.Run("権限のないコンポーネント取得拒否", func(t *testing.T) {
        otherModEntity := createEntityWithMod(t, "other-mod")
        
        _, err := api.Components().Get(otherModEntity, ecs.ComponentTypeSprite)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "permission denied")
    })
}

func TestModQueryAPI_Find(t *testing.T) {
    api := createTestModAPI(t, "test-mod")
    
    t.Run("MODエンティティのみ検索", func(t *testing.T) {
        // 自分のエンティティを作成
        myEntity, err := api.Entities().Create("my-entity")
        require.NoError(t, err)
        
        // 他MODとシステムエンティティを模擬作成
        createEntityWithMod(t, "other-mod")
        createSystemEntity(t)
        
        // クエリ実行 - 自分のエンティティのみ返却されるべき
        query := createTestQuery()
        results, err := api.Queries().Find(query)
        assert.NoError(t, err)
        assert.Len(t, results, 1)
        assert.Equal(t, myEntity, results[0])
    })
    
    t.Run("クエリ実行回数制限", func(t *testing.T) {
        query := createTestQuery()
        
        // 1000回まで成功
        for i := 0; i < 1000; i++ {
            _, err := api.Queries().Find(query)
            assert.NoError(t, err)
        }
        
        // 1001回目は失敗
        _, err := api.Queries().Find(query)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "query limit exceeded")
    })
}

func TestModSystemAPI_Register(t *testing.T) {
    api := createTestModAPI(t, "test-mod")
    
    t.Run("正常なシステム登録", func(t *testing.T) {
        system := createTestModSystem("test-system")
        err := api.Systems().Register(system)
        assert.NoError(t, err)
        
        registered := api.Systems().GetRegistered()
        assert.Contains(t, registered, "test-system")
    })
    
    t.Run("システム実行時間制限", func(t *testing.T) {
        // 10ms実行時間のシステム（制限5ms超過）
        longRunningSystem := createLongRunningModSystem("long-system", 10*time.Millisecond)
        err := api.Systems().Register(longRunningSystem)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "execution time exceeds limit")
    })
}

// セキュリティテスト

func TestModAPI_Security_PathTraversal(t *testing.T) {
    api := createTestModAPI(t, "malicious-mod")
    
    t.Run("パストラバーサル攻撃防御", func(t *testing.T) {
        // 悪意のあるタグでエンティティ作成試行
        maliciousTags := []string{
            "../../../etc/passwd",
            "..\\..\\..\\windows\\system32",
            "../../../../root/.ssh/id_rsa",
        }
        
        for _, tag := range maliciousTags {
            _, err := api.Entities().Create(tag)
            // タグは受け入れるが、ファイルアクセスは発生しない
            assert.NoError(t, err)
        }
        
        // しかし、ファイルシステムアクセスは完全ブロック
        // （この検証は実際の実装で行う）
    })
}

func TestModAPI_Security_SystemCommand(t *testing.T) {
    api := createTestModAPI(t, "malicious-mod")
    
    t.Run("システムコマンド実行防止", func(t *testing.T) {
        // 悪意のあるシステム登録試行
        maliciousSystem := createMaliciousSystem("rm -rf /")
        err := api.Systems().Register(maliciousSystem)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "security violation")
    })
}

func TestModAPI_ResourceLimits(t *testing.T) {
    api := createTestModAPI(t, "resource-test-mod")
    
    t.Run("メモリ使用量制限", func(t *testing.T) {
        // 大量のエンティティ作成でメモリ制限テスト
        for i := 0; i < 1000; i++ {
            _, err := api.Entities().Create("memory-test")
            if err != nil {
                assert.Contains(t, err.Error(), "memory limit exceeded")
                break
            }
        }
        
        // コンテキストのメモリ使用量確認
        ctx := api.GetContext()
        assert.True(t, ctx.MemoryUsage <= ctx.MaxMemory)
    })
}

// パフォーマンステスト

func BenchmarkModAPI_EntityCreation(b *testing.B) {
    api := createTestModAPI(b, "perf-test-mod")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        entityID, err := api.Entities().Create("perf-test")
        if err != nil {
            b.Fatal(err)
        }
        api.Entities().Delete(entityID)
    }
}

func BenchmarkModAPI_ComponentOperations(b *testing.B) {
    api := createTestModAPI(b, "perf-test-mod")
    entityID, _ := api.Entities().Create("perf-test")
    component := createTestSpriteComponent()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        api.Components().Add(entityID, component)
        api.Components().Get(entityID, ecs.ComponentTypeSprite)
        api.Components().Remove(entityID, ecs.ComponentTypeSprite)
    }
}

// テストヘルパー関数（まだ実装されていない）

func createTestModAPI(t testing.TB, modID string) ModECSAPI {
    // 実装されていないため、テスト実行時に失敗する
    panic("ModECSAPI implementation not found")
}

func createEntityWithMod(t testing.TB, modID string) ecs.EntityID {
    // 他MODのエンティティを模擬作成
    panic("Entity creation with mod not implemented")
}

func createSystemEntity(t testing.TB) ecs.EntityID {
    // システムエンティティを模擬作成
    panic("System entity creation not implemented")
}

func createTestSpriteComponent() ecs.Component {
    // テスト用Spriteコンポーネント作成
    panic("Test sprite component creation not implemented")
}

func createTestFileIOComponent() ecs.Component {
    // 禁止されたFileIOコンポーネント作成
    panic("FileIO component creation not implemented")
}

func createTestQuery() ecs.Query {
    // テスト用クエリ作成
    panic("Test query creation not implemented")
}

func createTestModSystem(id string) ModSystem {
    // テスト用MODシステム作成
    panic("Test mod system creation not implemented")
}

func createLongRunningModSystem(id string, duration time.Duration) ModSystem {
    // 長時間実行MODシステム作成
    panic("Long running mod system creation not implemented")
}

func createMaliciousSystem(command string) ModSystem {
    // 悪意のあるシステム作成
    panic("Malicious system creation not implemented")
}
```

### 3. エラー定義

実装先: `internal/core/ecs/mod/errors.go`

```go
package mod

import (
    "errors"
    "fmt"
)

var (
    // Entity関連エラー
    ErrEntityLimitExceeded = errors.New("entity limit exceeded")
    ErrEntityPermissionDenied = errors.New("entity permission denied")
    ErrSystemEntityAccess = errors.New("system entity access denied")
    
    // Component関連エラー
    ErrComponentNotAllowed = errors.New("component not allowed")
    ErrComponentPermissionDenied = errors.New("component permission denied")
    
    // Query関連エラー
    ErrQueryLimitExceeded = errors.New("query limit exceeded")
    ErrQueryTimeoutExceeded = errors.New("query timeout exceeded")
    
    // System関連エラー
    ErrSystemExecutionTimeExceeded = errors.New("system execution time exceeded")
    ErrSystemMemoryLimitExceeded = errors.New("system memory limit exceeded")
    ErrSecurityViolation = errors.New("security violation")
    
    // MOD関連エラー
    ErrMemoryLimitExceeded = errors.New("memory limit exceeded")
    ErrModNotFound = errors.New("mod not found")
    ErrModAlreadyExists = errors.New("mod already exists")
)

// SecurityError セキュリティ違反エラー
type SecurityError struct {
    ModID     string
    Operation string
    Reason    string
}

func (e *SecurityError) Error() string {
    return fmt.Sprintf("security violation in mod %s: %s (%s)", e.ModID, e.Operation, e.Reason)
}

// ResourceError リソース制限エラー
type ResourceError struct {
    ModID    string
    Resource string
    Current  int64
    Limit    int64
}

func (e *ResourceError) Error() string {
    return fmt.Sprintf("resource limit exceeded in mod %s: %s (%d/%d)", e.ModID, e.Resource, e.Current, e.Limit)
}
```

## Red段階実行手順

### 1. ファイル作成
```bash
# インターフェース定義
touch internal/core/ecs/mod/interfaces.go

# エラー定義
touch internal/core/ecs/mod/errors.go

# テストファイル
touch internal/core/ecs/mod/mod_api_test.go
```

### 2. テスト実行（失敗確認）
```bash
# 単体テスト実行（全て失敗することを確認）
go test ./internal/core/ecs/mod/ -v

# 期待される結果: 全テストケースがpanic/failureで失敗
```

### 3. 失敗状況の確認
- [ ] `TestModEntityAPI_Create`: panic "ModECSAPI implementation not found"
- [ ] `TestModEntityAPI_Delete`: panic "Entity creation with mod not implemented"  
- [ ] `TestModComponentAPI_Add`: panic "Test sprite component creation not implemented"
- [ ] `TestModComponentAPI_Get`: panic同上
- [ ] `TestModQueryAPI_Find`: panic "Test query creation not implemented"
- [ ] `TestModSystemAPI_Register`: panic "Test mod system creation not implemented"
- [ ] セキュリティテスト: panic各種ヘルパー関数未実装
- [ ] パフォーマンステスト: panic同上

## Red段階の成功基準

- [ ] 全てのテストファイルが作成される
- [ ] インターフェース定義が完成する
- [ ] エラー型定義が完成する
- [ ] テスト実行時に全テストケースが確実に失敗する
- [ ] 失敗理由が実装未完了であることが明確

## 次段階への準備

Red段階完了後、Green段階で以下を実装します：
1. `ModECSAPIImpl` 基本実装
2. セキュリティ制約の基本実装  
3. リソース制限の基本実装
4. テストヘルパー関数の実装

---

**作成日時**: 2025-08-08  
**段階**: TDD Red  
**期待結果**: 全テスト失敗 ✅