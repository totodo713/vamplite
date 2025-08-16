# ECSフレームワーク API仕様

## 概要

ECSフレームワークのAPIは、ゲーム開発者とMOD開発者向けに設計された高性能・型安全なインターフェースです。コアゲーム開発向けのフルアクセスAPIと、MOD開発向けの制限付きAPIの2つの層を提供します。

## コアAPI（ゲーム開発者向け）

### 1. EntityManager API

#### エンティティ基本操作

```go
// エンティティ作成
func (em *EntityManager) CreateEntity() EntityID
// 戻り値: 新しいEntityID
// 例外: メモリ不足時にpanicではなくエラーログ

// エンティティ削除  
func (em *EntityManager) DestroyEntity(entity EntityID) error
// パラメータ: entity - 削除対象のEntityID
// 戻り値: error - 無効なEntityIDの場合
// 副作用: 関連するすべてのコンポーネントを自動削除

// エンティティ有効性チェック
func (em *EntityManager) IsEntityValid(entity EntityID) bool
// パラメータ: entity - チェック対象のEntityID  
// 戻り値: bool - エンティティが存在し、アクティブかどうか

// エンティティ数取得
func (em *EntityManager) GetEntityCount() int
// 戻り値: 現在アクティブなエンティティ数
```

#### コンポーネント操作

```go
// コンポーネント追加
func (em *EntityManager) AddComponent(entity EntityID, component Component) error
// パラメータ: 
//   entity - 対象EntityID
//   component - 追加するComponent実装
// 戻り値: error - 無効なエンティティまたは重複コンポーネント
// 動作: 同じ型のコンポーネントが存在する場合は置換

// コンポーネント削除
func (em *EntityManager) RemoveComponent(entity EntityID, componentType ComponentType) error
// パラメータ:
//   entity - 対象EntityID
//   componentType - 削除するコンポーネント型
// 戻り値: error - 無効なエンティティまたは存在しないコンポーネント

// コンポーネント取得
func (em *EntityManager) GetComponent(entity EntityID, componentType ComponentType) (Component, error)
// パラメータ:
//   entity - 対象EntityID  
//   componentType - 取得するコンポーネント型
// 戻り値: 
//   Component - 該当コンポーネント
//   error - エンティティまたはコンポーネントが存在しない場合

// コンポーネント存在チェック
func (em *EntityManager) HasComponent(entity EntityID, componentType ComponentType) bool
// パラメータ:
//   entity - 対象EntityID
//   componentType - チェック対象コンポーネント型
// 戻り値: bool - コンポーネントが存在するかどうか
```

#### バッチ操作

```go
// バッチエンティティ作成
func (em *EntityManager) CreateEntities(count int) ([]EntityID, error)
// パラメータ: count - 作成するエンティティ数（最大10,000）
// 戻り値: 
//   []EntityID - 作成されたEntityIDのスライス
//   error - メモリ不足またはcount制限超過
// パフォーマンス: O(count)、単体作成より約30%高速

// バッチエンティティ削除
func (em *EntityManager) DestroyEntities(entities []EntityID) error
// パラメータ: entities - 削除対象EntityIDのスライス
// 戻り値: error - 無効なEntityIDを含む場合（部分的削除は実行される）
// パフォーマンス: O(n)、関連コンポーネントも一括削除
```

### 2. QueryEngine API

#### 基本クエリ

```go
// マスクベースクエリ
func (qe *QueryEngine) Query(mask ComponentMask) EntityIterator
// パラメータ: mask - 必要なコンポーネントのビットマスク
// 戻り値: EntityIterator - 条件に合致するエンティティのイテレータ
// パフォーマンス: O(1) キャッシュ済み、O(n) 初回実行

// 型指定クエリ
func (qe *QueryEngine) QueryWith(componentTypes ...ComponentType) EntityIterator  
// パラメータ: componentTypes - 必要なコンポーネント型の可変長引数
// 戻り値: EntityIterator - 条件に合致するエンティティのイテレータ
// 使用例: QueryWith(TransformComponentType, SpriteComponentType)

// 除外クエリ
func (qe *QueryEngine) QueryWithout(componentTypes ...ComponentType) EntityIterator
// パラメータ: componentTypes - 除外するコンポーネント型
// 戻り値: EntityIterator - 指定コンポーネントを持たないエンティティ
// 使用例: QueryWithout(DestroyedComponentType)
```

#### 高度なクエリ

