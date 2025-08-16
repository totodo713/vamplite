# TASK-401: パフォーマンス最適化 - テストケース仕様

## 概要

パフォーマンス最適化の全機能に対する包括的なテストケースを定義します。単体テスト、統合テスト、パフォーマンステスト、ベンチマークテスト、ストレステストを含む多層的なテスト戦略を実装します。

## テストカテゴリ

### 1. 単体テスト (Unit Tests)
### 2. 統合テスト (Integration Tests)  
### 3. パフォーマンステスト (Performance Tests)
### 4. ベンチマークテスト (Benchmark Tests)
### 5. ストレステスト (Stress Tests)
### 6. プロファイリングテスト (Profiling Tests)

---

## 1. 単体テスト (Unit Tests)

### UT-401-001: CPUキャッシュ効率最適化テスト

#### UT-401-001-A: SoAレイアウトメモリ配置テスト
```go
func TestOptimizedComponentStore_SoALayout(t *testing.T) {
    store := NewOptimizedComponentStore()
    
    // 連続するエンティティに対してコンポーネント追加
    entities := make([]EntityID, 1000)
    for i := 0; i < 1000; i++ {
        entities[i] = EntityID(i)
        store.AddTransform(entities[i], TransformComponent{
            Position: Vector3{X: float32(i), Y: 0, Z: 0},
        })
    }
    
    // メモリレイアウトの連続性確認
    transforms := store.GetTransformArray()
    assert.Equal(t, 1000, len(transforms))
    
    // キャッシュライン境界整列確認
    baseAddr := uintptr(unsafe.Pointer(&transforms[0]))
    assert.Equal(t, 0, baseAddr%64) // 64バイト境界整列
}
```

#### UT-401-001-B: メモリプリフェッチテスト
```go
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
    assert.Less(t, prefetchTime, normalTime*2)
}
```

### UT-401-002: SIMD命令活用テスト

#### UT-401-002-A: ベクトル演算テスト
```go
func TestSIMDTransformSystem_VectorOperations(t *testing.T) {
    system := NewSIMDTransformSystem()
    
    positions := []Vector3{
        {X: 1.0, Y: 2.0, Z: 3.0},
        {X: 4.0, Y: 5.0, Z: 6.0},
        {X: 7.0, Y: 8.0, Z: 9.0},
        {X: 10.0, Y: 11.0, Z: 12.0},
    }
    
    velocities := []Vector3{
        {X: 0.1, Y: 0.2, Z: 0.3},
        {X: 0.4, Y: 0.5, Z: 0.6},
        {X: 0.7, Y: 0.8, Z: 0.9},
        {X: 1.0, Y: 1.1, Z: 1.2},
    }
    
    deltaTime := float32(1.0/60.0)
    
    // SIMD演算実行
    system.UpdatePositions(positions, velocities, deltaTime)
    
    // 結果検証
    expected := Vector3{X: 1.0 + 0.1*deltaTime, Y: 2.0 + 0.2*deltaTime, Z: 3.0 + 0.3*deltaTime}
    assert.InDelta(t, expected.X, positions[0].X, 0.001)
    assert.InDelta(t, expected.Y, positions[0].Y, 0.001)
    assert.InDelta(t, expected.Z, positions[0].Z, 0.001)
}
```

#### UT-401-002-B: SIMD性能比較テスト
```go
func TestSIMDTransformSystem_PerformanceComparison(t *testing.T) {
    simdSystem := NewSIMDTransformSystem()
    scalarSystem := NewScalarTransformSystem()
    
    // 大量データ準備
    positions := make([]Vector3, 10000)
    velocities := make([]Vector3, 10000)
    for i := 0; i < 10000; i++ {
        positions[i] = Vector3{X: float32(i), Y: float32(i), Z: float32(i)}
        velocities[i] = Vector3{X: 0.1, Y: 0.2, Z: 0.3}
    }
    
    deltaTime := float32(1.0/60.0)
    
    // スカラー演算時間測定
    scalarPositions := make([]Vector3, len(positions))
    copy(scalarPositions, positions)
    
    start := time.Now()
    scalarSystem.UpdatePositions(scalarPositions, velocities, deltaTime)
    scalarTime := time.Since(start)
    
    // SIMD演算時間測定
    simdPositions := make([]Vector3, len(positions))
    copy(simdPositions, positions)
    
    start = time.Now()
    simdSystem.UpdatePositions(simdPositions, velocities, deltaTime)
    simdTime := time.Since(start)
    
    // SIMD性能向上確認（最低2倍向上期待）
    assert.Less(t, simdTime*2, scalarTime)
}
```

