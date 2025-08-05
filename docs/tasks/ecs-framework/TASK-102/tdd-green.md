# TASK-102: ComponentStore実装 - TDD Green Phase

## 概要

このフェーズでは、Red Phase で作成したテストを通すための**最小限の実装**を追加します。TDDの原則に従い、過度な実装は行わず、テストが通る最小限のコードのみを追加します。

## 実装する機能

### 1. バルク操作機能の最小実装

現在のSkipされているテストを通すため、基本的なバッチ操作を実装します。

#### AddComponentsBatch メソッド
```go
// AddComponentsBatch adds multiple components to multiple entities in batch
func (s *ComponentStore) AddComponentsBatch(entities []ecs.EntityID, components []ecs.Component) error {
    if len(entities) != len(components) {
        return fmt.Errorf("entities and components slice length mismatch: %d != %d", len(entities), len(components))
    }
    
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    // Simple implementation: add each component individually
    for i, entity := range entities {
        component := components[i]
        componentType := component.GetType()
        
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
    }
    
    return nil
}
```

#### RemoveComponentsBatch メソッド
```go
// RemoveComponentsBatch removes components of specified type from multiple entities
func (s *ComponentStore) RemoveComponentsBatch(entities []ecs.EntityID, componentType ecs.ComponentType) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    // Simple implementation: remove each component individually
    for _, entity := range entities {
        // Check if entity has this component
        if entityComponents, exists := s.entities[entity]; !exists || !entityComponents[componentType] {
            // Skip if entity doesn't have the component (not an error for batch operations)
            continue
        }
        
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
    }
    
    return nil
}
```

### 2. 複数コンポーネントクエリ機能の最小実装

#### GetEntitiesWithMultipleComponents メソッド
```go
// GetEntitiesWithMultipleComponents returns entities that have ALL specified component types
func (s *ComponentStore) GetEntitiesWithMultipleComponents(componentTypes []ecs.ComponentType) []ecs.EntityID {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    if len(componentTypes) == 0 {
        return []ecs.EntityID{}
    }
    
    // Start with entities that have the first component type
    firstType := componentTypes[0]
    if sparseSet, exists := s.sparseSets[firstType]; exists {
        candidates := sparseSet.ToSlice()
        
        // Filter entities that have all remaining component types
        var result []ecs.EntityID
        for _, entity := range candidates {
            hasAllComponents := true
            
            for _, componentType := range componentTypes[1:] {
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
    
    return []ecs.EntityID{}
}
```

### 3. 実装を追加

既存のComponentStoreファイルに上記の機能を追加します。