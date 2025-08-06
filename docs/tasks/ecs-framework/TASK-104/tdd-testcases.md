# TASK-104: 基本システム実装 - テストケース仕様

## テストケース概要

TASK-104の4つの基本システム（Movement, Rendering, Physics, Audio）およびBaseSystemの包括的なテストケースを定義します。TDD方式により、まずテストを実装し、その後にシステムを実装します。

## 1. BaseSystem テストケース

### 1.1 基本機能テスト

#### TestBaseSystem_Initialize
```go
// 基本初期化テスト
func TestBaseSystem_Initialize(t *testing.T) {
    system := createTestBaseSystem()
    world := createMockWorld()
    
    err := system.Initialize(world)
    
    assert.NoError(t, err)
    assert.True(t, system.IsEnabled())
    assert.NotNil(t, system.GetMetrics())
}
```

#### TestBaseSystem_GetType
```go
// システムタイプ取得テスト
func TestBaseSystem_GetType(t *testing.T) {
    system := createTestBaseSystem()
    
    systemType := system.GetType()
    
    assert.Equal(t, TestSystemType, systemType)
    assert.NotEmpty(t, string(systemType))
}
```

#### TestBaseSystem_Priority
```go
// プライオリティ管理テスト
func TestBaseSystem_Priority(t *testing.T) {
    system := createTestBaseSystem()
    expectedPriority := Priority(100)
    system.SetPriority(expectedPriority)
    
    priority := system.GetPriority()
    
    assert.Equal(t, expectedPriority, priority)
}
```

### 1.2 状態管理テスト

#### TestBaseSystem_EnableDisable
```go
// 有効/無効状態管理テスト
func TestBaseSystem_EnableDisable(t *testing.T) {
    system := createTestBaseSystem()
    
    // 初期状態: 有効
    assert.True(t, system.IsEnabled())
    
    // 無効化
    system.SetEnabled(false)
    assert.False(t, system.IsEnabled())
    
    // 再有効化
    system.SetEnabled(true)
    assert.True(t, system.IsEnabled())
}
```

### 1.3 メトリクス機能テスト

#### TestBaseSystem_Metrics
```go
// メトリクス収集テスト
func TestBaseSystem_Metrics(t *testing.T) {
    system := createTestBaseSystem()
    world := createMockWorld()
    system.Initialize(world)
    
    // Update実行前
    metrics := system.GetMetrics()
    assert.Equal(t, int64(0), metrics.UpdateCount)
    assert.Equal(t, time.Duration(0), metrics.TotalUpdateTime)
    
    // Update実行
    err := system.Update(world, 0.016) // 60FPS
    assert.NoError(t, err)
    
    // Update実行後
    metrics = system.GetMetrics()
    assert.Equal(t, int64(1), metrics.UpdateCount)
    assert.Greater(t, metrics.TotalUpdateTime, time.Duration(0))
    assert.Greater(t, metrics.AverageUpdateTime, time.Duration(0))
}
```

## 2. MovementSystem テストケース

### 2.1 基本機能テスト

#### TestMovementSystem_Interface
```go
// System インターフェース実装確認
func TestMovementSystem_Interface(t *testing.T) {
    system := NewMovementSystem()
    
    // System インターフェース実装確認
    var _ System = system
    
    assert.Equal(t, MovementSystemType, system.GetType())
    assert.Equal(t, MovementSystemPriority, system.GetPriority())
    
    deps := system.GetDependencies()
    assert.Empty(t, deps) // MovementSystemは依存なし
    
    required := system.GetRequiredComponents()
    assert.Contains(t, required, TransformComponentType)
}
```

#### TestMovementSystem_PositionUpdate
```go
// 位置更新処理テスト
func TestMovementSystem_PositionUpdate(t *testing.T) {
    system := NewMovementSystem()
    world := createWorldWithEntities()
    deltaTime := 0.016 // 60FPS
    
    // エンティティ作成
    entity := world.CreateEntity()
    transform := &TransformComponent{
        Position: Vector2{X: 0, Y: 0},
        Velocity: Vector2{X: 100, Y: 50}, // 100px/s, 50px/s
    }
    world.AddComponent(entity, transform)
    
    // Update実行
    err := system.Update(world, deltaTime)
    assert.NoError(t, err)
    
    // 位置確認: position += velocity * deltaTime
    updatedTransform := world.GetComponent(entity, TransformComponentType).(*TransformComponent)
    expectedX := 0 + 100*deltaTime // 1.6
    expectedY := 0 + 50*deltaTime  // 0.8
    
    assert.InDelta(t, expectedX, updatedTransform.Position.X, 0.001)
    assert.InDelta(t, expectedY, updatedTransform.Position.Y, 0.001)
}
```

