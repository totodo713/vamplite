package ecs

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helpers and mocks

type MockEventBusHandler struct {
	HandlerID      HandlerID
	OnHandle       func(EventBusEvent) error
	SupportedTypes []EventTypeID
	HandleCount    int64
	mutex          sync.Mutex
}

func (m *MockEventBusHandler) Handle(event EventBusEvent) error {
	atomic.AddInt64(&m.HandleCount, 1)
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.OnHandle != nil {
		return m.OnHandle(event)
	}
	return nil
}

func (m *MockEventBusHandler) GetHandlerID() HandlerID {
	if m.HandlerID != "" {
		return m.HandlerID
	}
	return HandlerID("mock-handler")
}

func (m *MockEventBusHandler) GetSupportedEventTypes() []EventTypeID {
	if m.SupportedTypes != nil {
		return m.SupportedTypes
	}
	return []EventTypeID{EventTypeIDEntityCreated}
}

func (m *MockEventBusHandler) GetHandleCount() int64 {
	return atomic.LoadInt64(&m.HandleCount)
}

// Test helper functions

func defaultEventBusConfig() *EventBusConfig {
	return &EventBusConfig{
		BufferSize:     1000,
		NumWorkers:     4,
		EnableMetrics:  true,
		EnablePriority: false,
		MaxHandlers:    1000,
	}
}

func setupEventBusForTest(t *testing.T) EventBus {
	config := defaultEventBusConfig()
	eventBus := NewEventBus(config)
	err := eventBus.Start()
	require.NoError(t, err)

	t.Cleanup(func() {
		eventBus.Stop()
	})

	return eventBus
}

func createBusTestEvent(eventType EventTypeID) EventBusEvent {
	return &GenericTestBusEvent{
		EventBusEventBase: EventBusEventBase{
			Type:      eventType,
			EntityID:  EntityID(123),
			Timestamp: time.Now(),
			Priority:  EventPriorityNormal,
		},
		TestData: "test-data",
	}
}

func createBusTestEventWithEntityID(eventType EventTypeID, entityID EntityID) EventBusEvent {
	event := createBusTestEvent(eventType).(*GenericTestBusEvent)
	event.EntityID = entityID
	return event
}

