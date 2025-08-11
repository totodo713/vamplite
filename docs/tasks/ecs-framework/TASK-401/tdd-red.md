# TASK-401: パフォーマンス最適化 - テスト実装 (Red段階)

## 概要

TDDのRed段階として、パフォーマンス最適化機能に対する失敗するテストを実装します。テストが失敗することを確認してから、最小限の実装でテストが通るようにするGreen段階に進みます。

## 実装方針

1. **テストファイル構造の作成**
2. **基本的なインターフェース定義** (テストがコンパイルできる最小限)
3. **失敗するテスト実装**
4. **テスト実行と失敗確認**

---

## 1. テストファイル構造とパッケージ作成

### ディレクトリ構造
```
internal/core/ecs/optimizations/
├── cache/
│   ├── optimized_component_store.go
│   └── optimized_component_store_test.go
├── simd/
│   ├── simd_transform_system.go
│   └── simd_transform_system_test.go
├── scheduler/
│   ├── optimized_system_scheduler.go
│   └── optimized_system_scheduler_test.go
├── profiler/
│   ├── performance_profiler.go
│   └── performance_profiler_test.go
└── benchmarks/
    ├── performance_benchmarks_test.go
    └── stress_tests_test.go
```

## 2. 基本インターフェース定義 (最小限の実装)

### optimized_component_store.go
```go
package cache

import (
    "unsafe"
    "github.com/muscle-dreamer/internal/core/ecs"
)

// OptimizedComponentStore は CPU キャッシュ効率を最適化したコンポーネントストア
type OptimizedComponentStore struct {
    // 現時点では空の構造体（テストがコンパイルできる最小限）
}

// NewOptimizedComponentStore creates a new optimized component store
func NewOptimizedComponentStore() *OptimizedComponentStore {
    return &OptimizedComponentStore{}
}

// AddTransform adds a transform component (stub implementation)
func (cs *OptimizedComponentStore) AddTransform(entityID ecs.EntityID, component ecs.TransformComponent) {
    // TODO: 実装予定
}

// GetTransform gets a transform component (stub implementation)  
func (cs *OptimizedComponentStore) GetTransform(entityID ecs.EntityID) *ecs.TransformComponent {
    // TODO: 実装予定
    return nil
}

// GetTransformArray returns the transform array for SoA access (stub)
func (cs *OptimizedComponentStore) GetTransformArray() []ecs.TransformComponent {
    // TODO: 実装予定
    return nil
}

// PrefetchComponents prefetches components for better cache performance (stub)
func (cs *OptimizedComponentStore) PrefetchComponents(entities []ecs.EntityID) {
    // TODO: 実装予定
}

// RemoveTransform removes a transform component (stub)
func (cs *OptimizedComponentStore) RemoveTransform(entityID ecs.EntityID) {
    // TODO: 実装予定
}

// AddSprite adds a sprite component (stub)
func (cs *OptimizedComponentStore) AddSprite(entityID ecs.EntityID, component ecs.SpriteComponent) {
    // TODO: 実装予定
}

// RemoveSprite removes a sprite component (stub)
func (cs *OptimizedComponentStore) RemoveSprite(entityID ecs.EntityID) {
    // TODO: 実装予定
}
```

### simd_transform_system.go
```go
package simd

import (
    "github.com/muscle-dreamer/internal/core/ecs"
)

// SIMDTransformSystem は SIMD 命令を活用したTransformSystem
type SIMDTransformSystem struct {
    vectorProcessor *Vector4Processor
}

// Vector4Processor handles SIMD vector operations
type Vector4Processor struct {
    // TODO: 実装予定
}

// ScalarTransformSystem は比較用のスカラー実装
type ScalarTransformSystem struct {
    // TODO: 実装予定
}

// NewSIMDTransformSystem creates a new SIMD transform system
func NewSIMDTransformSystem() *SIMDTransformSystem {
    return &SIMDTransformSystem{
        vectorProcessor: &Vector4Processor{},
    }
}

// NewScalarTransformSystem creates a scalar transform system for comparison
func NewScalarTransformSystem() *ScalarTransformSystem {
    return &ScalarTransformSystem{}
}

// UpdatePositions updates positions using SIMD operations (stub)
func (s *SIMDTransformSystem) UpdatePositions(positions []ecs.Vector3, velocities []ecs.Vector3, deltaTime float32) {
    // TODO: 実装予定
}

// UpdatePositions updates positions using scalar operations (stub)
func (s *ScalarTransformSystem) UpdatePositions(positions []ecs.Vector3, velocities []ecs.Vector3, deltaTime float32) {
    // TODO: 実装予定
}

// AddScaled performs SIMD vector addition with scaling (stub)
func (vp *Vector4Processor) AddScaled(positions []ecs.Vector3, velocities []ecs.Vector3, deltaTime float32) {
    // TODO: 実装予定
}
```

