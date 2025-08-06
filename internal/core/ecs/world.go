// Package ecs provides the core Entity Component System framework for Muscle Dreamer.
package ecs

// ==============================================
// World Interface - ECS統合管理の中核
// ==============================================

// World represents the main ECS world that manages all entities, components, and systems.
// It provides the primary interface for game logic to interact with the ECS framework.
type World interface {
	// Entity management
	CreateEntity() EntityID
	DestroyEntity(EntityID) error
	IsEntityValid(EntityID) bool
	GetEntityCount() int
	GetActiveEntities() []EntityID

	// Component management
	AddComponent(EntityID, Component) error
	RemoveComponent(EntityID, ComponentType) error
	GetComponent(EntityID, ComponentType) (Component, error)
	HasComponent(EntityID, ComponentType) bool
	GetComponents(EntityID) []Component

	// System management
	RegisterSystem(System) error
	UnregisterSystem(SystemType) error
	GetSystem(SystemType) (System, error)
	GetAllSystems() []System
	EnableSystem(SystemType) error
	DisableSystem(SystemType) error
	IsSystemEnabled(SystemType) bool

	// World operations
	Update(deltaTime float64) error
	Render(screen interface{}) error
	Shutdown() error

	// Performance monitoring
	GetMetrics() *PerformanceMetrics
	GetMemoryUsage() *MemoryUsage
	GetStorageStats() []StorageStats
	GetQueryStats() []QueryStats

	// Configuration
	GetConfig() *WorldConfig
	UpdateConfig(*WorldConfig) error

	// Events
	EmitEvent(Event) error
	Subscribe(EventType, EventHandler) error
	Unsubscribe(EventType, EventHandler) error

	// Query interface
	Query() QueryBuilder
	CreateQuery(QueryBuilder) QueryResult
	ExecuteQuery(QueryBuilder) QueryResult

	// Batch operations
	CreateEntities(count int) []EntityID
	DestroyEntities([]EntityID) error
	AddComponents(EntityID, []Component) error
	RemoveComponents(EntityID, []ComponentType) error

	// Serialization for saves/mods
	SerializeEntity(EntityID) ([]byte, error)
	DeserializeEntity([]byte) (EntityID, error)
	SerializeWorld() ([]byte, error)
	DeserializeWorld([]byte) error

	// Thread safety
	Lock()
	RLock()
	Unlock()
	RUnlock()
}

// ==============================================
// Component Interface
// ==============================================

// Component represents a data container that can be attached to entities.
// All components must implement this interface for type safety and serialization.
type Component interface {
	// GetType returns the component type for identification
	GetType() ComponentType

	// Clone creates a deep copy of the component
	Clone() Component

	// Validate ensures the component data is valid
	Validate() error

	// Size returns the memory size of the component in bytes
	Size() int

	// Serialize converts the component to bytes for persistence
	Serialize() ([]byte, error)

	// Deserialize loads component data from bytes
	Deserialize([]byte) error
}

// ==============================================
// System Interface
// ==============================================

// System represents a logic processor that operates on entities with specific components.
// Systems define the behavior of the ECS framework.
type System interface {
	// GetType returns the system type for identification
	GetType() SystemType

	// GetPriority returns the execution priority (higher executes first)
	GetPriority() Priority

	// GetDependencies returns systems that must execute before this one
	GetDependencies() []SystemType

	// GetRequiredComponents returns components this system operates on
	GetRequiredComponents() []ComponentType

	// Initialize sets up the system (called once)
	Initialize(World) error

	// Update processes entities with required components
	Update(World, float64) error

	// Render draws entities to screen (optional, for rendering systems)
	Render(World, interface{}) error

	// Shutdown cleans up system resources
	Shutdown() error

	// IsEnabled returns whether the system is currently active
	IsEnabled() bool

	// SetEnabled controls system execution
	SetEnabled(bool)

	// GetMetrics returns system performance data
	GetMetrics() *SystemMetrics

	// GetThreadSafety returns the thread safety level
	GetThreadSafety() ThreadSafetyLevel

	// CanRunInParallel returns true if system can run concurrently
	CanRunInParallel() bool
}

// ==============================================
// Event System Interface
// ==============================================

// Event represents an occurrence that systems can react to.
type Event interface {
	// GetType returns the event type for routing
	GetType() EventType

	// GetEntity returns the entity associated with this event (if any)
	GetEntity() EntityID

	// GetTimestamp returns when the event occurred
	GetTimestamp() int64

	// GetData returns event-specific data
	GetData() interface{}

	// Serialize converts event to bytes for logging/replay
	Serialize() ([]byte, error)
}

