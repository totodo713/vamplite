package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"muscle-dreamer/internal/core/ecs"
	"muscle-dreamer/internal/core/ecs/components"
)

func Test_MemoryPool_CreateAndInitialize(t *testing.T) {
	// Arrange & Act
	pool := NewMemoryPool(ecs.ComponentTypeTransform, 10)

	// Assert
	assert.NotNil(t, pool)
	assert.Equal(t, ecs.ComponentTypeTransform, pool.GetComponentType())
	assert.Equal(t, 0, pool.GetUsedCount())
	assert.Equal(t, 10, pool.GetCapacity())
	assert.Equal(t, 10, pool.GetAvailableCount())
}

func Test_MemoryPool_AcquireComponent(t *testing.T) {
	// Arrange
	pool := NewMemoryPool(ecs.ComponentTypeTransform, 5)

	// Act
	component, err := pool.Acquire()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, component)
	assert.Equal(t, ecs.ComponentTypeTransform, component.GetType())
	assert.Equal(t, 1, pool.GetUsedCount())
	assert.Equal(t, 4, pool.GetAvailableCount())
}

func Test_MemoryPool_AcquireExceedsCapacity(t *testing.T) {
	// Arrange
	pool := NewMemoryPool(ecs.ComponentTypeSprite, 2)

	// Acquire all components
	component1, _ := pool.Acquire()
	component2, _ := pool.Acquire()

	// Act - Try to acquire one more
	component3, err := pool.Acquire()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, component3)
	assert.Contains(t, err.Error(), "pool capacity exceeded")
	assert.Equal(t, 2, pool.GetUsedCount())
	assert.Equal(t, 0, pool.GetAvailableCount())

	// Ensure previously acquired components are still valid
	assert.NotNil(t, component1)
	assert.NotNil(t, component2)
}

func Test_MemoryPool_ReleaseComponent(t *testing.T) {
	// Arrange
	pool := NewMemoryPool(ecs.ComponentTypePhysics, 3)
	component, _ := pool.Acquire()

	// Act
	err := pool.Release(component)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 0, pool.GetUsedCount())
	assert.Equal(t, 3, pool.GetAvailableCount())
}

func Test_MemoryPool_ReleaseInvalidComponent(t *testing.T) {
	// Arrange
	pool := NewMemoryPool(ecs.ComponentTypeHealth, 3)
	invalidComponent := components.NewTransformComponent() // Wrong type

	// Act
	err := pool.Release(invalidComponent)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "component does not belong to this pool")
	assert.Equal(t, 0, pool.GetUsedCount())
	assert.Equal(t, 3, pool.GetAvailableCount())
}

func Test_MemoryPool_ReleaseNotAcquiredComponent(t *testing.T) {
	// Arrange
	pool := NewMemoryPool(ecs.ComponentTypeAI, 3)
	component1, _ := pool.Acquire()
	_, _ = pool.Acquire() // component2 not used further

	// Release component1
	pool.Release(component1)

	// Act - Try to release the same component again
	err := pool.Release(component1)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "component not currently acquired")
	assert.Equal(t, 1, pool.GetUsedCount()) // Only component2 should be acquired
	assert.Equal(t, 2, pool.GetAvailableCount())
}

func Test_MemoryPool_AcquireReleaseReuse(t *testing.T) {
	// Arrange
	pool := NewMemoryPool(ecs.ComponentTypeTransform, 2)

	// Act - Acquire, release, and acquire again
	component1, _ := pool.Acquire()
	originalPtr := component1

	pool.Release(component1)
	component2, _ := pool.Acquire()

	// Assert - Component should be reused
	assert.Equal(t, originalPtr, component2, "Pool should reuse released components")
	assert.Equal(t, 1, pool.GetUsedCount())
	assert.Equal(t, 1, pool.GetAvailableCount())
}

