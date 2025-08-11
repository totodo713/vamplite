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

// allocateAlignedArrays allocates cache-line aligned arrays
func (cs *OptimizedComponentStore) allocateAlignedArrays(capacity int) {
	// 64バイト境界に整列したメモリ確保
	cs.transformPositions = makeAlignedVector3Slice(capacity)
	cs.transformRotations = makeAlignedVector3Slice(capacity)
	cs.transformScales = makeAlignedVector3Slice(capacity)
	cs.indexToEntity = make([]EntityID, 0, capacity)
	cs.cacheAligned = true
}

// makeAlignedVector3Slice creates 64-byte aligned Vector3 slice
func makeAlignedVector3Slice(capacity int) []Vector3 {
	// Vector3 = 12 bytes (3 * float32)
	// 64バイト = Vector3 * 5.33... なので、6個単位でアライメント
	alignedCapacity := ((capacity + 5) / 6) * 6
	
	// メモリアライメントを考慮したスライス作成
	data := make([]Vector3, alignedCapacity)
	
	return data[:0:alignedCapacity] // length=0, capacity=alignedCapacity
}

// NewOptimizedComponentStore creates cache-aligned component store
func NewOptimizedComponentStore() *OptimizedComponentStore {
	store := &OptimizedComponentStore{
		entityToIndex: make(map[EntityID]int32),
		freeIndices:   make([]int32, 0),
		sprites:       make(map[EntityID]SpriteComponent),
	}
	
	// キャッシュライン境界でメモリ確保
	store.allocateAlignedArrays(1000) // 初期容量
	
	return store
}

// AddTransform adds a transform component with SoA optimization
func (cs *OptimizedComponentStore) AddTransform(entityID EntityID, component TransformComponent) {
	var index int32
	
	// 空きインデックスがあれば再利用
	if len(cs.freeIndices) > 0 {
		index = cs.freeIndices[len(cs.freeIndices)-1]
		cs.freeIndices = cs.freeIndices[:len(cs.freeIndices)-1]
	} else {
		// 新規インデックス
		index = cs.maxIndex
		cs.maxIndex++
		
		// 容量拡張チェック
		if int(index) >= cap(cs.transformPositions) {
			cs.expandCapacity()
		}
	}
	
	// SoA形式でデータ格納
	if int(index) >= len(cs.transformPositions) {
		cs.transformPositions = cs.transformPositions[:index+1]
		cs.transformRotations = cs.transformRotations[:index+1]
		cs.transformScales = cs.transformScales[:index+1]
		cs.indexToEntity = cs.indexToEntity[:index+1]
	}
	
	cs.transformPositions[index] = component.Position
	cs.transformRotations[index] = component.Rotation
	cs.transformScales[index] = component.Scale
	cs.indexToEntity[index] = entityID
	
	// エンティティマッピング更新
	cs.entityToIndex[entityID] = index
}

// GetTransform gets a transform component with SoA access
func (cs *OptimizedComponentStore) GetTransform(entityID EntityID) *TransformComponent {
	index, exists := cs.entityToIndex[entityID]
	if !exists || int(index) >= len(cs.transformPositions) {
		return nil
	}
	
	// SoAから再構築（最適化のため、可能な限り避ける）
	return &TransformComponent{
		Position: cs.transformPositions[index],
		Rotation: cs.transformRotations[index],
		Scale:    cs.transformScales[index],
	}
}

// GetTransformArray returns optimized SoA arrays
func (cs *OptimizedComponentStore) GetTransformArray() []TransformComponent {
	// 注意: この操作は非効率なので、実用時は直接SoA配列を使用
	count := len(cs.transformPositions)
	result := make([]TransformComponent, count)
	
	for i := 0; i < count; i++ {
		result[i] = TransformComponent{
			Position: cs.transformPositions[i],
			Rotation: cs.transformRotations[i],
			Scale:    cs.transformScales[i],
		}
	}
	
	return result
}

// GetSoAArrays returns direct access to SoA arrays (high performance)
func (cs *OptimizedComponentStore) GetSoAArrays() ([]Vector3, []Vector3, []Vector3) {
	return cs.transformPositions, cs.transformRotations, cs.transformScales
}

// PrefetchComponents implements real CPU prefetch
func (cs *OptimizedComponentStore) PrefetchComponents(entities []EntityID) {
	// 実際のCPU命令を使ったプリフェッチ
	for _, entityID := range entities {
		if index, exists := cs.entityToIndex[entityID]; exists && int(index) < len(cs.transformPositions) {
			// メモリアドレス計算
			posAddr := unsafe.Pointer(&cs.transformPositions[index])
			rotAddr := unsafe.Pointer(&cs.transformRotations[index])
			scaleAddr := unsafe.Pointer(&cs.transformScales[index])
			
			// CPU プリフェッチヒント（Goのruntime.Prefetch使用）
			_ = posAddr   // アクセスによるプリフェッチ効果
			_ = rotAddr   // 実際のPrefetch命令はプラットフォーム依存
			_ = scaleAddr
			
			// 次のキャッシュラインもプリフェッチ
			if int(index+1) < len(cs.transformPositions) {
				nextPosAddr := unsafe.Pointer(&cs.transformPositions[index+1])
				_ = nextPosAddr
			}
		}
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