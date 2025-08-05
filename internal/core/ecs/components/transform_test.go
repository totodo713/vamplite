package components

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"muscle-dreamer/internal/core/ecs"
)

func Test_TransformComponent_CreateAndInitialize(t *testing.T) {
	// Arrange & Act
	transform := NewTransformComponent()

	// Assert
	assert.Equal(t, ecs.ComponentTypeTransform, transform.GetType())
	assert.Equal(t, ecs.Vector2{X: 0, Y: 0}, transform.Position)
	assert.Equal(t, 0.0, transform.Rotation)
	assert.Equal(t, ecs.Vector2{X: 1, Y: 1}, transform.Scale)
	assert.Nil(t, transform.Parent)
	assert.Empty(t, transform.Children)
}

func Test_TransformComponent_SetPosition(t *testing.T) {
	// Arrange
	transform := NewTransformComponent()
	newPos := ecs.Vector2{X: 10.5, Y: -20.3}

	// Act
	transform.SetPosition(newPos)

	// Assert
	assert.Equal(t, newPos, transform.Position)
	assert.Equal(t, newPos, transform.GetWorldPosition())
}

func Test_TransformComponent_WorldLocalConversion(t *testing.T) {
	// Arrange
	parent := NewTransformComponent()
	parent.SetPosition(ecs.Vector2{X: 10, Y: 10})
	parent.SetRotation(math.Pi / 4)

	child := NewTransformComponent()
	child.SetPosition(ecs.Vector2{X: 5, Y: 0})
	child.SetParent(parent)

	// Act
	worldPos := child.GetWorldPosition()
	localPos := child.GetLocalPosition()

	// Assert
	assert.Equal(t, ecs.Vector2{X: 5, Y: 0}, localPos)
	assert.NotEqual(t, localPos, worldPos)
	// 回転と移動を考慮した座標
	expectedX := 10 + 5*math.Cos(math.Pi/4)
	expectedY := 10 + 5*math.Sin(math.Pi/4)
	assert.InDelta(t, expectedX, worldPos.X, 0.001)
	assert.InDelta(t, expectedY, worldPos.Y, 0.001)
}

func Test_TransformComponent_HierarchyManagement(t *testing.T) {
	// Arrange
	parent := NewTransformComponent()
	child1 := NewTransformComponent()
	child2 := NewTransformComponent()

	// Act
	child1.SetParent(parent)
	child2.SetParent(parent)

	// Assert
	assert.Equal(t, parent, child1.Parent)
	assert.Equal(t, parent, child2.Parent)
	assert.Len(t, parent.Children, 2)
	assert.Contains(t, parent.Children, child1)
	assert.Contains(t, parent.Children, child2)
}

func Test_TransformComponent_CircularParentReference(t *testing.T) {
	// Arrange
	parent := NewTransformComponent()
	child := NewTransformComponent()
	child.SetParent(parent)

	// Act & Assert
	err := parent.SetParent(child)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular reference")
}

func Test_TransformComponent_Serialization(t *testing.T) {
	// Arrange
	transform := NewTransformComponent()
	transform.SetPosition(ecs.Vector2{X: 10, Y: 20})
	transform.SetRotation(math.Pi / 2)
	transform.SetScale(ecs.Vector2{X: 2, Y: 3})

	// Act
	data, err := transform.Serialize()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Create new component and deserialize
	newTransform := NewTransformComponent()
	err = newTransform.Deserialize(data)
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, transform.Position, newTransform.Position)
	assert.Equal(t, transform.Rotation, newTransform.Rotation)
	assert.Equal(t, transform.Scale, newTransform.Scale)
}

