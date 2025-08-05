# TASK-102: ComponentStore実装 - TDD Red Phase

## 概要

このフェーズでは、要件定義とテストケース仕様に基づいて、現在の実装では**失敗する**テストを実装します。TDDの原則に従い、まず失敗するテストを書くことで、実装すべき機能を明確にします。

## 実装する失敗テスト

### 1. バルク操作テスト (新機能)

現在の実装にはバルク操作機能がないため、これらのテストは失敗します。

#### TC-102-005-001: バッチコンポーネント追加
```go
func TestComponentStore_AddComponentsBatch_Success(t *testing.T) {
    // Given: 登録済みストアと複数エンティティ
    store := setupStoreWithTransformComponent(t)
    entities := []ecs.EntityID{1, 2, 3, 4, 5}
    components := make([]ecs.Component, len(entities))
    
    for i := range entities {
        components[i] = &components.TransformComponent{
            Position: ecs.Vector3{X: float64(i), Y: float64(i*2), Z: float64(i*3)},
        }
    }
    
    // When: バッチでコンポーネント追加
    err := store.AddComponentsBatch(entities, components) // この機能は未実装
    
    // Then: 全て正常に追加される
    assert.NoError(t, err)
    
    for i, entity := range entities {
        assert.True(t, store.HasComponent(entity, components[i].GetType()))
        retrieved, err := store.GetComponent(entity, components[i].GetType())
        assert.NoError(t, err)
        assert.Equal(t, components[i], retrieved)
    }
}
```

#### TC-102-005-002: バッチ削除
```go
func TestComponentStore_RemoveComponentsBatch_Success(t *testing.T) {
    // Given: 複数エンティティにコンポーネントが追加された状態
    store := setupStoreWithTransformComponent(t)
    entities := []ecs.EntityID{1, 2, 3, 4, 5}
    
    for _, entity := range entities {
        component := &components.TransformComponent{Position: ecs.Vector3{X: 1, Y: 2, Z: 3}}
        store.AddComponent(entity, component)
    }
    
    // When: バッチで削除
    err := store.RemoveComponentsBatch(entities, ecs.ComponentTypeTransform) // 未実装
    
    // Then: 全て削除される
    assert.NoError(t, err)
    
    for _, entity := range entities {
        assert.False(t, store.HasComponent(entity, ecs.ComponentTypeTransform))
    }
}
```

### 2. 複雑クエリテスト (新機能)

現在の実装では単一コンポーネント型のクエリのみ対応しているため、複数型の組み合わせクエリは失敗します。

#### TC-102-006-002: 複数コンポーネント型クエリ
```go
func TestComponentStore_GetEntitiesWithMultipleComponents_Success(t *testing.T) {
    // Given: 複数のコンポーネント型を持つエンティティ
    store := setupMultiComponentStore(t)
    
    // エンティティ1: Transform + Sprite
    entity1 := ecs.EntityID(1)
    store.AddComponent(entity1, &components.TransformComponent{})
    store.AddComponent(entity1, &components.SpriteComponent{})
    
    // エンティティ2: Transformのみ
    entity2 := ecs.EntityID(2)
    store.AddComponent(entity2, &components.TransformComponent{})
    
    // エンティティ3: Transform + Sprite
    entity3 := ecs.EntityID(3)
    store.AddComponent(entity3, &components.TransformComponent{})
    store.AddComponent(entity3, &components.SpriteComponent{})
    
    // When: Transform + Sprite 両方を持つエンティティを取得
    entities := store.GetEntitiesWithMultipleComponents([]ecs.ComponentType{
        ecs.ComponentTypeTransform, ecs.ComponentTypeSprite,
    }) // この機能は未実装
    
    // Then: 該当するエンティティのみが返される
    expected := []ecs.EntityID{entity1, entity3}
    assert.ElementsMatch(t, expected, entities)
}
```

### 3. エラーハンドリング強化テスト

現在の実装では基本的なエラーハンドリングのみなので、より詳細なエラー情報が必要なテストは失敗します。

#### TC-102-ERROR-001: 詳細エラー情報
```go
func TestComponentStore_DetailedErrorHandling(t *testing.T) {
    // Given: 空のストア
    store := NewComponentStore()
    entity := ecs.EntityID(1)
    componentType := ecs.ComponentTypeTransform
    
    // When: 未登録型でコンポーネント取得
    _, err := store.GetComponent(entity, componentType)
    
    // Then: 詳細なエラー情報が含まれる
    assert.Error(t, err)
    
    var storeErr *ComponentStoreError
    assert.True(t, errors.As(err, &storeErr)) // ComponentStoreError型への変換
    assert.Equal(t, ErrorTypeComponentTypeNotRegistered, storeErr.Type)
    assert.Equal(t, entity, storeErr.EntityID)
    assert.Equal(t, componentType, storeErr.ComponentType)
}
```

### 4. パフォーマンス要件テスト

現在の実装ではパフォーマンス監視機能がないため、これらのテストは失敗します。

