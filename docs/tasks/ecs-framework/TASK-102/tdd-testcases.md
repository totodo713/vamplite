# TASK-102: ComponentStore実装 - テストケース仕様

## テスト戦略概要

この文書では、ComponentStore実装の全テストケースを定義します。TDDアプローチに従い、失敗するテストを先に実装し、段階的に機能を完成させます。

## テストケース分類

### 1. 単体テスト (Unit Tests)
- 個別機能の動作確認
- エラーハンドリングの検証
- エッジケースの処理確認

### 2. 統合テスト (Integration Tests)
- 他のECSコンポーネントとの連携
- 複合的な操作の検証

### 3. パフォーマンステスト (Performance Tests)
- レスポンス時間の測定
- メモリ使用量の確認
- スケーラビリティの検証

### 4. 並行性テスト (Concurrency Tests)
- マルチスレッド環境での安全性
- データレースの防止確認

## 単体テストケース詳細

### TC-102-001: コンポーネント型登録

#### TC-102-001-001: 正常な型登録
```go
func TestComponentStore_RegisterComponentType_Success(t *testing.T) {
    // Given: 空のComponentStore
    store := NewComponentStore()
    componentType := ecs.ComponentType("Transform")
    poolSize := 100
    
    // When: コンポーネント型を登録
    err := store.RegisterComponentType(componentType, poolSize)
    
    // Then: 正常に登録される
    assert.NoError(t, err)
    assert.True(t, store.IsRegistered(componentType))
    
    // And: 登録済み型一覧に含まれる
    registeredTypes := store.GetRegisteredTypes()
    assert.Contains(t, registeredTypes, componentType)
}
```

#### TC-102-001-002: 重複型登録エラー
```go
func TestComponentStore_RegisterComponentType_DuplicateError(t *testing.T) {
    // Given: すでに登録済みのコンポーネント型
    store := NewComponentStore()
    componentType := ecs.ComponentType("Transform")
    store.RegisterComponentType(componentType, 100)
    
    // When: 同じ型を再登録
    err := store.RegisterComponentType(componentType, 200)
    
    // Then: エラーが返される
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "already registered")
}
```

#### TC-102-001-003: 無効なプールサイズ
```go
func TestComponentStore_RegisterComponentType_InvalidPoolSize(t *testing.T) {
    // Given: ComponentStoreインスタンス
    store := NewComponentStore()
    componentType := ecs.ComponentType("Transform")
    
    // When: 無効なプールサイズで登録
    err := store.RegisterComponentType(componentType, -1)
    
    // Then: エラーが返される
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid pool size")
}
```

### TC-102-002: コンポーネント追加操作

#### TC-102-002-001: 正常なコンポーネント追加
```go
func TestComponentStore_AddComponent_Success(t *testing.T) {
    // Given: 登録済みコンポーネント型のストア
    store := setupStoreWithTransformComponent(t)
    entity := ecs.EntityID(1)
    component := &components.TransformComponent{
        Position: components.Vector3{X: 1.0, Y: 2.0, Z: 3.0},
    }
    
    // When: コンポーネントを追加
    err := store.AddComponent(entity, component)
    
    // Then: 正常に追加される
    assert.NoError(t, err)
    assert.True(t, store.HasComponent(entity, component.GetType()))
    
    // And: 追加されたコンポーネントを取得できる
    retrieved, err := store.GetComponent(entity, component.GetType())
    assert.NoError(t, err)
    assert.Equal(t, component, retrieved)
}
```

#### TC-102-002-002: 未登録型でのコンポーネント追加エラー
```go
func TestComponentStore_AddComponent_UnregisteredTypeError(t *testing.T) {
    // Given: 空のComponentStore
    store := NewComponentStore()
    entity := ecs.EntityID(1)
    component := &components.TransformComponent{}
    
    // When: 未登録型のコンポーネントを追加
    err := store.AddComponent(entity, component)
    
    // Then: エラーが返される
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not registered")
}
```

