# TASK-203: EventBus実装 - TDDテストケース仕様

## テストケース概要

**実装対象**: EventBus非同期イベント処理システム  
**テスト方針**: 基本機能・非同期処理・エラーハンドリング・統合・パフォーマンス  
**カバレッジ目標**: 95%以上  
**実行環境**: Go 1.22, テストフレームワーク標準testing  

## 基本機能テストケース

### TC-203-001: EventBusインターフェース基本動作

#### TC-203-001-01: EventBus初期化
```go
func TestEventBus_Initialize(t *testing.T) {
    // Given: EventBusConfig設定
    config := &EventBusConfig{
        BufferSize:    1000,
        NumWorkers:    4,
        EnableMetrics: true,
    }
    
    // When: EventBus作成
    eventBus := NewEventBus(config)
    
    // Then: 初期状態確認
    assert.NotNil(t, eventBus)
    assert.False(t, eventBus.IsRunning())
    assert.Equal(t, 0, len(eventBus.GetSubscriptions()))
}
```

#### TC-203-001-02: EventBus起動・停止
```go
func TestEventBus_StartStop(t *testing.T) {
    // Given: EventBus準備
    eventBus := NewEventBus(defaultConfig())
    
    // When: 起動
    err := eventBus.Start()
    
    // Then: 起動成功確認
    assert.NoError(t, err)
    assert.True(t, eventBus.IsRunning())
    
    // When: 停止
    err = eventBus.Stop()
    
    // Then: 停止成功確認  
    assert.NoError(t, err)
    assert.False(t, eventBus.IsRunning())
}
```

### TC-203-002: 基本イベント配信機能

#### TC-203-002-01: 同期イベント配信
```go
func TestEventBus_PublishSync(t *testing.T) {
    // Given: EventBus準備とハンドラー登録
    eventBus := setupEventBus(t)
    receivedEvents := make([]Event, 0)
    handler := &MockEventHandler{
        OnHandle: func(event Event) error {
            receivedEvents = append(receivedEvents, event)
            return nil
        },
    }
    
    subscriptionID, err := eventBus.Subscribe(EventTypeEntityCreated, handler)
    assert.NoError(t, err)
    
    // When: イベント配信
    event := &EntityCreatedEvent{
        EventBase: EventBase{
            Type:      EventTypeEntityCreated,
            EntityID:  EntityID(123),
            Timestamp: time.Now(),
            Priority:  EventPriorityNormal,
        },
        Components: []ComponentType{ComponentTypeTransform, ComponentTypeSprite},
    }
    
    err = eventBus.Publish(EventTypeEntityCreated, event)
    
    // Then: 配信成功確認
    assert.NoError(t, err)
    assert.Equal(t, 1, len(receivedEvents))
    assert.Equal(t, event, receivedEvents[0])
}
```

#### TC-203-002-02: 複数サブスクライバーへの配信
```go
func TestEventBus_PublishMultipleSubscribers(t *testing.T) {
    // Given: 複数ハンドラー登録
    eventBus := setupEventBus(t)
    receivedCounts := make([]int, 3)
    
    for i := 0; i < 3; i++ {
        idx := i
        handler := &MockEventHandler{
            OnHandle: func(event Event) error {
                receivedCounts[idx]++
                return nil
            },
        }
        _, err := eventBus.Subscribe(EventTypeComponentAdded, handler)
        assert.NoError(t, err)
    }
    
    // When: イベント配信
    event := createTestEvent(EventTypeComponentAdded)
    err := eventBus.Publish(EventTypeComponentAdded, event)
    
    // Then: 全ハンドラーに配信確認
    assert.NoError(t, err)
    for i := 0; i < 3; i++ {
        assert.Equal(t, 1, receivedCounts[i], "Handler %d should receive event", i)
    }
}
```

### TC-203-003: サブスクリプション管理

