package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"muscle-dreamer/internal/core/ecs"
	"muscle-dreamer/internal/core/ecs/components"
)

func Test_ComponentStore_CreateAndInitialize(t *testing.T) {
	// Arrange & Act
	store := NewComponentStore()

	// Assert
	assert.NotNil(t, store)
	assert.Equal(t, 0, store.GetEntityCount())
	assert.Equal(t, 0, len(store.GetRegisteredTypes()))
}

func Test_ComponentStore_RegisterComponentType(t *testing.T) {
	// Arrange
	store := NewComponentStore()

	// Act
	err := store.RegisterComponentType(ecs.ComponentTypeTransform, 100)

	// Assert
	assert.NoError(t, err)
	registeredTypes := store.GetRegisteredTypes()
	assert.Contains(t, registeredTypes, ecs.ComponentTypeTransform)
	assert.True(t, store.IsRegistered(ecs.ComponentTypeTransform))
}

func Test_ComponentStore_RegisterDuplicateComponentType(t *testing.T) {
	// Arrange
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypeSprite, 50)

	// Act
	err := store.RegisterComponentType(ecs.ComponentTypeSprite, 100)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func Test_ComponentStore_AddComponent(t *testing.T) {
	// Arrange
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypeTransform, 10)

	entityID := ecs.EntityID(123)
	transform := components.NewTransformComponent()
	transform.SetPosition(ecs.Vector2{X: 10, Y: 20})

	// Act
	err := store.AddComponent(entityID, transform)

	// Assert
	assert.NoError(t, err)
	assert.True(t, store.HasComponent(entityID, ecs.ComponentTypeTransform))
	assert.Equal(t, 1, store.GetEntityCount())
}

func Test_ComponentStore_AddComponentUnregisteredType(t *testing.T) {
	// Arrange
	store := NewComponentStore()
	entityID := ecs.EntityID(123)
	transform := components.NewTransformComponent()

	// Act
	err := store.AddComponent(entityID, transform)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")
}

func Test_ComponentStore_AddDuplicateComponent(t *testing.T) {
	// Arrange
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypePhysics, 10)

	entityID := ecs.EntityID(456)
	physics1 := components.NewPhysicsComponent()
	physics2 := components.NewPhysicsComponent()

	store.AddComponent(entityID, physics1)

	// Act
	err := store.AddComponent(entityID, physics2)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already has component")
}

func Test_ComponentStore_GetComponent(t *testing.T) {
	// Arrange
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypeHealth, 10)

	entityID := ecs.EntityID(789)
	health := components.NewHealthComponent(100)

	store.AddComponent(entityID, health)

	// Act
	retrievedComponent, err := store.GetComponent(entityID, ecs.ComponentTypeHealth)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, retrievedComponent)
	assert.Equal(t, ecs.ComponentTypeHealth, retrievedComponent.GetType())

	retrievedHealth := retrievedComponent.(*components.HealthComponent)
	assert.Equal(t, 100, retrievedHealth.MaxHealth)
}

func Test_ComponentStore_GetNonExistentComponent(t *testing.T) {
	// Arrange
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypeAI, 10)
	entityID := ecs.EntityID(999)

	// Act
	component, err := store.GetComponent(entityID, ecs.ComponentTypeAI)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, component)
	assert.Contains(t, err.Error(), "not found")
}

func Test_ComponentStore_RemoveComponent(t *testing.T) {
	// Arrange
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypeSprite, 10)

	entityID := ecs.EntityID(555)
	sprite := components.NewSpriteComponent()

	store.AddComponent(entityID, sprite)
	assert.True(t, store.HasComponent(entityID, ecs.ComponentTypeSprite))

	// Act
	err := store.RemoveComponent(entityID, ecs.ComponentTypeSprite)

	// Assert
	assert.NoError(t, err)
	assert.False(t, store.HasComponent(entityID, ecs.ComponentTypeSprite))
}

func Test_ComponentStore_RemoveNonExistentComponent(t *testing.T) {
	// Arrange
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypeTransform, 10)
	entityID := ecs.EntityID(777)

	// Act
	err := store.RemoveComponent(entityID, ecs.ComponentTypeTransform)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func Test_ComponentStore_GetAllComponents(t *testing.T) {
	// Arrange
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypeTransform, 10)
	store.RegisterComponentType(ecs.ComponentTypeSprite, 10)
	store.RegisterComponentType(ecs.ComponentTypePhysics, 10)

	entityID := ecs.EntityID(888)

	transform := components.NewTransformComponent()
	sprite := components.NewSpriteComponent()
	physics := components.NewPhysicsComponent()

	store.AddComponent(entityID, transform)
	store.AddComponent(entityID, sprite)
	store.AddComponent(entityID, physics)

	// Act
	allComponents := store.GetAllComponents(entityID)

	// Assert
	assert.Equal(t, 3, len(allComponents))

	componentTypes := make(map[ecs.ComponentType]bool)
	for _, component := range allComponents {
		componentTypes[component.GetType()] = true
	}

	assert.True(t, componentTypes[ecs.ComponentTypeTransform])
	assert.True(t, componentTypes[ecs.ComponentTypeSprite])
	assert.True(t, componentTypes[ecs.ComponentTypePhysics])
}