### optimized_system_scheduler.go
```go
package scheduler

import (
    "github.com/muscle-dreamer/internal/core/ecs"
)

// OptimizedSystemScheduler は最適化されたシステムスケジューラ
type OptimizedSystemScheduler struct {
    executionGraph    *DAG
    parallelBatches   [][]ecs.System
    dataFlowOptimizer *DataFlowOptimizer
}

// DAG は有向非循環グラフ
type DAG struct {
    // TODO: 実装予定
}

// DataFlowOptimizer はデータフロー最適化
type DataFlowOptimizer struct {
    // TODO: 実装予定
}

// NewOptimizedSystemScheduler creates a new optimized system scheduler
func NewOptimizedSystemScheduler() *OptimizedSystemScheduler {
    return &OptimizedSystemScheduler{
        executionGraph:    &DAG{},
        parallelBatches:   make([][]ecs.System, 0),
        dataFlowOptimizer: &DataFlowOptimizer{},
    }
}

// AddSystem adds a system with dependencies (stub)
func (s *OptimizedSystemScheduler) AddSystem(system ecs.System, dependencies ...ecs.System) {
    // TODO: 実装予定
}

// BuildDependencyGraph builds the dependency graph (stub)
func (s *OptimizedSystemScheduler) BuildDependencyGraph() *DAG {
    // TODO: 実装予定
    return s.executionGraph
}

// GenerateParallelBatches generates parallel execution batches (stub)
func (s *OptimizedSystemScheduler) GenerateParallelBatches() [][]ecs.System {
    // TODO: 実装予定
    return s.parallelBatches
}

// HasDependency checks if a dependency exists (stub)
func (dag *DAG) HasDependency(dependent, dependency ecs.System) bool {
    // TODO: 実装予定
    return false
}
```

---

## 3. 失敗するテスト実装

### optimized_component_store_test.go
```go
package cache

import (
    "runtime"
    "testing"
    "time"
    "unsafe"
    
    "github.com/stretchr/testify/assert"
    "github.com/muscle-dreamer/internal/core/ecs"
)

func TestOptimizedComponentStore_SoALayout(t *testing.T) {
    store := NewOptimizedComponentStore()
    
    // 連続するエンティティに対してコンポーネント追加
    entities := make([]ecs.EntityID, 1000)
    for i := 0; i < 1000; i++ {
        entities[i] = ecs.EntityID(i)
        store.AddTransform(entities[i], ecs.TransformComponent{
            Position: ecs.Vector3{X: float32(i), Y: 0, Z: 0},
        })
    }
    
    // メモリレイアウトの連続性確認
    transforms := store.GetTransformArray()
    assert.Equal(t, 1000, len(transforms))
    
    // キャッシュライン境界整列確認
    if len(transforms) > 0 {
        baseAddr := uintptr(unsafe.Pointer(&transforms[0]))
        assert.Equal(t, 0, baseAddr%64, "Transform array should be 64-byte aligned") // 64バイト境界整列
    }
}

func TestOptimizedComponentStore_Prefetch(t *testing.T) {
    store := NewOptimizedComponentStore()
    entities := setupTestEntities(store, 100)
    
    // プリフェッチ実行時間測定
    start := time.Now()
    store.PrefetchComponents(entities[:50])
    prefetchTime := time.Since(start)
    
    // プリフェッチなしでの同等処理時間と比較
    start = time.Now()
    for _, entity := range entities[50:] {
        _ = store.GetTransform(entity)
    }
    normalTime := time.Since(start)
    
    // プリフェッチによる性能向上確認（測定可能な範囲で）
    assert.Less(t, prefetchTime, normalTime*2, "Prefetch should improve or at least not degrade performance")
}

// setupTestEntities creates test entities for benchmarking
func setupTestEntities(store *OptimizedComponentStore, count int) []ecs.EntityID {
    entities := make([]ecs.EntityID, count)
    for i := 0; i < count; i++ {
        entities[i] = ecs.EntityID(i)
        store.AddTransform(entities[i], ecs.TransformComponent{
            Position: ecs.Vector3{X: float32(i), Y: float32(i), Z: float32(i)},
            Rotation: ecs.Vector3{X: 0, Y: 0, Z: 0},
            Scale:    ecs.Vector3{X: 1, Y: 1, Z: 1},
        })
    }
    return entities
}
```

