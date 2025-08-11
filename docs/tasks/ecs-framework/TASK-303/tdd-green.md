# TASK-303: Lua Bridge実装 - 最小実装 (Green段階)

## Green段階の目標

**TDD Green段階の目的**: 失敗するテストを通すための最小限の実装を作成する  
**実装方針**: 最小限で簡単な実装から始める → テスト通過 → 段階的に機能追加  
**実装優先順位**: コンパイル → 基本機能 → データ変換 → API統合  

## Phase 1: 基本実装・コンパイルエラー解決

### 基本的なLuaBridge実装

#### `internal/core/ecs/lua/lua_bridge.go`
```go
package lua

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/yuin/gopher-lua"
)

// LuaBridgeImpl - LuaBridgeインターフェースの実装
type LuaBridgeImpl struct{}

// NewLuaBridge - LuaBridge実装のコンストラクタ
func NewLuaBridge() LuaBridge {
	return &LuaBridgeImpl{}
}

// CreateVM - Lua VM作成
func (lb *LuaBridgeImpl) CreateVM(config *LuaVMConfig) (*LuaVM, error) {
	if config == nil {
		config = &LuaVMConfig{
			SandboxEnabled: false,
			ResourceLimits: &ResourceLimits{
				MaxExecutionTime: 100 * time.Millisecond,
				MaxMemoryUsage:   10 * 1024 * 1024, // 10MB
			},
		}
	}

	// Lua State作成
	state := lua.NewState()
	if state == nil {
		return nil, errors.New("failed to create Lua state")
	}

	// サンドボックス設定
	var sandbox *Sandbox
	if config.SandboxEnabled {
		sandbox = &Sandbox{
			FileSystemRestricted: true,
			NetworkRestricted:    true,
			OSCommandsBlocked:    true,
		}
		
		// サンドボックス適用
		err := applySandbox(state, sandbox)
		if err != nil {
			state.Close()
			return nil, fmt.Errorf("failed to apply sandbox: %w", err)
		}
	}

	vm := &LuaVM{
		state:       state,
		sandbox:     sandbox,
		permissions: config.Permissions,
		resources:   config.ResourceLimits,
	}

	return vm, nil
}

// DestroyVM - Lua VM削除
func (lb *LuaBridgeImpl) DestroyVM(vm *LuaVM) error {
	if vm == nil {
		return errors.New("vm is nil")
	}
	if vm.state == nil {
		return errors.New("vm state is nil")
	}

	vm.state.Close()
	return nil
}

// LoadScript - Luaスクリプト読み込み
func (lb *LuaBridgeImpl) LoadScript(vm *LuaVM, scriptPath string) (*LuaScript, error) {
	// 最小実装: 基本的なスクリプト情報のみ作成
	script := &LuaScript{
		path:   scriptPath,
		loaded: false,
		metadata: &ScriptMetadata{
			Name:       scriptPath,
			Version:    "1.0.0",
			APIVersion: "1.0.0",
		},
	}

	return script, nil
}

// UnloadScript - Luaスクリプトアンロード
func (lb *LuaBridgeImpl) UnloadScript(vm *LuaVM, script *LuaScript) error {
	if script == nil {
		return errors.New("script is nil")
	}
	
	script.loaded = false
	return nil
}

// ExecuteScript - Luaスクリプト実行
func (lb *LuaBridgeImpl) ExecuteScript(vm *LuaVM, script *LuaScript) error {
	if vm == nil || vm.state == nil {
		return errors.New("vm or vm state is nil")
	}
	if script == nil {
		return errors.New("script is nil")
	}

	// 最小実装: 空のスクリプト実行のみ対応
	err := vm.state.DoString("-- empty script")
	if err != nil {
		return fmt.Errorf("script execution failed: %w", err)
	}

	script.loaded = true
	return nil
}

// GoToLua - Go値をLua値に変換
func (lb *LuaBridgeImpl) GoToLua(vm *LuaVM, value interface{}) (lua.LValue, error) {
	if vm == nil || vm.state == nil {
		return nil, errors.New("vm or vm state is nil")
	}

	return convertGoToLua(vm.state, value)
}

// LuaToGo - Lua値をGo値に変換
func (lb *LuaBridgeImpl) LuaToGo(vm *LuaVM, value lua.LValue, target interface{}) error {
	if vm == nil || vm.state == nil {
		return errors.New("vm or vm state is nil")
	}

	return convertLuaToGo(value, target)
}

// RegisterECSAPI - ECS APIをLua VMに登録
func (lb *LuaBridgeImpl) RegisterECSAPI(vm *LuaVM, ecsAPI *ModECSAPI) error {
	if vm == nil || vm.state == nil {
		return errors.New("vm or vm state is nil")
	}
	if ecsAPI == nil {
		return errors.New("ecsAPI is nil")
	}

	// 最小実装: ECS APIテーブル作成
	ecsTable := vm.state.NewTable()
	
	// create_entity関数登録
	vm.state.SetField(ecsTable, "create_entity", vm.state.NewFunction(func(L *lua.LState) int {
		if ecsAPI == nil {
			L.Push(lua.LNil)
			return 1
		}
		
		entityID, err := (*ecsAPI).CreateEntity()
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		
		L.Push(lua.LNumber(entityID))
		return 1
	}))

	// entity_exists関数登録
	vm.state.SetField(ecsTable, "entity_exists", vm.state.NewFunction(func(L *lua.LState) int {
		entityID := EntityID(L.CheckNumber(1))
		exists := (*ecsAPI).EntityExists(entityID)
		L.Push(lua.LBool(exists))
		return 1
	}))

	// add_component関数登録
	vm.state.SetField(ecsTable, "add_component", vm.state.NewFunction(func(L *lua.LState) int {
		entityID := EntityID(L.CheckNumber(1))
		componentType := L.CheckString(2)
		dataTable := L.CheckTable(3)
		
		// Luaテーブルをマップに変換
		data := make(map[string]interface{})
		dataTable.ForEach(func(key, value lua.LValue) {
			if keyStr, ok := key.(lua.LString); ok {
				data[string(keyStr)] = luaValueToInterface(value)
			}
		})
		
		err := (*ecsAPI).AddComponent(entityID, componentType, data)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	// has_component関数登録
	vm.state.SetField(ecsTable, "has_component", vm.state.NewFunction(func(L *lua.LState) int {
		entityID := EntityID(L.CheckNumber(1))
		componentType := L.CheckString(2)
		has := (*ecsAPI).HasComponent(entityID, componentType)
		L.Push(lua.LBool(has))
		return 1
	}))

	// get_component関数登録
	vm.state.SetField(ecsTable, "get_component", vm.state.NewFunction(func(L *lua.LState) int {
		entityID := EntityID(L.CheckNumber(1))
		componentType := L.CheckString(2)
		
		component, err := (*ecsAPI).GetComponent(entityID, componentType)
		if err != nil || component == nil {
			L.Push(lua.LNil)
			return 1
		}
		
		luaValue, convertErr := convertGoToLua(L, component)
		if convertErr != nil {
			L.Push(lua.LNil)
			return 1
		}
		
		L.Push(luaValue)
		return 1
	}))

	// remove_component関数登録
	vm.state.SetField(ecsTable, "remove_component", vm.state.NewFunction(func(L *lua.LState) int {
		entityID := EntityID(L.CheckNumber(1))
		componentType := L.CheckString(2)
		
		err := (*ecsAPI).RemoveComponent(entityID, componentType)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	// query関数登録（簡易版）
	vm.state.SetField(ecsTable, "query", vm.state.NewFunction(func(L *lua.LState) int {
		queryBuilder := (*ecsAPI).QueryEntities()
		
		// QueryBuilderのLuaラッパーテーブル作成
		queryTable := L.NewTable()
		
		// with関数
		L.SetField(queryTable, "with", L.NewFunction(func(L *lua.LState) int {
			componentType := L.CheckString(1)
			queryBuilder = queryBuilder.With(componentType)
			L.Push(queryTable) // チェーンのため自身を返す
			return 1
		}))
		
		// without関数
		L.SetField(queryTable, "without", L.NewFunction(func(L *lua.LState) int {
			componentType := L.CheckString(1)
			queryBuilder = queryBuilder.Without(componentType)
			L.Push(queryTable) // チェーンのため自身を返す
			return 1
		}))
		
		// execute関数
		L.SetField(queryTable, "execute", L.NewFunction(func(L *lua.LState) int {
			entities, err := queryBuilder.Execute()
			if err != nil {
				L.Push(L.NewTable()) // 空のテーブルを返す
				return 1
			}
			
			// エンティティIDリストをLuaテーブルに変換
			entityTable := L.NewTable()
			for i, entityID := range entities {
				entityTable.RawSetInt(i+1, lua.LNumber(entityID)) // 1-indexed
			}
			
			L.Push(entityTable)
			return 1
		}))
		
		L.Push(queryTable)
		return 1
	}))

	// Global ecsテーブル設定
	vm.state.SetGlobal("ecs", ecsTable)
	
	return nil
}

// SetPermissions - API権限設定
func (lb *LuaBridgeImpl) SetPermissions(vm *LuaVM, permissions *APIPermissions) error {
	if vm == nil {
		return errors.New("vm is nil")
	}
	
	vm.permissions = permissions
	return nil
}

// ========== ヘルパー関数 ==========

// applySandbox - サンドボックス制限を適用
func applySandbox(state *lua.LState, sandbox *Sandbox) error {
	if sandbox == nil {
		return nil
	}

	// 危険な関数・ライブラリを無効化
	if sandbox.FileSystemRestricted {
		// io ライブラリを制限
		state.SetGlobal("io", lua.LNil)
		state.SetGlobal("dofile", lua.LNil)
		state.SetGlobal("loadfile", lua.LNil)
	}

	if sandbox.OSCommandsBlocked {
		// os ライブラリを制限
		state.SetGlobal("os", lua.LNil)
	}

	// debug ライブラリを無効化
	state.SetGlobal("debug", lua.LNil)
	
	// package ライブラリを制限
	state.SetGlobal("package", lua.LNil)
	state.SetGlobal("require", lua.LNil)

	return nil
}

// convertGoToLua - Go値をLua値に変換
func convertGoToLua(state *lua.LState, value interface{}) (lua.LValue, error) {
	if value == nil {
		return lua.LNil, nil
	}

	switch v := value.(type) {
	case string:
		return lua.LString(v), nil
	case int:
		return lua.LNumber(float64(v)), nil
	case int64:
		return lua.LNumber(float64(v)), nil
	case float32:
		return lua.LNumber(float64(v)), nil
	case float64:
		return lua.LNumber(v), nil
	case bool:
		return lua.LBool(v), nil
	case []string:
		table := state.NewTable()
		for i, item := range v {
			table.RawSetInt(i+1, lua.LString(item)) // 1-indexed
		}
		return table, nil
	case []int:
		table := state.NewTable()
		for i, item := range v {
			table.RawSetInt(i+1, lua.LNumber(float64(item)))
		}
		return table, nil
	case map[string]interface{}:
		table := state.NewTable()
		for key, val := range v {
			luaVal, err := convertGoToLua(state, val)
			if err != nil {
				return nil, err
			}
			table.RawSetString(key, luaVal)
		}
		return table, nil
	default:
		// reflectionを使用してstructを変換
		return convertStructToLua(state, value)
	}
}

// convertStructToLua - 構造体をLuaテーブルに変換（reflection使用）
func convertStructToLua(state *lua.LState, value interface{}) (lua.LValue, error) {
	v := reflect.ValueOf(value)
	t := reflect.TypeOf(value)

	// ポインタの場合は実体を取得
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("unsupported type: %T", value)
	}

	table := state.NewTable()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 非公開フィールドはスキップ
		if !field.CanInterface() {
			continue
		}

		// JSONタグからフィールド名を取得
		fieldName := fieldType.Name
		if tag := fieldType.Tag.Get("json"); tag != "" && tag != "-" {
			fieldName = tag
		}

		// フィールド値をLua値に変換
		luaVal, err := convertGoToLua(state, field.Interface())
		if err != nil {
			return nil, fmt.Errorf("failed to convert field %s: %w", fieldName, err)
		}

		table.RawSetString(fieldName, luaVal)
	}

	return table, nil
}

// convertLuaToGo - Lua値をGo値に変換
func convertLuaToGo(value lua.LValue, target interface{}) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return errors.New("target must be a pointer")
	}

	targetElem := targetValue.Elem()

	switch value.Type() {
	case lua.LTString:
		if targetElem.Kind() == reflect.String {
			targetElem.SetString(string(value.(lua.LString)))
			return nil
		}
	case lua.LTNumber:
		num := float64(value.(lua.LNumber))
		switch targetElem.Kind() {
		case reflect.Int:
			targetElem.SetInt(int64(num))
			return nil
		case reflect.Float64:
			targetElem.SetFloat(num)
			return nil
		}
	case lua.LTBool:
		if targetElem.Kind() == reflect.Bool {
			targetElem.SetBool(bool(value.(lua.LBool)))
			return nil
		}
	case lua.LTTable:
		// スライス変換をサポート
		if targetElem.Kind() == reflect.Slice {
			return convertLuaTableToSlice(value.(*lua.LTable), target)
		}
	case lua.LTNil:
		// nilの場合はゼロ値を設定
		targetElem.Set(reflect.Zero(targetElem.Type()))
		return nil
	}

	return fmt.Errorf("cannot convert Lua %s to Go %s", value.Type(), targetElem.Kind())
}

// convertLuaTableToSlice - LuaテーブルをGoスライスに変換
func convertLuaTableToSlice(table *lua.LTable, target interface{}) error {
	targetValue := reflect.ValueOf(target).Elem()
	elemType := targetValue.Type().Elem()

	var slice reflect.Value

	// テーブルを配列形式で処理（1-indexed）
	table.ForEach(func(key, value lua.LValue) {
		if !slice.IsValid() {
			slice = reflect.MakeSlice(targetValue.Type(), 0, 0)
		}

		elem := reflect.New(elemType).Elem()

		switch elemType.Kind() {
		case reflect.String:
			if value.Type() == lua.LTString {
				elem.SetString(string(value.(lua.LString)))
			}
		case reflect.Int:
			if value.Type() == lua.LTNumber {
				elem.SetInt(int64(float64(value.(lua.LNumber))))
			}
		case reflect.Float64:
			if value.Type() == lua.LTNumber {
				elem.SetFloat(float64(value.(lua.LNumber)))
			}
		}

		slice = reflect.Append(slice, elem)
	})

	if slice.IsValid() {
		targetValue.Set(slice)
	}

	return nil
}

// luaValueToInterface - Lua値をinterface{}に変換
func luaValueToInterface(value lua.LValue) interface{} {
	switch value.Type() {
	case lua.LTString:
		return string(value.(lua.LString))
	case lua.LTNumber:
		return float64(value.(lua.LNumber))
	case lua.LTBool:
		return bool(value.(lua.LBool))
	case lua.LTNil:
		return nil
	case lua.LTTable:
		// 簡易的なテーブル→マップ変換
		result := make(map[string]interface{})
		value.(*lua.LTable).ForEach(func(key, val lua.LValue) {
			if keyStr, ok := key.(lua.LString); ok {
				result[string(keyStr)] = luaValueToInterface(val)
			}
		})
		return result
	default:
		return nil
	}
}
```