func Test_MemoryPool_Clear(t *testing.T) {
	// Arrange
	pool := NewMemoryPool(ecs.ComponentTypeSprite, 5)

	// Acquire some components
	component1, _ := pool.Acquire()
	component2, _ := pool.Acquire()
	component3, _ := pool.Acquire()

	// Act
	pool.Clear()

	// Assert
	assert.Equal(t, 0, pool.GetUsedCount())
	assert.Equal(t, 5, pool.GetAvailableCount())

	// Previously acquired components should be invalid after clear
	err := pool.Release(component1)
	assert.Error(t, err)
	err = pool.Release(component2)
	assert.Error(t, err)
	err = pool.Release(component3)
	assert.Error(t, err)
}

func Test_MemoryPool_Resize(t *testing.T) {
	// Arrange
	pool := NewMemoryPool(ecs.ComponentTypePhysics, 3)

	// Acquire all components
	_, _ = pool.Acquire() // component1 not used further
	_, _ = pool.Acquire() // component2 not used further
	_, _ = pool.Acquire() // component3 not used further

	// Act - Resize to larger capacity
	err := pool.Resize(5)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 5, pool.GetCapacity())
	assert.Equal(t, 3, pool.GetUsedCount())
	assert.Equal(t, 2, pool.GetAvailableCount())

	// Should be able to acquire more components
	component4, err := pool.Acquire()
	assert.NoError(t, err)
	assert.NotNil(t, component4)
}

func Test_MemoryPool_ResizeToSmallerCapacity(t *testing.T) {
	// Arrange
	pool := NewMemoryPool(ecs.ComponentTypeHealth, 5)

	// Acquire some components
	_, _ = pool.Acquire() // component1 not used further
	_, _ = pool.Acquire() // component2 not used further
	_, _ = pool.Acquire() // component3 not used further

	// Act - Try to resize to smaller capacity than used count
	err := pool.Resize(2)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be smaller than used count")
	assert.Equal(t, 5, pool.GetCapacity()) // Should remain unchanged
}

func Test_MemoryPool_GetStatistics(t *testing.T) {
	// Arrange
	pool := NewMemoryPool(ecs.ComponentTypeAI, 10)

	// Acquire some components
	for i := 0; i < 7; i++ {
		pool.Acquire()
	}

	// Act
	stats := pool.GetStatistics()

	// Assert
	assert.Equal(t, ecs.ComponentTypeAI, stats.ComponentType)
	assert.Equal(t, 7, stats.UsedCount)
	assert.Equal(t, 10, stats.Capacity)
	assert.Equal(t, 3, stats.AvailableCount)
	assert.Equal(t, 0.7, stats.UsageRatio)
	assert.Greater(t, stats.MemoryUsed, int64(0))
	assert.Greater(t, stats.MemoryAllocated, stats.MemoryUsed)
}

func Test_MemoryPool_Performance_AcquireRelease(t *testing.T) {
	// Arrange
	pool := NewMemoryPool(ecs.ComponentTypeTransform, 1000)
	components := make([]ecs.Component, 0, 500)

	// Act - Acquire many components
	for i := 0; i < 500; i++ {
		component, err := pool.Acquire()
		assert.NoError(t, err)
		components = append(components, component)
	}

	// Assert - All acquired successfully
	assert.Equal(t, 500, pool.GetUsedCount())
	assert.Equal(t, 500, pool.GetAvailableCount())

	// Act - Release all components
	for _, component := range components {
		err := pool.Release(component)
		assert.NoError(t, err)
	}

	// Assert - All released successfully
	assert.Equal(t, 0, pool.GetUsedCount())
	assert.Equal(t, 1000, pool.GetAvailableCount())
}

func Benchmark_MemoryPool_Acquire(b *testing.B) {
	pool := NewMemoryPool(ecs.ComponentTypeTransform, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Acquire()
	}
}

func Benchmark_MemoryPool_AcquireRelease(b *testing.B) {
	pool := NewMemoryPool(ecs.ComponentTypeSprite, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		component, _ := pool.Acquire()
		pool.Release(component)
	}
}

func Benchmark_MemoryPool_Statistics(b *testing.B) {
	pool := NewMemoryPool(ecs.ComponentTypePhysics, 1000)

	// Prepare some data
	for i := 0; i < 500; i++ {
		pool.Acquire()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.GetStatistics()
	}
}
