# TASK-003: 基本コンポーネント実装 - テストケース定義

## テストケース概要

基本コンポーネント5種類（Transform, Sprite, Physics, Health, AI）の単体テストケースを定義します。各コンポーネントに対して機能テスト、性能テスト、エラーハンドリングテストを実施します。

## 1. TransformComponent テストケース

### 1.1 基本機能テスト

#### Test_TransformComponent_CreateAndInitialize
**目的**: コンポーネントの作成と初期化
```go
func Test_TransformComponent_CreateAndInitialize(t *testing.T) {
    // Arrange & Act
    transform := NewTransformComponent()
    
    // Assert
    assert.Equal(t, ComponentTypeTransform, transform.GetType())
    assert.Equal(t, Vector2{0, 0}, transform.Position)
    assert.Equal(t, 0.0, transform.Rotation)
    assert.Equal(t, Vector2{1, 1}, transform.Scale)
    assert.Nil(t, transform.Parent)
    assert.Empty(t, transform.Children)
}
```

#### Test_TransformComponent_SetPosition
**目的**: 位置設定の正常動作
```go
func Test_TransformComponent_SetPosition(t *testing.T) {
    // Arrange
    transform := NewTransformComponent()
    newPos := Vector2{10.5, -20.3}
    
    // Act
    transform.SetPosition(newPos)
    
    // Assert
    assert.Equal(t, newPos, transform.Position)
    assert.Equal(t, newPos, transform.GetWorldPosition())
}
```

#### Test_TransformComponent_WorldLocalConversion
**目的**: ワールド座標とローカル座標の変換
```go
func Test_TransformComponent_WorldLocalConversion(t *testing.T) {
    // Arrange
    parent := NewTransformComponent()
    parent.SetPosition(Vector2{10, 10})
    parent.SetRotation(math.Pi / 4)
    
    child := NewTransformComponent()
    child.SetPosition(Vector2{5, 0})
    child.SetParent(parent)
    
    // Act
    worldPos := child.GetWorldPosition()
    localPos := child.GetLocalPosition()
    
    // Assert
    assert.Equal(t, Vector2{5, 0}, localPos)
    assert.NotEqual(t, localPos, worldPos)
    // 回転と移動を考慮した座標
    expectedX := 10 + 5*math.Cos(math.Pi/4)
    expectedY := 10 + 5*math.Sin(math.Pi/4)
    assert.InDelta(t, expectedX, worldPos.X, 0.001)
    assert.InDelta(t, expectedY, worldPos.Y, 0.001)
}
```

#### Test_TransformComponent_HierarchyManagement
**目的**: 親子関係の管理
```go
func Test_TransformComponent_HierarchyManagement(t *testing.T) {
    // Arrange
    parent := NewTransformComponent()
    child1 := NewTransformComponent()
    child2 := NewTransformComponent()
    
    // Act
    child1.SetParent(parent)
    child2.SetParent(parent)
    
    // Assert
    assert.Equal(t, parent, child1.Parent)
    assert.Equal(t, parent, child2.Parent)
    assert.Len(t, parent.Children, 2)
    assert.Contains(t, parent.Children, child1)
    assert.Contains(t, parent.Children, child2)
}
```

### 1.2 性能テスト

#### Benchmark_TransformComponent_MatrixCalculation
**目的**: 行列計算の性能測定
```go
func Benchmark_TransformComponent_MatrixCalculation(b *testing.B) {
    transform := NewTransformComponent()
    transform.SetPosition(Vector2{100, 200})
    transform.SetRotation(math.Pi / 3)
    transform.SetScale(Vector2{2, 3})
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = transform.GetTransformMatrix()
    }
}
```

### 1.3 エラーハンドリングテスト

#### Test_TransformComponent_CircularParentReference
**目的**: 循環参照の検出
```go
func Test_TransformComponent_CircularParentReference(t *testing.T) {
    // Arrange
    parent := NewTransformComponent()
    child := NewTransformComponent()
    child.SetParent(parent)
    
    // Act & Assert
    err := parent.SetParent(child)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "circular reference")
}
```

## 2. SpriteComponent テストケース

### 2.1 基本機能テスト