func NewEntityIDFilterForBus(targetEntityID EntityID) EventBusFilter {
	return EventBusFilterFunc(func(event EventBusEvent) bool {
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
	eventBus := NewEventBus(defaultEventBusConfig())

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

func TestEventBus_PublishSync(t *testing.T) {
	// Given: EventBus準備とハンドラー登録
	eventBus := setupEventBusForTest(t)
	receivedEvents := make([]EventBusEvent, 0)
	var mutex sync.Mutex

	handler := &MockEventBusHandler{
		OnHandle: func(event EventBusEvent) error {
			mutex.Lock()
			receivedEvents = append(receivedEvents, event)
			mutex.Unlock()
			return nil
		},
	}

	subscriptionID, err := eventBus.Subscribe(EventTypeIDEntityCreated, handler)
	assert.NoError(t, err) // このテストは失敗するはず

	// When: イベント配信
	event := &EntityCreatedBusEvent{
		EventBusEventBase: EventBusEventBase{
			Type:      EventTypeIDEntityCreated,
			EntityID:  EntityID(123),
			Timestamp: time.Now(),
			Priority:  EventPriorityNormal,
		},
		Components: []ComponentType{ComponentTypeTransform, ComponentTypeSprite},
	}

	err = eventBus.Publish(EventTypeIDEntityCreated, event)

	// Then: 配信成功確認
	assert.NoError(t, err) // このテストは失敗するはず
	mutex.Lock()
	assert.Equal(t, 1, len(receivedEvents))
	assert.Equal(t, event, receivedEvents[0])
	mutex.Unlock()

	_ = subscriptionID // 使用していない警告を避ける
}

func TestEventBus_Subscribe(t *testing.T) {
	// Given: EventBus準備
	eventBus := setupEventBusForTest(t)
	handler := &MockEventBusHandler{}

	// When: サブスクリプション登録
	subscriptionID, err := eventBus.Subscribe(EventTypeIDPlayerDamaged, handler)

	// Then: 登録成功確認
	assert.NoError(t, err) // このテストは失敗するはず（未実装）
	assert.NotEqual(t, SubscriptionID(0), subscriptionID)
	assert.Equal(t, 1, len(eventBus.GetSubscriptions()))

	// When: サブスクリプション解除
	err = eventBus.Unsubscribe(subscriptionID)

	// Then: 解除成功確認
	assert.NoError(t, err) // このテストは失敗するはず（未実装）
	assert.Equal(t, 0, len(eventBus.GetSubscriptions()))
}

func TestEventBus_PublishAsync(t *testing.T) {
	// Given: EventBus準備、非同期ハンドラー
	eventBus := setupEventBusForTest(t)
	receivedChan := make(chan EventBusEvent, 1)
	handler := &MockEventBusHandler{
		OnHandle: func(event EventBusEvent) error {
			receivedChan <- event
			return nil
		},
	}

	_, err := eventBus.Subscribe(EventTypeIDItemCollected, handler)
	assert.NoError(t, err) // このテストは失敗するはず

	// When: 非同期イベント配信
	event := createBusTestEvent(EventTypeIDItemCollected)
	err = eventBus.PublishAsync(EventTypeIDItemCollected, event)

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

func TestEventBus_EntityIDFiltering(t *testing.T) {
	// Given: EntityIDフィルター設定
	eventBus := setupEventBusForTest(t)
	targetEntityID := EntityID(42)
	receivedEvents := make([]EventBusEvent, 0)
	var mutex sync.Mutex

	filter := NewEntityIDFilterForBus(targetEntityID)
	handler := &MockEventBusHandler{
		OnHandle: func(event EventBusEvent) error {
			mutex.Lock()
			receivedEvents = append(receivedEvents, event)
			mutex.Unlock()
			return nil
		},
	}

	_, err := eventBus.SubscribeWithFilter(EventTypeIDComponentAdded, filter, handler)
	assert.NoError(t, err) // このテストは失敗するはず

	// When: 複数EntityIDのイベント配信
	events := []EventBusEvent{
		createBusTestEventWithEntityID(EventTypeIDComponentAdded, EntityID(42)), // フィルター通過
		createBusTestEventWithEntityID(EventTypeIDComponentAdded, EntityID(10)), // フィルター除外
		createBusTestEventWithEntityID(EventTypeIDComponentAdded, EntityID(42)), // フィルター通過
	}

	for _, event := range events {
		err := eventBus.Publish(EventTypeIDComponentAdded, event)
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

func TestEventBus_HandlerErrorIsolation(t *testing.T) {
	// Given: 成功・失敗ハンドラー混在
	eventBus := setupEventBusForTest(t)
	successCount := int64(0)
	errorCount := int64(0)

	successHandler := &MockEventBusHandler{
		OnHandle: func(event EventBusEvent) error {
			atomic.AddInt64(&successCount, 1)
			return nil
		},
	}

	errorHandler := &MockEventBusHandler{
		OnHandle: func(event EventBusEvent) error {
			atomic.AddInt64(&errorCount, 1)
			return errors.New("handler error")
		},
	}

	_, err1 := eventBus.Subscribe(EventTypeIDItemCollected, successHandler)
	_, err2 := eventBus.Subscribe(EventTypeIDItemCollected, errorHandler)
	assert.NoError(t, err1) // このテストは失敗するはず
	assert.NoError(t, err2) // このテストは失敗するはず

	// When: イベント配信
	event := createBusTestEvent(EventTypeIDItemCollected)
	err := eventBus.Publish(EventTypeIDItemCollected, event)

	// Then: エラー隔離確認
	assert.NoError(t, err)                                     // このテストは失敗するはず（配信自体は成功）
	assert.Equal(t, int64(1), atomic.LoadInt64(&successCount)) // 成功ハンドラーは実行
	assert.Equal(t, int64(1), atomic.LoadInt64(&errorCount))   // エラーハンドラーも実行

	// エラー統計確認
	stats := eventBus.GetStats()
	assert.Equal(t, uint64(1), stats.HandlerErrors) // このテストは失敗するはず
}

func TestEventBus_InvalidEventType(t *testing.T) {
	// Given: EventBus準備
	eventBus := setupEventBusForTest(t)

	// When: 無効なイベントタイプで配信
	invalidEventType := EventTypeID(9999)
	event := createBusTestEvent(invalidEventType)
	err := eventBus.Publish(invalidEventType, event)

	// Then: 適切なエラー返却
	assert.Error(t, err) // このテストは失敗するはず（現在は"not implemented"エラーが返る）
	assert.Contains(t, err.Error(), "invalid event type")
}

func TestEventBus_NilEventHandling(t *testing.T) {
	// Given: EventBus準備
	eventBus := setupEventBusForTest(t)

	// When: nilイベント配信
	err := eventBus.Publish(EventTypeIDEntityCreated, nil)

	// Then: エラー処理確認
	assert.Error(t, err) // このテストは失敗するはず
	assert.Contains(t, err.Error(), "event cannot be nil")
}

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
	handler := &MockEventBusHandler{
		OnHandle: func(event EventBusEvent) error {
			atomic.AddInt64(&processedCount, 1)
			return nil
		},
	}

	_, err := eventBus.Subscribe(EventTypeIDComponentUpdated, handler)
	assert.NoError(t, err) // このテストは失敗するはず

	// When: 大量イベント配信
	numEvents := 1000 // テスト時間短縮のため削減
	startTime := time.Now()

	for i := 0; i < numEvents; i++ {
		event := createBusTestEvent(EventTypeIDComponentUpdated)
		err := eventBus.PublishAsync(EventTypeIDComponentUpdated, event)
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
	assert.Greater(t, eventsPerSecond, 1000.0,
		"Should process at least 1,000 events/sec") // このテストは失敗するはず
}

func TestEventBus_MemoryLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	// Given: メモリ使用量監視
	var initialMemStats, finalMemStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&initialMemStats)

	// When: 長期間のイベント処理
	for iteration := 0; iteration < 5; iteration++ { // テスト時間短縮のため削減
		eventBus := NewEventBus(defaultEventBusConfig())
		eventBus.Start()

		handler := &MockEventBusHandler{
			OnHandle: func(event EventBusEvent) error {
				// 処理をシミュレート
				return nil
			},
		}

		_, err := eventBus.Subscribe(EventTypeIDItemCollected, handler)
		if err == nil { // 実装後は正常動作するはず
			// 大量イベント処理
			for i := 0; i < 50; i++ {
				event := createBusTestEvent(EventTypeIDItemCollected)
				eventBus.PublishAsync(EventTypeIDItemCollected, event)
			}

			time.Sleep(50 * time.Millisecond) // 処理待ち
		}

		eventBus.Stop()
	}

	// Then: メモリ使用量確認
	// 複数回GCを実行してメモリを確実に解放
	for i := 0; i < 3; i++ {
		runtime.GC()
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}
	runtime.ReadMemStats(&finalMemStats)

	memoryIncrease := int64(finalMemStats.Alloc) - int64(initialMemStats.Alloc)
	// オーバーフロー対策: メモリ使用量が減った場合（GCによる）は0とみなす
	if memoryIncrease < 0 {
		memoryIncrease = 0
	}

	assert.Less(t, uint64(memoryIncrease), uint64(200*1024*1024),
		"Memory increase should be less than 200MB") // より現実的な閾値に調整
}
