package systems

import (
	"math"

	"muscle-dreamer/internal/core/ecs"
	"muscle-dreamer/internal/core/ecs/components"
)

// MovementSystem handles entity movement and position updates.
// It processes entities with TransformComponent to update positions
// based on velocity, acceleration, and boundary constraints.
type MovementSystem struct {
	*BaseSystem

	// Movement parameters
	maxSpeed float64
	boundary *Rectangle
}

// Rectangle represents a bounding rectangle for movement constraints.
type Rectangle struct {
	X, Y, Width, Height float64
}

// NewMovementSystem creates a new movement system.
func NewMovementSystem() *MovementSystem {
	return &MovementSystem{
		BaseSystem: NewBaseSystem(MovementSystemType, MovementSystemPriority),
		maxSpeed:   -1, // No limit by default
	}
}

// GetRequiredComponents returns the components this system operates on.
func (ms *MovementSystem) GetRequiredComponents() []ecs.ComponentType {
	return []ecs.ComponentType{ecs.ComponentTypeTransform}
}

// Initialize sets up the movement system.
func (ms *MovementSystem) Initialize(world ecs.World) error {
	// TODO: Implement initialization
	return ms.BaseSystem.Initialize(world)
}

// Update processes entity movement for the current frame.
func (ms *MovementSystem) Update(world ecs.World, deltaTime float64) error {
	if !ms.IsEnabled() {
		return nil
	}

	// 必要なコンポーネントを持つエンティティを取得
	result := world.Query().
		With(ecs.ComponentTypeTransform).
		With(ecs.ComponentTypePhysics).
		Execute()

	entities := result.GetEntities()

	for _, entity := range entities {
		// TransformComponent取得
		transformComp, err := world.GetComponent(entity, ecs.ComponentTypeTransform)
		if err != nil {
			continue
		}
		transform, ok := transformComp.(*components.TransformComponent)
		if !ok {
			continue
		}

		// PhysicsComponent取得
		physicsComp, err := world.GetComponent(entity, ecs.ComponentTypePhysics)
		if err != nil {
			continue
		}
		physics, ok := physicsComp.(*components.PhysicsComponent)
		if !ok {
			continue
		}

		// 加速度の適用（velocity += acceleration * deltaTime）
		physics.Velocity.X += physics.Acceleration.X * deltaTime
		physics.Velocity.Y += physics.Acceleration.Y * deltaTime

		// 速度制限の適用
		ms.limitSpeed(&physics.Velocity)

		// 位置の更新（position += velocity * deltaTime）
		transform.Position.X += physics.Velocity.X * deltaTime
		transform.Position.Y += physics.Velocity.Y * deltaTime

		// 回転の更新（PhysicsComponentにAngularVelocityフィールドが必要）
		// 現在のPhysicsComponentにAngularVelocityがないため、TransformComponentの回転は変更しない

		// 境界チェック
		ms.clampToBoundary(&transform.Position)
	}

	return ms.BaseSystem.Update(world, deltaTime)
}

// SetMaxSpeed sets the maximum movement speed limit.
func (ms *MovementSystem) SetMaxSpeed(maxSpeed float64) {
	ms.maxSpeed = maxSpeed
}

// GetMaxSpeed returns the current maximum speed limit.
func (ms *MovementSystem) GetMaxSpeed() float64 {
	return ms.maxSpeed
}

// SetBoundary sets movement boundary constraints.
func (ms *MovementSystem) SetBoundary(x, y, width, height float64) {
	ms.boundary = &Rectangle{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

// GetBoundary returns the current movement boundary.
func (ms *MovementSystem) GetBoundary() *Rectangle {
	return ms.boundary
}

// limitSpeed applies speed constraints to a velocity vector.
func (ms *MovementSystem) limitSpeed(velocity *ecs.Vector2) {
	if ms.maxSpeed <= 0 {
		return
	}

	speed := math.Sqrt(velocity.X*velocity.X + velocity.Y*velocity.Y)
	if speed > ms.maxSpeed {
		scale := ms.maxSpeed / speed
		velocity.X *= scale
		velocity.Y *= scale
	}
}

// clampToBoundary constrains position within boundary limits.
func (ms *MovementSystem) clampToBoundary(position *ecs.Vector2) {
	if ms.boundary == nil {
		return
	}

	if position.X < ms.boundary.X {
		position.X = ms.boundary.X
	} else if position.X > ms.boundary.X+ms.boundary.Width {
		position.X = ms.boundary.X + ms.boundary.Width
	}

	if position.Y < ms.boundary.Y {
		position.Y = ms.boundary.Y
	} else if position.Y > ms.boundary.Y+ms.boundary.Height {
		position.Y = ms.boundary.Y + ms.boundary.Height
	}
}
