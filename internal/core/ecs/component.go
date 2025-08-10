// Package ecs provides the core Entity Component System framework for Muscle Dreamer.
package ecs

import (
	"reflect"
	"time"
)

// ==============================================
// ComponentStore Interface - コンポーネント管理・ストレージ
// ==============================================

// ComponentStore manages component storage and retrieval with high performance.
// It uses Structure of Arrays (SoA) memory layout and sparse sets for efficiency.
type ComponentStore interface {
	// Core component operations
	AddComponent(EntityID, Component) error
	RemoveComponent(EntityID, ComponentType) error
	GetComponent(EntityID, ComponentType) (Component, error)
	HasComponent(EntityID, ComponentType) bool
	GetComponentUnsafe(EntityID, ComponentType) Component

	// Bulk operations for performance
	GetComponents(EntityID) []Component
	GetComponentTypes(EntityID) []ComponentType
	RemoveAllComponents(EntityID) error
	SetComponentsFromArchetype(EntityID, ArchetypeID, []Component) error

	// Type management and registration
	RegisterComponentType(ComponentType, func() Component) error
	UnregisterComponentType(ComponentType) error
	GetRegisteredTypes() []ComponentType
	IsTypeRegistered(ComponentType) bool
	GetComponentFactory(ComponentType) (func() Component, error)

	// Query support for high-performance entity filtering
	GetEntitiesWith(ComponentType) []EntityID
	GetEntitiesWithAll([]ComponentType) []EntityID
	GetEntitiesWithAny([]ComponentType) []EntityID
	GetEntitiesWithout(ComponentType) []EntityID
	GetEntitiesWithArchetype(ArchetypeID) []EntityID

	// Storage optimization and memory management
	Compact() error
	CompactType(ComponentType) error
	GetStorageStats() []StorageStats
	GetStorageStatsForType(ComponentType) *StorageStats
	GetFragmentation() float64
	GetFragmentationForType(ComponentType) float64

	// Batch operations for efficiency
	AddComponents(EntityID, []Component) error
	RemoveComponents(EntityID, []ComponentType) error
	UpdateComponents(EntityID, []Component) error
	CloneComponents(from EntityID, to EntityID) error

	// Component data access patterns for systems
	GetComponentArray(ComponentType) []Component
	GetComponentArrayForEntities(ComponentType, []EntityID) []Component
	GetSparseSet(ComponentType) SparseSet
	GetDenseArray(ComponentType) []Component

	// Serialization for save/load and mods
	SerializeComponent(EntityID, ComponentType) ([]byte, error)
	DeserializeComponent(EntityID, ComponentType, []byte) error
	SerializeEntity(EntityID) (map[ComponentType][]byte, error)
	DeserializeEntity(EntityID, map[ComponentType][]byte) error

	// Change tracking for optimizations
	MarkDirty(EntityID, ComponentType)
	IsDirty(EntityID, ComponentType) bool
	GetDirtyComponents() map[EntityID][]ComponentType
	ClearDirtyFlags()
	EnableChangeTracking(bool)

	// Events for reactive systems
	OnComponentAdded(func(EntityID, ComponentType, Component)) error
	OnComponentRemoved(func(EntityID, ComponentType)) error
	OnComponentChanged(func(EntityID, ComponentType, Component, Component)) error

	// Memory pools and allocation
	GetMemoryPool(ComponentType) MemoryPool
	CreateMemoryPool(ComponentType, int) error
	PreallocateComponents(ComponentType, int) error
	GetPoolStats() map[ComponentType]*MemoryPoolStats

	// Thread safety
	Lock()
	RLock()
	Unlock()
	RUnlock()
	LockType(ComponentType)
	RLockType(ComponentType)
	UnlockType(ComponentType)
	RUnlockType(ComponentType)

	// Debug and validation
	ValidateIntegrity() error
	GetDebugInfo() *ComponentStoreDebugInfo
	GetTypeInfo(ComponentType) *ComponentTypeInfo
}

// ==============================================
// Sparse Set Interface for High-Performance Storage
// ==============================================

