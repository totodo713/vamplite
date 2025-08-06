# TASK-103: SystemManager実装 - Green段階 (最小実装)

## Green段階の目的

TDDのGreen段階では、**テストを通すための最小限の実装**を行います。過度な最適化や機能追加は避け、単純で動作する実装を目指します。

## 実装方針

1. **シンプルな実装**: マップベースの単純な管理
2. **同期的実行**: 並列実行は後の段階で実装
3. **基本的なエラーハンドリング**: nilチェックと重複登録チェック
4. **依存関係の基本実装**: 循環依存検出の最小実装

## 実装コード

### SystemManager実装の更新

#### ファイル: `internal/core/ecs/system_manager.go`

```go
package ecs

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ==============================================
// Error Definitions
// ==============================================

var (
	ErrSystemNotFound      = errors.New("system not found")
	ErrSystemAlreadyExists = errors.New("system already exists")
	ErrNilSystem           = errors.New("cannot register nil system")
	ErrCyclicDependency    = errors.New("cyclic dependency detected")
	ErrInvalidDependency   = errors.New("invalid dependency")
)

// ==============================================
// SystemManager Implementation
// ==============================================

// SystemManagerImpl implements the SystemManager interface
type SystemManagerImpl struct {
	// System storage and state
	systems      map[SystemType]System
	systemStates map[SystemType]bool // true = enabled, false = disabled
	
	// Priority management
	systemPriorities map[SystemType]Priority
	
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
		systemPriorities:   make(map[SystemType]Priority),
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
// System Registration and Lifecycle
// ==============================================

func (sm *SystemManagerImpl) RegisterSystem(system System) error {
	if system == nil {
		return ErrNilSystem
	}
	
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	systemType := system.GetType()
	
	// Check if system already exists
	if _, exists := sm.systems[systemType]; exists {
		return ErrSystemAlreadyExists
	}
	
	// Register the system
	sm.systems[systemType] = system
	sm.systemStates[systemType] = true // Enabled by default
	sm.systemPriorities[systemType] = system.GetPriority()
	
	// Initialize empty dependencies
	sm.dependencies[systemType] = []SystemType{}
	sm.dependents[systemType] = []SystemType{}
	
	// Update execution order
	sm.executionOrder = append(sm.executionOrder, systemType)
	
	// Initialize metrics
	sm.metrics[systemType] = &SystemMetrics{
		SystemType: systemType,
	}
	
	return nil
}

func (sm *SystemManagerImpl) RegisterSystemWithPriority(system System, priority Priority) error {
	if system == nil {
		return ErrNilSystem
	}
	
	// Register system normally first
	if err := sm.RegisterSystem(system); err != nil {
		return err
	}
	
	// Override priority
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sm.systemPriorities[system.GetType()] = priority
	
	return nil
}

func (sm *SystemManagerImpl) UnregisterSystem(systemType SystemType) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	// Check if system exists
	if _, exists := sm.systems[systemType]; !exists {
		return ErrSystemNotFound
	}
	
	// Remove from all data structures
	delete(sm.systems, systemType)
	delete(sm.systemStates, systemType)
	delete(sm.systemPriorities, systemType)
	delete(sm.dependencies, systemType)
	delete(sm.dependents, systemType)
	delete(sm.metrics, systemType)
	delete(sm.systemErrors, systemType)
	
	// Remove from execution order
	newOrder := make([]SystemType, 0, len(sm.executionOrder)-1)
	for _, st := range sm.executionOrder {
		if st != systemType {
			newOrder = append(newOrder, st)
		}
	}
	sm.executionOrder = newOrder
	
	// Remove from dependencies of other systems
	for st, deps := range sm.dependencies {
		newDeps := make([]SystemType, 0)
		for _, dep := range deps {
			if dep != systemType {
				newDeps = append(newDeps, dep)
			}
		}
		sm.dependencies[st] = newDeps
	}
	
	// Remove from dependents of other systems
	for st, deps := range sm.dependents {
		newDeps := make([]SystemType, 0)
		for _, dep := range deps {
			if dep != systemType {
				newDeps = append(newDeps, dep)
			}
		}
		sm.dependents[st] = newDeps
	}
	
	return nil
}

func (sm *SystemManagerImpl) GetSystem(systemType SystemType) (System, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	system, exists := sm.systems[systemType]
	if !exists {
		return nil, ErrSystemNotFound
	}
	
	return system, nil
}

func (sm *SystemManagerImpl) GetAllSystems() []System {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	systems := make([]System, 0, len(sm.systems))
	for _, system := range sm.systems {
		systems = append(systems, system)
	}
	
	return systems
}

func (sm *SystemManagerImpl) GetSystemCount() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	return len(sm.systems)
}

// ==============================================
// System State Management
// ==============================================

func (sm *SystemManagerImpl) EnableSystem(systemType SystemType) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	if _, exists := sm.systems[systemType]; !exists {
		return ErrSystemNotFound
	}
	
	sm.systemStates[systemType] = true
	return nil
}

func (sm *SystemManagerImpl) DisableSystem(systemType SystemType) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	if _, exists := sm.systems[systemType]; !exists {
		return ErrSystemNotFound
	}
	
	sm.systemStates[systemType] = false
	return nil
}

func (sm *SystemManagerImpl) IsSystemEnabled(systemType SystemType) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	enabled, exists := sm.systemStates[systemType]
	if !exists {
		return false
	}
	
	return enabled
}

func (sm *SystemManagerImpl) GetEnabledSystems() []SystemType {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	enabled := make([]SystemType, 0)
	for systemType, isEnabled := range sm.systemStates {
		if isEnabled {
			enabled = append(enabled, systemType)
		}
	}
	
	return enabled
}

func (sm *SystemManagerImpl) GetDisabledSystems() []SystemType {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	disabled := make([]SystemType, 0)
	for systemType, isEnabled := range sm.systemStates {
		if !isEnabled {
			disabled = append(disabled, systemType)
		}
	}
	
	return disabled
}

// ==============================================
// Dependency Management
// ==============================================

func (sm *SystemManagerImpl) SetSystemDependency(dependent SystemType, dependency SystemType) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	// Check both systems exist
	if _, exists := sm.systems[dependent]; !exists {
		return fmt.Errorf("%w: dependent system %s not found", ErrInvalidDependency, dependent)
	}
	if _, exists := sm.systems[dependency]; !exists {
		return fmt.Errorf("%w: dependency system %s not found", ErrInvalidDependency, dependency)
	}
	
	// Check for cyclic dependency
	if sm.wouldCreateCycle(dependent, dependency) {
		return ErrCyclicDependency
	}
	
	// Add dependency
	sm.dependencies[dependent] = append(sm.dependencies[dependent], dependency)
	sm.dependents[dependency] = append(sm.dependents[dependency], dependent)
	
	return nil
}

func (sm *SystemManagerImpl) RemoveSystemDependency(dependent SystemType, dependency SystemType) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	// Remove from dependencies
	if deps, exists := sm.dependencies[dependent]; exists {
		newDeps := make([]SystemType, 0)
		for _, dep := range deps {
			if dep != dependency {
				newDeps = append(newDeps, dep)
			}
		}
		sm.dependencies[dependent] = newDeps
	}
	
	// Remove from dependents
	if deps, exists := sm.dependents[dependency]; exists {
		newDeps := make([]SystemType, 0)
		for _, dep := range deps {
			if dep != dependent {
				newDeps = append(newDeps, dep)
			}
		}
		sm.dependents[dependency] = newDeps
	}
	
	return nil
}

func (sm *SystemManagerImpl) GetSystemDependencies(systemType SystemType) []SystemType {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	deps := sm.dependencies[systemType]
	result := make([]SystemType, len(deps))
	copy(result, deps)
	return result
}

func (sm *SystemManagerImpl) GetSystemDependents(systemType SystemType) []SystemType {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	deps := sm.dependents[systemType]
	result := make([]SystemType, len(deps))
	copy(result, deps)
	return result
}

// wouldCreateCycle checks if adding a dependency would create a cycle
func (sm *SystemManagerImpl) wouldCreateCycle(dependent, dependency SystemType) bool {
	// Simple DFS to detect cycle
	visited := make(map[SystemType]bool)
	return sm.hasCycleDFS(dependency, dependent, visited)
}

func (sm *SystemManagerImpl) hasCycleDFS(current, target SystemType, visited map[SystemType]bool) bool {
	if current == target {
		return true
	}
	
	if visited[current] {
		return false
	}
	
	visited[current] = true
	
	for _, dep := range sm.dependencies[current] {
		if sm.hasCycleDFS(dep, target, visited) {
			return true
		}
	}
	
	return false
}

// ==============================================
// System Execution
// ==============================================

func (sm *SystemManagerImpl) UpdateSystems(world World, deltaTime float64) error {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	// Execute systems in order
	for _, systemType := range sm.executionOrder {
		// Skip disabled systems
		if !sm.systemStates[systemType] {
			continue
		}
		
		system := sm.systems[systemType]
		if system == nil {
			continue
		}
		
		// Execute Update
		if err := system.Update(world, deltaTime); err != nil {
			// Store error but continue execution
			sm.systemErrors[systemType] = append(sm.systemErrors[systemType], err)
			
			// Call error handler if set
			if sm.errorHandler != nil {
				if handlerErr := sm.errorHandler(systemType, err); handlerErr != nil {
					return handlerErr
				}
			}
		}
	}
	
	return nil
}

func (sm *SystemManagerImpl) RenderSystems(world World, renderer interface{}) error {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	// Execute systems in order
	for _, systemType := range sm.executionOrder {
		// Skip disabled systems
		if !sm.systemStates[systemType] {
			continue
		}
		
		system := sm.systems[systemType]
		if system == nil {
			continue
		}
		
		// Execute Render
		if err := system.Render(world, renderer); err != nil {
			// Store error but continue execution
			sm.systemErrors[systemType] = append(sm.systemErrors[systemType], err)
			
			// Call error handler if set
			if sm.errorHandler != nil {
				if handlerErr := sm.errorHandler(systemType, err); handlerErr != nil {
					return handlerErr
				}
			}
		}
	}
	
	return nil
}

func (sm *SystemManagerImpl) InitializeSystems(world World) error {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	for _, systemType := range sm.executionOrder {
		system := sm.systems[systemType]
		if system == nil {
			continue
		}
		
		if err := system.Initialize(world); err != nil {
			return fmt.Errorf("failed to initialize system %s: %w", systemType, err)
		}
	}
	
	return nil
}

func (sm *SystemManagerImpl) ShutdownSystems() error {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	// Shutdown in reverse order
	for i := len(sm.executionOrder) - 1; i >= 0; i-- {
		systemType := sm.executionOrder[i]
		system := sm.systems[systemType]
		if system == nil {
			continue
		}
		
		if err := system.Shutdown(); err != nil {
			// Log error but continue shutdown
			sm.systemErrors[systemType] = append(sm.systemErrors[systemType], err)
		}
	}
	
	return nil
}

// ==============================================
// System Filtering and Queries
// ==============================================

func (sm *SystemManagerImpl) GetSystemsByPriority(priority Priority) []System {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	systems := make([]System, 0)
	for systemType, systemPriority := range sm.systemPriorities {
		if systemPriority == priority {
			if system, exists := sm.systems[systemType]; exists {
				systems = append(systems, system)
			}
		}
	}
	
	return systems
}

func (sm *SystemManagerImpl) GetSystemsByComponent(componentType ComponentType) []System {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	systems := make([]System, 0)
	for _, system := range sm.systems {
		requiredComponents := system.GetRequiredComponents()
		for _, required := range requiredComponents {
			if required == componentType {
				systems = append(systems, system)
				break
			}
		}
	}
	
	return systems
}

func (sm *SystemManagerImpl) GetSystemsByThreadSafety(threadSafety ThreadSafetyLevel) []System {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	systems := make([]System, 0)
	for _, system := range sm.systems {
		if system.GetThreadSafety() == threadSafety {
			systems = append(systems, system)
		}
	}
	
	return systems
}

func (sm *SystemManagerImpl) FindSystemsByPredicate(predicate func(System) bool) []System {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	systems := make([]System, 0)
	for _, system := range sm.systems {
		if predicate(system) {
			systems = append(systems, system)
		}
	}
	
	return systems
}

// ==============================================
// Additional Required Methods - Basic Implementation
// ==============================================

func (sm *SystemManagerImpl) GetExecutionOrder() []SystemType {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	result := make([]SystemType, len(sm.executionOrder))
	copy(result, sm.executionOrder)
	return result
}

func (sm *SystemManagerImpl) ValidateExecutionOrder() error {
	// Basic validation - check all systems are in order
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	for _, systemType := range sm.executionOrder {
		if _, exists := sm.systems[systemType]; !exists {
			return fmt.Errorf("system %s in execution order but not registered", systemType)
		}
	}
	
	return nil
}

func (sm *SystemManagerImpl) RecomputeExecutionOrder() error {
	// Simple implementation - no dependency ordering yet
	return nil
}

func (sm *SystemManagerImpl) SetParallelExecution(enabled bool) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.parallelExecution = enabled
}

func (sm *SystemManagerImpl) IsParallelExecutionEnabled() bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.parallelExecution
}

func (sm *SystemManagerImpl) GetParallelGroups() [][]SystemType {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	result := make([][]SystemType, len(sm.parallelGroups))
	for i, group := range sm.parallelGroups {
		result[i] = make([]SystemType, len(group))
		copy(result[i], group)
	}
	return result
}

func (sm *SystemManagerImpl) SetMaxParallelSystems(max int) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.maxParallelSystems = max
}

func (sm *SystemManagerImpl) GetMaxParallelSystems() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.maxParallelSystems
}

func (sm *SystemManagerImpl) GetSystemMetrics(systemType SystemType) (*SystemMetrics, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	metrics, exists := sm.metrics[systemType]
	if !exists {
		return nil, ErrSystemNotFound
	}
	
	return metrics, nil
}

func (sm *SystemManagerImpl) GetAllSystemMetrics() map[SystemType]*SystemMetrics {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	result := make(map[SystemType]*SystemMetrics)
	for k, v := range sm.metrics {
		result[k] = v
	}
	return result
}

func (sm *SystemManagerImpl) ResetSystemMetrics() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	for systemType := range sm.metrics {
		sm.metrics[systemType] = &SystemMetrics{
			SystemType: systemType,
		}
	}
	
	return nil
}

func (sm *SystemManagerImpl) EnableProfiling(enabled bool) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.profilingEnabled = enabled
}

func (sm *SystemManagerImpl) IsProfilingEnabled() bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.profilingEnabled
}

// ==============================================
// Stub implementations for remaining methods
// ==============================================

func (sm *SystemManagerImpl) RegisterSystems(systems []System) error {
	for _, system := range systems {
		if err := sm.RegisterSystem(system); err != nil {
			return err
		}
	}
	return nil
}

func (sm *SystemManagerImpl) UnregisterSystems(systemTypes []SystemType) error {
	for _, systemType := range systemTypes {
		if err := sm.UnregisterSystem(systemType); err != nil {
			return err
		}
	}
	return nil
}

func (sm *SystemManagerImpl) EnableSystems(systemTypes []SystemType) error {
	for _, systemType := range systemTypes {
		if err := sm.EnableSystem(systemType); err != nil {
			return err
		}
	}
	return nil
}

func (sm *SystemManagerImpl) DisableSystems(systemTypes []SystemType) error {
	for _, systemType := range systemTypes {
		if err := sm.DisableSystem(systemType); err != nil {
			return err
		}
	}
	return nil
}

func (sm *SystemManagerImpl) SetErrorHandler(handler func(SystemType, error) error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.errorHandler = handler
}

func (sm *SystemManagerImpl) GetSystemErrors() map[SystemType][]error {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	result := make(map[SystemType][]error)
	for k, v := range sm.systemErrors {
		errors := make([]error, len(v))
		copy(errors, v)
		result[k] = errors
	}
	return result
}

func (sm *SystemManagerImpl) ClearSystemErrors(systemType SystemType) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	delete(sm.systemErrors, systemType)
	return nil
}

func (sm *SystemManagerImpl) GetFailedSystems() []SystemType {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	failed := make([]SystemType, 0)
	for systemType, errors := range sm.systemErrors {
		if len(errors) > 0 {
			failed = append(failed, systemType)
		}
	}
	return failed
}

func (sm *SystemManagerImpl) SetSystemTimeout(systemType SystemType, timeout time.Duration) error {
	// TODO: Implement in refactor phase
	return nil
}

func (sm *SystemManagerImpl) GetSystemTimeout(systemType SystemType) time.Duration {
	// TODO: Implement in refactor phase
	return sm.globalTimeout
}

func (sm *SystemManagerImpl) SetGlobalTimeout(timeout time.Duration) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.globalTimeout = timeout
}

func (sm *SystemManagerImpl) GetGlobalTimeout() time.Duration {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.globalTimeout
}

func (sm *SystemManagerImpl) SerializeSystemState() ([]byte, error) {
	// TODO: Implement in refactor phase
	return nil, nil
}

func (sm *SystemManagerImpl) DeserializeSystemState(data []byte) error {
	// TODO: Implement in refactor phase
	return nil
}

func (sm *SystemManagerImpl) SaveSystemConfiguration(path string) error {
	// TODO: Implement in refactor phase
	return nil
}

func (sm *SystemManagerImpl) LoadSystemConfiguration(path string) error {
	// TODO: Implement in refactor phase
	return nil
}

func (sm *SystemManagerImpl) Lock() {
	sm.mutex.Lock()
}

func (sm *SystemManagerImpl) RLock() {
	sm.mutex.RLock()
}

func (sm *SystemManagerImpl) Unlock() {
	sm.mutex.Unlock()
}

func (sm *SystemManagerImpl) RUnlock() {
	sm.mutex.RUnlock()
}

func (sm *SystemManagerImpl) ValidateIntegrity() error {
	return sm.ValidateExecutionOrder()
}

func (sm *SystemManagerImpl) GetDebugInfo() *SystemManagerDebugInfo {
	// TODO: Implement in refactor phase
	return nil
}

func (sm *SystemManagerImpl) DumpExecutionOrder() string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	result := "Execution Order:\n"
	for i, systemType := range sm.executionOrder {
		enabled := "disabled"
		if sm.systemStates[systemType] {
			enabled = "enabled"
		}
		result += fmt.Sprintf("  %d. %s (%s)\n", i+1, systemType, enabled)
	}
	return result
}

func (sm *SystemManagerImpl) GetDependencyGraph() *DependencyGraph {
	// TODO: Implement in refactor phase
	return nil
}
```