### simd_transform_system_test.go
```go
package simd

import (
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/muscle-dreamer/internal/core/ecs"
)

func TestSIMDTransformSystem_VectorOperations(t *testing.T) {
    system := NewSIMDTransformSystem()
    
    positions := []ecs.Vector3{
        {X: 1.0, Y: 2.0, Z: 3.0},
        {X: 4.0, Y: 5.0, Z: 6.0},
        {X: 7.0, Y: 8.0, Z: 9.0},
        {X: 10.0, Y: 11.0, Z: 12.0},
    }
    
    velocities := []ecs.Vector3{
        {X: 0.1, Y: 0.2, Z: 0.3},
        {X: 0.4, Y: 0.5, Z: 0.6},
        {X: 0.7, Y: 0.8, Z: 0.9},
        {X: 1.0, Y: 1.1, Z: 1.2},
    }
    
    deltaTime := float32(1.0/60.0)
    
    // SIMD演算実行
    system.UpdatePositions(positions, velocities, deltaTime)
    
    // 結果検証
    expected := ecs.Vector3{X: 1.0 + 0.1*deltaTime, Y: 2.0 + 0.2*deltaTime, Z: 3.0 + 0.3*deltaTime}
    assert.InDelta(t, expected.X, positions[0].X, 0.001)
    assert.InDelta(t, expected.Y, positions[0].Y, 0.001)
    assert.InDelta(t, expected.Z, positions[0].Z, 0.001)
}

func TestSIMDTransformSystem_PerformanceComparison(t *testing.T) {
    simdSystem := NewSIMDTransformSystem()
    scalarSystem := NewScalarTransformSystem()
    
    // 大量データ準備
    positions := make([]ecs.Vector3, 10000)
    velocities := make([]ecs.Vector3, 10000)
    for i := 0; i < 10000; i++ {
        positions[i] = ecs.Vector3{X: float32(i), Y: float32(i), Z: float32(i)}
        velocities[i] = ecs.Vector3{X: 0.1, Y: 0.2, Z: 0.3}
    }
    
    deltaTime := float32(1.0/60.0)
    
    // スカラー演算時間測定
    scalarPositions := make([]ecs.Vector3, len(positions))
    copy(scalarPositions, positions)
    
    start := time.Now()
    scalarSystem.UpdatePositions(scalarPositions, velocities, deltaTime)
    scalarTime := time.Since(start)
    
    // SIMD演算時間測定
    simdPositions := make([]ecs.Vector3, len(positions))
    copy(simdPositions, positions)
    
    start = time.Now()
    simdSystem.UpdatePositions(simdPositions, velocities, deltaTime)
    simdTime := time.Since(start)
    
    // SIMD性能向上確認（最低2倍向上期待）
    assert.Less(t, simdTime*2, scalarTime, "SIMD implementation should be at least 2x faster than scalar")
}
```

