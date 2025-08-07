# TASK-202: MemoryManager実装 - テストケース仕様

## テストカテゴリ

### 1. ObjectPool基本機能テスト

#### Test_ObjectPool_Creation
```go
// プールの作成と基本プロパティの確認
func Test_ObjectPool_Creation(t *testing.T) {
    // Given: プールパラメータ
    poolName := "TestPool"
    objectSize := 64
    initialCapacity := 100
    
    // When: プールを作成
    pool := NewObjectPool(poolName, objectSize, initialCapacity)
    
    // Then: プロパティが正しく設定されている
    assert.NotNil(t, pool)
    assert.Equal(t, objectSize, pool.ObjectSize())
    assert.Equal(t, initialCapacity, pool.Capacity())
    assert.Equal(t, 0, pool.Size()) // 初期状態では使用中オブジェクトは0
}
```

#### Test_ObjectPool_GetPut
```go
// オブジェクトの取得と返却
func Test_ObjectPool_GetPut(t *testing.T) {
    // Given: 初期化されたプール
    pool := NewObjectPool("TestPool", 64, 10)
    
    // When: オブジェクトを取得
    ptr1, err := pool.Get()
    assert.NoError(t, err)
    assert.NotNil(t, ptr1)
    assert.Equal(t, 1, pool.Size())
    
    // When: 別のオブジェクトを取得
    ptr2, err := pool.Get()
    assert.NoError(t, err)
    assert.NotNil(t, ptr2)
    assert.NotEqual(t, ptr1, ptr2) // 異なるアドレス
    assert.Equal(t, 2, pool.Size())
    
    // When: オブジェクトを返却
    err = pool.Put(ptr1)
    assert.NoError(t, err)
    assert.Equal(t, 1, pool.Size())
    
    // When: 同じオブジェクトを再取得
    ptr3, err := pool.Get()
    assert.NoError(t, err)
    assert.Equal(t, ptr1, ptr3) // 再利用される
}
```

#### Test_ObjectPool_Overflow
```go
// プール容量を超えた場合の自動拡張
func Test_ObjectPool_Overflow(t *testing.T) {
    // Given: 小さな容量のプール
    pool := NewObjectPool("TestPool", 32, 2)
    
    // When: 容量を超えてオブジェクトを取得
    ptrs := make([]unsafe.Pointer, 5)
    for i := 0; i < 5; i++ {
        ptr, err := pool.Get()
        assert.NoError(t, err)
        assert.NotNil(t, ptr)
        ptrs[i] = ptr
    }
    
    // Then: プールが自動拡張される
    assert.True(t, pool.Capacity() >= 5)
    assert.Equal(t, 5, pool.Size())
}
```

### 2. MemoryManager基本機能テスト

#### Test_MemoryManager_CreatePool
```go
// メモリマネージャーでのプール作成
func Test_MemoryManager_CreatePool(t *testing.T) {
    // Given: メモリマネージャー
    mm := NewMemoryManager()
    
    // When: プールを作成
    err := mm.CreatePool("EntityPool", 128, 1000)
    assert.NoError(t, err)
    
    // Then: プールが取得できる
    pool, err := mm.GetPool("EntityPool")
    assert.NoError(t, err)
    assert.NotNil(t, pool)
    
    // When: 同じ名前でプールを作成
    err = mm.CreatePool("EntityPool", 64, 500)
    
    // Then: エラーが返される
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "already exists")
}
```

#### Test_MemoryManager_Allocate
```go
// 直接メモリ割り当て
func Test_MemoryManager_Allocate(t *testing.T) {
    // Given: メモリマネージャー
    mm := NewMemoryManager()
    
    // When: メモリを割り当て
    ptr, err := mm.Allocate(256)
    assert.NoError(t, err)
    assert.NotNil(t, ptr)
    
    // Then: メモリ使用量が増加
    usage := mm.GetMemoryUsage()
    assert.True(t, usage.Allocated >= 256)
    
    // When: メモリを解放
    err = mm.Deallocate(ptr)
    assert.NoError(t, err)
}
```

#### Test_MemoryManager_AllocateAligned
```go
// アラインメント付きメモリ割り当て
func Test_MemoryManager_AllocateAligned(t *testing.T) {
    // Given: メモリマネージャー
    mm := NewMemoryManager()
    
    // When: 64バイトアラインメントでメモリを割り当て
    ptr, err := mm.AllocateAligned(100, 64)
    assert.NoError(t, err)
    assert.NotNil(t, ptr)
    
    // Then: アドレスが64の倍数
    address := uintptr(ptr)
    assert.Equal(t, uintptr(0), address%64)
}
```

### 3. GC制御テスト

