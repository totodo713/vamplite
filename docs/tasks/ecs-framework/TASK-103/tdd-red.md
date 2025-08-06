# TASK-103: SystemManager実装 - Red段階 (失敗するテスト実装)

## Red段階の目的

TDDのRed段階では、仕様を満たす失敗するテストを実装します。これにより：
1. 実装すべき機能の明確化
2. テストの妥当性確認
3. 最小限の実装への指針提供

## 実装するテストファイル

### 1. SystemManager基本テストファイル作成

#### ファイル: `internal/core/ecs/system_manager_test.go`

```go
package ecs

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// ==============================================
// Mock System Implementation for Testing
// ==============================================

type MockSystem struct {
	systemType     SystemType
	updateCalled   int
	renderCalled   int
	initCalled     int
	shutdownCalled int
	shouldError    bool
	execTime       time.Duration
	mutex          sync.Mutex
}

func NewMockSystem(systemType SystemType) *MockSystem {
	return &MockSystem{
		systemType: systemType,
		execTime:   1 * time.Millisecond,
	}
}

func (s *MockSystem) GetType() SystemType {
	return s.systemType
}

func (s *MockSystem) GetPriority() Priority {
	return PriorityNormal
}

func (s *MockSystem) GetThreadSafety() ThreadSafetyLevel {
	return ThreadSafetyReadWrite
}

func (s *MockSystem) Update(world World, deltaTime float64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	time.Sleep(s.execTime) // Simulate work
	s.updateCalled++
	
	if s.shouldError {
		return errors.New("mock system error")
	}
	return nil
}

func (s *MockSystem) Render(world World, renderer interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	time.Sleep(s.execTime) // Simulate work
	s.renderCalled++
	
	if s.shouldError {
		return errors.New("mock system render error")
	}
	return nil
}

func (s *MockSystem) Initialize(world World) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.initCalled++
	
	if s.shouldError {
		return errors.New("mock system init error")
	}
	return nil
}

func (s *MockSystem) Shutdown() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.shutdownCalled++
	
	if s.shouldError {
		return errors.New("mock system shutdown error")
	}
	return nil
}

func (s *MockSystem) GetDependencies() []SystemType {
	return []SystemType{}
}

func (s *MockSystem) GetRequiredComponents() []ComponentType {
	return []ComponentType{}
}

func (s *MockSystem) GetCallCounts() (update, render, init, shutdown int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.updateCalled, s.renderCalled, s.initCalled, s.shutdownCalled
}

func (s *MockSystem) SetShouldError(shouldError bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.shouldError = shouldError
}

func (s *MockSystem) SetExecutionTime(duration time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.execTime = duration
}

// ==============================================
// Mock World Implementation for Testing
// ==============================================

type MockWorld struct{}

func NewMockWorld() *MockWorld {
	return &MockWorld{}
}

func (w *MockWorld) CreateEntity() EntityID { return EntityID(1) }
func (w *MockWorld) DestroyEntity(EntityID) error { return nil }
func (w *MockWorld) IsEntityValid(EntityID) bool { return true }
func (w *MockWorld) GetEntityCount() int { return 0 }
func (w *MockWorld) GetAllEntities() []EntityID { return []EntityID{} }
func (w *MockWorld) AddComponent(EntityID, Component) error { return nil }
func (w *MockWorld) RemoveComponent(EntityID, ComponentType) error { return nil }
func (w *MockWorld) GetComponent(EntityID, ComponentType) (Component, error) { return nil, nil }
func (w *MockWorld) HasComponent(EntityID, ComponentType) bool { return false }
func (w *MockWorld) GetComponentStore() ComponentStore { return nil }
func (w *MockWorld) GetEntityManager() EntityManager { return nil }
func (w *MockWorld) GetSystemManager() SystemManager { return nil }
func (w *MockWorld) GetEventBus() EventBus { return nil }
func (w *MockWorld) Update(float64) error { return nil }
func (w *MockWorld) Render(interface{}) error { return nil }
func (w *MockWorld) Initialize() error { return nil }
func (w *MockWorld) Shutdown() error { return nil }

// ==============================================
// TC-SM-001: システム登録機能テスト
// ==============================================

// TC-SM-001-01: 正常なシステム登録
func TestSystemManager_RegisterSystem_Success(t *testing.T) {
	// Given: 新しいSystemManagerとMockSystem
	sm := NewSystemManager()
	system := NewMockSystem(SystemType("TestSystem"))

	// When: システムを登録する
	err := sm.RegisterSystem(system)

	// Then: システムが正常に登録され、GetSystemで取得できる
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	retrievedSystem, err := sm.GetSystem(SystemType("TestSystem"))
	if err != nil {
		t.Errorf("Expected no error when getting system, got %v", err)
	}
	if retrievedSystem != system {
		t.Errorf("Expected retrieved system to be the same as registered")
	}

	if sm.GetSystemCount() != 1 {
		t.Errorf("Expected system count to be 1, got %d", sm.GetSystemCount())
	}
}

// TC-SM-001-02: 重複システム登録拒否
func TestSystemManager_RegisterSystem_DuplicateError(t *testing.T) {
	// Given: 既に登録済みのシステムがあるSystemManager
	sm := NewSystemManager()
	system1 := NewMockSystem(SystemType("TestSystem"))
	system2 := NewMockSystem(SystemType("TestSystem"))
	
	sm.RegisterSystem(system1)

	// When: 同じ型のシステムを再度登録しようとする
	err := sm.RegisterSystem(system2)

	// Then: エラーが返され、元のシステムが保持される
	if err == nil {
		t.Error("Expected error for duplicate system registration")
	}

	retrievedSystem, _ := sm.GetSystem(SystemType("TestSystem"))
	if retrievedSystem != system1 {
		t.Error("Expected original system to be preserved")
	}

	if sm.GetSystemCount() != 1 {
		t.Errorf("Expected system count to remain 1, got %d", sm.GetSystemCount())
	}
}

// TC-SM-001-03: 優先度付きシステム登録
func TestSystemManager_RegisterSystemWithPriority_Success(t *testing.T) {
	// Given: SystemManagerと異なる優先度のシステム
	sm := NewSystemManager()
	systemHigh := NewMockSystem(SystemType("HighPrioritySystem"))
	systemLow := NewMockSystem(SystemType("LowPrioritySystem"))

	// When: 優先度付きでシステムを登録する
	err1 := sm.RegisterSystemWithPriority(systemHigh, PriorityHigh)
	err2 := sm.RegisterSystemWithPriority(systemLow, PriorityLow)

	// Then: 優先度順でシステムが管理される
	if err1 != nil || err2 != nil {
		t.Errorf("Expected no errors, got %v, %v", err1, err2)
	}

	systems := sm.GetSystemsByPriority(PriorityHigh)
	if len(systems) != 1 || systems[0] != systemHigh {
		t.Error("Expected high priority system to be registered with correct priority")
	}

	systems = sm.GetSystemsByPriority(PriorityLow)
	if len(systems) != 1 || systems[0] != systemLow {
		t.Error("Expected low priority system to be registered with correct priority")
	}
}

// TC-SM-001-04: nilシステム登録エラー
func TestSystemManager_RegisterSystem_NilSystemError(t *testing.T) {
	// Given: SystemManager
	sm := NewSystemManager()

	// When: nilシステムを登録しようとする
	err := sm.RegisterSystem(nil)

	// Then: エラーが返される
	if err == nil {
		t.Error("Expected error for nil system registration")
	}

	if sm.GetSystemCount() != 0 {
		t.Errorf("Expected system count to be 0, got %d", sm.GetSystemCount())
	}
}

// TC-SM-001-05: システム登録解除
func TestSystemManager_UnregisterSystem_Success(t *testing.T) {
	// Given: 登録済みシステムがあるSystemManager
	sm := NewSystemManager()
	system := NewMockSystem(SystemType("TestSystem"))
	sm.RegisterSystem(system)

	// When: システムを登録解除する
	err := sm.UnregisterSystem(SystemType("TestSystem"))

	// Then: システムが削除され、GetSystemでエラーが返される
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_, err = sm.GetSystem(SystemType("TestSystem"))
	if err == nil {
		t.Error("Expected error when getting unregistered system")
	}

	if sm.GetSystemCount() != 0 {
		t.Errorf("Expected system count to be 0, got %d", sm.GetSystemCount())
	}
}

// ==============================================
// TC-SM-002: システム状態管理テスト
// ==============================================

// TC-SM-002-01: システム有効化・無効化
func TestSystemManager_EnableDisableSystem_Success(t *testing.T) {
	// Given: 登録済みシステムがあるSystemManager
	sm := NewSystemManager()
	system := NewMockSystem(SystemType("TestSystem"))
	sm.RegisterSystem(system)

	// When: システムを無効化・有効化する
	err1 := sm.DisableSystem(SystemType("TestSystem"))
	enabled1 := sm.IsSystemEnabled(SystemType("TestSystem"))
	
	err2 := sm.EnableSystem(SystemType("TestSystem"))
	enabled2 := sm.IsSystemEnabled(SystemType("TestSystem"))

	// Then: IsSystemEnabledが正しい状態を返す
	if err1 != nil || err2 != nil {
		t.Errorf("Expected no errors, got %v, %v", err1, err2)
	}

	if enabled1 {
		t.Error("Expected system to be disabled")
	}

	if !enabled2 {
		t.Error("Expected system to be enabled")
	}
}

// TC-SM-002-02: 有効・無効システム一覧取得
func TestSystemManager_GetEnabledDisabledSystems_Success(t *testing.T) {
	// Given: 有効・無効システムが混在するSystemManager
	sm := NewSystemManager()
	system1 := NewMockSystem(SystemType("EnabledSystem"))
	system2 := NewMockSystem(SystemType("DisabledSystem"))
	
	sm.RegisterSystem(system1)
	sm.RegisterSystem(system2)
	sm.DisableSystem(SystemType("DisabledSystem"))

	// When: 有効・無効システム一覧を取得する
	enabledSystems := sm.GetEnabledSystems()
	disabledSystems := sm.GetDisabledSystems()

	// Then: 正しいシステム一覧が返される
	if len(enabledSystems) != 1 || enabledSystems[0] != SystemType("EnabledSystem") {
		t.Errorf("Expected 1 enabled system 'EnabledSystem', got %v", enabledSystems)
	}

	if len(disabledSystems) != 1 || disabledSystems[0] != SystemType("DisabledSystem") {
		t.Errorf("Expected 1 disabled system 'DisabledSystem', got %v", disabledSystems)
	}
}

// ==============================================
// TC-SM-003: 依存関係設定・検証テスト
// ==============================================

// TC-SM-003-01: 依存関係設定
func TestSystemManager_SetSystemDependency_Success(t *testing.T) {
	// Given: 2つの登録済みシステムがあるSystemManager
	sm := NewSystemManager()
	systemA := NewMockSystem(SystemType("SystemA"))
	systemB := NewMockSystem(SystemType("SystemB"))
	
	sm.RegisterSystem(systemA)
	sm.RegisterSystem(systemB)

	// When: システムAがシステムBに依存するよう設定する
	err := sm.SetSystemDependency(SystemType("SystemA"), SystemType("SystemB"))

	// Then: 依存関係が設定され、GetSystemDependenciesで取得できる
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	dependencies := sm.GetSystemDependencies(SystemType("SystemA"))
	if len(dependencies) != 1 || dependencies[0] != SystemType("SystemB") {
		t.Errorf("Expected SystemA to depend on SystemB, got %v", dependencies)
	}

	dependents := sm.GetSystemDependents(SystemType("SystemB"))
	if len(dependents) != 1 || dependents[0] != SystemType("SystemA") {
		t.Errorf("Expected SystemB to have SystemA as dependent, got %v", dependents)
	}
}

// TC-SM-003-02: 循環依存検出・拒否
func TestSystemManager_SetSystemDependency_CyclicError(t *testing.T) {
	// Given: A→B依存が設定済みのSystemManager
	sm := NewSystemManager()
	systemA := NewMockSystem(SystemType("SystemA"))
	systemB := NewMockSystem(SystemType("SystemB"))
	
	sm.RegisterSystem(systemA)
	sm.RegisterSystem(systemB)
	sm.SetSystemDependency(SystemType("SystemA"), SystemType("SystemB"))

	// When: B→A依存を設定しようとする
	err := sm.SetSystemDependency(SystemType("SystemB"), SystemType("SystemA"))

	// Then: 循環依存エラーが返される
	if err == nil {
		t.Error("Expected error for cyclic dependency")
	}

	// A→Bの依存関係は保持される
	dependencies := sm.GetSystemDependencies(SystemType("SystemA"))
	if len(dependencies) != 1 || dependencies[0] != SystemType("SystemB") {
		t.Error("Expected original dependency to be preserved")
	}

	// B→Aの依存関係は設定されない
	dependencies = sm.GetSystemDependencies(SystemType("SystemB"))
	if len(dependencies) != 0 {
		t.Error("Expected no dependencies for SystemB")
	}
}

// ==============================================
// TC-SM-005: システム実行テスト
// ==============================================

// TC-SM-005-01: Update実行テスト
func TestSystemManager_UpdateSystems_Success(t *testing.T) {
	// Given: 複数の登録済みシステムがあるSystemManager
	sm := NewSystemManager()
	world := NewMockWorld()
	
	system1 := NewMockSystem(SystemType("System1"))
	system2 := NewMockSystem(SystemType("System2"))
	
	sm.RegisterSystem(system1)
	sm.RegisterSystem(system2)

	// When: UpdateSystemsを呼び出す
	err := sm.UpdateSystems(world, 0.016)

	// Then: 全システムのUpdateが実行順序通りに呼び出される
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	update1, _, _, _ := system1.GetCallCounts()
	update2, _, _, _ := system2.GetCallCounts()

	if update1 != 1 {
		t.Errorf("Expected system1.Update to be called once, got %d", update1)
	}
	if update2 != 1 {
		t.Errorf("Expected system2.Update to be called once, got %d", update2)
	}
}

// ==============================================
// Helper Function: SystemManager Constructor
// ==============================================

// NewSystemManager creates a new SystemManager instance
// NOTE: This will fail until we implement SystemManagerImpl
func NewSystemManager() SystemManager {
	// This will be implemented in the Green phase
	panic("SystemManager not implemented yet - this is expected in Red phase")
}
```