#### TestMovementSystem_RotationUpdate
```go
// 回転更新処理テスト
func TestMovementSystem_RotationUpdate(t *testing.T) {
    system := NewMovementSystem()
    world := createWorldWithEntities()
    deltaTime := 0.016
    
    entity := world.CreateEntity()
    transform := &TransformComponent{
        Rotation:        0,
        AngularVelocity: math.Pi, // 180度/秒
    }
    world.AddComponent(entity, transform)
    
    err := system.Update(world, deltaTime)
    assert.NoError(t, err)
    
    updatedTransform := world.GetComponent(entity, TransformComponentType).(*TransformComponent)
    expectedRotation := math.Pi * deltaTime
    
    assert.InDelta(t, expectedRotation, updatedTransform.Rotation, 0.001)
}
```

### 2.2 境界チェックテスト

#### TestMovementSystem_BoundaryCheck
```go
// 境界チェック機能テスト
func TestMovementSystem_BoundaryCheck(t *testing.T) {
    system := NewMovementSystem()
    system.SetBoundary(0, 0, 800, 600) // 画面サイズ設定
    world := createWorldWithEntities()
    
    // 画面外への移動エンティティ
    entity := world.CreateEntity()
    transform := &TransformComponent{
        Position: Vector2{X: 790, Y: 300},
        Velocity: Vector2{X: 1000, Y: 0}, // 右方向高速移動
    }
    world.AddComponent(entity, transform)
    
    err := system.Update(world, 0.016)
    assert.NoError(t, err)
    
    // 境界でクランプされることを確認
    updatedTransform := world.GetComponent(entity, TransformComponentType).(*TransformComponent)
    assert.LessOrEqual(t, updatedTransform.Position.X, float64(800))
}
```

### 2.3 加速度・速度制限テスト

#### TestMovementSystem_Acceleration
```go
// 加速度適用テスト
func TestMovementSystem_Acceleration(t *testing.T) {
    system := NewMovementSystem()
    world := createWorldWithEntities()
    
    entity := world.CreateEntity()
    transform := &TransformComponent{
        Position:     Vector2{X: 0, Y: 0},
        Velocity:     Vector2{X: 0, Y: 0},
        Acceleration: Vector2{X: 100, Y: -200}, // 右・上方向加速度
    }
    world.AddComponent(entity, transform)
    
    // 複数フレーム実行
    for i := 0; i < 10; i++ {
        err := system.Update(world, 0.016)
        assert.NoError(t, err)
    }
    
    updatedTransform := world.GetComponent(entity, TransformComponentType).(*TransformComponent)
    
    // 速度が加速度により増加していることを確認
    assert.Greater(t, updatedTransform.Velocity.X, float64(0))
    assert.Less(t, updatedTransform.Velocity.Y, float64(0))
    
    // 位置も変化していることを確認
    assert.Greater(t, updatedTransform.Position.X, float64(0))
}
```

#### TestMovementSystem_MaxSpeed
```go
// 最大速度制限テスト
func TestMovementSystem_MaxSpeed(t *testing.T) {
    system := NewMovementSystem()
    system.SetMaxSpeed(200) // 最大速度200px/s
    world := createWorldWithEntities()
    
    entity := world.CreateEntity()
    transform := &TransformComponent{
        Velocity: Vector2{X: 500, Y: 300}, // 制限を超えた速度
    }
    world.AddComponent(entity, transform)
    
    err := system.Update(world, 0.016)
    assert.NoError(t, err)
    
    updatedTransform := world.GetComponent(entity, TransformComponentType).(*TransformComponent)
    speed := math.Sqrt(updatedTransform.Velocity.X*updatedTransform.Velocity.X + 
                      updatedTransform.Velocity.Y*updatedTransform.Velocity.Y)
    
    assert.LessOrEqual(t, speed, 200.1) // 小数誤差を考慮
}
```

## 3. RenderingSystem テストケース

### 3.1 基本機能テスト

#### TestRenderingSystem_Interface
```go
// System インターフェース実装確認
func TestRenderingSystem_Interface(t *testing.T) {
    system := NewRenderingSystem()
    
    var _ System = system
    
    assert.Equal(t, RenderingSystemType, system.GetType())
    assert.Equal(t, RenderingSystemPriority, system.GetPriority())
    
    required := system.GetRequiredComponents()
    assert.Contains(t, required, TransformComponentType)
    assert.Contains(t, required, SpriteComponentType)
}
```