### UT-401-003: システム実行順序最適化テスト

#### UT-401-003-A: 依存関係グラフ生成テスト
```go
func TestOptimizedSystemScheduler_DependencyGraph(t *testing.T) {
    scheduler := NewOptimizedSystemScheduler()
    
    // システム依存関係設定
    physicsSystem := &PhysicsSystem{}
    renderSystem := &RenderSystem{}
    transformSystem := &TransformSystem{}
    
    scheduler.AddSystem(transformSystem)
    scheduler.AddSystem(physicsSystem, transformSystem) // physicsはtransformに依存
    scheduler.AddSystem(renderSystem, transformSystem)   // renderはtransformに依存
    
    // 依存関係グラフ生成
    graph := scheduler.BuildDependencyGraph()
    
    // 依存関係検証
    assert.True(t, graph.HasDependency(physicsSystem, transformSystem))
    assert.True(t, graph.HasDependency(renderSystem, transformSystem))
    assert.False(t, graph.HasDependency(transformSystem, physicsSystem))
}
```

#### UT-401-003-B: 並列実行バッチ生成テスト
```go
func TestOptimizedSystemScheduler_ParallelBatches(t *testing.T) {
    scheduler := NewOptimizedSystemScheduler()
    
    // 複雑な依存関係を持つシステム群設定
    setupComplexSystemDependencies(scheduler)
    
    // 並列実行バッチ生成
    batches := scheduler.GenerateParallelBatches()
    
    // バッチ数が適切であることを確認
    assert.GreaterOrEqual(t, len(batches), 1)
    assert.LessOrEqual(t, len(batches), 10) // 過度な分割防止
    
    // 各バッチ内のシステムが並列実行可能であることを確認
    for _, batch := range batches {
        verifyParallelExecutionSafety(t, batch)
    }
}
```

---

## 2. 統合テスト (Integration Tests)

### IT-401-001: 最適化機能統合テスト

#### IT-401-001-A: ECS全体最適化統合テスト
```go
func TestOptimizedECS_FullIntegration(t *testing.T) {
    world := CreateOptimizedWorld()
    
    // 複雑なシナリオセットアップ
    entities := make([]EntityID, 5000)
    for i := 0; i < 5000; i++ {
        entities[i] = world.CreateEntity()
        world.AddComponent(entities[i], &TransformComponent{})
        world.AddComponent(entities[i], &SpriteComponent{})
        if i%2 == 0 {
            world.AddComponent(entities[i], &PhysicsComponent{})
        }
    }
    
    // 複数フレーム実行
    totalTime := time.Duration(0)
    frameCount := 100
    
    for i := 0; i < frameCount; i++ {
        start := time.Now()
        world.Update(1.0/60.0)
        frameTime := time.Since(start)
        totalTime += frameTime
        
        // 各フレームが目標時間内であることを確認
        assert.Less(t, frameTime, 16*time.Millisecond)
    }
    
    // 平均フレーム時間確認
    averageFrameTime := totalTime / time.Duration(frameCount)
    assert.Less(t, averageFrameTime, 15*time.Millisecond)
}
```

### IT-401-002: システム間連携最適化テスト

#### IT-401-002-A: データフロー最適化テスト
```go
func TestOptimizedDataFlow_SystemInteraction(t *testing.T) {
    world := CreateOptimizedWorld()
    
    // データフロー追跡システム設定
    dataFlowTracker := NewDataFlowTracker()
    world.AddSystem(&TransformSystem{Tracker: dataFlowTracker})
    world.AddSystem(&PhysicsSystem{Tracker: dataFlowTracker})
    world.AddSystem(&RenderSystem{Tracker: dataFlowTracker})
    
    // テストシナリオ実行
    setupDataFlowTestScenario(world)
    world.Update(1.0/60.0)
    
    // データフローの効率性検証
    flowStats := dataFlowTracker.GetStatistics()
    assert.Less(t, flowStats.UnnecessaryDataCopies, 5) // 不要なデータコピー最小化
    assert.Greater(t, flowStats.CacheHitRatio, 0.9)    // キャッシュヒット率90%以上
}
```