### 2. 基本的なSystemManager構造体の空実装

#### ファイル: `internal/core/ecs/system_manager.go`

```go
package ecs

import (
	"sync"
	"time"
)

// ==============================================
// SystemManager Implementation
// ==============================================

// SystemManagerImpl implements the SystemManager interface
type SystemManagerImpl struct {
	// System storage and state
	systems      map[SystemType]System
	systemStates map[SystemType]bool // true = enabled, false = disabled
	
	// Dependency management
	dependencies map[SystemType][]SystemType
	dependents   map[SystemType][]SystemType
	
	// Execution control
	executionOrder []SystemType
	parallelGroups [][]SystemType
	
	// Configuration
	parallelExecution  bool
	maxParallelSystems int
	globalTimeout      time.Duration
	
	// Performance monitoring
	metrics          map[SystemType]*SystemMetrics
	profilingEnabled bool
	
	// Error handling
	errorHandler func(SystemType, error) error
	systemErrors map[SystemType][]error
	
	// Thread safety
	mutex sync.RWMutex
}

// NewSystemManager creates a new SystemManager instance
func NewSystemManager() SystemManager {
	return &SystemManagerImpl{
		systems:            make(map[SystemType]System),
		systemStates:       make(map[SystemType]bool),
		dependencies:       make(map[SystemType][]SystemType),
		dependents:         make(map[SystemType][]SystemType),
		executionOrder:     make([]SystemType, 0),
		parallelGroups:     make([][]SystemType, 0),
		parallelExecution:  false,
		maxParallelSystems: 1,
		globalTimeout:      30 * time.Second,
		metrics:            make(map[SystemType]*SystemMetrics),
		profilingEnabled:   false,
		systemErrors:       make(map[SystemType][]error),
	}
}

// ==============================================
// System Registration and Lifecycle - STUB IMPLEMENTATIONS
// ==============================================

func (sm *SystemManagerImpl) RegisterSystem(system System) error {
	// TODO: Implement in Green phase
	panic("RegisterSystem not implemented")
}

func (sm *SystemManagerImpl) RegisterSystemWithPriority(system System, priority Priority) error {
	// TODO: Implement in Green phase
	panic("RegisterSystemWithPriority not implemented")
}

func (sm *SystemManagerImpl) UnregisterSystem(systemType SystemType) error {
	// TODO: Implement in Green phase
	panic("UnregisterSystem not implemented")
}

func (sm *SystemManagerImpl) GetSystem(systemType SystemType) (System, error) {
	// TODO: Implement in Green phase
	panic("GetSystem not implemented")
}

func (sm *SystemManagerImpl) GetAllSystems() []System {
	// TODO: Implement in Green phase
	panic("GetAllSystems not implemented")
}

func (sm *SystemManagerImpl) GetSystemCount() int {
	// TODO: Implement in Green phase
	panic("GetSystemCount not implemented")
}

// ==============================================
// System Execution Control - STUB IMPLEMENTATIONS
// ==============================================

func (sm *SystemManagerImpl) UpdateSystems(world World, deltaTime float64) error {
	// TODO: Implement in Green phase
	panic("UpdateSystems not implemented")
}

func (sm *SystemManagerImpl) RenderSystems(world World, renderer interface{}) error {
	// TODO: Implement in Green phase
	panic("RenderSystems not implemented")
}

func (sm *SystemManagerImpl) InitializeSystems(world World) error {
	// TODO: Implement in Green phase
	panic("InitializeSystems not implemented")
}

func (sm *SystemManagerImpl) ShutdownSystems() error {
	// TODO: Implement in Green phase
	panic("ShutdownSystems not implemented")
}

// ==============================================
// System State Management - STUB IMPLEMENTATIONS
// ==============================================

func (sm *SystemManagerImpl) EnableSystem(systemType SystemType) error {
	// TODO: Implement in Green phase
	panic("EnableSystem not implemented")
}

func (sm *SystemManagerImpl) DisableSystem(systemType SystemType) error {
	// TODO: Implement in Green phase
	panic("DisableSystem not implemented")
}

func (sm *SystemManagerImpl) IsSystemEnabled(systemType SystemType) bool {
	// TODO: Implement in Green phase
	panic("IsSystemEnabled not implemented")
}

func (sm *SystemManagerImpl) GetEnabledSystems() []SystemType {
	// TODO: Implement in Green phase
	panic("GetEnabledSystems not implemented")
}

func (sm *SystemManagerImpl) GetDisabledSystems() []SystemType {
	// TODO: Implement in Green phase
	panic("GetDisabledSystems not implemented")
}

// ==============================================
// Dependency Management - STUB IMPLEMENTATIONS
// ==============================================

func (sm *SystemManagerImpl) SetSystemDependency(dependent SystemType, dependency SystemType) error {
	// TODO: Implement in Green phase
	panic("SetSystemDependency not implemented")
}

func (sm *SystemManagerImpl) RemoveSystemDependency(dependent SystemType, dependency SystemType) error {
	// TODO: Implement in Green phase
	panic("RemoveSystemDependency not implemented")
}

func (sm *SystemManagerImpl) GetSystemDependencies(systemType SystemType) []SystemType {
	// TODO: Implement in Green phase
	panic("GetSystemDependencies not implemented")
}

func (sm *SystemManagerImpl) GetSystemDependents(systemType SystemType) []SystemType {
	// TODO: Implement in Green phase
	panic("GetSystemDependents not implemented")
}

func (sm *SystemManagerImpl) GetExecutionOrder() []SystemType {
	// TODO: Implement in Green phase
	panic("GetExecutionOrder not implemented")
}

func (sm *SystemManagerImpl) ValidateExecutionOrder() error {
	// TODO: Implement in Green phase
	panic("ValidateExecutionOrder not implemented")
}

func (sm *SystemManagerImpl) RecomputeExecutionOrder() error {
	// TODO: Implement in Green phase
	panic("RecomputeExecutionOrder not implemented")
}

// ==============================================
// System Filtering and Queries - STUB IMPLEMENTATIONS
// ==============================================

func (sm *SystemManagerImpl) GetSystemsByComponent(componentType ComponentType) []System {
	// TODO: Implement in Green phase
	panic("GetSystemsByComponent not implemented")
}

func (sm *SystemManagerImpl) GetSystemsByPriority(priority Priority) []System {
	// TODO: Implement in Green phase
	panic("GetSystemsByPriority not implemented")
}

func (sm *SystemManagerImpl) GetSystemsByThreadSafety(threadSafety ThreadSafetyLevel) []System {
	// TODO: Implement in Green phase
	panic("GetSystemsByThreadSafety not implemented")
}

func (sm *SystemManagerImpl) FindSystemsByPredicate(predicate func(System) bool) []System {
	// TODO: Implement in Green phase
	panic("FindSystemsByPredicate not implemented")
}

// ==============================================
// Parallel Execution Management - STUB IMPLEMENTATIONS
// ==============================================

func (sm *SystemManagerImpl) SetParallelExecution(enabled bool) {
	// TODO: Implement in Green phase
	panic("SetParallelExecution not implemented")
}

func (sm *SystemManagerImpl) IsParallelExecutionEnabled() bool {
	// TODO: Implement in Green phase
	panic("IsParallelExecutionEnabled not implemented")
}

func (sm *SystemManagerImpl) GetParallelGroups() [][]SystemType {
	// TODO: Implement in Green phase
	panic("GetParallelGroups not implemented")
}

func (sm *SystemManagerImpl) SetMaxParallelSystems(max int) {
	// TODO: Implement in Green phase
	panic("SetMaxParallelSystems not implemented")
}

func (sm *SystemManagerImpl) GetMaxParallelSystems() int {
	// TODO: Implement in Green phase
	panic("GetMaxParallelSystems not implemented")
}

// ==============================================
// Performance Monitoring - STUB IMPLEMENTATIONS
// ==============================================

func (sm *SystemManagerImpl) GetSystemMetrics(systemType SystemType) (*SystemMetrics, error) {
	// TODO: Implement in Green phase
	panic("GetSystemMetrics not implemented")
}

func (sm *SystemManagerImpl) GetAllSystemMetrics() map[SystemType]*SystemMetrics {
	// TODO: Implement in Green phase
	panic("GetAllSystemMetrics not implemented")
}

func (sm *SystemManagerImpl) ResetSystemMetrics() error {
	// TODO: Implement in Green phase
	panic("ResetSystemMetrics not implemented")
}

func (sm *SystemManagerImpl) EnableProfiling(enabled bool) {
	// TODO: Implement in Green phase
	panic("EnableProfiling not implemented")
}

func (sm *SystemManagerImpl) IsProfilingEnabled() bool {
	// TODO: Implement in Green phase
	panic("IsProfilingEnabled not implemented")
}

// ==============================================
// Additional Interface Methods - STUB IMPLEMENTATIONS
// ==============================================

func (sm *SystemManagerImpl) RegisterSystems(systems []System) error {
	panic("RegisterSystems not implemented")
}

func (sm *SystemManagerImpl) UnregisterSystems(systemTypes []SystemType) error {
	panic("UnregisterSystems not implemented")
}

func (sm *SystemManagerImpl) EnableSystems(systemTypes []SystemType) error {
	panic("EnableSystems not implemented")
}

func (sm *SystemManagerImpl) DisableSystems(systemTypes []SystemType) error {
	panic("DisableSystems not implemented")
}

func (sm *SystemManagerImpl) SetErrorHandler(handler func(SystemType, error) error) {
	panic("SetErrorHandler not implemented")
}

func (sm *SystemManagerImpl) GetSystemErrors() map[SystemType][]error {
	panic("GetSystemErrors not implemented")
}

func (sm *SystemManagerImpl) ClearSystemErrors(systemType SystemType) error {
	panic("ClearSystemErrors not implemented")
}

func (sm *SystemManagerImpl) GetFailedSystems() []SystemType {
	panic("GetFailedSystems not implemented")
}

func (sm *SystemManagerImpl) SetSystemTimeout(systemType SystemType, timeout time.Duration) error {
	panic("SetSystemTimeout not implemented")
}

func (sm *SystemManagerImpl) GetSystemTimeout(systemType SystemType) time.Duration {
	panic("GetSystemTimeout not implemented")
}

func (sm *SystemManagerImpl) SetGlobalTimeout(timeout time.Duration) {
	panic("SetGlobalTimeout not implemented")
}

func (sm *SystemManagerImpl) GetGlobalTimeout() time.Duration {
	panic("GetGlobalTimeout not implemented")
}

func (sm *SystemManagerImpl) SerializeSystemState() ([]byte, error) {
	panic("SerializeSystemState not implemented")
}

func (sm *SystemManagerImpl) DeserializeSystemState(data []byte) error {
	panic("DeserializeSystemState not implemented")
}

func (sm *SystemManagerImpl) SaveSystemConfiguration(path string) error {
	panic("SaveSystemConfiguration not implemented")
}

func (sm *SystemManagerImpl) LoadSystemConfiguration(path string) error {
	panic("LoadSystemConfiguration not implemented")
}

func (sm *SystemManagerImpl) Lock() {
	panic("Lock not implemented")
}

func (sm *SystemManagerImpl) RLock() {
	panic("RLock not implemented")
}

func (sm *SystemManagerImpl) Unlock() {
	panic("Unlock not implemented")
}

func (sm *SystemManagerImpl) RUnlock() {
	panic("RUnlock not implemented")
}

func (sm *SystemManagerImpl) ValidateIntegrity() error {
	panic("ValidateIntegrity not implemented")
}

func (sm *SystemManagerImpl) GetDebugInfo() *SystemManagerDebugInfo {
	panic("GetDebugInfo not implemented")
}

func (sm *SystemManagerImpl) DumpExecutionOrder() string {
	panic("DumpExecutionOrder not implemented")
}

func (sm *SystemManagerImpl) GetDependencyGraph() *DependencyGraph {
	panic("GetDependencyGraph not implemented")
}
```