#### Test_MemoryManager_GCControl
```go
// GC制御機能
func Test_MemoryManager_GCControl(t *testing.T) {
    // Given: メモリマネージャー
    mm := NewMemoryManager()
    
    // When: GCしきい値を設定
    err := mm.SetGCThreshold(10 * 1024 * 1024) // 10MB
    assert.NoError(t, err)
    
    // When: 大量のメモリを割り当て
    ptrs := make([]unsafe.Pointer, 100)
    for i := 0; i < 100; i++ {
        ptr, _ := mm.Allocate(1024 * 100) // 100KB each
        ptrs[i] = ptr
    }
    
    // When: 手動でGCをトリガー
    initialStats := mm.GetGCStats()
    err = mm.TriggerGC()
    assert.NoError(t, err)
    
    // Then: GC統計が更新される
    newStats := mm.GetGCStats()
    assert.True(t, newStats.NumGC > initialStats.NumGC)
}
```

### 4. メモリ監視テスト

#### Test_MemoryManager_UsageTracking
```go
// メモリ使用量追跡
func Test_MemoryManager_UsageTracking(t *testing.T) {
    // Given: メモリマネージャー
    mm := NewMemoryManager()
    
    // When: 初期状態を確認
    usage := mm.GetMemoryUsage()
    initialAllocated := usage.Allocated
    
    // When: メモリを割り当て
    ptrs := make([]unsafe.Pointer, 10)
    for i := 0; i < 10; i++ {
        ptr, _ := mm.Allocate(1024)
        ptrs[i] = ptr
    }
    
    // Then: 使用量が増加
    usage = mm.GetMemoryUsage()
    assert.True(t, usage.Allocated > initialAllocated)
    assert.True(t, usage.Used >= 10*1024)
    
    // When: メモリを解放
    for _, ptr := range ptrs {
        mm.Deallocate(ptr)
    }
    
    // Then: 使用量が減少
    usage = mm.GetMemoryUsage()
    assert.True(t, usage.Used < 10*1024)
}
```

#### Test_MemoryManager_MemoryLimit
```go
// メモリ制限機能
func Test_MemoryManager_MemoryLimit(t *testing.T) {
    // Given: メモリ制限を設定したマネージャー
    mm := NewMemoryManager()
    err := mm.SetMemoryLimit(1024 * 1024) // 1MB
    assert.NoError(t, err)
    
    // When: 制限内でメモリを割り当て
    ptr1, err := mm.Allocate(512 * 1024) // 512KB
    assert.NoError(t, err)
    assert.NotNil(t, ptr1)
    
    ptr2, err := mm.Allocate(256 * 1024) // 256KB
    assert.NoError(t, err)
    assert.NotNil(t, ptr2)
    
    // When: 制限を超えて割り当て
    ptr3, err := mm.Allocate(512 * 1024) // 512KB (total would be 1.25MB)
    
    // Then: エラーが返される
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "memory limit exceeded")
    assert.Nil(t, ptr3)
}
```

#### Test_MemoryManager_WarningCallback
```go
// メモリ警告コールバック
func Test_MemoryManager_WarningCallback(t *testing.T) {
    // Given: メモリマネージャーと警告コールバック
    mm := NewMemoryManager()
    mm.SetMemoryLimit(1024 * 1024) // 1MB
    
    warningTriggered := false
    mm.RegisterMemoryWarningCallback(0.8, func() {
        warningTriggered = true
    })
    
    // When: 80%を超えるメモリを使用
    ptr, err := mm.Allocate(850 * 1024) // 850KB (> 80% of 1MB)
    assert.NoError(t, err)
    assert.NotNil(t, ptr)
    
    // Then: 警告コールバックが呼ばれる
    assert.True(t, warningTriggered)
}
```

### 5. リーク検出テスト

#### Test_MemoryManager_LeakDetection
```go
// メモリリーク検出
func Test_MemoryManager_LeakDetection(t *testing.T) {
    // Given: リーク検出を有効にしたマネージャー
    mm := NewMemoryManager()
    mm.EnableLeakDetection(true)
    
    // When: メモリを割り当てて一部を解放しない
    ptr1, _ := mm.Allocate(100)
    ptr2, _ := mm.Allocate(200)
    ptr3, _ := mm.Allocate(300)
    
    mm.Deallocate(ptr2) // ptr1とptr3は解放しない
    
    // When: リークレポートを取得
    report := mm.GetLeakReport()
    
    // Then: リークが検出される
    assert.Equal(t, 2, report.TotalLeaks)
    assert.Equal(t, uint64(400), report.LeakedBytes) // 100 + 300
    assert.Len(t, report.Leaks, 2)
}
```

### 6. パフォーマンステスト