// SparseSet provides O(1) insertion, deletion, and lookup for entity-component mapping.
// This is a key data structure for high-performance ECS implementations.
type SparseSet interface {
	// Core operations
	Insert(EntityID, Component) error
	Remove(EntityID) bool
	Get(EntityID) (Component, bool)
	Contains(EntityID) bool

	// Bulk operations
	GetAll() []Component
	GetEntities() []EntityID
	GetPairs() []EntityComponentPair
	Clear()

	// Iteration support
	ForEach(func(EntityID, Component))
	ForEachEntity(func(EntityID))

	// Memory management
	Reserve(int)
	Compact()
	GetCapacity() int
	GetSize() int
	GetMemoryUsage() int64

	// Serialization
	Serialize() ([]byte, error)
	Deserialize([]byte) error

	// Statistics
	GetStats() *SparseSetStats
}

// EntityComponentPair represents an entity-component association.
type EntityComponentPair struct {
	EntityID  EntityID  `json:"entityId"`
	Component Component `json:"component"`
}

// SparseSetStats contains performance statistics for sparse sets.
type SparseSetStats struct {
	Size          int     `json:"size"`
	Capacity      int     `json:"capacity"`
	MemoryUsage   int64   `json:"memoryUsageBytes"`
	LoadFactor    float64 `json:"loadFactor"`
	Fragmentation float64 `json:"fragmentation"`
	AccessCount   int64   `json:"accessCount"`
	InsertCount   int64   `json:"insertCount"`
	RemoveCount   int64   `json:"removeCount"`
}

// ==============================================
// Memory Pool Interface for Component Allocation
// ==============================================

// MemoryPool manages pre-allocated memory blocks for components.
type MemoryPool interface {
	// Allocation
	Allocate() (Component, error)
	AllocateBatch(int) ([]Component, error)
	Deallocate(Component) error
	DeallocateBatch([]Component) error

	// Pool management
	Grow(int) error
	Shrink(int) error
	GetSize() int
	GetCapacity() int
	GetFreeCount() int

	// Statistics
	GetStats() *MemoryPoolStats
	GetFragmentation() float64
	GetHitRate() float64

	// Cleanup
	Clear() error
	Compact() error
}

// MemoryPoolStats contains statistics about memory pool usage.
type MemoryPoolStats struct {
	ComponentType  ComponentType `json:"componentType"`
	TotalAllocated int           `json:"totalAllocated"`
	CurrentUsed    int           `json:"currentUsed"`
	PeakUsed       int           `json:"peakUsed"`
	PoolCapacity   int           `json:"poolCapacity"`
	MemoryUsage    int64         `json:"memoryUsageBytes"`
	AllocCount     int64         `json:"allocCount"`
	DeallocCount   int64         `json:"deallocCount"`
	HitRate        float64       `json:"hitRate"`
	Fragmentation  float64       `json:"fragmentation"`
	LastGrowth     time.Time     `json:"lastGrowth"`
}

// ==============================================
// Component Type Information and Metadata
// ==============================================

// ComponentTypeInfo contains metadata about a registered component type.
type ComponentTypeInfo struct {
	Type         ComponentType                   `json:"type"`
	ReflectType  reflect.Type                    `json:"-"`
	Size         int                             `json:"sizeBytes"`
	Alignment    int                             `json:"alignment"`
	IsPointer    bool                            `json:"isPointer"`
	Factory      func() Component                `json:"-"`
	Validator    func(Component) error           `json:"-"`
	Serializer   func(Component) ([]byte, error) `json:"-"`
	Deserializer func([]byte) (Component, error) `json:"-"`

	// Statistics
	InstanceCount int64     `json:"instanceCount"`
	TotalMemory   int64     `json:"totalMemoryBytes"`
	CreatedAt     time.Time `json:"createdAt"`
	LastAccessed  time.Time `json:"lastAccessed"`
	AccessCount   int64     `json:"accessCount"`
}

