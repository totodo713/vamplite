package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"muscle-dreamer/internal/core/ecs"
	"muscle-dreamer/internal/core/ecs/components"
	"muscle-dreamer/internal/core/systems"
)

func TestRenderingSystem_Interface(t *testing.T) {
	system := systems.NewRenderingSystem()

	var _ ecs.System = system

	assert.Equal(t, systems.RenderingSystemType, system.GetType())
	assert.Equal(t, systems.RenderingSystemPriority, system.GetPriority())

	required := system.GetRequiredComponents()
	assert.Contains(t, required, ecs.ComponentTypeTransform)
	assert.Contains(t, required, ecs.ComponentTypeSprite)
}

func TestRenderingSystem_BasicRendering(t *testing.T) {
	system := systems.NewRenderingSystem()
	world := createWorldWithEntities()
	mockRenderer := &MockRenderer{}

	// 描画エンティティ作成
	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 100, Y: 200},
		Scale:    ecs.Vector2{X: 2, Y: 2},
		Rotation: 0.785, // 45度回転 (π/4)
	}
	sprite := &components.SpriteComponent{
		TextureID: "player_sprite",
		ZOrder:    0,
		Visible:   true,
		Color:     ecs.Color{R: 255, G: 255, B: 255, A: 255},
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, sprite)

	system.Initialize(world)

	// Render実行
	err := system.Render(world, mockRenderer)
	assert.NoError(t, err)

	// モックレンダラーが呼ばれたことを確認
	assert.Equal(t, 1, mockRenderer.DrawCallCount)
	assert.Equal(t, "player_sprite", mockRenderer.LastTexture)
}

func TestRenderingSystem_ZOrder(t *testing.T) {
	system := systems.NewRenderingSystem()
	world := createWorldWithEntities()
	mockRenderer := &MockRenderer{}

	// 異なるZ-Orderのエンティティ作成
	entities := []ecs.EntityID{}
	zOrders := []int{3, 1, 5, 2} // 描画順序: 1, 2, 3, 5

	for i, z := range zOrders {
		entity := world.CreateEntity()
		transform := &components.TransformComponent{
			Position: ecs.Vector2{X: float64(i * 50), Y: 0},
			Scale:    ecs.Vector2{X: 1, Y: 1},
		}
		sprite := &components.SpriteComponent{
			ZOrder:    z,
			TextureID: fmt.Sprintf("sprite_%d", i),
			Visible:   true,
			Color:     ecs.Color{R: 255, G: 255, B: 255, A: 255},
		}
		world.AddComponent(entity, transform)
		world.AddComponent(entity, sprite)
		entities = append(entities, entity)
	}

	system.Initialize(world)

	err := system.Render(world, mockRenderer)
	assert.NoError(t, err)

	// 描画順序確認（Z-Orderが低いものから順に描画される）
	expectedOrder := []string{"sprite_1", "sprite_3", "sprite_0", "sprite_2"}
	assert.Equal(t, expectedOrder, mockRenderer.DrawOrder)
}

