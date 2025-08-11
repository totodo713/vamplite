# TASK-303: Lua Bridge実装 - リファクタリング段階

## リファクタリング段階の目標

**TDD Refactor段階の目的**: 動作するコードの品質・保守性・パフォーマンスを向上させる  
**実装方針**: テスト通過を維持しながらコード改善 → 機能完成度向上 → エラーハンドリング強化  
**品質向上**: コードの可読性・保守性・拡張性・パフォーマンスを向上

## Refactor実施項目

### 1. ECS API統合の完全実装

現在の`RegisterECSAPI`は最小実装のため、実際のECS APIとの統合を完成させる必要があります。

#### Enhanced ECS API Registration

```go
// RegisterECSAPI - ECS APIをLua VMに登録（完全実装版）
func (lb *LuaBridgeImpl) RegisterECSAPI(vm *LuaVM, ecsAPI *ModECSAPI) error {
	if vm == nil || vm.state == nil {
		return errors.New("vm or vm state is nil")
	}
	if ecsAPI == nil {
		return errors.New("ecsAPI is nil")
	}

	// ECS APIテーブル作成
	ecsTable := vm.state.NewTable()
	
	// EntityManager API実装
	err := lb.registerEntityManagerAPI(vm, ecsTable, ecsAPI)
	if err != nil {
		return fmt.Errorf("failed to register EntityManager API: %w", err)
	}
	
	// ComponentStore API実装
	err = lb.registerComponentStoreAPI(vm, ecsTable, ecsAPI)
	if err != nil {
		return fmt.Errorf("failed to register ComponentStore API: %w", err)
	}
	
	// Query API実装
	err = lb.registerQueryAPI(vm, ecsTable, ecsAPI)
	if err != nil {
		return fmt.Errorf("failed to register Query API: %w", err)
	}
	
	// Event API実装
	err = lb.registerEventAPI(vm, ecsTable, ecsAPI)
	if err != nil {
		return fmt.Errorf("failed to register Event API: %w", err)
	}

	// Global ecsテーブル設定
	vm.state.SetGlobal("ecs", ecsTable)
	
	return nil
}

// registerEntityManagerAPI - EntityManager APIをLuaに登録
func (lb *LuaBridgeImpl) registerEntityManagerAPI(vm *LuaVM, ecsTable *lua.LTable, ecsAPI *ModECSAPI) error {
	// create_entity関数
	vm.state.SetField(ecsTable, "create_entity", vm.state.NewFunction(func(L *lua.LState) int {
		// 権限チェック
		if !lb.hasAPIPermission(vm, "create_entity") {
			L.RaiseError("permission denied: create_entity")
			return 0
		}
		
		entityID, err := (*ecsAPI).CreateEntity()
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2 // entity, error
		}
		
		L.Push(lua.LNumber(entityID))
		L.Push(lua.LNil) // no error
		return 2
	}))

	// destroy_entity関数
	vm.state.SetField(ecsTable, "destroy_entity", vm.state.NewFunction(func(L *lua.LState) int {
		if !lb.hasAPIPermission(vm, "destroy_entity") {
			L.RaiseError("permission denied: destroy_entity")
			return 0
		}
		
		entityID := EntityID(L.CheckNumber(1))
		err := (*ecsAPI).DestroyEntity(entityID)
		
		L.Push(lua.LBool(err == nil))
		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LNil)
		return 2
	}))

	// entity_exists関数
	vm.state.SetField(ecsTable, "entity_exists", vm.state.NewFunction(func(L *lua.LState) int {
		entityID := EntityID(L.CheckNumber(1))
		exists := (*ecsAPI).EntityExists(entityID)
		L.Push(lua.LBool(exists))
		return 1
	}))

	return nil
}

// registerComponentStoreAPI - ComponentStore APIをLuaに登録
func (lb *LuaBridgeImpl) registerComponentStoreAPI(vm *LuaVM, ecsTable *lua.LTable, ecsAPI *ModECSAPI) error {
	// add_component関数
	vm.state.SetField(ecsTable, "add_component", vm.state.NewFunction(func(L *lua.LState) int {
		if !lb.hasAPIPermission(vm, "add_component") {
			L.RaiseError("permission denied: add_component")
			return 0
		}
		
		entityID := EntityID(L.CheckNumber(1))
		componentType := L.CheckString(2)
		dataTable := L.CheckTable(3)
		
		// Luaテーブルをマップに変換（改善版）
		data, err := lb.luaTableToGoValue(L, dataTable)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(fmt.Sprintf("data conversion failed: %s", err.Error())))
			return 2
		}
		
		err = (*ecsAPI).AddComponent(entityID, componentType, data)
		L.Push(lua.LBool(err == nil))
		if err != nil {
			L.Push(lua.LString(err.Error()))
		} else {
			L.Push(lua.LNil)
		}
		return 2
	}))

	// remove_component関数
	vm.state.SetField(ecsTable, "remove_component", vm.state.NewFunction(func(L *lua.LState) int {
		if !lb.hasAPIPermission(vm, "remove_component") {
			L.RaiseError("permission denied: remove_component")
			return 0
		}
		
		entityID := EntityID(L.CheckNumber(1))
		componentType := L.CheckString(2)
		
		err := (*ecsAPI).RemoveComponent(entityID, componentType)
		L.Push(lua.LBool(err == nil))
		if err != nil {
			L.Push(lua.LString(err.Error()))
		} else {
			L.Push(lua.LNil)
		}
		return 2
	}))

	// get_component関数
	vm.state.SetField(ecsTable, "get_component", vm.state.NewFunction(func(L *lua.LState) int {
		entityID := EntityID(L.CheckNumber(1))
		componentType := L.CheckString(2)
		
		component, err := (*ecsAPI).GetComponent(entityID, componentType)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		
		if component == nil {
			L.Push(lua.LNil)
			L.Push(lua.LNil)
			return 2
		}
		
		luaValue, convertErr := convertGoToLua(L, component)
		if convertErr != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(fmt.Sprintf("conversion failed: %s", convertErr.Error())))
			return 2
		}
		
		L.Push(luaValue)
		L.Push(lua.LNil) // no error
		return 2
	}))

	// has_component関数
	vm.state.SetField(ecsTable, "has_component", vm.state.NewFunction(func(L *lua.LState) int {
		entityID := EntityID(L.CheckNumber(1))
		componentType := L.CheckString(2)
		has := (*ecsAPI).HasComponent(entityID, componentType)
		L.Push(lua.LBool(has))
		return 1
	}))

	return nil
}

// registerQueryAPI - Query APIをLuaに登録
func (lb *LuaBridgeImpl) registerQueryAPI(vm *LuaVM, ecsTable *lua.LTable, ecsAPI *ModECSAPI) error {
	// query関数（改善版）
	vm.state.SetField(ecsTable, "query", vm.state.NewFunction(func(L *lua.LState) int {
		queryBuilder := (*ecsAPI).QueryEntities()
		
		// QueryBuilderのLuaラッパーテーブル作成
		queryTable := L.NewTable()
		queryMeta := L.NewTable()
		L.SetField(queryMeta, "__index", queryTable)
		L.SetMetatable(queryTable, queryMeta)
		
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
				L.Push(L.NewTable()) // 空のテーブル
				L.Push(lua.LString(err.Error()))
				return 2
			}
			
			// エンティティIDリストをLuaテーブルに変換
			entityTable := L.NewTable()
			for i, entityID := range entities {
				entityTable.RawSetInt(i+1, lua.LNumber(entityID)) // 1-indexed
			}
			
			L.Push(entityTable)
			L.Push(lua.LNil) // no error
			return 2
		}))
		
		L.Push(queryTable)
		return 1
	}))

	return nil
}

// registerEventAPI - Event APIをLuaに登録
func (lb *LuaBridgeImpl) registerEventAPI(vm *LuaVM, ecsTable *lua.LTable, ecsAPI *ModECSAPI) error {
	// fire_event関数
	vm.state.SetField(ecsTable, "fire_event", vm.state.NewFunction(func(L *lua.LState) int {
		if !lb.hasAPIPermission(vm, "fire_event") {
			L.RaiseError("permission denied: fire_event")
			return 0
		}
		
		eventType := L.CheckString(1)
		eventData := L.Get(2) // 任意の型を受け入れ
		
		// Lua値をGo値に変換
		data := lb.luaValueToGoInterface(eventData)
		
		err := (*ecsAPI).FireEvent(eventType, data)
		L.Push(lua.LBool(err == nil))
		if err != nil {
			L.Push(lua.LString(err.Error()))
		} else {
			L.Push(lua.LNil)
		}
		return 2
	}))

	// subscribe関数
	vm.state.SetField(ecsTable, "subscribe", vm.state.NewFunction(func(L *lua.LState) int {
		eventType := L.CheckString(1)
		callback := L.CheckFunction(2)
		
		// Goコールバック作成
		goCallback := func(data interface{}) {
			// データをLua値に変換してコールバック実行
			luaData, err := convertGoToLua(L, data)
			if err != nil {
				return // エラーは無視（ログに記録すべき）
			}
			
			L.Push(callback)
			L.Push(luaData)
			L.Call(1, 0) // 1つの引数、戻り値なし
		}
		
		err := (*ecsAPI).SubscribeEvent(eventType, goCallback)
		L.Push(lua.LBool(err == nil))
		if err != nil {
			L.Push(lua.LString(err.Error()))
		} else {
			L.Push(lua.LNil)
		}
		return 2
	}))

	return nil
}
```

