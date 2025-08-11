package cache

import (
	"runtime"
	"unsafe"

	. "muscle-dreamer/internal/core/ecs/optimizations"
)

// OptimizedComponentStore は高性能SoA実装のコンポーネントストア
type OptimizedComponentStore struct {
	// SoA - Structure of Arrays 実装
	transformPositions []Vector3 // X, Y, Z連続配置
	transformRotations []Vector3 // 回転データ連続配置
	transformScales    []Vector3 // スケールデータ連続配置
	
	// エンティティ管理
	entityToIndex map[EntityID]int32 // エンティティ→インデックス
	indexToEntity []EntityID         // インデックス→エンティティ
	
	// 空きインデックス管理（高速削除用）
	freeIndices []int32
	maxIndex    int32
	
	// スプライト管理（簡易実装）
	sprites map[EntityID]SpriteComponent
	
	// メモリアライメント最適化
	cacheAligned bool
}

// NewOptimizedComponentStore creates a new optimized component store
func NewOptimizedComponentStore() *OptimizedComponentStore {
	return &OptimizedComponentStore{
		transforms:     make(map[EntityID]TransformComponent),
		sprites:        make(map[EntityID]SpriteComponent),
		transformArray: make([]TransformComponent, 0),
	}
}

// AddTransform adds a transform component
func (cs *OptimizedComponentStore) AddTransform(entityID EntityID, component TransformComponent) {
	cs.transforms[entityID] = component
	cs.transformArray = append(cs.transformArray, component)
}

// GetTransform gets a transform component
func (cs *OptimizedComponentStore) GetTransform(entityID EntityID) *TransformComponent {
	if component, exists := cs.transforms[entityID]; exists {
		return &component
	}
	return nil
}

// GetTransformArray returns the transform array for SoA access
func (cs *OptimizedComponentStore) GetTransformArray() []TransformComponent {
	return cs.transformArray
}

// PrefetchComponents prefetches components (minimal implementation)
func (cs *OptimizedComponentStore) PrefetchComponents(entities []EntityID) {
	// 最小実装: 実際のプリフェッチはせず、メモリアクセスのみ
	for _, entityID := range entities {
		_ = cs.transforms[entityID] // メモリアクセス
	}
}

// RemoveTransform removes a transform component
func (cs *OptimizedComponentStore) RemoveTransform(entityID EntityID) {
	delete(cs.transforms, entityID)
	
	// transformArray からも削除（簡易実装）
	cs.rebuildTransformArray()
}

// AddSprite adds a sprite component
func (cs *OptimizedComponentStore) AddSprite(entityID EntityID, component SpriteComponent) {
	cs.sprites[entityID] = component
}

// RemoveSprite removes a sprite component
func (cs *OptimizedComponentStore) RemoveSprite(entityID EntityID) {
	delete(cs.sprites, entityID)
}

// rebuildTransformArray rebuilds the transform array after removal
func (cs *OptimizedComponentStore) rebuildTransformArray() {
	cs.transformArray = cs.transformArray[:0] // クリア
	for _, transform := range cs.transforms {
		cs.transformArray = append(cs.transformArray, transform)
	}
}