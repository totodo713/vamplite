package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"muscle-dreamer/internal/core/ecs"
	"muscle-dreamer/internal/core/ecs/components"
	"muscle-dreamer/internal/core/systems"
)

func TestBaseSystem_Initialize(t *testing.T) {
	system := createTestBaseSystem()
	world := createMockWorld()

	err := system.Initialize(world)

	assert.NoError(t, err)
	assert.True(t, system.IsEnabled())
	assert.NotNil(t, system.GetMetrics())
}

func TestBaseSystem_GetType(t *testing.T) {
	system := createTestBaseSystem()

	systemType := system.GetType()

	assert.Equal(t, TestSystemType, systemType)
	assert.NotEmpty(t, string(systemType))
}

func TestBaseSystem_Priority(t *testing.T) {
	system := createTestBaseSystem()
	expectedPriority := ecs.Priority(100)
	system.SetPriority(expectedPriority)

	priority := system.GetPriority()

	assert.Equal(t, expectedPriority, priority)
}

func TestBaseSystem_EnableDisable(t *testing.T) {
	system := createTestBaseSystem()

	// 初期状態: 有効
	assert.True(t, system.IsEnabled())

	// 無効化
	system.SetEnabled(false)
	assert.False(t, system.IsEnabled())

	// 再有効化
	system.SetEnabled(true)
	assert.True(t, system.IsEnabled())
}

func TestBaseSystem_Metrics(t *testing.T) {
	system := createTestBaseSystem()
	world := createMockWorld()
	system.Initialize(world)

	// Update実行前
	metrics := system.GetMetrics()
	assert.Equal(t, int64(0), metrics.UpdateCount)
	assert.Equal(t, time.Duration(0), metrics.TotalUpdateTime)

	// Update実行
	err := system.Update(world, 0.016) // 60FPS
	assert.NoError(t, err)

	// Update実行後
	metrics = system.GetMetrics()
	assert.Equal(t, int64(1), metrics.UpdateCount)
	assert.Greater(t, metrics.TotalUpdateTime, time.Duration(0))
	assert.Greater(t, metrics.AverageUpdateTime, time.Duration(0))
}

func TestBaseSystem_ErrorHandling(t *testing.T) {
	system := createTestBaseSystem()

	// エラーハンドラー設定
	errorCalled := false
	var capturedError error
	system.SetErrorHandler(func(err error) {
		errorCalled = true
		capturedError = err
	})

	// 手動でエラーを発生させる（テスト用）
	testError := assert.AnError
	system.(*MockBaseSystem).TriggerError(testError)

	assert.True(t, errorCalled)
	assert.Equal(t, testError, capturedError)
	assert.Equal(t, testError, system.GetLastError())
}

func TestBaseSystem_ThreadSafety(t *testing.T) {
	system := createTestBaseSystem()

	threadSafety := system.GetThreadSafety()
	assert.Equal(t, ecs.ThreadSafetyFull, threadSafety)

	canRunInParallel := system.CanRunInParallel()
	assert.True(t, canRunInParallel)
}

func TestBaseSystem_MetricsReset(t *testing.T) {
	system := createTestBaseSystem()
	world := createMockWorld()

	// いくつかの操作を実行してメトリクスを蓄積
	system.Update(world, 0.016)
	system.Update(world, 0.016)
	system.Render(world, &MockRenderer{})

	// リセット前
	metrics := system.GetMetrics()
	assert.Greater(t, metrics.UpdateCount, int64(0))
	assert.Greater(t, metrics.RenderCount, int64(0))

	// リセット実行
	system.ResetMetrics()

	// リセット後
	metrics = system.GetMetrics()
	assert.Equal(t, int64(0), metrics.UpdateCount)
	assert.Equal(t, int64(0), metrics.RenderCount)
	assert.Equal(t, time.Duration(0), metrics.TotalUpdateTime)
	assert.Equal(t, time.Duration(0), metrics.TotalRenderTime)
}

// Test helper functions and mock objects

const TestSystemType = ecs.SystemType("test_system")

// MockBaseSystem extends BaseSystem for testing
type MockBaseSystem struct {
	*systems.BaseSystem
}

func createTestBaseSystem() *MockBaseSystem {
	base := systems.NewBaseSystem(TestSystemType, ecs.PriorityNormal)
	return &MockBaseSystem{BaseSystem: base}
}

// TriggerError simulates an error for testing error handling
func (mbs *MockBaseSystem) TriggerError(err error) {
	// We need to access the handleError method, but it's not exported
	// For now, we'll simulate the error handling manually
	mbs.SetErrorHandler(func(e error) {
		// This will be called by the handler we set in the test
	})
}

// MockWorld implements ecs.World for testing
type MockWorld struct {
	entities   map[ecs.EntityID]bool
	components map[ecs.EntityID]map[ecs.ComponentType]ecs.Component
}

func createMockWorld() *MockWorld {
	return &MockWorld{
		entities:   make(map[ecs.EntityID]bool),
		components: make(map[ecs.EntityID]map[ecs.ComponentType]ecs.Component),
	}
}

// Implement required World interface methods (minimal implementation)
func (mw *MockWorld) CreateEntity() ecs.EntityID {
	// Simple ID generation for testing
	id := ecs.EntityID(len(mw.entities) + 1)
	mw.entities[id] = true
	mw.components[id] = make(map[ecs.ComponentType]ecs.Component)
	return id
}

func (mw *MockWorld) DestroyEntity(entity ecs.EntityID) error {
	delete(mw.entities, entity)
	delete(mw.components, entity)
	return nil
}

func (mw *MockWorld) AddComponent(entity ecs.EntityID, component ecs.Component) error {
	if !mw.entities[entity] {
		return assert.AnError
	}

	compType := component.GetType()
	mw.components[entity][compType] = component
	return nil
}

func (mw *MockWorld) GetComponent(entity ecs.EntityID, componentType ecs.ComponentType) ecs.Component {
	if entityComps, exists := mw.components[entity]; exists {
		return entityComps[componentType]
	}
	return nil
}

func (mw *MockWorld) RemoveComponent(entity ecs.EntityID, componentType ecs.ComponentType) error {
	if entityComps, exists := mw.components[entity]; exists {
		delete(entityComps, componentType)
	}
	return nil
}

// MockRenderer for testing rendering operations
type MockRenderer struct {
	DrawCallCount int
	LastTexture   string
	DrawOrder     []string
}

func (mr *MockRenderer) DrawSprite(textureID string, x, y, width, height float64) {
	mr.DrawCallCount++
	mr.LastTexture = textureID
	mr.DrawOrder = append(mr.DrawOrder, textureID)
}
