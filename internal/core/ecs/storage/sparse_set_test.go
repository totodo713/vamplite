package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"muscle-dreamer/internal/core/ecs"
)

func Test_SparseSet_CreateAndInitialize(t *testing.T) {
	// Arrange & Act
	sparseSet := NewSparseSet()

	// Assert
	assert.NotNil(t, sparseSet)
	assert.Equal(t, 0, sparseSet.Size())
	assert.True(t, sparseSet.IsEmpty())
}

func Test_SparseSet_AddEntity(t *testing.T) {
	// Arrange
	sparseSet := NewSparseSet()
	entityID := ecs.EntityID(123)

	// Act
	err := sparseSet.Add(entityID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, sparseSet.Contains(entityID))
	assert.Equal(t, 1, sparseSet.Size())
	assert.False(t, sparseSet.IsEmpty())
}

func Test_SparseSet_AddDuplicateEntity(t *testing.T) {
	// Arrange
	sparseSet := NewSparseSet()
	entityID := ecs.EntityID(123)
	err := sparseSet.Add(entityID)
	assert.NoError(t, err)

	// Act
	err = sparseSet.Add(entityID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.Equal(t, 1, sparseSet.Size())
}

func Test_SparseSet_RemoveEntity(t *testing.T) {
	// Arrange
	sparseSet := NewSparseSet()
	entityID := ecs.EntityID(456)
	err := sparseSet.Add(entityID)
	assert.NoError(t, err)

	// Act
	err = sparseSet.Remove(entityID)

	// Assert
	assert.NoError(t, err)
	assert.False(t, sparseSet.Contains(entityID))
	assert.Equal(t, 0, sparseSet.Size())
	assert.True(t, sparseSet.IsEmpty())
}

func Test_SparseSet_RemoveNonExistentEntity(t *testing.T) {
	// Arrange
	sparseSet := NewSparseSet()
	entityID := ecs.EntityID(789)

	// Act
	err := sparseSet.Remove(entityID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func Test_SparseSet_GetIndex(t *testing.T) {
	// Arrange
	sparseSet := NewSparseSet()
	entities := []ecs.EntityID{100, 200, 300}

	for _, entity := range entities {
		err := sparseSet.Add(entity)
		assert.NoError(t, err)
	}

	// Act & Assert
	for i, entity := range entities {
		index, err := sparseSet.GetIndex(entity)
		assert.NoError(t, err)
		assert.Equal(t, i, index)
	}
}

func Test_SparseSet_GetEntityByIndex(t *testing.T) {
	// Arrange
	sparseSet := NewSparseSet()
	entities := []ecs.EntityID{100, 200, 300}

	for _, entity := range entities {
		err := sparseSet.Add(entity)
		assert.NoError(t, err)
	}

	// Act & Assert
	for i, expectedEntity := range entities {
		entity, err := sparseSet.GetEntityByIndex(i)
		assert.NoError(t, err)
		assert.Equal(t, expectedEntity, entity)
	}
}

func Test_SparseSet_GetEntityByInvalidIndex(t *testing.T) {
	// Arrange
	sparseSet := NewSparseSet()
	err := sparseSet.Add(ecs.EntityID(100))
	assert.NoError(t, err)

	// Act
	_, err = sparseSet.GetEntityByIndex(10)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")
}

func Test_SparseSet_IterateEntities(t *testing.T) {
	// Arrange
	sparseSet := NewSparseSet()
	expectedEntities := []ecs.EntityID{100, 200, 300, 400}

	for _, entity := range expectedEntities {
		sparseSet.Add(entity)
	}

	// Act
	var actualEntities []ecs.EntityID
	sparseSet.Iterate(func(entity ecs.EntityID) bool {
		actualEntities = append(actualEntities, entity)
		return true // Continue iteration
	})

	// Assert
	assert.Equal(t, len(expectedEntities), len(actualEntities))
	for _, expected := range expectedEntities {
		assert.Contains(t, actualEntities, expected)
	}
}

func Test_SparseSet_IterateWithStop(t *testing.T) {
	// Arrange
	sparseSet := NewSparseSet()
	entities := []ecs.EntityID{100, 200, 300, 400, 500}

	for _, entity := range entities {
		err := sparseSet.Add(entity)
		assert.NoError(t, err)
	}

	// Act
	var visitedEntities []ecs.EntityID
	sparseSet.Iterate(func(entity ecs.EntityID) bool {
		visitedEntities = append(visitedEntities, entity)
		return len(visitedEntities) < 3 // Stop after 3 entities
	})

	// Assert
	assert.Equal(t, 3, len(visitedEntities))
}

func Test_SparseSet_Clear(t *testing.T) {
	// Arrange
	sparseSet := NewSparseSet()
	entities := []ecs.EntityID{100, 200, 300}

	for _, entity := range entities {
		err := sparseSet.Add(entity)
		assert.NoError(t, err)
	}
	assert.Equal(t, 3, sparseSet.Size())

	// Act
	sparseSet.Clear()

	// Assert
	assert.Equal(t, 0, sparseSet.Size())
	assert.True(t, sparseSet.IsEmpty())
	for _, entity := range entities {
		assert.False(t, sparseSet.Contains(entity))
	}
}

func Test_SparseSet_ToSlice(t *testing.T) {
	// Arrange
	sparseSet := NewSparseSet()
	expectedEntities := []ecs.EntityID{100, 200, 300}

	for _, entity := range expectedEntities {
		sparseSet.Add(entity)
	}

	// Act
	entities := sparseSet.ToSlice()

	// Assert
	assert.Equal(t, len(expectedEntities), len(entities))
	for _, expected := range expectedEntities {
		assert.Contains(t, entities, expected)
	}
}

func Test_SparseSet_Performance_LargeDataset(t *testing.T) {
	// Arrange
	sparseSet := NewSparseSet()
	entityCount := 10000

	// Act - Add entities
	for i := 0; i < entityCount; i++ {
		err := sparseSet.Add(ecs.EntityID(i))
		assert.NoError(t, err)
	}

	// Assert - All entities added
	assert.Equal(t, entityCount, sparseSet.Size())

	// Act - Check contains (should be O(1))
	for i := 0; i < entityCount; i++ {
		assert.True(t, sparseSet.Contains(ecs.EntityID(i)))
	}

	// Act - Remove every other entity
	for i := 0; i < entityCount; i += 2 {
		err := sparseSet.Remove(ecs.EntityID(i))
		assert.NoError(t, err)
	}

	// Assert - Half entities remain
	assert.Equal(t, entityCount/2, sparseSet.Size())

	// Remaining entities should be odd numbers
	for i := 1; i < entityCount; i += 2 {
		assert.True(t, sparseSet.Contains(ecs.EntityID(i)))
	}
}

func Benchmark_SparseSet_Add(b *testing.B) {
	sparseSet := NewSparseSet()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sparseSet.Add(ecs.EntityID(i))
	}
}

func Benchmark_SparseSet_Contains(b *testing.B) {
	sparseSet := NewSparseSet()

	// Prepare data
	for i := 0; i < 10000; i++ {
		_ = sparseSet.Add(ecs.EntityID(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sparseSet.Contains(ecs.EntityID(i % 10000))
	}
}

func Benchmark_SparseSet_Remove(b *testing.B) {
	// Prepare data
	sparseSet := NewSparseSet()
	for i := 0; i < b.N; i++ {
		_ = sparseSet.Add(ecs.EntityID(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sparseSet.Remove(ecs.EntityID(i))
	}
}

func Benchmark_SparseSet_Iterate(b *testing.B) {
	sparseSet := NewSparseSet()

	// Prepare data
	for i := 0; i < 10000; i++ {
		_ = sparseSet.Add(ecs.EntityID(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sparseSet.Iterate(func(entity ecs.EntityID) bool {
			return true // Continue iteration
		})
	}
}
