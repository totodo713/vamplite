package components

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"muscle-dreamer/internal/core/ecs"
)

func Test_PhysicsComponent_CreateAndInitialize(t *testing.T) {
	// Arrange & Act
	physics := NewPhysicsComponent()

	// Assert
	assert.Equal(t, ecs.ComponentTypePhysics, physics.GetType())
	assert.Equal(t, ecs.Vector2{X: 0, Y: 0}, physics.Velocity)
	assert.Equal(t, ecs.Vector2{X: 0, Y: 0}, physics.Acceleration)
	assert.Equal(t, 1.0, physics.Mass)
	assert.Equal(t, 0.0, physics.Friction)
	assert.False(t, physics.Gravity)
	assert.False(t, physics.IsStatic)
	assert.Equal(t, 10000.0, physics.MaxSpeed)
}

func Test_PhysicsComponent_ApplyForce(t *testing.T) {
	// Arrange
	physics := NewPhysicsComponent()
	physics.Mass = 2.0
	force := ecs.Vector2{X: 10, Y: 0}
	deltaTime := 0.016 // 60 FPS

	// Act
	physics.ApplyForce(force, deltaTime)

	// Assert
	expectedAccel := ecs.Vector2{X: 5, Y: 0} // F = ma, a = F/m = 10/2
	assert.Equal(t, expectedAccel, physics.Acceleration)
}

func Test_PhysicsComponent_UpdateVelocity(t *testing.T) {
	// Arrange
	physics := NewPhysicsComponent()
	physics.Acceleration = ecs.Vector2{X: 5, Y: 0}
	deltaTime := 0.016

	// Act
	physics.UpdateVelocity(deltaTime)

	// Assert
	expectedVelocity := ecs.Vector2{X: 0.08, Y: 0} // v = a * t = 5 * 0.016
	assert.InDelta(t, expectedVelocity.X, physics.Velocity.X, 0.001)
	assert.InDelta(t, expectedVelocity.Y, physics.Velocity.Y, 0.001)
}

func Test_PhysicsComponent_FrictionApplication(t *testing.T) {
	// Arrange
	physics := NewPhysicsComponent()
	physics.Velocity = ecs.Vector2{X: 10, Y: 0}
	physics.Friction = 0.1
	deltaTime := 0.016

	// Act
	physics.ApplyFriction(deltaTime)

	// Assert
	assert.Less(t, physics.Velocity.X, 10.0)
	assert.GreaterOrEqual(t, physics.Velocity.X, 0.0)
}

func Test_PhysicsComponent_MaxSpeedLimit(t *testing.T) {
	// Arrange
	physics := NewPhysicsComponent()
	physics.MaxSpeed = 50.0
	physics.Velocity = ecs.Vector2{X: 100, Y: 0} // Exceeds max speed

	// Act
	physics.ApplySpeedLimit()

	// Assert
	velocity := math.Sqrt(float64(physics.Velocity.X*physics.Velocity.X + physics.Velocity.Y*physics.Velocity.Y))
	assert.LessOrEqual(t, velocity, physics.MaxSpeed+0.001)
}

func Test_PhysicsComponent_StaticObject(t *testing.T) {
	// Arrange
	physics := NewPhysicsComponent()
	physics.IsStatic = true
	originalVelocity := physics.Velocity

	// Act
	physics.ApplyForce(ecs.Vector2{X: 100, Y: 100}, 0.016)

	// Assert - static objects should not move
	assert.Equal(t, originalVelocity, physics.Velocity)
}

func Test_PhysicsComponent_GravityApplication(t *testing.T) {
	// Arrange
	physics := NewPhysicsComponent()
	physics.Gravity = true
	deltaTime := 0.016
	gravityForce := ecs.Vector2{X: 0, Y: -9.81} // Standard gravity

	// Act
	physics.ApplyGravity(gravityForce, deltaTime)

	// Assert
	expectedAccelY := gravityForce.Y // a = F/m, mass = 1
	assert.Equal(t, expectedAccelY, physics.Acceleration.Y)
}

func Test_PhysicsComponent_Serialization(t *testing.T) {
	// Arrange
	physics := NewPhysicsComponent()
	physics.Velocity = ecs.Vector2{X: 5, Y: 10}
	physics.Mass = 2.5
	physics.Friction = 0.3
	physics.Gravity = true
	physics.MaxSpeed = 100.0

	// Act
	data, err := physics.Serialize()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Create new component and deserialize
	newPhysics := NewPhysicsComponent()
	err = newPhysics.Deserialize(data)
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, physics.Velocity, newPhysics.Velocity)
	assert.Equal(t, physics.Mass, newPhysics.Mass)
	assert.Equal(t, physics.Friction, newPhysics.Friction)
	assert.Equal(t, physics.Gravity, newPhysics.Gravity)
	assert.Equal(t, physics.MaxSpeed, newPhysics.MaxSpeed)
}

func Test_PhysicsComponent_Clone(t *testing.T) {
	// Arrange
	original := NewPhysicsComponent()
	original.Mass = 3.0
	original.Friction = 0.5

	// Act
	cloned := original.Clone()

	// Assert
	assert.NotSame(t, original, cloned)
	clonedPhysics := cloned.(*PhysicsComponent)
	assert.Equal(t, original.Mass, clonedPhysics.Mass)
	assert.Equal(t, original.Friction, clonedPhysics.Friction)
}

func Test_PhysicsComponent_Validate(t *testing.T) {
	// Arrange
	physics := NewPhysicsComponent()

	// Act & Assert - valid state
	err := physics.Validate()
	assert.NoError(t, err)

	// Invalid mass (negative)
	physics.Mass = -1.0
	err = physics.Validate()
	assert.Error(t, err)

	// Invalid friction (negative)
	physics.Mass = 1.0
	physics.Friction = -0.5
	err = physics.Validate()
	assert.Error(t, err)
}

func Test_PhysicsComponent_Size(t *testing.T) {
	// Arrange
	physics := NewPhysicsComponent()

	// Act
	size := physics.Size()

	// Assert
	assert.LessOrEqual(t, size, 56, "PhysicsComponent size should be <= 56 bytes")
	assert.Greater(t, size, 0, "PhysicsComponent size should be > 0")
}

func Benchmark_PhysicsComponent_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewPhysicsComponent()
	}
}

func Benchmark_PhysicsComponent_ForceApplication(b *testing.B) {
	physics := NewPhysicsComponent()
	force := ecs.Vector2{X: 10, Y: 5}
	deltaTime := 0.016

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		physics.ApplyForce(force, deltaTime)
	}
}