#### TC-203-003-01: サブスクリプション登録・解除
```go
func TestEventBus_SubscriptionManagement(t *testing.T) {
    // Given: EventBus準備
    eventBus := setupEventBus(t)
    handler := &MockEventHandler{}
    
    // When: サブスクリプション登録
    subscriptionID, err := eventBus.Subscribe(EventTypePlayerDamaged, handler)
    
    // Then: 登録成功確認
    assert.NoError(t, err)
    assert.NotEqual(t, SubscriptionID(0), subscriptionID)
    assert.Equal(t, 1, len(eventBus.GetSubscriptions()))
    
    // When: サブスクリプション解除
    err = eventBus.Unsubscribe(subscriptionID)
    
    // Then: 解除成功確認
    assert.NoError(t, err)
    assert.Equal(t, 0, len(eventBus.GetSubscriptions()))
}
```

#### TC-203-003-02: 重複サブスクリプション防止
```go
func TestEventBus_PreventDuplicateSubscription(t *testing.T) {
    // Given: EventBus準備、同一ハンドラー
    eventBus := setupEventBus(t)
    handler := &MockEventHandler{HandlerID: HandlerID("test-handler")}
    
    // When: 同一ハンドラーを複数回登録
    sub1, err1 := eventBus.Subscribe(EventTypeEnemyDefeated, handler)
    sub2, err2 := eventBus.Subscribe(EventTypeEnemyDefeated, handler)
    
    // Then: 重複登録処理確認
    assert.NoError(t, err1)
    assert.NoError(t, err2) // または重複エラー返却
    assert.Equal(t, sub1, sub2) // 同じサブスクリプションIDまたは異なるID
    assert.Equal(t, 1, len(eventBus.GetSubscriptions())) // 重複防止確認
}
```

## 非同期処理テストケース

### TC-203-004: 非同期イベント配信

#### TC-203-004-01: 基本非同期配信
```go
func TestEventBus_PublishAsync(t *testing.T) {
    // Given: EventBus準備、非同期ハンドラー
    eventBus := setupEventBus(t)
    receivedChan := make(chan Event, 1)
    handler := &MockEventHandler{
        OnHandle: func(event Event) error {
            receivedChan <- event
            return nil
        },
    }
    
    _, err := eventBus.Subscribe(EventTypeItemCollected, handler)
    assert.NoError(t, err)
    
    // When: 非同期イベント配信
    event := createTestEvent(EventTypeItemCollected)
    err = eventBus.PublishAsync(EventTypeItemCollected, event)
    
    // Then: 非同期配信成功確認
    assert.NoError(t, err)
    
    // Wait for async processing
    select {
    case receivedEvent := <-receivedChan:
        assert.Equal(t, event, receivedEvent)
    case <-time.After(time.Second):
        t.Fatal("Timeout waiting for async event")
    }
}
```

#### TC-203-004-02: バックプレッシャー処理
```go
func TestEventBus_Backpressure(t *testing.T) {
    // Given: 小さなバッファサイズのEventBus
    config := &EventBusConfig{
        BufferSize: 2, // 意図的に小さく設定
        NumWorkers: 1,
    }
    eventBus := NewEventBus(config)
    eventBus.Start()
    defer eventBus.Stop()
    
    blockingHandler := &MockEventHandler{
        OnHandle: func(event Event) error {
            time.Sleep(100 * time.Millisecond) // 処理遅延
            return nil
        },
    }
    
    _, err := eventBus.Subscribe(EventTypePlayerDamaged, blockingHandler)
    assert.NoError(t, err)
    
    // When: バッファを超えるイベント配信
    for i := 0; i < 10; i++ {
        event := createTestEvent(EventTypePlayerDamaged)
        err := eventBus.PublishAsync(EventTypePlayerDamaged, event)
        
        if i < 2 {
            // バッファ内は成功
            assert.NoError(t, err)
        } else {
            // バックプレッシャー適用
            // エラー返却またはブロッキング動作確認
            if err != nil {
                assert.Equal(t, ErrQueueFull, err)
            }
        }
    }
}
```

### TC-203-005: ワーカープール動作