#### TestRenderingSystem_BasicRendering
```go
// 基本描画処理テスト
func TestRenderingSystem_BasicRendering(t *testing.T) {
    system := NewRenderingSystem()
    world := createWorldWithEntities()
    mockRenderer := &MockRenderer{}
    
    // 描画エンティティ作成
    entity := world.CreateEntity()
    transform := &TransformComponent{
        Position: Vector2{X: 100, Y: 200},
        Scale:    Vector2{X: 2, Y: 2},
        Rotation: math.Pi / 4, // 45度回転
    }
    sprite := &SpriteComponent{
        TextureID: "player_sprite",
        Width:     32,
        Height:    32,
    }
    world.AddComponent(entity, transform)
    world.AddComponent(entity, sprite)
    
    // Render実行
    err := system.Render(world, mockRenderer)
    assert.NoError(t, err)
    
    // モックレンダラーが呼ばれたことを確認
    assert.Equal(t, 1, mockRenderer.DrawCallCount)
    assert.Equal(t, "player_sprite", mockRenderer.LastTexture)
}
```

### 3.2 Z-Order描画順序テスト

#### TestRenderingSystem_ZOrder
```go
// Z-Order描画順序テスト
func TestRenderingSystem_ZOrder(t *testing.T) {
    system := NewRenderingSystem()
    world := createWorldWithEntities()
    mockRenderer := &MockRenderer{}
    
    // 異なるZ-Orderのエンティティ作成
    entities := []EntityID{}
    zOrders := []int{3, 1, 5, 2} // 描画順序: 1, 2, 3, 5
    
    for i, z := range zOrders {
        entity := world.CreateEntity()
        transform := &TransformComponent{Position: Vector2{X: float64(i * 50), Y: 0}}
        sprite := &SpriteComponent{ZOrder: z, TextureID: fmt.Sprintf("sprite_%d", i)}
        world.AddComponent(entity, transform)
        world.AddComponent(entity, sprite)
        entities = append(entities, entity)
    }
    
    err := system.Render(world, mockRenderer)
    assert.NoError(t, err)
    
    // 描画順序確認
    expectedOrder := []string{"sprite_1", "sprite_3", "sprite_0", "sprite_2"}
    assert.Equal(t, expectedOrder, mockRenderer.DrawOrder)
}
```

### 3.3 カリングテスト

#### TestRenderingSystem_ViewportCulling
```go
// ビューポートカリングテスト
func TestRenderingSystem_ViewportCulling(t *testing.T) {
    system := NewRenderingSystem()
    system.SetViewport(0, 0, 800, 600)
    world := createWorldWithEntities()
    mockRenderer := &MockRenderer{}
    
    // 画面内エンティティ
    visibleEntity := world.CreateEntity()
    visibleTransform := &TransformComponent{Position: Vector2{X: 400, Y: 300}}
    visibleSprite := &SpriteComponent{TextureID: "visible", Width: 32, Height: 32}
    world.AddComponent(visibleEntity, visibleTransform)
    world.AddComponent(visibleEntity, visibleSprite)
    
    // 画面外エンティティ
    hiddenEntity := world.CreateEntity()
    hiddenTransform := &TransformComponent{Position: Vector2{X: -100, Y: -100}}
    hiddenSprite := &SpriteComponent{TextureID: "hidden", Width: 32, Height: 32}
    world.AddComponent(hiddenEntity, hiddenTransform)
    world.AddComponent(hiddenEntity, hiddenSprite)
    
    err := system.Render(world, mockRenderer)
    assert.NoError(t, err)
    
    // 画面内のみ描画されることを確認
    assert.Equal(t, 1, mockRenderer.DrawCallCount)
    assert.Equal(t, "visible", mockRenderer.LastTexture)
}
```

## 4. PhysicsSystem テストケース

### 4.1 基本機能テスト

#### TestPhysicsSystem_Interface
```go
// System インターフェース実装確認
func TestPhysicsSystem_Interface(t *testing.T) {
    system := NewPhysicsSystem()
    
    var _ System = system
    
    assert.Equal(t, PhysicsSystemType, system.GetType())
    
    required := system.GetRequiredComponents()
    assert.Contains(t, required, TransformComponentType)
    assert.Contains(t, required, PhysicsComponentType)
}
```

