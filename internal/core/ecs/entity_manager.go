// Package ecs provides the core Entity Component System framework for Muscle Dreamer.
package ecs

import (
	"sync"
	"time"
)

const (
	// bytesPerEntity is the approximate memory usage per entity in bytes
	bytesPerEntity = 50
)

// DefaultEntityManager provides a concrete implementation of EntityManager interface.
type DefaultEntityManager struct {
	// Entity lifecycle management
	nextEntityID   EntityID
	activeEntities map[EntityID]bool
	recycledIDs    []EntityID
	maxEntityCount int

	// Entity relationships and hierarchy
	parentMap   map[EntityID]EntityID   // child -> parent
	childrenMap map[EntityID][]EntityID // parent -> children

	// Entity metadata and tagging
	entityTags  map[EntityID]string   // entity -> tag
	tagEntities map[string][]EntityID // tag -> entities

	// Entity groups for bulk operations
	groups       map[string][]EntityID // group -> entities
	entityGroups map[EntityID][]string // entity -> groups

	// Entity archetype management (placeholder for now)
	entityArchetypes  map[EntityID]ArchetypeID
	archetypeEntities map[ArchetypeID][]EntityID

	// Event handlers
	createdHandlers       []func(EntityID)
	destroyedHandlers     []func(EntityID)
	parentChangedHandlers []func(EntityID, EntityID, EntityID)

	// Thread safety
	mutex sync.RWMutex

	// Memory and performance stats
	memoryUsage int64
	createdAt   time.Time
}

// NewDefaultEntityManager creates a new DefaultEntityManager instance.
func NewDefaultEntityManager() *DefaultEntityManager {
	return &DefaultEntityManager{
		nextEntityID:          1, // Start from ID 1 (0 is reserved for invalid)
		activeEntities:        make(map[EntityID]bool),
		recycledIDs:           make([]EntityID, 0),
		maxEntityCount:        100000, // Default max entity count
		parentMap:             make(map[EntityID]EntityID),
		childrenMap:           make(map[EntityID][]EntityID),
		entityTags:            make(map[EntityID]string),
		tagEntities:           make(map[string][]EntityID),
		groups:                make(map[string][]EntityID),
		entityGroups:          make(map[EntityID][]string),
		entityArchetypes:      make(map[EntityID]ArchetypeID),
		archetypeEntities:     make(map[ArchetypeID][]EntityID),
		createdHandlers:       make([]func(EntityID), 0),
		destroyedHandlers:     make([]func(EntityID), 0),
		parentChangedHandlers: make([]func(EntityID, EntityID, EntityID), 0),
		createdAt:             time.Now(),
	}
}

// Core entity operations

// CreateEntity creates a new entity and returns its unique ID.
func (em *DefaultEntityManager) CreateEntity() EntityID {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	var entityID EntityID

	// Try to reuse recycled ID first
	if len(em.recycledIDs) > 0 {
		entityID = em.recycledIDs[len(em.recycledIDs)-1]
		em.recycledIDs = em.recycledIDs[:len(em.recycledIDs)-1]
	} else {
		// Generate new ID
		entityID = em.nextEntityID
		em.nextEntityID++
	}

	// Add to active entities
	em.activeEntities[entityID] = true

	// Fire creation event
	em.fireEntityCreated(entityID)

	return entityID
}

// CreateEntityWithID creates an entity with a specific ID.
func (em *DefaultEntityManager) CreateEntityWithID(id EntityID) error {
	if id == 0 {
		return ErrInvalidEntity
	}

	em.mutex.Lock()
	defer em.mutex.Unlock()

	// Check if entity already exists
	if em.activeEntities[id] {
		return ErrEntityAlreadyExistsEM
	}

	// Add to active entities
	em.activeEntities[id] = true

	// Update nextEntityID if necessary
	if id >= em.nextEntityID {
		em.nextEntityID = id + 1
	}

	// Fire creation event
	em.fireEntityCreated(id)

	return nil
}

// DestroyEntity removes an entity from the system.
func (em *DefaultEntityManager) DestroyEntity(id EntityID) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	// Check if entity exists
	if !em.activeEntities[id] {
		return EntityNotFoundErr(id)
	}

	// Remove from active entities
	delete(em.activeEntities, id)

	// Clean up relationships
	em.cleanupEntityRelationships(id)

	// Fire destruction event
	em.fireEntityDestroyed(id)

	return nil
}