#### TC-203-005-01: 並列イベント処理
```go
func TestEventBus_WorkerPoolConcurrency(t *testing.T) {
    // Given: マルチワーカー設定
    config := &EventBusConfig{
        BufferSize: 1000,
        NumWorkers: 4,
    }
    eventBus := NewEventBus(config)
    eventBus.Start()
    defer eventBus.Stop()
    
    processedEvents := &sync.Map{}
    handler := &MockEventHandler{
        OnHandle: func(event Event) error {
            goroutineID := getGoroutineID() // テスト用ヘルパー
            processedEvents.Store(event.GetEntityID(), goroutineID)
            time.Sleep(10 * time.Millisecond) // 処理時間シミュレート
            return nil
        },
    }
    
    _, err := eventBus.Subscribe(EventTypeComponentUpdated, handler)
    assert.NoError(t, err)
    
    // When: 大量イベント並列配信
    numEvents := 100
    for i := 0; i < numEvents; i++ {
        event := createTestEventWithEntityID(EventTypeComponentUpdated, EntityID(i))
        err := eventBus.PublishAsync(EventTypeComponentUpdated, event)
        assert.NoError(t, err)
    }
    
    // Then: 並列処理確認
    time.Sleep(2 * time.Second) // 処理完了待ち
    
    goroutineIDs := make(map[int]bool)
    processedEvents.Range(func(key, value interface{}) bool {
        goroutineID := value.(int)
        goroutineIDs[goroutineID] = true
        return true
    })
    
    // 複数goroutineで処理されたことを確認
    assert.True(t, len(goroutineIDs) > 1, "Events should be processed by multiple goroutines")
}
```

## イベントフィルタリングテストケース

### TC-203-006: フィルタリング機能

#### TC-203-006-01: EntityIDフィルタリング
```go
func TestEventBus_EntityIDFiltering(t *testing.T) {
    // Given: EntityIDフィルター設定
    eventBus := setupEventBus(t)
    targetEntityID := EntityID(42)
    receivedEvents := make([]Event, 0)
    
    filter := NewEntityIDFilter(targetEntityID)
    handler := &MockEventHandler{
        OnHandle: func(event Event) error {
            receivedEvents = append(receivedEvents, event)
            return nil
        },
    }
    
    _, err := eventBus.SubscribeWithFilter(EventTypeComponentAdded, filter, handler)
    assert.NoError(t, err)
    
    // When: 複数EntityIDのイベント配信
    events := []Event{
        createTestEventWithEntityID(EventTypeComponentAdded, EntityID(42)), // フィルター通過
        createTestEventWithEntityID(EventTypeComponentAdded, EntityID(10)), // フィルター除外
        createTestEventWithEntityID(EventTypeComponentAdded, EntityID(42)), // フィルター通過
    }
    
    for _, event := range events {
        err := eventBus.Publish(EventTypeComponentAdded, event)
        assert.NoError(t, err)
    }
    
    // Then: フィルター結果確認
    assert.Equal(t, 2, len(receivedEvents))
    for _, event := range receivedEvents {
        assert.Equal(t, targetEntityID, event.GetEntityID())
    }
}
```

#### TC-203-006-02: カスタムフィルター
```go
func TestEventBus_CustomFilter(t *testing.T) {
    // Given: カスタムフィルター（高優先度のみ通過）
    eventBus := setupEventBus(t)
    receivedEvents := make([]Event, 0)
    
    highPriorityFilter := EventFilterFunc(func(event Event) bool {
        return event.GetPriority() == EventPriorityHigh
    })
    
    handler := &MockEventHandler{
        OnHandle: func(event Event) error {
            receivedEvents = append(receivedEvents, event)
            return nil
        },
    }
    
    _, err := eventBus.SubscribeWithFilter(EventTypePlayerDamaged, highPriorityFilter, handler)
    assert.NoError(t, err)
    
    // When: 異なる優先度のイベント配信
    events := []Event{
        createTestEventWithPriority(EventTypePlayerDamaged, EventPriorityLow),    // 除外
        createTestEventWithPriority(EventTypePlayerDamaged, EventPriorityHigh),   // 通過
        createTestEventWithPriority(EventTypePlayerDamaged, EventPriorityNormal), // 除外
        createTestEventWithPriority(EventTypePlayerDamaged, EventPriorityHigh),   // 通過
    }
    
    for _, event := range events {
        err := eventBus.Publish(EventTypePlayerDamaged, event)
        assert.NoError(t, err)
    }
    
    // Then: 高優先度イベントのみ受信確認
    assert.Equal(t, 2, len(receivedEvents))
    for _, event := range receivedEvents {
        assert.Equal(t, EventPriorityHigh, event.GetPriority())
    }
}
```

