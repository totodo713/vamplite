package query

import (
	"muscle-dreamer/internal/core/ecs"
)

// ComponentBitSet represents component presence using bitset operations
// 最大64種類のコンポーネントをサポート
type ComponentBitSet uint64

// componentTypeToBitPosition maps component types to bit positions
var componentTypeToBitPosition = map[ecs.ComponentType]int{
	ecs.ComponentTypeTransform: 0,
	ecs.ComponentTypeSprite:    1,
	ecs.ComponentTypePhysics:   2,
	ecs.ComponentTypeHealth:    3,
	ecs.ComponentTypeAI:        4,
	ecs.ComponentTypeInventory: 5,
	ecs.ComponentTypeAudio:     6,
	ecs.ComponentTypeInput:     7,
	ecs.ComponentTypeDisabled:  8,
	ecs.ComponentTypeEnergy:    9,
	ecs.ComponentTypeDead:      10,
}

// NewComponentBitSet creates a new empty bitset
func NewComponentBitSet() ComponentBitSet {
	return ComponentBitSet(0)
}

// NewComponentBitSetWithComponents creates a bitset with specified components
func NewComponentBitSetWithComponents(componentTypes ...ecs.ComponentType) ComponentBitSet {
	bitset := NewComponentBitSet()
	for _, componentType := range componentTypes {
		bitset = bitset.Set(componentType)
	}
	return bitset
}

// Set sets the bit for the given component type and returns a new ComponentBitSet.
//
// If the component type is not registered, the operation is ignored and the
// original bitset is returned unchanged. This ensures that invalid component
// types do not cause errors during runtime.
//
// Performance: O(1) - constant time operation using bit manipulation.
func (b ComponentBitSet) Set(componentType ecs.ComponentType) ComponentBitSet {
	position, exists := getComponentBitPositionSafe(componentType)
	if !exists {
		// 無効なコンポーネントタイプは無視
		return b
	}
	return b | (1 << position)
}

// SetMany sets multiple component type bits at once
func (b ComponentBitSet) SetMany(componentTypes ...ecs.ComponentType) ComponentBitSet {
	result := b
	for _, componentType := range componentTypes {
		result = result.Set(componentType)
	}
	return result
}

// Clear clears the bit for the given component type
func (b ComponentBitSet) Clear(componentType ecs.ComponentType) ComponentBitSet {
	position, exists := getComponentBitPositionSafe(componentType)
	if !exists {
		return b
	}
	return b &^ (1 << position)
}

// ClearMany clears multiple component type bits at once
func (b ComponentBitSet) ClearMany(componentTypes ...ecs.ComponentType) ComponentBitSet {
	result := b
	for _, componentType := range componentTypes {
		result = result.Clear(componentType)
	}
	return result
}

// Has checks if the bit for the given component type is set
func (b ComponentBitSet) Has(componentType ecs.ComponentType) bool {
	position, exists := getComponentBitPositionSafe(componentType)
	if !exists {
		// 無効なコンポーネントタイプは常にfalse
		return false
	}
	return (b & (1 << position)) != 0
}

// HasAll checks if all specified component types are set
func (b ComponentBitSet) HasAll(componentTypes ...ecs.ComponentType) bool {
	for _, componentType := range componentTypes {
		if !b.Has(componentType) {
			return false
		}
	}
	return true
}

// HasAny checks if any of the specified component types are set
func (b ComponentBitSet) HasAny(componentTypes ...ecs.ComponentType) bool {
	for _, componentType := range componentTypes {
		if b.Has(componentType) {
			return true
		}
	}
	return false
}

// And performs bitwise AND operation
func (b ComponentBitSet) And(other ComponentBitSet) ComponentBitSet {
	return b & other
}

// Or performs bitwise OR operation
func (b ComponentBitSet) Or(other ComponentBitSet) ComponentBitSet {
	return b | other
}

// Intersects checks if this bitset intersects with another
func (b ComponentBitSet) Intersects(other ComponentBitSet) bool {
	return (b & other) != 0
}

// IsSubsetOf checks if this bitset is a subset of another
func (b ComponentBitSet) IsSubsetOf(other ComponentBitSet) bool {
	return (b & other) == b
}

// IsSupersetOf checks if this bitset is a superset of another
func (b ComponentBitSet) IsSupersetOf(other ComponentBitSet) bool {
	return other.IsSubsetOf(b)
}

// Equals checks if two bitsets are equal
func (b ComponentBitSet) Equals(other ComponentBitSet) bool {
	return b == other
}

// GetSetComponentTypes returns all component types that are set
func (b ComponentBitSet) GetSetComponentTypes() []ecs.ComponentType {
	var result []ecs.ComponentType

	for componentType, position := range componentTypeToBitPosition {
		if (b & (1 << position)) != 0 {
			result = append(result, componentType)
		}
	}

	return result
}

// ForEachSet executes a function for each set component type
func (b ComponentBitSet) ForEachSet(fn func(ecs.ComponentType)) {
	for _, componentType := range b.GetSetComponentTypes() {
		fn(componentType)
	}
}

// getComponentBitPositionSafe returns the bit position with error handling
func getComponentBitPositionSafe(componentType ecs.ComponentType) (int, bool) {
	position, exists := componentTypeToBitPosition[componentType]
	return position, exists
}

// getComponentBitPosition returns the bit position for a component type
// Legacy function for backward compatibility
func getComponentBitPosition(componentType ecs.ComponentType) int {
	if position, exists := getComponentBitPositionSafe(componentType); exists {
		return position
	}
	return -1 // 無効なコンポーネントタイプ
}
