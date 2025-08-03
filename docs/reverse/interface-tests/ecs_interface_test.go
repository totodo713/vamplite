// ========================================================
// ECS Interface Test Suite
// Entity Component System インターフェーステスト
// ========================================================

package interfaces_test

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"muscle-dreamer/docs/reverse/interfaces"
	"testing"
	"time"
)

// ========================================================
// Mock Implementations for Testing
// ========================================================

// MockEntityManager - EntityManager のモック実装
type MockEntityManager struct {
	mock.Mock
	entities   map[interfaces.EntityID]bool
	components map[interfaces.EntityID]map[interfaces.ComponentType]interfaces.Component
	nextID     interfaces.EntityID
}

func NewMockEntityManager() *MockEntityManager {
	return &MockEntityManager{
		entities:   make(map[interfaces.EntityID]bool),
		components: make(map[interfaces.EntityID]map[interfaces.ComponentType]interfaces.Component),
		nextID:     1,
	}
}

func (m *MockEntityManager) CreateEntity() interfaces.EntityID {
	args := m.Called()

	id := m.nextID
	m.nextID++
	m.entities[id] = true
	m.components[id] = make(map[interfaces.ComponentType]interfaces.Component)

	if args.Get(0) != nil {
		return args.Get(0).(interfaces.EntityID)
	}
	return id
}

func (m *MockEntityManager) DestroyEntity(id interfaces.EntityID) {
	m.Called(id)
	delete(m.entities, id)
	delete(m.components, id)
}

func (m *MockEntityManager) GetComponent(id interfaces.EntityID, componentType interfaces.ComponentType) interfaces.Component {
	args := m.Called(id, componentType)

	if comps, exists := m.components[id]; exists {
		if comp, found := comps[componentType]; found {
			return comp
		}
	}

	return args.Get(0)
}

func (m *MockEntityManager) AddComponent(id interfaces.EntityID, component interfaces.Component) {
	m.Called(id, component)

	if _, exists := m.components[id]; !exists {
		m.components[id] = make(map[interfaces.ComponentType]interfaces.Component)
	}
	m.components[id][component.GetType()] = component
}

func (m *MockEntityManager) RemoveComponent(id interfaces.EntityID, componentType interfaces.ComponentType) {
	m.Called(id, componentType)

	if comps, exists := m.components[id]; exists {
		delete(comps, componentType)
	}
}

func (m *MockEntityManager) HasComponent(id interfaces.EntityID, componentType interfaces.ComponentType) bool {
	args := m.Called(id, componentType)

	if comps, exists := m.components[id]; exists {
		_, found := comps[componentType]
		return found
	}

	if args.Get(0) != nil {
		return args.Bool(0)
	}
	return false
}

func (m *MockEntityManager) GetEntitiesWith(componentType interfaces.ComponentType) []interfaces.EntityID {
	args := m.Called(componentType)

	var result []interfaces.EntityID
	for entityID, comps := range m.components {
		if _, has := comps[componentType]; has {
			result = append(result, entityID)
		}
	}

	if args.Get(0) != nil {
		return args.Get(0).([]interfaces.EntityID)
	}
	return result
}

// MockSystem - System のモック実装
type MockSystem struct {
	mock.Mock
	SystemType   interfaces.SystemType
	UpdateCalled bool
	LastDelta    float64
}

func NewMockSystem(systemType interfaces.SystemType) *MockSystem {
	return &MockSystem{
		SystemType: systemType,
	}
}

func (m *MockSystem) Update(deltaTime float64, entities interfaces.EntityManager) error {
	args := m.Called(deltaTime, entities)
	m.UpdateCalled = true
	m.LastDelta = deltaTime
	return args.Error(0)
}

func (m *MockSystem) GetType() interfaces.SystemType {
	return m.SystemType
}

// MockSystemManager - SystemManager のモック実装
type MockSystemManager struct {
	mock.Mock
	systems []interfaces.System
}

func NewMockSystemManager() *MockSystemManager {
	return &MockSystemManager{
		systems: make([]interfaces.System, 0),
	}
}