// EventType represents different types of events in the system.
type EventType string

// Common event types
const (
	EventEntityCreated    EventType = "entity_created"
	EventEntityDestroyed  EventType = "entity_destroyed"
	EventComponentAdded   EventType = "component_added"
	EventComponentRemoved EventType = "component_removed"
	EventSystemError      EventType = "system_error"
	EventPerformanceAlert EventType = "performance_alert"
)

// EventHandler is a function that processes events.
type EventHandler func(Event) error

// Note: QueryBuilder and QueryResult interfaces are defined in query.go

// ==============================================
// Performance Monitoring Interface
// ==============================================

// SystemMetrics contains performance data for a specific system.
type SystemMetrics struct {
	SystemType       SystemType `json:"system_type"`
	ExecutionCount   int64      `json:"execution_count"`
	TotalTime        int64      `json:"total_time_ns"`
	AverageTime      int64      `json:"average_time_ns"`
	MaxTime          int64      `json:"max_time_ns"`
	MinTime          int64      `json:"min_time_ns"`
	ErrorCount       int64      `json:"error_count"`
	LastExecution    int64      `json:"last_execution_ns"`
	EntitysProcessed int64      `json:"entities_processed"`
	MemoryAllocated  int64      `json:"memory_allocated_bytes"`
}

// ==============================================
// Resource Management Interface
// ==============================================

// ResourceManager manages game resources (textures, sounds, etc.) for ECS.
type ResourceManager interface {
	// Load resources
	LoadTexture(string) error
	LoadSound(string) error
	LoadFont(string) error

	// Get resources
	GetTexture(string) interface{}
	GetSound(string) interface{}
	GetFont(string) interface{}

	// Unload resources
	UnloadTexture(string) error
	UnloadSound(string) error
	UnloadAll() error

	// Resource info
	GetLoadedResources() []string
	GetMemoryUsage() int64
}

// ==============================================
// Serialization Interface
// ==============================================

// Serializable provides save/load functionality for ECS data.
type Serializable interface {
	// Serialize converts the object to bytes
	Serialize() ([]byte, error)

	// Deserialize loads object data from bytes
	Deserialize([]byte) error

	// GetVersion returns serialization format version
	GetVersion() int
}

// ==============================================
// Factory Interface
// ==============================================

// WorldFactory creates and configures ECS worlds.
type WorldFactory interface {
	// CreateWorld creates a new world with configuration
	CreateWorld(WorldConfig) (World, error)

	// CreateDefaultWorld creates a world with default settings
	CreateDefaultWorld() (World, error)

	// RegisterComponentType registers a component type for factory creation
	RegisterComponentType(ComponentType, func() Component) error

	// RegisterSystemType registers a system type for factory creation
	RegisterSystemType(SystemType, func() System) error

	// CreateComponent creates a component by type
	CreateComponent(ComponentType) (Component, error)

	// CreateSystem creates a system by type
	CreateSystem(SystemType) (System, error)
}

// ==============================================
// Debug Interface
// ==============================================

// DebugInfo provides debugging information about the ECS state.
type DebugInfo struct {
	EntityCount     int                   `json:"entity_count"`
	ComponentCounts map[ComponentType]int `json:"component_counts"`
	SystemStates    map[SystemType]bool   `json:"system_states"`
	MemoryUsage     int64                 `json:"memory_usage_bytes"`
	QueryCacheSize  int                   `json:"query_cache_size"`
	EventQueueSize  int                   `json:"event_queue_size"`
	Performance     *PerformanceMetrics   `json:"performance"`
}

// Debugger provides debugging capabilities for the ECS framework.
type Debugger interface {
	// GetDebugInfo returns current ECS state
	GetDebugInfo() *DebugInfo

	// DumpEntities exports all entities to a readable format
	DumpEntities() (string, error)

	// ValidateIntegrity checks ECS data consistency
	ValidateIntegrity() error

	// GetEntityDetails returns detailed info about a specific entity
	GetEntityDetails(EntityID) (string, error)

	// TraceSystem enables/disables system execution tracing
	TraceSystem(SystemType, bool)

	// GetSystemTrace returns execution trace for a system
	GetSystemTrace(SystemType) ([]string, error)
}
