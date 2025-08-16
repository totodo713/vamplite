package optimizations

// EntityID represents a unique entity identifier (matching existing definition).
type EntityID uint64

// ComponentType represents a component type identifier.
type ComponentType uint16

// Vector3 represents a 3D vector.
type Vector3 struct {
	X, Y, Z float32
}

// TransformComponent represents position, rotation, and scale
type TransformComponent struct {
	Position Vector3
	Rotation Vector3
	Scale    Vector3
}

// SpriteComponent represents sprite rendering data
type SpriteComponent struct {
	TextureID int32
	Color     Color
	Visible   bool
}

// PhysicsComponent represents physics properties
type PhysicsComponent struct {
	Velocity     Vector3
	Acceleration Vector3
	Mass         float32
}

// Color represents RGBA color values
type Color struct {
	R, G, B, A float32
}

// System represents a game system
type System interface {
	Update(world World, deltaTime float32)
}

// World represents the ECS world interface (minimal for testing)
type World interface {
	CreateEntity() EntityID
	AddComponent(entityID EntityID, component interface{})
	Update(deltaTime float32)
}