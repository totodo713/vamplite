// ==============================================
// ECS Framework - Go Interface Definitions
// Muscle Dreamer Game Engine
// Generated: 2025-08-03
// ==============================================

package ecs

import (
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
)

// ==============================================
// 1. Core ECS Types
// ==============================================

// EntityID represents a unique entity identifier
type EntityID uint64

// ComponentType represents the type of a component
type ComponentType string

// SystemType represents the type of a system
type SystemType string

// Priority defines execution priority for systems
type Priority int

const (
	PriorityLowest  Priority = 0
	PriorityLow     Priority = 25
	PriorityNormal  Priority = 50
	PriorityHigh    Priority = 75
	PriorityHighest Priority = 100
)

// ==============================================
// 2. Core Interfaces
// ==============================================

// World represents the main ECS world container
type World interface {
	// Entity management
	CreateEntity() EntityID
	DestroyEntity(EntityID) error
	IsEntityValid(EntityID) bool
	GetEntityCount() int

	// Component management
	AddComponent(EntityID, Component) error
	RemoveComponent(EntityID, ComponentType) error
	GetComponent(EntityID, ComponentType) (Component, error)
	HasComponent(EntityID, ComponentType) bool

	// System management
	RegisterSystem(System) error
	UnregisterSystem(SystemType) error
	GetSystem(SystemType) (System, error)

	// Query operations
	Query(QueryBuilder) QueryResult
	CreateQuery(ComponentType, ...ComponentType) QueryResult

	// World operations
	Update(deltaTime float64) error
	Render(screen *ebiten.Image) error
	Clear() error

	// Performance monitoring
	GetMetrics() *PerformanceMetrics
}

// EntityManager manages entity lifecycle
type EntityManager interface {
	CreateEntity() EntityID
	DestroyEntity(EntityID) error
	IsValid(EntityID) bool
	GetActiveEntities() []EntityID
	GetEntityCount() int
	RecycleEntity(EntityID) error

	// Entity relationships
	SetParent(child EntityID, parent EntityID) error
	GetParent(EntityID) (EntityID, bool)
	GetChildren(EntityID) []EntityID

	// Entity metadata
	SetTag(EntityID, string) error
	GetTag(EntityID) (string, bool)
	FindByTag(string) []EntityID
}

// ComponentStore manages component storage and access
type ComponentStore interface {
	// Component operations
	AddComponent(EntityID, Component) error
	RemoveComponent(EntityID, ComponentType) error
	GetComponent(EntityID, ComponentType) (Component, error)
	HasComponent(EntityID, ComponentType) bool

	// Bulk operations
	GetComponents(EntityID) []Component
	RemoveAllComponents(EntityID) error

	// Type management
	RegisterComponentType(ComponentType, func() Component) error
	GetRegisteredTypes() []ComponentType

	// Storage optimization
	Compact() error
	GetStorageStats() StorageStats
}

// SystemManager manages system registration and execution
type SystemManager interface {
	// System registration
	RegisterSystem(System) error
	UnregisterSystem(SystemType) error
	GetSystem(SystemType) (System, error)
	GetAllSystems() []System

	// Execution management
	UpdateSystems(deltaTime float64) error
	RenderSystems(screen *ebiten.Image) error

	// Dependency management
	SetSystemDependency(SystemType, SystemType) error
	GetExecutionOrder() []SystemType
	ValidateExecutionOrder() error

	// System state
	EnableSystem(SystemType) error
	DisableSystem(SystemType) error
	IsSystemEnabled(SystemType) bool
}

// QueryEngine provides efficient entity querying
type QueryEngine interface {
	// Query creation
	CreateQuery(QueryBuilder) QueryResult
	CacheQuery(string, QueryBuilder) QueryResult

	// Query execution
	Execute(QueryBuilder) QueryResult
	ExecuteCached(string) QueryResult

	// Query optimization
	OptimizeQueries() error
	ClearQueryCache() error
	GetQueryStats() QueryStats

	// Real-time updates
	UpdateQueryCache(EntityID, ComponentType, bool) error
}

