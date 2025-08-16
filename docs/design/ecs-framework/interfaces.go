// Package ecs provides the core Entity Component System interfaces for Muscle Dreamer
package ecs

import (
	"context"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

// =============================================================================
// Core Types
// =============================================================================

// EntityID represents a unique identifier for an entity
type EntityID struct {
	Index      uint32 // Index in the entity array
	Generation uint32 // Generation number for entity recycling
}

// ComponentType represents the type identifier for a component
type ComponentType uint64

// SystemType represents the type identifier for a system
type SystemType string

// ComponentMask represents a bitmask for component types
type ComponentMask uint64

// =============================================================================
// Entity Management
// =============================================================================

// EntityManager manages entity lifecycle and component associations
type EntityManager interface {
	// Entity Operations
	CreateEntity() EntityID
	DestroyEntity(entity EntityID) error
	IsEntityValid(entity EntityID) bool
	GetEntityCount() int

	// Component Operations
	AddComponent(entity EntityID, component Component) error
	RemoveComponent(entity EntityID, componentType ComponentType) error
	GetComponent(entity EntityID, componentType ComponentType) (Component, error)
	HasComponent(entity EntityID, componentType ComponentType) bool

	// Batch Operations
	CreateEntities(count int) ([]EntityID, error)
	DestroyEntities(entities []EntityID) error

	// Query Operations
	Query(mask ComponentMask) EntityIterator
	QueryWith(componentTypes ...ComponentType) EntityIterator
	QueryWithout(componentTypes ...ComponentType) EntityIterator

	// Memory Management
	Compact() error
	GetMemoryStats() MemoryStats
}

// Component represents the base interface for all components
type Component interface {
	GetType() ComponentType
	Clone() Component
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}

// MemoryStats provides memory usage information
type MemoryStats struct {
	EntityCount      int
	ComponentCount   int
	MemoryUsed       int64
	MemoryAllocated  int64
	FragmentedMemory int64
	PoolMemoryUsed   int64
	PoolMemoryTotal  int64
}

// =============================================================================
// Component Store
// =============================================================================

// ComponentStore manages component data storage and retrieval
type ComponentStore interface {
	// Component Type Registration
	RegisterComponentType(componentType ComponentType, factory ComponentFactory) error
	IsComponentTypeRegistered(componentType ComponentType) bool
	GetRegisteredComponentTypes() []ComponentType

	// Component Storage
	StoreComponent(entity EntityID, component Component) error
	RetrieveComponent(entity EntityID, componentType ComponentType) (Component, error)
	DeleteComponent(entity EntityID, componentType ComponentType) error

	// Batch Operations
	GetComponentsOfType(componentType ComponentType) ComponentIterator
	GetEntityComponents(entity EntityID) []Component

	// Memory Management
	DefragmentStorage(componentType ComponentType) error
	GetStorageStats(componentType ComponentType) StorageStats
}

// ComponentFactory creates new instances of components
type ComponentFactory interface {
	CreateComponent() Component
	GetComponentType() ComponentType
	GetComponentSize() int
}

// StorageStats provides storage information for a specific component type
type StorageStats struct {
	ComponentType     ComponentType
	ComponentCount    int
	MemoryUsed        int64
	MemoryWasted      int64
	FragmentationRate float64
}

// =============================================================================
// System Management
// =============================================================================

// SystemManager manages system registration and execution
type SystemManager interface {
	// System Registration
	RegisterSystem(system System) error
	UnregisterSystem(systemType SystemType) error
	GetRegisteredSystems() []SystemType
	IsSystemRegistered(systemType SystemType) bool

	// System Execution
	UpdateSystems(ctx context.Context, deltaTime time.Duration) error
	RenderSystems(ctx context.Context, screen *ebiten.Image) error

	// System Dependencies
	SetSystemDependency(dependent, dependency SystemType) error
	RemoveSystemDependency(dependent, dependency SystemType) error
	GetExecutionOrder() []SystemType

	// System State Management
	EnableSystem(systemType SystemType) error
	DisableSystem(systemType SystemType) error
	IsSystemEnabled(systemType SystemType) bool

	// Performance Monitoring
	GetSystemPerformance(systemType SystemType) SystemPerformance
	GetOverallPerformance() OverallPerformance
}

// System represents a game system that operates on entities with specific components
type System interface {
	// System Identity
	GetType() SystemType
	GetRequiredComponents() []ComponentType
	GetOptionalComponents() []ComponentType

	// System Lifecycle
	Initialize(entityManager EntityManager, componentStore ComponentStore) error
	Update(ctx context.Context, deltaTime time.Duration) error
	Cleanup() error

	// System State
	IsEnabled() bool
	SetEnabled(enabled bool)

	// Performance Monitoring
	GetPerformanceMetrics() SystemPerformance
}

// RenderSystem represents a system that performs rendering operations
type RenderSystem interface {
	System
	Render(ctx context.Context, screen *ebiten.Image) error
	GetRenderOrder() int
}

// SystemPerformance provides performance metrics for a system
type SystemPerformance struct {
	SystemType         SystemType
	LastUpdateTime     time.Duration
	AverageUpdateTime  time.Duration
	MaxUpdateTime      time.Duration
	UpdateCount        int64
	ErrorCount         int64
	EntitiesProcessed  int
	ComponentsAccessed int64
}

// OverallPerformance provides overall system performance metrics
type OverallPerformance struct {
	TotalUpdateTime    time.Duration
	SystemCount        int
	EntitiesProcessed  int
	ComponentsAccessed int64
	MemoryUsage        int64
	GCPressure         float64
	FrameTime          time.Duration
	TargetFrameTime    time.Duration
}

// =============================================================================
// Query System
// =============================================================================

// QueryEngine provides efficient entity querying capabilities
type QueryEngine interface {
	// Query Building
	NewQuery() QueryBuilder
	CreateCachedQuery(name string, mask ComponentMask) error
	GetCachedQuery(name string) (EntityIterator, error)

	// Query Execution
	ExecuteQuery(mask ComponentMask) EntityIterator
	ExecuteComplexQuery(query ComplexQuery) EntityIterator

	// Index Management
	RebuildIndices() error
	GetIndexStats() IndexStats
}

// QueryBuilder provides a fluent interface for building complex queries
type QueryBuilder interface {
	With(componentTypes ...ComponentType) QueryBuilder
	Without(componentTypes ...ComponentType) QueryBuilder
	WithAny(componentTypes ...ComponentType) QueryBuilder
	WithAll(componentTypes ...ComponentType) QueryBuilder
	Limit(count int) QueryBuilder
	Execute() EntityIterator
}

// ComplexQuery represents a complex entity query
type ComplexQuery struct {
	RequiredComponents  []ComponentType
	ForbiddenComponents []ComponentType
	AnyOfComponents     []ComponentType
	AllOfComponents     []ComponentType
	Limit               int
	Offset              int
}

// IndexStats provides information about query indices
type IndexStats struct {
	IndexCount       int
	CachedQueryCount int
	IndexMemoryUsage int64
	IndexHitRate     float64
	IndexMissRate    float64
	RebuildCount     int64
}

// =============================================================================
// Iterators
// =============================================================================

// EntityIterator provides iteration over entities
type EntityIterator interface {
	Next() bool
	Entity() EntityID
	Components() []Component
	ComponentsOfType(componentType ComponentType) []Component
	Count() int
	Reset()
	Close() error
}

// ComponentIterator provides iteration over components
type ComponentIterator interface {
	Next() bool
	Component() Component
	Entity() EntityID
	Count() int
	Reset()
	Close() error
}

// =============================================================================
// Predefined Components
// =============================================================================

// TransformComponent manages entity position, rotation, and scale
type TransformComponent struct {
	Position Vector2
	Rotation float64 // Rotation in radians
	Scale    Vector2
}

func (t *TransformComponent) GetType() ComponentType { return TransformComponentType }
func (t *TransformComponent) Clone() Component {
	return &TransformComponent{
		Position: t.Position,
		Rotation: t.Rotation,
		Scale:    t.Scale,
	}
}
func (t *TransformComponent) Serialize() ([]byte, error)    { /* Implementation */ return nil, nil }
func (t *TransformComponent) Deserialize(data []byte) error { /* Implementation */ return nil }

// SpriteComponent manages sprite rendering
type SpriteComponent struct {
	Image      *ebiten.Image
	SourceRect Rectangle
	Color      colorm.ColorM
	Visible    bool
	Layer      int
	FlipX      bool
	FlipY      bool
	Opacity    float64
}

func (s *SpriteComponent) GetType() ComponentType { return SpriteComponentType }
func (s *SpriteComponent) Clone() Component {
	return &SpriteComponent{
		Image:      s.Image,
		SourceRect: s.SourceRect,
		Color:      s.Color,
		Visible:    s.Visible,
		Layer:      s.Layer,
		FlipX:      s.FlipX,
		FlipY:      s.FlipY,
		Opacity:    s.Opacity,
	}
}
func (s *SpriteComponent) Serialize() ([]byte, error)    { /* Implementation */ return nil, nil }
func (s *SpriteComponent) Deserialize(data []byte) error { /* Implementation */ return nil }

// VelocityComponent manages entity movement
type VelocityComponent struct {
	Velocity     Vector2
	MaxSpeed     float64
	Acceleration Vector2
	Friction     float64
}

func (v *VelocityComponent) GetType() ComponentType { return VelocityComponentType }
func (v *VelocityComponent) Clone() Component {
	return &VelocityComponent{
		Velocity:     v.Velocity,
		MaxSpeed:     v.MaxSpeed,
		Acceleration: v.Acceleration,
		Friction:     v.Friction,
	}
}
func (v *VelocityComponent) Serialize() ([]byte, error)    { /* Implementation */ return nil, nil }
func (v *VelocityComponent) Deserialize(data []byte) error { /* Implementation */ return nil }

// HealthComponent manages entity health and damage
type HealthComponent struct {
	Current        int
	Maximum        int
	Regeneration   float64
	Invulnerable   bool
	LastDamageTime time.Time
}

func (h *HealthComponent) GetType() ComponentType { return HealthComponentType }
func (h *HealthComponent) Clone() Component {
	return &HealthComponent{
		Current:        h.Current,
		Maximum:        h.Maximum,
		Regeneration:   h.Regeneration,
		Invulnerable:   h.Invulnerable,
		LastDamageTime: h.LastDamageTime,
	}
}
func (h *HealthComponent) Serialize() ([]byte, error)    { /* Implementation */ return nil, nil }
func (h *HealthComponent) Deserialize(data []byte) error { /* Implementation */ return nil }

// CollisionComponent manages entity collision detection
type CollisionComponent struct {
	Bounds    Rectangle
	Layer     int
	Mask      int
	IsTrigger bool
	IsStatic  bool
	Material  PhysicsMaterial
}

func (c *CollisionComponent) GetType() ComponentType { return CollisionComponentType }
func (c *CollisionComponent) Clone() Component {
	return &CollisionComponent{
		Bounds:    c.Bounds,
		Layer:     c.Layer,
		Mask:      c.Mask,
		IsTrigger: c.IsTrigger,
		IsStatic:  c.IsStatic,
		Material:  c.Material,
	}
}
func (c *CollisionComponent) Serialize() ([]byte, error)    { /* Implementation */ return nil, nil }
func (c *CollisionComponent) Deserialize(data []byte) error { /* Implementation */ return nil }

// =============================================================================
// Utility Types
// =============================================================================

// Vector2 represents a 2D vector
type Vector2 struct {
	X, Y float64
}

// Rectangle represents a 2D rectangle
type Rectangle struct {
	X, Y, Width, Height float64
}

// PhysicsMaterial defines physical properties for collision
type PhysicsMaterial struct {
	Friction    float64
	Restitution float64
	Density     float64
}

// =============================================================================
// Component Type Constants
// =============================================================================

const (
	TransformComponentType ComponentType = 1 << iota
	SpriteComponentType
	VelocityComponentType
	HealthComponentType
	CollisionComponentType
	InputComponentType
	AudioComponentType
	AIComponentType
	AnimationComponentType
	ParticleComponentType
	// Add more component types as needed
)

// =============================================================================
// System Type Constants
// =============================================================================

const (
	MovementSystemType  SystemType = "movement"
	RenderSystemType    SystemType = "render"
	PhysicsSystemType   SystemType = "physics"
	InputSystemType     SystemType = "input"
	AudioSystemType     SystemType = "audio"
	AISystemType        SystemType = "ai"
	AnimationSystemType SystemType = "animation"
	ParticleSystemType  SystemType = "particle"
	CollisionSystemType SystemType = "collision"
	// Add more system types as needed
)

// =============================================================================
// Events and Messaging
// =============================================================================

// EventManager manages event publishing and subscription
type EventManager interface {
	Subscribe(eventType string, handler EventHandler) error
	Unsubscribe(eventType string, handler EventHandler) error
	Publish(event Event) error
	PublishAsync(event Event) error
}

// Event represents a game event
type Event interface {
	GetType() string
	GetTimestamp() time.Time
	GetData() interface{}
}

// EventHandler handles specific types of events
type EventHandler interface {
	HandleEvent(event Event) error
	GetSupportedEventTypes() []string
}

// =============================================================================
// MOD API Interfaces (Sandboxed)
// =============================================================================

// ModECSAPI provides a restricted ECS API for mods
type ModECSAPI interface {
	// Read-only Entity Operations
	IsEntityValid(entity EntityID) bool
	GetEntityCount() int

	// Read-only Component Operations
	GetComponent(entity EntityID, componentType ComponentType) (Component, error)
	HasComponent(entity EntityID, componentType ComponentType) bool

	// Restricted Query Operations
	QueryReadOnly(mask ComponentMask) EntityIterator
	QueryWithReadOnly(componentTypes ...ComponentType) EntityIterator

	// Limited Component Creation (only for mod-owned entities)
	CreateModEntity() (EntityID, error)
	AddModComponent(entity EntityID, component Component) error

	// Event System Access
	SubscribeToEvents(eventType string, handler EventHandler) error
	PublishModEvent(event Event) error
}

// ModSystem represents a system implemented by a mod
type ModSystem interface {
	System
	GetModID() string
	GetPermissions() ModPermissions
}

// ModPermissions defines what a mod is allowed to do
type ModPermissions struct {
	CanCreateEntities     bool
	CanModifyComponents   bool
	CanAccessSaveData     bool
	CanAccessNetwork      bool
	AllowedComponentTypes []ComponentType
	AllowedSystemTypes    []SystemType
}

// =============================================================================
// Performance and Profiling
// =============================================================================

// Profiler provides performance profiling capabilities
type Profiler interface {
	StartProfiling(name string) ProfileHandle
	StopProfiling(handle ProfileHandle)
	GetProfileResult(name string) ProfileResult
	GetAllProfileResults() map[string]ProfileResult
	ResetProfileData()
}

// ProfileHandle represents an active profiling session
type ProfileHandle interface {
	GetName() string
	GetStartTime() time.Time
	Stop()
}

// ProfileResult contains profiling data
type ProfileResult struct {
	Name            string
	TotalTime       time.Duration
	AverageTime     time.Duration
	MinTime         time.Duration
	MaxTime         time.Duration
	CallCount       int64
	MemoryAllocated int64
}

// =============================================================================
// Serialization and Persistence
// =============================================================================

// WorldSerializer handles saving and loading of the entire ECS world
type WorldSerializer interface {
	SerializeWorld() ([]byte, error)
	DeserializeWorld(data []byte) error
	SerializeEntity(entity EntityID) ([]byte, error)
	DeserializeEntity(data []byte) (EntityID, error)
	GetSaveVersion() int
	IsCompatibleVersion(version int) bool
}

// ComponentSerializer handles component-specific serialization
type ComponentSerializer interface {
	SerializeComponent(component Component) ([]byte, error)
	DeserializeComponent(componentType ComponentType, data []byte) (Component, error)
	RegisterComponentSerializer(componentType ComponentType, serializer func(Component) ([]byte, error), deserializer func([]byte) (Component, error))
}
