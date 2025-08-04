package components

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"muscle-dreamer/internal/core/ecs"
)

func Test_Components_Serialization(t *testing.T) {
	components := []ecs.Component{
		NewTransformComponent(),
		NewSpriteComponent(),
		NewPhysicsComponent(),
		NewHealthComponent(100),
		NewAIComponent(),
	}

	for _, component := range components {
		t.Run(string(component.GetType()), func(t *testing.T) {
			// Act
			data, err := component.Serialize()
			assert.NoError(t, err)
			assert.NotEmpty(t, data)

			// Create new component and deserialize
			newComponent := createComponentByType(component.GetType())
			err = newComponent.Deserialize(data)
			assert.NoError(t, err)

			// Verify data integrity
			newData, err := newComponent.Serialize()
			assert.NoError(t, err)
			assert.Equal(t, data, newData)
		})
	}
}

func Test_Components_MemoryUsage(t *testing.T) {
	components := map[ecs.ComponentType]ecs.Component{
		ecs.ComponentTypeTransform: NewTransformComponent(),
		ecs.ComponentTypeSprite:    NewSpriteComponent(),
		ecs.ComponentTypePhysics:   NewPhysicsComponent(),
		ecs.ComponentTypeHealth:    NewHealthComponent(100),
		ecs.ComponentTypeAI:        NewAIComponent(),
	}

	expectedSizes := map[ecs.ComponentType]int{
		ecs.ComponentTypeTransform: 40,
		ecs.ComponentTypeSprite:    64,
		ecs.ComponentTypePhysics:   56,
		ecs.ComponentTypeHealth:    72,
		ecs.ComponentTypeAI:        96,
	}

	for componentType, component := range components {
		t.Run(string(componentType), func(t *testing.T) {
			actualSize := component.Size()
			expectedSize := expectedSizes[componentType]
			assert.LessOrEqual(t, actualSize, expectedSize,
				"Component %s size %d exceeds limit %d", componentType, actualSize, expectedSize)
		})
	}
}

func Test_Components_InterfaceCompliance(t *testing.T) {
	components := []ecs.Component{
		NewTransformComponent(),
		NewSpriteComponent(),
		NewPhysicsComponent(),
		NewHealthComponent(100),
		NewAIComponent(),
	}

	for _, component := range components {
		t.Run(string(component.GetType()), func(t *testing.T) {
			// Test GetType
			assert.NotEmpty(t, component.GetType())

			// Test Clone
			cloned := component.Clone()
			assert.NotSame(t, component, cloned)
			assert.Equal(t, component.GetType(), cloned.GetType())

			// Test Validate
			err := component.Validate()
			assert.NoError(t, err)

			// Test Size
			size := component.Size()
			assert.Greater(t, size, 0)

			// Test Serialize/Deserialize
			data, err := component.Serialize()
			assert.NoError(t, err)
			assert.NotEmpty(t, data)

			err = component.Deserialize(data)
			assert.NoError(t, err)
		})
	}
}

func Test_Components_TypeConstants(t *testing.T) {
	expectedTypes := []ecs.ComponentType{
		ecs.ComponentTypeTransform,
		ecs.ComponentTypeSprite,
		ecs.ComponentTypePhysics,
		ecs.ComponentTypeHealth,
		ecs.ComponentTypeAI,
	}

	components := []ecs.Component{
		NewTransformComponent(),
		NewSpriteComponent(),
		NewPhysicsComponent(),
		NewHealthComponent(100),
		NewAIComponent(),
	}

	for i, component := range components {
		assert.Equal(t, expectedTypes[i], component.GetType())
	}
}

func Benchmark_Components_Creation(b *testing.B) {
	benchmarks := []struct {
		name    string
		factory func() ecs.Component
	}{
		{"Transform", func() ecs.Component { return NewTransformComponent() }},
		{"Sprite", func() ecs.Component { return NewSpriteComponent() }},
		{"Physics", func() ecs.Component { return NewPhysicsComponent() }},
		{"Health", func() ecs.Component { return NewHealthComponent(100) }},
		{"AI", func() ecs.Component { return NewAIComponent() }},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = bm.factory()
			}
		})
	}
}

func Benchmark_Components_Serialization(b *testing.B) {
	components := []ecs.Component{
		NewTransformComponent(),
		NewSpriteComponent(),
		NewPhysicsComponent(),
		NewHealthComponent(100),
		NewAIComponent(),
	}

	for _, component := range components {
		b.Run(string(component.GetType()), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				data, _ := component.Serialize()
				_ = data
			}
		})
	}
}

func Benchmark_Components_Clone(b *testing.B) {
	components := []ecs.Component{
		NewTransformComponent(),
		NewSpriteComponent(),
		NewPhysicsComponent(),
		NewHealthComponent(100),
		NewAIComponent(),
	}

	for _, component := range components {
		b.Run(string(component.GetType()), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = component.Clone()
			}
		})
	}
}

// Helper function to create components by type for testing
func createComponentByType(componentType ecs.ComponentType) ecs.Component {
	switch componentType {
	case ecs.ComponentTypeTransform:
		return NewTransformComponent()
	case ecs.ComponentTypeSprite:
		return NewSpriteComponent()
	case ecs.ComponentTypePhysics:
		return NewPhysicsComponent()
	case ecs.ComponentTypeHealth:
		return NewHealthComponent(100)
	case ecs.ComponentTypeAI:
		return NewAIComponent()
	default:
		panic("Unknown component type: " + string(componentType))
	}
}
