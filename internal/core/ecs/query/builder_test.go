package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"muscle-dreamer/internal/core/ecs"
)

// TestQueryBuilderImpl_BasicOperations tests basic query builder operations
func TestQueryBuilderImpl_BasicOperations(t *testing.T) {
	t.Run("With adds required components", func(t *testing.T) {
		qb := NewQueryBuilder()

		result := qb.With(ecs.ComponentTypeTransform).
			With(ecs.ComponentTypeSprite)

		impl := result.(*QueryBuilderImpl)
		assert.True(t, impl.requiredComponents.Has(ecs.ComponentTypeTransform))
		assert.True(t, impl.requiredComponents.Has(ecs.ComponentTypeSprite))
	})

	t.Run("Without adds excluded components", func(t *testing.T) {
		qb := NewQueryBuilder()

		result := qb.Without(ecs.ComponentTypeAI).
			Without(ecs.ComponentTypeDisabled)

		impl := result.(*QueryBuilderImpl)
		assert.True(t, impl.excludedComponents.Has(ecs.ComponentTypeAI))
		assert.True(t, impl.excludedComponents.Has(ecs.ComponentTypeDisabled))
	})

	t.Run("WithAll adds multiple required components", func(t *testing.T) {
		qb := NewQueryBuilder()
		components := []ecs.ComponentType{
			ecs.ComponentTypeTransform,
			ecs.ComponentTypeSprite,
			ecs.ComponentTypePhysics,
		}

		result := qb.WithAll(components)

		impl := result.(*QueryBuilderImpl)
		for _, comp := range components {
			assert.True(t, impl.requiredComponents.Has(comp))
		}
	})

	t.Run("WithAny adds optional components", func(t *testing.T) {
		qb := NewQueryBuilder()
		components := []ecs.ComponentType{
			ecs.ComponentTypeHealth,
			ecs.ComponentTypeEnergy,
		}

		result := qb.WithAny(components)

		impl := result.(*QueryBuilderImpl)
		for _, comp := range components {
			assert.True(t, impl.optionalComponents.Has(comp))
		}
	})

	t.Run("WithNone adds excluded components", func(t *testing.T) {
		qb := NewQueryBuilder()
		components := []ecs.ComponentType{
			ecs.ComponentTypeDisabled,
			ecs.ComponentTypeDead,
		}

		result := qb.WithNone(components)

		impl := result.(*QueryBuilderImpl)
		for _, comp := range components {
			assert.True(t, impl.excludedComponents.Has(comp))
		}
	})
}

// TestQueryBuilderImpl_FilteringOptions tests query filtering options
func TestQueryBuilderImpl_FilteringOptions(t *testing.T) {
	t.Run("Where adds custom filter", func(t *testing.T) {
		qb := NewQueryBuilder()

		filterFunc := func(id ecs.EntityID, components []ecs.Component) bool {
			return id > 100
		}

		result := qb.Where(filterFunc)

		impl := result.(*QueryBuilderImpl)
		require.NotNil(t, impl.customFilter)
	})

	t.Run("WhereComponent adds component filter", func(t *testing.T) {
		qb := NewQueryBuilder()

		filterFunc := func(c ecs.Component) bool {
			return c != nil
		}

		result := qb.WhereComponent(ecs.ComponentTypeTransform, filterFunc)

		impl := result.(*QueryBuilderImpl)
		require.NotNil(t, impl.componentFilters)
		require.NotNil(t, impl.componentFilters[ecs.ComponentTypeTransform])
	})

	t.Run("WhereEntity adds entity filter", func(t *testing.T) {
		qb := NewQueryBuilder()

		filterFunc := func(id ecs.EntityID) bool {
			return id%2 == 0
		}

		result := qb.WhereEntity(filterFunc)

		impl := result.(*QueryBuilderImpl)
		require.NotNil(t, impl.entityFilter)
	})
}