---

## 3. パフォーマンステスト (Performance Tests)

### PT-401-001: フレームレート性能テスト

#### PT-401-001-A: 10,000エンティティ@60FPSテスト
```go
func TestPerformance_10000Entities60FPS(t *testing.T) {
    if testing.Short() {
        t.Skip("Performance test skipped in short mode")
    }
    
    world := CreateOptimizedWorld()
    
    // 10,000エンティティ作成
    entities := make([]EntityID, 10000)
    for i := 0; i < 10000; i++ {
        entities[i] = world.CreateEntity()
        world.AddComponent(entities[i], &TransformComponent{
            Position: Vector3{X: rand.Float32(), Y: rand.Float32(), Z: rand.Float32()},
            Rotation: Vector3{X: rand.Float32(), Y: rand.Float32(), Z: rand.Float32()},
            Scale:    Vector3{X: 1.0, Y: 1.0, Z: 1.0},
        })
        world.AddComponent(entities[i], &SpriteComponent{
            TextureID: rand.Int31n(100),
            Color:     Color{R: 1.0, G: 1.0, B: 1.0, A: 1.0},
        })
        if i%3 == 0 {
            world.AddComponent(entities[i], &PhysicsComponent{
                Velocity:     Vector3{X: rand.Float32(), Y: rand.Float32(), Z: 0},
                Acceleration: Vector3{X: 0, Y: -9.8, Z: 0},
            })
        }
    }
    
    // 60FPS性能測定（10秒間）
    frameCount := 600
    totalFrameTime := time.Duration(0)
    maxFrameTime := time.Duration(0)
    frameTimeTarget := 16667 * time.Microsecond // 16.667ms
    
    for i := 0; i < frameCount; i++ {
        start := time.Now()
        world.Update(1.0/60.0)
        frameTime := time.Since(start)
        
        totalFrameTime += frameTime
        if frameTime > maxFrameTime {
            maxFrameTime = frameTime
        }
        
        // 個別フレーム時間確認
        if frameTime > frameTimeTarget {
            t.Logf("Frame %d exceeded target: %v > %v", i, frameTime, frameTimeTarget)
        }
    }
    
    averageFrameTime := totalFrameTime / time.Duration(frameCount)
    
    // パフォーマンス要件確認
    assert.Less(t, averageFrameTime, frameTimeTarget,
        "Average frame time %v exceeds target %v", averageFrameTime, frameTimeTarget)
    assert.Less(t, maxFrameTime, frameTimeTarget*2,
        "Maximum frame time %v exceeds acceptable limit %v", maxFrameTime, frameTimeTarget*2)
    
    t.Logf("Performance Results:")
    t.Logf("  Average frame time: %v", averageFrameTime)
    t.Logf("  Maximum frame time: %v", maxFrameTime)
    t.Logf("  Entities processed: %d", len(entities))
}
```

### PT-401-002: メモリ使用量テスト

