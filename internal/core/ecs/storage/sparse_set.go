package storage

import (
	"fmt"

	"muscle-dreamer/internal/core/ecs"
)

// SparseSet provides O(1) entity operations for ECS component storage
type SparseSet struct {
	// sparse maps entity ID to dense index
	sparse map[ecs.EntityID]int

	// dense stores entity IDs in contiguous memory
	dense []ecs.EntityID

	// size tracks the number of entities
	size int
}

// NewSparseSet creates a new sparse set for entity storage
func NewSparseSet() *SparseSet {
	return &SparseSet{
		sparse: make(map[ecs.EntityID]int),
		dense:  make([]ecs.EntityID, 0, 1000), // Initial capacity
		size:   0,
	}
}

// Add adds an entity to the sparse set (O(1) operation)
func (s *SparseSet) Add(entity ecs.EntityID) error {
	if _, exists := s.sparse[entity]; exists {
		return fmt.Errorf("entity %d already exists in sparse set", entity)
	}

	// Add to dense array
	if s.size >= len(s.dense) {
		s.dense = append(s.dense, entity)
	} else {
		s.dense[s.size] = entity
	}

	// Map entity to its dense index
	s.sparse[entity] = s.size
	s.size++

	return nil
}

// Remove removes an entity from the sparse set (O(1) operation)
func (s *SparseSet) Remove(entity ecs.EntityID) error {
	index, exists := s.sparse[entity]
	if !exists {
		return fmt.Errorf("entity %d not found in sparse set", entity)
	}

	// Get the last entity in dense array
	lastIndex := s.size - 1
	lastEntity := s.dense[lastIndex]

	// Move last entity to the removed entity's position
	s.dense[index] = lastEntity
	s.sparse[lastEntity] = index

	// Remove the entity from sparse map
	delete(s.sparse, entity)
	s.size--

	return nil
}

// Contains checks if an entity exists in the sparse set (O(1) operation)
func (s *SparseSet) Contains(entity ecs.EntityID) bool {
	_, exists := s.sparse[entity]
	return exists
}

// Size returns the number of entities in the sparse set
func (s *SparseSet) Size() int {
	return s.size
}

// IsEmpty returns true if the sparse set is empty
func (s *SparseSet) IsEmpty() bool {
	return s.size == 0
}

// GetIndex returns the dense index of an entity (O(1) operation)
func (s *SparseSet) GetIndex(entity ecs.EntityID) (int, error) {
	index, exists := s.sparse[entity]
	if !exists {
		return -1, fmt.Errorf("entity %d not found in sparse set", entity)
	}
	return index, nil
}

// GetEntityByIndex returns the entity at a specific dense index
func (s *SparseSet) GetEntityByIndex(index int) (ecs.EntityID, error) {
	if index < 0 || index >= s.size {
		return ecs.InvalidEntityID, fmt.Errorf("index %d out of range [0, %d)", index, s.size)
	}
	return s.dense[index], nil
}

// Iterate iterates over all entities in the sparse set
// The callback function should return true to continue iteration, false to stop
func (s *SparseSet) Iterate(callback func(ecs.EntityID) bool) {
	for i := 0; i < s.size; i++ {
		if !callback(s.dense[i]) {
			break
		}
	}
}

// Clear removes all entities from the sparse set
func (s *SparseSet) Clear() {
	s.sparse = make(map[ecs.EntityID]int)
	s.size = 0
	// Keep the dense slice allocated but reset its logical size
}

// ToSlice returns all entities as a slice (creates a copy)
func (s *SparseSet) ToSlice() []ecs.EntityID {
	result := make([]ecs.EntityID, s.size)
	copy(result, s.dense[:s.size])
	return result
}

// Capacity returns the current capacity of the dense array
func (s *SparseSet) Capacity() int {
	return cap(s.dense)
}

// Reserve ensures the sparse set can hold at least the specified number of entities
func (s *SparseSet) Reserve(capacity int) {
	if capacity > cap(s.dense) {
		newDense := make([]ecs.EntityID, s.size, capacity)
		copy(newDense, s.dense[:s.size])
		s.dense = newDense
	}
}