#### Test_SpriteComponent_CreateAndInitialize
```go
func Test_SpriteComponent_CreateAndInitialize(t *testing.T) {
    // Arrange & Act
    sprite := NewSpriteComponent()
    
    // Assert
    assert.Equal(t, ComponentTypeSprite, sprite.GetType())
    assert.Empty(t, sprite.TextureID)
    assert.Equal(t, AABB{}, sprite.SourceRect)
    assert.Equal(t, Color{255, 255, 255, 255}, sprite.Color)
    assert.Equal(t, 0, sprite.ZOrder)
    assert.True(t, sprite.Visible)
    assert.False(t, sprite.FlipX)
    assert.False(t, sprite.FlipY)
}
```

#### Test_SpriteComponent_SetTexture
```go
func Test_SpriteComponent_SetTexture(t *testing.T) {
    // Arrange
    sprite := NewSpriteComponent()
    textureID := "player_texture"
    sourceRect := AABB{Min: Vector2{0, 0}, Max: Vector2{32, 32}}
    
    // Act
    sprite.SetTexture(textureID, sourceRect)
    
    // Assert
    assert.Equal(t, textureID, sprite.TextureID)
    assert.Equal(t, sourceRect, sprite.SourceRect)
}
```

### 2.2 描画順序テスト

#### Test_SpriteComponent_ZOrderSorting
```go
func Test_SpriteComponent_ZOrderSorting(t *testing.T) {
    // Arrange
    sprites := []*SpriteComponent{
        NewSpriteComponent(),
        NewSpriteComponent(),
        NewSpriteComponent(),
    }
    sprites[0].ZOrder = 10
    sprites[1].ZOrder = 5
    sprites[2].ZOrder = 15
    
    // Act
    sort.Slice(sprites, func(i, j int) bool {
        return sprites[i].ZOrder < sprites[j].ZOrder
    })
    
    // Assert
    assert.Equal(t, 5, sprites[0].ZOrder)
    assert.Equal(t, 10, sprites[1].ZOrder)
    assert.Equal(t, 15, sprites[2].ZOrder)
}
```

## 3. PhysicsComponent テストケース

### 3.1 基本機能テスト

#### Test_PhysicsComponent_CreateAndInitialize
```go
func Test_PhysicsComponent_CreateAndInitialize(t *testing.T) {
    // Arrange & Act
    physics := NewPhysicsComponent()
    
    // Assert
    assert.Equal(t, ComponentTypePhysics, physics.GetType())
    assert.Equal(t, Vector2{0, 0}, physics.Velocity)
    assert.Equal(t, Vector2{0, 0}, physics.Acceleration)
    assert.Equal(t, 1.0, physics.Mass)
    assert.Equal(t, 0.0, physics.Friction)
    assert.False(t, physics.Gravity)
    assert.False(t, physics.IsStatic)
    assert.Equal(t, math.Inf(1), physics.MaxSpeed)
}
```

#### Test_PhysicsComponent_ApplyForce
```go
func Test_PhysicsComponent_ApplyForce(t *testing.T) {
    // Arrange
    physics := NewPhysicsComponent()
    physics.Mass = 2.0
    force := Vector2{10, 0}
    deltaTime := 0.016 // 60 FPS
    
    // Act
    physics.ApplyForce(force, deltaTime)
    
    // Assert
    expectedAccel := Vector2{5, 0} // F = ma, a = F/m = 10/2
    assert.Equal(t, expectedAccel, physics.Acceleration)
}
```

#### Test_PhysicsComponent_UpdateVelocity
```go
func Test_PhysicsComponent_UpdateVelocity(t *testing.T) {
    // Arrange
    physics := NewPhysicsComponent()
    physics.Acceleration = Vector2{5, 0}
    deltaTime := 0.016
    
    // Act
    physics.UpdateVelocity(deltaTime)
    
    // Assert
    expectedVelocity := Vector2{0.08, 0} // v = a * t = 5 * 0.016
    assert.InDelta(t, expectedVelocity.X, physics.Velocity.X, 0.001)
    assert.InDelta(t, expectedVelocity.Y, physics.Velocity.Y, 0.001)
}
```

### 3.2 物理計算テスト