#### PT-401-002-A: メモリ効率テスト
```go
func TestPerformance_MemoryUsage(t *testing.T) {
    if testing.Short() {
        t.Skip("Performance test skipped in short mode")
    }
    
    // 初期メモリ状況記録
    runtime.GC()
    runtime.GC() // 確実にGCを実行
    
    var initialMemStats runtime.MemStats
    runtime.ReadMemStats(&initialMemStats)
    
    world := CreateOptimizedWorld()
    
    // 10,000エンティティでメモリ使用量測定
    entityCount := 10000
    entities := make([]EntityID, entityCount)
    
    for i := 0; i < entityCount; i++ {
        entities[i] = world.CreateEntity()
        world.AddComponent(entities[i], &TransformComponent{})
        world.AddComponent(entities[i], &SpriteComponent{})
        world.AddComponent(entities[i], &PhysicsComponent{})
    }
    
    // メモリ使用量測定
    runtime.GC()
    runtime.GC()
    
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    
    memoryUsed := memStats.Alloc - initialMemStats.Alloc
    memoryPerEntity := memoryUsed / uint64(entityCount)
    
    // メモリ効率要件確認
    maxTotalMemory := uint64(256 * 1024 * 1024) // 256MB
    maxMemoryPerEntity := uint64(100)           // 100B per entity
    
    assert.Less(t, memoryUsed, maxTotalMemory,
        "Total memory usage %d bytes exceeds limit %d bytes", memoryUsed, maxTotalMemory)
    assert.Less(t, memoryPerEntity, maxMemoryPerEntity,
        "Memory per entity %d bytes exceeds limit %d bytes", memoryPerEntity, maxMemoryPerEntity)
    
    t.Logf("Memory Usage Results:")
    t.Logf("  Total memory used: %d bytes (%.2f MB)", memoryUsed, float64(memoryUsed)/(1024*1024))
    t.Logf("  Memory per entity: %d bytes", memoryPerEntity)
    t.Logf("  Entity count: %d", entityCount)
}
```

### PT-401-003: システム実行時間テスト

#### PT-401-003-A: システム実行時間制限テスト
```go
func TestPerformance_SystemExecutionTime(t *testing.T) {
    if testing.Short() {
        t.Skip("Performance test skipped in short mode")
    }
    
    world := CreateOptimizedWorld()
    
    // 50システム登録
    systems := make([]System, 50)
    for i := 0; i < 50; i++ {
        systems[i] = CreateTestSystem(i)
        world.AddSystem(systems[i])
    }
    
    // 複雑なテストデータ設定
    setupComplexSystemScenario(world, 5000) // 5000エンティティ
    
    // システム実行時間測定
    executionTimes := make([]time.Duration, 100)
    maxExecutionTime := 10 * time.Millisecond
    
    for i := 0; i < 100; i++ {
        start := time.Now()
        world.UpdateSystems(1.0/60.0)
        executionTimes[i] = time.Since(start)
        
        assert.Less(t, executionTimes[i], maxExecutionTime,
            "System execution time %v exceeds limit %v on iteration %d", 
            executionTimes[i], maxExecutionTime, i)
    }
    
    // 統計情報
    totalTime := time.Duration(0)
    maxTime := time.Duration(0)
    for _, execTime := range executionTimes {
        totalTime += execTime
        if execTime > maxTime {
            maxTime = execTime
        }
    }
    averageTime := totalTime / time.Duration(len(executionTimes))
    
    t.Logf("System Execution Time Results:")
    t.Logf("  Average execution time: %v", averageTime)
    t.Logf("  Maximum execution time: %v", maxTime)
    t.Logf("  Systems count: %d", len(systems))
}
```

---

## 4. ベンチマークテスト (Benchmark Tests)

### BT-401-001: コンポーネント操作ベンチマーク

#### BT-401-001-A: コンポーネント追加・削除ベンチマーク
```go
func BenchmarkComponentStore_AddRemove(b *testing.B) {
    store := NewOptimizedComponentStore()
    entities := make([]EntityID, 1000)
    
    for i := 0; i < 1000; i++ {
        entities[i] = EntityID(i)
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        entityID := entities[i%1000]
        
        // コンポーネント追加
        store.AddTransform(entityID, TransformComponent{})
        store.AddSprite(entityID, SpriteComponent{})
        
        // コンポーネント削除
        store.RemoveTransform(entityID)
        store.RemoveSprite(entityID)
    }
}
```

#### BT-401-001-B: コンポーネントクエリベンチマーク
```go
func BenchmarkQueryEngine_ComplexQuery(b *testing.B) {
    world := CreateOptimizedWorld()
    
    // テストデータセットアップ
    for i := 0; i < 10000; i++ {
        entity := world.CreateEntity()
        world.AddComponent(entity, &TransformComponent{})
        if i%2 == 0 {
            world.AddComponent(entity, &SpriteComponent{})
        }
        if i%3 == 0 {
            world.AddComponent(entity, &PhysicsComponent{})
        }
    }
    
    query := world.Query().
        With(ComponentTypeTransform).
        With(ComponentTypeSprite).
        Without(ComponentTypePhysics)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        results := query.Execute()
        _ = results // 結果を使用
    }
}
```