func (m *MockSystemManager) RegisterSystem(system interfaces.System) {
	m.Called(system)
	m.systems = append(m.systems, system)
}

func (m *MockSystemManager) UnregisterSystem(systemType interfaces.SystemType) {
	m.Called(systemType)

	for i, sys := range m.systems {
		if sys.GetType() == systemType {
			m.systems = append(m.systems[:i], m.systems[i+1:]...)
			break
		}
	}
}

func (m *MockSystemManager) UpdateSystems(deltaTime float64, entities interfaces.EntityManager) error {
	args := m.Called(deltaTime, entities)

	for _, system := range m.systems {
		if err := system.Update(deltaTime, entities); err != nil {
			return err
		}
	}

	return args.Error(0)
}

func (m *MockSystemManager) RenderSystems(screen *ebiten.Image) {
	m.Called(screen)
}

// ========================================================
// Interface Contract Tests
// ========================================================

// TestEntityManagerInterface - EntityManager インターフェース契約テスト
func TestEntityManagerInterface(t *testing.T) {
	em := NewMockEntityManager()

	t.Run("CreateEntity", func(t *testing.T) {
		em.On("CreateEntity").Return(interfaces.EntityID(1))

		entity := em.CreateEntity()
		assert.NotEqual(t, interfaces.EntityID(0), entity)

		em.AssertExpectations(t)
	})

	t.Run("EntityLifecycle", func(t *testing.T) {
		entity := interfaces.EntityID(1)
		component := &MockTransformComponent{X: 100, Y: 200}

		em.On("AddComponent", entity, component).Return()
		em.On("HasComponent", entity, interfaces.ComponentTransform).Return(true)
		em.On("GetComponent", entity, interfaces.ComponentTransform).Return(component)
		em.On("RemoveComponent", entity, interfaces.ComponentTransform).Return()
		em.On("DestroyEntity", entity).Return()

		// コンポーネント追加
		em.AddComponent(entity, component)

		// コンポーネント存在確認
		hasComponent := em.HasComponent(entity, interfaces.ComponentTransform)
		assert.True(t, hasComponent)

		// コンポーネント取得
		retrieved := em.GetComponent(entity, interfaces.ComponentTransform)
		assert.Equal(t, component, retrieved)

		// コンポーネント削除
		em.RemoveComponent(entity, interfaces.ComponentTransform)

		// エンティティ削除
		em.DestroyEntity(entity)

		em.AssertExpectations(t)
	})

	t.Run("GetEntitiesWith", func(t *testing.T) {
		expectedEntities := []interfaces.EntityID{1, 2, 3}

		em.On("GetEntitiesWith", interfaces.ComponentTransform).Return(expectedEntities)

		entities := em.GetEntitiesWith(interfaces.ComponentTransform)
		assert.Equal(t, expectedEntities, entities)

		em.AssertExpectations(t)
	})
}

// TestSystemInterface - System インターフェース契約テスト
func TestSystemInterface(t *testing.T) {
	t.Run("SystemExecution", func(t *testing.T) {
		system := NewMockSystem(interfaces.SystemMovement)
		em := NewMockEntityManager()

		system.On("Update", 0.016, em).Return(nil)

		err := system.Update(0.016, em)
		assert.NoError(t, err)
		assert.True(t, system.UpdateCalled)
		assert.Equal(t, 0.016, system.LastDelta)

		systemType := system.GetType()
		assert.Equal(t, interfaces.SystemMovement, systemType)

		system.AssertExpectations(t)
	})
}

