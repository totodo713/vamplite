package components

import (
	"encoding/json"
	"errors"
	"math"

	"muscle-dreamer/internal/core/ecs"
)

// TransformComponent handles entity position, rotation, and scale
type TransformComponent struct {
	Position ecs.Vector2 `json:"position"`
	Rotation float64     `json:"rotation"`
	Scale    ecs.Vector2 `json:"scale"`

	// Hierarchy
	Parent   *TransformComponent   `json:"-"`
	Children []*TransformComponent `json:"-"`

	// Cache for performance
	dirty           bool            `json:"-"`
	transformMatrix TransformMatrix `json:"-"`
}

// NewTransformComponent creates a new transform component with default values
func NewTransformComponent() *TransformComponent {
	return &TransformComponent{
		Position: ecs.Vector2{X: 0, Y: 0},
		Rotation: 0.0,
		Scale:    ecs.Vector2{X: 1, Y: 1},
		Children: make([]*TransformComponent, 0),
		dirty:    true,
	}
}

// GetType returns the component type
func (t *TransformComponent) GetType() ecs.ComponentType {
	return ecs.ComponentTypeTransform
}

// SetPosition sets the local position
func (t *TransformComponent) SetPosition(position ecs.Vector2) {
	t.Position = position
	t.markDirty()
}

// SetRotation sets the rotation in radians
func (t *TransformComponent) SetRotation(rotation float64) {
	t.Rotation = rotation
	t.markDirty()
}

// SetScale sets the scale
func (t *TransformComponent) SetScale(scale ecs.Vector2) {
	t.Scale = scale
	t.markDirty()
}

// GetPosition returns the local position (alias for compatibility)
func (t *TransformComponent) GetPosition() ecs.Vector2 {
	return t.Position
}

// GetLocalPosition returns the local position
func (t *TransformComponent) GetLocalPosition() ecs.Vector2 {
	return t.Position
}

// GetWorldPosition returns the world position
func (t *TransformComponent) GetWorldPosition() ecs.Vector2 {
	if t.Parent == nil {
		return t.Position
	}

	parentWorldPos := t.Parent.GetWorldPosition()
	parentRotation := t.Parent.GetWorldRotation()
	parentScale := t.Parent.GetWorldScale()

	// Apply parent transformation
	cos := math.Cos(parentRotation)
	sin := math.Sin(parentRotation)

	// Rotate and scale the local position
	worldX := (float32(t.Position.X)*float32(cos)-float32(t.Position.Y)*float32(sin))*parentScale.X + parentWorldPos.X
	worldY := (float32(t.Position.X)*float32(sin)+float32(t.Position.Y)*float32(cos))*parentScale.Y + parentWorldPos.Y

	return ecs.Vector2{X: worldX, Y: worldY}
}

// GetWorldRotation returns the world rotation
func (t *TransformComponent) GetWorldRotation() float64 {
	if t.Parent == nil {
		return t.Rotation
	}
	return t.Parent.GetWorldRotation() + t.Rotation
}

// GetWorldScale returns the world scale
func (t *TransformComponent) GetWorldScale() ecs.Vector2 {
	if t.Parent == nil {
		return t.Scale
	}
	parentScale := t.Parent.GetWorldScale()
	return ecs.Vector2{
		X: t.Scale.X * parentScale.X,
		Y: t.Scale.Y * parentScale.Y,
	}
}

// SetParent sets the parent transform
func (t *TransformComponent) SetParent(parent *TransformComponent) error {
	if parent == t {
		return errors.New("cannot set self as parent")
	}

	// Check for circular reference
	if t.isAncestor(parent) {
		return errors.New("circular reference detected")
	}

	// Also check if parent would create a cycle by having t as an ancestor
	if parent != nil && parent.isAncestor(t) {
		return errors.New("circular reference detected")
	}

	// Remove from current parent
	if t.Parent != nil {
		t.Parent.removeChild(t)
	}

	// Set new parent
	t.Parent = parent
	if parent != nil {
		parent.addChild(t)
	}

	t.markDirty()
	return nil
}

