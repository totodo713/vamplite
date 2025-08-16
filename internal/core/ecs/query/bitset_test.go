package query

import (
	"testing"

	"muscle-dreamer/internal/core/ecs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponentBitSet_BasicOperations(t *testing.T) {
	t.Run("A-001-01: 新しいビットセットは初期状態で全ビットが0", func(t *testing.T) {
		bitset := NewComponentBitSet()

		// 基本コンポーネントタイプでテスト
		componentTypes := []ecs.ComponentType{
			ecs.ComponentTypeTransform,
			ecs.ComponentTypeSprite,
			ecs.ComponentTypePhysics,
			ecs.ComponentTypeHealth,
		}

		for _, componentType := range componentTypes {
			assert.False(t, bitset.Has(componentType),
				"初期状態では%sビットは0である必要があります", componentType)
		}
	})

	t.Run("A-001-02: Set操作で指定ビットが1になる", func(t *testing.T) {
		bitset := NewComponentBitSet()

		bitset = bitset.Set(ecs.ComponentTypeTransform)
		assert.True(t, bitset.Has(ecs.ComponentTypeTransform),
			"Set操作後はTransformビットが1になる必要があります")
		assert.False(t, bitset.Has(ecs.ComponentTypeSprite),
			"他のビットは変更されない必要があります")
	})

	t.Run("A-001-03: Clear操作で指定ビットが0になる", func(t *testing.T) {
		bitset := NewComponentBitSet()
		bitset = bitset.Set(ecs.ComponentTypeTransform)
		bitset = bitset.Set(ecs.ComponentTypeSprite)

		bitset = bitset.Clear(ecs.ComponentTypeTransform)
		assert.False(t, bitset.Has(ecs.ComponentTypeTransform),
			"Clear操作後はTransformビットが0になる必要があります")
		assert.True(t, bitset.Has(ecs.ComponentTypeSprite),
			"他のビットは変更されない必要があります")
	})

	t.Run("A-001-04: Has操作で正しいビット状態を返す", func(t *testing.T) {
		bitset := NewComponentBitSet()

		// 複数のビットをセット
		bitset = bitset.Set(ecs.ComponentTypeTransform)
		bitset = bitset.Set(ecs.ComponentTypePhysics)

		assert.True(t, bitset.Has(ecs.ComponentTypeTransform))
		assert.False(t, bitset.Has(ecs.ComponentTypeSprite))
		assert.True(t, bitset.Has(ecs.ComponentTypePhysics))
		assert.False(t, bitset.Has(ecs.ComponentTypeHealth))
	})

	t.Run("A-001-05: 同じビット位置に複数回Set/Clearしても正常動作", func(t *testing.T) {
		bitset := NewComponentBitSet()

		// 複数回Set
		bitset = bitset.Set(ecs.ComponentTypeTransform)
		bitset = bitset.Set(ecs.ComponentTypeTransform)
		bitset = bitset.Set(ecs.ComponentTypeTransform)
		assert.True(t, bitset.Has(ecs.ComponentTypeTransform))

		// 複数回Clear
		bitset = bitset.Clear(ecs.ComponentTypeTransform)
		bitset = bitset.Clear(ecs.ComponentTypeTransform)
		assert.False(t, bitset.Has(ecs.ComponentTypeTransform))
	})
}