func TestRenderingSystem_ViewportCulling(t *testing.T) {
	system := systems.NewRenderingSystem()
	system.SetViewport(0, 0, 800, 600)
	world := createWorldWithEntities()
	mockRenderer := &MockRenderer{}

	// 画面内エンティティ
	visibleEntity := world.CreateEntity()
	visibleTransform := &components.TransformComponent{
		Position: ecs.Vector2{X: 400, Y: 300},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	visibleSprite := &components.SpriteComponent{
		TextureID: "visible",
		ZOrder:    0,
		Visible:   true,
		Color:     ecs.Color{R: 255, G: 255, B: 255, A: 255},
		SourceRect: ecs.AABB{
			Min: ecs.Vector2{X: 0, Y: 0},
			Max: ecs.Vector2{X: 32, Y: 32}, // 32x32 sprite
		},
	}
	world.AddComponent(visibleEntity, visibleTransform)
	world.AddComponent(visibleEntity, visibleSprite)

	// 画面外エンティティ
	hiddenEntity := world.CreateEntity()
	hiddenTransform := &components.TransformComponent{
		Position: ecs.Vector2{X: -100, Y: -100},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	hiddenSprite := &components.SpriteComponent{
		TextureID: "hidden",
		ZOrder:    0,
		Visible:   true,
		Color:     ecs.Color{R: 255, G: 255, B: 255, A: 255},
		SourceRect: ecs.AABB{
			Min: ecs.Vector2{X: 0, Y: 0},
			Max: ecs.Vector2{X: 32, Y: 32},
		},
	}
	world.AddComponent(hiddenEntity, hiddenTransform)
	world.AddComponent(hiddenEntity, hiddenSprite)

	system.Initialize(world)

	err := system.Render(world, mockRenderer)
	assert.NoError(t, err)

	// 画面内のみ描画されることを確認
	assert.Equal(t, 1, mockRenderer.DrawCallCount)
	assert.Equal(t, "visible", mockRenderer.LastTexture)
}

func TestRenderingSystem_VisibilityFlag(t *testing.T) {
	system := systems.NewRenderingSystem()
	world := createWorldWithEntities()
	mockRenderer := &MockRenderer{}

	// 非表示エンティティ
	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 100, Y: 100},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	sprite := &components.SpriteComponent{
		TextureID: "invisible_sprite",
		ZOrder:    0,
		Visible:   false, // 非表示に設定
		Color:     ecs.Color{R: 255, G: 255, B: 255, A: 255},
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, sprite)

	system.Initialize(world)

	err := system.Render(world, mockRenderer)
	assert.NoError(t, err)

	// 非表示フラグにより描画されないことを確認
	assert.Equal(t, 0, mockRenderer.DrawCallCount)
}

func TestRenderingSystem_CameraTransform(t *testing.T) {
	system := systems.NewRenderingSystem()
	world := createWorldWithEntities()
	mockRenderer := &MockRenderer{}

	// カメラ設定
	cameraPos := ecs.Vector2{X: 100, Y: 50}
	system.SetCamera(cameraPos, 2.0, 0.0) // 2倍ズーム

	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 200, Y: 150},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	sprite := &components.SpriteComponent{
		TextureID: "test_sprite",
		ZOrder:    0,
		Visible:   true,
		Color:     ecs.Color{R: 255, G: 255, B: 255, A: 255},
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, sprite)

	system.Initialize(world)

	err := system.Render(world, mockRenderer)
	assert.NoError(t, err)

	// カメラ変換が適用されることを確認
	camera := system.GetCamera()
	assert.Equal(t, cameraPos, camera.Position)
	assert.Equal(t, 2.0, camera.Zoom)
}

func TestRenderingSystem_ViewportSettings(t *testing.T) {
	system := systems.NewRenderingSystem()

	// ビューポート設定
	system.SetViewport(10, 20, 640, 480)
	viewport := system.GetViewport()

	assert.NotNil(t, viewport)
	assert.Equal(t, 10.0, viewport.X)
	assert.Equal(t, 20.0, viewport.Y)
	assert.Equal(t, 640.0, viewport.Width)
	assert.Equal(t, 480.0, viewport.Height)
}

func TestRenderingSystem_EmptyScene(t *testing.T) {
	system := systems.NewRenderingSystem()
	world := createWorldWithEntities() // エンティティなしのワールド
	mockRenderer := &MockRenderer{}

	system.Initialize(world)

	err := system.Render(world, mockRenderer)
	assert.NoError(t, err)

	// エンティティがない場合、描画呼び出しもない
	assert.Equal(t, 0, mockRenderer.DrawCallCount)
}

// Enhanced MockRenderer for rendering system tests

type MockRenderCall struct {
	TextureID     string
	X, Y          float64
	Width, Height float64
	Scale         ecs.Vector2
	Rotation      float64
	Color         ecs.Color
}

type EnhancedMockRenderer struct {
	*MockRenderer
	RenderCalls []MockRenderCall
}

func NewEnhancedMockRenderer() *EnhancedMockRenderer {
	return &EnhancedMockRenderer{
		MockRenderer: &MockRenderer{},
		RenderCalls:  make([]MockRenderCall, 0),
	}
}

func (emr *EnhancedMockRenderer) DrawSpriteWithTransform(
	textureID string, x, y, width, height float64,
	scale ecs.Vector2, rotation float64, color ecs.Color) {

	emr.DrawSprite(textureID, ecs.Vector2{X: x, Y: y}, scale, rotation, 0)

	call := MockRenderCall{
		TextureID: textureID,
		X:         x, Y: y,
		Width: width, Height: height,
		Scale:    scale,
		Rotation: rotation,
		Color:    color,
	}
	emr.RenderCalls = append(emr.RenderCalls, call)
}