```go
// クエリビルダー
func (qe *QueryEngine) NewQuery() QueryBuilder
// 戻り値: QueryBuilder - 複雑な条件を組み立てるビルダー

// QueryBuilder メソッドチェーン
type QueryBuilder interface {
    With(componentTypes ...ComponentType) QueryBuilder
    Without(componentTypes ...ComponentType) QueryBuilder  
    WithAny(componentTypes ...ComponentType) QueryBuilder
    WithAll(componentTypes ...ComponentType) QueryBuilder
    Limit(count int) QueryBuilder
    Execute() EntityIterator
}

// 使用例:
entities := queryEngine.NewQuery().
    With(TransformComponentType, VelocityComponentType).
    Without(DestroyedComponentType).
    WithAny(PlayerComponentType, EnemyComponentType).
    Limit(100).
    Execute()
```

#### キャッシュ管理

```go
// 名前付きクエリキャッシュ作成
func (qe *QueryEngine) CreateCachedQuery(name string, mask ComponentMask) error
// パラメータ:
//   name - クエリの識別名
//   mask - キャッシュするクエリのマスク
// 戻り値: error - 重複名またはメモリ不足

// キャッシュされたクエリの実行
func (qe *QueryEngine) GetCachedQuery(name string) (EntityIterator, error)
// パラメータ: name - キャッシュされたクエリ名
// 戻り値:
//   EntityIterator - クエリ結果のイテレータ
//   error - 存在しないクエリ名
// パフォーマンス: O(1) 常時
```

### 3. SystemManager API

#### システム登録管理

```go
// システム登録
func (sm *SystemManager) RegisterSystem(system System) error
// パラメータ: system - 登録するSystem実装
// 戻り値: error - 重複登録または依存関係エラー
// 動作: 依存関係を自動解析し、実行順序を決定

// システム登録解除
func (sm *SystemManager) UnregisterSystem(systemType SystemType) error
// パラメータ: systemType - 解除するシステム型
// 戻り値: error - 存在しないシステムまたは依存関係エラー

// システム存在チェック
func (sm *SystemManager) IsSystemRegistered(systemType SystemType) bool
// パラメータ: systemType - チェック対象システム型
// 戻り値: bool - システムが登録済みかどうか

// 登録済みシステム一覧
func (sm *SystemManager) GetRegisteredSystems() []SystemType
// 戻り値: []SystemType - 登録済みシステム型のスライス
```

#### システム実行制御

```go
// システム更新実行
func (sm *SystemManager) UpdateSystems(ctx context.Context, deltaTime time.Duration) error
// パラメータ:
//   ctx - キャンセレーション可能なコンテキスト
//   deltaTime - 前フレームからの経過時間
// 戻り値: error - システム実行エラー（すべてのシステムエラーを集約）
// 動作: 依存順序に従って順次またはパラレル実行

// レンダーシステム実行
func (sm *SystemManager) RenderSystems(ctx context.Context, screen *ebiten.Image) error
// パラメータ:
//   ctx - キャンセレーション可能なコンテキスト  
//   screen - 描画対象スクリーン
// 戻り値: error - レンダリングエラー
// 動作: レンダーオーダーに従って順次実行
```

#### システム依存関係管理

```go
// システム依存関係設定
func (sm *SystemManager) SetSystemDependency(dependent, dependency SystemType) error
// パラメータ:
//   dependent - 依存するシステム
//   dependency - 依存されるシステム  
// 戻り値: error - 循環依存または存在しないシステム

// 実行順序取得
func (sm *SystemManager) GetExecutionOrder() []SystemType
// 戻り値: []SystemType - 依存関係を解決した実行順序
```

### 4. EventManager API

#### イベント発行・購読

```go
// イベント購読
func (em *EventManager) Subscribe(eventType string, handler EventHandler) error
// パラメータ:
//   eventType - 購読するイベント型
//   handler - イベントハンドラー実装
// 戻り値: error - 無効なイベント型またはハンドラー

// イベント購読解除
func (em *EventManager) Unsubscribe(eventType string, handler EventHandler) error
// パラメータ:
//   eventType - 購読解除するイベント型
//   handler - 解除対象ハンドラー
// 戻り値: error - 未登録のハンドラー

// 同期イベント発行
func (em *EventManager) Publish(event Event) error
// パラメータ: event - 発行するイベント
// 戻り値: error - ハンドラー実行エラー
// 動作: すべてのハンドラーを順次実行

// 非同期イベント発行
func (em *EventManager) PublishAsync(event Event) error
// パラメータ: event - 発行するイベント
// 戻り値: error - キューイングエラー
// 動作: イベントをキューに追加、後でバッチ処理
```

## MOD API（制限付きアクセス）