// ==============================================
// 3. Component Interface
// ==============================================

// Component represents a data container
type Component interface {
	GetType() ComponentType
	Clone() Component
	Validate() error
}

// Serializable components can be saved/loaded
type Serializable interface {
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}

// Networked components can be synchronized
type Networked interface {
	GetNetworkID() uint32
	SetNetworkID(uint32)
	IsDirty() bool
	SetDirty(bool)
}

// ==============================================
// 4. System Interface
// ==============================================

// System represents a logic processor
type System interface {
	GetType() SystemType
	GetPriority() Priority
	GetDependencies() []SystemType

	// Lifecycle
	Initialize(World) error
	Update(deltaTime float64) error
	Shutdown() error

	// State management
	IsEnabled() bool
	SetEnabled(bool)

	// Performance
	GetExecutionTime() time.Duration
	GetUpdateCount() uint64
}

// RenderSystem can render to screen
type RenderSystem interface {
	System
	Render(screen *ebiten.Image) error
}

// ThreadSafeSystem can run in parallel
type ThreadSafeSystem interface {
	System
	CanRunInParallel() bool
	GetThreadSafetyLevel() ThreadSafetyLevel
}

type ThreadSafetyLevel int

const (
	ThreadSafetyNone ThreadSafetyLevel = iota
	ThreadSafetyRead
	ThreadSafetyWrite
	ThreadSafetyFull
)

// ==============================================
// 5. Query System
// ==============================================

// QueryBuilder constructs entity queries
type QueryBuilder interface {
	With(ComponentType) QueryBuilder
	Without(ComponentType) QueryBuilder
	WithTag(string) QueryBuilder
	WithParent(EntityID) QueryBuilder
	WithChild(EntityID) QueryBuilder
	Build() Query
}

// Query represents a compiled query
type Query interface {
	Execute() QueryResult
	GetSignature() string
	IsValid() bool
	GetComponentTypes() []ComponentType
}

// QueryResult provides query results
type QueryResult interface {
	GetEntities() []EntityID
	GetCount() int
	IsEmpty() bool

	// Iteration
	ForEach(func(EntityID) bool) error
	ForEachWithComponents(func(EntityID, []Component) bool) error

	// Filtering
	Filter(func(EntityID) bool) QueryResult

	// Sorting
	Sort(func(EntityID, EntityID) bool) QueryResult
}

// ==============================================
// 6. Memory Management
// ==============================================

// MemoryManager manages ECS memory allocation
type MemoryManager interface {
	// Memory pools
	AllocateBlock(size int) ([]byte, error)
	DeallocateBlock([]byte) error

	// Memory statistics
	GetMemoryUsage() MemoryUsage
	GetPoolStats() []PoolStats

	// Garbage collection
	Compact() error
	ForceGC() error

	// Memory limits
	SetMemoryLimit(int64) error
	GetMemoryLimit() int64
}

// MemoryPool manages fixed-size memory blocks
type MemoryPool interface {
	Allocate() ([]byte, error)
	Deallocate([]byte) error
	GetBlockSize() int
	GetFreeBlocks() int
	GetUsedBlocks() int
}

// ==============================================
// 7. Events and Messaging
// ==============================================

// EventBus handles event communication
type EventBus interface {
	// Event publishing
	Publish(Event) error
	PublishAsync(Event) error

	// Event subscription
	Subscribe(EventType, EventHandler) (SubscriptionID, error)
	Unsubscribe(SubscriptionID) error

	// Event filtering
	SubscribeWithFilter(EventType, EventFilter, EventHandler) (SubscriptionID, error)

	// Bus management
	Clear() error
	GetSubscriberCount(EventType) int
}

// Event represents a system event
type Event interface {
	GetType() EventType
	GetTimestamp() time.Time
	GetSource() interface{}
	GetData() interface{}
}

// EventHandler processes events
type EventHandler func(Event) error

