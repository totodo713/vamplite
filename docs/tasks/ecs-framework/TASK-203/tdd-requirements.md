# TASK-203: EventBus実装 - TDD要件定義

## タスク概要

**実装対象**: 非同期イベント処理システム（EventBus）  
**要件リンク**: REQ-201, REQ-202  
**依存タスク**: TASK-202（MemoryManager実装）完了済み  
**実装タイプ**: TDDプロセス  

## 機能要件

### F-203-001: 基本イベント配信機能
- **要件**: イベント送信者と受信者の疎結合な通信
- **詳細**: 
  - イベントタイプに基づく配信
  - 複数の受信者への同時配信
  - 型安全なイベント処理
- **実装場所**: `internal/core/ecs/event_bus.go`

### F-203-002: 非同期イベント処理
- **要件**: メインスレッドをブロックしない非同期配信
- **詳細**:
  - チャンネルベースの非同期通信
  - バッファリング機能
  - バックプレッシャー対応
- **実装場所**: `internal/core/ecs/event_bus.go`

### F-203-003: イベントフィルタリング・ルーティング
- **要件**: 条件に基づくイベントの選択的配信
- **詳細**:
  - EntityIDによるフィルタリング
  - イベントタイプによるフィルタリング  
  - カスタムフィルター関数対応
- **実装場所**: `internal/core/ecs/event_filter.go`

### F-203-004: サブスクリプション管理
- **要件**: 動的なイベント受信者の登録・解除
- **詳細**:
  - サブスクライバー自動登録
  - 登録解除（unsubscribe）機能
  - 重複登録の防止
- **実装場所**: `internal/core/ecs/subscription_manager.go`

### F-203-005: イベント優先度・順序保証
- **要件**: 重要なイベントの優先処理と順序保証
- **詳細**:
  - 優先度キューによるイベント処理
  - 同一優先度内での順序保証（FIFO）
  - システム間の依存関係を考慮した順序制御
- **実装場所**: `internal/core/ecs/priority_queue.go`

## 非機能要件

### NFR-203-001: パフォーマンス要件
- **目標レイテンシ**: イベント配信 < 1ms
- **目標スループット**: 10,000 events/秒
- **メモリ使用量**: バッファサイズの動的調整
- **CPU使用率**: イベント処理のオーバーヘッド < 5%

### NFR-203-002: スケーラビリティ要件
- **サブスクライバー数**: 最大1,000個のイベントハンドラー
- **イベントタイプ数**: 最大100種類のイベント型
- **バッファサイズ**: 設定可能（デフォルト1,000件）

### NFR-203-003: 信頼性要件
- **イベントロス率**: 0%（バックプレッシャー適用）
- **エラー隔離**: 1つのハンドラーエラーが他に影響しない
- **メモリリーク**: 24時間連続稼働でリークなし

### NFR-203-004: 保守性要件
- **コードカバレッジ**: 95%以上
- **ログ出力**: イベント配信・エラーの詳細ログ
- **監視機能**: イベント処理統計の収集

## インターフェース設計

### EventBusインターフェース
```go
type EventBus interface {
    // イベント配信
    Publish(eventType EventType, event Event) error
    PublishAsync(eventType EventType, event Event) error
    
    // サブスクリプション管理
    Subscribe(eventType EventType, handler EventHandler) (SubscriptionID, error)
    Unsubscribe(subscriptionID SubscriptionID) error
    
    // フィルタリング
    SubscribeWithFilter(eventType EventType, filter EventFilter, handler EventHandler) (SubscriptionID, error)
    
    // 制御
    Start() error
    Stop() error
    Flush() error
    
    // 統計情報
    GetStats() EventBusStats
}
```

### Eventインターフェース
```go
type Event interface {
    GetType() EventType
    GetEntityID() EntityID
    GetTimestamp() time.Time
    GetPriority() EventPriority
    Validate() error
}
```

### EventHandlerインターフェース
```go
type EventHandler interface {
    Handle(event Event) error
    GetHandlerID() HandlerID
    GetSupportedEventTypes() []EventType
}
```

## データ構造設計

### イベントタイプ定義
```go
type EventType uint32

const (
    EventTypeEntityCreated EventType = iota
    EventTypeEntityDestroyed
    EventTypeComponentAdded
    EventTypeComponentRemoved
    EventTypeComponentUpdated
    EventTypeSystemStarted
    EventTypeSystemStopped
    // ゲーム固有のイベント
    EventTypePlayerDamaged
    EventTypeEnemyDefeated
    EventTypeItemCollected
)
```

### 具体的イベント実装
```go
// ECS基本イベント
type EntityCreatedEvent struct {
    EventBase
    EntityID EntityID
    Components []ComponentType
}

type ComponentAddedEvent struct {
    EventBase
    EntityID EntityID
    ComponentType ComponentType
    ComponentData interface{}
}

// ゲームイベント
type PlayerDamagedEvent struct {
    EventBase
    PlayerEntity EntityID
    Damage float64
    DamageSource EntityID
    DamageType DamageType
}
```

## 実装クラス設計