#### TestPhysicsSystem_Gravity
```go
// 重力適用テスト
func TestPhysicsSystem_Gravity(t *testing.T) {
    system := NewPhysicsSystem()
    system.SetGravity(Vector2{X: 0, Y: 980}) // 重力 9.8m/s²
    world := createWorldWithEntities()
    
    entity := world.CreateEntity()
    transform := &TransformComponent{Position: Vector2{X: 0, Y: 0}}
    physics := &PhysicsComponent{
        Mass:     1.0,
        Velocity: Vector2{X: 0, Y: 0},
    }
    world.AddComponent(entity, transform)
    world.AddComponent(entity, physics)
    
    // 1秒間シミュレーション
    for i := 0; i < 60; i++ { // 60FPS
        err := system.Update(world, 0.016)
        assert.NoError(t, err)
    }
    
    updatedPhysics := world.GetComponent(entity, PhysicsComponentType).(*PhysicsComponent)
    updatedTransform := world.GetComponent(entity, TransformComponentType).(*TransformComponent)
    
    // 重力により下方向速度が増加
    assert.Less(t, updatedPhysics.Velocity.Y, float64(-500))
    
    // 落下により位置が下方向に移動
    assert.Less(t, updatedTransform.Position.Y, float64(-1000))
}
```

### 4.2 衝突検出テスト

#### TestPhysicsSystem_AABB_Collision
```go
// AABB衝突検出テスト
func TestPhysicsSystem_AABB_Collision(t *testing.T) {
    system := NewPhysicsSystem()
    world := createWorldWithEntities()
    
    // 移動オブジェクト
    movingEntity := world.CreateEntity()
    movingTransform := &TransformComponent{
        Position: Vector2{X: 0, Y: 100},
        Velocity: Vector2{X: 200, Y: 0}, // 右方向移動
    }
    movingPhysics := &PhysicsComponent{
        BoundingBox: Rectangle{Width: 32, Height: 32},
        IsDynamic:   true,
    }
    world.AddComponent(movingEntity, movingTransform)
    world.AddComponent(movingEntity, movingPhysics)
    
    // 静的オブジェクト
    staticEntity := world.CreateEntity()
    staticTransform := &TransformComponent{Position: Vector2{X: 100, Y: 100}}
    staticPhysics := &PhysicsComponent{
        BoundingBox: Rectangle{Width: 32, Height: 32},
        IsDynamic:   false,
    }
    world.AddComponent(staticEntity, staticTransform)
    world.AddComponent(staticEntity, staticPhysics)
    
    // 衝突が発生するまでシミュレーション
    collisionDetected := false
    for i := 0; i < 30; i++ {
        err := system.Update(world, 0.016)
        assert.NoError(t, err)
        
        // 衝突イベントチェック
        collisions := system.GetCollisions()
        if len(collisions) > 0 {
            collisionDetected = true
            break
        }
    }
    
    assert.True(t, collisionDetected, "衝突が検出されませんでした")
}
```

### 4.3 物理応答テスト

#### TestPhysicsSystem_Restitution
```go
// 反発係数テスト
func TestPhysicsSystem_Restitution(t *testing.T) {
    system := NewPhysicsSystem()
    world := createWorldWithEntities()
    
    entity := world.CreateEntity()
    transform := &TransformComponent{
        Position: Vector2{X: 100, Y: 0},
        Velocity: Vector2{X: 0, Y: -500}, // 下方向高速移動
    }
    physics := &PhysicsComponent{
        BoundingBox: Rectangle{Width: 32, Height: 32},
        Restitution: 0.8, // 80%反発
        IsDynamic:   true,
    }
    world.AddComponent(entity, transform)
    world.AddComponent(entity, physics)
    
    // 地面との衝突をシミュレーション
    system.AddStaticCollider(Rectangle{X: 0, Y: 500, Width: 800, Height: 100})
    
    // 衝突・反発をシミュレーション
    for i := 0; i < 100; i++ {
        err := system.Update(world, 0.016)
        assert.NoError(t, err)
    }
    
    updatedPhysics := world.GetComponent(entity, PhysicsComponentType).(*PhysicsComponent)
    
    // 反発により上方向速度を持つことを確認
    assert.Greater(t, updatedPhysics.Velocity.Y, float64(0))
    
    // 反発速度は元の速度 * 反発係数程度
    assert.Less(t, updatedPhysics.Velocity.Y, 500*0.9) // 誤差を考慮
}
```

## 5. AudioSystem テストケース

### 5.1 基本機能テスト