### Mock ECS API実装

テストで使用するMock ECS APIを実装します。

#### `internal/core/ecs/lua/mock_ecs_api.go`
```go
package lua

import (
	"errors"
)

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
	if !m.entities[id] {
		return errors.New("entity not found")
	}
	delete(m.entities, id)
	delete(m.components, id)
	return nil
}

func (m *MockECSAPI) EntityExists(id EntityID) bool {
	return m.entities[id]
}

func (m *MockECSAPI) AddComponent(entityID EntityID, componentType string, data interface{}) error {
	if !m.entities[entityID] {
		return errors.New("entity not found")
	}
	m.components[entityID][componentType] = data
	return nil
}

func (m *MockECSAPI) RemoveComponent(entityID EntityID, componentType string) error {
	if !m.entities[entityID] {
		return errors.New("entity not found")
	}
	delete(m.components[entityID], componentType)
	return nil
}

func (m *MockECSAPI) GetComponent(entityID EntityID, componentType string) (interface{}, error) {
	if !m.entities[entityID] {
		return nil, errors.New("entity not found")
	}
	component, exists := m.components[entityID][componentType]
	if !exists {
		return nil, nil
	}
	return component, nil
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
	api          *MockECSAPI
	withTypes    []string
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

## テスト実行・Green段階確認

最小実装が完了したので、テストを実行してGreen段階を確認しましょう。

```bash
cd internal/core/ecs/lua && go test -v
```

期待される結果：
- コンパイルエラー解決
- 基本的なテストケースの通過
- 一部のテストは実装不完全で失敗する可能性（正常な段階的実装）

## 段階的機能追加

基本テストが通過した後、以下の順序で機能を段階的に追加します：

### Phase 2: ECS API統合の完成
- Lua APIテストの通過
- Event APIの実装
- クエリエンジン統合の完成

### Phase 3: 高度なデータ変換
- 複雑な構造体変換の完成
- エラーハンドリングの改善
- パフォーマンス最適化

### Phase 4: サンドボックス・セキュリティ
- セキュリティ制限の完全実装
- リソース制限の実装
- エラー処理の完成

## 実装上の注意事項

### 最小実装の原則
- **テストが通る最小限の実装のみ**を行う
- **過度な最適化や機能追加は避ける**
- **次のテストケースが失敗したら機能追加する**

### エラーハンドリング
- **基本的なnilチェック**は実装する
- **詳細なバリデーション**は必要最小限
- **エラーメッセージ**は簡潔に

### パフォーマンス考慮
- **現段階では基本的な動作のみ重視**
- **パフォーマンス最適化は後のRefactor段階で実施**
- **メモリリークは基本的な対策のみ実施**

---

**Green段階の目標**: 全基本テストケースの通過確認後、次のRefactor段階に進みます。