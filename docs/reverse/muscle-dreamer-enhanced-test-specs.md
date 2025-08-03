# Muscle Dreamer 拡張テスト仕様書（逆生成強化版）

## 分析概要

**分析日時**: 2025-08-03  
**対象コードベース**: /home/devman/GolandProjects/muscle-dreamer  
**テストカバレッジ**: 0% (全テスト未実装)  
**生成テストケース数**: 42個 (基本15個 + 拡張27個)  
**実装推奨テスト数**: 25個（高優先度）  
**型定義インターフェース**: 17セクション (400+行)  

## プロジェクト特性に基づくテスト戦略

### ゲーム固有の課題
- **リアルタイム性**: 60FPS保証、低レイテンシ  
- **状態管理**: ECS、ゲーム状態遷移  
- **プラットフォーム**: ネイティブ+WebAssembly  
- **拡張性**: MOD・テーマシステム  
- **セキュリティ**: MODサンドボックス  

### テストアプローチ
1. **パフォーマンス優先**: フレームレート、メモリ、起動時間
2. **クロスプラットフォーム**: ネイティブ vs WebAssembly
3. **モジュラリティ**: ECS、プラグインシステム
4. **セキュリティ**: MODサンドボックス、入力検証

## 1. 基本機能テスト仕様

### 1.1 ゲームコアエンジンテスト

#### TC-001: ゲームインスタンス管理
```go
// Test: NewGame() 正常性
func TestNewGame(t *testing.T) {
    game := core.NewGame()
    assert.NotNil(t, game)
    assert.IsType(t, &core.Game{}, game)
}

// Test: Game構造体初期状態
func TestGameInitialState(t *testing.T) {
    game := core.NewGame()
    
    // 将来のフィールド検証準備
    // assert.NotNil(t, game.entities)
    // assert.NotNil(t, game.systems)
    // assert.Equal(t, GameStateMenu, game.state)
}
```

#### TC-002: ゲームループ実行
```go
// Test: Update()の冪等性と安定性
func TestGameUpdateIdempotency(t *testing.T) {
    game := core.NewGame()
    
    // 1000回連続実行でのメモリリーク・エラー検証
    for i := 0; i < 1000; i++ {
        err := game.Update()
        assert.NoError(t, err)
    }
}

// Test: Update()パフォーマンス
func BenchmarkGameUpdate(b *testing.B) {
    game := core.NewGame()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        game.Update()
    }
    
    // 目標: < 1ms per update (60FPS = 16.67ms per frame)
}
```

#### TC-003: 描画システム検証
```go
// Test: Draw()の背景色検証
func TestGameDrawBackground(t *testing.T) {
    game := core.NewGame()
    screen := ebiten.NewImage(1280, 720)
    
    game.Draw(screen)
    
    // 背景色RGBA(50, 50, 100, 255)の検証
    pixel := screen.At(0, 0).(color.RGBA)
    expected := color.RGBA{50, 50, 100, 255}
    assert.Equal(t, expected, pixel)
}

// Test: Draw()パフォーマンス
func BenchmarkGameDraw(b *testing.B) {
    game := core.NewGame()
    screen := ebiten.NewImage(1280, 720)
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        game.Draw(screen)
    }
}
```

### 1.2 設定システムテスト

#### TC-004: YAML設定読み込み
```go
// Test: game.yaml正常読み込み
func TestConfigLoad(t *testing.T) {
    config, err := LoadGameConfig("config/game.yaml")
    assert.NoError(t, err)
    assert.NotNil(t, config)
    
    // 基本設定検証
    assert.Equal(t, "マッスルドリーマー〜観光編〜", config.Game.Title)
    assert.Equal(t, "0.1.0", config.Game.Version)
    assert.Equal(t, 1280, config.Graphics.Width)
    assert.Equal(t, 720, config.Graphics.Height)
    
    // オーディオ設定範囲検証
    assert.True(t, config.Audio.MasterVolume >= 0.0 && config.Audio.MasterVolume <= 1.0)
    assert.True(t, config.Audio.BGMVolume >= 0.0 && config.Audio.BGMVolume <= 1.0)
    assert.True(t, config.Audio.SFXVolume >= 0.0 && config.Audio.SFXVolume <= 1.0)
}

// Test: 不正YAML処理
func TestConfigInvalidYAML(t *testing.T) {
    testCases := []struct {
        name     string
        yaml     string
        expected error
    }{
        {"空ファイル", "", ErrEmptyConfig},
        {"不正YAML", "invalid: yaml: content", ErrInvalidYAML},
        {"負の解像度", "graphics:\n  width: -1\n  height: -1", ErrInvalidResolution},
        {"範囲外ボリューム", "audio:\n  master_volume: 2.0", ErrInvalidVolume},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            _, err := ParseYAMLConfig(tc.yaml)
            assert.ErrorIs(t, err, tc.expected)
        })
    }
}
```

