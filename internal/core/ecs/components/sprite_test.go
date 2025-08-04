package components

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"muscle-dreamer/internal/core/ecs"
)

func Test_SpriteComponent_CreateAndInitialize(t *testing.T) {
	// Arrange & Act
	sprite := NewSpriteComponent()

	// Assert
	assert.Equal(t, ecs.ComponentTypeSprite, sprite.GetType())
	assert.Empty(t, sprite.TextureID)
	assert.Equal(t, ecs.AABB{}, sprite.SourceRect)
	assert.Equal(t, ecs.Color{R: 255, G: 255, B: 255, A: 255}, sprite.Color)
	assert.Equal(t, 0, sprite.ZOrder)
	assert.True(t, sprite.Visible)
	assert.False(t, sprite.FlipX)
	assert.False(t, sprite.FlipY)
}

func Test_SpriteComponent_SetTexture(t *testing.T) {
	// Arrange
	sprite := NewSpriteComponent()
	textureID := "player_texture"
	sourceRect := ecs.AABB{Min: ecs.Vector2{X: 0, Y: 0}, Max: ecs.Vector2{X: 32, Y: 32}}

	// Act
	sprite.SetTexture(textureID, sourceRect)

	// Assert
	assert.Equal(t, textureID, sprite.TextureID)
	assert.Equal(t, sourceRect, sprite.SourceRect)
}

func Test_SpriteComponent_ZOrderSorting(t *testing.T) {
	// Arrange
	sprites := []*SpriteComponent{
		NewSpriteComponent(),
		NewSpriteComponent(),
		NewSpriteComponent(),
	}
	sprites[0].ZOrder = 10
	sprites[1].ZOrder = 5
	sprites[2].ZOrder = 15

	// Act
	sort.Slice(sprites, func(i, j int) bool {
		return sprites[i].ZOrder < sprites[j].ZOrder
	})

	// Assert
	assert.Equal(t, 5, sprites[0].ZOrder)
	assert.Equal(t, 10, sprites[1].ZOrder)
	assert.Equal(t, 15, sprites[2].ZOrder)
}

func Test_SpriteComponent_SetColor(t *testing.T) {
	// Arrange
	sprite := NewSpriteComponent()
	color := ecs.Color{R: 128, G: 64, B: 192, A: 200}

	// Act
	sprite.SetColor(color)

	// Assert
	assert.Equal(t, color, sprite.Color)
}

func Test_SpriteComponent_SetVisibility(t *testing.T) {
	// Arrange
	sprite := NewSpriteComponent()
	assert.True(t, sprite.Visible)

	// Act
	sprite.SetVisible(false)

	// Assert
	assert.False(t, sprite.Visible)
}

func Test_SpriteComponent_Flip(t *testing.T) {
	// Arrange
	sprite := NewSpriteComponent()

	// Act
	sprite.SetFlipX(true)
	sprite.SetFlipY(true)

	// Assert
	assert.True(t, sprite.FlipX)
	assert.True(t, sprite.FlipY)
}

func Test_SpriteComponent_Serialization(t *testing.T) {
	// Arrange
	sprite := NewSpriteComponent()
	sprite.TextureID = "test_texture"
	sprite.SourceRect = ecs.AABB{Min: ecs.Vector2{X: 10, Y: 20}, Max: ecs.Vector2{X: 50, Y: 60}}
	sprite.Color = ecs.Color{R: 100, G: 150, B: 200, A: 250}
	sprite.ZOrder = 5
	sprite.Visible = false
	sprite.FlipX = true

	// Act
	data, err := sprite.Serialize()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Create new component and deserialize
	newSprite := NewSpriteComponent()
	err = newSprite.Deserialize(data)
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, sprite.TextureID, newSprite.TextureID)
	assert.Equal(t, sprite.SourceRect, newSprite.SourceRect)
	assert.Equal(t, sprite.Color, newSprite.Color)
	assert.Equal(t, sprite.ZOrder, newSprite.ZOrder)
	assert.Equal(t, sprite.Visible, newSprite.Visible)
	assert.Equal(t, sprite.FlipX, newSprite.FlipX)
}

func Test_SpriteComponent_Clone(t *testing.T) {
	// Arrange
	original := NewSpriteComponent()
	original.TextureID = "original_texture"
	original.ZOrder = 10

	// Act
	cloned := original.Clone()

	// Assert
	assert.NotSame(t, original, cloned)
	clonedSprite := cloned.(*SpriteComponent)
	assert.Equal(t, original.TextureID, clonedSprite.TextureID)
	assert.Equal(t, original.ZOrder, clonedSprite.ZOrder)
}

func Test_SpriteComponent_Validate(t *testing.T) {
	// Arrange
	sprite := NewSpriteComponent()

	// Act & Assert - valid state
	err := sprite.Validate()
	assert.NoError(t, err)

	// Invalid source rect (negative size)
	sprite.SourceRect = ecs.AABB{
		Min: ecs.Vector2{X: 10, Y: 10},
		Max: ecs.Vector2{X: 5, Y: 5},
	}
	err = sprite.Validate()
	assert.Error(t, err)
}

func Test_SpriteComponent_Size(t *testing.T) {
	// Arrange
	sprite := NewSpriteComponent()

	// Act
	size := sprite.Size()

	// Assert
	assert.LessOrEqual(t, size, 64, "SpriteComponent size should be <= 64 bytes")
	assert.Greater(t, size, 0, "SpriteComponent size should be > 0")
}

func Benchmark_SpriteComponent_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewSpriteComponent()
	}
}

func Benchmark_SpriteComponent_Serialization(b *testing.B) {
	sprite := NewSpriteComponent()
	sprite.TextureID = "benchmark_texture"
	sprite.ZOrder = 5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := sprite.Serialize()
		_ = data
	}
}