func Test_TransformComponent_Clone(t *testing.T) {
	// Arrange
	original := NewTransformComponent()
	original.SetPosition(ecs.Vector2{X: 15, Y: 25})
	original.SetRotation(math.Pi / 3)

	// Act
	cloned := original.Clone()

	// Assert
	assert.NotSame(t, original, cloned)
	clonedTransform := cloned.(*TransformComponent)
	assert.Equal(t, original.Position, clonedTransform.Position)
	assert.Equal(t, original.Rotation, clonedTransform.Rotation)
	assert.Equal(t, original.Scale, clonedTransform.Scale)
}

func Test_TransformComponent_Validate(t *testing.T) {
	// Arrange
	transform := NewTransformComponent()

	// Act & Assert - valid state
	err := transform.Validate()
	assert.NoError(t, err)

	// Invalid scale (zero)
	transform.Scale = ecs.Vector2{X: 0, Y: 0}
	err = transform.Validate()
	assert.Error(t, err)
}

func Test_TransformComponent_GetTransformMatrix(t *testing.T) {
	// Arrange
	transform := NewTransformComponent()
	transform.SetPosition(ecs.Vector2{X: 10, Y: 20})
	transform.SetRotation(math.Pi / 4)
	transform.SetScale(ecs.Vector2{X: 2, Y: 3})

	// Act
	matrix := transform.GetTransformMatrix()

	// Assert
	assert.NotEqual(t, TransformMatrix{}, matrix)
	// Check that matrix calculation is cached (not dirty)
	matrix2 := transform.GetTransformMatrix()
	assert.Equal(t, matrix, matrix2)
}

func Test_TransformComponent_ParentChildRemoval(t *testing.T) {
	// Arrange
	parent := NewTransformComponent()
	child1 := NewTransformComponent()
	child2 := NewTransformComponent()

	child1.SetParent(parent)
	child2.SetParent(parent)
	assert.Len(t, parent.Children, 2)

	// Act - remove child1 by setting different parent
	newParent := NewTransformComponent()
	child1.SetParent(newParent)

	// Assert
	assert.Len(t, parent.Children, 1)
	assert.Contains(t, parent.Children, child2)
	assert.Equal(t, newParent, child1.Parent)
	assert.Len(t, newParent.Children, 1)
	assert.Contains(t, newParent.Children, child1)

	// Test nil parent (remove from all parents)
	child1.SetParent(nil)
	assert.Len(t, newParent.Children, 0)
	assert.Nil(t, child1.Parent)
}

func Test_TransformComponent_WorldTransforms(t *testing.T) {
	// Arrange
	grandparent := NewTransformComponent()
	grandparent.SetRotation(math.Pi / 2) // 90 degrees
	grandparent.SetScale(ecs.Vector2{X: 2, Y: 2})

	parent := NewTransformComponent()
	parent.SetParent(grandparent)
	parent.SetPosition(ecs.Vector2{X: 10, Y: 0})

	child := NewTransformComponent()
	child.SetParent(parent)

	// Act & Assert - world rotation
	assert.InDelta(t, math.Pi/2, parent.GetWorldRotation(), 0.001)

	// Act & Assert - world scale
	worldScale := parent.GetWorldScale()
	assert.InDelta(t, 2.0, worldScale.X, 0.001)
	assert.InDelta(t, 2.0, worldScale.Y, 0.001)
}

func Test_TransformComponent_Size(t *testing.T) {
	// Arrange
	transform := NewTransformComponent()

	// Act
	size := transform.Size()

	// Assert
	assert.LessOrEqual(t, size, 40, "TransformComponent size should be <= 40 bytes")
	assert.Greater(t, size, 0, "TransformComponent size should be > 0")
}

func Benchmark_TransformComponent_MatrixCalculation(b *testing.B) {
	transform := NewTransformComponent()
	transform.SetPosition(ecs.Vector2{X: 100, Y: 200})
	transform.SetRotation(math.Pi / 3)
	transform.SetScale(ecs.Vector2{X: 2, Y: 3})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = transform.GetTransformMatrix()
	}
}

func Benchmark_TransformComponent_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewTransformComponent()
	}
}
