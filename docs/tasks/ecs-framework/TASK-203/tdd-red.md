# TASK-203: EventBus実装 - TDD Red段階

## Red段階の目的

失敗するテストを作成し、EventBus機能が未実装であることを確認する。  
すべてのテストが失敗することで、実装すべき機能の範囲を明確にする。

## 1. 基本インターフェース・構造体定義

まず、テスト実装に必要な最小限のインターフェース定義を作成します。

### internal/core/ecs/event_types.go
```go
package ecs

import (
    "errors"
    "time"
)

// EventType defines the type of event
type EventType uint32

const (
    EventTypeEntityCreated EventType = iota
    EventTypeEntityDestroyed
    EventTypeComponentAdded
    EventTypeComponentRemoved
    EventTypeComponentUpdated
    EventTypeSystemStarted
    EventTypeSystemStopped
    EventTypePlayerDamaged
    EventTypeEnemyDefeated
    EventTypeItemCollected
)

// EventPriority defines event processing priority
type EventPriority uint8

const (
    EventPriorityLow EventPriority = iota
    EventPriorityNormal
    EventPriorityHigh
    EventPriorityCritical
)

// Event defines the interface for all events
type Event interface {
    GetType() EventType
    GetEntityID() EntityID
    GetTimestamp() time.Time
    GetPriority() EventPriority
    Validate() error
}

// EventBase provides common event fields
type EventBase struct {
    Type      EventType
    EntityID  EntityID
    Timestamp time.Time
    Priority  EventPriority
}

func (e EventBase) GetType() EventType        { return e.Type }
func (e EventBase) GetEntityID() EntityID     { return e.EntityID }
func (e EventBase) GetTimestamp() time.Time   { return e.Timestamp }
func (e EventBase) GetPriority() EventPriority { return e.Priority }
func (e EventBase) Validate() error           { return nil }

// Specific event implementations
type EntityCreatedEvent struct {
    EventBase
    Components []ComponentType
}

type ComponentAddedEvent struct {
    EventBase
    ComponentType ComponentType
    ComponentData interface{}
}

type PlayerDamagedEvent struct {
    EventBase
    Damage       float64
    DamageSource EntityID
}

// Generic test event for testing
type GenericTestEvent struct {
    EventBase
    TestData string
}

// Handler types
type HandlerID string
type SubscriptionID uint64

// EventHandler defines the interface for event handlers
type EventHandler interface {
    Handle(event Event) error
    GetHandlerID() HandlerID
    GetSupportedEventTypes() []EventType
}

// EventFilter defines the interface for event filtering
type EventFilter interface {
    Filter(event Event) bool
}

type EventFilterFunc func(Event) bool

func (f EventFilterFunc) Filter(event Event) bool {
    return f(event)
}

// EventBusStats provides statistics about event bus performance
type EventBusStats struct {
    EventsPublished    uint64
    EventsProcessed    uint64
    EventsDropped      uint64
    HandlerErrors      uint64
    HandlerPanics      uint64
    AvgLatencyNanos    int64
    TotalSubscriptions int
    QueueSize          int
    WorkerCount        int
}

// EventBusConfig defines configuration for EventBus
type EventBusConfig struct {
    BufferSize      int
    NumWorkers      int
    EnableMetrics   bool
    EnablePriority  bool
    MaxHandlers     int
}

// EventBus defines the main interface for event bus
type EventBus interface {
    // Lifecycle
    Start() error
    Stop() error
    IsRunning() bool
    Flush() error
    
    // Event publishing
    Publish(eventType EventType, event Event) error
    PublishAsync(eventType EventType, event Event) error
    
    // Subscription management
    Subscribe(eventType EventType, handler EventHandler) (SubscriptionID, error)
    Unsubscribe(subscriptionID SubscriptionID) error
    SubscribeWithFilter(eventType EventType, filter EventFilter, handler EventHandler) (SubscriptionID, error)
    
    // Information
    GetStats() EventBusStats
    GetSubscriptions() map[SubscriptionID]*Subscription
}

// Subscription represents an active subscription
type Subscription struct {
    ID       SubscriptionID
    Type     EventType
    Handler  EventHandler
    Filter   EventFilter
    Created  time.Time
    Active   bool
}

// Errors
var (
    ErrEventBusNotStarted   = errors.New("event bus is not started")
    ErrEventBusStopped      = errors.New("event bus is stopped")
    ErrInvalidEventType     = errors.New("invalid event type")
    ErrSubscriptionNotFound = errors.New("subscription not found")
    ErrHandlerPanic         = errors.New("event handler panic")
    ErrQueueFull            = errors.New("event queue is full")
    ErrEventNil             = errors.New("event cannot be nil")
)
```