#### TC-102-002-003: 重複コンポーネント追加エラー
```go
func TestComponentStore_AddComponent_DuplicateError(t *testing.T) {
    // Given: すでにコンポーネントが追加されたエンティティ
    store := setupStoreWithTransformComponent(t)
    entity := ecs.EntityID(1)
    component1 := &components.TransformComponent{Position: components.Vector3{X: 1, Y: 2, Z: 3}}
    component2 := &components.TransformComponent{Position: components.Vector3{X: 4, Y: 5, Z: 6}}
    store.AddComponent(entity, component1)
    
    // When: 同じ型の別のコンポーネントを追加
    err := store.AddComponent(entity, component2)
    
    // Then: エラーが返される
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "already has component")
}
```

### TC-102-003: コンポーネント取得操作

#### TC-102-003-001: 正常なコンポーネント取得
```go
func TestComponentStore_GetComponent_Success(t *testing.T) {
    // Given: コンポーネントが追加されたエンティティ
    store, entity, originalComponent := setupEntityWithTransform(t)
    
    // When: コンポーネントを取得
    retrieved, err := store.GetComponent(entity, originalComponent.GetType())
    
    // Then: 正常に取得される
    assert.NoError(t, err)
    assert.Equal(t, originalComponent, retrieved)
}
```

#### TC-102-003-002: 存在しないコンポーネント取得エラー
```go
func TestComponentStore_GetComponent_NotFoundError(t *testing.T) {
    // Given: 登録済みだが空のストア
    store := setupStoreWithTransformComponent(t)
    entity := ecs.EntityID(999)
    componentType := ecs.ComponentType("Transform")
    
    // When: 存在しないエンティティのコンポーネント取得
    retrieved, err := store.GetComponent(entity, componentType)
    
    // Then: エラーが返される
    assert.Error(t, err)
    assert.Nil(t, retrieved)
    assert.Contains(t, err.Error(), "not found")
}
```

#### TC-102-003-003: 型安全性確認
```go
func TestComponentStore_GetComponent_TypeSafety(t *testing.T) {
    // Given: Transformコンポーネントが追加されたエンティティ
    store, entity, _ := setupEntityWithTransform(t)
    
    // When: 異なる型でコンポーネント取得を試行
    retrieved, err := store.GetComponent(entity, ecs.ComponentType("Sprite"))
    
    // Then: エラーが返される
    assert.Error(t, err)
    assert.Nil(t, retrieved)
}
```

### TC-102-004: コンポーネント削除操作

#### TC-102-004-001: 正常なコンポーネント削除
```go
func TestComponentStore_RemoveComponent_Success(t *testing.T) {
    // Given: コンポーネントが追加されたエンティティ
    store, entity, component := setupEntityWithTransform(t)
    
    // When: コンポーネントを削除
    err := store.RemoveComponent(entity, component.GetType())
    
    // Then: 正常に削除される
    assert.NoError(t, err)
    assert.False(t, store.HasComponent(entity, component.GetType()))
    
    // And: 削除後は取得できない
    _, err = store.GetComponent(entity, component.GetType())
    assert.Error(t, err)
}
```

#### TC-102-004-002: 存在しないコンポーネント削除エラー
```go
func TestComponentStore_RemoveComponent_NotFoundError(t *testing.T) {
    // Given: 空のストア
    store := setupStoreWithTransformComponent(t)
    entity := ecs.EntityID(999)
    componentType := ecs.ComponentType("Transform")
    
    // When: 存在しないコンポーネントを削除
    err := store.RemoveComponent(entity, componentType)
    
    // Then: エラーが返される
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
}
```

### TC-102-005: バルク操作