### BT-401-002: システム実行ベンチマーク

#### BT-401-002-A: 単一システムベンチマーク
```go
func BenchmarkTransformSystem_Update(b *testing.B) {
    world := CreateOptimizedWorld()
    system := &OptimizedTransformSystem{}
    world.AddSystem(system)
    
    // 10,000エンティティ作成
    for i := 0; i < 10000; i++ {
        entity := world.CreateEntity()
        world.AddComponent(entity, &TransformComponent{
            Position: Vector3{X: float32(i), Y: 0, Z: 0},
        })
        world.AddComponent(entity, &PhysicsComponent{
            Velocity: Vector3{X: 1.0, Y: 0, Z: 0},
        })
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        system.Update(world, 1.0/60.0)
    }
}
```

### BT-401-003: 並列処理ベンチマーク

#### BT-401-003-A: 並列システム実行ベンチマーク
```go
func BenchmarkParallelSystems_Execution(b *testing.B) {
    world := CreateOptimizedWorld()
    
    // 並列実行可能な独立システム群追加
    for i := 0; i < 10; i++ {
        world.AddSystem(CreateIndependentSystem(i))
    }
    
    setupBenchmarkEntities(world, 5000)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        world.UpdateSystemsParallel(1.0/60.0)
    }
}
```

---

## 5. ストレステスト (Stress Tests)

### ST-401-001: 大量エンティティストレステスト

#### ST-401-001-A: 100,000エンティティ処理テスト
```go
func TestStress_100000Entities(t *testing.T) {
    if testing.Short() {
        t.Skip("Stress test skipped in short mode")
    }
    
    world := CreateOptimizedWorld()
    
    // 100,000エンティティ作成
    entityCount := 100000
    t.Logf("Creating %d entities...", entityCount)
    
    startTime := time.Now()
    entities := make([]EntityID, entityCount)
    
    for i := 0; i < entityCount; i++ {
        entities[i] = world.CreateEntity()
        world.AddComponent(entities[i], &TransformComponent{})
        
        if i%10 == 0 && i%(entityCount/100) == 0 {
            progress := float64(i) / float64(entityCount) * 100
            t.Logf("Progress: %.1f%%", progress)
        }
    }
    
    creationTime := time.Since(startTime)
    t.Logf("Entity creation completed in %v", creationTime)
    
    // メモリ使用量確認
    runtime.GC()
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    
    memoryUsed := memStats.Alloc
    memoryPerEntity := memoryUsed / uint64(entityCount)
    
    t.Logf("Memory usage: %d bytes total, %d bytes per entity", memoryUsed, memoryPerEntity)
    
    // 処理性能確認（1時間連続実行は現実的でないため10分に短縮）
    testDuration := 10 * time.Minute
    endTime := time.Now().Add(testDuration)
    frameCount := 0
    
    t.Logf("Starting %v stress test...", testDuration)
    
    for time.Now().Before(endTime) {
        frameStart := time.Now()
        world.Update(1.0/60.0)
        frameTime := time.Since(frameStart)
        frameCount++
        
        // フレーム時間が極端に長い場合はアラート
        if frameTime > 50*time.Millisecond {
            t.Logf("Warning: Frame %d took %v", frameCount, frameTime)
        }
        
        // 定期的な進捗報告
        if frameCount%1800 == 0 { // 30秒毎
            elapsed := time.Since(time.Now().Add(-testDuration))
            remaining := testDuration - elapsed
            t.Logf("Stress test progress: %v remaining, frame %d", remaining, frameCount)
        }
    }
    
    // 最終結果確認
    finalMemStats := runtime.MemStats{}
    runtime.ReadMemStats(&finalMemStats)
    
    memoryLeak := finalMemStats.Alloc - memStats.Alloc
    t.Logf("Stress test completed:")
    t.Logf("  Total frames: %d", frameCount)
    t.Logf("  Memory leak: %d bytes", memoryLeak)
    
    // メモリリーク許容範囲確認（50MB以下）
    maxAllowedLeak := uint64(50 * 1024 * 1024)
    assert.Less(t, memoryLeak, maxAllowedLeak,
        "Memory leak %d bytes exceeds limit %d bytes", memoryLeak, maxAllowedLeak)
}
```