// TestQueryBuilderImpl_ResultModifiers tests result modification options
func TestQueryBuilderImpl_ResultModifiers(t *testing.T) {
	t.Run("Limit sets maximum results", func(t *testing.T) {
		qb := NewQueryBuilder()

		result := qb.Limit(100)

		impl := result.(*QueryBuilderImpl)
		assert.Equal(t, 100, impl.limit)
	})

	t.Run("Offset sets starting position", func(t *testing.T) {
		qb := NewQueryBuilder()

		result := qb.Offset(50)

		impl := result.(*QueryBuilderImpl)
		assert.Equal(t, 50, impl.offset)
	})

	t.Run("OrderBy sets sorting function", func(t *testing.T) {
		qb := NewQueryBuilder()

		sortFunc := func(a, b ecs.EntityID) bool {
			return a < b
		}

		result := qb.OrderBy(sortFunc)

		impl := result.(*QueryBuilderImpl)
		require.NotNil(t, impl.orderByFunc)
	})
}

// TestQueryBuilderImpl_PerformanceOptions tests performance optimization options
func TestQueryBuilderImpl_PerformanceOptions(t *testing.T) {
	t.Run("Cache sets cache key", func(t *testing.T) {
		qb := NewQueryBuilder()

		result := qb.Cache("enemies-query")

		impl := result.(*QueryBuilderImpl)
		assert.Equal(t, "enemies-query", impl.cacheKey)
	})

	t.Run("CacheFor sets cache duration", func(t *testing.T) {
		qb := NewQueryBuilder()
		duration := 5 * time.Minute

		result := qb.CacheFor(duration)

		impl := result.(*QueryBuilderImpl)
		assert.Equal(t, duration, impl.cacheDuration)
	})

	t.Run("UseBitset enables bitset optimization", func(t *testing.T) {
		qb := NewQueryBuilder()

		result := qb.UseBitset(true)

		impl := result.(*QueryBuilderImpl)
		assert.True(t, impl.useBitset)
	})

	t.Run("UseIndex sets index hint", func(t *testing.T) {
		qb := NewQueryBuilder()

		result := qb.UseIndex("spatial-index")

		impl := result.(*QueryBuilderImpl)
		assert.Equal(t, "spatial-index", impl.indexHint)
	})
}

// TestQueryBuilderImpl_SpatialQueries tests spatial query capabilities
func TestQueryBuilderImpl_SpatialQueries(t *testing.T) {
	t.Run("WithinRadius sets radius constraint", func(t *testing.T) {
		qb := NewQueryBuilder()
		center := ecs.Vector2{X: 100, Y: 200}

		result := qb.WithinRadius(center, 50.0)

		impl := result.(*QueryBuilderImpl)
		assert.Equal(t, center, impl.spatialCenter)
		assert.Equal(t, 50.0, impl.spatialRadius)
		assert.Equal(t, SpatialFilterRadius, impl.spatialFilter)
	})

	t.Run("WithinBounds sets bounds constraint", func(t *testing.T) {
		qb := NewQueryBuilder()
		bounds := ecs.AABB{
			Min: ecs.Vector2{X: 0, Y: 0},
			Max: ecs.Vector2{X: 100, Y: 100},
		}

		result := qb.WithinBounds(bounds)

		impl := result.(*QueryBuilderImpl)
		assert.Equal(t, bounds, impl.spatialBounds)
		assert.Equal(t, SpatialFilterBounds, impl.spatialFilter)
	})

	t.Run("Intersects sets intersection constraint", func(t *testing.T) {
		qb := NewQueryBuilder()
		bounds := ecs.AABB{
			Min: ecs.Vector2{X: 50, Y: 50},
			Max: ecs.Vector2{X: 150, Y: 150},
		}

		result := qb.Intersects(bounds)

		impl := result.(*QueryBuilderImpl)
		assert.Equal(t, bounds, impl.spatialBounds)
		assert.Equal(t, SpatialFilterIntersects, impl.spatialFilter)
	})

	t.Run("Nearest sets nearest constraint", func(t *testing.T) {
		qb := NewQueryBuilder()
		point := ecs.Vector2{X: 100, Y: 100}

		result := qb.Nearest(point, 10)

		impl := result.(*QueryBuilderImpl)
		assert.Equal(t, point, impl.spatialCenter)
		assert.Equal(t, 10, impl.nearestCount)
		assert.Equal(t, SpatialFilterNearest, impl.spatialFilter)
	})
}