// ComponentRegistry manages component type registration and metadata.
type ComponentRegistry interface {
	// Type registration
	RegisterType(ComponentType, func() Component) error
	RegisterTypeWithInfo(ComponentType, *ComponentTypeInfo) error
	UnregisterType(ComponentType) error

	// Type information
	GetTypeInfo(ComponentType) (*ComponentTypeInfo, error)
	GetAllTypes() []ComponentType
	IsRegistered(ComponentType) bool
	GetFactory(ComponentType) (func() Component, error)

	// Type validation
	ValidateComponent(Component) error
	ValidateType(ComponentType, Component) error

	// Type serialization
	SerializeType(ComponentType, Component) ([]byte, error)
	DeserializeType(ComponentType, []byte) (Component, error)

	// Statistics
	GetTypeStats() map[ComponentType]*ComponentTypeStats
	GetRegistryStats() *ComponentRegistryStats
}

// ComponentTypeStats contains usage statistics for a component type.
type ComponentTypeStats struct {
	Type             ComponentType `json:"type"`
	InstanceCount    int64         `json:"instanceCount"`
	TotalMemory      int64         `json:"totalMemoryBytes"`
	AverageSize      float64       `json:"averageSizeBytes"`
	AccessCount      int64         `json:"accessCount"`
	CreationCount    int64         `json:"creationCount"`
	DestructionCount int64         `json:"destructionCount"`
	LastAccessed     time.Time     `json:"lastAccessed"`
	FirstCreated     time.Time     `json:"firstCreated"`
}

// ComponentRegistryStats contains overall registry statistics.
type ComponentRegistryStats struct {
	RegisteredTypes  int                                   `json:"registeredTypes"`
	TotalInstances   int64                                 `json:"totalInstances"`
	TotalMemory      int64                                 `json:"totalMemoryBytes"`
	TypeStats        map[ComponentType]*ComponentTypeStats `json:"typeStats"`
	MostUsedTypes    []ComponentType                       `json:"mostUsedTypes"`
	LeastUsedTypes   []ComponentType                       `json:"leastUsedTypes"`
	MemoryHeavyTypes []ComponentType                       `json:"memoryHeavyTypes"`
}

// ==============================================
// Component Store Debug Information
// ==============================================

// ComponentStoreDebugInfo provides comprehensive debugging information.
type ComponentStoreDebugInfo struct {
	RegisteredTypes       int                                  `json:"registered_types"`
	TotalComponents       int64                                `json:"total_components"`
	TotalMemoryUsage      int64                                `json:"total_memory_usage_bytes"`
	StorageStats          []StorageStats                       `json:"storage_stats"`
	MemoryPoolStats       map[ComponentType]*MemoryPoolStats   `json:"memory_pool_stats"`
	SparseSetStats        map[ComponentType]*SparseSetStats    `json:"sparse_set_stats"`
	TypeInfo              map[ComponentType]*ComponentTypeInfo `json:"type_info"`
	DirtyComponents       map[EntityID][]ComponentType         `json:"dirty_components"`
	ChangeTrackingEnabled bool                                 `json:"change_tracking_enabled"`
	ThreadSafetyEnabled   bool                                 `json:"thread_safety_enabled"`
}

// ==============================================
// Component Change Tracking
// ==============================================

// ChangeTracker tracks component modifications for optimization.
type ChangeTracker interface {
	// Change tracking
	MarkDirty(EntityID, ComponentType)
	MarkClean(EntityID, ComponentType)
	IsDirty(EntityID, ComponentType) bool
	GetDirtyEntities(ComponentType) []EntityID
	GetDirtyComponents(EntityID) []ComponentType

	// Bulk operations
	MarkAllDirty(ComponentType)
	MarkAllClean(ComponentType)
	ClearAllDirty()

	// Change detection
	HasChanges() bool
	GetChangeCount() int64
	GetChangesForType(ComponentType) int64

	// Events
	OnComponentChanged(func(EntityID, ComponentType)) error
	OnEntityChanged(func(EntityID)) error

	// Configuration
	Enable(bool)
	IsEnabled() bool
	SetGranularity(ComponentChangeGranularity)
}

// ComponentChangeGranularity defines the level of change tracking detail.
type ComponentChangeGranularity int

const (
	ChangeGranularityNone      ComponentChangeGranularity = iota // No tracking
	ChangeGranularityComponent                                   // Track component-level changes
	ChangeGranularityField                                       // Track field-level changes
	ChangeGranularityValue                                       // Track value-level changes
)

