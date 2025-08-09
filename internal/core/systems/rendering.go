package systems

import (
	"sort"

	"muscle-dreamer/internal/core/ecs"
	"muscle-dreamer/internal/core/ecs/components"
)

// RenderingSystem handles entity rendering and drawing operations.
// It processes entities with both TransformComponent and SpriteComponent
// to render sprites on screen with proper positioning and layering.
type RenderingSystem struct {
	*BaseSystem

	// Rendering parameters
	viewport *Rectangle
	camera   *Camera
}

// Camera represents the rendering camera/viewport.
type Camera struct {
	Position ecs.Vector2
	Zoom     float64
	Rotation float64
}

// RenderableEntity holds data for rendering an entity.
type RenderableEntity struct {
	EntityID  ecs.EntityID
	Transform *components.TransformComponent
	Sprite    *components.SpriteComponent
	ZOrder    int
}

// NewRenderingSystem creates a new rendering system.
func NewRenderingSystem() *RenderingSystem {
	return &RenderingSystem{
		BaseSystem: NewBaseSystem(RenderingSystemType, RenderingSystemPriority),
		camera: &Camera{
			Position: ecs.Vector2{X: 0, Y: 0},
			Zoom:     1.0,
			Rotation: 0.0,
		},
	}
}

// GetRequiredComponents returns the components this system operates on.
func (rs *RenderingSystem) GetRequiredComponents() []ecs.ComponentType {
	return []ecs.ComponentType{
		ecs.ComponentTypeTransform,
		ecs.ComponentTypeSprite,
	}
}

// Initialize sets up the rendering system.
func (rs *RenderingSystem) Initialize(world ecs.World) error {
	// TODO: Implement initialization
	return rs.BaseSystem.Initialize(world)
}

// Render draws all renderable entities to the screen.
func (rs *RenderingSystem) Render(world ecs.World, renderer interface{}) error {
	if !rs.IsEnabled() {
		return nil
	}

	// Get entities with both Transform and Sprite components
	result := world.Query().
		With(ecs.ComponentTypeTransform).
		With(ecs.ComponentTypeSprite).
		Execute()

	entities := result.GetEntities()
	renderables := make([]RenderableEntity, 0, len(entities))

	// Collect renderable entities
	for _, entity := range entities {
		transformComp, err := world.GetComponent(entity, ecs.ComponentTypeTransform)
		if err != nil {
			continue
		}
		spriteComp, err := world.GetComponent(entity, ecs.ComponentTypeSprite)
		if err != nil {
			continue
		}

		transform, ok := transformComp.(*components.TransformComponent)
		if !ok {
			continue
		}
		sprite, ok := spriteComp.(*components.SpriteComponent)
		if !ok {
			continue
		}

		// Skip if not visible
		if !sprite.Visible {
			continue
		}

		// Viewport culling
		if !rs.isInViewport(transform, sprite) {
			continue
		}

		renderable := RenderableEntity{
			EntityID:  entity,
			Transform: transform,
			Sprite:    sprite,
			ZOrder:    sprite.ZOrder,
		}
		renderables = append(renderables, renderable)
	}

	// Sort by Z-Order
	rs.sortByZOrder(renderables)

	// For testing, check if renderer is a MockRenderer type
	if mockRenderer, ok := renderer.(interface {
		DrawSprite(textureID string, position, scale ecs.Vector2, rotation float64, zOrder int)
	}); ok {
		// Render each entity using the mock interface
		for _, renderable := range renderables {
			mockRenderer.DrawSprite(
				renderable.Sprite.TextureID,
				renderable.Transform.Position,
				renderable.Transform.Scale,
				renderable.Transform.Rotation,
				renderable.Sprite.ZOrder,
			)
		}
	}

	return rs.BaseSystem.Render(world, renderer)
}

// SetViewport sets the rendering viewport dimensions.
func (rs *RenderingSystem) SetViewport(x, y, width, height float64) {
	rs.viewport = &Rectangle{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

// GetViewport returns the current rendering viewport.
func (rs *RenderingSystem) GetViewport() *Rectangle {
	return rs.viewport
}

// SetCamera sets the camera position and properties.
func (rs *RenderingSystem) SetCamera(position ecs.Vector2, zoom, rotation float64) {
	rs.camera.Position = position
	rs.camera.Zoom = zoom
	rs.camera.Rotation = rotation
}

// GetCamera returns the current camera settings.
func (rs *RenderingSystem) GetCamera() *Camera {
	return rs.camera
}

// isInViewport checks if an entity is within the viewport bounds.
func (rs *RenderingSystem) isInViewport(transform *components.TransformComponent, sprite *components.SpriteComponent) bool {
	if rs.viewport == nil {
		return true // No culling if no viewport is set
	}

	// Simple AABB check against viewport
	spriteWidth := sprite.SourceRect.Max.X - sprite.SourceRect.Min.X
	spriteHeight := sprite.SourceRect.Max.Y - sprite.SourceRect.Min.Y

	entityLeft := transform.Position.X
	entityRight := transform.Position.X + spriteWidth
	entityTop := transform.Position.Y
	entityBottom := transform.Position.Y + spriteHeight

	viewportLeft := rs.viewport.X
	viewportRight := rs.viewport.X + rs.viewport.Width
	viewportTop := rs.viewport.Y
	viewportBottom := rs.viewport.Y + rs.viewport.Height

	return !(entityRight < viewportLeft ||
		entityLeft > viewportRight ||
		entityBottom < viewportTop ||
		entityTop > viewportBottom)
}

// sortByZOrder sorts renderable entities by their Z-order for proper layering.
func (rs *RenderingSystem) sortByZOrder(entities []RenderableEntity) {
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ZOrder < entities[j].ZOrder
	})
}
