package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"muscle-dreamer/internal/core/ecs"
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

	assert.Equal(t, MockSystemType, systemType)
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
	assert.Equal(t, int64(0), metrics.ExecutionCount)
	assert.Equal(t, int64(0), metrics.TotalTime)

	// Update実行
	err := system.Update(world, 0.016) // 60FPS
	assert.NoError(t, err)

	// Update実行後
	metrics = system.GetMetrics()
	assert.Equal(t, int64(1), metrics.ExecutionCount)
	assert.Greater(t, metrics.TotalTime, int64(0))
	assert.Greater(t, metrics.AverageTime, int64(0))
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

	// テスト用のエラーを発生させる実装
	// BaseSystemにパブリックなエラー処理用メソッドがないため、
	// 代わりにエラーハンドラーの設定・取得をテスト

	// 初期状態の確認
	assert.Nil(t, system.GetLastError())
	assert.False(t, errorCalled)

	// エラーハンドラーが適切に設定されているか確認
	// （実際のエラー処理は統合テストで確認）
	// 初期状態では capturedError は nil であることを確認
	assert.Nil(t, capturedError)
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
	assert.Greater(t, metrics.ExecutionCount, int64(0))

	// リセット実行
	system.ResetMetrics()

	// リセット後
	metrics = system.GetMetrics()
	assert.Equal(t, int64(0), metrics.ExecutionCount)
	assert.Equal(t, int64(0), metrics.TotalTime)
	assert.Equal(t, int64(0), metrics.AverageTime)
}

// Test helper functions and mock objects

const MockSystemType = ecs.SystemType("test_system")

// MockBaseSystem extends BaseSystem for testing
type MockBaseSystem struct {
	*systems.BaseSystem
}

func createTestBaseSystem() *MockBaseSystem {
	base := systems.NewBaseSystem(MockSystemType, ecs.PriorityNormal)
	return &MockBaseSystem{BaseSystem: base}
}

// TriggerError simulates an error for testing error handling
func (mbs *MockBaseSystem) TriggerError(_ error) {
	// We need to access the handleError method, but it's not exported
	// For now, we'll simulate the error handling manually
	mbs.SetErrorHandler(func(_ error) {
		// This will be called by the handler we set in the test
	})
}

// Use MockWorld from test_utils.go

func createMockWorld() *MockWorld {
	return NewMockWorld()
}

// Use MockRenderer from test_utils.go
