package tests

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"muscle-dreamer/internal/core/ecs"
	"muscle-dreamer/internal/core/ecs/components"
	"muscle-dreamer/internal/core/systems"
)

func TestPhysicsSystem_Interface(t *testing.T) {
	system := systems.NewPhysicsSystem()

	var _ ecs.System = system

	assert.Equal(t, systems.PhysicsSystemType, system.GetType())

	required := system.GetRequiredComponents()
	assert.Contains(t, required, ecs.ComponentTypeTransform)
	assert.Contains(t, required, ecs.ComponentTypePhysics)
}

func TestPhysicsSystem_Gravity(t *testing.T) {
	system := systems.NewPhysicsSystem()
	system.SetGravity(ecs.Vector2{X: 0, Y: 980}) // 重力 9.8m/s²
	world := createWorldWithEntities()

	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 0, Y: 0},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		Mass:     1.0,
		Velocity: ecs.Vector2{X: 0, Y: 0},
		Gravity:  true, // 重力の影響を受ける
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)

	system.Initialize(world)

	// 1秒間シミュレーション
	for i := 0; i < 60; i++ { // 60FPS
		err := system.Update(world, 0.016)
		assert.NoError(t, err)
	}

	updatedPhysics := world.GetComponent(entity, ecs.ComponentTypePhysics).(*components.PhysicsComponent)
	updatedTransform := world.GetComponent(entity, ecs.ComponentTypeTransform).(*components.TransformComponent)

	// 重力により下方向速度が増加
	assert.Less(t, updatedPhysics.Velocity.Y, float64(-500))

	// 落下により位置が下方向に移動
	assert.Less(t, updatedTransform.Position.Y, float64(-1000))
}

func TestPhysicsSystem_GravityDisabled(t *testing.T) {
	system := systems.NewPhysicsSystem()
	system.SetGravity(ecs.Vector2{X: 0, Y: 980})
	world := createWorldWithEntities()

	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 0, Y: 0},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		Mass:     1.0,
		Velocity: ecs.Vector2{X: 0, Y: 0},
		Gravity:  false, // 重力の影響を受けない
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)

	system.Initialize(world)

	initialVelocity := physics.Velocity

	// 複数フレーム実行
	for i := 0; i < 10; i++ {
		err := system.Update(world, 0.016)
		assert.NoError(t, err)
	}

	updatedPhysics := world.GetComponent(entity, ecs.ComponentTypePhysics).(*components.PhysicsComponent)

	// 重力が無効なので速度は変化しない
	assert.Equal(t, initialVelocity.Y, updatedPhysics.Velocity.Y)
}

func TestPhysicsSystem_StaticObjects(t *testing.T) {
	system := systems.NewPhysicsSystem()
	system.SetGravity(ecs.Vector2{X: 0, Y: 980})
	world := createWorldWithEntities()

	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 100, Y: 100},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		Mass:     1.0,
		Velocity: ecs.Vector2{X: 100, Y: 100},
		IsStatic: true, // 静的オブジェクト
		Gravity:  true,
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)

	system.Initialize(world)

	initialPosition := transform.Position
	initialVelocity := physics.Velocity

	// 複数フレーム実行
	for i := 0; i < 10; i++ {
		err := system.Update(world, 0.016)
		assert.NoError(t, err)
	}

	updatedPhysics := world.GetComponent(entity, ecs.ComponentTypePhysics).(*components.PhysicsComponent)
	updatedTransform := world.GetComponent(entity, ecs.ComponentTypeTransform).(*components.TransformComponent)

	// 静的オブジェクトなので位置と速度は変化しない
	assert.Equal(t, initialPosition.X, updatedTransform.Position.X)
	assert.Equal(t, initialPosition.Y, updatedTransform.Position.Y)
	assert.Equal(t, initialVelocity.X, updatedPhysics.Velocity.X)
	assert.Equal(t, initialVelocity.Y, updatedPhysics.Velocity.Y)
}

func TestPhysicsSystem_MaxSpeed(t *testing.T) {
	system := systems.NewPhysicsSystem()
	world := createWorldWithEntities()

	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 0, Y: 0},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		Mass:     1.0,
		Velocity: ecs.Vector2{X: 1000, Y: 1000}, // 非常に高い速度
		MaxSpeed: 300,                           // 最大速度制限
		IsStatic: false,
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)

	system.Initialize(world)

	err := system.Update(world, 0.016)
	assert.NoError(t, err)

	updatedPhysics := world.GetComponent(entity, ecs.ComponentTypePhysics).(*components.PhysicsComponent)
	speed := math.Sqrt(updatedPhysics.Velocity.X*updatedPhysics.Velocity.X +
		updatedPhysics.Velocity.Y*updatedPhysics.Velocity.Y)

	// 最大速度制限が適用される
	assert.LessOrEqual(t, speed, physics.MaxSpeed+0.1) // 小数誤差を考慮
}

func TestPhysicsSystem_FrictionAndDrag(t *testing.T) {
	system := systems.NewPhysicsSystem()
	world := createWorldWithEntities()

	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 0, Y: 0},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		Mass:     1.0,
		Velocity: ecs.Vector2{X: 100, Y: 100},
		Friction: 0.9, // 高い摩擦
		Gravity:  false,
		IsStatic: false,
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)

	system.Initialize(world)

	initialSpeed := math.Sqrt(physics.Velocity.X*physics.Velocity.X +
		physics.Velocity.Y*physics.Velocity.Y)

	// 複数フレーム実行（摩擦により減速）
	for i := 0; i < 60; i++ {
		err := system.Update(world, 0.016)
		assert.NoError(t, err)
	}

	updatedPhysics := world.GetComponent(entity, ecs.ComponentTypePhysics).(*components.PhysicsComponent)
	finalSpeed := math.Sqrt(updatedPhysics.Velocity.X*updatedPhysics.Velocity.X +
		updatedPhysics.Velocity.Y*updatedPhysics.Velocity.Y)

	// 摩擦により速度が減少
	assert.Less(t, finalSpeed, initialSpeed)
}