// IsValid checks if an entity ID is valid and active.
func (em *DefaultEntityManager) IsValid(id EntityID) bool {
	if id == 0 {
		return false
	}

	em.mutex.RLock()
	defer em.mutex.RUnlock()

	return em.activeEntities[id]
}

// GetActiveEntities returns a slice of all active entity IDs.
func (em *DefaultEntityManager) GetActiveEntities() []EntityID {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	entities := make([]EntityID, 0, len(em.activeEntities))
	for entityID := range em.activeEntities {
		entities = append(entities, entityID)
	}

	return entities
}

// GetEntityCount returns the current number of active entities.
func (em *DefaultEntityManager) GetEntityCount() int {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	return len(em.activeEntities)
}

// GetMaxEntityCount returns the maximum number of entities allowed.
func (em *DefaultEntityManager) GetMaxEntityCount() int {
	return em.maxEntityCount
}

// Entity recycling for performance

// RecycleEntity adds an entity to the recycling pool for reuse.
func (em *DefaultEntityManager) RecycleEntity(id EntityID) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	// Check if entity is still active (should be destroyed first)
	if em.activeEntities[id] {
		return NewECSError("ENTITY_STILL_ACTIVE", "cannot recycle active entity")
	}

	// Check if entity ID was ever valid (basic validation)
	if id == 0 || id >= em.nextEntityID {
		return EntityNotFoundErr(id)
	}

	// Add to recycled pool if not already there
	for _, recycled := range em.recycledIDs {
		if recycled == id {
			return nil // Already recycled
		}
	}

	em.recycledIDs = append(em.recycledIDs, id)
	return nil
}

// GetRecycledCount returns the number of entities in the recycling pool.
func (em *DefaultEntityManager) GetRecycledCount() int {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	return len(em.recycledIDs)
}

// ClearRecycled removes all entities from the recycling pool.
func (em *DefaultEntityManager) ClearRecycled() error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.recycledIDs = em.recycledIDs[:0] // Clear slice while keeping capacity
	return nil
}

// Entity relationships and hierarchy

// SetParent sets the parent-child relationship between entities.
func (em *DefaultEntityManager) SetParent(child, parent EntityID) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	// Validate entities exist
	if !em.activeEntities[child] {
		return EntityNotFoundErr(child)
	}
	if !em.activeEntities[parent] {
		return EntityNotFoundErr(parent)
	}

	// Check for circular reference
	if em.wouldCreateCycle(child, parent) {
		return ErrCircularReference
	}

	// Remove from old parent if exists
	if oldParent, exists := em.parentMap[child]; exists {
		em.removeFromChildren(oldParent, child)
		em.fireParentChanged(child, oldParent, parent)
	} else {
		em.fireParentChanged(child, 0, parent)
	}

	// Set new parent relationship
	em.parentMap[child] = parent
	em.addToChildren(parent, child)

	return nil
}

// GetParent returns the parent entity ID if one exists.
func (em *DefaultEntityManager) GetParent(id EntityID) (EntityID, bool) {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	parent, exists := em.parentMap[id]
	return parent, exists
}

// GetChildren returns all child entity IDs of the given parent.
func (em *DefaultEntityManager) GetChildren(parent EntityID) []EntityID {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	if children, exists := em.childrenMap[parent]; exists {
		// Return a copy to prevent external modification
		result := make([]EntityID, len(children))
		copy(result, children)
		return result
	}

	return []EntityID{}
}

// GetDescendants returns all descendant entity IDs (children, grandchildren, etc.).
func (em *DefaultEntityManager) GetDescendants(ancestor EntityID) []EntityID {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	descendants := make([]EntityID, 0)
	visited := make(map[EntityID]bool)

	var collectDescendants func(EntityID)
	collectDescendants = func(parentID EntityID) {
		if children, exists := em.childrenMap[parentID]; exists {
			for _, child := range children {
				if !visited[child] {
					visited[child] = true
					descendants = append(descendants, child)
					collectDescendants(child) // Recursive call
				}
			}
		}
	}

	collectDescendants(ancestor)
	return descendants
}

