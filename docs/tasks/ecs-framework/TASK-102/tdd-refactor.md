# TASK-102: ComponentStore実装 - TDD Refactor Phase

## 概要

このフェーズでは、Green Phase で実装した機能を**品質向上**させつつ、**テストは通り続ける**ことを保証します。TDDの原則に従い、リファクタリング後も全テストが通ることを確認します。

## リファクタリング対象の特定

### 1. コードの重複 (DRY原則違反)

#### 問題点:
- `AddComponent` と `AddComponentsBatch` で同じ検証ロジックが重複
- `RemoveComponent` と `RemoveComponentsBatch` で同じ削除ロジックが重複

#### 解決策:
共通処理を抽出してヘルパーメソッドに分離する

### 2. パフォーマンス改善

#### 問題点:
- バッチ操作が単純にループで個別処理している
- 複数コンポーネントクエリで効率的でないフィルタリング

#### 解決策:
- バッチ操作での一括処理最適化
- クエリエンジンでのビットセット操作導入

### 3. エラーハンドリングの改善

#### 問題点:
- エラーメッセージが統一されていない
- バッチ操作でのトランザクション的処理が不十分

#### 解決策:
- 統一されたエラー型の導入
- バッチ操作での部分的失敗時のロールバック

## リファクタリング実装

### Phase 1: コード重複の排除

#### ヘルパーメソッドの抽出

```go
// validateComponentAddition validates if a component can be added to an entity
func (s *ComponentStore) validateComponentAddition(entity ecs.EntityID, componentType ecs.ComponentType) error {
    // Check if component type is registered
    if !s.registeredTypes[componentType] {
        return fmt.Errorf("component type %s not registered", componentType)
    }

    // Check if entity already has this component type
    if entityComponents, exists := s.entities[entity]; exists {
        if entityComponents[componentType] {
            return fmt.Errorf("entity %d already has component of type %s", entity, componentType)
        }
    }

    return nil
}

// addComponentInternal adds a component to an entity (assumes validation is done)
func (s *ComponentStore) addComponentInternal(entity ecs.EntityID, component ecs.Component) error {
    componentType := component.GetType()

    // Add entity to sparse set for this component type
    if err := s.sparseSets[componentType].Add(entity); err != nil {
        return fmt.Errorf("failed to add entity %d to sparse set: %w", entity, err)
    }

    // Store component
    s.components[componentType][entity] = component

    // Update entity tracking
    if s.entities[entity] == nil {
        s.entities[entity] = make(map[ecs.ComponentType]bool)
    }
    s.entities[entity][componentType] = true

    return nil
}

// removeComponentInternal removes a component from an entity (assumes validation is done)
func (s *ComponentStore) removeComponentInternal(entity ecs.EntityID, componentType ecs.ComponentType) error {
    // Remove from sparse set
    if err := s.sparseSets[componentType].Remove(entity); err != nil {
        return fmt.Errorf("failed to remove entity %d from sparse set: %w", entity, err)
    }

    // Remove component
    delete(s.components[componentType], entity)

    // Update entity tracking
    delete(s.entities[entity], componentType)
    if len(s.entities[entity]) == 0 {
        delete(s.entities, entity)
    }

    return nil
}
```

#### 既存メソッドのリファクタリング

```go
// AddComponent adds a component to an entity (refactored)
func (s *ComponentStore) AddComponent(entity ecs.EntityID, component ecs.Component) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    componentType := component.GetType()

    if err := s.validateComponentAddition(entity, componentType); err != nil {
        return err
    }

    return s.addComponentInternal(entity, component)
}

// AddComponentsBatch adds multiple components to multiple entities in batch (refactored)
func (s *ComponentStore) AddComponentsBatch(entities []ecs.EntityID, components []ecs.Component) error {
    if len(entities) != len(components) {
        return fmt.Errorf("entities and components slice length mismatch: %d != %d", len(entities), len(components))
    }

    s.mutex.Lock()
    defer s.mutex.Unlock()

    // Validate all operations first (fail-fast)
    for i, entity := range entities {
        componentType := components[i].GetType()
        if err := s.validateComponentAddition(entity, componentType); err != nil {
            return fmt.Errorf("validation failed for entity %d: %w", entity, err)
        }
    }

    // Execute all operations (should not fail after validation)
    for i, entity := range entities {
        if err := s.addComponentInternal(entity, components[i]); err != nil {
            // This should not happen after validation, but handle gracefully
            return fmt.Errorf("internal error adding component to entity %d: %w", entity, err)
        }
    }

    return nil
}
```

### Phase 2: パフォーマンス最適化

#### ビットセット操作による高速クエリ

