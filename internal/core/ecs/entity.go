// Package ecs provides the core Entity Component System framework for Muscle Dreamer.
package ecs

import (
	"sync"
	"time"
)

// ==============================================
// EntityManager Interface - エンティティライフサイクル管理
// ==============================================

// EntityManager manages the lifecycle of entities in the ECS world.
// It handles entity creation, destruction, relationships, and metadata.
type EntityManager interface {
	// Core entity operations
	CreateEntity() EntityID
	CreateEntityWithID(EntityID) error
	DestroyEntity(EntityID) error
	IsValid(EntityID) bool
	GetActiveEntities() []EntityID
	GetEntityCount() int
	GetMaxEntityCount() int

	// Entity recycling for performance
	RecycleEntity(EntityID) error
	GetRecycledCount() int
	ClearRecycled() error

	// Entity relationships and hierarchy
	SetParent(child EntityID, parent EntityID) error
	GetParent(EntityID) (EntityID, bool)
	GetChildren(EntityID) []EntityID
	GetDescendants(EntityID) []EntityID
	GetAncestors(EntityID) []EntityID
	RemoveFromParent(EntityID) error
	IsAncestor(ancestor EntityID, descendant EntityID) bool

	// Entity metadata and tagging
	SetTag(EntityID, string) error
	GetTag(EntityID) (string, bool)
	RemoveTag(EntityID) error
	FindByTag(string) []EntityID
	GetAllTags() []string

	// Entity groups for bulk operations
	CreateGroup(string) error
	AddToGroup(EntityID, string) error
	RemoveFromGroup(EntityID, string) error
	GetGroup(string) []EntityID
	GetEntityGroups(EntityID) []string
	DestroyGroup(string) error

	// Entity lifecycle events
	OnEntityCreated(func(EntityID)) error
	OnEntityDestroyed(func(EntityID)) error
	OnParentChanged(func(EntityID, EntityID, EntityID)) error

	// Batch operations for performance
	CreateEntities(count int) []EntityID
	DestroyEntities([]EntityID) error
	ValidateEntities([]EntityID) []EntityID

	// Entity archetype management (for performance optimization)
	GetArchetype(EntityID) ArchetypeID
	GetEntitiesByArchetype(ArchetypeID) []EntityID
	GetArchetypeCount() int

	// Memory and performance management
	Compact() error
	GetFragmentation() float64
	GetMemoryUsage() int64
	GetPoolStats() *EntityPoolStats

	// Serialization
	SerializeEntity(EntityID) (*EntityData, error)
	DeserializeEntity(*EntityData) (EntityID, error)
	SerializeBatch([]EntityID) ([]*EntityData, error)
	DeserializeBatch([]*EntityData) ([]EntityID, error)

	// Thread safety
	Lock()
	RLock()
	Unlock()
	RUnlock()

	// Debug and validation
	ValidateIntegrity() error
	GetDebugInfo() *EntityManagerDebugInfo
}

// ==============================================
// Entity Data Structures
// ==============================================

// ArchetypeID represents a unique combination of component types.
// Entities with the same archetype share the same component types.
type ArchetypeID uint32

// EntityData represents serializable entity information.
type EntityData struct {
	ID         EntityID                 `json:"id"`
	Components map[ComponentType][]byte `json:"components"`
	Parent     EntityID                 `json:"parent,omitempty"`
	Children   []EntityID               `json:"children,omitempty"`
	Tag        string                   `json:"tag,omitempty"`
	Groups     []string                 `json:"groups,omitempty"`
	Archetype  ArchetypeID              `json:"archetype"`
	CreatedAt  time.Time                `json:"created_at"`
	ModifiedAt time.Time                `json:"modified_at"`
}

// EntityPoolStats contains statistics about entity memory pools.
type EntityPoolStats struct {
	TotalEntities    int     `json:"total_entities"`
	ActiveEntities   int     `json:"active_entities"`
	RecycledEntities int     `json:"recycled_entities"`
	PoolCapacity     int     `json:"pool_capacity"`
	MemoryUsed       int64   `json:"memory_used_bytes"`
	MemoryReserved   int64   `json:"memory_reserved_bytes"`
	Fragmentation    float64 `json:"fragmentation"`
	HitRate          float64 `json:"hit_rate"`
}

// EntityManagerDebugInfo provides debugging information about entity management.
type EntityManagerDebugInfo struct {
	EntityCount     int                 `json:"entity_count"`
	MaxEntityID     EntityID            `json:"max_entity_id"`
	RecycledCount   int                 `json:"recycled_count"`
	ArchetypeCount  int                 `json:"archetype_count"`
	TagCount        int                 `json:"tag_count"`
	GroupCount      int                 `json:"group_count"`
	HierarchyDepth  int                 `json:"max_hierarchy_depth"`
	MemoryUsage     int64               `json:"memory_usage_bytes"`
	PoolStats       *EntityPoolStats    `json:"pool_stats"`
	ArchetypeStats  map[ArchetypeID]int `json:"archetype_stats"`
	TagDistribution map[string]int      `json:"tag_distribution"`
	GroupSizes      map[string]int      `json:"group_sizes"`
}