## テスト実行と検証

Green段階の実装完了後、テストを実行して通ることを確認します。

```bash
cd /home/devman/GolandProjects/muscle-dreamer
go test ./internal/core/ecs -v -run "TestSystemManager"
```

## Green段階の確認ポイント

### ✅ 成功条件

1. **全テストが通る**: 基本機能のテストがすべてパス
2. **最小限の実装**: 過度な最適化を避けた単純な実装
3. **データ整合性**: 内部データ構造の一貫性維持
4. **スレッドセーフ**: 基本的な排他制御の実装

### 実装された機能

1. **システム登録・削除**
   - 単一システムの登録・削除
   - 優先度付き登録
   - 重複チェック、nilチェック

2. **システム状態管理**
   - 有効化・無効化
   - 状態クエリ

3. **依存関係管理**
   - 依存関係の設定・削除
   - 循環依存検出（DFSベース）
   - 依存・被依存リスト取得

4. **システム実行**
   - 順次Update/Render実行
   - エラーハンドリング
   - Initialize/Shutdown

5. **システムクエリ**
   - 優先度別フィルタリング
   - コンポーネント別フィルタリング
   - カスタム述語によるフィルタリング

## 次のステップ

Green段階完了後、**Refactor段階**で以下の改善を行います：

1. **パフォーマンス最適化**
   - 並列実行の実装
   - キャッシュ機構の追加

2. **依存関係の高度な管理**
   - トポロジカルソートによる実行順序決定
   - 並列実行グループの自動生成

3. **メトリクス収集**
   - 実行時間計測
   - メモリ使用量追跡

4. **エラーハンドリングの強化**
   - タイムアウト処理
   - リトライ機構

5. **永続化機能**
   - 設定の保存・読み込み
   - 状態のシリアライズ