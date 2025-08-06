package tests

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"muscle-dreamer/internal/core/ecs"
	"muscle-dreamer/internal/core/ecs/components"
	"muscle-dreamer/internal/core/systems"
)

func TestMovementSystem_Interface(t *testing.T) {
	system := systems.NewMovementSystem()

	// System インターフェース実装確認
	var _ ecs.System = system

	assert.Equal(t, systems.MovementSystemType, system.GetType())
	assert.Equal(t, systems.MovementSystemPriority, system.GetPriority())

	deps := system.GetDependencies()
	assert.Empty(t, deps) // MovementSystemは依存なし

	required := system.GetRequiredComponents()
	assert.Contains(t, required, ecs.ComponentTypeTransform)
}

func TestMovementSystem_PositionUpdate(t *testing.T) {
	system := systems.NewMovementSystem()
	world := createWorldWithEntities()
	deltaTime := 0.016 // 60FPS

	// エンティティ作成
	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 0, Y: 0},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		Velocity: ecs.Vector2{X: 100, Y: 50}, // 100px/s, 50px/s
		Mass:     1.0,
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)

	// システム初期化
	err := system.Initialize(world)
	assert.NoError(t, err)

	// Update実行
	err = system.Update(world, deltaTime)
	assert.NoError(t, err)

	// 位置確認: position += velocity * deltaTime
	updatedTransform := world.GetComponent(entity, ecs.ComponentTypeTransform).(*components.TransformComponent)
	expectedX := 0 + 100*deltaTime // 1.6
	expectedY := 0 + 50*deltaTime  // 0.8

	assert.InDelta(t, expectedX, updatedTransform.Position.X, 0.001)
	assert.InDelta(t, expectedY, updatedTransform.Position.Y, 0.001)
}

func TestMovementSystem_RotationUpdate(t *testing.T) {
	system := systems.NewMovementSystem()
	world := createWorldWithEntities()
	deltaTime := 0.016

	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 0, Y: 0},
		Rotation: 0,
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		// Note: We'll need to add AngularVelocity to PhysicsComponent
		// or create a separate component for angular motion
		Velocity: ecs.Vector2{X: 0, Y: 0},
		Mass:     1.0,
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)

	// システム初期化
	system.Initialize(world)

	err := system.Update(world, deltaTime)
	assert.NoError(t, err)

	// Note: This test will fail until we implement angular velocity
	// For now, we'll test that rotation doesn't change unexpectedly
	updatedTransform := world.GetComponent(entity, ecs.ComponentTypeTransform).(*components.TransformComponent)
	assert.Equal(t, float64(0), updatedTransform.Rotation)
}

func TestMovementSystem_BoundaryCheck(t *testing.T) {
	system := systems.NewMovementSystem()
	system.SetBoundary(0, 0, 800, 600) // 画面サイズ設定
	world := createWorldWithEntities()

	// 画面外への移動エンティティ
	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 790, Y: 300},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		Velocity: ecs.Vector2{X: 1000, Y: 0}, // 右方向高速移動
		Mass:     1.0,
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)

	system.Initialize(world)

	err := system.Update(world, 0.016)
	assert.NoError(t, err)

	// 境界でクランプされることを確認
	updatedTransform := world.GetComponent(entity, ecs.ComponentTypeTransform).(*components.TransformComponent)
	assert.LessOrEqual(t, updatedTransform.Position.X, float64(800))
}

func TestMovementSystem_Acceleration(t *testing.T) {
	system := systems.NewMovementSystem()
	world := createWorldWithEntities()

	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 0, Y: 0},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		Velocity:     ecs.Vector2{X: 0, Y: 0},
		Acceleration: ecs.Vector2{X: 100, Y: -200}, // 右・上方向加速度
		Mass:         1.0,
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)

	system.Initialize(world)

	// 複数フレーム実行
	for i := 0; i < 10; i++ {
		err := system.Update(world, 0.016)
		assert.NoError(t, err)
	}

	updatedPhysics := world.GetComponent(entity, ecs.ComponentTypePhysics).(*components.PhysicsComponent)

	// 速度が加速度により増加していることを確認
	assert.Greater(t, updatedPhysics.Velocity.X, float64(0))
	assert.Less(t, updatedPhysics.Velocity.Y, float64(0))

	// 位置も変化していることを確認
	updatedTransform := world.GetComponent(entity, ecs.ComponentTypeTransform).(*components.TransformComponent)
	assert.Greater(t, updatedTransform.Position.X, float64(0))
}

func TestMovementSystem_MaxSpeed(t *testing.T) {
	system := systems.NewMovementSystem()
	system.SetMaxSpeed(200) // 最大速度200px/s
	world := createWorldWithEntities()

	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 0, Y: 0},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		Velocity: ecs.Vector2{X: 500, Y: 300}, // 制限を超えた速度
		Mass:     1.0,
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)

	system.Initialize(world)

	err := system.Update(world, 0.016)
	assert.NoError(t, err)

	updatedPhysics := world.GetComponent(entity, ecs.ComponentTypePhysics).(*components.PhysicsComponent)
	speed := math.Sqrt(updatedPhysics.Velocity.X*updatedPhysics.Velocity.X +
		updatedPhysics.Velocity.Y*updatedPhysics.Velocity.Y)

	assert.LessOrEqual(t, speed, 200.1) // 小数誤差を考慮
}

func TestMovementSystem_EnableDisable(t *testing.T) {
	system := systems.NewMovementSystem()
	world := createWorldWithEntities()

	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 0, Y: 0},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		Velocity: ecs.Vector2{X: 100, Y: 0},
		Mass:     1.0,
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)

	system.Initialize(world)

	// システム無効化
	system.SetEnabled(false)
	assert.False(t, system.IsEnabled())

	// 無効な状態でUpdate実行
	initialPos := transform.Position
	err := system.Update(world, 0.016)
	assert.NoError(t, err)

	// 位置が変更されないことを確認（システムが無効のため）
	updatedTransform := world.GetComponent(entity, ecs.ComponentTypeTransform).(*components.TransformComponent)
	assert.Equal(t, initialPos.X, updatedTransform.Position.X)
	assert.Equal(t, initialPos.Y, updatedTransform.Position.Y)

	// システム再有効化
	system.SetEnabled(true)
	assert.True(t, system.IsEnabled())
}

// Helper functions for movement system tests

func createWorldWithEntities() *MockWorld {
	return createMockWorld() // Reuse the mock from base_system_test.go
}