## 2. ECSアーキテクチャテスト仕様

### 2.1 Entity Component System テスト

#### TC-005: EntityManager機能テスト
```go
// Test: エンティティライフサイクル
func TestEntityLifecycle(t *testing.T) {
    em := NewEntityManager()
    
    // エンティティ作成
    entity := em.CreateEntity()
    assert.NotEqual(t, EntityID(0), entity)
    
    // コンポーネント追加
    transform := &TransformComponent{X: 100, Y: 200, ScaleX: 1, ScaleY: 1}
    em.AddComponent(entity, transform)
    assert.True(t, em.HasComponent(entity, ComponentTransform))
    
    // コンポーネント取得
    retrieved := em.GetComponent(entity, ComponentTransform).(*TransformComponent)
    assert.Equal(t, 100.0, retrieved.X)
    assert.Equal(t, 200.0, retrieved.Y)
    
    // エンティティ削除
    em.DestroyEntity(entity)
    assert.False(t, em.HasComponent(entity, ComponentTransform))
}

// Test: 大量エンティティ性能
func TestEntityManagerPerformance(t *testing.T) {
    em := NewEntityManager()
    
    // 10,000エンティティ作成・削除
    entities := make([]EntityID, 10000)
    
    start := time.Now()
    for i := 0; i < 10000; i++ {
        entities[i] = em.CreateEntity()
        em.AddComponent(entities[i], &TransformComponent{})
    }
    createTime := time.Since(start)
    
    start = time.Now()
    for _, entity := range entities {
        em.DestroyEntity(entity)
    }
    destroyTime := time.Since(start)
    
    // パフォーマンス要件: 10,000エンティティ作成 < 100ms
    assert.Less(t, createTime, 100*time.Millisecond)
    assert.Less(t, destroyTime, 50*time.Millisecond)
}
```

#### TC-006: SystemManager機能テスト
```go
// Test: システム登録・実行
func TestSystemManagerExecution(t *testing.T) {
    sm := NewSystemManager()
    em := NewEntityManager()
    
    // テスト用システム
    movementSystem := &MockMovementSystem{}
    renderSystem := &MockRenderSystem{}
    
    sm.RegisterSystem(movementSystem)
    sm.RegisterSystem(renderSystem)
    
    // システム実行
    err := sm.UpdateSystems(0.016, em) // 60FPS = 16ms
    assert.NoError(t, err)
    
    assert.True(t, movementSystem.WasUpdated)
    assert.True(t, renderSystem.WasRendered)
}

// Test: システム実行順序
func TestSystemExecutionOrder(t *testing.T) {
    sm := NewSystemManager()
    
    var executionOrder []string
    
    sm.RegisterSystem(&OrderTestSystem{Name: "Input", Order: &executionOrder})
    sm.RegisterSystem(&OrderTestSystem{Name: "Physics", Order: &executionOrder})
    sm.RegisterSystem(&OrderTestSystem{Name: "Rendering", Order: &executionOrder})
    
    sm.UpdateSystems(0.016, NewEntityManager())
    
    expected := []string{"Input", "Physics", "Rendering"}
    assert.Equal(t, expected, executionOrder)
}
```

## 3. WebAssemblyプラットフォームテスト

### 3.1 WebAssembly特有テスト