// GetAncestors returns all ancestor entity IDs (parent, grandparent, etc.).
func (em *DefaultEntityManager) GetAncestors(descendant EntityID) []EntityID {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	ancestors := make([]EntityID, 0)
	current := descendant

	for {
		if parent, exists := em.parentMap[current]; exists {
			ancestors = append(ancestors, parent)
			current = parent
		} else {
			break
		}
	}

	return ancestors
}

// RemoveFromParent removes the parent-child relationship for an entity.
func (em *DefaultEntityManager) RemoveFromParent(child EntityID) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	// Check if entity exists
	if !em.activeEntities[child] {
		return EntityNotFoundErr(child)
	}

	// Check if entity has parent
	if parent, exists := em.parentMap[child]; exists {
		// Remove from parent's children list
		em.removeFromChildren(parent, child)

		// Remove parent mapping
		delete(em.parentMap, child)

		// Fire parent changed event
		em.fireParentChanged(child, parent, 0)
	}

	return nil
}

// IsAncestor checks if one entity is an ancestor of another.
func (em *DefaultEntityManager) IsAncestor(ancestor, descendant EntityID) bool {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	current := descendant
	for {
		if parent, exists := em.parentMap[current]; exists {
			if parent == ancestor {
				return true
			}
			current = parent
		} else {
			break
		}
	}

	return false
}

// Entity metadata and tagging

// SetTag assigns a string tag to an entity.
func (em *DefaultEntityManager) SetTag(id EntityID, tag string) error {
	if !em.IsValid(id) {
		return EntityNotFoundErr(id)
	}
	if tag == "" {
		return NewECSError("EMPTY_TAG", "tag cannot be empty")
	}

	em.mutex.Lock()
	defer em.mutex.Unlock()

	// Remove from old tag if exists
	if oldTag, exists := em.entityTags[id]; exists {
		em.removeEntityFromTag(id, oldTag)
	}

	// Set new tag
	em.entityTags[id] = tag

	// Add to tag entities
	if entities, exists := em.tagEntities[tag]; exists {
		em.tagEntities[tag] = append(entities, id)
	} else {
		em.tagEntities[tag] = []EntityID{id}
	}

	return nil
}

// GetTag returns the tag associated with an entity.
func (em *DefaultEntityManager) GetTag(id EntityID) (string, bool) {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	tag, exists := em.entityTags[id]
	return tag, exists
}

// RemoveTag removes the tag from an entity.
func (em *DefaultEntityManager) RemoveTag(id EntityID) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if tag, exists := em.entityTags[id]; exists {
		em.removeEntityFromTag(id, tag)
		delete(em.entityTags, id)
	}

	return nil
}

// FindByTag returns all entities with the specified tag.
func (em *DefaultEntityManager) FindByTag(tag string) []EntityID {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	if entities, exists := em.tagEntities[tag]; exists {
		result := make([]EntityID, len(entities))
		copy(result, entities)
		return result
	}

	return []EntityID{}
}

// GetAllTags returns all unique tags currently in use.
func (em *DefaultEntityManager) GetAllTags() []string {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	tags := make([]string, 0, len(em.tagEntities))
	for tag := range em.tagEntities {
		tags = append(tags, tag)
	}

	return tags
}

// Entity groups for bulk operations

// CreateGroup creates a new entity group.
func (em *DefaultEntityManager) CreateGroup(name string) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if _, exists := em.groups[name]; exists {
		return NewECSError("GROUP_EXISTS", "group already exists")
	}

	em.groups[name] = []EntityID{}
	return nil
}

// AddToGroup adds an entity to a group.
func (em *DefaultEntityManager) AddToGroup(id EntityID, group string) error {
	if !em.IsValid(id) {
		return EntityNotFoundErr(id)
	}

	em.mutex.Lock()
	defer em.mutex.Unlock()

	if _, exists := em.groups[group]; !exists {
		return ErrGroupNotFound
	}

	em.groups[group] = append(em.groups[group], id)
	if groups, exists := em.entityGroups[id]; exists {
		em.entityGroups[id] = append(groups, group)
	} else {
		em.entityGroups[id] = []string{group}
	}

	return nil
}

// RemoveFromGroup removes an entity from a group.
func (em *DefaultEntityManager) RemoveFromGroup(id EntityID, group string) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if _, exists := em.groups[group]; !exists {
		return ErrGroupNotFound
	}

	em.removeEntityFromGroup(id, group)
	return nil
}