func TestComponentBitSet_LogicalOperations(t *testing.T) {
	t.Run("A-002-01: AND演算で共通ビットのみ1になる", func(t *testing.T) {
		// Transform + Sprite (0b0011)
		bitsetA := NewComponentBitSet().
			Set(ecs.ComponentTypeTransform).
			Set(ecs.ComponentTypeSprite)

		// Physics + Health (0b1100)
		bitsetB := NewComponentBitSet().
			Set(ecs.ComponentTypePhysics).
			Set(ecs.ComponentTypeHealth)

		result := bitsetA.And(bitsetB)

		// AND結果は0b0000になるはず
		assert.False(t, result.Has(ecs.ComponentTypeTransform))
		assert.False(t, result.Has(ecs.ComponentTypeSprite))
		assert.False(t, result.Has(ecs.ComponentTypePhysics))
		assert.False(t, result.Has(ecs.ComponentTypeHealth))
	})

	t.Run("A-002-02: OR演算でいずれかが1なら1になる", func(t *testing.T) {
		// Transform + Sprite
		bitsetA := NewComponentBitSet().
			Set(ecs.ComponentTypeTransform).
			Set(ecs.ComponentTypeSprite)

		// Physics + Health
		bitsetB := NewComponentBitSet().
			Set(ecs.ComponentTypePhysics).
			Set(ecs.ComponentTypeHealth)

		result := bitsetA.Or(bitsetB)

		// OR結果は全ビット1になるはず
		assert.True(t, result.Has(ecs.ComponentTypeTransform))
		assert.True(t, result.Has(ecs.ComponentTypeSprite))
		assert.True(t, result.Has(ecs.ComponentTypePhysics))
		assert.True(t, result.Has(ecs.ComponentTypeHealth))
	})

	t.Run("A-002-03: Transform+Physics AND Transform+Sprite", func(t *testing.T) {
		// Transform + Physics (0b0101)
		bitsetA := NewComponentBitSet().
			Set(ecs.ComponentTypeTransform).
			Set(ecs.ComponentTypePhysics)

		// Transform + Sprite (0b0011)
		bitsetB := NewComponentBitSet().
			Set(ecs.ComponentTypeTransform).
			Set(ecs.ComponentTypeSprite)

		result := bitsetA.And(bitsetB)

		// AND結果はTransformのみ (0b0001)
		assert.True(t, result.Has(ecs.ComponentTypeTransform))
		assert.False(t, result.Has(ecs.ComponentTypeSprite))
		assert.False(t, result.Has(ecs.ComponentTypePhysics))
		assert.False(t, result.Has(ecs.ComponentTypeHealth))
	})
}

func TestComponentBitSet_BoundaryValues(t *testing.T) {
	t.Run("A-003-01: ComponentType to bit position mapping", func(t *testing.T) {
		// 各ComponentTypeが一意のビット位置にマップされることを確認
		positions := make(map[int]ecs.ComponentType)
		componentTypes := []ecs.ComponentType{
			ecs.ComponentTypeTransform,
			ecs.ComponentTypeSprite,
			ecs.ComponentTypePhysics,
			ecs.ComponentTypeHealth,
			ecs.ComponentTypeAI,
			ecs.ComponentTypeInventory,
			ecs.ComponentTypeAudio,
			ecs.ComponentTypeInput,
		}

		for _, componentType := range componentTypes {
			position := getComponentBitPosition(componentType)
			require.True(t, position >= 0 && position < 64,
				"ビット位置は0-63の範囲内である必要があります: %s -> %d", componentType, position)

			if prevType, exists := positions[position]; exists {
				t.Errorf("重複するビット位置: %s と %s が同じ位置 %d", componentType, prevType, position)
			}
			positions[position] = componentType
		}
	})

	t.Run("A-003-03: 存在しないComponentTypeの処理", func(t *testing.T) {
		bitset := NewComponentBitSet()
		invalidComponentType := ecs.ComponentTypeFromString("invalid_component_type")

		// 存在しないコンポーネントタイプは-1を返すかエラーになるべき
		position := getComponentBitPosition(invalidComponentType)
		assert.Equal(t, -1, position, "存在しないComponentTypeは-1を返すべき")

		// Set/Has操作は安全に処理される必要がある
		assert.NotPanics(t, func() {
			bitset.Set(invalidComponentType)
			bitset.Has(invalidComponentType)
		})
	})
}