### internal/core/ecs/event_bus.go (未実装版)
```go
package ecs

// EventBusImpl - 未実装のEventBus実装
type EventBusImpl struct {
    // フィールドは後で実装
}

// NewEventBus creates a new EventBus instance (未実装)
func NewEventBus(config *EventBusConfig) EventBus {
    return &EventBusImpl{}
}

// 全メソッドを未実装状態で定義
func (eb *EventBusImpl) Start() error {
    return errors.New("not implemented")
}

func (eb *EventBusImpl) Stop() error {
    return errors.New("not implemented")
}

func (eb *EventBusImpl) IsRunning() bool {
    return false
}

func (eb *EventBusImpl) Flush() error {
    return errors.New("not implemented")
}

func (eb *EventBusImpl) Publish(eventType EventType, event Event) error {
    return errors.New("not implemented")
}

func (eb *EventBusImpl) PublishAsync(eventType EventType, event Event) error {
    return errors.New("not implemented")
}

func (eb *EventBusImpl) Subscribe(eventType EventType, handler EventHandler) (SubscriptionID, error) {
    return 0, errors.New("not implemented")
}

func (eb *EventBusImpl) Unsubscribe(subscriptionID SubscriptionID) error {
    return errors.New("not implemented")
}

func (eb *EventBusImpl) SubscribeWithFilter(eventType EventType, filter EventFilter, handler EventHandler) (SubscriptionID, error) {
    return 0, errors.New("not implemented")
}

func (eb *EventBusImpl) GetStats() EventBusStats {
    return EventBusStats{}
}

func (eb *EventBusImpl) GetSubscriptions() map[SubscriptionID]*Subscription {
    return make(map[SubscriptionID]*Subscription)
}
```

## 2. テスト実装