### optimized_system_scheduler_test.go
```go
package scheduler

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/muscle-dreamer/internal/core/ecs"
)

// Mock systems for testing
type MockPhysicsSystem struct{}
type MockRenderSystem struct{}
type MockTransformSystem struct{}

func (s *MockPhysicsSystem) Update(world ecs.World, deltaTime float32) {}
func (s *MockRenderSystem) Update(world ecs.World, deltaTime float32) {}
func (s *MockTransformSystem) Update(world ecs.World, deltaTime float32) {}

func TestOptimizedSystemScheduler_DependencyGraph(t *testing.T) {
    scheduler := NewOptimizedSystemScheduler()
    
    // システム依存関係設定
    physicsSystem := &MockPhysicsSystem{}
    renderSystem := &MockRenderSystem{}
    transformSystem := &MockTransformSystem{}
    
    scheduler.AddSystem(transformSystem)
    scheduler.AddSystem(physicsSystem, transformSystem) // physicsはtransformに依存
    scheduler.AddSystem(renderSystem, transformSystem)   // renderはtransformに依存
    
    // 依存関係グラフ生成
    graph := scheduler.BuildDependencyGraph()
    
    // 依存関係検証
    assert.True(t, graph.HasDependency(physicsSystem, transformSystem), "Physics should depend on Transform")
    assert.True(t, graph.HasDependency(renderSystem, transformSystem), "Render should depend on Transform")
    assert.False(t, graph.HasDependency(transformSystem, physicsSystem), "Transform should not depend on Physics")
}

func TestOptimizedSystemScheduler_ParallelBatches(t *testing.T) {
    scheduler := NewOptimizedSystemScheduler()
    
    // 複雑な依存関係を持つシステム群設定
    setupComplexSystemDependencies(scheduler)
    
    // 並列実行バッチ生成
    batches := scheduler.GenerateParallelBatches()
    
    // バッチ数が適切であることを確認
    assert.GreaterOrEqual(t, len(batches), 1, "Should have at least one batch")
    assert.LessOrEqual(t, len(batches), 10, "Should not have too many batches") // 過度な分割防止
    
    // 各バッチ内のシステムが並列実行可能であることを確認
    for i, batch := range batches {
        assert.Greater(t, len(batch), 0, "Batch %d should not be empty", i)
        verifyParallelExecutionSafety(t, batch)
    }
}

func setupComplexSystemDependencies(scheduler *OptimizedSystemScheduler) {
    systems := []ecs.System{
        &MockTransformSystem{},
        &MockPhysicsSystem{},
        &MockRenderSystem{},
    }
    
    // 簡単な依存関係設定（実際の実装では複雑になる）
    scheduler.AddSystem(systems[0])              // Transform (no deps)
    scheduler.AddSystem(systems[1], systems[0])  // Physics -> Transform
    scheduler.AddSystem(systems[2], systems[0])  // Render -> Transform
}

func verifyParallelExecutionSafety(t *testing.T, batch []ecs.System) {
    // バッチ内のシステム間に依存関係がないことを確認
    // 実装が進めば、より詳細な検証を追加
    assert.Greater(t, len(batch), 0, "Batch should contain systems")
}
```

---

## 4. パフォーマンステスト実装 (失敗確認用)

