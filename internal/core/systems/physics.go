package systems

import (
	"math"

	"muscle-dreamer/internal/core/ecs"
	"muscle-dreamer/internal/core/ecs/components"
)

// PhysicsSystem handles physics simulation including collision detection,
// gravity, and physics responses for entities.
type PhysicsSystem struct {
	*BaseSystem

	// Physics parameters
	gravity         ecs.Vector2
	staticColliders []Collider
	collisions      []Collision
	fixedTimeStep   float64
	accumulator     float64
}

// Collider represents a collision shape.
type Collider struct {
	Bounds    Rectangle
	IsTrigger bool
	Material  PhysicsMaterial
}

// PhysicsMaterial defines physics properties.
type PhysicsMaterial struct {
	Friction    float64
	Restitution float64
	Density     float64
}

// Collision represents a collision event between two entities.
type Collision struct {
	EntityA      ecs.EntityID
	EntityB      ecs.EntityID
	ContactPoint ecs.Vector2
	Normal       ecs.Vector2
	Depth        float64
	Timestamp    int64
}

// NewPhysicsSystem creates a new physics system.
func NewPhysicsSystem() *PhysicsSystem {
	return &PhysicsSystem{
		BaseSystem:      NewBaseSystem(PhysicsSystemType, PhysicsSystemPriority),
		gravity:         ecs.Vector2{X: 0, Y: 9.8 * 100}, // Default gravity (downward)
		staticColliders: make([]Collider, 0),
		collisions:      make([]Collision, 0),
		fixedTimeStep:   1.0 / 60.0, // 60Hz physics
	}
}

// GetRequiredComponents returns the components this system operates on.
func (ps *PhysicsSystem) GetRequiredComponents() []ecs.ComponentType {
	return []ecs.ComponentType{
		ecs.ComponentTypeTransform,
		ecs.ComponentTypePhysics,
	}
}

// Initialize sets up the physics system.
func (ps *PhysicsSystem) Initialize(world ecs.World) error {
	// TODO: Implement initialization
	return ps.BaseSystem.Initialize(world)
}

// Update processes physics simulation for the current frame.
func (ps *PhysicsSystem) Update(world ecs.World, deltaTime float64) error {
	// TODO: Implement physics processing
	return ps.BaseSystem.Update(world, deltaTime)
}

// SetGravity sets the global gravity vector.
func (ps *PhysicsSystem) SetGravity(gravity ecs.Vector2) {
	ps.gravity = gravity
}

// GetGravity returns the current gravity vector.
func (ps *PhysicsSystem) GetGravity() ecs.Vector2 {
	return ps.gravity
}

// AddStaticCollider adds a static collision shape to the world.
func (ps *PhysicsSystem) AddStaticCollider(bounds Rectangle) {
	collider := Collider{
		Bounds:    bounds,
		IsTrigger: false,
		Material: PhysicsMaterial{
			Friction:    0.5,
			Restitution: 0.3,
			Density:     1.0,
		},
	}
	ps.staticColliders = append(ps.staticColliders, collider)
}

// GetStaticColliders returns all static colliders.
func (ps *PhysicsSystem) GetStaticColliders() []Collider {
	return ps.staticColliders
}

// GetCollisions returns collisions detected in the last update.
func (ps *PhysicsSystem) GetCollisions() []Collision {
	return ps.collisions
}

// ClearCollisions clears the collision list.
func (ps *PhysicsSystem) ClearCollisions() {
	ps.collisions = ps.collisions[:0]
}

// SetFixedTimeStep sets the fixed physics timestep.
func (ps *PhysicsSystem) SetFixedTimeStep(timeStep float64) {
	ps.fixedTimeStep = timeStep
}

// GetFixedTimeStep returns the current fixed timestep.
func (ps *PhysicsSystem) GetFixedTimeStep() float64 {
	return ps.fixedTimeStep
}

// checkAABBCollision performs Axis-Aligned Bounding Box collision detection.
func (ps *PhysicsSystem) checkAABBCollision(boundsA, boundsB Rectangle) bool {
	return !(boundsA.X+boundsA.Width < boundsB.X ||
		boundsB.X+boundsB.Width < boundsA.X ||
		boundsA.Y+boundsA.Height < boundsB.Y ||
		boundsB.Y+boundsB.Height < boundsA.Y)
}

// resolveCollision applies collision response between two entities.
func (ps *PhysicsSystem) resolveCollision(collision *Collision, world ecs.World) {
	// TODO: Implement collision resolution
}

// applyGravity applies gravity to a physics component.
func (ps *PhysicsSystem) applyGravity(physics *components.PhysicsComponent, deltaTime float64) {
	if physics.Mass <= 0 || physics.IsStatic {
		return
	}

	// Apply gravitational acceleration
	physics.Velocity.X += ps.gravity.X * deltaTime
	physics.Velocity.Y += ps.gravity.Y * deltaTime
}

// applyDrag applies air resistance/drag to velocity.
func (ps *PhysicsSystem) applyDrag(physics *components.PhysicsComponent, deltaTime float64) {
	dragCoeff := 0.98 // Air resistance coefficient
	physics.Velocity.X *= math.Pow(dragCoeff, deltaTime)
	physics.Velocity.Y *= math.Pow(dragCoeff, deltaTime)
}