#### Test_PhysicsComponent_FrictionApplication
```go
func Test_PhysicsComponent_FrictionApplication(t *testing.T) {
    // Arrange
    physics := NewPhysicsComponent()
    physics.Velocity = Vector2{10, 0}
    physics.Friction = 0.1
    deltaTime := 0.016
    
    // Act
    physics.ApplyFriction(deltaTime)
    
    // Assert
    assert.Less(t, physics.Velocity.X, 10.0)
    assert.GreaterOrEqual(t, physics.Velocity.X, 0.0)
}
```

## 4. HealthComponent テストケース

### 4.1 基本機能テスト

#### Test_HealthComponent_CreateAndInitialize
```go
func Test_HealthComponent_CreateAndInitialize(t *testing.T) {
    // Arrange & Act
    health := NewHealthComponent(100)
    
    // Assert
    assert.Equal(t, ComponentTypeHealth, health.GetType())
    assert.Equal(t, 100, health.CurrentHealth)
    assert.Equal(t, 100, health.MaxHealth)
    assert.Equal(t, 0, health.Shield)
    assert.False(t, health.IsInvincible)
    assert.Zero(t, health.LastDamageTime)
    assert.Equal(t, 0.0, health.RegenerationRate)
    assert.Empty(t, health.StatusEffects)
}
```

#### Test_HealthComponent_TakeDamage
```go
func Test_HealthComponent_TakeDamage(t *testing.T) {
    // Arrange
    health := NewHealthComponent(100)
    damage := 25
    
    // Act
    actualDamage := health.TakeDamage(damage)
    
    // Assert
    assert.Equal(t, 75, health.CurrentHealth)
    assert.Equal(t, 25, actualDamage)
    assert.NotZero(t, health.LastDamageTime)
}
```

#### Test_HealthComponent_TakeDamageWithShield
```go
func Test_HealthComponent_TakeDamageWithShield(t *testing.T) {
    // Arrange
    health := NewHealthComponent(100)
    health.Shield = 30
    damage := 50
    
    // Act
    actualDamage := health.TakeDamage(damage)
    
    // Assert
    assert.Equal(t, 80, health.CurrentHealth) // 50 - 30 shield = 20 damage
    assert.Equal(t, 0, health.Shield)
    assert.Equal(t, 20, actualDamage)
}
```

### 4.2 体力回復テスト

#### Test_HealthComponent_Regeneration
```go
func Test_HealthComponent_Regeneration(t *testing.T) {
    // Arrange
    health := NewHealthComponent(100)
    health.CurrentHealth = 50
    health.RegenerationRate = 5.0 // 5 HP per second
    deltaTime := 0.5 // 0.5 seconds
    
    // Act
    health.UpdateRegeneration(deltaTime)
    
    // Assert
    expectedHealth := 52.5
    assert.InDelta(t, expectedHealth, float64(health.CurrentHealth), 0.1)
}
```

## 5. AIComponent テストケース

### 5.1 基本機能テスト

#### Test_AIComponent_CreateAndInitialize
```go
func Test_AIComponent_CreateAndInitialize(t *testing.T) {
    // Arrange & Act
    ai := NewAIComponent()
    
    // Assert
    assert.Equal(t, ComponentTypeAI, ai.GetType())
    assert.Equal(t, AIStateIdle, ai.State)
    assert.Equal(t, InvalidEntityID, ai.Target)
    assert.Empty(t, ai.PatrolPoints)
    assert.Equal(t, 50.0, ai.DetectionRadius)
    assert.Equal(t, 10.0, ai.AttackRange)
    assert.Equal(t, 100.0, ai.Speed)
    assert.Equal(t, AIBehaviorNeutral, ai.Behavior)
    assert.Zero(t, ai.LastStateChange)
}
```

#### Test_AIComponent_StateTransition
```go
func Test_AIComponent_StateTransition(t *testing.T) {
    // Arrange
    ai := NewAIComponent()
    assert.Equal(t, AIStateIdle, ai.State)
    
    // Act
    ai.SetState(AIStatePatrol)
    
    // Assert
    assert.Equal(t, AIStatePatrol, ai.State)
    assert.NotZero(t, ai.LastStateChange)
}
```

### 5.2 AI行動テスト

