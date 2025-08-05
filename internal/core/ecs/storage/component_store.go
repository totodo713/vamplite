package storage

import (
	"fmt"
	"sync"

	"muscle-dreamer/internal/core/ecs"
)

// ComponentStore manages component storage across all component types
type ComponentStore struct {
	// Component storage by type
	sparseSets  map[ecs.ComponentType]*SparseSet
	memoryPools map[ecs.ComponentType]*MemoryPool
	components  map[ecs.ComponentType]map[ecs.EntityID]ecs.Component

	// Entity tracking
	entities map[ecs.EntityID]map[ecs.ComponentType]bool

	// Thread safety
	mutex sync.RWMutex

	// Configuration
	registeredTypes map[ecs.ComponentType]bool
}

// NewComponentStore creates a new component store
func NewComponentStore() *ComponentStore {
	return &ComponentStore{
		sparseSets:      make(map[ecs.ComponentType]*SparseSet),
		memoryPools:     make(map[ecs.ComponentType]*MemoryPool),
		components:      make(map[ecs.ComponentType]map[ecs.EntityID]ecs.Component),
		entities:        make(map[ecs.EntityID]map[ecs.ComponentType]bool),
		registeredTypes: make(map[ecs.ComponentType]bool),
	}
}

// RegisterComponentType registers a component type with the store
func (s *ComponentStore) RegisterComponentType(componentType ecs.ComponentType, poolSize int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.registeredTypes[componentType] {
		return fmt.Errorf("component type %s already registered", componentType)
	}

	// Initialize storage structures for this component type
	s.sparseSets[componentType] = NewSparseSet()
	s.memoryPools[componentType] = NewMemoryPool(componentType, poolSize)
	s.components[componentType] = make(map[ecs.EntityID]ecs.Component)
	s.registeredTypes[componentType] = true

	return nil
}

// IsRegistered checks if a component type is registered
func (s *ComponentStore) IsRegistered(componentType ecs.ComponentType) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.registeredTypes[componentType]
}

// GetRegisteredTypes returns all registered component types
func (s *ComponentStore) GetRegisteredTypes() []ecs.ComponentType {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	types := make([]ecs.ComponentType, 0, len(s.registeredTypes))
	for componentType := range s.registeredTypes {
		types = append(types, componentType)
	}
	return types
}

// AddComponent adds a component to an entity (refactored)
func (s *ComponentStore) AddComponent(entity ecs.EntityID, component ecs.Component) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	componentType := component.GetType()

	if err := s.validateComponentAddition(entity, componentType); err != nil {
		return err
	}

	return s.addComponentInternal(entity, component)
}

// GetComponent retrieves a component from an entity
func (s *ComponentStore) GetComponent(entity ecs.EntityID, componentType ecs.ComponentType) (ecs.Component, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Check if component type is registered
	if !s.registeredTypes[componentType] {
		return nil, fmt.Errorf("component type %s not registered", componentType)
	}

	// Check if entity has this component
	component, exists := s.components[componentType][entity]
	if !exists {
		return nil, fmt.Errorf("component of type %s not found for entity %d", componentType, entity)
	}

	return component, nil
}

// HasComponent checks if an entity has a specific component type
func (s *ComponentStore) HasComponent(entity ecs.EntityID, componentType ecs.ComponentType) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if entityComponents, exists := s.entities[entity]; exists {
		return entityComponents[componentType]
	}
	return false
}

// RemoveComponent removes a component from an entity
func (s *ComponentStore) RemoveComponent(entity ecs.EntityID, componentType ecs.ComponentType) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if entity has this component
	if entityComponents, exists := s.entities[entity]; !exists || !entityComponents[componentType] {
		return fmt.Errorf("component of type %s not found for entity %d", componentType, entity)
	}

	// Remove from sparse set
	if err := s.sparseSets[componentType].Remove(entity); err != nil {
		return fmt.Errorf("failed to remove entity from sparse set: %w", err)
	}

	// Remove component
	delete(s.components[componentType], entity)

	// Update entity tracking
	delete(s.entities[entity], componentType)
	if len(s.entities[entity]) == 0 {
		delete(s.entities, entity)
	}

	return nil
}