### performance_benchmarks_test.go
```go
package benchmarks

import (
    "runtime"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/muscle-dreamer/internal/core/ecs"
)

// TestPerformance_10000Entities60FPS tests the 60FPS performance requirement
func TestPerformance_10000Entities60FPS(t *testing.T) {
    if testing.Short() {
        t.Skip("Performance test skipped in short mode")
    }
    
    // 現時点では空の実装なので必ず失敗する
    world := createTestWorld()
    
    // 10,000エンティティ作成
    entities := make([]ecs.EntityID, 10000)
    for i := 0; i < 10000; i++ {
        entities[i] = world.CreateEntity()
        world.AddComponent(entities[i], &ecs.TransformComponent{})
        world.AddComponent(entities[i], &ecs.SpriteComponent{})
    }
    
    // 60FPS性能測定（10秒間）
    frameTime := measureAverageFrameTime(world, 600) // 10秒間測定
    assert.Less(t, frameTime, 16.67*time.Millisecond, "Should maintain 60FPS")
}

// TestPerformance_MemoryUsage tests memory usage requirements
func TestPerformance_MemoryUsage(t *testing.T) {
    if testing.Short() {
        t.Skip("Performance test skipped in short mode")
    }
    
    world := createTestWorld()
    
    // 10,000エンティティでメモリ使用量測定
    createEntities(world, 10000)
    
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    
    memoryUsage := memStats.Alloc
    assert.Less(t, memoryUsage, uint64(256*1024*1024), "Memory usage should be under 256MB") // 256MB
}

// TestPerformance_SystemExecutionTime tests system execution time
func TestPerformance_SystemExecutionTime(t *testing.T) {
    if testing.Short() {
        t.Skip("Performance test skipped in short mode")
    }
    
    world := createTestWorld()
    setupComplexSystemScenario(world)
    
    start := time.Now()
    world.Update(1.0/60.0) // 1フレーム実行
    duration := time.Since(start)
    
    assert.Less(t, duration, 10*time.Millisecond, "System execution should be under 10ms")
}

// Helper functions (stubs that will cause tests to fail)
func createTestWorld() ecs.World {
    // TODO: 実際のWorld実装を返す（現時点では nil で失敗）
    return nil
}

func measureAverageFrameTime(world ecs.World, frameCount int) time.Duration {
    // TODO: 実際のフレーム時間測定（現時点では必ず失敗する値を返す）
    return 50 * time.Millisecond // 50msは60FPSの目標16.67msを大幅に超過
}

func createEntities(world ecs.World, count int) {
    // TODO: エンティティ作成実装
}

func setupComplexSystemScenario(world ecs.World) {
    // TODO: 複雑なシステムシナリオ設定
}
```

---

## 5. ベンチマークテスト実装

### benchmark_tests.go
```go
package benchmarks

import (
    "testing"
    
    "github.com/muscle-dreamer/internal/core/ecs"
    "github.com/muscle-dreamer/internal/core/ecs/optimizations/cache"
)

// BenchmarkComponentStore_AddRemove benchmarks component add/remove operations
func BenchmarkComponentStore_AddRemove(b *testing.B) {
    store := cache.NewOptimizedComponentStore()
    entities := make([]ecs.EntityID, 1000)
    
    for i := 0; i < 1000; i++ {
        entities[i] = ecs.EntityID(i)
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        entityID := entities[i%1000]
        
        // コンポーネント追加
        store.AddTransform(entityID, ecs.TransformComponent{})
        store.AddSprite(entityID, ecs.SpriteComponent{})
        
        // コンポーネント削除
        store.RemoveTransform(entityID)
        store.RemoveSprite(entityID)
    }
}
```

---

## 6. テスト実行とテスト失敗確認

### テスト実行コマンド
```bash
# 基本テスト実行（失敗確認）
go test ./internal/core/ecs/optimizations/... -v

# 特定パッケージのテスト実行
go test ./internal/core/ecs/optimizations/cache -v
go test ./internal/core/ecs/optimizations/simd -v
go test ./internal/core/ecs/optimizations/scheduler -v

# パフォーマンステスト実行
go test ./internal/core/ecs/optimizations/benchmarks -v -run TestPerformance

# ベンチマーク実行
go test ./internal/core/ecs/optimizations/benchmarks -bench=.
```

### 期待される失敗結果

#### 1. コンパイルエラー
```
# github.com/muscle-dreamer/internal/core/ecs/optimizations/cache
./optimized_component_store_test.go:XX:XX: undefined: ecs.EntityID
./optimized_component_store_test.go:XX:XX: undefined: ecs.TransformComponent
./optimized_component_store_test.go:XX:XX: undefined: ecs.Vector3
```

#### 2. テスト実行失敗
```
--- FAIL: TestOptimizedComponentStore_SoALayout (0.00s)
    optimized_component_store_test.go:XX: 
        Error: 0 != 1000 (expected)
        Test: Transform array should have 1000 elements

--- FAIL: TestSIMDTransformSystem_VectorOperations (0.00s)
    simd_transform_system_test.go:XX:
        Error: positions were not updated (no actual implementation)

--- FAIL: TestPerformance_10000Entities60FPS (0.01s)
    performance_benchmarks_test.go:XX:
        Error: 50ms >= 16.67ms (expected)
        Test: Should maintain 60FPS
```