// TestQueryBuilderImpl_HierarchicalQueries tests hierarchical query capabilities
func TestQueryBuilderImpl_HierarchicalQueries(t *testing.T) {
	t.Run("Children sets children filter", func(t *testing.T) {
		qb := NewQueryBuilder()
		parentID := ecs.EntityID(42)

		result := qb.Children(parentID)

		impl := result.(*QueryBuilderImpl)
		assert.Equal(t, parentID, impl.hierarchyRoot)
		assert.Equal(t, HierarchyFilterChildren, impl.hierarchyFilter)
	})

	t.Run("Descendants sets descendants filter", func(t *testing.T) {
		qb := NewQueryBuilder()
		rootID := ecs.EntityID(1)

		result := qb.Descendants(rootID)

		impl := result.(*QueryBuilderImpl)
		assert.Equal(t, rootID, impl.hierarchyRoot)
		assert.Equal(t, HierarchyFilterDescendants, impl.hierarchyFilter)
	})

	t.Run("Ancestors sets ancestors filter", func(t *testing.T) {
		qb := NewQueryBuilder()
		childID := ecs.EntityID(100)

		result := qb.Ancestors(childID)

		impl := result.(*QueryBuilderImpl)
		assert.Equal(t, childID, impl.hierarchyRoot)
		assert.Equal(t, HierarchyFilterAncestors, impl.hierarchyFilter)
	})

	t.Run("Siblings sets siblings filter", func(t *testing.T) {
		qb := NewQueryBuilder()
		entityID := ecs.EntityID(50)

		result := qb.Siblings(entityID)

		impl := result.(*QueryBuilderImpl)
		assert.Equal(t, entityID, impl.hierarchyRoot)
		assert.Equal(t, HierarchyFilterSiblings, impl.hierarchyFilter)
	})
}

// TestQueryBuilderImpl_Serialization tests query serialization
func TestQueryBuilderImpl_Serialization(t *testing.T) {
	t.Run("ToString generates string representation", func(t *testing.T) {
		qb := NewQueryBuilder().
			With(ecs.ComponentTypeTransform).
			With(ecs.ComponentTypeSprite).
			Without(ecs.ComponentTypeDisabled).
			Limit(100)

		str := qb.ToString()

		assert.NotEmpty(t, str)
		assert.Contains(t, str, "required")
		assert.Contains(t, str, "excluded")
		assert.Contains(t, str, "limit")
	})

	t.Run("ToHash generates consistent hash", func(t *testing.T) {
		qb1 := NewQueryBuilder().
			With(ecs.ComponentTypeTransform).
			With(ecs.ComponentTypeSprite)

		qb2 := NewQueryBuilder().
			With(ecs.ComponentTypeTransform).
			With(ecs.ComponentTypeSprite)

		hash1 := qb1.ToHash()
		hash2 := qb2.ToHash()

		assert.Equal(t, hash1, hash2)
	})

	t.Run("GetSignature generates unique signature", func(t *testing.T) {
		qb := NewQueryBuilder().
			With(ecs.ComponentTypeTransform).
			Without(ecs.ComponentTypeDisabled)

		sig := qb.GetSignature()

		assert.NotEmpty(t, sig)
	})

	t.Run("Clone creates independent copy", func(t *testing.T) {
		original := NewQueryBuilder().
			With(ecs.ComponentTypeTransform).
			Limit(50)

		clone := original.Clone()

		// Modify original
		original.With(ecs.ComponentTypeSprite)

		// Clone should not be affected
		origImpl := original.(*QueryBuilderImpl)
		cloneImpl := clone.(*QueryBuilderImpl)

		assert.True(t, origImpl.requiredComponents.Has(ecs.ComponentTypeSprite))
		assert.False(t, cloneImpl.requiredComponents.Has(ecs.ComponentTypeSprite))
	})
}