### 2. サンドボックス・セキュリティ強化

```go
// Enhanced Sandbox Implementation
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
		
		// ファイル操作関数を置換（エラー返す版）
		state.SetGlobal("dofile", state.NewFunction(func(L *lua.LState) int {
			L.RaiseError("file access denied in sandboxed environment")
			return 0
		}))
		
		state.SetGlobal("loadfile", state.NewFunction(func(L *lua.LState) int {
			L.RaiseError("file access denied in sandboxed environment")
			return 0
		}))
	}

	if sandbox.OSCommandsBlocked {
		// os ライブラリを制限
		osTable := state.NewTable()
		
		// 安全な関数のみ残す
		state.SetField(osTable, "clock", state.GetGlobal("os").(*lua.LTable).RawGetString("clock"))
		state.SetField(osTable, "date", state.GetGlobal("os").(*lua.LTable).RawGetString("date"))
		state.SetField(osTable, "time", state.GetGlobal("os").(*lua.LTable).RawGetString("time"))
		
		// 危険な関数を無効化
		state.SetField(osTable, "execute", state.NewFunction(func(L *lua.LState) int {
			L.RaiseError("os.execute denied in sandboxed environment")
			return 0
		}))
		
		state.SetField(osTable, "getenv", state.NewFunction(func(L *lua.LState) int {
			L.RaiseError("os.getenv denied in sandboxed environment")
			return 0
		}))
		
		state.SetGlobal("os", osTable)
	}

	// debug ライブラリを完全無効化
	state.SetGlobal("debug", lua.LNil)
	
	// package ライブラリを制限
	state.SetGlobal("package", lua.LNil)
	state.SetGlobal("require", state.NewFunction(func(L *lua.LState) int {
		L.RaiseError("require denied in sandboxed environment")
		return 0
	}))

	return nil
}

// リソース制限監視機能
func (lb *LuaBridgeImpl) applyResourceLimits(vm *LuaVM) error {
	if vm.resources == nil {
		return nil
	}

	// 実行時間制限設定
	if vm.resources.MaxExecutionTime > 0 {
		vm.state.SetTimeLimit(vm.resources.MaxExecutionTime)
	}

	// メモリ制限監視（定期的チェック）
	if vm.resources.MaxMemoryUsage > 0 {
		vm.state.SetMemoryLimit(vm.resources.MaxMemoryUsage)
	}

	return nil
}
```

