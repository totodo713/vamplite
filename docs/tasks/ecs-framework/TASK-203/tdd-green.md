# TASK-203: EventBus実装 - TDD Green段階

## Green段階概要

Red段階で作成した失敗するテストを、最小限の実装でパスさせる段階です。
過剰な実装や最適化は行わず、テストが通る最小限のコードを実装します。

## 実装対象テスト

以下のテストを成功させるための最小実装を行います：

### 基本機能テスト
- `TestEventBus_Initialize` ✅（すでに成功）
- `TestEventBus_StartStop` - Start/Stop機能の基本実装
- `TestEventBus_PublishSync` - 同期イベント配信の基本実装  
- `TestEventBus_Subscribe` - サブスクリプション管理の基本実装

### 拡張機能テスト（順次対応）
- `TestEventBus_PublishAsync` - 非同期配信実装
- `TestEventBus_EntityIDFiltering` - フィルタリング実装
- `TestEventBus_HandlerErrorIsolation` - エラー分離実装
- `TestEventBus_InvalidEventType` - エラーハンドリング実装
- `TestEventBus_NilEventHandling` - Nilイベント対応
- `TestEventBus_HighThroughput` - 高スループット対応（簡易版）
- `TestEventBus_MemoryLeak` - メモリ管理実装（簡易版）

## 実装戦略

### Phase 1: 基本ライフサイクル実装
1. **EventBusImpl構造体の実装**
   - 必要な内部フィールドを追加
   - 基本状態管理

2. **Start/Stop実装**
   - IsRunning状態管理
   - 基本的な開始・停止処理

### Phase 2: サブスクリプション管理実装
1. **Subscribe/Unsubscribe実装**
   - サブスクリプション保存と管理
   - 一意なSubscriptionIDの生成

2. **GetSubscriptions実装**
   - アクティブなサブスクリプション一覧の提供

### Phase 3: イベント配信実装
1. **Publish（同期）実装**
   - 登録されたハンドラーへのイベント配信
   - 基本的なエラーハンドリング

2. **PublishAsync（非同期）実装**  
   - ワーカーゴルーチンベースの配信
   - チャネルベースのキュー実装

### Phase 4: フィルタリングとエラーハンドリング実装
1. **SubscribeWithFilter実装**
   - フィルター条件付きサブスクリプション

2. **エラー処理実装**
   - 各種バリデーション
   - 統計情報収集

## 実装詳細

### 1. EventBusImpl構造体の拡張

```go
type EventBusImpl struct {
    config      *EventBusConfig
    isRunning   bool
    mutex       sync.RWMutex
    
    // サブスクリプション管理
    subscriptions     map[SubscriptionID]*EventBusSubscription
    subscriptionsByType map[EventTypeID][]*EventBusSubscription
    nextSubscriptionID  SubscriptionID
    
    // 非同期処理用
    eventQueue  chan *queuedEvent
    workerWG    sync.WaitGroup
    stopChan    chan struct{}
    
    // 統計情報
    stats EventBusStats
}

type queuedEvent struct {
    EventType EventTypeID
    Event     EventBusEvent
}
```

### 2. 基本メソッドの実装

#### Start/Stop実装
```go
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
```

#### Subscribe/Unsubscribe実装
```go
func (eb *EventBusImpl) Subscribe(eventType EventTypeID, handler EventBusHandler) (SubscriptionID, error) {
    return eb.SubscribeWithFilter(eventType, nil, handler)
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
```

#### Publish実装
```go
func (eb *EventBusImpl) Publish(eventType EventTypeID, event EventBusEvent) error {
    if event == nil {
        return ErrEventNil
    }
    
    eb.mutex.RLock()
    if !eb.isRunning {
        eb.mutex.RUnlock()
        return ErrEventBusNotStarted
    }
    
    subscriptions := eb.subscriptionsByType[eventType]
    eb.mutex.RUnlock()
    
    if eb.config.EnableMetrics {
        eb.stats.EventsPublished++
    }
    
    for _, subscription := range subscriptions {
        if subscription.Filter != nil && !subscription.Filter.Filter(event) {
            continue
        }
        
        if err := subscription.Handler.Handle(event); err != nil {
            if eb.config.EnableMetrics {
                eb.stats.HandlerErrors++
            }
        } else {
            if eb.config.EnableMetrics {
                eb.stats.EventsProcessed++
            }
        }
    }
    
    return nil
}

func (eb *EventBusImpl) PublishAsync(eventType EventTypeID, event EventBusEvent) error {
    if event == nil {
        return ErrEventNil
    }
    
    eb.mutex.RLock()
    if !eb.isRunning {
        eb.mutex.RUnlock()
        return ErrEventBusNotStarted
    }
    eb.mutex.RUnlock()
    
    queuedEvt := &queuedEvent{
        EventType: eventType,
        Event:     event,
    }
    
    select {
    case eb.eventQueue <- queuedEvt:
        if eb.config.EnableMetrics {
            eb.stats.EventsPublished++
        }
        return nil
    default:
        if eb.config.EnableMetrics {
            eb.stats.EventsDropped++
        }
        return ErrQueueFull
    }
}
```

#### Worker実装
```go
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

func (eb *EventBusImpl) processQueuedEvent(queuedEvt *queuedEvent) {
    eb.mutex.RLock()
    subscriptions := eb.subscriptionsByType[queuedEvt.EventType]
    eb.mutex.RUnlock()
    
    for _, subscription := range subscriptions {
        if subscription.Filter != nil && !subscription.Filter.Filter(queuedEvt.Event) {
            continue
        }
        
        if err := subscription.Handler.Handle(queuedEvt.Event); err != nil {
            if eb.config.EnableMetrics {
                eb.stats.HandlerErrors++
            }
        } else {
            if eb.config.EnableMetrics {
                eb.stats.EventsProcessed++
            }
        }
    }
}
```

## 実装順序

1. **構造体フィールドの追加**
2. **NewEventBus関数の更新**
3. **Start/Stop/IsRunningメソッドの実装**
4. **Subscribe/Unsubscribeメソッドの実装**
5. **Publish（同期）メソッドの実装**
6. **PublishAsync（非同期）とWorkerの実装**
7. **エラーハンドリングとバリデーションの追加**
8. **GetStats/GetSubscriptionsメソッドの実装**

## テスト実行順序

各実装完了後に対応するテストを実行し、段階的に成功させていきます：

1. `TestEventBus_StartStop`
2. `TestEventBus_Subscribe`  
3. `TestEventBus_PublishSync`
4. `TestEventBus_PublishAsync`
5. `TestEventBus_EntityIDFiltering`
6. 残りのテスト

## 制限事項（Green段階）

- パフォーマンスの最適化は行わない
- 複雑なエラー回復は実装しない
- 詳細な統計情報収集は最小限
- メモリ使用量の最適化は行わない

これらの最適化は、Refactor段階で実装します。

## 成功基準

- 全ての基本テストケースがパス
- 「not implemented」エラーが解消
- コードが最小限で理解しやすい
- 次のRefactor段階への準備ができている

## 次の段階

Green段階完了後、tdd-refactor.md でコードの改善と最適化を行います。