func TestPhysicsSystem_StaticColliders(t *testing.T) {
	system := systems.NewPhysicsSystem()
	world := createWorldWithEntities()

	// 静的コライダー（地面）を追加
	groundBounds := systems.Rectangle{X: 0, Y: 500, Width: 800, Height: 100}
	system.AddStaticCollider(groundBounds)

	// 落下するオブジェクト
	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 400, Y: 0},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		Mass:     1.0,
		Velocity: ecs.Vector2{X: 0, Y: 0},
		Gravity:  true,
		IsStatic: false,
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)

	system.Initialize(world)
	system.SetGravity(ecs.Vector2{X: 0, Y: 500}) // 下向き重力

	// 地面に衝突するまでシミュレーション
	for i := 0; i < 120; i++ { // 2秒間
		err := system.Update(world, 0.016)
		assert.NoError(t, err)

		// 衝突チェック
		collisions := system.GetCollisions()
		if len(collisions) > 0 {
			// 衝突が発生した
			break
		}
	}

	// 静的コライダーが正しく追加されていることを確認
	staticColliders := system.GetStaticColliders()
	assert.Len(t, staticColliders, 1)
	assert.Equal(t, groundBounds, staticColliders[0].Bounds)
}

func TestPhysicsSystem_FixedTimeStep(t *testing.T) {
	system := systems.NewPhysicsSystem()

	// デフォルトの固定タイムステップを確認
	defaultTimeStep := system.GetFixedTimeStep()
	assert.Equal(t, 1.0/60.0, defaultTimeStep) // 60Hz

	// タイムステップ変更
	newTimeStep := 1.0 / 120.0 // 120Hz
	system.SetFixedTimeStep(newTimeStep)

	assert.Equal(t, newTimeStep, system.GetFixedTimeStep())
}

func TestPhysicsSystem_Acceleration(t *testing.T) {
	system := systems.NewPhysicsSystem()
	world := createWorldWithEntities()

	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 0, Y: 0},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		Mass:         1.0,
		Velocity:     ecs.Vector2{X: 0, Y: 0},
		Acceleration: ecs.Vector2{X: 50, Y: 100},
		Gravity:      false, // 重力を無効にして加速度のみテスト
		IsStatic:     false,
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)

	system.Initialize(world)

	// 1秒間の加速度適用
	for i := 0; i < 60; i++ {
		err := system.Update(world, 0.016)
		assert.NoError(t, err)
	}

	updatedPhysics := world.GetComponent(entity, ecs.ComponentTypePhysics).(*components.PhysicsComponent)

	// 加速度により速度が増加
	assert.Greater(t, updatedPhysics.Velocity.X, float64(40)) // 50 * 0.016 * 60 = 48
	assert.Greater(t, updatedPhysics.Velocity.Y, float64(80)) // 100 * 0.016 * 60 = 96
}

func TestPhysicsSystem_CollisionDetection_AABB(t *testing.T) {
	system := systems.NewPhysicsSystem()
	world := createWorldWithEntities()

	// 移動オブジェクト
	movingEntity := world.CreateEntity()
	movingTransform := &components.TransformComponent{
		Position: ecs.Vector2{X: 0, Y: 100},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	movingPhysics := &components.PhysicsComponent{
		Mass:     1.0,
		Velocity: ecs.Vector2{X: 200, Y: 0}, // 右方向移動
		IsStatic: false,
	}
	world.AddComponent(movingEntity, movingTransform)
	world.AddComponent(movingEntity, movingPhysics)

	// 静的オブジェクト（障害物）として静的コライダーを追加
	obstacleRect := systems.Rectangle{X: 100, Y: 90, Width: 20, Height: 20}
	system.AddStaticCollider(obstacleRect)

	system.Initialize(world)

	// 衝突が発生するまでシミュレーション
	collisionDetected := false
	for i := 0; i < 30; i++ {
		system.ClearCollisions() // 前フレームの衝突をクリア
		err := system.Update(world, 0.016)
		assert.NoError(t, err)

		// 衝突イベントチェック
		collisions := system.GetCollisions()
		if len(collisions) > 0 {
			collisionDetected = true
			assert.Equal(t, movingEntity, collisions[0].EntityA)
			break
		}
	}

	assert.True(t, collisionDetected, "衝突が検出されませんでした")
}

func TestPhysicsSystem_CollisionClearance(t *testing.T) {
	system := systems.NewPhysicsSystem()

	// 衝突データを手動で追加
	collision := systems.Collision{
		EntityA:      ecs.EntityID(1),
		EntityB:      ecs.EntityID(2),
		ContactPoint: ecs.Vector2{X: 50, Y: 50},
		Normal:       ecs.Vector2{X: 1, Y: 0},
		Depth:        5.0,
		Timestamp:    12345,
	}

	// Note: このテストは実装時に適切なメソッドで衝突を追加する必要がある
	// 現在は GetCollisions() が空の配列を返すことを確認
	collisions := system.GetCollisions()
	assert.Empty(t, collisions)

	// Clear実行
	system.ClearCollisions()
	collisions = system.GetCollisions()
	assert.Empty(t, collisions)
}