#### Test_AIComponent_PatrolBehavior
```go
func Test_AIComponent_PatrolBehavior(t *testing.T) {
    // Arrange
    ai := NewAIComponent()
    patrolPoints := []Vector2{
        {0, 0}, {100, 0}, {100, 100}, {0, 100},
    }
    ai.SetPatrolPoints(patrolPoints)
    
    // Act
    ai.SetState(AIStatePatrol)
    nextPoint := ai.GetNextPatrolPoint()
    
    // Assert
    assert.Equal(t, AIStatePatrol, ai.State)
    assert.Equal(t, patrolPoints[0], nextPoint)
}
```

## 6. 統合テスト

### Test_Components_Serialization
```go
func Test_Components_Serialization(t *testing.T) {
    components := []Component{
        NewTransformComponent(),
        NewSpriteComponent(),
        NewPhysicsComponent(),
        NewHealthComponent(100),
        NewAIComponent(),
    }
    
    for _, component := range components {
        // Act
        data, err := component.Serialize()
        assert.NoError(t, err)
        assert.NotEmpty(t, data)
        
        // Create new component and deserialize
        newComponent := createComponentByType(component.GetType())
        err = newComponent.Deserialize(data)
        assert.NoError(t, err)
        
        // Verify data integrity
        newData, err := newComponent.Serialize()
        assert.NoError(t, err)
        assert.Equal(t, data, newData)
    }
}
```

### Test_Components_MemoryUsage
```go
func Test_Components_MemoryUsage(t *testing.T) {
    components := map[ComponentType]Component{
        ComponentTypeTransform: NewTransformComponent(),
        ComponentTypeSprite:    NewSpriteComponent(),
        ComponentTypePhysics:   NewPhysicsComponent(),
        ComponentTypeHealth:    NewHealthComponent(100),
        ComponentTypeAI:        NewAIComponent(),
    }
    
    expectedSizes := map[ComponentType]int{
        ComponentTypeTransform: 40,
        ComponentTypeSprite:    64,
        ComponentTypePhysics:   56,
        ComponentTypeHealth:    72,
        ComponentTypeAI:        96,
    }
    
    for componentType, component := range components {
        actualSize := component.Size()
        expectedSize := expectedSizes[componentType]
        assert.LessOrEqual(t, actualSize, expectedSize,
            "Component %s size %d exceeds limit %d", componentType, actualSize, expectedSize)
    }
}
```

## 7. ベンチマークテスト

### Benchmark_Components_Creation
```go
func Benchmark_Components_Creation(b *testing.B) {
    benchmarks := []struct {
        name    string
        factory func() Component
    }{
        {"Transform", func() Component { return NewTransformComponent() }},
        {"Sprite", func() Component { return NewSpriteComponent() }},
        {"Physics", func() Component { return NewPhysicsComponent() }},
        {"Health", func() Component { return NewHealthComponent(100) }},
        {"AI", func() Component { return NewAIComponent() }},
    }
    
    for _, bm := range benchmarks {
        b.Run(bm.name, func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                _ = bm.factory()
            }
        })
    }
}
```

### Benchmark_Components_Serialization
```go
func Benchmark_Components_Serialization(b *testing.B) {
    components := []Component{
        NewTransformComponent(),
        NewSpriteComponent(),
        NewPhysicsComponent(),
        NewHealthComponent(100),
        NewAIComponent(),
    }
    
    for _, component := range components {
        b.Run(string(component.GetType()), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                data, _ := component.Serialize()
                _ = data
            }
        })
    }
}
```

## テスト実行計画

### テストファイル構成
- `internal/core/ecs/components/transform_test.go`
- `internal/core/ecs/components/sprite_test.go`
- `internal/core/ecs/components/physics_test.go`
- `internal/core/ecs/components/health_test.go`
- `internal/core/ecs/components/ai_test.go`
- `internal/core/ecs/components/integration_test.go`

### テスト実行コマンド
```bash
# 全テスト実行
go test ./internal/core/ecs/components/... -v

# カバレッジ測定
go test ./internal/core/ecs/components/... -cover -coverprofile=coverage.out

# ベンチマーク実行
go test ./internal/core/ecs/components/... -bench=. -benchmem

# レースコンディション検出
go test ./internal/core/ecs/components/... -race
```

## 成功基準

- [ ] 全テストケースが成功（PASS）
- [ ] コードカバレッジ > 95%
- [ ] ベンチマーク要件を満たす
- [ ] レースコンディション検出なし
- [ ] メモリリーク検出なし