func Test_ComponentStore_GetAllComponentsEmptyEntity(t *testing.T) {
	// Arrange
	store := NewComponentStore()
	entityID := ecs.EntityID(999)

	// Act
	allComponents := store.GetAllComponents(entityID)

	// Assert
	assert.Empty(t, allComponents)
}

func Test_ComponentStore_RemoveEntity(t *testing.T) {
	// Arrange
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypeTransform, 10)
	store.RegisterComponentType(ecs.ComponentTypeSprite, 10)

	entityID := ecs.EntityID(111)

	transform := components.NewTransformComponent()
	sprite := components.NewSpriteComponent()

	store.AddComponent(entityID, transform)
	store.AddComponent(entityID, sprite)

	assert.True(t, store.HasComponent(entityID, ecs.ComponentTypeTransform))
	assert.True(t, store.HasComponent(entityID, ecs.ComponentTypeSprite))

	// Act
	removedCount := store.RemoveEntity(entityID)

	// Assert
	assert.Equal(t, 2, removedCount)
	assert.False(t, store.HasComponent(entityID, ecs.ComponentTypeTransform))
	assert.False(t, store.HasComponent(entityID, ecs.ComponentTypeSprite))
	assert.Equal(t, 0, store.GetEntityCount())
}

func Test_ComponentStore_GetEntitiesWithComponent(t *testing.T) {
	// Arrange
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypeHealth, 10)

	entities := []ecs.EntityID{100, 200, 300, 400}

	// Add health component to some entities
	for i, entityID := range entities[:3] {
		health := components.NewHealthComponent((i + 1) * 50)
		store.AddComponent(entityID, health)
	}

	// Act
	entitiesWithHealth := store.GetEntitiesWithComponent(ecs.ComponentTypeHealth)

	// Assert
	assert.Equal(t, 3, len(entitiesWithHealth))
	for _, entityID := range entities[:3] {
		assert.Contains(t, entitiesWithHealth, entityID)
	}
	assert.NotContains(t, entitiesWithHealth, entities[3])
}

func Test_ComponentStore_GetComponentCount(t *testing.T) {
	// Arrange
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypeAI, 10)
	store.RegisterComponentType(ecs.ComponentTypePhysics, 10)

	entities := []ecs.EntityID{1, 2, 3, 4, 5}

	// Add AI components to 3 entities
	for _, entityID := range entities[:3] {
		ai := components.NewAIComponent()
		store.AddComponent(entityID, ai)
	}

	// Add Physics components to 2 entities
	for _, entityID := range entities[:2] {
		physics := components.NewPhysicsComponent()
		store.AddComponent(entityID, physics)
	}

	// Act & Assert
	assert.Equal(t, 3, store.GetComponentCount(ecs.ComponentTypeAI))
	assert.Equal(t, 2, store.GetComponentCount(ecs.ComponentTypePhysics))
	assert.Equal(t, 0, store.GetComponentCount(ecs.ComponentTypeSprite)) // Not registered/added
}

func Test_ComponentStore_Clear(t *testing.T) {
	// Arrange
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypeTransform, 10)
	store.RegisterComponentType(ecs.ComponentTypeSprite, 10)

	entities := []ecs.EntityID{1, 2, 3}

	for _, entityID := range entities {
		transform := components.NewTransformComponent()
		sprite := components.NewSpriteComponent()
		store.AddComponent(entityID, transform)
		store.AddComponent(entityID, sprite)
	}

	assert.Equal(t, 3, store.GetEntityCount())

	// Act
	store.Clear()

	// Assert
	assert.Equal(t, 0, store.GetEntityCount())
	for _, entityID := range entities {
		assert.False(t, store.HasComponent(entityID, ecs.ComponentTypeTransform))
		assert.False(t, store.HasComponent(entityID, ecs.ComponentTypeSprite))
	}
}

