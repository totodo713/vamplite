package components

import (
	"encoding/json"
	"errors"

	"muscle-dreamer/internal/core/ecs"
)

// SpriteComponent handles 2D sprite rendering information
type SpriteComponent struct {
	TextureID  string    `json:"texture_id"`
	SourceRect ecs.AABB  `json:"source_rect"`
	Color      ecs.Color `json:"color"`
	ZOrder     int       `json:"z_order"`
	Visible    bool      `json:"visible"`
	FlipX      bool      `json:"flip_x"`
	FlipY      bool      `json:"flip_y"`
}

// NewSpriteComponent creates a new sprite component with default values
func NewSpriteComponent() *SpriteComponent {
	return &SpriteComponent{
		TextureID:  "",
		SourceRect: ecs.AABB{},
		Color:      ecs.Color{R: 255, G: 255, B: 255, A: 255},
		ZOrder:     0,
		Visible:    true,
		FlipX:      false,
		FlipY:      false,
	}
}

// GetType returns the component type
func (s *SpriteComponent) GetType() ecs.ComponentType {
	return ecs.ComponentTypeSprite
}

// SetTexture sets the texture ID and source rectangle
func (s *SpriteComponent) SetTexture(textureID string, sourceRect ecs.AABB) {
	s.TextureID = textureID
	s.SourceRect = sourceRect
}

// SetColor sets the sprite color
func (s *SpriteComponent) SetColor(color ecs.Color) {
	s.Color = color
}

// SetVisible sets the visibility
func (s *SpriteComponent) SetVisible(visible bool) {
	s.Visible = visible
}

// SetFlipX sets horizontal flip
func (s *SpriteComponent) SetFlipX(flip bool) {
	s.FlipX = flip
}

// SetFlipY sets vertical flip
func (s *SpriteComponent) SetFlipY(flip bool) {
	s.FlipY = flip
}

// Clone creates a deep copy of the component
func (s *SpriteComponent) Clone() ecs.Component {
	return &SpriteComponent{
		TextureID:  s.TextureID,
		SourceRect: s.SourceRect,
		Color:      s.Color,
		ZOrder:     s.ZOrder,
		Visible:    s.Visible,
		FlipX:      s.FlipX,
		FlipY:      s.FlipY,
	}
}

// Validate ensures the component data is valid
func (s *SpriteComponent) Validate() error {
	// Check if source rect is valid (max >= min)
	if s.SourceRect.Max.X < s.SourceRect.Min.X || s.SourceRect.Max.Y < s.SourceRect.Min.Y {
		return errors.New("invalid source rectangle: max must be >= min")
	}
	return nil
}

// Size returns the memory size of the component in bytes
func (s *SpriteComponent) Size() int {
	// Approximate size calculation
	return 48 + len(s.TextureID) // Base struct + string length
}

// Serialize converts the component to bytes
func (s *SpriteComponent) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

// Deserialize loads component data from bytes
func (s *SpriteComponent) Deserialize(data []byte) error {
	return json.Unmarshal(data, s)
}