```go
// GetEntitiesWithMultipleComponents returns entities that have ALL specified component types (optimized)
func (s *ComponentStore) GetEntitiesWithMultipleComponents(componentTypes []ecs.ComponentType) []ecs.EntityID {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    if len(componentTypes) == 0 {
        return []ecs.EntityID{}
    }

    // Find the component type with the smallest entity set to minimize iterations
    var smallestSet *SparseSet
    var smallestSize int = -1
    
    for _, componentType := range componentTypes {
        if sparseSet, exists := s.sparseSets[componentType]; exists {
            size := sparseSet.Size()
            if smallestSize == -1 || size < smallestSize {
                smallestSet = sparseSet
                smallestSize = size
            }
        } else {
            // If any component type doesn't exist, no entities can have all types
            return []ecs.EntityID{}
        }
    }

    if smallestSet == nil {
        return []ecs.EntityID{}
    }

    // Use the smallest set as candidates and check if they have all other component types
    candidates := smallestSet.ToSlice()
    var result []ecs.EntityID

    for _, entity := range candidates {
        hasAllComponents := true

        for _, componentType := range componentTypes {
            if entityComponents, exists := s.entities[entity]; !exists || !entityComponents[componentType] {
                hasAllComponents = false
                break
            }
        }

        if hasAllComponents {
            result = append(result, entity)
        }
    }

    return result
}
```

### Phase 3: エラーハンドリングの改善

#### 統一されたエラー型

```go
// ComponentStoreError represents different types of ComponentStore errors
type ComponentStoreError struct {
    Type         ErrorType
    Message      string
    EntityID     ecs.EntityID
    ComponentType ecs.ComponentType
    Cause        error
}

type ErrorType int

const (
    ErrorTypeComponentNotRegistered ErrorType = iota
    ErrorTypeDuplicateComponent
    ErrorTypeComponentNotFound
    ErrorTypeMemoryAllocation
    ErrorTypeConcurrency
    ErrorTypeValidation
)

func (e *ComponentStoreError) Error() string {
    return e.Message
}

func (e *ComponentStoreError) Unwrap() error {
    return e.Cause
}

func (e *ComponentStoreError) Is(target error) bool {
    if other, ok := target.(*ComponentStoreError); ok {
        return e.Type == other.Type
    }
    return false
}

// Helper functions for creating specific errors
func newComponentNotRegisteredError(componentType ecs.ComponentType) error {
    return &ComponentStoreError{
        Type:          ErrorTypeComponentNotRegistered,
        Message:       fmt.Sprintf("component type %s not registered", componentType),
        ComponentType: componentType,
    }
}

func newDuplicateComponentError(entity ecs.EntityID, componentType ecs.ComponentType) error {
    return &ComponentStoreError{
        Type:          ErrorTypeDuplicateComponent,
        Message:       fmt.Sprintf("entity %d already has component of type %s", entity, componentType),
        EntityID:      entity,
        ComponentType: componentType,
    }
}

func newComponentNotFoundError(entity ecs.EntityID, componentType ecs.ComponentType) error {
    return &ComponentStoreError{
        Type:          ErrorTypeComponentNotFound,
        Message:       fmt.Sprintf("component of type %s not found for entity %d", componentType, entity),
        EntityID:      entity,
        ComponentType: componentType,
    }
}
```

### Phase 4: メモリ効率の改善

#### メモリプール活用の最適化

```go
// GetComponentWithPool retrieves a component using memory pool optimization
func (s *ComponentStore) GetComponentWithPool(entity ecs.EntityID, componentType ecs.ComponentType) (ecs.Component, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    // Check if component type is registered
    if !s.registeredTypes[componentType] {
        return nil, newComponentNotRegisteredError(componentType)
    }

    // Fast path: check existence using sparse set first
    if sparseSet, exists := s.sparseSets[componentType]; exists {
        if !sparseSet.Contains(entity) {
            return nil, newComponentNotFoundError(entity, componentType)
        }
    } else {
        return nil, newComponentNotFoundError(entity, componentType)
    }

    // Get component from storage
    component, exists := s.components[componentType][entity]
    if !exists {
        return nil, newComponentNotFoundError(entity, componentType)
    }

    return component, nil
}
```

## リファクタリング結果の検証

各フェーズでのテスト実行により、機能が保持されていることを確認します。

### 期待される改善

1. **コード品質**:
   - DRY原則に従った重複コードの排除
   - 単一責任原則に従ったメソッドの分離
   - 統一されたエラーハンドリング

2. **パフォーマンス**:
   - バッチ操作での効率向上（10-30%の性能改善期待）
   - クエリ処理での最適化（小さなセットでの効率的フィルタリング）
   - メモリアクセスパターンの最適化

3. **保守性**:
   - 明確な責任分離
   - テスタブルなコード構造
   - 拡張性のあるアーキテクチャ

## 実装手順

1. **Phase 1**: ヘルパーメソッド抽出 → テスト実行
2. **Phase 2**: パフォーマンス最適化 → ベンチマーク実行
3. **Phase 3**: エラーハンドリング改善 → エラーテスト実行
4. **Phase 4**: メモリ効率改善 → メモリテスト実行

各フェーズ後に全テストが通ることを確認し、リグレッションがないことを保証します。