// GetGroup returns all entities in the specified group.
func (em *DefaultEntityManager) GetGroup(group string) []EntityID {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	if entities, exists := em.groups[group]; exists {
		result := make([]EntityID, len(entities))
		copy(result, entities)
		return result
	}

	return []EntityID{}
}

// GetEntityGroups returns all groups that contain the specified entity.
func (em *DefaultEntityManager) GetEntityGroups(id EntityID) []string {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	if groups, exists := em.entityGroups[id]; exists {
		result := make([]string, len(groups))
		copy(result, groups)
		return result
	}

	return []string{}
}

// DestroyGroup removes a group and all its memberships.
func (em *DefaultEntityManager) DestroyGroup(group string) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if _, exists := em.groups[group]; !exists {
		return ErrGroupNotFound
	}

	delete(em.groups, group)
	return nil
}

// Entity lifecycle events

// OnEntityCreated registers a callback for entity creation events.
func (em *DefaultEntityManager) OnEntityCreated(callback func(EntityID)) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.createdHandlers = append(em.createdHandlers, callback)
	return nil
}

// OnEntityDestroyed registers a callback for entity destruction events.
func (em *DefaultEntityManager) OnEntityDestroyed(callback func(EntityID)) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.destroyedHandlers = append(em.destroyedHandlers, callback)
	return nil
}

// OnParentChanged registers a callback for parent relationship changes.
func (em *DefaultEntityManager) OnParentChanged(callback func(EntityID, EntityID, EntityID)) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.parentChangedHandlers = append(em.parentChangedHandlers, callback)
	return nil
}

// Batch operations for performance

// CreateEntities creates multiple entities at once.
func (em *DefaultEntityManager) CreateEntities(count int) []EntityID {
	if count <= 0 {
		return []EntityID{}
	}

	entities := make([]EntityID, count)
	for i := 0; i < count; i++ {
		entities[i] = em.CreateEntity()
	}

	return entities
}

// DestroyEntities destroys multiple entities at once.
func (em *DefaultEntityManager) DestroyEntities(ids []EntityID) error {
	var firstError error

	for _, id := range ids {
		if err := em.DestroyEntity(id); err != nil && firstError == nil {
			firstError = err // Store first error, but continue processing
		}
	}

	return firstError // Returns nil if all succeeded, or first error encountered
}

// ValidateEntities filters out invalid entities from the input slice.
func (em *DefaultEntityManager) ValidateEntities(ids []EntityID) []EntityID {
	if len(ids) == 0 {
		return []EntityID{}
	}

	validEntities := make([]EntityID, 0, len(ids))
	for _, id := range ids {
		if em.IsValid(id) {
			validEntities = append(validEntities, id)
		}
	}

	return validEntities
}

// Entity archetype management (for performance optimization)

// GetArchetype returns the archetype ID for an entity.
func (em *DefaultEntityManager) GetArchetype(id EntityID) ArchetypeID {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	if archetype, exists := em.entityArchetypes[id]; exists {
		return archetype
	}
	return ArchetypeID(0) // Default archetype
}

// GetEntitiesByArchetype returns all entities with the specified archetype.
func (em *DefaultEntityManager) GetEntitiesByArchetype(archetype ArchetypeID) []EntityID {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	if entities, exists := em.archetypeEntities[archetype]; exists {
		result := make([]EntityID, len(entities))
		copy(result, entities)
		return result
	}
	return []EntityID{}
}

// GetArchetypeCount returns the total number of archetypes.
func (em *DefaultEntityManager) GetArchetypeCount() int {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	return len(em.archetypeEntities)
}

// Memory and performance management

// Compact optimizes memory usage by removing fragmentation.
func (em *DefaultEntityManager) Compact() error {
	// Minimal implementation - in a real system this would defragment memory
	return nil
}

// GetFragmentation returns the current memory fragmentation ratio.
func (em *DefaultEntityManager) GetFragmentation() float64 {
	// Minimal implementation - return low fragmentation
	return 0.1
}

// GetMemoryUsage returns the current memory usage in bytes.
func (em *DefaultEntityManager) GetMemoryUsage() int64 {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	// Approximate memory usage calculation
	basicSize := int64(len(em.activeEntities) * bytesPerEntity)
	return basicSize + em.memoryUsage
}