#### TC-102-P001-001: レスポンス時間要件
```go
func TestComponentStore_PerformanceRequirements(t *testing.T) {
    // Given: 10,000個のコンポーネントが追加されたストア
    store := setupLargeStore(t, 10000)
    componentType := ecs.ComponentTypeTransform
    
    // When: パフォーマンス測定
    start := time.Now()
    for i := 0; i < 1000; i++ {
        entityID := ecs.EntityID(i % 10000)
        _, err := store.GetComponent(entityID, componentType)
        assert.NoError(t, err)
    }
    elapsed := time.Since(start)
    
    // Then: 平均レスポンス時間が1ms未満
    avgResponseTime := elapsed / 1000
    assert.Less(t, avgResponseTime, time.Millisecond, 
        "Average response time: %v (should be < 1ms)", avgResponseTime)
}
```

### 5. メモリ効率テスト

現在の実装では詳細なメモリ使用量監視機能がないため、これらのテストは失敗する可能性があります。

#### TC-102-P002-001: メモリ使用量要件
```go
func TestComponentStore_MemoryEfficiencyRequirement(t *testing.T) {
    // Given: メモリ使用量測定開始
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // When: 10,000個のコンポーネントを追加
    store := setupStoreWithTransformComponent(t)
    for i := 0; i < 10000; i++ {
        entity := ecs.EntityID(i)
        component := &components.TransformComponent{
            Position: ecs.Vector3{X: float64(i), Y: float64(i), Z: float64(i)},
        }
        store.AddComponent(entity, component)
    }
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    // Then: エンティティあたり100B以下の要件確認
    memoryUsed := m2.Alloc - m1.Alloc
    memoryPerEntity := memoryUsed / 10000
    
    assert.LessOrEqual(t, memoryPerEntity, uint64(100), 
        "Memory per entity: %d bytes (should be ≤ 100)", memoryPerEntity)
}
```

### 6. 並行性安全テスト

現在の実装の並行性制御をより厳密にテストします。

#### TC-102-C001-002: データレース検出
```go
func TestComponentStore_ConcurrentAccess_DataRaceDetection(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping concurrent test in short mode")
    }
    
    // Given: ストアと共有エンティティ
    store := setupStoreWithTransformComponent(t)
    entity := ecs.EntityID(1)
    numWorkers := 100
    numOperations := 1000
    
    var wg sync.WaitGroup
    errors := make(chan error, numWorkers)
    
    // When: 同時に読み書き操作
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            
            for j := 0; j < numOperations; j++ {
                if workerID%2 == 0 {
                    // Writer
                    component := &components.TransformComponent{
                        Position: ecs.Vector3{X: float64(j), Y: float64(j), Z: float64(j)},
                    }
                    if err := store.AddComponent(ecs.EntityID(workerID*numOperations+j), component); err != nil {
                        select {
                        case errors <- err:
                        default:
                        }
                    }
                } else {
                    // Reader
                    if store.HasComponent(entity, ecs.ComponentTypeTransform) {
                        if _, err := store.GetComponent(entity, ecs.ComponentTypeTransform); err != nil {
                            select {
                            case errors <- err:
                            default:
                            }
                        }
                    }
                }
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    // Then: データレース検出ツール(-race)でエラーなし
    for err := range errors {
        if !strings.Contains(err.Error(), "not found") { // 正常なエラーは除外
            t.Errorf("Unexpected error during concurrent access: %v", err)
        }
    }
}
```

### 7. 統合テスト

現在の実装では他のECSコンポーネントとの統合は限定的なため、これらのテストは失敗する可能性があります。

#### TC-102-I001-001: QueryEngine統合
```go
func TestComponentStore_QueryEngineIntegration(t *testing.T) {
    // Given: ComponentStoreとQueryEngineの統合環境
    componentStore := NewComponentStore()
    queryEngine := ecs.NewQueryEngine(componentStore) // QueryEngineは未実装
    
    componentStore.RegisterComponentType(ecs.ComponentTypeTransform, 100)
    componentStore.RegisterComponentType(ecs.ComponentTypeSprite, 100)
    
    // エンティティ追加
    for i := 0; i < 100; i++ {
        entity := ecs.EntityID(i)
        componentStore.AddComponent(entity, &components.TransformComponent{})
        
        if i%2 == 0 {
            componentStore.AddComponent(entity, &components.SpriteComponent{})
        }
    }
    
    // When: QueryEngineを使用した複雑クエリ
    query := queryEngine.NewQuery().
        With(ecs.ComponentTypeTransform).
        With(ecs.ComponentTypeSprite).
        Build()
    
    entities := query.Execute() // この統合機能は未実装
    
    // Then: 正しい結果が返される
    assert.Equal(t, 50, len(entities)) // 偶数番号のエンティティのみ
}
```

## 失敗テスト実装

上記のテストケースを既存のテストファイルに追加し、現在の実装では失敗することを確認します。

### テスト実装ファイル更新