#### Benchmark_ObjectPool_GetPut
```go
// プール操作のベンチマーク
func Benchmark_ObjectPool_GetPut(b *testing.B) {
    pool := NewObjectPool("BenchPool", 64, 1000)
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            ptr, _ := pool.Get()
            pool.Put(ptr)
        }
    })
}
```

#### Benchmark_MemoryManager_Allocate
```go
// 直接割り当てのベンチマーク
func Benchmark_MemoryManager_Allocate(b *testing.B) {
    mm := NewMemoryManager()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ptr, _ := mm.Allocate(128)
        mm.Deallocate(ptr)
    }
}
```

#### Benchmark_MemoryManager_WithPool
```go
// プール経由割り当てのベンチマーク
func Benchmark_MemoryManager_WithPool(b *testing.B) {
    mm := NewMemoryManager()
    mm.CreatePool("BenchPool", 128, 10000)
    pool, _ := mm.GetPool("BenchPool")
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            ptr, _ := pool.Get()
            pool.Put(ptr)
        }
    })
}
```

### 7. ストレステスト

#### Test_MemoryManager_LongRunning
```go
// 長時間実行テスト
func Test_MemoryManager_LongRunning(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping long-running test")
    }
    
    mm := NewMemoryManager()
    mm.EnableLeakDetection(true)
    
    // 1時間実行
    duration := 1 * time.Hour
    start := time.Now()
    
    for time.Since(start) < duration {
        // ランダムな操作を実行
        size := rand.Intn(1024) + 1
        ptr, err := mm.Allocate(size)
        if err == nil {
            time.Sleep(time.Microsecond)
            mm.Deallocate(ptr)
        }
        
        // 定期的にメトリクスを確認
        if time.Since(start).Seconds() > 0 && 
           int(time.Since(start).Seconds())%60 == 0 {
            metrics := mm.GetMetrics()
            report := mm.GetLeakReport()
            
            // メモリリークチェック
            assert.True(t, report.LeakedBytes < 50*1024*1024) // < 50MB
            
            // 断片化チェック
            assert.True(t, metrics.FragmentationRate < 0.05) // < 5%
        }
    }
}
```

### 8. 統合テスト

#### Test_MemoryManager_WithECS
```go
// ECSシステムとの統合テスト
func Test_MemoryManager_WithECS(t *testing.T) {
    // Given: メモリマネージャーとECSワールド
    mm := NewMemoryManager()
    
    // プールを事前作成
    mm.CreatePool("EntityPool", 64, 10000)
    mm.CreatePool("TransformPool", 48, 10000)
    mm.CreatePool("SpritePool", 32, 10000)
    
    world := ecs.NewWorld(ecs.WithMemoryManager(mm))
    
    // When: 大量のエンティティを作成・削除
    entities := make([]ecs.EntityID, 1000)
    for i := 0; i < 1000; i++ {
        entity := world.CreateEntity()
        world.AddComponent(entity, &TransformComponent{})
        world.AddComponent(entity, &SpriteComponent{})
        entities[i] = entity
    }
    
    // Then: メモリ使用量が適切
    usage := mm.GetMemoryUsage()
    assert.True(t, usage.Pools["EntityPool"].InUse > 0)
    assert.True(t, usage.Pools["TransformPool"].InUse > 0)
    assert.True(t, usage.Pools["SpritePool"].InUse > 0)
    
    // When: エンティティを削除
    for _, entity := range entities {
        world.DestroyEntity(entity)
    }
    
    // Then: メモリが解放される
    usage = mm.GetMemoryUsage()
    assert.Equal(t, 0, usage.Pools["EntityPool"].InUse)
    assert.Equal(t, 0, usage.Pools["TransformPool"].InUse)
    assert.Equal(t, 0, usage.Pools["SpritePool"].InUse)
}
```

## テスト実行順序

1. **基本機能テスト** - ObjectPoolとMemoryManagerの基本動作確認
2. **GC制御テスト** - GC制御機能の動作確認
3. **監視機能テスト** - メモリ監視・制限機能の確認
4. **リーク検出テスト** - メモリリーク検出機能の確認
5. **パフォーマンステスト** - 性能目標の達成確認
6. **ストレステスト** - 長時間安定性の確認
7. **統合テスト** - ECSシステムとの連携確認

## 成功基準

- [ ] 全単体テストが合格
- [ ] メモリ割り当て速度 < 100ns (ベンチマーク確認)
- [ ] プールヒット率 > 95% (メトリクス確認)
- [ ] メモリ断片化率 < 5% (長時間テスト確認)
- [ ] 24時間実行でのリーク < 50MB (ストレステスト確認)
- [ ] GC停止時間 < 1ms (GC統計確認)