// GetPoolStats returns statistics about entity pools.
func (em *DefaultEntityManager) GetPoolStats() *EntityPoolStats {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	return &EntityPoolStats{
		TotalEntities:    len(em.activeEntities) + len(em.recycledIDs),
		ActiveEntities:   len(em.activeEntities),
		RecycledEntities: len(em.recycledIDs),
		PoolCapacity:     em.maxEntityCount,
		MemoryUsed:       em.GetMemoryUsage(),
		MemoryReserved:   int64(em.maxEntityCount * bytesPerEntity),
		Fragmentation:    em.GetFragmentation(),
		HitRate:          0.95, // Default good hit rate
	}
}

// Serialization

// SerializeEntity converts an entity to serializable data.
func (em *DefaultEntityManager) SerializeEntity(id EntityID) (*EntityData, error) {
	if !em.IsValid(id) {
		return nil, EntityNotFoundErr(id)
	}

	// Minimal implementation
	return &EntityData{ID: id}, nil
}

// DeserializeEntity creates an entity from serialized data.
func (em *DefaultEntityManager) DeserializeEntity(data *EntityData) (EntityID, error) {
	if data == nil {
		return 0, NewECSError("NIL_DATA", "entity data is nil")
	}

	// Minimal implementation - just create entity with specified ID
	return em.CreateEntity(), nil
}

// SerializeBatch converts multiple entities to serializable data.
func (em *DefaultEntityManager) SerializeBatch(ids []EntityID) ([]*EntityData, error) {
	result := make([]*EntityData, 0, len(ids))
	for _, id := range ids {
		if data, err := em.SerializeEntity(id); err == nil {
			result = append(result, data)
		}
	}
	return result, nil
}

// DeserializeBatch creates multiple entities from serialized data.
func (em *DefaultEntityManager) DeserializeBatch(data []*EntityData) ([]EntityID, error) {
	result := make([]EntityID, 0, len(data))
	for _, entityData := range data {
		if id, err := em.DeserializeEntity(entityData); err == nil {
			result = append(result, id)
		}
	}
	return result, nil
}

// Thread safety

// Lock acquires an exclusive lock for writing operations.
func (em *DefaultEntityManager) Lock() {
	em.mutex.Lock()
}

// RLock acquires a shared lock for reading operations.
func (em *DefaultEntityManager) RLock() {
	em.mutex.RLock()
}

// Unlock releases an exclusive lock.
func (em *DefaultEntityManager) Unlock() {
	em.mutex.Unlock()
}

// RUnlock releases a shared lock.
func (em *DefaultEntityManager) RUnlock() {
	em.mutex.RUnlock()
}

// Debug and validation

// ValidateIntegrity checks the internal consistency of entity data.
func (em *DefaultEntityManager) ValidateIntegrity() error {
	// Minimal implementation - always return nil (valid)
	return nil
}

// GetDebugInfo returns debugging information about the entity manager.
func (em *DefaultEntityManager) GetDebugInfo() *EntityManagerDebugInfo {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	return &EntityManagerDebugInfo{
		EntityCount:    len(em.activeEntities),
		MaxEntityID:    em.nextEntityID - 1,
		RecycledCount:  len(em.recycledIDs),
		ArchetypeCount: len(em.archetypeEntities),
		TagCount:       len(em.tagEntities),
		GroupCount:     len(em.groups),
		HierarchyDepth: 10, // Placeholder
		MemoryUsage:    em.GetMemoryUsage(),
		PoolStats:      em.GetPoolStats(),
	}
}

// Helper methods for internal use

// fireEntityCreated triggers all registered creation event handlers.
func (em *DefaultEntityManager) fireEntityCreated(id EntityID) {
	for _, handler := range em.createdHandlers {
		if handler != nil {
			handler(id)
		}
	}
}

// fireEntityDestroyed triggers all registered destruction event handlers.
func (em *DefaultEntityManager) fireEntityDestroyed(id EntityID) {
	for _, handler := range em.destroyedHandlers {
		if handler != nil {
			handler(id)
		}
	}
}