// ==============================================
// Entity Relationship Management
// ==============================================

// EntityRelationship represents parent-child relationships between entities.
type EntityRelationship struct {
	Parent   EntityID   `json:"parent"`
	Children []EntityID `json:"children"`
	Depth    int        `json:"depth"`
}

// HierarchyManager manages entity parent-child relationships.
type HierarchyManager interface {
	// Relationship operations
	SetParent(child EntityID, parent EntityID) error
	RemoveParent(EntityID) error
	GetParent(EntityID) (EntityID, bool)
	GetChildren(EntityID) []EntityID
	HasChildren(EntityID) bool

	// Hierarchy traversal
	GetDescendants(EntityID) []EntityID
	GetAncestors(EntityID) []EntityID
	GetSiblings(EntityID) []EntityID
	GetRoot(EntityID) EntityID
	GetDepth(EntityID) int
	GetMaxDepth() int

	// Hierarchy validation
	IsAncestor(ancestor EntityID, descendant EntityID) bool
	WouldCreateCycle(child EntityID, parent EntityID) bool
	ValidateHierarchy() error

	// Hierarchy events
	OnHierarchyChanged(func(EntityID, EntityID, EntityID)) error

	// Batch operations
	SetParents(map[EntityID]EntityID) error
	RemoveParents([]EntityID) error

	// Serialization
	SerializeHierarchy() ([]byte, error)
	DeserializeHierarchy([]byte) error

	// Debug and statistics
	GetHierarchyStats() *HierarchyStats
	GetDebugInfo() string
}

// HierarchyStats contains statistics about entity hierarchies.
type HierarchyStats struct {
	TotalRelationships int     `json:"total_relationships"`
	MaxDepth           int     `json:"max_depth"`
	AverageDepth       float64 `json:"average_depth"`
	RootEntities       int     `json:"root_entities"`
	LeafEntities       int     `json:"leaf_entities"`
	OrphanedEntities   int     `json:"orphaned_entities"`
}

// ==============================================
// Entity Archetype System
// ==============================================

// ArchetypeManager manages entity archetypes for performance optimization.
// Entities with the same component types share the same archetype.
type ArchetypeManager interface {
	// Archetype operations
	GetArchetype(componentTypes []ComponentType) ArchetypeID
	CreateArchetype(componentTypes []ComponentType) ArchetypeID
	GetArchetypeComponents(ArchetypeID) []ComponentType
	GetArchetypeEntities(ArchetypeID) []EntityID

	// Entity archetype management
	SetEntityArchetype(EntityID, ArchetypeID) error
	GetEntityArchetype(EntityID) (ArchetypeID, bool)
	UpdateEntityArchetype(EntityID, []ComponentType) error

	// Archetype queries
	GetArchetypesWithComponent(ComponentType) []ArchetypeID
	GetArchetypesWithoutComponent(ComponentType) []ArchetypeID
	GetArchetypesWithAll([]ComponentType) []ArchetypeID
	GetArchetypesWithAny([]ComponentType) []ArchetypeID

	// Statistics and optimization
	GetArchetypeStats() map[ArchetypeID]*ArchetypeStats
	GetMostCommonArchetypes(int) []ArchetypeID
	OptimizeArchetypes() error
	CompactArchetypes() error

	// Serialization
	SerializeArchetypes() ([]byte, error)
	DeserializeArchetypes([]byte) error

	// Debug information
	GetDebugInfo() *ArchetypeManagerDebugInfo
}

// ArchetypeStats contains statistics about a specific archetype.
type ArchetypeStats struct {
	ID             ArchetypeID     `json:"id"`
	ComponentTypes []ComponentType `json:"component_types"`
	EntityCount    int             `json:"entity_count"`
	MemoryUsage    int64           `json:"memory_usage_bytes"`
	AccessCount    int64           `json:"access_count"`
	LastAccessed   time.Time       `json:"last_accessed"`
	CreatedAt      time.Time       `json:"created_at"`
}

// ArchetypeManagerDebugInfo provides debugging information about archetypes.
type ArchetypeManagerDebugInfo struct {
	ArchetypeCount     int                             `json:"archetype_count"`
	TotalEntities      int                             `json:"total_entities"`
	MemoryUsage        int64                           `json:"memory_usage_bytes"`
	MostCommon         []ArchetypeID                   `json:"most_common_archetypes"`
	LeastCommon        []ArchetypeID                   `json:"least_common_archetypes"`
	ArchetypeStats     map[ArchetypeID]*ArchetypeStats `json:"archetype_stats"`
	ComponentFrequency map[ComponentType]int           `json:"component_frequency"`
}

// ==============================================
// Entity Event System
// ==============================================

// EntityEvent represents events related to entity lifecycle.
type EntityEvent struct {
	Type      EntityEventType `json:"type"`
	EntityID  EntityID        `json:"entity_id"`
	Timestamp time.Time       `json:"timestamp"`
	Data      interface{}     `json:"data,omitempty"`
}