## 優先度・順序保証テストケース  

### TC-203-007: イベント優先度処理

#### TC-203-007-01: 優先度キュー動作
```go
func TestEventBus_PriorityQueue(t *testing.T) {
    // Given: 優先度キュー有効なEventBus
    config := &EventBusConfig{
        BufferSize:      1000,
        NumWorkers:      1, // 順序確認のため1つに制限
        EnablePriority:  true,
    }
    eventBus := NewEventBus(config)
    eventBus.Start()
    defer eventBus.Stop()
    
    processedOrder := make([]EventPriority, 0)
    handler := &MockEventHandler{
        OnHandle: func(event Event) error {
            processedOrder = append(processedOrder, event.GetPriority())
            return nil
        },
    }
    
    _, err := eventBus.Subscribe(EventTypeSystemStarted, handler)
    assert.NoError(t, err)
    
    // When: 異なる優先度のイベント配信（逆順）
    priorities := []EventPriority{
        EventPriorityLow,
        EventPriorityHigh,
        EventPriorityNormal,
        EventPriorityCritical,
    }
    
    for _, priority := range priorities {
        event := createTestEventWithPriority(EventTypeSystemStarted, priority)
        err := eventBus.PublishAsync(EventTypeSystemStarted, event)
        assert.NoError(t, err)
    }
    
    time.Sleep(500 * time.Millisecond) // 処理完了待ち
    
    // Then: 優先度順の処理確認
    expectedOrder := []EventPriority{
        EventPriorityCritical,
        EventPriorityHigh,
        EventPriorityNormal,
        EventPriorityLow,
    }
    assert.Equal(t, expectedOrder, processedOrder)
}
```

#### TC-203-007-02: 同一優先度内FIFO順序
```go
func TestEventBus_SamePriorityFIFO(t *testing.T) {
    // Given: 同一優先度イベントの順序確認用
    eventBus := setupEventBusWithWorkers(1) // 順序保証のため単一ワーカー
    processedEntityIDs := make([]EntityID, 0)
    
    handler := &MockEventHandler{
        OnHandle: func(event Event) error {
            processedEntityIDs = append(processedEntityIDs, event.GetEntityID())
            return nil
        },
    }
    
    _, err := eventBus.Subscribe(EventTypeEnemyDefeated, handler)
    assert.NoError(t, err)
    
    // When: 同一優先度の連続イベント配信
    expectedOrder := []EntityID{100, 101, 102, 103, 104}
    for _, entityID := range expectedOrder {
        event := createTestEventWithPriorityAndEntity(
            EventTypeEnemyDefeated, 
            EventPriorityNormal, 
            entityID,
        )
        err := eventBus.PublishAsync(EventTypeEnemyDefeated, event)
        assert.NoError(t, err)
    }
    
    time.Sleep(200 * time.Millisecond)
    
    // Then: FIFO順序確認
    assert.Equal(t, expectedOrder, processedEntityIDs)
}
```

## エラーハンドリングテストケース

### TC-203-008: ハンドラーエラー処理

#### TC-203-008-01: ハンドラーエラー隔離
```go
func TestEventBus_HandlerErrorIsolation(t *testing.T) {
    // Given: 成功・失敗ハンドラー混在
    eventBus := setupEventBus(t)
    successCount := 0
    errorCount := 0
    
    successHandler := &MockEventHandler{
        OnHandle: func(event Event) error {
            successCount++
            return nil
        },
    }
    
    errorHandler := &MockEventHandler{
        OnHandle: func(event Event) error {
            errorCount++
            return errors.New("handler error")
        },
    }
    
    _, err1 := eventBus.Subscribe(EventTypeItemCollected, successHandler)
    _, err2 := eventBus.Subscribe(EventTypeItemCollected, errorHandler)
    assert.NoError(t, err1)
    assert.NoError(t, err2)
    
    // When: イベント配信
    event := createTestEvent(EventTypeItemCollected)
    err := eventBus.Publish(EventTypeItemCollected, event)
    
    // Then: エラー隔離確認
    assert.NoError(t, err) // 配信自体は成功
    assert.Equal(t, 1, successCount) // 成功ハンドラーは実行
    assert.Equal(t, 1, errorCount)   // エラーハンドラーも実行
    
    // エラー統計確認
    stats := eventBus.GetStats()
    assert.Equal(t, uint64(1), stats.HandlerErrors)
}
```