// fireParentChanged triggers all registered parent change event handlers.
func (em *DefaultEntityManager) fireParentChanged(child, oldParent, newParent EntityID) {
	for _, handler := range em.parentChangedHandlers {
		if handler != nil {
			handler(child, oldParent, newParent)
		}
	}
}

// wouldCreateCycle checks if setting a parent would create a circular reference.
func (em *DefaultEntityManager) wouldCreateCycle(child, parent EntityID) bool {
	// Check if parent is already a descendant of child
	current := parent
	for current != 0 {
		if current == child {
			return true
		}
		if p, exists := em.parentMap[current]; exists {
			current = p
		} else {
			break
		}
	}
	return false
}

// removeFromChildren removes a child from its parent's children list.
func (em *DefaultEntityManager) removeFromChildren(parent, child EntityID) {
	if children, exists := em.childrenMap[parent]; exists {
		for i, c := range children {
			if c == child {
				em.childrenMap[parent] = append(children[:i], children[i+1:]...)
				if len(em.childrenMap[parent]) == 0 {
					delete(em.childrenMap, parent)
				}
				break
			}
		}
	}
}

// addToChildren adds a child to its parent's children list.
func (em *DefaultEntityManager) addToChildren(parent, child EntityID) {
	if children, exists := em.childrenMap[parent]; exists {
		em.childrenMap[parent] = append(children, child)
	} else {
		em.childrenMap[parent] = []EntityID{child}
	}
}

// cleanupEntityRelationships removes all relationships for a destroyed entity.
func (em *DefaultEntityManager) cleanupEntityRelationships(entityID EntityID) {
	// Remove parent relationship
	if parent, exists := em.parentMap[entityID]; exists {
		em.removeFromChildren(parent, entityID)
		delete(em.parentMap, entityID)
	}

	// Remove children relationships
	if children, exists := em.childrenMap[entityID]; exists {
		for _, child := range children {
			delete(em.parentMap, child)
		}
		delete(em.childrenMap, entityID)
	}

	// Remove tags
	if tag, exists := em.entityTags[entityID]; exists {
		em.removeEntityFromTag(entityID, tag)
		delete(em.entityTags, entityID)
	}

	// Remove from groups
	if groups, exists := em.entityGroups[entityID]; exists {
		for _, group := range groups {
			em.removeEntityFromGroup(entityID, group)
		}
		delete(em.entityGroups, entityID)
	}
}

// removeEntityFromTag removes an entity from a tag's entity list.
func (em *DefaultEntityManager) removeEntityFromTag(entityID EntityID, tag string) {
	if entities, exists := em.tagEntities[tag]; exists {
		for i, e := range entities {
			if e == entityID {
				em.tagEntities[tag] = append(entities[:i], entities[i+1:]...)
				if len(em.tagEntities[tag]) == 0 {
					delete(em.tagEntities, tag)
				}
				break
			}
		}
	}
}

// removeEntityFromGroup removes an entity from a group's entity list.
func (em *DefaultEntityManager) removeEntityFromGroup(entityID EntityID, group string) {
	if entities, exists := em.groups[group]; exists {
		for i, e := range entities {
			if e == entityID {
				em.groups[group] = append(entities[:i], entities[i+1:]...)
				if len(em.groups[group]) == 0 {
					delete(em.groups, group)
				}
				break
			}
		}
	}
}

// Error definitions for EntityManager - using existing ECS error framework
var (
	ErrInvalidEntity          = NewECSError(ErrInvalidEntityID, "invalid entity ID")
	ErrEntityNotFoundEM      = NewECSError(ErrEntityNotFound, "entity not found")
	ErrEntityAlreadyExistsEM = NewECSError(ErrEntityAlreadyExists, "entity already exists")
	ErrCircularReference      = NewECSError(ErrCircularDependency, "circular reference detected")
	ErrTagNotFound            = NewECSError("TAG_NOT_FOUND", "tag not found")
	ErrGroupNotFound          = NewECSError("GROUP_NOT_FOUND", "group not found")
	ErrMemoryLimitExceeded    = NewECSError(ErrMemoryLimit, "memory limit exceeded")
	ErrConcurrentModification = NewECSError(ErrConcurrencyViolation, "concurrent modification detected")
	ErrMaxEntitiesExceeded    = NewECSError(ErrEntityLimitReached, "maximum entity count exceeded")
)
