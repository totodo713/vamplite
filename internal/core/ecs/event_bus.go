package ecs

import (
	"sync"
	"time"
)

// EventBusImpl - EventBus実装（TDD Green段階）
type EventBusImpl struct {
	config    *EventBusConfig
	isRunning bool
	mutex     sync.RWMutex

	// サブスクリプション管理
	subscriptions       map[SubscriptionID]*EventBusSubscription
	subscriptionsByType map[EventTypeID][]*EventBusSubscription
	nextSubscriptionID  SubscriptionID

	// 非同期処理用
	eventQueue chan *queuedEvent
	workerWG   sync.WaitGroup
	stopChan   chan struct{}

	// 統計情報（atomic操作用）
	statsAtomic *EventBusStatsAtomic

	// メモリプール
	eventPool sync.Pool
}

// queuedEvent represents an event in the async queue
type queuedEvent struct {
	EventType EventTypeID
	Event     EventBusEvent
}

// NewEventBus creates a new EventBus instance
func NewEventBus(config *EventBusConfig) EventBus {
	if config == nil {
		config = DefaultEventBusConfig()
	}

	return &EventBusImpl{
		config:              config,
		isRunning:           false,
		subscriptions:       make(map[SubscriptionID]*EventBusSubscription),
		subscriptionsByType: make(map[EventTypeID][]*EventBusSubscription),
		nextSubscriptionID:  0,
		statsAtomic:         NewEventBusStatsAtomic(),
		eventPool: sync.Pool{
			New: func() interface{} {
				return &queuedEvent{}
			},
		},
	}
}

// 全メソッドを未実装状態で定義（TDD Red段階）

func (eb *EventBusImpl) Start() error {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if eb.isRunning {
		return nil // 既に起動済み
	}

	eb.isRunning = true
	eb.stopChan = make(chan struct{})
	eb.eventQueue = make(chan *queuedEvent, eb.config.BufferSize)

	// ワーカーゴルーチン起動
	for i := 0; i < eb.config.NumWorkers; i++ {
		eb.workerWG.Add(1)
		go eb.worker()
	}

	return nil
}

func (eb *EventBusImpl) Stop() error {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if !eb.isRunning {
		return nil // 既に停止済み
	}

	eb.isRunning = false
	close(eb.stopChan)
	close(eb.eventQueue)

	eb.workerWG.Wait()
	return nil
}

func (eb *EventBusImpl) IsRunning() bool {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	return eb.isRunning
}

func (eb *EventBusImpl) Flush() error {
	eb.mutex.RLock()
	if !eb.isRunning {
		eb.mutex.RUnlock()
		return ErrEventBusNotStarted
	}

	queueSize := len(eb.eventQueue)
	eb.mutex.RUnlock()

	// 既存のキューが空になるまで待機
	for queueSize > 0 {
		time.Sleep(10 * time.Millisecond)
		eb.mutex.RLock()
		queueSize = len(eb.eventQueue)
		eb.mutex.RUnlock()
	}

	return nil
}

func (eb *EventBusImpl) Publish(eventType EventTypeID, event EventBusEvent) error {
	if event == nil {
		return ErrEventNil
	}

	// イベントタイプの有効性チェック
	if !isValidEventType(eventType) {
		return ErrInvalidEventType
	}

	eb.mutex.RLock()
	if !eb.isRunning {
		eb.mutex.RUnlock()
		return ErrEventBusNotStarted
	}

	subscriptions := eb.subscriptionsByType[eventType]
	eb.mutex.RUnlock()

	if eb.config.EnableMetrics {
		eb.statsAtomic.eventsPublished.Add(1)
	}

	for _, subscription := range subscriptions {
		if subscription.Filter != nil && !subscription.Filter.Filter(event) {
			continue
		}

		if err := subscription.Handler.Handle(event); err != nil {
			if eb.config.EnableMetrics {
				eb.statsAtomic.handlerErrors.Add(1)
			}
		} else {
			if eb.config.EnableMetrics {
				eb.statsAtomic.eventsProcessed.Add(1)
			}
		}
	}

	return nil
}