#### TestAudioSystem_Interface
```go
// System インターフェース実装確認
func TestAudioSystem_Interface(t *testing.T) {
    system := NewAudioSystem()
    
    var _ System = system
    
    assert.Equal(t, AudioSystemType, system.GetType())
    
    required := system.GetRequiredComponents()
    assert.Contains(t, required, AudioComponentType)
}
```

#### TestAudioSystem_PlaySound
```go
// 音声再生テスト
func TestAudioSystem_PlaySound(t *testing.T) {
    system := NewAudioSystem()
    world := createWorldWithEntities()
    mockAudioEngine := &MockAudioEngine{}
    system.SetAudioEngine(mockAudioEngine)
    
    entity := world.CreateEntity()
    audio := &AudioComponent{
        SoundID:  "jump_sound",
        Volume:   0.8,
        IsPlaying: true,
        IsLoop:   false,
    }
    world.AddComponent(entity, audio)
    
    err := system.Update(world, 0.016)
    assert.NoError(t, err)
    
    // オーディオエンジンに再生要求が送られることを確認
    assert.Equal(t, 1, mockAudioEngine.PlayCallCount)
    assert.Equal(t, "jump_sound", mockAudioEngine.LastSoundID)
    assert.InDelta(t, 0.8, mockAudioEngine.LastVolume, 0.01)
}
```

### 5.2 3Dオーディオテスト

#### TestAudioSystem_3DAudio
```go
// 3Dオーディオ（距離減衰）テスト
func TestAudioSystem_3DAudio(t *testing.T) {
    system := NewAudioSystem()
    system.SetListener(Vector2{X: 0, Y: 0}) // リスナー位置
    world := createWorldWithEntities()
    mockAudioEngine := &MockAudioEngine{}
    system.SetAudioEngine(mockAudioEngine)
    
    // 近距離オーディオソース
    nearEntity := world.CreateEntity()
    nearTransform := &TransformComponent{Position: Vector2{X: 10, Y: 0}}
    nearAudio := &AudioComponent{
        SoundID:     "near_sound",
        Volume:      1.0,
        IsPlaying:   true,
        Is3D:        true,
        MaxDistance: 100,
    }
    world.AddComponent(nearEntity, nearTransform)
    world.AddComponent(nearEntity, nearAudio)
    
    // 遠距離オーディオソース
    farEntity := world.CreateEntity()
    farTransform := &TransformComponent{Position: Vector2{X: 90, Y: 0}}
    farAudio := &AudioComponent{
        SoundID:     "far_sound",
        Volume:      1.0,
        IsPlaying:   true,
        Is3D:        true,
        MaxDistance: 100,
    }
    world.AddComponent(farEntity, farTransform)
    world.AddComponent(farEntity, farAudio)
    
    err := system.Update(world, 0.016)
    assert.NoError(t, err)
    
    // 距離による音量減衰を確認
    assert.Equal(t, 2, mockAudioEngine.PlayCallCount)
    
    // 近距離音源の方が大きい音量で再生される
    assert.Greater(t, mockAudioEngine.VolumeHistory[0], mockAudioEngine.VolumeHistory[1])
}
```

## 6. 統合テストケース

### 6.1 システム連携テスト

#### TestSystemsIntegration_MovementToPhysics
```go
// MovementSystem → PhysicsSystem 連携テスト
func TestSystemsIntegration_MovementToPhysics(t *testing.T) {
    movementSystem := NewMovementSystem()
    physicsSystem := NewPhysicsSystem()
    world := createWorldWithEntities()
    
    // 両システム初期化
    movementSystem.Initialize(world)
    physicsSystem.Initialize(world)
    
    // 移動・物理コンポーネント持ちエンティティ
    entity := world.CreateEntity()
    transform := &TransformComponent{
        Position: Vector2{X: 0, Y: 0},
        Velocity: Vector2{X: 100, Y: 0},
    }
    physics := &PhysicsComponent{
        Mass: 1.0,
        BoundingBox: Rectangle{Width: 32, Height: 32},
        IsDynamic: true,
    }
    world.AddComponent(entity, transform)
    world.AddComponent(entity, physics)
    
    // システム順次実行
    for i := 0; i < 60; i++ {
        err := movementSystem.Update(world, 0.016)
        assert.NoError(t, err)
        
        err = physicsSystem.Update(world, 0.016)
        assert.NoError(t, err)
    }
    
    // 連携動作確認
    updatedTransform := world.GetComponent(entity, TransformComponentType).(*TransformComponent)
    assert.Greater(t, updatedTransform.Position.X, float64(50)) // 移動確認
}
```