// TestSystemManagerInterface - SystemManager インターフェース契約テスト
func TestSystemManagerInterface(t *testing.T) {
	sm := NewMockSystemManager()

	t.Run("SystemRegistration", func(t *testing.T) {
		system1 := NewMockSystem(interfaces.SystemMovement)
		system2 := NewMockSystem(interfaces.SystemRendering)

		sm.On("RegisterSystem", system1).Return()
		sm.On("RegisterSystem", system2).Return()

		sm.RegisterSystem(system1)
		sm.RegisterSystem(system2)

		assert.Len(t, sm.systems, 2)

		sm.AssertExpectations(t)
	})

	t.Run("SystemUnregistration", func(t *testing.T) {
		system := NewMockSystem(interfaces.SystemMovement)
		sm.systems = []interfaces.System{system}

		sm.On("UnregisterSystem", interfaces.SystemMovement).Return()

		sm.UnregisterSystem(interfaces.SystemMovement)

		assert.Len(t, sm.systems, 0)

		sm.AssertExpectations(t)
	})

	t.Run("SystemsUpdate", func(t *testing.T) {
		system1 := NewMockSystem(interfaces.SystemMovement)
		system2 := NewMockSystem(interfaces.SystemRendering)
		em := NewMockEntityManager()

		system1.On("Update", 0.016, em).Return(nil)
		system2.On("Update", 0.016, em).Return(nil)
		sm.On("UpdateSystems", 0.016, em).Return(nil)

		sm.systems = []interfaces.System{system1, system2}

		err := sm.UpdateSystems(0.016, em)
		assert.NoError(t, err)

		assert.True(t, system1.UpdateCalled)
		assert.True(t, system2.UpdateCalled)

		sm.AssertExpectations(t)
	})
}

// ========================================================
// Component Interface Tests
// ========================================================

// MockTransformComponent - TransformComponent のモック
type MockTransformComponent struct {
	X, Y     float64
	Rotation float64
	ScaleX   float64
	ScaleY   float64
}

func (t *MockTransformComponent) GetType() interfaces.ComponentType {
	return interfaces.ComponentTransform
}

// MockSpriteComponent - SpriteComponent のモック
type MockSpriteComponent struct {
	Width  int
	Height int
}

func (s *MockSpriteComponent) GetType() interfaces.ComponentType {
	return interfaces.ComponentSprite
}

// TestComponentInterface - Component インターフェース契約テスト
func TestComponentInterface(t *testing.T) {
	t.Run("TransformComponent", func(t *testing.T) {
		transform := &MockTransformComponent{
			X: 100, Y: 200, Rotation: 0, ScaleX: 1, ScaleY: 1,
		}

		componentType := transform.GetType()
		assert.Equal(t, interfaces.ComponentTransform, componentType)
		assert.Equal(t, 100.0, transform.X)
		assert.Equal(t, 200.0, transform.Y)
	})

	t.Run("SpriteComponent", func(t *testing.T) {
		sprite := &MockSpriteComponent{
			Width: 32, Height: 32,
		}

		componentType := sprite.GetType()
		assert.Equal(t, interfaces.ComponentSprite, componentType)
		assert.Equal(t, 32, sprite.Width)
		assert.Equal(t, 32, sprite.Height)
	})
}

// ========================================================
// Integration Tests
// ========================================================

// TestECSIntegration - ECS統合テスト
func TestECSIntegration(t *testing.T) {
	em := NewMockEntityManager()
	sm := NewMockSystemManager()

	t.Run("CompleteECSWorkflow", func(t *testing.T) {
		// エンティティ作成
		entity := em.CreateEntity()

		// コンポーネント追加
		transform := &MockTransformComponent{X: 100, Y: 200}
		sprite := &MockSpriteComponent{Width: 32, Height: 32}

		em.AddComponent(entity, transform)
		em.AddComponent(entity, sprite)

		// システム作成・登録
		movementSystem := NewMockSystem(interfaces.SystemMovement)
		renderSystem := NewMockSystem(interfaces.SystemRendering)

		movementSystem.On("Update", mock.AnythingOfType("float64"), em).Return(nil)
		renderSystem.On("Update", mock.AnythingOfType("float64"), em).Return(nil)

		sm.RegisterSystem(movementSystem)
		sm.RegisterSystem(renderSystem)

		// システム実行
		sm.On("UpdateSystems", mock.AnythingOfType("float64"), em).Return(nil)
		err := sm.UpdateSystems(0.016, em)

		assert.NoError(t, err)
		assert.True(t, movementSystem.UpdateCalled)
		assert.True(t, renderSystem.UpdateCalled)
	})
}

// ========================================================
// Performance Tests
// ========================================================