func Test_ComponentStore_GetStorageStatistics(t *testing.T) {
	// Arrange
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypeTransform, 100)
	store.RegisterComponentType(ecs.ComponentTypeSprite, 50)

	// Add some components
	for i := 0; i < 30; i++ {
		entityID := ecs.EntityID(i)
		transform := components.NewTransformComponent()
		store.AddComponent(entityID, transform)

		if i < 20 {
			sprite := components.NewSpriteComponent()
			store.AddComponent(entityID, sprite)
		}
	}

	// Act
	stats := store.GetStorageStatistics()

	// Assert
	assert.Equal(t, 2, len(stats))

	// Find transform and sprite stats
	var transformStats, spriteStats *ecs.StorageStats
	for _, stat := range stats {
		if stat.ComponentType == ecs.ComponentTypeTransform {
			transformStats = stat
		} else if stat.ComponentType == ecs.ComponentTypeSprite {
			spriteStats = stat
		}
	}

	assert.NotNil(t, transformStats)
	assert.NotNil(t, spriteStats)
	assert.Equal(t, 30, transformStats.ComponentCount)
	assert.Equal(t, 20, spriteStats.ComponentCount)
}

func Test_ComponentStore_Performance_LargeDataset(t *testing.T) {
	// Arrange
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypeTransform, 5000)
	store.RegisterComponentType(ecs.ComponentTypeSprite, 3000)
	store.RegisterComponentType(ecs.ComponentTypePhysics, 2000)

	entityCount := 1000

	// Act - Add many entities with multiple components
	for i := 0; i < entityCount; i++ {
		entityID := ecs.EntityID(i)

		// All entities get transform
		transform := components.NewTransformComponent()
		transform.SetPosition(ecs.Vector2{X: float64(i), Y: float64(i * 2)})
		err := store.AddComponent(entityID, transform)
		assert.NoError(t, err)

		// 70% get sprite
		if i < entityCount*7/10 {
			sprite := components.NewSpriteComponent()
			err := store.AddComponent(entityID, sprite)
			assert.NoError(t, err)
		}

		// 50% get physics
		if i < entityCount/2 {
			physics := components.NewPhysicsComponent()
			err := store.AddComponent(entityID, physics)
			assert.NoError(t, err)
		}
	}

	// Assert - Performance expectations
	assert.Equal(t, entityCount, store.GetEntityCount())
	assert.Equal(t, entityCount, store.GetComponentCount(ecs.ComponentTypeTransform))
	assert.Equal(t, entityCount*7/10, store.GetComponentCount(ecs.ComponentTypeSprite))
	assert.Equal(t, entityCount/2, store.GetComponentCount(ecs.ComponentTypePhysics))

	// Test query performance
	entitiesWithTransform := store.GetEntitiesWithComponent(ecs.ComponentTypeTransform)
	assert.Equal(t, entityCount, len(entitiesWithTransform))

	entitiesWithSprite := store.GetEntitiesWithComponent(ecs.ComponentTypeSprite)
	assert.Equal(t, entityCount*7/10, len(entitiesWithSprite))
}

func Benchmark_ComponentStore_AddComponent(b *testing.B) {
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypeTransform, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entityID := ecs.EntityID(i)
		transform := components.NewTransformComponent()
		store.AddComponent(entityID, transform)
	}
}

func Benchmark_ComponentStore_GetComponent(b *testing.B) {
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypeSprite, 10000)

	// Prepare data
	for i := 0; i < 10000; i++ {
		entityID := ecs.EntityID(i)
		sprite := components.NewSpriteComponent()
		store.AddComponent(entityID, sprite)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entityID := ecs.EntityID(i % 10000)
		store.GetComponent(entityID, ecs.ComponentTypeSprite)
	}
}

func Benchmark_ComponentStore_HasComponent(b *testing.B) {
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypePhysics, 10000)

	// Prepare data
	for i := 0; i < 10000; i++ {
		entityID := ecs.EntityID(i)
		physics := components.NewPhysicsComponent()
		store.AddComponent(entityID, physics)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entityID := ecs.EntityID(i % 10000)
		store.HasComponent(entityID, ecs.ComponentTypePhysics)
	}
}

func Benchmark_ComponentStore_GetEntitiesWithComponent(b *testing.B) {
	store := NewComponentStore()
	store.RegisterComponentType(ecs.ComponentTypeHealth, 10000)

	// Prepare data
	for i := 0; i < 10000; i++ {
		entityID := ecs.EntityID(i)
		health := components.NewHealthComponent(100)
		store.AddComponent(entityID, health)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.GetEntitiesWithComponent(ecs.ComponentTypeHealth)
	}
}