#### TC-102-005-001: 複数エンティティへのコンポーネント一括追加
```go
func TestComponentStore_AddComponentsBatch_Success(t *testing.T) {
    // Given: 登録済みストアと複数エンティティ
    store := setupStoreWithTransformComponent(t)
    entities := []ecs.EntityID{1, 2, 3, 4, 5}
    components := make([]ecs.Component, len(entities))
    
    for i, entity := range entities {
        components[i] = &components.TransformComponent{
            Position: components.Vector3{X: float64(i), Y: float64(i*2), Z: float64(i*3)},
        }
    }
    
    // When: バッチでコンポーネント追加
    err := store.AddComponentsBatch(entities, components)
    
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

#### TC-102-005-002: バッチ処理でのエラー処理
```go
func TestComponentStore_AddComponentsBatch_PartialError(t *testing.T) {
    // Given: 一部のエンティティに既存コンポーネントがある状態
    store := setupStoreWithTransformComponent(t)
    entities := []ecs.EntityID{1, 2, 3}
    
    // エンティティ2に先にコンポーネント追加
    store.AddComponent(entities[1], &components.TransformComponent{})
    
    components := []ecs.Component{
        &components.TransformComponent{Position: components.Vector3{X: 1, Y: 1, Z: 1}},
        &components.TransformComponent{Position: components.Vector3{X: 2, Y: 2, Z: 2}}, // 重複エラー
        &components.TransformComponent{Position: components.Vector3{X: 3, Y: 3, Z: 3}},
    }
    
    // When: バッチ処理実行
    err := store.AddComponentsBatch(entities, components)
    
    // Then: エラーが返される
    assert.Error(t, err)
    
    // And: 部分的な変更はロールバックされる
    assert.False(t, store.HasComponent(entities[0], components[0].GetType()))
    assert.False(t, store.HasComponent(entities[2], components[2].GetType()))
}
```

### TC-102-006: クエリ操作

#### TC-102-006-001: 特定コンポーネント型を持つエンティティ取得
```go
func TestComponentStore_GetEntitiesWithComponent_Success(t *testing.T) {
    // Given: 複数エンティティにコンポーネントが追加された状態
    store := setupStoreWithTransformComponent(t)
    expectedEntities := []ecs.EntityID{1, 3, 5, 7, 9}
    componentType := ecs.ComponentType("Transform")
    
    for _, entity := range expectedEntities {
        store.AddComponent(entity, &components.TransformComponent{})
    }
    
    // 別のエンティティに別の型のコンポーネント追加
    store.RegisterComponentType(ecs.ComponentType("Sprite"), 100)
    store.AddComponent(ecs.EntityID(2), &components.SpriteComponent{})
    
    // When: Transform型を持つエンティティを取得
    entities := store.GetEntitiesWithComponent(componentType)
    
    // Then: 正しいエンティティが返される
    assert.ElementsMatch(t, expectedEntities, entities)
}
```

#### TC-102-006-002: 複数コンポーネント型の組み合わせクエリ
```go
func TestComponentStore_GetEntitiesWithMultipleComponents_Success(t *testing.T) {
    // Given: 複数のコンポーネント型を持つエンティティ
    store := setupMultiComponentStore(t)
    transformType := ecs.ComponentType("Transform")
    spriteType := ecs.ComponentType("Sprite")
    
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
    
    // When: Transform + Sprite の両方を持つエンティティを取得
    entities := store.GetEntitiesWithComponents([]ecs.ComponentType{transformType, spriteType})
    
    // Then: 該当するエンティティのみが返される
    expected := []ecs.EntityID{entity1, entity3}
    assert.ElementsMatch(t, expected, entities)
}
```

### TC-102-007: エンティティ管理

#### TC-102-007-001: エンティティの全コンポーネント取得
```go
func TestComponentStore_GetAllComponents_Success(t *testing.T) {
    // Given: 複数コンポーネントを持つエンティティ
    store := setupMultiComponentStore(t)
    entity := ecs.EntityID(1)
    
    transformComp := &components.TransformComponent{Position: components.Vector3{X: 1, Y: 2, Z: 3}}
    spriteComp := &components.SpriteComponent{TextureID: "texture1"}
    
    store.AddComponent(entity, transformComp)
    store.AddComponent(entity, spriteComp)
    
    // When: エンティティの全コンポーネントを取得
    allComponents := store.GetAllComponents(entity)
    
    // Then: 全てのコンポーネントが返される
    assert.Len(t, allComponents, 2)
    assert.Contains(t, allComponents, transformComp)
    assert.Contains(t, allComponents, spriteComp)
}
```

#### TC-102-007-002: エンティティ削除とクリーンアップ
```go
func TestComponentStore_RemoveEntity_Success(t *testing.T) {
    // Given: 複数コンポーネントを持つエンティティ
    store := setupMultiComponentStore(t)
    entity := ecs.EntityID(1)
    
    store.AddComponent(entity, &components.TransformComponent{})
    store.AddComponent(entity, &components.SpriteComponent{})
    
    // When: エンティティを削除
    removedCount := store.RemoveEntity(entity)
    
    // Then: 全コンポーネントが削除される
    assert.Equal(t, 2, removedCount)
    assert.Equal(t, 0, len(store.GetAllComponents(entity)))
    
    // And: エンティティ追跡からも削除される
    assert.Equal(t, 0, store.GetEntityCount())
}
```

## パフォーマンステストケース

### TC-102-P001: 大量コンポーネント処理性能

#### TC-102-P001-001: 10,000コンポーネント追加性能
```go
func BenchmarkComponentStore_AddComponent_10K(b *testing.B) {
    store := setupStoreWithTransformComponent(b)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        for j := 0; j < 10000; j++ {
            entity := ecs.EntityID(i*10000 + j)
            component := &components.TransformComponent{
                Position: components.Vector3{X: float64(j), Y: float64(j), Z: float64(j)},
            }
            store.AddComponent(entity, component)
        }
    }
}
```

#### TC-102-P001-002: 検索性能測定
```go
func BenchmarkComponentStore_GetComponent_Performance(b *testing.B) {
    // Given: 10,000個のコンポーネントが追加されたストア
    store := setupLargeStore(b, 10000)
    componentType := ecs.ComponentType("Transform")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        entityID := ecs.EntityID(i % 10000)
        _, err := store.GetComponent(entityID, componentType)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### TC-102-P002: メモリ使用量測定

#### TC-102-P002-001: メモリ効率測定
```go
func TestComponentStore_MemoryEfficiency(t *testing.T) {
    // Given: メモリ使用量測定開始
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // When: 大量のコンポーネントを追加
    store := setupStoreWithTransformComponent(t)
    for i := 0; i < 10000; i++ {
        entity := ecs.EntityID(i)
        component := &components.TransformComponent{
            Position: components.Vector3{X: float64(i), Y: float64(i), Z: float64(i)},
        }
        store.AddComponent(entity, component)
    }
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    // Then: メモリ使用量が基準以下
    memoryUsed := m2.Alloc - m1.Alloc
    memoryPerEntity := memoryUsed / 10000
    
    // エンティティあたり100B以下の要件確認
    assert.LessOrEqual(t, memoryPerEntity, uint64(100), 
        "Memory per entity: %d bytes (should be ≤ 100)", memoryPerEntity)
}
```

## 並行性テストケース

### TC-102-C001: 読み書き競合テスト

#### TC-102-C001-001: 並行読み取り安全性
```go
func TestComponentStore_ConcurrentRead_Safety(t *testing.T) {
    // Given: コンポーネントが追加されたストア
    store := setupStoreWithTransformComponent(t)
    entity := ecs.EntityID(1)
    component := &components.TransformComponent{Position: components.Vector3{X: 1, Y: 2, Z: 3}}
    store.AddComponent(entity, component)
    
    // When: 複数ゴルーチンで同時読み取り
    numGoroutines := 100
    numReadsPerGoroutine := 1000
    var wg sync.WaitGroup
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < numReadsPerGoroutine; j++ {
                retrieved, err := store.GetComponent(entity, component.GetType())
                assert.NoError(t, err)
                assert.Equal(t, component, retrieved)
            }
        }()
    }
    
    wg.Wait()
    
    // Then: データ破損なし
    retrieved, err := store.GetComponent(entity, component.GetType())
    assert.NoError(t, err)
    assert.Equal(t, component, retrieved)
}
```

#### TC-102-C001-002: 読み書き混在テスト
```go
func TestComponentStore_ConcurrentReadWrite_Safety(t *testing.T) {
    store := setupStoreWithTransformComponent(t)
    numWorkers := 10
    numOperations := 1000
    
    var wg sync.WaitGroup
    
    // Writer goroutines
    for i := 0; i < numWorkers/2; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for j := 0; j < numOperations; j++ {
                entity := ecs.EntityID(workerID*numOperations + j)
                component := &components.TransformComponent{
                    Position: components.Vector3{X: float64(j), Y: float64(j), Z: float64(j)},
                }
                err := store.AddComponent(entity, component)
                assert.NoError(t, err)
            }
        }(i)
    }
    
    // Reader goroutines
    for i := numWorkers / 2; i < numWorkers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for j := 0; j < numOperations; j++ {
                entity := ecs.EntityID(j)
                if store.HasComponent(entity, ecs.ComponentType("Transform")) {
                    _, err := store.GetComponent(entity, ecs.ComponentType("Transform"))
                    if err != nil && !strings.Contains(err.Error(), "not found") {
                        t.Errorf("Unexpected error: %v", err)
                    }
                }
            }
        }(i)
    }
    
    wg.Wait()
}
```

## 統合テストケース

### TC-102-I001: EntityManager統合

#### TC-102-I001-001: エンティティライフサイクル統合
```go
func TestComponentStore_EntityManagerIntegration(t *testing.T) {
    // Given: EntityManagerとComponentStoreの統合環境
    entityManager := ecs.NewEntityManager()
    componentStore := NewComponentStore()
    componentStore.RegisterComponentType(ecs.ComponentType("Transform"), 100)
    
    // When: エンティティ作成→コンポーネント追加→エンティティ削除の流れ
    entity := entityManager.CreateEntity()
    component := &components.TransformComponent{Position: components.Vector3{X: 1, Y: 2, Z: 3}}
    
    err := componentStore.AddComponent(entity, component)
    assert.NoError(t, err)
    
    // エンティティ削除時のクリーンアップ
    componentStore.RemoveEntity(entity)
    entityManager.RemoveEntity(entity)
    
    // Then: 完全にクリーンアップされる
    assert.False(t, componentStore.HasComponent(entity, component.GetType()))
    assert.False(t, entityManager.IsValid(entity))
}
```

## テストヘルパー関数

### セットアップ関数群

```go
// 基本的なTransformコンポーネント登録済みストア作成
func setupStoreWithTransformComponent(t testing.TB) *ComponentStore {
    store := NewComponentStore()
    err := store.RegisterComponentType(ecs.ComponentType("Transform"), 100)
    require.NoError(t, err)
    return store
}

// 複数コンポーネント型登録済みストア作成
func setupMultiComponentStore(t testing.TB) *ComponentStore {
    store := NewComponentStore()
    store.RegisterComponentType(ecs.ComponentType("Transform"), 100)
    store.RegisterComponentType(ecs.ComponentType("Sprite"), 100)
    store.RegisterComponentType(ecs.ComponentType("Physics"), 100)
    return store
}

// エンティティ+コンポーネント設定済み状態作成
func setupEntityWithTransform(t testing.TB) (*ComponentStore, ecs.EntityID, ecs.Component) {
    store := setupStoreWithTransformComponent(t)
    entity := ecs.EntityID(1)
    component := &components.TransformComponent{
        Position: components.Vector3{X: 1.0, Y: 2.0, Z: 3.0},
    }
    
    err := store.AddComponent(entity, component)
    require.NoError(t, err)
    
    return store, entity, component
}

// 大量データ設定済みストア作成
func setupLargeStore(t testing.TB, entityCount int) *ComponentStore {
    store := setupStoreWithTransformComponent(t)
    
    for i := 0; i < entityCount; i++ {
        entity := ecs.EntityID(i)
        component := &components.TransformComponent{
            Position: components.Vector3{X: float64(i), Y: float64(i), Z: float64(i)},
        }
        err := store.AddComponent(entity, component)
        require.NoError(t, err)
    }
    
    return store
}
```

## テスト実行順序

### Phase 1: 基本機能テスト
1. TC-102-001: コンポーネント型登録
2. TC-102-002: コンポーネント追加
3. TC-102-003: コンポーネント取得
4. TC-102-004: コンポーネント削除

### Phase 2: 高度機能テスト
1. TC-102-005: バルク操作
2. TC-102-006: クエリ操作
3. TC-102-007: エンティティ管理

### Phase 3: 品質保証テスト
1. TC-102-P001, P002: パフォーマンステスト
2. TC-102-C001: 並行性テスト
3. TC-102-I001: 統合テスト

## 完了条件

- [ ] 全単体テストケース実装・通過
- [ ] 全統合テストケース実装・通過
- [ ] パフォーマンステスト実装・基準達成
- [ ] 並行性テスト実装・安全性確認
- [ ] コードカバレッジ95%以上達成
- [ ] 静的解析ツールでの検証通過

---

## 次のステップ

このテストケース仕様を基に、次は **tdd-red.md** で失敗するテストの実装を行います。