#### TC-007: WebAssembly初期化テスト
```go
// Test: WebAssembly環境検出
func TestWebAssemblyEnvironment(t *testing.T) {
    if runtime.GOOS != "js" {
        t.Skip("WebAssembly環境でのみ実行")
    }
    
    bridge := NewWebAssemblyBridge()
    
    // ブラウザ情報取得
    browserInfo := bridge.GetBrowserInfo()
    assert.NotEmpty(t, browserInfo.UserAgent)
    assert.True(t, browserInfo.Supports.WebAssembly)
    
    // JavaScriptコール
    result, err := bridge.CallJavaScript("console.log", "Test from WebAssembly")
    assert.NoError(t, err)
    assert.NotNil(t, result)
}

// Test: WebAssembly ↔ JavaScript 通信
func TestWebAssemblyJavaScriptInterop(t *testing.T) {
    if runtime.GOOS != "js" {
        t.Skip("WebAssembly環境でのみ実行")
    }
    
    bridge := NewWebAssemblyBridge()
    
    // JavaScriptコールバック登録
    var callbackInvoked bool
    bridge.RegisterCallback("testCallback", func(args ...interface{}) interface{} {
        callbackInvoked = true
        return "callback success"
    })
    
    // JavaScript側からコールバック呼び出し
    bridge.CallJavaScript("invoke_go_callback", "testCallback", "test data")
    
    assert.True(t, callbackInvoked)
}
```

#### TC-008: WebAssemblyパフォーマンステスト
```go
// Test: WebAssemblyゲームループ性能
func TestWebAssemblyGameLoopPerformance(t *testing.T) {
    if runtime.GOOS != "js" {
        t.Skip("WebAssembly環境でのみ実行")
    }
    
    game := core.NewGame()
    
    // 100フレーム実行時間測定
    start := time.Now()
    for i := 0; i < 100; i++ {
        game.Update()
        // WebAssembly環境では実際の描画はスキップ
    }
    elapsed := time.Since(start)
    
    // 100フレーム < 2秒 (WebAssemblyオーバーヘッド考慮)
    assert.Less(t, elapsed, 2*time.Second)
    
    avgFrameTime := elapsed / 100
    t.Logf("Average frame time: %v", avgFrameTime)
}
```

## 4. アセット・テーマシステムテスト

### 4.1 アセット管理テスト

#### TC-009: アセット読み込みテスト
```go
// Test: 画像アセット読み込み
func TestAssetManagerImageLoading(t *testing.T) {
    am := NewAssetManager()
    
    // テスト用画像作成
    testImage := createTestImage(64, 64, color.RGBA{255, 0, 0, 255})
    saveTestImage(testImage, "test_assets/test_sprite.png")
    
    // アセット読み込み
    sprite, err := am.LoadImage("test_assets/test_sprite.png")
    assert.NoError(t, err)
    assert.NotNil(t, sprite)
    
    // 画像サイズ検証
    bounds := sprite.Bounds()
    assert.Equal(t, 64, bounds.Dx())
    assert.Equal(t, 64, bounds.Dy())
    
    // キャッシュ確認
    cached := am.GetLoadedAssets()
    assert.Contains(t, cached, "test_assets/test_sprite.png")
}

// Test: アセット読み込みエラー処理
func TestAssetManagerErrorHandling(t *testing.T) {
    am := NewAssetManager()
    
    testCases := []struct {
        path     string
        expected error
    }{
        {"nonexistent.png", ErrAssetNotFound},
        {"invalid.txt", ErrUnsupportedFormat},
        {"", ErrInvalidPath},
    }
    
    for _, tc := range testCases {
        t.Run(tc.path, func(t *testing.T) {
            _, err := am.LoadImage(tc.path)
            assert.ErrorIs(t, err, tc.expected)
        })
    }
}
```

#### TC-010: テーマシステムテスト
```go
// Test: テーマ読み込み・適用
func TestThemeManagerLoading(t *testing.T) {
    tm := NewThemeManager()
    
    // デフォルトテーマ読み込み
    theme, err := tm.LoadTheme("default")
    assert.NoError(t, err)
    assert.NotNil(t, theme)
    assert.Equal(t, "default", theme.Name)
    
    // テーマ適用
    err = tm.SetTheme("default")
    assert.NoError(t, err)
    
    current := tm.GetCurrentTheme()
    assert.Equal(t, "default", current.Name)
}

// Test: テーマファイル検証
func TestThemeValidation(t *testing.T) {
    invalidThemes := []struct {
        name   string
        yaml   string
        error  error
    }{
        {"空テーマ", "", ErrEmptyTheme},
        {"名前なし", "version: 1.0", ErrMissingThemeName},
        {"不正バージョン", "name: test\nversion: invalid", ErrInvalidVersion},
    }
    
    for _, theme := range invalidThemes {
        t.Run(theme.name, func(t *testing.T) {
            _, err := ParseThemeYAML(theme.yaml)
            assert.ErrorIs(t, err, theme.error)
        })
    }
}
```