func (eb *EventBusImpl) PublishAsync(eventType EventTypeID, event EventBusEvent) error {
	if event == nil {
		return ErrEventNil
	}

	// イベントタイプの有効性チェック
	if !isValidEventType(eventType) {
		return ErrInvalidEventType
	}

	eb.mutex.RLock()
	if !eb.isRunning {
		eb.mutex.RUnlock()
		return ErrEventBusNotStarted
	}
	eb.mutex.RUnlock()

	queuedEvt := eb.eventPool.Get().(*queuedEvent)
	queuedEvt.EventType = eventType
	queuedEvt.Event = event

	select {
	case eb.eventQueue <- queuedEvt:
		if eb.config.EnableMetrics {
			eb.statsAtomic.eventsPublished.Add(1)
		}
		return nil
	default:
		if eb.config.EnableMetrics {
			eb.statsAtomic.eventsDropped.Add(1)
		}
		return ErrQueueFull
	}
}

func (eb *EventBusImpl) Subscribe(eventType EventTypeID, handler EventBusHandler) (SubscriptionID, error) {
	return eb.SubscribeWithFilter(eventType, nil, handler)
}

func (eb *EventBusImpl) Unsubscribe(subscriptionID SubscriptionID) error {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	subscription, exists := eb.subscriptions[subscriptionID]
	if !exists {
		return ErrSubscriptionNotFound
	}

	// subscriptionsから削除
	delete(eb.subscriptions, subscriptionID)

	// subscriptionsByTypeからも削除
	eventType := subscription.Type
	subscriptions := eb.subscriptionsByType[eventType]
	for i, sub := range subscriptions {
		if sub.ID == subscriptionID {
			eb.subscriptionsByType[eventType] = append(subscriptions[:i], subscriptions[i+1:]...)
			break
		}
	}

	return nil
}

func (eb *EventBusImpl) SubscribeWithFilter(eventType EventTypeID, filter EventBusFilter, handler EventBusHandler) (SubscriptionID, error) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if !eb.isRunning {
		return 0, ErrEventBusNotStarted
	}

	eb.nextSubscriptionID++
	subscriptionID := eb.nextSubscriptionID

	subscription := &EventBusSubscription{
		ID:      subscriptionID,
		Type:    eventType,
		Handler: handler,
		Filter:  filter,
		Created: time.Now(),
		Active:  true,
	}

	eb.subscriptions[subscriptionID] = subscription
	eb.subscriptionsByType[eventType] = append(eb.subscriptionsByType[eventType], subscription)

	return subscriptionID, nil
}

func (eb *EventBusImpl) GetStats() EventBusStats {
	eb.mutex.RLock()
	totalSubs := len(eb.subscriptions)
	queueSize := 0
	if eb.eventQueue != nil {
		queueSize = len(eb.eventQueue)
	}
	workerCount := eb.config.NumWorkers
	eb.mutex.RUnlock()

	return eb.statsAtomic.ToStats(totalSubs, queueSize, workerCount)
}

func (eb *EventBusImpl) GetSubscriptions() map[SubscriptionID]*EventBusSubscription {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()

	// コピーして返す
	result := make(map[SubscriptionID]*EventBusSubscription)
	for id, sub := range eb.subscriptions {
		result[id] = sub
	}
	return result
}

// worker processes events from the async queue
func (eb *EventBusImpl) worker() {
	defer eb.workerWG.Done()

	for {
		select {
		case queuedEvt, ok := <-eb.eventQueue:
			if !ok {
				return // チャネルが閉じられた
			}
			eb.processQueuedEvent(queuedEvt)
		case <-eb.stopChan:
			return // 停止要求
		}
	}
}

// processQueuedEvent handles a single queued event
func (eb *EventBusImpl) processQueuedEvent(queuedEvt *queuedEvent) {
	defer func() {
		// メモリプールにオブジェクトを戻す
		queuedEvt.Event = nil // 参照を切る
		eb.eventPool.Put(queuedEvt)
	}()

	eb.mutex.RLock()
	subscriptions := eb.subscriptionsByType[queuedEvt.EventType]
	eb.mutex.RUnlock()

	for _, subscription := range subscriptions {
		if subscription.Filter != nil && !subscription.Filter.Filter(queuedEvt.Event) {
			continue
		}
		if err := subscription.Handler.Handle(queuedEvt.Event); err != nil {
			if eb.config.EnableMetrics {
				eb.statsAtomic.handlerErrors.Add(1)
			}
		} else {
			if eb.config.EnableMetrics {
				eb.statsAtomic.eventsProcessed.Add(1)
			}
		}
	}
}

// isValidEventType checks if the given event type is valid
func isValidEventType(eventType EventTypeID) bool {
	// Green段階では簡単な範囲チェック
	return eventType <= EventTypeIDItemCollected
}