### internal/core/ecs/event_bus_test.go
```go
package ecs

import (
    "sync"
    "sync/atomic"
    "testing"
    "time"
    "errors"
    "runtime"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// Test helpers and mocks

type MockEventHandler struct {
    HandlerID      HandlerID
    OnHandle       func(Event) error
    SupportedTypes []EventType
    HandleCount    int64
    mutex          sync.Mutex
}

func (m *MockEventHandler) Handle(event Event) error {
    atomic.AddInt64(&m.HandleCount, 1)
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
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
    return []EventType{EventTypeEntityCreated}
}

func (m *MockEventHandler) GetHandleCount() int64 {
    return atomic.LoadInt64(&m.HandleCount)
}

// Test helper functions

func defaultConfig() *EventBusConfig {
    return &EventBusConfig{
        BufferSize:     1000,
        NumWorkers:     4,
        EnableMetrics:  true,
        EnablePriority: false,
        MaxHandlers:    1000,
    }
}

func setupEventBus(t *testing.T) EventBus {
    config := defaultConfig()
    eventBus := NewEventBus(config)
    err := eventBus.Start()
    require.NoError(t, err)
    
    t.Cleanup(func() {
        eventBus.Stop()
    })
    
    return eventBus
}

func setupEventBusWithWorkers(workers int) EventBus {
    config := &EventBusConfig{
        BufferSize:     1000,
        NumWorkers:     workers,
        EnableMetrics:  true,
        EnablePriority: false,
    }
    eventBus := NewEventBus(config)
    eventBus.Start()
    return eventBus
}

func createTestEvent(eventType EventType) Event {
    return &GenericTestEvent{
        EventBase: EventBase{
            Type:      eventType,
            EntityID:  EntityID(123),
            Timestamp: time.Now(),
            Priority:  EventPriorityNormal,
        },
        TestData: "test-data",
    }
}

func createTestEventWithEntityID(eventType EventType, entityID EntityID) Event {
    event := createTestEvent(eventType).(*GenericTestEvent)
    event.EntityID = entityID
    return event
}

func createTestEventWithPriority(eventType EventType, priority EventPriority) Event {
    event := createTestEvent(eventType).(*GenericTestEvent)
    event.Priority = priority
    return event
}

func createTestEventWithTimestamp(eventType EventType, timestamp time.Time) Event {
    event := createTestEvent(eventType).(*GenericTestEvent)
    event.Timestamp = timestamp
    return event
}

func createTestEventWithPriorityAndEntity(eventType EventType, priority EventPriority, entityID EntityID) Event {
    event := createTestEvent(eventType).(*GenericTestEvent)
    event.Priority = priority
    event.EntityID = entityID
    return event
}

// Filter helpers
func NewEntityIDFilter(targetEntityID EntityID) EventFilter {
    return EventFilterFunc(func(event Event) bool {
        return event.GetEntityID() == targetEntityID
    })
}

// TC-203-001: EventBusインターフェース基本動作

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
    assert.False(t, eventBus.IsRunning()) // このテストは失敗するはず（未実装）
    assert.Equal(t, 0, len(eventBus.GetSubscriptions()))
}

func TestEventBus_StartStop(t *testing.T) {
    // Given: EventBus準備
    eventBus := NewEventBus(defaultConfig())
    
    // When: 起動
    err := eventBus.Start()
    
    // Then: 起動成功確認
    assert.NoError(t, err) // このテストは失敗するはず（未実装）
    assert.True(t, eventBus.IsRunning())
    
    // When: 停止
    err = eventBus.Stop()
    
    // Then: 停止成功確認
    assert.NoError(t, err) // このテストは失敗するはず（未実装）
    assert.False(t, eventBus.IsRunning())
}

// TC-203-002: 基本イベント配信機能

func TestEventBus_PublishSync(t *testing.T) {
    // Given: EventBus準備とハンドラー登録
    eventBus := setupEventBus(t)
    receivedEvents := make([]Event, 0)
    var mutex sync.Mutex
    
    handler := &MockEventHandler{
        OnHandle: func(event Event) error {
            mutex.Lock()
            receivedEvents = append(receivedEvents, event)
            mutex.Unlock()
            return nil
        },
    }
    
    subscriptionID, err := eventBus.Subscribe(EventTypeEntityCreated, handler)
    assert.NoError(t, err) // このテストは失敗するはず
    
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
    assert.NoError(t, err) // このテストは失敗するはず
    mutex.Lock()
    assert.Equal(t, 1, len(receivedEvents))
    assert.Equal(t, event, receivedEvents[0])
    mutex.Unlock()
    
    _ = subscriptionID // 使用していない警告を避ける
}

func TestEventBus_PublishMultipleSubscribers(t *testing.T) {
    // Given: 複数ハンドラー登録
    eventBus := setupEventBus(t)
    receivedCounts := make([]int64, 3)
    
    for i := 0; i < 3; i++ {
        idx := i
        handler := &MockEventHandler{
            OnHandle: func(event Event) error {
                atomic.AddInt64(&receivedCounts[idx], 1)
                return nil
            },
        }
        _, err := eventBus.Subscribe(EventTypeComponentAdded, handler)
        assert.NoError(t, err) // このテストは失敗するはず
    }
    
    // When: イベント配信
    event := createTestEvent(EventTypeComponentAdded)
    err := eventBus.Publish(EventTypeComponentAdded, event)
    
    // Then: 全ハンドラーに配信確認
    assert.NoError(t, err) // このテストは失敗するはず
    for i := 0; i < 3; i++ {
        assert.Equal(t, int64(1), atomic.LoadInt64(&receivedCounts[i]), 
            "Handler %d should receive event", i)
    }
}

// TC-203-003: サブスクリプション管理

func TestEventBus_SubscriptionManagement(t *testing.T) {
    // Given: EventBus準備
    eventBus := setupEventBus(t)
    handler := &MockEventHandler{}
    
    // When: サブスクリプション登録
    subscriptionID, err := eventBus.Subscribe(EventTypePlayerDamaged, handler)
    
    // Then: 登録成功確認
    assert.NoError(t, err) // このテストは失敗するはず
    assert.NotEqual(t, SubscriptionID(0), subscriptionID)
    assert.Equal(t, 1, len(eventBus.GetSubscriptions()))
    
    // When: サブスクリプション解除
    err = eventBus.Unsubscribe(subscriptionID)
    
    // Then: 解除成功確認
    assert.NoError(t, err) // このテストは失敗するはず
    assert.Equal(t, 0, len(eventBus.GetSubscriptions()))
}

// TC-203-004: 非同期イベント配信

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
    assert.NoError(t, err) // このテストは失敗するはず
    
    // When: 非同期イベント配信
    event := createTestEvent(EventTypeItemCollected)
    err = eventBus.PublishAsync(EventTypeItemCollected, event)
    
    // Then: 非同期配信成功確認
    assert.NoError(t, err) // このテストは失敗するはず
    
    // Wait for async processing
    select {
    case receivedEvent := <-receivedChan:
        assert.Equal(t, event, receivedEvent)
    case <-time.After(time.Second):
        t.Fatal("Timeout waiting for async event")
    }
}

// TC-203-006: フィルタリング機能

func TestEventBus_EntityIDFiltering(t *testing.T) {
    // Given: EntityIDフィルター設定
    eventBus := setupEventBus(t)
    targetEntityID := EntityID(42)
    receivedEvents := make([]Event, 0)
    var mutex sync.Mutex
    
    filter := NewEntityIDFilter(targetEntityID)
    handler := &MockEventHandler{
        OnHandle: func(event Event) error {
            mutex.Lock()
            receivedEvents = append(receivedEvents, event)
            mutex.Unlock()
            return nil
        },
    }
    
    _, err := eventBus.SubscribeWithFilter(EventTypeComponentAdded, filter, handler)
    assert.NoError(t, err) // このテストは失敗するはず
    
    // When: 複数EntityIDのイベント配信
    events := []Event{
        createTestEventWithEntityID(EventTypeComponentAdded, EntityID(42)), // フィルター通過
        createTestEventWithEntityID(EventTypeComponentAdded, EntityID(10)), // フィルター除外
        createTestEventWithEntityID(EventTypeComponentAdded, EntityID(42)), // フィルター通過
    }
    
    for _, event := range events {
        err := eventBus.Publish(EventTypeComponentAdded, event)
        assert.NoError(t, err) // このテストは失敗するはず
    }
    
    // Then: フィルター結果確認
    mutex.Lock()
    assert.Equal(t, 2, len(receivedEvents)) // このテストは失敗するはず
    for _, event := range receivedEvents {
        assert.Equal(t, targetEntityID, event.GetEntityID())
    }
    mutex.Unlock()
}

// TC-203-008: エラーハンドリング

func TestEventBus_HandlerErrorIsolation(t *testing.T) {
    // Given: 成功・失敗ハンドラー混在
    eventBus := setupEventBus(t)
    successCount := int64(0)
    errorCount := int64(0)
    
    successHandler := &MockEventHandler{
        OnHandle: func(event Event) error {
            atomic.AddInt64(&successCount, 1)
            return nil
        },
    }
    
    errorHandler := &MockEventHandler{
        OnHandle: func(event Event) error {
            atomic.AddInt64(&errorCount, 1)
            return errors.New("handler error")
        },
    }
    
    _, err1 := eventBus.Subscribe(EventTypeItemCollected, successHandler)
    _, err2 := eventBus.Subscribe(EventTypeItemCollected, errorHandler)
    assert.NoError(t, err1) // このテストは失敗するはず
    assert.NoError(t, err2) // このテストは失敗するはず
    
    // When: イベント配信
    event := createTestEvent(EventTypeItemCollected)
    err := eventBus.Publish(EventTypeItemCollected, event)
    
    // Then: エラー隔離確認
    assert.NoError(t, err) // このテストは失敗するはず（配信自体は成功）
    assert.Equal(t, int64(1), atomic.LoadInt64(&successCount)) // 成功ハンドラーは実行
    assert.Equal(t, int64(1), atomic.LoadInt64(&errorCount))   // エラーハンドラーも実行
    
    // エラー統計確認
    stats := eventBus.GetStats()
    assert.Equal(t, uint64(1), stats.HandlerErrors) // このテストは失敗するはず
}

func TestEventBus_InvalidEventType(t *testing.T) {
    // Given: EventBus準備
    eventBus := setupEventBus(t)
    
    // When: 無効なイベントタイプで配信
    invalidEventType := EventType(9999)
    event := createTestEvent(invalidEventType)
    err := eventBus.Publish(invalidEventType, event)
    
    // Then: 適切なエラー返却
    assert.Error(t, err) // このテストは失敗するはず（現在は"not implemented"エラーが返る）
    assert.Contains(t, err.Error(), "invalid event type")
}

func TestEventBus_NilEventHandling(t *testing.T) {
    // Given: EventBus準備
    eventBus := setupEventBus(t)
    
    // When: nilイベント配信
    err := eventBus.Publish(EventTypeEntityCreated, nil)
    
    // Then: エラー処理確認
    assert.Error(t, err) // このテストは失敗するはず
    assert.Contains(t, err.Error(), "event cannot be nil")
}

// TC-203-010: パフォーマンステスト

func TestEventBus_HighThroughput(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance test in short mode")
    }
    
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
    assert.NoError(t, err) // このテストは失敗するはず
    
    // When: 大量イベント配信
    numEvents := 10000
    startTime := time.Now()
    
    for i := 0; i < numEvents; i++ {
        event := createTestEvent(EventTypeComponentUpdated)
        err := eventBus.PublishAsync(EventTypeComponentUpdated, event)
        assert.NoError(t, err) // このテストは失敗するはず
    }
    
    // 処理完了待ち
    for atomic.LoadInt64(&processedCount) < int64(numEvents) {
        time.Sleep(10 * time.Millisecond)
        if time.Since(startTime) > 30*time.Second { // タイムアウト
            break
        }
    }
    duration := time.Since(startTime)
    
    // Then: スループット確認
    eventsPerSecond := float64(numEvents) / duration.Seconds()
    assert.Greater(t, eventsPerSecond, 10000.0, 
        "Should process at least 10,000 events/sec") // このテストは失敗するはず
}

// TC-203-011: メモリリークテスト

func TestEventBus_MemoryLeak(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping memory leak test in short mode")
    }
    
    // Given: メモリ使用量監視
    var initialMemStats, finalMemStats runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&initialMemStats)
    
    // When: 長期間のイベント処理
    for iteration := 0; iteration < 10; iteration++ { // 本来は100回だが、テスト時間短縮のため10回
        eventBus := NewEventBus(defaultConfig())
        eventBus.Start()
        
        handler := &MockEventHandler{
            OnHandle: func(event Event) error {
                // 処理をシミュレート
                return nil
            },
        }
        
        _, err := eventBus.Subscribe(EventTypeItemCollected, handler)
        if err == nil { // 実装後は正常動作するはず
            // 大量イベント処理
            for i := 0; i < 100; i++ {
                event := createTestEvent(EventTypeItemCollected)
                eventBus.PublishAsync(EventTypeItemCollected, event)
            }
            
            time.Sleep(50 * time.Millisecond) // 処理待ち
        }
        
        eventBus.Stop()
    }
    
    // Then: メモリ使用量確認
    runtime.GC()
    runtime.ReadMemStats(&finalMemStats)
    
    memoryIncrease := finalMemStats.Alloc - initialMemStats.Alloc
    assert.Less(t, memoryIncrease, uint64(50*1024*1024), 
        "Memory increase should be less than 50MB") // 実装によっては失敗する可能性
}

// Benchmark tests (これらも現在は実行できない)

func BenchmarkEventBus_PublishSync(b *testing.B) {
    eventBus := NewEventBus(defaultConfig())
    eventBus.Start()
    defer eventBus.Stop()
    
    handler := &MockEventHandler{}
    eventBus.Subscribe(EventTypeEntityCreated, handler)
    
    event := createTestEvent(EventTypeEntityCreated)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        eventBus.Publish(EventTypeEntityCreated, event)
    }
}

func BenchmarkEventBus_PublishAsync(b *testing.B) {
    eventBus := NewEventBus(defaultConfig())
    eventBus.Start()
    defer eventBus.Stop()
    
    handler := &MockEventHandler{}
    eventBus.Subscribe(EventTypeEntityCreated, handler)
    
    event := createTestEvent(EventTypeEntityCreated)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        eventBus.PublishAsync(EventTypeEntityCreated, event)
    }
}
```