### ST-401-002: システム負荷ストレステスト

#### ST-401-002-A: 高負荷システム実行テスト
```go
func TestStress_HighLoadSystems(t *testing.T) {
    if testing.Short() {
        t.Skip("Stress test skipped in short mode")
    }
    
    world := CreateOptimizedWorld()
    
    // 負荷の高いシステム群追加
    systems := []System{
        &HighComplexityPhysicsSystem{},
        &DetailedRenderingSystem{},
        &AIDecisionSystem{},
        &NetworkSyncSystem{},
        &AudioProcessingSystem{},
    }
    
    for _, system := range systems {
        world.AddSystem(system)
    }
    
    // 高負荷テストシナリオ設定
    setupHighLoadScenario(world, 10000)
    
    // 30分間の高負荷実行テスト
    testDuration := 30 * time.Minute
    endTime := time.Now().Add(testDuration)
    
    maxFrameTime := time.Duration(0)
    frameTimeSum := time.Duration(0)
    frameCount := 0
    frameTimeTarget := 16670 * time.Microsecond
    
    for time.Now().Before(endTime) {
        frameStart := time.Now()
        world.Update(1.0/60.0)
        frameTime := time.Since(frameStart)
        
        frameCount++
        frameTimeSum += frameTime
        
        if frameTime > maxFrameTime {
            maxFrameTime = frameTime
        }
        
        // フレーム時間オーバー時の記録
        if frameTime > frameTimeTarget*2 {
            t.Logf("Frame %d exceeded 2x target: %v", frameCount, frameTime)
        }
    }
    
    averageFrameTime := frameTimeSum / time.Duration(frameCount)
    
    t.Logf("High load stress test results:")
    t.Logf("  Total frames: %d", frameCount)
    t.Logf("  Average frame time: %v", averageFrameTime)
    t.Logf("  Maximum frame time: %v", maxFrameTime)
    
    // 性能劣化が許容範囲内であることを確認
    assert.Less(t, averageFrameTime, frameTimeTarget*1.5,
        "Average frame time degraded beyond acceptable limit")
}
```

---

## 6. プロファイリングテスト (Profiling Tests)

### PF-401-001: CPUプロファイリングテスト

#### PF-401-001-A: ホットスポット特定テスト
```go
func TestProfiling_CPUHotspots(t *testing.T) {
    if testing.Short() {
        t.Skip("Profiling test skipped in short mode")
    }
    
    // CPUプロファイリング開始
    cpuProfileFile, err := os.Create("cpu_profile.prof")
    require.NoError(t, err)
    defer cpuProfileFile.Close()
    
    err = pprof.StartCPUProfile(cpuProfileFile)
    require.NoError(t, err)
    defer pprof.StopCPUProfile()
    
    world := CreateOptimizedWorld()
    setupProfilingTestScenario(world, 10000)
    
    // プロファイリング実行（5分間）
    endTime := time.Now().Add(5 * time.Minute)
    for time.Now().Before(endTime) {
        world.Update(1.0/60.0)
    }
    
    t.Log("CPU profiling completed. Analyze with: go tool pprof cpu_profile.prof")
}
```

### PF-401-002: メモリプロファイリングテスト