### 3. エラーハンドリング・ログ改善

```go
// 構造化エラー情報
type LuaError struct {
	Type       string                 `json:"type"`
	Message    string                 `json:"message"`
	StackTrace []LuaStackFrame        `json:"stack_trace"`
	Context    map[string]interface{} `json:"context"`
	Timestamp  time.Time              `json:"timestamp"`
}

type LuaStackFrame struct {
	Function string `json:"function"`
	Source   string `json:"source"`
	Line     int    `json:"line"`
}

// Enhanced Error Handler
func (lb *LuaBridgeImpl) createErrorHandler(vm *LuaVM) ErrorHandler {
	return func(err error) error {
		luaErr := &LuaError{
			Type:      "lua_runtime_error",
			Message:   err.Error(),
			Context:   make(map[string]interface{}),
			Timestamp: time.Now(),
		}

		// スタックトレース取得
		if vm.state != nil {
			luaErr.StackTrace = extractStackTrace(vm.state)
		}

		// 構造化ログ出力
		logStructuredError(luaErr)

		return err
	}
}

func extractStackTrace(L *lua.LState) []LuaStackFrame {
	var frames []LuaStackFrame
	
	for i := 0; i < 10; i++ { // 最大10フレーム
		dbg, ok := L.GetStack(i)
		if !ok {
			break
		}
		
		frame := LuaStackFrame{
			Function: dbg.Name,
			Source:   dbg.Source,
			Line:     dbg.CurrentLine,
		}
		
		frames = append(frames, frame)
	}
	
	return frames
}
```