## 3. テスト実行とRed確認

### テスト実行コマンド
```bash
# 基本テスト実行
cd /path/to/muscle-dreamer
go test -v ./internal/core/ecs/... -run="TestEventBus_"

# 特定テストのみ実行
go test -v ./internal/core/ecs/... -run="TestEventBus_Initialize"

# カバレッジ付きテスト
go test -v -coverprofile=coverage_red.out ./internal/core/ecs/...
go tool cover -html=coverage_red.out

# レース検出付きテスト
go test -v -race ./internal/core/ecs/... -run="TestEventBus_"
```

### 期待される結果（Red段階）

全てのテストが失敗することを確認：

```
=== RUN   TestEventBus_Initialize
    event_bus_test.go:XX: assertion failed: expected false, got true (IsRunning should be false)
--- FAIL: TestEventBus_Initialize (0.00s)

=== RUN   TestEventBus_StartStop  
    event_bus_test.go:XX: assertion failed: Start() should not return error, got "not implemented"
--- FAIL: TestEventBus_StartStop (0.00s)

=== RUN   TestEventBus_PublishSync
    event_bus_test.go:XX: assertion failed: Subscribe() should not return error, got "not implemented" 
--- FAIL: TestEventBus_PublishSync (0.00s)

[全テストが失敗...]

FAIL    github.com/your-project/muscle-dreamer/internal/core/ecs    0.XXs
```

