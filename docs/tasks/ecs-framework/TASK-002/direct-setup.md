# TASK-002: コアインターフェース定義 - 直接実装

## 実装目標

ECSフレームワークの中核となるインターフェースを定義し、型安全で高性能なEntity Component Systemの基盤を構築する。

## 実装ステップ

### Step 1: 基本型定義 (`types.go`)

ECSフレームワークの基本型とエラー型を定義：

- `EntityID` - エンティティの一意識別子
- `ComponentType` - コンポーネント種別
- `SystemType` - システム種別  
- `Priority` - システム実行優先度
- 基本定数・エラーコード定義

### Step 2: Worldインターフェース (`world.go`)

ECS統合管理の中核インターフェース：

```go
type World interface {
    // Entity management
    CreateEntity() EntityID
    DestroyEntity(EntityID) error
    IsEntityValid(EntityID) bool
    
    // Component management
    AddComponent(EntityID, Component) error
    RemoveComponent(EntityID, ComponentType) error
    GetComponent(EntityID, ComponentType) (Component, error)
    
    // System management
    RegisterSystem(System) error
    UnregisterSystem(SystemType) error
    
    // World operations
    Update(deltaTime float64) error
    Render(screen *ebiten.Image) error
    
    // Performance monitoring
    GetMetrics() *PerformanceMetrics
}
```

### Step 3: EntityManagerインターフェース (`entity.go`)

エンティティライフサイクル管理：

```go
type EntityManager interface {
    CreateEntity() EntityID
    DestroyEntity(EntityID) error
    IsValid(EntityID) bool
    GetActiveEntities() []EntityID
    GetEntityCount() int
    RecycleEntity(EntityID) error
    
    // Entity relationships
    SetParent(child EntityID, parent EntityID) error
    GetParent(EntityID) (EntityID, bool)
    GetChildren(EntityID) []EntityID
    
    // Entity metadata
    SetTag(EntityID, string) error
    GetTag(EntityID) (string, bool)
    FindByTag(string) []EntityID
}
```

### Step 4: ComponentStoreインターフェース (`component.go`)

コンポーネント管理・ストレージ：

```go
type ComponentStore interface {
    // Component operations
    AddComponent(EntityID, Component) error
    RemoveComponent(EntityID, ComponentType) error
    GetComponent(EntityID, ComponentType) (Component, error)
    HasComponent(EntityID, ComponentType) bool
    
    // Bulk operations
    GetComponents(EntityID) []Component
    RemoveAllComponents(EntityID) error
    
    // Type management
    RegisterComponentType(ComponentType, func() Component) error
    GetRegisteredTypes() []ComponentType
    
    // Storage optimization
    Compact() error
    GetStorageStats() StorageStats
}
```

### Step 5: SystemManagerインターフェース (`system.go`)

システム登録・実行管理：

```go
type SystemManager interface {
    // System registration
    RegisterSystem(System) error
    UnregisterSystem(SystemType) error
    GetSystem(SystemType) (System, error)
    GetAllSystems() []System
    
    // Execution management
    UpdateSystems(deltaTime float64) error
    RenderSystems(screen *ebiten.Image) error
    
    // Dependency management
    SetSystemDependency(SystemType, SystemType) error
    GetExecutionOrder() []SystemType
    ValidateExecutionOrder() error
    
    // System state
    EnableSystem(SystemType) error
    DisableSystem(SystemType) error
    IsSystemEnabled(SystemType) bool
}
```

### Step 6: QueryEngineインターフェース (`query.go`)

高速エンティティクエリシステム：

```go
type QueryEngine interface {
    // Query creation
    CreateQuery(QueryBuilder) QueryResult
    CacheQuery(string, QueryBuilder) QueryResult
    
    // Query execution
    Execute(QueryBuilder) QueryResult
    ExecuteCached(string) QueryResult
    
    // Query optimization
    OptimizeQueries() error
    ClearQueryCache() error
    GetQueryStats() QueryStats
    
    // Real-time updates
    UpdateQueryCache(EntityID, ComponentType, bool) error
}
```

### Step 7: エラー定義 (`errors.go`)

ECSフレームワーク特化エラー：

```go
type ECSError struct {
    Code      string    `json:"code"`
    Message   string    `json:"message"`
    Component string    `json:"component,omitempty"`
    Entity    EntityID  `json:"entity,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}
```

## 実装品質基準

### 型安全性
- 全インターフェースにおける型安全性確保
- ジェネリクス活用による実行時エラー防止
- コンパイル時型チェック最大化

### パフォーマンス設計
- メモリ効率を考慮したインターフェース設計
- CPUキャッシュ効率を意識したデータ構造
- 大量データ処理に適した設計

### 拡張性
- 将来の機能追加に対応可能な設計
- MODシステムとの統合を考慮
- プラグイン機能への対応

## 実装開始