#### TC-203-008-02: ハンドラーpanic回復
```go
func TestEventBus_HandlerPanicRecovery(t *testing.T) {
    // Given: panicハンドラー
    eventBus := setupEventBus(t)
    normalExecuted := false
    
    panicHandler := &MockEventHandler{
        OnHandle: func(event Event) error {
            panic("test panic")
        },
    }
    
    normalHandler := &MockEventHandler{
        OnHandle: func(event Event) error {
            normalExecuted = true
            return nil
        },
    }
    
    _, err1 := eventBus.Subscribe(EventTypePlayerDamaged, panicHandler)
    _, err2 := eventBus.Subscribe(EventTypePlayerDamaged, normalHandler)
    assert.NoError(t, err1)
    assert.NoError(t, err2)
    
    // When: イベント配信
    event := createTestEvent(EventTypePlayerDamaged)
    err := eventBus.Publish(EventTypePlayerDamaged, event)
    
    // Then: panic回復とサービス継続確認
    assert.NoError(t, err) // 配信は成功
    assert.True(t, normalExecuted) // 他のハンドラーは実行
    
    stats := eventBus.GetStats()
    assert.Equal(t, uint64(1), stats.HandlerPanics)
}
```

### TC-203-009: 無効データ処理

#### TC-203-009-01: 無効イベントタイプ
```go
func TestEventBus_InvalidEventType(t *testing.T) {
    // Given: EventBus準備
    eventBus := setupEventBus(t)
    
    // When: 無効なイベントタイプで配信
    invalidEventType := EventType(9999)
    event := createTestEvent(invalidEventType)
    err := eventBus.Publish(invalidEventType, event)
    
    // Then: 適切なエラー返却
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid event type")
}
```

#### TC-203-009-02: nilイベント処理
```go
func TestEventBus_NilEventHandling(t *testing.T) {
    // Given: EventBus準備
    eventBus := setupEventBus(t)
    
    // When: nilイベント配信
    err := eventBus.Publish(EventTypeEntityCreated, nil)
    
    // Then: エラー処理確認
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "event cannot be nil")
}
```

## パフォーマンステストケース

### TC-203-010: スループット・レイテンシ

#### TC-203-010-01: 高スループットテスト
```go
func TestEventBus_HighThroughput(t *testing.T) {
    // Given: 高性能設定EventBus
    config := &EventBusConfig{
        BufferSize: 10000,
        NumWorkers: 8,
    }
    eventBus := NewEventBus(config)
    eventBus.Start()
    defer eventBus.Stop()
    
    processedCount := int64(0)
    handler := &MockEventHandler{
        OnHandle: func(event Event) error {
            atomic.AddInt64(&processedCount, 1)
            return nil
        },
    }
    
    _, err := eventBus.Subscribe(EventTypeComponentUpdated, handler)
    assert.NoError(t, err)
    
    // When: 大量イベント配信
    numEvents := 10000
    startTime := time.Now()
    
    for i := 0; i < numEvents; i++ {
        event := createTestEvent(EventTypeComponentUpdated)
        err := eventBus.PublishAsync(EventTypeComponentUpdated, event)
        assert.NoError(t, err)
    }
    
    // 処理完了待ち
    for atomic.LoadInt64(&processedCount) < int64(numEvents) {
        time.Sleep(10 * time.Millisecond)
    }
    duration := time.Since(startTime)
    
    // Then: スループット確認
    eventsPerSecond := float64(numEvents) / duration.Seconds()
    assert.Greater(t, eventsPerSecond, 10000.0, "Should process at least 10,000 events/sec")
}
```