// EventFilter filters events
type EventFilter func(Event) bool

type EventType string
type SubscriptionID uint64

// ==============================================
// 8. Performance Monitoring
// ==============================================

// MetricsCollector gathers performance data
type MetricsCollector interface {
	// Metrics recording
	RecordEntityCount(int)
	RecordSystemExecutionTime(SystemType, time.Duration)
	RecordMemoryUsage(int64)
	RecordFrameTime(time.Duration)

	// Metrics retrieval
	GetMetrics() *PerformanceMetrics
	GetMetricsHistory(time.Duration) []PerformanceMetrics

	// Thresholds
	SetThreshold(MetricType, float64) error
	GetThresholds() map[MetricType]float64

	// Alerts
	EnableAlerts(bool)
	GetAlerts() []PerformanceAlert
}

type MetricType string

const (
	MetricEntityCount     MetricType = "entity_count"
	MetricSystemExecTime  MetricType = "system_exec_time"
	MetricMemoryUsage     MetricType = "memory_usage"
	MetricFrameTime       MetricType = "frame_time"
	MetricQueryTime       MetricType = "query_time"
	MetricComponentAccess MetricType = "component_access"
)

// ==============================================
// 9. MOD System Integration
// ==============================================

// ModECSAPI provides limited ECS access for mods
type ModECSAPI interface {
	// Limited entity operations
	CreateEntity() (EntityID, error)
	DestroyEntity(EntityID) error

	// Restricted component access
	AddAllowedComponent(EntityID, Component) error
	GetAllowedComponent(EntityID, ComponentType) (Component, error)

	// Safe queries
	QueryEntities(ComponentType) ([]EntityID, error)

	// Resource limits
	GetEntityLimit() int
	GetMemoryLimit() int64
	GetUsedMemory() int64
}

// ModSecurityValidator validates mod operations
type ModSecurityValidator interface {
	ValidateComponentType(ComponentType) error
	ValidateSystemType(SystemType) error
	ValidateEntityOperation(EntityID, string) error
	ValidateMemoryUsage(int64) error

	// Permission checks
	HasPermission(ModID, Permission) bool
	GrantPermission(ModID, Permission) error
	RevokePermission(ModID, Permission) error
}

type ModID string
type Permission string

const (
	PermissionCreateEntity    Permission = "create_entity"
	PermissionModifyComponent Permission = "modify_component"
	PermissionQueryEntity     Permission = "query_entity"
	PermissionPlayAudio       Permission = "play_audio"
	PermissionModifyUI        Permission = "modify_ui"
)

// ==============================================
// 10. Data Structures
// ==============================================

// PerformanceMetrics contains performance data
type PerformanceMetrics struct {
	EntityCount    int           `json:"entity_count"`
	ComponentCount int           `json:"component_count"`
	SystemCount    int           `json:"system_count"`
	MemoryUsage    int64         `json:"memory_usage"`
	FrameTime      time.Duration `json:"frame_time"`
	UpdateTime     time.Duration `json:"update_time"`
	RenderTime     time.Duration `json:"render_time"`
	QueryTime      time.Duration `json:"query_time"`
	GCTime         time.Duration `json:"gc_time"`
	Timestamp      time.Time     `json:"timestamp"`
}

// StorageStats contains component storage statistics
type StorageStats struct {
	ComponentType  ComponentType `json:"component_type"`
	ComponentCount int           `json:"component_count"`
	MemoryUsed     int64         `json:"memory_used"`
	MemoryReserved int64         `json:"memory_reserved"`
	Fragmentation  float64       `json:"fragmentation"`
}

// QueryStats contains query performance statistics
type QueryStats struct {
	QuerySignature string        `json:"query_signature"`
	ExecutionCount int64         `json:"execution_count"`
	AverageTime    time.Duration `json:"average_time"`
	LastExecuted   time.Time     `json:"last_executed"`
	CacheHitRate   float64       `json:"cache_hit_rate"`
}