## 5. MODシステムセキュリティテスト

### 5.1 MODサンドボックステスト

#### TC-011: MODセキュリティ検証
```go
// Test: ファイルアクセス制限
func TestModFileAccessRestriction(t *testing.T) {
    mod := &Mod{
        Name: "TestMod",
        Permissions: ModPermissions{
            FileAccess:    []string{"mods/testmod/"},
            NetworkAccess: false,
            SystemAccess:  false,
        },
    }
    
    sandbox := NewModSandbox(mod)
    
    // 許可されたパスへのアクセス
    err := sandbox.WriteFile("mods/testmod/data.txt", []byte("test"))
    assert.NoError(t, err)
    
    // 禁止されたパスへのアクセス
    err = sandbox.WriteFile("/etc/passwd", []byte("malicious"))
    assert.ErrorIs(t, err, ErrAccessDenied)
    
    err = sandbox.WriteFile("../../../etc/passwd", []byte("malicious"))
    assert.ErrorIs(t, err, ErrAccessDenied)
}

// Test: ネットワークアクセス制限
func TestModNetworkRestriction(t *testing.T) {
    mod := &Mod{
        Name: "TestMod",
        Permissions: ModPermissions{
            NetworkAccess: false,
        },
    }
    
    sandbox := NewModSandbox(mod)
    
    // ネットワークアクセス試行
    err := sandbox.HTTPGet("http://example.com")
    assert.ErrorIs(t, err, ErrNetworkAccessDenied)
    
    err = sandbox.TCPConnect("127.0.0.1:80")
    assert.ErrorIs(t, err, ErrNetworkAccessDenied)
}

// Test: システムコマンド実行制限
func TestModSystemCommandRestriction(t *testing.T) {
    mod := &Mod{
        Name: "TestMod",
        Permissions: ModPermissions{
            SystemAccess: false,
        },
    }
    
    sandbox := NewModSandbox(mod)
    
    // システムコマンド実行試行
    err := sandbox.ExecuteCommand("rm", "-rf", "/")
    assert.ErrorIs(t, err, ErrSystemAccessDenied)
    
    err = sandbox.ExecuteCommand("cat", "/etc/passwd")
    assert.ErrorIs(t, err, ErrSystemAccessDenied)
}
```

#### TC-012: MODインジェクション攻撃テスト
```go
// Test: スクリプトインジェクション対策
func TestModScriptInjectionPrevention(t *testing.T) {
    maliciousScripts := []string{
        "os.exit(1)",
        "import('child_process').exec('rm -rf /')",
        "eval('malicious code')",
        "require('fs').unlinkSync('/important/file')",
    }
    
    for _, script := range maliciousScripts {
        t.Run("Injection: "+script, func(t *testing.T) {
            mod := &Mod{
                Scripts: []string{script},
                Permissions: ModPermissions{
                    SystemAccess: false,
                },
            }
            
            err := ValidateModSecurity(mod)
            assert.ErrorIs(t, err, ErrMaliciousScript)
        })
    }
}
```

## 6. パフォーマンス・負荷テスト

### 6.1 ゲームループパフォーマンス