### 1. ModECSAPI（サンドボックス化）

#### 制限付きエンティティ操作

```go
// エンティティ有効性チェック（読み取り専用）
func (api *ModECSAPI) IsEntityValid(entity EntityID) bool
// パラメータ: entity - チェック対象EntityID
// 戻り値: bool - エンティティの有効性
// 制限: MODが作成したエンティティのみアクセス可能

// エンティティ数取得（読み取り専用）
func (api *ModECSAPI) GetEntityCount() int
// 戻り値: MODが認識可能なエンティティ数
// 制限: MOD権限内のエンティティのみカウント

// MOD専用エンティティ作成
func (api *ModECSAPI) CreateModEntity() (EntityID, error)
// 戻り値:
//   EntityID - 新規作成されたMOD所有エンティティ
//   error - 権限不足またはリソース制限
// 制限: MOD毎のエンティティ数制限あり（デフォルト1,000個）
```

#### 制限付きコンポーネント操作

```go
// コンポーネント取得（読み取り専用）
func (api *ModECSAPI) GetComponent(entity EntityID, componentType ComponentType) (Component, error)
// パラメータ:
//   entity - 対象EntityID
//   componentType - 取得するコンポーネント型
// 戻り値:
//   Component - 読み取り専用コンポーネント
//   error - アクセス権限不足
// 制限: MODが許可されたコンポーネント型のみ

// コンポーネント存在チェック
func (api *ModECSAPI) HasComponent(entity EntityID, componentType ComponentType) bool
// パラメータ:
//   entity - 対象EntityID
//   componentType - チェック対象コンポーネント型
// 戻り値: bool - コンポーネント存在とアクセス権限
// 制限: アクセス不可な場合は常にfalse

// MODコンポーネント追加
func (api *ModECSAPI) AddModComponent(entity EntityID, component Component) error
// パラメータ:
//   entity - 対象EntityID（MOD所有エンティティのみ）
//   component - 追加するコンポーネント
// 戻り値: error - 権限不足またはコンポーネント制限
// 制限: MOD定義コンポーネントのみ追加可能
```

#### 制限付きクエリ

```go
// 読み取り専用クエリ
func (api *ModECSAPI) QueryReadOnly(mask ComponentMask) EntityIterator
// パラメータ: mask - クエリマスク
// 戻り値: EntityIterator - 読み取り専用イテレータ
// 制限: MODがアクセス可能なエンティティ・コンポーネントのみ

// 型指定読み取り専用クエリ
func (api *ModECSAPI) QueryWithReadOnly(componentTypes ...ComponentType) EntityIterator
// パラメータ: componentTypes - クエリ対象コンポーネント型
// 戻り値: EntityIterator - 読み取り専用イテレータ
// 制限: 許可されたコンポーネント型のみ
```

### 2. MODイベントAPI

```go
// MODイベント購読
func (api *ModECSAPI) SubscribeToEvents(eventType string, handler EventHandler) error
// パラメータ:
//   eventType - 購読するイベント型
//   handler - イベントハンドラー
// 戻り値: error - 権限不足または制限超過
// 制限: MOD許可イベント型のみ（ゲーム状態変更イベントは除外）

// MODイベント発行
func (api *ModECSAPI) PublishModEvent(event Event) error
// パラメータ: event - 発行するMODイベント
// 戻り値: error - 権限不足またはイベント制限
// 制限: MOD名前空間内のイベントのみ発行可能
```

## APIエラーハンドリング

### エラー型定義

```go
type ECSError struct {
    Type        ErrorType
    Message     string
    EntityID    EntityID
    SystemType  SystemType
    Timestamp   time.Time
    StackTrace  []string
}

type ErrorType int

const (
    ErrorTypeInvalidEntity ErrorType = iota
    ErrorTypeInvalidComponent
    ErrorTypeInvalidSystem
    ErrorTypePermissionDenied
    ErrorTypeResourceExhausted
    ErrorTypeInternalError
)
```

### エラーレスポンス例

```go
// エンティティ関連エラー
var (
    ErrEntityNotFound    = &ECSError{Type: ErrorTypeInvalidEntity, Message: "entity not found"}
    ErrEntityDestroyed   = &ECSError{Type: ErrorTypeInvalidEntity, Message: "entity already destroyed"}
    ErrInvalidEntityID   = &ECSError{Type: ErrorTypeInvalidEntity, Message: "invalid entity ID"}
)

// コンポーネント関連エラー
var (
    ErrComponentNotFound     = &ECSError{Type: ErrorTypeInvalidComponent, Message: "component not found"}
    ErrComponentAlreadyExists = &ECSError{Type: ErrorTypeInvalidComponent, Message: "component already exists"}
    ErrInvalidComponentType  = &ECSError{Type: ErrorTypeInvalidComponent, Message: "invalid component type"}
)

// MOD権限エラー
var (
    ErrPermissionDenied      = &ECSError{Type: ErrorTypePermissionDenied, Message: "permission denied"}
    ErrResourceLimitExceeded = &ECSError{Type: ErrorTypeResourceExhausted, Message: "resource limit exceeded"}
)
```