// MemoryUsage contains memory usage information
type MemoryUsage struct {
	TotalAllocated int64     `json:"total_allocated"`
	TotalReserved  int64     `json:"total_reserved"`
	TotalFree      int64     `json:"total_free"`
	Fragmentation  float64   `json:"fragmentation"`
	GCCount        int64     `json:"gc_count"`
	LastGCTime     time.Time `json:"last_gc_time"`
}

// PoolStats contains memory pool statistics
type PoolStats struct {
	BlockSize   int     `json:"block_size"`
	TotalBlocks int     `json:"total_blocks"`
	FreeBlocks  int     `json:"free_blocks"`
	UsedBlocks  int     `json:"used_blocks"`
	Utilization float64 `json:"utilization"`
}

// PerformanceAlert represents a performance warning
type PerformanceAlert struct {
	Type      MetricType    `json:"type"`
	Severity  AlertSeverity `json:"severity"`
	Message   string        `json:"message"`
	Value     float64       `json:"value"`
	Threshold float64       `json:"threshold"`
	Timestamp time.Time     `json:"timestamp"`
}

type AlertSeverity int

const (
	AlertInfo AlertSeverity = iota
	AlertWarning
	AlertError
	AlertCritical
)

// ==============================================
// 11. Basic Component Types
// ==============================================

// TransformComponent handles position, rotation, and scale
type TransformComponent struct {
	Position Vector2 `json:"position"`
	Rotation float64 `json:"rotation"`
	Scale    Vector2 `json:"scale"`

	// Hierarchy
	Parent   EntityID   `json:"parent,omitempty"`
	Children []EntityID `json:"children,omitempty"`

	// Cached world transform
	WorldTransform Matrix3x3 `json:"-"`
	dirty          bool
	mutex          sync.RWMutex
}

func (t *TransformComponent) GetType() ComponentType { return "transform" }

// SpriteComponent handles sprite rendering
type SpriteComponent struct {
	Image   *ebiten.Image `json:"-"`
	ImageID string        `json:"image_id"`
	Color   color.RGBA    `json:"color"`
	ZOrder  int           `json:"z_order"`
	Visible bool          `json:"visible"`

	// Sprite sheet animation
	FrameIndex  int     `json:"frame_index"`
	FrameCount  int     `json:"frame_count"`
	FrameWidth  int     `json:"frame_width"`
	FrameHeight int     `json:"frame_height"`
	AnimSpeed   float64 `json:"anim_speed"`

	// Rendering options
	FlipH bool `json:"flip_h"`
	FlipV bool `json:"flip_v"`
}

func (s *SpriteComponent) GetType() ComponentType { return "sprite" }

// PhysicsComponent handles physics properties
type PhysicsComponent struct {
	Velocity     Vector2 `json:"velocity"`
	Acceleration Vector2 `json:"acceleration"`
	Mass         float64 `json:"mass"`
	Friction     float64 `json:"friction"`
	Restitution  float64 `json:"restitution"`

	// Physics state
	Kinematic bool `json:"kinematic"`
	Static    bool `json:"static"`
	Enabled   bool `json:"enabled"`

	// Forces
	AppliedForces []Vector2 `json:"-"`
}

func (p *PhysicsComponent) GetType() ComponentType { return "physics" }

// HealthComponent handles entity health/damage
type HealthComponent struct {
	Current int `json:"current"`
	Maximum int `json:"maximum"`
	Armor   int `json:"armor"`
	Shield  int `json:"shield"`

	// Health state
	Invulnerable bool      `json:"invulnerable"`
	LastDamage   time.Time `json:"last_damage"`

	// Regeneration
	RegenRate  float64 `json:"regen_rate"`
	RegenDelay float64 `json:"regen_delay"`
}

func (h *HealthComponent) GetType() ComponentType { return "health" }