// TestECSPerformance - ECSパフォーマンステスト
func TestECSPerformance(t *testing.T) {
	t.Run("EntityCreationPerformance", func(t *testing.T) {
		em := NewMockEntityManager()

		const entityCount = 10000

		start := time.Now()
		for i := 0; i < entityCount; i++ {
			entity := em.CreateEntity()
			transform := &MockTransformComponent{
				X: float64(i), Y: float64(i),
			}
			em.AddComponent(entity, transform)
		}
		elapsed := time.Since(start)

		// 10,000エンティティ作成 < 100ms
		assert.Less(t, elapsed, 100*time.Millisecond)

		t.Logf("Created %d entities in %v", entityCount, elapsed)
	})

	t.Run("SystemUpdatePerformance", func(t *testing.T) {
		sm := NewMockSystemManager()
		em := NewMockEntityManager()

		// 10個のシステム作成
		for i := 0; i < 10; i++ {
			system := NewMockSystem(interfaces.SystemType("System" + string(rune(i))))
			system.On("Update", mock.AnythingOfType("float64"), em).Return(nil)
			sm.RegisterSystem(system)
		}

		sm.On("UpdateSystems", mock.AnythingOfType("float64"), em).Return(nil)

		// 1000回システム更新
		start := time.Now()
		for i := 0; i < 1000; i++ {
			sm.UpdateSystems(0.016, em)
		}
		elapsed := time.Since(start)

		// 1000回更新 < 1秒
		assert.Less(t, elapsed, time.Second)

		t.Logf("1000 system updates in %v", elapsed)
	})
}

// ========================================================
// Error Handling Tests
// ========================================================

// MockErrorSystem - エラーを発生させるシステム
type MockErrorSystem struct {
	SystemType  interfaces.SystemType
	ShouldError bool
	ErrorMsg    string
}

func (m *MockErrorSystem) Update(deltaTime float64, entities interfaces.EntityManager) error {
	if m.ShouldError {
		return errors.New(m.ErrorMsg)
	}
	return nil
}

func (m *MockErrorSystem) GetType() interfaces.SystemType {
	return m.SystemType
}

// TestErrorHandling - エラーハンドリングテスト
func TestErrorHandling(t *testing.T) {
	t.Run("SystemUpdateError", func(t *testing.T) {
		sm := NewMockSystemManager()
		em := NewMockEntityManager()

		errorSystem := &MockErrorSystem{
			SystemType:  interfaces.SystemMovement,
			ShouldError: true,
			ErrorMsg:    "Test system error",
		}

		normalSystem := NewMockSystem(interfaces.SystemRendering)
		normalSystem.On("Update", mock.AnythingOfType("float64"), em).Return(nil)

		sm.systems = []interfaces.System{errorSystem, normalSystem}
		sm.On("UpdateSystems", mock.AnythingOfType("float64"), em).Return(errors.New("Test system error"))

		err := sm.UpdateSystems(0.016, em)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Test system error")
	})
}

// ========================================================
// Benchmark Tests
// ========================================================

// BenchmarkEntityCreation - エンティティ作成ベンチマーク
func BenchmarkEntityCreation(b *testing.B) {
	em := NewMockEntityManager()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		entity := em.CreateEntity()
		transform := &MockTransformComponent{
			X: float64(i), Y: float64(i),
		}
		em.AddComponent(entity, transform)
	}
}

// BenchmarkSystemUpdate - システム更新ベンチマーク
func BenchmarkSystemUpdate(b *testing.B) {
	system := NewMockSystem(interfaces.SystemMovement)
	em := NewMockEntityManager()

	system.On("Update", mock.AnythingOfType("float64"), em).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		system.Update(0.016, em)
	}
}

// BenchmarkComponentQuery - コンポーネントクエリベンチマーク
func BenchmarkComponentQuery(b *testing.B) {
	em := NewMockEntityManager()

	// 1000エンティティを事前作成
	for i := 0; i < 1000; i++ {
		entity := em.CreateEntity()
		em.AddComponent(entity, &MockTransformComponent{})
	}

	em.On("GetEntitiesWith", interfaces.ComponentTransform).Return(make([]interfaces.EntityID, 1000))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		em.GetEntitiesWith(interfaces.ComponentTransform)
	}
}