// EntityEventType represents different types of entity events.
type EntityEventType string

const (
	EntityEventCreated          EntityEventType = "created"
	EntityEventDestroyed        EntityEventType = "destroyed"
	EntityEventParentChanged    EntityEventType = "parent_changed"
	EntityEventTagChanged       EntityEventType = "tag_changed"
	EntityEventGroupAdded       EntityEventType = "group_added"
	EntityEventGroupRemoved     EntityEventType = "group_removed"
	EntityEventArchetypeChanged EntityEventType = "archetype_changed"
)

// EntityEventHandler processes entity events.
type EntityEventHandler func(*EntityEvent) error

// EntityEventBus manages entity-related events.
type EntityEventBus interface {
	// Event publishing
	PublishEvent(*EntityEvent) error
	PublishEventAsync(*EntityEvent) error

	// Event subscription
	Subscribe(EntityEventType, EntityEventHandler) error
	Unsubscribe(EntityEventType, EntityEventHandler) error
	SubscribeAll(EntityEventHandler) error

	// Event filtering
	SubscribeFiltered(EntityEventType, func(*EntityEvent) bool, EntityEventHandler) error

	// Event history
	GetEventHistory(EntityID) []*EntityEvent
	ClearEventHistory(EntityID) error
	GetAllEvents() []*EntityEvent

	// Configuration
	SetMaxHistorySize(int)
	SetAsyncBufferSize(int)
	EnableHistory(bool)

	// Statistics
	GetEventStats() *EntityEventStats
}

// EntityEventStats contains statistics about entity events.
type EntityEventStats struct {
	TotalEvents     int64                     `json:"total_events"`
	EventsByType    map[EntityEventType]int64 `json:"events_by_type"`
	SubscriberCount int                       `json:"subscriber_count"`
	HistorySize     int                       `json:"history_size"`
	QueueSize       int                       `json:"queue_size"`
	ProcessedEvents int64                     `json:"processed_events"`
	FailedEvents    int64                     `json:"failed_events"`
}

// ==============================================
// Thread-Safe Entity Collections
// ==============================================

// EntitySet provides thread-safe entity collections for high-performance access.
type EntitySet struct {
	entities map[EntityID]bool
	mutex    sync.RWMutex
	version  uint64 // For change detection
}

// NewEntitySet creates a new thread-safe entity set.
func NewEntitySet() *EntitySet {
	return &EntitySet{
		entities: make(map[EntityID]bool),
		version:  0,
	}
}

// Add adds an entity to the set.
func (s *EntitySet) Add(entity EntityID) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.entities[entity] = true
	s.version++
}

// Remove removes an entity from the set.
func (s *EntitySet) Remove(entity EntityID) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	existed := s.entities[entity]
	delete(s.entities, entity)
	if existed {
		s.version++
	}
	return existed
}

// Contains checks if an entity exists in the set.
func (s *EntitySet) Contains(entity EntityID) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.entities[entity]
}

// ToSlice returns all entities as a slice.
func (s *EntitySet) ToSlice() []EntityID {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entities := make([]EntityID, 0, len(s.entities))
	for entity := range s.entities {
		entities = append(entities, entity)
	}
	return entities
}

// Len returns the number of entities in the set.
func (s *EntitySet) Len() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.entities)
}

// Clear removes all entities from the set.
func (s *EntitySet) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.entities = make(map[EntityID]bool)
	s.version++
}

// GetVersion returns the current version for change detection.
func (s *EntitySet) GetVersion() uint64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.version
}

// Clone creates a deep copy of the entity set.
func (s *EntitySet) Clone() *EntitySet {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	cloned := &EntitySet{
		entities: make(map[EntityID]bool, len(s.entities)),
		version:  s.version,
	}

	for entity := range s.entities {
		cloned.entities[entity] = true
	}

	return cloned
}

// Union returns a new set containing entities from both sets.
func (s *EntitySet) Union(other *EntitySet) *EntitySet {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	other.mutex.RLock()
	defer other.mutex.RUnlock()

	result := NewEntitySet()

	for entity := range s.entities {
		result.entities[entity] = true
	}

	for entity := range other.entities {
		result.entities[entity] = true
	}

	return result
}

// Intersection returns a new set containing entities present in both sets.
func (s *EntitySet) Intersection(other *EntitySet) *EntitySet {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	other.mutex.RLock()
	defer other.mutex.RUnlock()

	result := NewEntitySet()

	for entity := range s.entities {
		if other.entities[entity] {
			result.entities[entity] = true
		}
	}

	return result
}

// Difference returns a new set containing entities in this set but not in other.
func (s *EntitySet) Difference(other *EntitySet) *EntitySet {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	other.mutex.RLock()
	defer other.mutex.RUnlock()

	result := NewEntitySet()

	for entity := range s.entities {
		if !other.entities[entity] {
			result.entities[entity] = true
		}
	}

	return result
}