// TestQueryBuilderImpl_ComplexQueries tests complex query scenarios
func TestQueryBuilderImpl_ComplexQueries(t *testing.T) {
	t.Run("Combines multiple constraints", func(t *testing.T) {
		qb := NewQueryBuilder().
			With(ecs.ComponentTypeTransform).
			With(ecs.ComponentTypeSprite).
			Without(ecs.ComponentTypeDisabled).
			WithAny([]ecs.ComponentType{
				ecs.ComponentTypeHealth,
				ecs.ComponentTypeEnergy,
			}).
			WithinRadius(ecs.Vector2{X: 100, Y: 100}, 50).
			Limit(20).
			Cache("complex-query")

		impl := qb.(*QueryBuilderImpl)

		// Verify all constraints are set
		assert.True(t, impl.requiredComponents.Has(ecs.ComponentTypeTransform))
		assert.True(t, impl.requiredComponents.Has(ecs.ComponentTypeSprite))
		assert.True(t, impl.excludedComponents.Has(ecs.ComponentTypeDisabled))
		assert.True(t, impl.optionalComponents.Has(ecs.ComponentTypeHealth))
		assert.Equal(t, 50.0, impl.spatialRadius)
		assert.Equal(t, 20, impl.limit)
		assert.Equal(t, "complex-query", impl.cacheKey)
	})

	t.Run("Fluent interface maintains builder", func(t *testing.T) {
		qb := NewQueryBuilder()

		result := qb.
			With(ecs.ComponentTypeTransform).
			Without(ecs.ComponentTypeDisabled).
			Limit(10).
			Offset(5).
			Cache("fluent-test")

		assert.NotNil(t, result)
		assert.IsType(t, &QueryBuilderImpl{}, result)
	})
}

// TestQueryBuilderImpl_EdgeCases tests edge cases and error conditions
func TestQueryBuilderImpl_EdgeCases(t *testing.T) {
	t.Run("Empty query builder", func(t *testing.T) {
		qb := NewQueryBuilder()

		impl := qb.(*QueryBuilderImpl)
		assert.Equal(t, 0, impl.requiredComponents.Count())
		assert.Equal(t, 0, impl.excludedComponents.Count())
		assert.Equal(t, -1, impl.limit) // No limit
		assert.Equal(t, 0, impl.offset)
	})

	t.Run("Negative limit is ignored", func(t *testing.T) {
		qb := NewQueryBuilder().Limit(-10)

		impl := qb.(*QueryBuilderImpl)
		assert.Equal(t, -1, impl.limit) // Should remain as no limit
	})

	t.Run("Negative offset is ignored", func(t *testing.T) {
		qb := NewQueryBuilder().Offset(-5)

		impl := qb.(*QueryBuilderImpl)
		assert.Equal(t, 0, impl.offset) // Should remain as 0
	})

	t.Run("Empty component arrays", func(t *testing.T) {
		qb := NewQueryBuilder().
			WithAll([]ecs.ComponentType{}).
			WithAny([]ecs.ComponentType{}).
			WithNone([]ecs.ComponentType{})

		impl := qb.(*QueryBuilderImpl)
		assert.Equal(t, 0, impl.requiredComponents.Count())
		assert.Equal(t, 0, impl.optionalComponents.Count())
		assert.Equal(t, 0, impl.excludedComponents.Count())
	})

	t.Run("Duplicate components are handled", func(t *testing.T) {
		qb := NewQueryBuilder().
			With(ecs.ComponentTypeTransform).
			With(ecs.ComponentTypeTransform) // Duplicate

		impl := qb.(*QueryBuilderImpl)
		// Should only be set once
		assert.Equal(t, 1, impl.requiredComponents.Count())
	})
}

// TestQueryBuilderImpl_ThreadSafety tests thread safety considerations
func TestQueryBuilderImpl_ThreadSafety(t *testing.T) {
	t.Run("Clone creates independent instances", func(t *testing.T) {
		original := NewQueryBuilder().With(ecs.ComponentTypeTransform)

		// Create multiple clones
		clones := make([]ecs.QueryBuilder, 10)
		for i := range clones {
			clones[i] = original.Clone()
		}

		// Modify each clone differently
		for i, clone := range clones {
			if i%2 == 0 {
				clone.With(ecs.ComponentTypeSprite)
			} else {
				clone.Without(ecs.ComponentTypeDisabled)
			}
		}

		// Original should remain unchanged
		origImpl := original.(*QueryBuilderImpl)
		assert.Equal(t, 1, origImpl.requiredComponents.Count())
		assert.Equal(t, 0, origImpl.excludedComponents.Count())
	})
}