func TestComponentBitSet_ExtendedOperations(t *testing.T) {
	t.Run("NewComponentBitSetWithComponents", func(t *testing.T) {
		bitset := NewComponentBitSetWithComponents(
			ecs.ComponentTypeTransform,
			ecs.ComponentTypeSprite,
			ecs.ComponentTypePhysics,
		)

		assert.True(t, bitset.Has(ecs.ComponentTypeTransform))
		assert.True(t, bitset.Has(ecs.ComponentTypeSprite))
		assert.True(t, bitset.Has(ecs.ComponentTypePhysics))
		assert.False(t, bitset.Has(ecs.ComponentTypeHealth))
	})

	t.Run("HasAll複数コンポーネント", func(t *testing.T) {
		bitset := NewComponentBitSetWithComponents(
			ecs.ComponentTypeTransform,
			ecs.ComponentTypeSprite,
			ecs.ComponentTypePhysics,
		)

		assert.True(t, bitset.HasAll(ecs.ComponentTypeTransform, ecs.ComponentTypeSprite))
		assert.True(t, bitset.HasAll(ecs.ComponentTypeTransform))
		assert.False(t, bitset.HasAll(ecs.ComponentTypeTransform, ecs.ComponentTypeHealth))
	})

	t.Run("HasAny複数コンポーネント", func(t *testing.T) {
		bitset := NewComponentBitSetWithComponents(ecs.ComponentTypeTransform)

		assert.True(t, bitset.HasAny(ecs.ComponentTypeTransform, ecs.ComponentTypeSprite))
		assert.True(t, bitset.HasAny(ecs.ComponentTypeTransform))
		assert.False(t, bitset.HasAny(ecs.ComponentTypeSprite, ecs.ComponentTypeHealth))
	})

	t.Run("集合演算の正確性", func(t *testing.T) {
		bitsetA := NewComponentBitSetWithComponents(ecs.ComponentTypeTransform, ecs.ComponentTypeSprite)
		bitsetB := NewComponentBitSetWithComponents(ecs.ComponentTypeTransform, ecs.ComponentTypePhysics)

		// Intersection
		intersection := bitsetA.And(bitsetB)
		assert.True(t, intersection.Has(ecs.ComponentTypeTransform))
		assert.False(t, intersection.Has(ecs.ComponentTypeSprite))
		assert.False(t, intersection.Has(ecs.ComponentTypePhysics))

		// Union
		union := bitsetA.Or(bitsetB)
		assert.True(t, union.Has(ecs.ComponentTypeTransform))
		assert.True(t, union.Has(ecs.ComponentTypeSprite))
		assert.True(t, union.Has(ecs.ComponentTypePhysics))

		// Subset/Superset
		assert.True(t, intersection.IsSubsetOf(bitsetA))
		assert.True(t, bitsetA.IsSupersetOf(intersection))
		assert.True(t, intersection.IsSubsetOf(union))

		// Intersects
		assert.True(t, bitsetA.Intersects(bitsetB))

		// Equals
		assert.True(t, bitsetA.Equals(bitsetA))
		assert.False(t, bitsetA.Equals(bitsetB))
	})

	t.Run("SetMany and ClearMany", func(t *testing.T) {
		bitset := NewComponentBitSet()

		// SetManyのテスト
		bitset = bitset.SetMany(
			ecs.ComponentTypeTransform,
			ecs.ComponentTypeSprite,
			ecs.ComponentTypePhysics,
		)

		assert.True(t, bitset.HasAll(
			ecs.ComponentTypeTransform,
			ecs.ComponentTypeSprite,
			ecs.ComponentTypePhysics,
		))

		// ClearManyのテスト
		bitset = bitset.ClearMany(ecs.ComponentTypeSprite, ecs.ComponentTypePhysics)
		assert.True(t, bitset.Has(ecs.ComponentTypeTransform))
		assert.False(t, bitset.Has(ecs.ComponentTypeSprite))
		assert.False(t, bitset.Has(ecs.ComponentTypePhysics))
	})

	t.Run("GetSetComponentTypes", func(t *testing.T) {
		bitset := NewComponentBitSetWithComponents(
			ecs.ComponentTypeTransform,
			ecs.ComponentTypeSprite,
		)

		setTypes := bitset.GetSetComponentTypes()
		assert.Len(t, setTypes, 2)

		// setTypesにTransformとSpriteが含まれていることを確認
		found := make(map[ecs.ComponentType]bool)
		for _, componentType := range setTypes {
			found[componentType] = true
		}

		assert.True(t, found[ecs.ComponentTypeTransform])
		assert.True(t, found[ecs.ComponentTypeSprite])
	})

	t.Run("ForEachSet", func(t *testing.T) {
		bitset := NewComponentBitSetWithComponents(
			ecs.ComponentTypeTransform,
			ecs.ComponentTypeSprite,
		)

		var callCount int
		var calledTypes []ecs.ComponentType

		bitset.ForEachSet(func(componentType ecs.ComponentType) {
			callCount++
			calledTypes = append(calledTypes, componentType)
		})

		assert.Equal(t, 2, callCount)
		assert.Len(t, calledTypes, 2)
	})
}