// AIComponent handles AI behavior
type AIComponent struct {
	BehaviorType string                 `json:"behavior_type"`
	State        string                 `json:"state"`
	Target       EntityID               `json:"target"`
	Parameters   map[string]interface{} `json:"parameters"`

	// AI timing
	NextUpdate time.Time `json:"next_update"`
	UpdateRate float64   `json:"update_rate"`

	// AI state
	Enabled bool `json:"enabled"`
}

func (ai *AIComponent) GetType() ComponentType { return "ai" }

// ==============================================
// 12. Utility Types
// ==============================================

// Vector2 represents a 2D vector
type Vector2 struct {
	X, Y float64 `json:"x,y"`
}

// Matrix3x3 represents a 3x3 transformation matrix
type Matrix3x3 struct {
	M00, M01, M02 float64
	M10, M11, M12 float64
	M20, M21, M22 float64
}

// AABB represents an axis-aligned bounding box
type AABB struct {
	Min Vector2 `json:"min"`
	Max Vector2 `json:"max"`
}

// ==============================================
// 13. Error Types
// ==============================================

// ECSError represents an ECS-specific error
type ECSError struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Component string    `json:"component,omitempty"`
	Entity    EntityID  `json:"entity,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *ECSError) Error() string {
	return e.Message
}

// Common error codes
const (
	ErrEntityNotFound     = "ENTITY_NOT_FOUND"
	ErrComponentNotFound  = "COMPONENT_NOT_FOUND"
	ErrSystemNotFound     = "SYSTEM_NOT_FOUND"
	ErrInvalidEntityID    = "INVALID_ENTITY_ID"
	ErrComponentExists    = "COMPONENT_EXISTS"
	ErrSystemExists       = "SYSTEM_EXISTS"
	ErrCircularDependency = "CIRCULAR_DEPENDENCY"
	ErrMemoryLimit        = "MEMORY_LIMIT_EXCEEDED"
	ErrPermissionDenied   = "PERMISSION_DENIED"
	ErrResourceExhausted  = "RESOURCE_EXHAUSTED"
)

// ==============================================
// 14. Factory Functions
// ==============================================

// WorldConfig contains world initialization parameters
type WorldConfig struct {
	MaxEntities    int           `json:"max_entities"`
	MemoryLimit    int64         `json:"memory_limit"`
	EnableMetrics  bool          `json:"enable_metrics"`
	EnableEvents   bool          `json:"enable_events"`
	ThreadPoolSize int           `json:"thread_pool_size"`
	QueryCacheSize int           `json:"query_cache_size"`
	GCInterval     time.Duration `json:"gc_interval"`
}

// NewWorld creates a new ECS world with the given configuration
func NewWorld(config WorldConfig) (World, error) {
	// Implementation would be provided by the concrete implementation
	return nil, nil
}

// NewWorldWithDefaults creates a new ECS world with default settings
func NewWorldWithDefaults() (World, error) {
	return NewWorld(WorldConfig{
		MaxEntities:    10000,
		MemoryLimit:    256 * 1024 * 1024, // 256MB
		EnableMetrics:  true,
		EnableEvents:   true,
		ThreadPoolSize: 4,
		QueryCacheSize: 1000,
		GCInterval:     time.Second * 30,
	})
}

// ==============================================
// Usage Examples (Comments)
// ==============================================

/*
Example Usage:

// Create world
world, err := NewWorldWithDefaults()
if err != nil {
    log.Fatal(err)
}

// Create entity with components
playerEntity := world.CreateEntity()
world.AddComponent(playerEntity, &TransformComponent{
    Position: Vector2{X: 100, Y: 100},
    Scale:    Vector2{X: 1, Y: 1},
})
world.AddComponent(playerEntity, &SpriteComponent{
    ImageID: "player.png",
    Color:   color.RGBA{255, 255, 255, 255},
    Visible: true,
})

// Register systems
world.RegisterSystem(&MovementSystem{})
world.RegisterSystem(&RenderingSystem{})

// Game loop
for {
    world.Update(deltaTime)
    world.Render(screen)
}
*/