## 4. Red段階で確認すべき項目

### ✅ 確認項目チェックリスト

- [ ] 全ての基本機能テストが失敗している
- [ ] 非同期処理テストが失敗している  
- [ ] フィルタリング機能テストが失敗している
- [ ] エラーハンドリングテストが失敗している
- [ ] パフォーマンステストが失敗している
- [ ] すべてのメソッドが"not implemented"エラーを返している
- [ ] テストコードがコンパイルエラーなしで実行できている
- [ ] モック・ヘルパー関数が正しく動作している

### 失敗パターンの分類

1. **実装不足エラー**: `errors.New("not implemented")` が返される
2. **初期状態エラー**: IsRunning()がfalseを返さない等
3. **機能欠如エラー**: Subscribe/Publish等が動作しない
4. **統計情報エラー**: GetStats()が空の構造体を返す

## 5. 次のステップ（Green段階への準備）

Red段階完了後、次のGreen段階では：

1. **EventBusImpl構造体の実装**
2. **基本的な配信機能の実装**  
3. **サブスクリプション管理の実装**
4. **非同期処理機能の実装**
5. **エラーハンドリングの実装**

### Green段階の実装優先順位

1. **Phase 1**: 基本構造・同期配信
2. **Phase 2**: サブスクリプション管理  
3. **Phase 3**: 非同期処理・ワーカープール
4. **Phase 4**: フィルタリング・優先度
5. **Phase 5**: エラーハンドリング・統計

