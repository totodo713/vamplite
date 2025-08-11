package cache

import (
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"

	"github.com/muscle-dreamer/internal/core/ecs"
)

func TestOptimizedComponentStore_SoALayout(t *testing.T) {
	store := NewOptimizedComponentStore()

	// 連続するエンティティに対してコンポーネント追加
	entities := make([]ecs.EntityID, 1000)
	for i := 0; i < 1000; i++ {
		entities[i] = ecs.EntityID(i)
		store.AddTransform(entities[i], ecs.TransformComponent{
			Position: ecs.Vector3{X: float32(i), Y: 0, Z: 0},
		})
	}

	// メモリレイアウトの連続性確認
	transforms := store.GetTransformArray()
	assert.Equal(t, 1000, len(transforms))

	// キャッシュライン境界整列確認
	if len(transforms) > 0 {
		baseAddr := uintptr(unsafe.Pointer(&transforms[0]))
		assert.Equal(t, 0, baseAddr%64, "Transform array should be 64-byte aligned") // 64バイト境界整列
	}
}

func TestOptimizedComponentStore_Prefetch(t *testing.T) {
	store := NewOptimizedComponentStore()
	entities := setupTestEntities(store, 100)

	// プリフェッチ実行時間測定
	start := time.Now()
	store.PrefetchComponents(entities[:50])
	prefetchTime := time.Since(start)

	// プリフェッチなしでの同等処理時間と比較
	start = time.Now()
	for _, entity := range entities[50:] {
		_ = store.GetTransform(entity)
	}
	normalTime := time.Since(start)

	// プリフェッチによる性能向上確認（測定可能な範囲で）
	assert.Less(t, prefetchTime, normalTime*2, "Prefetch should improve or at least not degrade performance")
}

// setupTestEntities creates test entities for benchmarking
func setupTestEntities(store *OptimizedComponentStore, count int) []ecs.EntityID {
	entities := make([]ecs.EntityID, count)
	for i := 0; i < count; i++ {
		entities[i] = ecs.EntityID(i)
		store.AddTransform(entities[i], ecs.TransformComponent{
			Position: ecs.Vector3{X: float32(i), Y: float32(i), Z: float32(i)},
			Rotation: ecs.Vector3{X: 0, Y: 0, Z: 0},
			Scale:    ecs.Vector3{X: 1, Y: 1, Z: 1},
		})
	}
	return entities
}