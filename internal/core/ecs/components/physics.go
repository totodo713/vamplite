package components

import (
	"encoding/json"
	"errors"
	"math"

	"muscle-dreamer/internal/core/ecs"
)

// PhysicsComponent handles physics simulation parameters
type PhysicsComponent struct {
	Velocity     ecs.Vector2 `json:"velocity"`
	Acceleration ecs.Vector2 `json:"acceleration"`
	Mass         float64     `json:"mass"`
	Friction     float64     `json:"friction"`
	Gravity      bool        `json:"gravity"`
	IsStatic     bool        `json:"isStatic"`
	MaxSpeed     float64     `json:"maxSpeed"`
}

// NewPhysicsComponent creates a new physics component with default values
func NewPhysicsComponent() *PhysicsComponent {
	return &PhysicsComponent{
		Velocity:     ecs.Vector2{X: 0, Y: 0},
		Acceleration: ecs.Vector2{X: 0, Y: 0},
		Mass:         1.0,
		Friction:     0.0,
		Gravity:      false,
		IsStatic:     false,
		MaxSpeed:     10000.0, // Large finite value instead of infinity
	}
}

// GetType returns the component type
func (p *PhysicsComponent) GetType() ecs.ComponentType {
	return ecs.ComponentTypePhysics
}

// ApplyForce applies a force to the physics body
func (p *PhysicsComponent) ApplyForce(force ecs.Vector2, _ float64) {
	if p.IsStatic || p.Mass <= 0 {
		return
	}

	// F = ma, so a = F/m
	p.Acceleration.X = force.X / float32(p.Mass)
	p.Acceleration.Y = force.Y / float32(p.Mass)
}

// UpdateVelocity updates velocity based on acceleration
func (p *PhysicsComponent) UpdateVelocity(deltaTime float64) {
	if p.IsStatic {
		return
	}

	// v = v0 + at
	p.Velocity.X += p.Acceleration.X * float32(deltaTime)
	p.Velocity.Y += p.Acceleration.Y * float32(deltaTime)
}

// ApplyFriction applies friction to the velocity
func (p *PhysicsComponent) ApplyFriction(deltaTime float64) {
	if p.IsStatic || p.Friction <= 0 {
		return
	}

	// Simple friction model: v = v * (1 - friction * dt)
	frictionFactor := 1.0 - (p.Friction * deltaTime)
	if frictionFactor < 0 {
		frictionFactor = 0
	}

	p.Velocity.X *= float32(frictionFactor)
	p.Velocity.Y *= float32(frictionFactor)
}

// ApplySpeedLimit applies maximum speed limitation
func (p *PhysicsComponent) ApplySpeedLimit() {
	if p.IsStatic || math.IsInf(p.MaxSpeed, 1) {
		return
	}

	speed := math.Sqrt(float64(p.Velocity.X*p.Velocity.X + p.Velocity.Y*p.Velocity.Y))
	if speed > p.MaxSpeed {
		// Normalize and scale to max speed
		scale := float32(p.MaxSpeed / speed)
		p.Velocity.X *= scale
		p.Velocity.Y *= scale
	}
}

// ApplyGravity applies gravitational force
func (p *PhysicsComponent) ApplyGravity(gravityForce ecs.Vector2, _ float64) {
	if p.IsStatic || !p.Gravity {
		return
	}

	p.Acceleration.X += gravityForce.X
	p.Acceleration.Y += gravityForce.Y
}

// Clone creates a deep copy of the component
func (p *PhysicsComponent) Clone() ecs.Component {
	return &PhysicsComponent{
		Velocity:     p.Velocity,
		Acceleration: p.Acceleration,
		Mass:         p.Mass,
		Friction:     p.Friction,
		Gravity:      p.Gravity,
		IsStatic:     p.IsStatic,
		MaxSpeed:     p.MaxSpeed,
	}
}

// Validate ensures the component data is valid
func (p *PhysicsComponent) Validate() error {
	if p.Mass < 0 {
		return errors.New("mass cannot be negative")
	}
	if p.Friction < 0 {
		return errors.New("friction cannot be negative")
	}
	if p.MaxSpeed < 0 {
		return errors.New("max speed cannot be negative")
	}
	return nil
}

// Size returns the memory size of the component in bytes
func (p *PhysicsComponent) Size() int {
	// Approximate size calculation
	return 56 // Velocity(16) + Acceleration(16) + Mass(8) + Friction(8) + MaxSpeed(8) + bools
}

// Serialize converts the component to bytes
func (p *PhysicsComponent) Serialize() ([]byte, error) {
	return json.Marshal(p)
}

// Deserialize loads component data from bytes
func (p *PhysicsComponent) Deserialize(data []byte) error {
	return json.Unmarshal(data, p)
}