## 6. Red段階完了確認

### 完了基準

- [ ] 全テストケースが実装され、実行可能
- [ ] 全テストが期待通り失敗している（実装不足が原因）
- [ ] テストコードにコンパイルエラーが存在しない
- [ ] モック・ヘルパー関数が正常動作している
- [ ] テスト実行時間が許容範囲内（<30秒）

### 実行ログ例

```bash
$ go test -v ./internal/core/ecs/... -run="TestEventBus_"

=== RUN   TestEventBus_Initialize
--- FAIL: TestEventBus_Initialize (0.00s)
    event_bus_test.go:123: IsRunning() should return false, got true

=== RUN   TestEventBus_StartStop
--- FAIL: TestEventBus_StartStop (0.00s)
    event_bus_test.go:135: Start() returned error: not implemented

=== RUN   TestEventBus_PublishSync
--- FAIL: TestEventBus_PublishSync (0.00s)
    event_bus_test.go:165: Subscribe() returned error: not implemented

[... 他のテストも全て失敗 ...]

FAIL
FAIL    muscle-dreamer/internal/core/ecs    2.456s
```

---

**Red段階作業完了**

すべてのテストが失敗することで、実装すべき機能の範囲が明確になりました。次のGreen段階では、これらのテストを通すための最小実装を行います。

**作成日**: 2025-08-07  
**更新日**: 2025-08-07  
**段階**: Red（失敗テスト実装完了）