#### PF-401-002-A: メモリ使用パターン解析テスト
```go
func TestProfiling_MemoryUsage(t *testing.T) {
    if testing.Short() {
        t.Skip("Profiling test skipped in short mode")
    }
    
    world := CreateOptimizedWorld()
    setupProfilingTestScenario(world, 10000)
    
    // 初期メモリプロファイル取得
    runtime.GC()
    initialProfile := getMemoryProfile()
    
    // 負荷実行
    for i := 0; i < 3600; i++ { // 1分間（60FPS）
        world.Update(1.0/60.0)
    }
    
    // 最終メモリプロファイル取得
    runtime.GC()
    finalProfile := getMemoryProfile()
    
    // メモリプロファイル保存
    memProfileFile, err := os.Create("mem_profile.prof")
    require.NoError(t, err)
    defer memProfileFile.Close()
    
    err = pprof.WriteHeapProfile(memProfileFile)
    require.NoError(t, err)
    
    // メモリ使用パターン解析
    memoryGrowth := finalProfile.TotalAlloc - initialProfile.TotalAlloc
    t.Logf("Memory growth during test: %d bytes", memoryGrowth)
    t.Log("Memory profiling completed. Analyze with: go tool pprof mem_profile.prof")
    
    // 異常なメモリ増加の検出
    maxAcceptableGrowth := uint64(100 * 1024 * 1024) // 100MB
    assert.Less(t, memoryGrowth, maxAcceptableGrowth,
        "Memory growth %d exceeds acceptable limit %d", memoryGrowth, maxAcceptableGrowth)
}
```

---

## テスト実行設定

### テスト環境設定
```go
func init() {
    // パフォーマンステスト用設定
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // メモリ設定
    debug.SetGCPercent(100) // GC頻度調整
}
```

### ヘルパー関数

#### テストデータ生成
```go
func setupTestEntities(store *OptimizedComponentStore, count int) []EntityID {
    entities := make([]EntityID, count)
    for i := 0; i < count; i++ {
        entities[i] = EntityID(i)
        store.AddTransform(entities[i], TransformComponent{
            Position: Vector3{X: rand.Float32(), Y: rand.Float32(), Z: rand.Float32()},
        })
    }
    return entities
}

func setupComplexSystemScenario(world *World, entityCount int) {
    for i := 0; i < entityCount; i++ {
        entity := world.CreateEntity()
        world.AddComponent(entity, &TransformComponent{})
        world.AddComponent(entity, &SpriteComponent{})
        
        if i%3 == 0 {
            world.AddComponent(entity, &PhysicsComponent{})
        }
        if i%5 == 0 {
            world.AddComponent(entity, &AIComponent{})
        }
    }
}
```

---

## テスト実行コマンド

### 基本テスト実行
```bash
# 全テスト実行
go test ./internal/core/ecs/optimizations/... -v

# パフォーマンステストのみ
go test ./internal/core/ecs/optimizations/... -v -run TestPerformance

# ベンチマークテスト実行
go test ./internal/core/ecs/optimizations/... -bench=. -benchmem

# ストレステスト実行
go test ./internal/core/ecs/optimizations/... -v -run TestStress -timeout=2h

# プロファイリングテスト実行
go test ./internal/core/ecs/optimizations/... -v -run TestProfiling -cpuprofile=cpu.prof -memprofile=mem.prof
```

### CI/CD設定
```yaml
performance_tests:
  - name: "Performance Tests"
    run: go test -v -run TestPerformance -timeout=30m
    
benchmark_tests:
  - name: "Benchmark Tests"  
    run: go test -bench=. -benchmem -count=3
    
stress_tests:
  - name: "Stress Tests"
    run: go test -v -run TestStress -timeout=2h
    when: nightly
```

---

## 成功基準

### 必須パフォーマンス基準
- [ ] フレーム時間<16.67ms（60FPS）
- [ ] エンティティ作成1000個/フレーム
- [ ] クエリ実行時間<1ms  
- [ ] メモリ使用量<256MB（10,000エンティティ）
- [ ] システム実行時間<10ms

### 品質基準
- [ ] 全テストケース通過率100%
- [ ] メモリリーク<50MB/24h
- [ ] 長期安定性確保（24時間連続実行）

### ベンチマーク基準
- [ ] 競合ライブラリと同等以上の性能
- [ ] CPU効率向上（SIMD活用による4倍性能向上）
- [ ] メモリ効率向上（SoAによる2倍アクセス性能向上）

---

**作成日**: 2025-08-11  
**最終更新**: 2025-08-11  
**承認者**: ECSアーキテクト  
**レビュー状況**: ✅ レビュー完了