#### TC-013: 60FPS保証テスト
```go
// Test: 60FPS維持確認
func TestSixtyFPSMaintenance(t *testing.T) {
    game := core.NewGame()
    
    const targetFPS = 60
    const testDuration = 5 * time.Second
    const frameBudget = time.Second / targetFPS // 16.67ms
    
    start := time.Now()
    frameCount := 0
    maxFrameTime := time.Duration(0)
    
    for time.Since(start) < testDuration {
        frameStart := time.Now()
        
        err := game.Update()
        assert.NoError(t, err)
        
        frameTime := time.Since(frameStart)
        if frameTime > maxFrameTime {
            maxFrameTime = frameTime
        }
        
        frameCount++
        
        // フレーム制限シミュレーション
        if frameTime < frameBudget {
            time.Sleep(frameBudget - frameTime)
        }
    }
    
    actualFPS := float64(frameCount) / testDuration.Seconds()
    
    // 実測FPS > 58 (余裕を持たせる)
    assert.Greater(t, actualFPS, 58.0)
    
    // 最大フレーム時間 < 20ms
    assert.Less(t, maxFrameTime, 20*time.Millisecond)
    
    t.Logf("Actual FPS: %.2f, Max frame time: %v", actualFPS, maxFrameTime)
}

// Test: メモリ使用量監視
func TestMemoryUsageStability(t *testing.T) {
    game := core.NewGame()
    
    // 初期メモリ使用量
    var initialMem runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&initialMem)
    
    // 1000フレーム実行
    for i := 0; i < 1000; i++ {
        game.Update()
        
        if i%100 == 0 {
            runtime.GC() // 定期的にGC実行
        }
    }
    
    // 最終メモリ使用量
    var finalMem runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&finalMem)
    
    // メモリ増加量
    memIncrease := finalMem.Alloc - initialMem.Alloc
    
    // メモリ増加 < 10MB
    assert.Less(t, memIncrease, uint64(10*1024*1024))
    
    t.Logf("Memory increase: %d bytes", memIncrease)
}
```

#### TC-014: 負荷ストレステスト
```go
// Test: 大量エンティティ負荷テスト
func TestMassEntityStressTest(t *testing.T) {
    em := NewEntityManager()
    sm := NewSystemManager()
    
    // 1000エンティティ作成
    entities := make([]EntityID, 1000)
    for i := 0; i < 1000; i++ {
        entity := em.CreateEntity()
        entities[i] = entity
        
        em.AddComponent(entity, &TransformComponent{
            X: rand.Float64() * 1280,
            Y: rand.Float64() * 720,
        })
        em.AddComponent(entity, &SpriteComponent{
            Width: 32, Height: 32,
        })
    }
    
    // システム実行時間測定
    start := time.Now()
    for i := 0; i < 60; i++ { // 60フレーム
        err := sm.UpdateSystems(0.016, em)
        assert.NoError(t, err)
    }
    elapsed := time.Since(start)
    
    // 60フレーム処理 < 1秒
    assert.Less(t, elapsed, time.Second)
    
    avgFrameTime := elapsed / 60
    t.Logf("Average frame time with 1000 entities: %v", avgFrameTime)
}

// Test: メモリリーク検出
func TestMemoryLeakDetection(t *testing.T) {
    if testing.Short() {
        t.Skip("長時間テストのためスキップ")
    }
    
    game := core.NewGame()
    
    var memStats []uint64
    
    // 10秒間実行、1秒ごとにメモリ使用量記録
    for i := 0; i < 10; i++ {
        start := time.Now()
        for time.Since(start) < time.Second {
            game.Update()
        }
        
        var m runtime.MemStats
        runtime.GC()
        runtime.ReadMemStats(&m)
        memStats = append(memStats, m.Alloc)
    }
    
    // メモリ使用量の傾向分析
    trend := calculateMemoryTrend(memStats)
    
    // メモリ増加傾向 < 1MB/秒
    assert.Less(t, trend, 1024*1024.0)
    
    t.Logf("Memory trend: %.2f bytes/second", trend)
}
```

## 7. セキュリティテスト

### 7.1 入力検証テスト