#### TC-203-010-02: レイテンシ測定
```go
func TestEventBus_Latency(t *testing.T) {
    // Given: レイテンシ測定用設定
    eventBus := setupEventBus(t)
    latencies := make([]time.Duration, 0, 1000)
    var mu sync.Mutex
    
    handler := &MockEventHandler{
        OnHandle: func(event Event) error {
            receiveTime := time.Now()
            sendTime := event.GetTimestamp()
            latency := receiveTime.Sub(sendTime)
            
            mu.Lock()
            latencies = append(latencies, latency)
            mu.Unlock()
            return nil
        },
    }
    
    _, err := eventBus.Subscribe(EventTypePlayerDamaged, handler)
    assert.NoError(t, err)
    
    // When: イベント配信とレイテンシ測定
    for i := 0; i < 1000; i++ {
        event := createTestEventWithTimestamp(EventTypePlayerDamaged, time.Now())
        err := eventBus.PublishAsync(EventTypePlayerDamaged, event)
        assert.NoError(t, err)
    }
    
    time.Sleep(2 * time.Second) // 処理完了待ち
    
    // Then: レイテンシ分析
    mu.Lock()
    assert.Equal(t, 1000, len(latencies))
    
    // 平均レイテンシ計算
    var totalLatency time.Duration
    for _, latency := range latencies {
        totalLatency += latency
    }
    avgLatency := totalLatency / time.Duration(len(latencies))
    mu.Unlock()
    
    assert.Less(t, avgLatency, time.Millisecond, "Average latency should be less than 1ms")
}
```

### TC-203-011: メモリ使用量テスト

#### TC-203-011-01: メモリリークテスト  
```go
func TestEventBus_MemoryLeak(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping memory leak test in short mode")
    }
    
    // Given: メモリ使用量監視
    var initialMemStats, finalMemStats runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&initialMemStats)
    
    // When: 長期間のイベント処理
    for iteration := 0; iteration < 100; iteration++ {
        eventBus := setupEventBus(t)
        
        handler := &MockEventHandler{
            OnHandle: func(event Event) error {
                // 処理をシミュレート
                return nil
            },
        }
        
        _, err := eventBus.Subscribe(EventTypeItemCollected, handler)
        assert.NoError(t, err)
        
        // 大量イベント処理
        for i := 0; i < 1000; i++ {
            event := createTestEvent(EventTypeItemCollected)
            err := eventBus.PublishAsync(EventTypeItemCollected, event)
            assert.NoError(t, err)
        }
        
        time.Sleep(100 * time.Millisecond) // 処理待ち
        eventBus.Stop()
    }
    
    // Then: メモリ使用量確認
    runtime.GC()
    runtime.ReadMemStats(&finalMemStats)
    
    memoryIncrease := finalMemStats.Alloc - initialMemStats.Alloc
    assert.Less(t, memoryIncrease, uint64(50*1024*1024), "Memory increase should be less than 50MB")
}
```

## 統合テストケース

### TC-203-012: ECS統合テスト

#### TC-203-012-01: EntityManagerとの統合
```go
func TestEventBus_EntityManagerIntegration(t *testing.T) {
    // Given: EntityManagerとEventBusの統合環境
    entityManager := setupEntityManager(t)
    eventBus := setupEventBus(t)
    
    createdEvents := make([]EntityCreatedEvent, 0)
    destroyedEvents := make([]EntityDestroyedEvent, 0)
    
    createdHandler := &MockEventHandler{
        OnHandle: func(event Event) error {
            if e, ok := event.(*EntityCreatedEvent); ok {
                createdEvents = append(createdEvents, *e)
            }
            return nil
        },
    }
    
    destroyedHandler := &MockEventHandler{
        OnHandle: func(event Event) error {
            if e, ok := event.(*EntityDestroyedEvent); ok {
                destroyedEvents = append(destroyedEvents, *e)
            }
            return nil
        },
    }
    
    _, err1 := eventBus.Subscribe(EventTypeEntityCreated, createdHandler)
    _, err2 := eventBus.Subscribe(EventTypeEntityDestroyed, destroyedHandler)
    assert.NoError(t, err1)
    assert.NoError(t, err2)
    
    // EntityManagerにEventBus設定
    entityManager.SetEventBus(eventBus)
    
    // When: エンティティライフサイクル操作
    entityID1 := entityManager.CreateEntity()
    entityID2 := entityManager.CreateEntity()
    entityManager.DestroyEntity(entityID1)
    
    time.Sleep(100 * time.Millisecond) // イベント処理待ち
    
    // Then: 統合動作確認
    assert.Equal(t, 2, len(createdEvents))
    assert.Equal(t, 1, len(destroyedEvents))
    assert.Equal(t, entityID1, destroyedEvents[0].EntityID)
}
```