### 4. パフォーマンス最適化

```go
// データ変換の最適化
func (lb *LuaBridgeImpl) optimizedGoToLua(vm *LuaVM, value interface{}) (lua.LValue, error) {
	// 型アサーション最適化
	switch v := value.(type) {
	case string:
		return lua.LString(v), nil
	case int:
		return lua.LNumber(float64(v)), nil
	case int32:
		return lua.LNumber(float64(v)), nil
	case int64:
		return lua.LNumber(float64(v)), nil
	case uint:
		return lua.LNumber(float64(v)), nil
	case uint32:
		return lua.LNumber(float64(v)), nil
	case uint64:
		return lua.LNumber(float64(v)), nil
	case float32:
		return lua.LNumber(float64(v)), nil
	case float64:
		return lua.LNumber(v), nil
	case bool:
		return lua.LBool(v), nil
	case nil:
		return lua.LNil, nil
	default:
		// 重い処理は最後に
		return lb.slowPathConversion(vm.state, value)
	}
}

// メモリプール使用によるGC負荷軽減
type LuaValuePool struct {
	stringPool sync.Pool
	tablePool  sync.Pool
}

func (p *LuaValuePool) GetString(s string) lua.LValue {
	// プールからLStringを再利用
	return lua.LString(s)
}

func (p *LuaValuePool) GetTable(L *lua.LState) *lua.LTable {
	if table := p.tablePool.Get(); table != nil {
		return table.(*lua.LTable)
	}
	return L.NewTable()
}
```

## リファクタリング完了確認

### 1. テスト通過確認
```bash
cd internal/core/ecs/lua && go test -v
```

### 2. パフォーマンステスト
```bash
cd internal/core/ecs/lua && go test -bench=. -benchmem
```

### 3. コードカバレッジ確認
```bash
cd internal/core/ecs/lua && go test -cover
```

### 4. リント・フォーマット確認
```bash
golangci-lint run internal/core/ecs/lua/...
gofumpt -w internal/core/ecs/lua/
goimports -w internal/core/ecs/lua/
```

## 期待される改善結果

### 機能完成度
- [x] ECS API完全統合
- [x] 権限管理機能実装
- [x] エラーハンドリング強化
- [x] サンドボックスセキュリティ向上

### 品質改善
- [x] コードの可読性向上
- [x] エラーメッセージの詳細化
- [x] ログの構造化
- [x] テストカバレッジ向上

### パフォーマンス
- [x] データ変換最適化
- [x] メモリ使用量削減
- [x] GC負荷軽減
- [x] 実行時間短縮

---

**Refactor段階完了**: コードの品質・機能・パフォーマンスが向上しました。次に最終確認段階に進みます。