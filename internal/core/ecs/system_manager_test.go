package ecs

import (
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
	return ThreadSafetyWrite
}

func (s *MockSystem) CanRunInParallel() bool {
	return true
}

func (s *MockSystem) IsEnabled() bool {
	return true
}

func (s *MockSystem) SetEnabled(_ bool) {
	// Mock implementation
}

func (s *MockSystem) GetMetrics() *SystemMetrics {
	return &SystemMetrics{}
}

func (s *MockSystem) Update(_ World, _ float64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	time.Sleep(s.execTime) // Simulate work
	s.updateCalled++

	if s.shouldError {
		return errors.New("mock system error")
	}
	return nil
}

func (s *MockSystem) Render(_ World, _ interface{}) error {
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

// Minimal implementation of World interface for testing
func (w *MockWorld) CreateEntity() EntityID                        { return EntityID(1) }
func (w *MockWorld) DestroyEntity(EntityID) error                  { return nil }
func (w *MockWorld) IsEntityValid(EntityID) bool                   { return true }
func (w *MockWorld) GetEntityCount() int                           { return 0 }
func (w *MockWorld) GetActiveEntities() []EntityID                 { return []EntityID{} }
func (w *MockWorld) AddComponent(EntityID, Component) error        { return nil }
func (w *MockWorld) RemoveComponent(EntityID, ComponentType) error { return nil }
func (w *MockWorld) GetComponent(EntityID, ComponentType) (Component, error) {
	return nil, errors.New(ErrComponentNotFound)
}
func (w *MockWorld) HasComponent(EntityID, ComponentType) bool        { return false }
func (w *MockWorld) GetComponents(EntityID) []Component               { return []Component{} }
func (w *MockWorld) RegisterSystem(System) error                      { return nil }
func (w *MockWorld) UnregisterSystem(SystemType) error                { return nil }
func (w *MockWorld) GetSystem(SystemType) (System, error)             { return nil, errors.New(ErrSystemNotFound) }
func (w *MockWorld) GetAllSystems() []System                          { return []System{} }
func (w *MockWorld) EnableSystem(SystemType) error                    { return nil }
func (w *MockWorld) DisableSystem(SystemType) error                   { return nil }
func (w *MockWorld) IsSystemEnabled(SystemType) bool                  { return true }
func (w *MockWorld) Update(float64) error                             { return nil }
func (w *MockWorld) Render(interface{}) error                         { return nil }
func (w *MockWorld) Shutdown() error                                  { return nil }
func (w *MockWorld) GetMetrics() *PerformanceMetrics                  { return nil }
func (w *MockWorld) GetMemoryUsage() *MemoryUsage                     { return nil }
func (w *MockWorld) GetStorageStats() []StorageStats                  { return []StorageStats{} }
func (w *MockWorld) GetQueryStats() []QueryStats                      { return []QueryStats{} }
func (w *MockWorld) GetConfig() *WorldConfig                          { return nil }
func (w *MockWorld) UpdateConfig(*WorldConfig) error                  { return nil }
func (w *MockWorld) EmitEvent(Event) error                            { return nil }
func (w *MockWorld) Subscribe(EventType, EventHandler) error          { return nil }
func (w *MockWorld) Unsubscribe(EventType, EventHandler) error        { return nil }
func (w *MockWorld) Query() QueryBuilder                              { return nil }
func (w *MockWorld) CreateQuery(QueryBuilder) QueryResult             { return nil }
func (w *MockWorld) ExecuteQuery(QueryBuilder) QueryResult            { return nil }
func (w *MockWorld) CreateEntities(int) []EntityID                    { return []EntityID{} }
func (w *MockWorld) DestroyEntities([]EntityID) error                 { return nil }
func (w *MockWorld) AddComponents(EntityID, []Component) error        { return nil }
func (w *MockWorld) RemoveComponents(EntityID, []ComponentType) error { return nil }
func (w *MockWorld) SerializeEntity(EntityID) ([]byte, error)         { return nil, nil }
func (w *MockWorld) DeserializeEntity([]byte) (EntityID, error)       { return EntityID(0), nil }
func (w *MockWorld) SerializeWorld() ([]byte, error)                  { return nil, nil }
func (w *MockWorld) DeserializeWorld([]byte) error                    { return nil }
func (w *MockWorld) Lock()                                            {}
func (w *MockWorld) RLock()                                           {}
func (w *MockWorld) Unlock()                                          {}
func (w *MockWorld) RUnlock()                                         {}

// ==============================================
// TC-SM-001: システム登録機能テスト
// ==============================================

// TC-SM-001-01: 正常なシステム登録
func TestSystemManager_RegisterSystem_Success(t *testing.T) {
	// Given: 新しいSystemManagerとMockSystem
	sm := NewSystemManager()
	system := NewMockSystem(SystemTypeFromString("TestSystem"))

	// When: システムを登録する
	err := sm.RegisterSystem(system)
	// Then: システムが正常に登録され、GetSystemで取得できる
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	retrievedSystem, err := sm.GetSystem(SystemTypeFromString("TestSystem"))
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
	system1 := NewMockSystem(SystemTypeFromString("TestSystem"))
	system2 := NewMockSystem(SystemTypeFromString("TestSystem"))

	sm.RegisterSystem(system1)

	// When: 同じ型のシステムを再度登録しようとする
	err := sm.RegisterSystem(system2)

	// Then: エラーが返され、元のシステムが保持される
	if err == nil {
		t.Error("Expected error for duplicate system registration")
	}

	retrievedSystem, _ := sm.GetSystem(SystemTypeFromString("TestSystem"))
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
	systemHigh := NewMockSystem(SystemTypeFromString("HighPrioritySystem"))
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
	system := NewMockSystem(SystemTypeFromString("TestSystem"))
	sm.RegisterSystem(system)

	// When: システムを登録解除する
	err := sm.UnregisterSystem(SystemTypeFromString("TestSystem"))
	// Then: システムが削除され、GetSystemでエラーが返される
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_, err = sm.GetSystem(SystemTypeFromString("TestSystem"))
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
	system := NewMockSystem(SystemTypeFromString("TestSystem"))
	sm.RegisterSystem(system)

	// When: システムを無効化・有効化する
	err1 := sm.DisableSystem(SystemTypeFromString("TestSystem"))
	enabled1 := sm.IsSystemEnabled(SystemTypeFromString("TestSystem"))

	err2 := sm.EnableSystem(SystemTypeFromString("TestSystem"))
	enabled2 := sm.IsSystemEnabled(SystemTypeFromString("TestSystem"))

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