// ==============================================
// Structure of Arrays (SoA) Implementation
// ==============================================

// SoAStorage implements Structure of Arrays storage for cache-friendly access.
type SoAStorage interface {
	// Array-based access
	GetArray(ComponentType) interface{}
	GetEntityArray() []EntityID
	GetComponentArray(ComponentType) []Component

	// Indexed access
	GetByIndex(ComponentType, int) (Component, error)
	SetByIndex(ComponentType, int, Component) error
	GetEntityByIndex(int) EntityID

	// Bulk operations
	GetRange(ComponentType, int, int) []Component
	SetRange(ComponentType, int, []Component) error
	SwapElements(int, int) error

	// Array management
	GrowArray(ComponentType, int) error
	ShrinkArray(ComponentType, int) error
	CompactArrays() error

	// Memory layout optimization
	GetStride(ComponentType) int
	GetAlignment(ComponentType) int
	IsContiguous(ComponentType) bool
	GetCacheEfficiency() float64

	// Statistics
	GetArrayStats(ComponentType) *ArrayStats
	GetLayoutStats() *LayoutStats
}

// ArrayStats contains statistics about a specific component array.
type ArrayStats struct {
	ComponentType ComponentType `json:"component_type"`
	Length        int           `json:"length"`
	Capacity      int           `json:"capacity"`
	ElementSize   int           `json:"element_size_bytes"`
	TotalSize     int64         `json:"total_size_bytes"`
	Stride        int           `json:"stride"`
	Alignment     int           `json:"alignment"`
	Fragmentation float64       `json:"fragmentation"`
	CacheHitRate  float64       `json:"cache_hit_rate"`
}

// LayoutStats contains statistics about the overall memory layout.
type LayoutStats struct {
	TotalArrays        int     `json:"total_arrays"`
	TotalMemory        int64   `json:"total_memory_bytes"`
	AverageUtilization float64 `json:"average_utilization"`
	CacheEfficiency    float64 `json:"cache_efficiency"`
	Fragmentation      float64 `json:"fragmentation"`
	OptimalLayout      bool    `json:"optimal_layout"`
}

// ==============================================
// Component Validation and Safety
// ==============================================

// ComponentValidator provides validation for component data integrity.
type ComponentValidator interface {
	// Validation rules
	AddValidationRule(ComponentType, func(Component) error) error
	RemoveValidationRule(ComponentType) error
	GetValidationRules(ComponentType) []func(Component) error

	// Validation execution
	ValidateComponent(Component) error
	ValidateEntity(EntityID) error
	ValidateAll() error
	ValidateType(ComponentType) error

	// Batch validation
	ValidateComponents([]Component) []error
	ValidateEntities([]EntityID) []error

	// Custom validators
	RegisterCustomValidator(string, func(interface{}) error) error
	ApplyCustomValidator(string, interface{}) error

	// Validation statistics
	GetValidationStats() *ValidationStats
	GetFailureReport() []*ValidationFailure
}

// ValidationStats contains statistics about component validation.
type ValidationStats struct {
	TotalValidations      int64                   `json:"total_validations"`
	SuccessfulValidations int64                   `json:"successful_validations"`
	FailedValidations     int64                   `json:"failed_validations"`
	ValidationsByType     map[ComponentType]int64 `json:"validations_by_type"`
	FailuresByType        map[ComponentType]int64 `json:"failures_by_type"`
	AverageTime           float64                 `json:"average_time_ns"`
	LastValidation        time.Time               `json:"last_validation"`
}

// ValidationFailure represents a component validation failure.
type ValidationFailure struct {
	EntityID      EntityID           `json:"entity_id"`
	ComponentType ComponentType      `json:"component_type"`
	FieldName     string             `json:"field_name,omitempty"`
	ErrorMessage  string             `json:"error_message"`
	Timestamp     time.Time          `json:"timestamp"`
	Severity      ValidationSeverity `json:"severity"`
}

// ValidationSeverity represents the severity level of validation failures.
type ValidationSeverity int

const (
	ValidationSeverityInfo ValidationSeverity = iota
	ValidationSeverityWarning
	ValidationSeverityError
	ValidationSeverityCritical
)