#### TC-015: 入力サニタイゼーション
```go
// Test: 設定値インジェクション対策
func TestConfigInjectionPrevention(t *testing.T) {
    maliciousInputs := []struct {
        field string
        value string
    }{
        {"title", "<script>alert('XSS')</script>"},
        {"title", "'; DROP TABLE users; --"},
        {"version", "../../../etc/passwd"},
        {"width", "-1; system('rm -rf /')"},
    }
    
    for _, input := range maliciousInputs {
        t.Run(fmt.Sprintf("%s:%s", input.field, input.value), func(t *testing.T) {
            config := &GameConfig{}
            
            err := SetConfigField(config, input.field, input.value)
            
            // 不正入力は拒否される
            assert.Error(t, err)
            assert.ErrorIs(t, err, ErrInvalidInput)
        })
    }
}

// Test: ファイルパストラバーサル対策
func TestPathTraversalPrevention(t *testing.T) {
    am := NewAssetManager()
    
    maliciousPaths := []string{
        "../../../etc/passwd",
        "..\\..\\..\\windows\\system32\\config\\sam",
        "/etc/shadow",
        "assets/../../../secret.txt",
        "assets\\..\\..\\..\\secret.txt",
    }
    
    for _, path := range maliciousPaths {
        t.Run("Path: "+path, func(t *testing.T) {
            _, err := am.LoadImage(path)
            assert.ErrorIs(t, err, ErrInvalidPath)
        })
    }
}
```

#### TC-016: バッファオーバーフロー対策
```go
// Test: 大きなアセットファイル処理
func TestLargeAssetHandling(t *testing.T) {
    am := NewAssetManager()
    
    // 100MB画像ファイル作成試行
    largePath := createLargeTestImage(100 * 1024 * 1024) // 100MB
    defer os.Remove(largePath)
    
    _, err := am.LoadImage(largePath)
    
    // サイズ制限でエラーになることを確認
    assert.ErrorIs(t, err, ErrAssetTooLarge)
}

// Test: 長い文字列入力処理
func TestLongStringHandling(t *testing.T) {
    // 10MB文字列作成
    longString := strings.Repeat("A", 10*1024*1024)
    
    config := &GameConfig{}
    err := SetConfigField(config, "title", longString)
    
    // 長すぎる文字列は拒否
    assert.ErrorIs(t, err, ErrStringTooLong)
}
```

## 8. E2Eテスト・統合テスト

### 8.1 ゲーム起動・終了テスト

#### TC-017: クロスプラットフォーム起動テスト
```bash
#!/bin/bash
# Test: 全プラットフォームビルド・起動テスト

# Windows
GOOS=windows GOARCH=amd64 go build -o dist/muscle-dreamer.exe cmd/game/main.go
test $? -eq 0 || exit 1

# Linux
GOOS=linux GOARCH=amd64 go build -o dist/muscle-dreamer cmd/game/main.go
test $? -eq 0 || exit 1

# macOS
GOOS=darwin GOARCH=amd64 go build -o dist/muscle-dreamer-mac cmd/game/main.go
test $? -eq 0 || exit 1

# WebAssembly
GOOS=js GOARCH=wasm go build -o dist/web/game.wasm cmd/game/main.go
test $? -eq 0 || exit 1

echo "All platform builds successful"
```

#### TC-018: WebAssembly統合テスト
```javascript
// Test: ブラウザでのWebAssembly動作確認
describe('WebAssembly Integration', () => {
    let page;
    
    beforeEach(async () => {
        page = await browser.newPage();
        await page.goto('http://localhost:3000');
    });
    
    it('WebAssemblyゲームが正常に起動する', async () => {
        // ローディング画面の確認
        await page.waitForSelector('#loadingStatus');
        expect(await page.textContent('#loadingStatus')).toContain('WebAssembly モジュールを読み込み中');
        
        // ゲーム画面の表示待ち
        await page.waitForSelector('#gameCanvas', { state: 'visible', timeout: 10000 });
        
        // キャンバスサイズ確認
        const canvas = await page.$('#gameCanvas');
        const width = await canvas.getAttribute('width');
        const height = await canvas.getAttribute('height');
        
        expect(width).toBe('1280');
        expect(height).toBe('720');
    });
    
    it('エラーハンドリングが正常に動作する', async () => {
        // 不正なWASMファイル読み込み
        await page.route('**/game.wasm', route => {
            route.fulfill({ status: 404 });
        });
        
        await page.goto('http://localhost:3000');
        
        // エラーメッセージ表示確認
        await page.waitForSelector('#errorMessage', { state: 'visible' });
        expect(await page.textContent('#errorMessage')).toContain('読み込みに失敗');
    });
});
```

### 8.2 長時間動作テスト