// GetTransformMatrix returns the transformation matrix
func (t *TransformComponent) GetTransformMatrix() TransformMatrix {
	if t.dirty {
		t.calculateTransformMatrix()
		t.dirty = false
	}
	return t.transformMatrix
}

// isAncestor checks if the given transform is an ancestor
func (t *TransformComponent) isAncestor(ancestor *TransformComponent) bool {
	current := t.Parent
	for current != nil {
		if current == ancestor {
			return true
		}
		current = current.Parent
	}
	return false
}

// addChild adds a child transform
func (t *TransformComponent) addChild(child *TransformComponent) {
	for _, existing := range t.Children {
		if existing == child {
			return // Already a child
		}
	}
	t.Children = append(t.Children, child)
}

// removeChild removes a child transform
func (t *TransformComponent) removeChild(child *TransformComponent) {
	for i, existing := range t.Children {
		if existing == child {
			t.Children = append(t.Children[:i], t.Children[i+1:]...)
			return
		}
	}
}

// markDirty marks the transform and all children as dirty
func (t *TransformComponent) markDirty() {
	t.markDirtyRecursive(make(map[*TransformComponent]bool))
}

// markDirtyRecursive marks the transform and all children as dirty with cycle detection
func (t *TransformComponent) markDirtyRecursive(visited map[*TransformComponent]bool) {
	if visited[t] {
		return // Avoid infinite recursion
	}

	visited[t] = true
	t.dirty = true

	for _, child := range t.Children {
		if child != nil {
			child.markDirtyRecursive(visited)
		}
	}
}

// calculateTransformMatrix calculates the transformation matrix
func (t *TransformComponent) calculateTransformMatrix() {
	cos := math.Cos(t.Rotation)
	sin := math.Sin(t.Rotation)

	// 2D transformation matrix in column-major order
	// [sx*cos  -sy*sin  tx]
	// [sx*sin   sy*cos  ty]
	// [0        0       1 ]
	t.transformMatrix = TransformMatrix{
		t.Scale.X * cos, t.Scale.X * sin, 0,
		-t.Scale.Y * sin, t.Scale.Y * cos, 0,
		t.Position.X, t.Position.Y, 1,
	}
}

// Clone creates a deep copy of the component
func (t *TransformComponent) Clone() ecs.Component {
	clone := &TransformComponent{
		Position: t.Position,
		Rotation: t.Rotation,
		Scale:    t.Scale,
		Children: make([]*TransformComponent, 0),
		dirty:    true,
	}
	return clone
}

// Validate ensures the component data is valid
func (t *TransformComponent) Validate() error {
	if t.Scale.X == 0 || t.Scale.Y == 0 {
		return errors.New("scale cannot be zero")
	}
	return nil
}

// Size returns the memory size of the component in bytes
func (t *TransformComponent) Size() int {
	// Approximate size calculation
	return 32 // Position(16) + Rotation(8) + Scale(16) - excluding pointers and slices
}

// Serialize converts the component to bytes
func (t *TransformComponent) Serialize() ([]byte, error) {
	data := struct {
		Position ecs.Vector2 `json:"position"`
		Rotation float64     `json:"rotation"`
		Scale    ecs.Vector2 `json:"scale"`
	}{
		Position: t.Position,
		Rotation: t.Rotation,
		Scale:    t.Scale,
	}
	return json.Marshal(data)
}

// Deserialize loads component data from bytes
func (t *TransformComponent) Deserialize(data []byte) error {
	var serialData struct {
		Position ecs.Vector2 `json:"position"`
		Rotation float64     `json:"rotation"`
		Scale    ecs.Vector2 `json:"scale"`
	}

	if err := json.Unmarshal(data, &serialData); err != nil {
		return err
	}

	t.Position = serialData.Position
	t.Rotation = serialData.Rotation
	t.Scale = serialData.Scale
	t.markDirty()

	return nil
}