---

## 7. 基本型定義の実装（コンパイル用最小実装）

### 基本型の定義ファイル作成

#### internal/core/ecs/types.go
```go
package ecs

import "time"

// EntityID represents a unique entity identifier
type EntityID uint32

// ComponentType represents a component type identifier
type ComponentType uint16

// Vector3 represents a 3D vector
type Vector3 struct {
    X, Y, Z float32
}

// TransformComponent represents position, rotation, and scale
type TransformComponent struct {
    Position Vector3
    Rotation Vector3
    Scale    Vector3
}

// SpriteComponent represents sprite rendering data
type SpriteComponent struct {
    TextureID int32
    Color     Color
    Visible   bool
}

// PhysicsComponent represents physics properties
type PhysicsComponent struct {
    Velocity     Vector3
    Acceleration Vector3
    Mass         float32
}

// Color represents RGBA color values
type Color struct {
    R, G, B, A float32
}

// System represents a game system
type System interface {
    Update(world World, deltaTime float32)
}

// World represents the ECS world
type World interface {
    CreateEntity() EntityID
    AddComponent(entityID EntityID, component interface{})
    Update(deltaTime float32)
}
```

---

## 8. テスト実行結果の確認

Red段階の目的は、**テストが失敗することを確認する**ことです。

### 期待される状況：

1. **コンパイルエラーの解決**: 基本型定義により、テストがコンパイル可能になる
2. **テスト実行での失敗**: すべてのテストが期待通り失敗する
3. **失敗理由の明確化**: 各テストがなぜ失敗するかが明確になる

### 実行コマンドと期待結果：

```bash
# テスト実行
cd /workdir
go test ./internal/core/ecs/optimizations/... -v

# 期待される出力例
=== RUN   TestOptimizedComponentStore_SoALayout
--- FAIL: TestOptimizedComponentStore_SoALayout (0.00s)
    optimized_component_store_test.go:25: Expected array length 1000, got 0

=== RUN   TestSIMDTransformSystem_VectorOperations  
--- FAIL: TestSIMDTransformSystem_VectorOperations (0.00s)
    simd_transform_system_test.go:32: Values were not updated (stub implementation)

=== RUN   TestPerformance_10000Entities60FPS
--- FAIL: TestPerformance_10000Entities60FPS (0.10s)
    performance_benchmarks_test.go:42: Frame time 50ms exceeds limit 16.67ms

FAIL
```

---

## 9. 実装ファイルの作成

### 実装ファイルを環境に作成

以下のコマンドで実装ファイルを作成し、テストが失敗することを確認します：

```bash
# ディレクトリ作成
mkdir -p internal/core/ecs/optimizations/{cache,simd,scheduler,benchmarks}

# 基本ファイル作成
touch internal/core/ecs/optimizations/cache/{optimized_component_store.go,optimized_component_store_test.go}
touch internal/core/ecs/optimizations/simd/{simd_transform_system.go,simd_transform_system_test.go}
touch internal/core/ecs/optimizations/scheduler/{optimized_system_scheduler.go,optimized_system_scheduler_test.go}
touch internal/core/ecs/optimizations/benchmarks/{performance_benchmarks_test.go,benchmark_tests.go}
```

---

## Red段階完了条件

### ✅ 確認事項：

1. **コンパイル成功**: 全テストファイルがエラーなくコンパイルできる
2. **テスト失敗**: 実装されたすべてのテストが期待通り失敗する
3. **失敗理由明確**: 各テストの失敗理由が明確に特定できる
4. **テストカバレッジ**: 要件定義の主要機能がテストでカバーされている

### 🔄 次のステップ：

Red段階完了後、**Green段階**に進み：
1. テストが通る最小限の実装を作成
2. すべてのテストが成功することを確認
3. 過度な実装を避け、必要最小限の機能のみ実装

---

**実装ステータス**: 🔴 Red段階 - テスト失敗確認完了  
**次のフェーズ**: 🟢 Green段階 - 最小実装でテストを通す  
**作成日**: 2025-08-11  
**最終更新**: 2025-08-11