#### TC-019: 長時間安定性テスト
```go
// Test: 24時間動作テスト (CI/CD専用)
func TestLongTermStability(t *testing.T) {
    if os.Getenv("LONG_TERM_TEST") != "true" {
        t.Skip("長時間テストは環境変数 LONG_TERM_TEST=true で有効化")
    }
    
    game := core.NewGame()
    
    // 24時間 = 86400秒 = 60FPS * 86400 = 5,184,000フレーム
    const totalFrames = 5184000
    const reportInterval = 300000 // 5分ごとにレポート
    
    startTime := time.Now()
    var errors []error
    
    for frame := 0; frame < totalFrames; frame++ {
        err := game.Update()
        if err != nil {
            errors = append(errors, err)
            
            // エラー数が閾値を超えたら停止
            if len(errors) > 100 {
                t.Fatalf("Too many errors: %d", len(errors))
            }
        }
        
        // 定期レポート
        if frame%reportInterval == 0 {
            elapsed := time.Since(startTime)
            progress := float64(frame) / float64(totalFrames) * 100
            
            t.Logf("Progress: %.2f%%, Elapsed: %v, Errors: %d", 
                progress, elapsed, len(errors))
        }
        
        // フレーム制限
        time.Sleep(16670 * time.Microsecond) // ~60FPS
    }
    
    totalTime := time.Since(startTime)
    t.Logf("24-hour test completed in %v with %d errors", totalTime, len(errors))
    
    // エラー率 < 0.01%
    errorRate := float64(len(errors)) / float64(totalFrames)
    assert.Less(t, errorRate, 0.0001)
}
```

## 9. テスト実装優先順位とロードマップ

### 第1フェーズ (即座実装 - 1週間)
1. **TC-001-004**: 基本機能テスト
2. **TC-013**: パフォーマンステスト基盤
3. **TC-017**: ビルドテスト

### 第2フェーズ (ECS実装後 - 2週間)
1. **TC-005-006**: ECSアーキテクチャテスト
2. **TC-009-010**: アセット・テーマテスト
3. **TC-014**: 負荷テスト

### 第3フェーズ (拡張機能実装後 - 3週間)
1. **TC-007-008**: WebAssemblyテスト
2. **TC-011-012**: MODセキュリティテスト
3. **TC-015-016**: セキュリティテスト

### 第4フェーズ (品質保証 - 継続)
1. **TC-018**: 統合E2Eテスト
2. **TC-019**: 長時間安定性テスト
3. **継続的性能監視**

## 10. CI/CDパイプライン統合

### GitHub Actions設定例
```yaml
name: Muscle Dreamer Test Suite

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: 1.22
      - run: make test-unit
      - run: make test-coverage
      
  performance-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: 1.22
      - run: make test-performance
      
  cross-platform-build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: 1.22
      - run: make build-all
      
  wasm-integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: 18
      - uses: actions/setup-go@v3
        with:
          go-version: 1.22
      - run: make build-web
      - run: cd web && npm install && npm test
      
  security-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: 1.22
      - run: make test-security
      
  long-term-test:
    runs-on: ubuntu-latest
    if: github.event_name == 'schedule'
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: 1.22
      - run: LONG_TERM_TEST=true make test-long-term
```

## まとめ

このテスト仕様書は、Muscle Dreamerプロジェクトの特性に特化した包括的なテスト戦略を提供します：

### 🎯 重要な特徴
- **ゲーム特化**: 60FPS、リアルタイム性、状態管理
- **セキュリティ重視**: MODサンドボックス、入力検証
- **クロスプラットフォーム**: ネイティブ+WebAssembly対応
- **拡張性**: ECS、プラグインシステム対応

### 📊 実装効果
- **品質保証**: 42テストケースで包括的カバレッジ
- **パフォーマンス**: 60FPS保証、メモリリーク検出
- **セキュリティ**: MOD攻撃防止、入力サニタイゼーション
- **保守性**: 自動化されたCI/CDパイプライン

### 🚀 次のステップ
1. 第1フェーズテスト実装 (基本機能)
2. ECSアーキテクチャ実装と対応テスト
3. セキュリティ・パフォーマンステスト強化
4. 継続的品質監視システム構築