#### TC-203-012-02: SystemManagerとの統合
```go
func TestEventBus_SystemManagerIntegration(t *testing.T) {
    // Given: SystemManagerとEventBus統合
    systemManager := setupSystemManager(t)
    eventBus := setupEventBus(t)
    
    systemEvents := make([]Event, 0)
    handler := &MockEventHandler{
        OnHandle: func(event Event) error {
            systemEvents = append(systemEvents, event)
            return nil
        },
    }
    
    _, err := eventBus.Subscribe(EventTypeSystemStarted, handler)
    assert.NoError(t, err)
    
    systemManager.SetEventBus(eventBus)
    
    // When: システム操作
    movementSystem := &MockSystem{Name: "MovementSystem"}
    err = systemManager.RegisterSystem(movementSystem)
    assert.NoError(t, err)
    
    err = systemManager.StartSystem("MovementSystem")
    assert.NoError(t, err)
    
    time.Sleep(100 * time.Millisecond)
    
    // Then: システムイベント確認
    assert.GreaterOrEqual(t, len(systemEvents), 1)
    // システム開始イベントの詳細確認
}
```

## モック・ヘルパー実装

### MockEventHandler
```go
type MockEventHandler struct {
    HandlerID     HandlerID
    OnHandle      func(Event) error
    SupportedTypes []EventType
}

func (m *MockEventHandler) Handle(event Event) error {
    if m.OnHandle != nil {
        return m.OnHandle(event)
    }
    return nil
}

func (m *MockEventHandler) GetHandlerID() HandlerID {
    if m.HandlerID != "" {
        return m.HandlerID
    }
    return HandlerID("mock-handler")
}

func (m *MockEventHandler) GetSupportedEventTypes() []EventType {
    if m.SupportedTypes != nil {
        return m.SupportedTypes
    }
    return []EventType{EventTypeEntityCreated} // デフォルト
}
```

### テストヘルパー関数
```go
func setupEventBus(t *testing.T) EventBus {
    config := &EventBusConfig{
        BufferSize:    1000,
        NumWorkers:    4,
        EnableMetrics: true,
        EnablePriority: false,
    }
    eventBus := NewEventBus(config)
    err := eventBus.Start()
    require.NoError(t, err)
    
    t.Cleanup(func() {
        eventBus.Stop()
    })
    
    return eventBus
}

func createTestEvent(eventType EventType) Event {
    return &GenericTestEvent{
        EventBase: EventBase{
            Type:      eventType,
            EntityID:  EntityID(rand.Intn(1000)),
            Timestamp: time.Now(),
            Priority:  EventPriorityNormal,
        },
    }
}

func createTestEventWithEntityID(eventType EventType, entityID EntityID) Event {
    event := createTestEvent(eventType)
    event.(*GenericTestEvent).EntityID = entityID
    return event
}

func createTestEventWithPriority(eventType EventType, priority EventPriority) Event {
    event := createTestEvent(eventType)
    event.(*GenericTestEvent).Priority = priority
    return event
}
```

## テスト実行・品質基準

### カバレッジ目標
- **行カバレッジ**: 95%以上
- **分岐カバレッジ**: 90%以上
- **関数カバレッジ**: 100%

### パフォーマンス基準
- **スループット**: 10,000 events/秒以上
- **レイテンシ**: 平均 < 1ms、99%tile < 5ms
- **メモリ使用量**: 長期実行でリーク < 50MB

### 実行コマンド
```bash
# 全テスト実行
go test -v ./internal/core/ecs/...

# カバレッジ付きテスト
go test -v -coverprofile=coverage.out ./internal/core/ecs/...
go tool cover -html=coverage.out

# ベンチマーク実行
go test -v -bench=. -benchmem ./internal/core/ecs/...

# レースコンディション検出
go test -v -race ./internal/core/ecs/...

# メモリリークテスト
go test -v -memprofile=mem.out ./internal/core/ecs/...
```

---

**作成日**: 2025-08-07  
**更新日**: 2025-08-07  
**レビュー状態**: 初版作成完了  
**承認者**: TBD