## Red段階テスト実行

上記のファイルを作成後、テストを実行して**すべてのテストが失敗する**ことを確認します。

```bash
cd /home/devman/GolandProjects/muscle-dreamer
go test ./internal/core/ecs -v -run "TestSystemManager"
```

### 期待される結果

- **全テストが失敗**: `panic` によりテストが失敗する
- **テスト構造の妥当性確認**: テストケースが期待通りに実行される
- **Mock実装の動作確認**: MockSystemとMockWorldが正常に動作する

## 実装ファイル作成アクション

### 1. テストファイル作成
```go
// internal/core/ecs/system_manager_test.go
// 上記のテスト実装を配置
```

### 2. SystemManager実装ファイル作成
```go
// internal/core/ecs/system_manager.go
// 上記のスタブ実装を配置
```

### 3. テスト実行・確認
```bash
# テストが期待通り失敗することを確認
go test ./internal/core/ecs -v -run "TestSystemManager"
```

## Red段階の確認ポイント

### ✅ 成功条件
1. **全テストが失敗**: すべてのテストケースが`panic`で失敗する
2. **テストカバレッジ**: 主要機能（登録、状態管理、依存関係、実行）をカバー
3. **Mock実装**: テスト用のMockSystemとMockWorldが正常動作
4. **エラーメッセージ**: 明確で一貫したpanicメッセージ

### ❌ 失敗条件
1. **テストが成功**: 実装前にテストが通る（テスト設計エラー）
2. **コンパイルエラー**: 型定義やインターフェース不整合
3. **実行時エラー**: panic以外の予期しないエラー

## 次のステップ

Red段階完了後、**Green段階**で最小限の実装を行い、テストを通すようにします。

---

**実装優先順位**:
1. **Phase 1**: システム登録・状態管理の基本実装
2. **Phase 2**: 依存関係管理の基礎実装  
3. **Phase 3**: システム実行制御の実装
4. **Phase 4**: 並列実行・パフォーマンス監視
5. **Phase 5**: 高度機能・最適化