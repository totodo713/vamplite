package mod

import (
	"sync"
	"time"

	"muscle-dreamer/internal/core/ecs"
)

// PerformanceMonitor パフォーマンス監視
type PerformanceMonitor struct {
	mu               sync.RWMutex
	apiCallDurations map[string][]time.Duration
	memorySnapshots  []int64
	queryFrequency   map[string]int
}

// NewPerformanceMonitor 新しいパフォーマンス監視を作成
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		apiCallDurations: make(map[string][]time.Duration),
		memorySnapshots:  make([]int64, 0),
		queryFrequency:   make(map[string]int),
	}
}

// RecordAPICall API呼び出し時間を記録
func (p *PerformanceMonitor) RecordAPICall(operation string, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.apiCallDurations[operation] = append(p.apiCallDurations[operation], duration)
}

// RecordMemorySnapshot メモリ使用量のスナップショットを記録
func (p *PerformanceMonitor) RecordMemorySnapshot(usage int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.memorySnapshots = append(p.memorySnapshots, usage)
}

// GetAverageAPICallTime 平均API呼び出し時間を取得
func (p *PerformanceMonitor) GetAverageAPICallTime(operation string) time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()

	durations := p.apiCallDurations[operation]
	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total / time.Duration(len(durations))
}

// EntityPool エンティティ用オブジェクトプール
type EntityPool struct {
	pool chan *ModEntityAPIImpl
	size int
}

// NewEntityPool 新しいエンティティプールを作成
func NewEntityPool(size int) *EntityPool {
	return &EntityPool{
		pool: make(chan *ModEntityAPIImpl, size),
		size: size,
	}
}

// Get プールからエンティティAPIを取得
func (p *EntityPool) Get() *ModEntityAPIImpl {
	select {
	case entity := <-p.pool:
		return entity
	default:
		return &ModEntityAPIImpl{
			entities:      make(map[ecs.EntityID][]string),
			ownedEntities: make(map[ecs.EntityID]bool),
		}
	}
}

// Put エンティティAPIをプールに返却
func (p *EntityPool) Put(entity *ModEntityAPIImpl) {
	// リセット処理
	for k := range entity.entities {
		delete(entity.entities, k)
	}
	for k := range entity.ownedEntities {
		delete(entity.ownedEntities, k)
	}
	entity.nextID = 0

	select {
	case p.pool <- entity:
	default:
		// プール満杯時は破棄
	}
}

// ComponentCache コンポーネントキャッシュ
type ComponentCache struct {
	mu     sync.RWMutex
	cache  map[cacheKey]ecs.Component
	hits   int64
	misses int64
}

type cacheKey struct {
	entityID      ecs.EntityID
	componentType ecs.ComponentType
}

// NewComponentCache 新しいコンポーネントキャッシュを作成
func NewComponentCache() *ComponentCache {
	return &ComponentCache{
		cache: make(map[cacheKey]ecs.Component),
	}
}

// Get キャッシュからコンポーネントを取得
func (c *ComponentCache) Get(entityID ecs.EntityID, componentType ecs.ComponentType) (ecs.Component, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := cacheKey{entityID: entityID, componentType: componentType}
	component, exists := c.cache[key]

	if exists {
		c.hits++
	} else {
		c.misses++
	}

	return component, exists
}

// Set コンポーネントをキャッシュに設定
func (c *ComponentCache) Set(entityID ecs.EntityID, componentType ecs.ComponentType, component ecs.Component) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey{entityID: entityID, componentType: componentType}
	c.cache[key] = component
}

// Remove キャッシュからコンポーネントを削除
func (c *ComponentCache) Remove(entityID ecs.EntityID, componentType ecs.ComponentType) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey{entityID: entityID, componentType: componentType}
	delete(c.cache, key)
}

// GetHitRate キャッシュヒット率を取得
func (c *ComponentCache) GetHitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	if total == 0 {
		return 0
	}
	return float64(c.hits) / float64(total)
}
