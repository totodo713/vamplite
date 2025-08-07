package query

import (
	"muscle-dreamer/internal/core/ecs"
)

// componentTypeToIndex maps ComponentType to bit position
var componentTypeToIndex = map[ecs.ComponentType]int{
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

// indexToComponentType maps bit position to ComponentType
var indexToComponentType = map[int]ecs.ComponentType{
	0:  ecs.ComponentTypeTransform,
	1:  ecs.ComponentTypeSprite,
	2:  ecs.ComponentTypePhysics,
	3:  ecs.ComponentTypeHealth,
	4:  ecs.ComponentTypeAI,
	5:  ecs.ComponentTypeInventory,
	6:  ecs.ComponentTypeAudio,
	7:  ecs.ComponentTypeInput,
	8:  ecs.ComponentTypeDisabled,
	9:  ecs.ComponentTypeEnergy,
	10: ecs.ComponentTypeDead,
}

// GetComponentTypeFromPosition returns the ComponentType for a bit position
func GetComponentTypeFromPosition(pos int) ecs.ComponentType {
	if ct, ok := indexToComponentType[pos]; ok {
		return ct
	}
	return ecs.InvalidComponentType
}

// RegisterComponentType registers a new component type with a bit position
// This function is thread-safe and can be called during initialization
func RegisterComponentType(ct ecs.ComponentType, position int) {
	if position < 0 || position >= 64 {
		panic("component bit position must be between 0 and 63")
	}

	// Check for conflicts
	if existingPos, exists := componentTypeToIndex[ct]; exists && existingPos != position {
		panic("component type already registered with different position")
	}

	if existingCT, exists := indexToComponentType[position]; exists && existingCT != ct {
		panic("bit position already used by different component type")
	}

	// Register the mapping
	componentTypeToIndex[ct] = position
	indexToComponentType[position] = ct
}