// GetAllComponents returns all components for an entity
func (s *ComponentStore) GetAllComponents(entity ecs.EntityID) []ecs.Component {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entityComponents, exists := s.entities[entity]
	if !exists {
		return []ecs.Component{}
	}

	components := make([]ecs.Component, 0, len(entityComponents))
	for componentType := range entityComponents {
		if component, exists := s.components[componentType][entity]; exists {
			components = append(components, component)
		}
	}

	return components
}

// RemoveEntity removes all components from an entity
func (s *ComponentStore) RemoveEntity(entity ecs.EntityID) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	entityComponents, exists := s.entities[entity]
	if !exists {
		return 0
	}

	removedCount := 0
	for componentType := range entityComponents {
		// Remove from sparse set
		s.sparseSets[componentType].Remove(entity)

		// Remove component
		delete(s.components[componentType], entity)
		removedCount++
	}

	// Remove entity tracking
	delete(s.entities, entity)

	return removedCount
}

// GetEntitiesWithComponent returns all entities that have a specific component type
func (s *ComponentStore) GetEntitiesWithComponent(componentType ecs.ComponentType) []ecs.EntityID {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if sparseSet, exists := s.sparseSets[componentType]; exists {
		return sparseSet.ToSlice()
	}
	return []ecs.EntityID{}
}

// GetComponentCount returns the number of components of a specific type
func (s *ComponentStore) GetComponentCount(componentType ecs.ComponentType) int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if sparseSet, exists := s.sparseSets[componentType]; exists {
		return sparseSet.Size()
	}
	return 0
}

// GetEntityCount returns the total number of entities with at least one component
func (s *ComponentStore) GetEntityCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.entities)
}

// Clear removes all components and entities from the store
func (s *ComponentStore) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Clear all sparse sets
	for _, sparseSet := range s.sparseSets {
		sparseSet.Clear()
	}

	// Clear all memory pools
	for _, memoryPool := range s.memoryPools {
		memoryPool.Clear()
	}

	// Clear component storage
	for componentType := range s.components {
		s.components[componentType] = make(map[ecs.EntityID]ecs.Component)
	}

	// Clear entity tracking
	s.entities = make(map[ecs.EntityID]map[ecs.ComponentType]bool)
}

// GetStorageStatistics returns storage statistics for all component types
func (s *ComponentStore) GetStorageStatistics() []*ecs.StorageStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stats := make([]*ecs.StorageStats, 0, len(s.registeredTypes))

	for componentType := range s.registeredTypes {
		sparseSet := s.sparseSets[componentType]
		componentCount := sparseSet.Size()

		// Calculate memory usage estimates
		var memoryUsed, memoryReserved int64
		if componentCount > 0 {
			// Estimate based on component size
			for entity := range s.components[componentType] {
				component := s.components[componentType][entity]
				memoryUsed += int64(component.Size())
				break // Just use first component for size estimation
			}
			memoryUsed *= int64(componentCount)
		}

		memoryReserved = int64(sparseSet.Capacity() * 64) // Rough estimate

		stat := &ecs.StorageStats{
			ComponentType:  componentType,
			ComponentCount: componentCount,
			MemoryUsed:     memoryUsed,
			MemoryReserved: memoryReserved,
			Fragmentation:  0.0, // TODO: Calculate fragmentation
			AccessCount:    0,   // TODO: Track access count
			CacheHitRate:   1.0, // TODO: Track cache hit rate
		}

		stats = append(stats, stat)
	}

	return stats
}

// =============================================================================
// HELPER METHODS - TDD Refactor Phase
// =============================================================================

// validateComponentAddition validates if a component can be added to an entity
func (s *ComponentStore) validateComponentAddition(entity ecs.EntityID, componentType ecs.ComponentType) error {
	// Check if component type is registered
	if !s.registeredTypes[componentType] {
		return fmt.Errorf("component type %s not registered", componentType)
	}

	// Check if entity already has this component type
	if entityComponents, exists := s.entities[entity]; exists {
		if entityComponents[componentType] {
			return fmt.Errorf("entity %d already has component of type %s", entity, componentType)
		}
	}

	return nil
}

// addComponentInternal adds a component to an entity (assumes validation is done)
func (s *ComponentStore) addComponentInternal(entity ecs.EntityID, component ecs.Component) error {
	componentType := component.GetType()

	// Add entity to sparse set for this component type
	if err := s.sparseSets[componentType].Add(entity); err != nil {
		return fmt.Errorf("failed to add entity %d to sparse set: %w", entity, err)
	}

	// Store component
	s.components[componentType][entity] = component

	// Update entity tracking
	if s.entities[entity] == nil {
		s.entities[entity] = make(map[ecs.ComponentType]bool)
	}
	s.entities[entity][componentType] = true

	return nil
}