### EventBusImpl構造体
```go
type EventBusImpl struct {
    // イベントキューと配信
    eventQueue    chan *QueuedEvent
    subscribers   map[EventType][]Subscriber
    subscriptions map[SubscriptionID]*Subscription
    
    // 制御
    running       atomic.Bool
    workerPool    *WorkerPool
    ctx           context.Context
    cancel        context.CancelFunc
    
    // 統計・監視
    stats         EventBusStats
    logger        *log.Logger
    
    // 設定
    config        *EventBusConfig
    
    // 同期
    mutex         sync.RWMutex
}
```

### WorkerPool設計
```go
type WorkerPool struct {
    workers    []*Worker
    workQueue  chan Work
    numWorkers int
    ctx        context.Context
}

type Worker struct {
    id         int
    workQueue  chan Work
    workerPool chan chan Work
    ctx        context.Context
}

type Work struct {
    Event     Event
    Handler   EventHandler
    Timestamp time.Time
}
```

## エラーハンドリング設計

### エラータイプ定義
```go
var (
    ErrEventBusNotStarted    = errors.New("event bus is not started")
    ErrEventBusStopped       = errors.New("event bus is stopped")
    ErrInvalidEventType      = errors.New("invalid event type")
    ErrSubscriptionNotFound  = errors.New("subscription not found")
    ErrHandlerPanic          = errors.New("event handler panic")
    ErrQueueFull             = errors.New("event queue is full")
)
```

### エラー処理方針
- **配信エラー**: ログ出力＋統計に記録、他のハンドラーに影響なし
- **ハンドラーpanicエラー**: recover＋ログ出力、サービス継続
- **キュー満杯エラー**: バックプレッシャー適用、配信待ちまたは破棄

## テスト要件

### 単体テスト要件

#### 基本機能テスト
- [ ] イベント配信・受信の基本動作
- [ ] 複数サブスクライバーへの配信
- [ ] サブスクリプション登録・解除
- [ ] イベントフィルタリング機能
- [ ] 優先度キューの動作

#### 非同期処理テスト
- [ ] 非同期配信の動作確認
- [ ] バックプレッシャー処理
- [ ] ワーカープールの動作
- [ ] 同時配信の安全性

#### エラーハンドリングテスト
- [ ] ハンドラーエラー時の隔離
- [ ] ハンドラーpanicからの回復
- [ ] 無効なイベント処理
- [ ] キュー満杯時の処理

### 統合テスト要件

#### システム統合テスト
- [ ] ECSコアシステムとの統合
- [ ] MemoryManagerとの連携
- [ ] SystemManagerとの連携

#### パフォーマンステスト
- [ ] 大量イベント処理（10,000 events/秒）
- [ ] レイテンシ測定（< 1ms）
- [ ] メモリ使用量監視
- [ ] 長期実行安定性テスト（24時間）

### ストレステスト要件
- [ ] 大量サブスクライバー（1,000個）
- [ ] 大量イベントタイプ（100種類）  
- [ ] 高頻度配信（並行1,000 goroutine）
- [ ] メモリ制限下での動作

## 実装フェーズ

### Phase 1: 基本構造 (1日目)
- EventBusインターフェース定義
- 基本イベント構造体実装
- サブスクリプション管理機能
- 基本配信機能（同期）

### Phase 2: 非同期処理 (2日目) 
- ワーカープール実装
- 非同期配信機能
- バックプレッシャー制御
- エラーハンドリング強化

### Phase 3: 高度機能 (3日目)
- イベントフィルタリング
- 優先度キュー
- 統計・監視機能
- パフォーマンス最適化

## 受け入れ基準

### 機能完成基準
- [ ] 全インターフェース実装完了
- [ ] 基本イベント配信動作確認
- [ ] 非同期処理安全性確保
- [ ] フィルタリング機能動作確認
- [ ] エラー隔離機能動作確認

### 品質基準
- [ ] 単体テストカバレッジ95%以上
- [ ] 統合テスト全通過
- [ ] パフォーマンステスト目標達成
- [ ] メモリリークテスト通過
- [ ] ストレステスト通過

### ドキュメント基準
- [ ] API仕様書作成
- [ ] 使用例・サンプルコード作成
- [ ] パフォーマンス特性ドキュメント
- [ ] トラブルシューティングガイド

## リスク要因と対策

### 技術リスク
- **並行アクセス競合**: mutex/atomic操作による排他制御
- **メモリリーク**: 適切なリソース解放と監視
- **デッドロック**: ロック順序の統一と timeout適用

### パフォーマンスリスク  
- **高レイテンシ**: プロファイリングによるボトルネック特定
- **CPU使用率**: 効率的なデータ構造とアルゴリズム選択
- **メモリ使用量**: オブジェクトプーリングと適切なGC制御

## 次フェーズとの連携

### TASK-204 (MetricsCollector) との連携
- イベントバス統計情報の提供
- メトリクス収集API対応
- パフォーマンス監視データ連携

### TASK-301 (ModECSAPI) との連携  
- MOD向けイベントAPI提供
- セキュリティ制限機能
- イベント権限管理

---

**作成日**: 2025-08-07  
**更新日**: 2025-08-07  
**レビュー状態**: 初版作成完了  
**承認者**: TBD