## パフォーマンス仕様

### API性能目標

| API操作 | 目標時間 | 測定条件 | 最適化手法 |
|---------|----------|----------|------------|
| CreateEntity() | <100ns | 単体実行 | オブジェクトプール |
| DestroyEntity() | <200ns | 単体実行 | 遅延削除 |
| AddComponent() | <150ns | 基本コンポーネント | メモリプール |
| GetComponent() | <50ns | キャッシュヒット | ハッシュテーブル |
| Query() | <1ms | 10,000エンティティ | ビットマスク |
| UpdateSystems() | <10ms | 50システム | 並列実行 |

### 同時アクセス性能

```go
// 並行性能仕様
type ConcurrencySpec struct {
    // 読み取り操作の並列性
    MaxReadConcurrency  int // 無制限
    ReadLockType        string // RWMutex.RLock
    
    // 書き込み操作の排他性
    WriteSerialiation   bool // true
    WriteLockType       string // RWMutex.Lock
    
    // MODアクセス制限
    MaxModConcurrency   int // 10 concurrent mods
    ModRateLimiting     bool // true
}
```

## API使用例

### 基本的なゲームループ

```go
func GameLoop(entityManager EntityManager, systemManager SystemManager) error {
    ctx := context.Background()
    
    for {
        start := time.Now()
        
        // システム更新
        if err := systemManager.UpdateSystems(ctx, deltaTime); err != nil {
            log.Printf("System update error: %v", err)
        }
        
        // レンダリング
        if err := systemManager.RenderSystems(ctx, screen); err != nil {
            log.Printf("Render error: %v", err)
        }
        
        // フレーム時間制御
        elapsed := time.Since(start)
        if elapsed < targetFrameTime {
            time.Sleep(targetFrameTime - elapsed)
        }
    }
}
```

### プレイヤーエンティティ作成

```go
func CreatePlayer(em EntityManager) (EntityID, error) {
    // エンティティ作成
    player := em.CreateEntity()
    
    // コンポーネント追加
    transform := &TransformComponent{
        Position: Vector2{X: 100, Y: 100},
        Rotation: 0,
        Scale:    Vector2{X: 1, Y: 1},
    }
    
    sprite := &SpriteComponent{
        Image:   playerImage,
        Visible: true,
        Layer:   1,
    }
    
    health := &HealthComponent{
        Current: 100,
        Maximum: 100,
    }
    
    if err := em.AddComponent(player, transform); err != nil {
        return player, err
    }
    if err := em.AddComponent(player, sprite); err != nil {
        return player, err
    }
    if err := em.AddComponent(player, health); err != nil {
        return player, err
    }
    
    return player, nil
}
```

### システム実装例

```go
type MovementSystem struct {
    entityManager EntityManager
    enabled       bool
}

func (s *MovementSystem) Update(ctx context.Context, deltaTime time.Duration) error {
    if !s.enabled {
        return nil
    }
    
    // 移動可能なエンティティを検索
    entities := s.entityManager.QueryWith(
        TransformComponentType,
        VelocityComponentType,
    )
    defer entities.Close()
    
    deltaSeconds := float64(deltaTime) / float64(time.Second)
    
    for entities.Next() {
        entityID := entities.Entity()
        
        // コンポーネント取得
        transform, err := s.entityManager.GetComponent(entityID, TransformComponentType)
        if err != nil {
            continue
        }
        
        velocity, err := s.entityManager.GetComponent(entityID, VelocityComponentType)
        if err != nil {
            continue
        }
        
        // 位置更新
        t := transform.(*TransformComponent)
        v := velocity.(*VelocityComponent)
        
        t.Position.X += v.Velocity.X * deltaSeconds
        t.Position.Y += v.Velocity.Y * deltaSeconds
        
        // 更新されたコンポーネントを保存
        s.entityManager.AddComponent(entityID, t)
    }
    
    return nil
}
```

このAPI仕様により、ECSフレームワークは高性能で安全な開発環境を提供し、コアゲーム開発とMOD開発の両方をサポートします。