// removeComponentInternal removes a component from an entity (assumes validation is done)
func (s *ComponentStore) removeComponentInternal(entity ecs.EntityID, componentType ecs.ComponentType) error {
	// Remove from sparse set
	if err := s.sparseSets[componentType].Remove(entity); err != nil {
		return fmt.Errorf("failed to remove entity %d from sparse set: %w", entity, err)
	}

	// Remove component
	delete(s.components[componentType], entity)

	// Update entity tracking
	delete(s.entities[entity], componentType)
	if len(s.entities[entity]) == 0 {
		delete(s.entities, entity)
	}

	return nil
}

// =============================================================================
// BULK OPERATIONS - TDD Green Phase Implementation (Refactored)
// =============================================================================

// AddComponentsBatch adds multiple components to multiple entities in batch (refactored)
func (s *ComponentStore) AddComponentsBatch(entities []ecs.EntityID, components []ecs.Component) error {
	if len(entities) != len(components) {
		return fmt.Errorf("entities and components slice length mismatch: %d != %d", len(entities), len(components))
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Validate all operations first (fail-fast)
	for i, entity := range entities {
		componentType := components[i].GetType()
		if err := s.validateComponentAddition(entity, componentType); err != nil {
			return fmt.Errorf("validation failed for entity %d: %w", entity, err)
		}
	}

	// Execute all operations (should not fail after validation)
	for i, entity := range entities {
		if err := s.addComponentInternal(entity, components[i]); err != nil {
			// This should not happen after validation, but handle gracefully
			return fmt.Errorf("internal error adding component to entity %d: %w", entity, err)
		}
	}

	return nil
}

// RemoveComponentsBatch removes components of specified type from multiple entities
func (s *ComponentStore) RemoveComponentsBatch(entities []ecs.EntityID, componentType ecs.ComponentType) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Simple implementation: remove each component individually
	for _, entity := range entities {
		// Check if entity has this component
		if entityComponents, exists := s.entities[entity]; !exists || !entityComponents[componentType] {
			// Skip if entity doesn't have the component (not an error for batch operations)
			continue
		}

		// Remove from sparse set
		if err := s.sparseSets[componentType].Remove(entity); err != nil {
			return fmt.Errorf("failed to remove entity %d from sparse set: %w", entity, err)
		}

		// Remove component
		delete(s.components[componentType], entity)

		// Update entity tracking
		delete(s.entities[entity], componentType)
		if len(s.entities[entity]) == 0 {
			delete(s.entities, entity)
		}
	}

	return nil
}

// =============================================================================
// COMPLEX QUERY OPERATIONS - TDD Green Phase Implementation
// =============================================================================

// GetEntitiesWithMultipleComponents returns entities that have ALL specified component types (optimized)
func (s *ComponentStore) GetEntitiesWithMultipleComponents(componentTypes []ecs.ComponentType) []ecs.EntityID {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if len(componentTypes) == 0 {
		return []ecs.EntityID{}
	}

	// Find the component type with the smallest entity set to minimize iterations
	var smallestSet *SparseSet
	var smallestSize int = -1

	for _, componentType := range componentTypes {
		if sparseSet, exists := s.sparseSets[componentType]; exists {
			size := sparseSet.Size()
			if smallestSize == -1 || size < smallestSize {
				smallestSet = sparseSet
				smallestSize = size
			}
		} else {
			// If any component type doesn't exist, no entities can have all types
			return []ecs.EntityID{}
		}
	}

	if smallestSet == nil {
		return []ecs.EntityID{}
	}

	// Use the smallest set as candidates and check if they have all other component types
	candidates := smallestSet.ToSlice()
	var result []ecs.EntityID

	for _, entity := range candidates {
		hasAllComponents := true

		for _, componentType := range componentTypes {
			if entityComponents, exists := s.entities[entity]; !exists || !entityComponents[componentType] {
				hasAllComponents = false
				break
			}
		}

		if hasAllComponents {
			result = append(result, entity)
		}
	}

	return result
}