### 6.2 パフォーマンステスト

#### TestSystemsPerformance_10000Entities
```go
// 10,000エンティティパフォーマンステスト
func TestSystemsPerformance_10000Entities(t *testing.T) {
    systems := []System{
        NewMovementSystem(),
        NewRenderingSystem(),
        NewPhysicsSystem(),
        NewAudioSystem(),
    }
    
    world := createWorldWithEntities()
    
    // 全システム初期化
    for _, system := range systems {
        err := system.Initialize(world)
        assert.NoError(t, err)
    }
    
    // 10,000エンティティ作成
    for i := 0; i < 10000; i++ {
        entity := world.CreateEntity()
        
        transform := &TransformComponent{
            Position: Vector2{X: rand.Float64() * 800, Y: rand.Float64() * 600},
            Velocity: Vector2{X: rand.Float64()*200 - 100, Y: rand.Float64()*200 - 100},
        }
        sprite := &SpriteComponent{TextureID: fmt.Sprintf("sprite_%d", i%10)}
        
        world.AddComponent(entity, transform)
        world.AddComponent(entity, sprite)
        
        // 50%の確率で物理コンポーネント追加
        if i%2 == 0 {
            physics := &PhysicsComponent{Mass: 1.0, BoundingBox: Rectangle{Width: 16, Height: 16}}
            world.AddComponent(entity, physics)
        }
    }
    
    // パフォーマンス測定
    start := time.Now()
    
    for i := 0; i < 60; i++ { // 1秒間シミュレーション
        for _, system := range systems {
            if system.GetType() == RenderingSystemType {
                continue // Render系は除外（モックレンダラー未設定のため）
            }
            
            err := system.Update(world, 0.016)
            assert.NoError(t, err)
        }
    }
    
    elapsed := time.Since(start)
    
    // パフォーマンス要件：1秒間のシミュレーションが2秒以内で完了
    assert.Less(t, elapsed, 2*time.Second)
    
    // 各システムのメトリクス確認
    for _, system := range systems {
        if system.GetType() == RenderingSystemType {
            continue
        }
        
        metrics := system.GetMetrics()
        avgUpdateTime := metrics.AverageUpdateTime
        
        // 各システムの平均実行時間が10ms以下
        assert.Less(t, avgUpdateTime, 10*time.Millisecond,
            "System %s average update time: %v", system.GetType(), avgUpdateTime)
    }
}
```

## 7. エラーハンドリングテスト

### 7.1 不正データハンドリング

#### TestSystems_InvalidComponents
```go
// 不正コンポーネントハンドリングテスト
func TestSystems_InvalidComponents(t *testing.T) {
    systems := []System{
        NewMovementSystem(),
        NewPhysicsSystem(),
    }
    
    world := createWorldWithEntities()
    
    // 全システム初期化
    for _, system := range systems {
        system.Initialize(world)
    }
    
    // 無効な値を持つコンポーネント作成
    entity := world.CreateEntity()
    invalidTransform := &TransformComponent{
        Position: Vector2{X: math.NaN(), Y: math.Inf(1)}, // NaN, Inf
        Velocity: Vector2{X: math.Inf(-1), Y: math.NaN()},
    }
    world.AddComponent(entity, invalidTransform)
    
    // システム実行してエラーハンドリング確認
    for _, system := range systems {
        err := system.Update(world, 0.016)
        
        // エラーが適切に処理される（パニックしない）
        if err != nil {
            assert.Contains(t, err.Error(), "invalid") // エラーメッセージ確認
        }
    }
}
```

## テスト実装要件

### テストユーティリティ
1. **Mock オブジェクト**: World, Renderer, AudioEngine
2. **テストヘルパー**: エンティティ作成、コンポーネント設定
3. **アサーション**: 数値精度、時間計測、メトリクス確認

### テストデータ
1. **境界値**: 最大/最小座標、速度、音量
2. **異常値**: NaN, Infinity, nil ポインタ
3. **パフォーマンスデータ**: 大量エンティティ、長時間実行

### 実行要件
- **並列実行**: `go test -race` によるデータ競合検出
- **カバレッジ**: `go test -cover` で90%以上
- **ベンチマーク**: `go test -bench` でパフォーマンス測定

---

## まとめ

この詳細なテストケース仕様により、TASK-104の基本システム実装を包括的にテストできます。TDD方式により、まずこれらのテストを実装し、その後にシステムを段階的に実装していきます。

**次のステップ**: テスト実装（Red段階）の開始