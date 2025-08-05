package storage

import (
	"errors"
	"fmt"

	"muscle-dreamer/internal/core/ecs"
	"muscle-dreamer/internal/core/ecs/components"
)

// PoolStatistics contains memory pool usage statistics
type PoolStatistics struct {
	ComponentType   ecs.ComponentType `json:"component_type"`
	UsedCount       int               `json:"used_count"`
	Capacity        int               `json:"capacity"`
	AvailableCount  int               `json:"available_count"`
	UsageRatio      float64           `json:"usage_ratio"`
	MemoryUsed      int64             `json:"memory_used"`
	MemoryAllocated int64             `json:"memory_allocated"`
}

// MemoryPool provides efficient component allocation and deallocation
type MemoryPool struct {
	componentType ecs.ComponentType
	pool          []ecs.Component
	available     []ecs.Component
	used          map[ecs.Component]bool
	capacity      int
}

// NewMemoryPool creates a new memory pool for a specific component type
func NewMemoryPool(componentType ecs.ComponentType, capacity int) *MemoryPool {
	pool := &MemoryPool{
		componentType: componentType,
		pool:          make([]ecs.Component, 0, capacity),
		available:     make([]ecs.Component, 0, capacity),
		used:          make(map[ecs.Component]bool),
		capacity:      capacity,
	}

	// Pre-allocate components
	for i := 0; i < capacity; i++ {
		component := pool.createComponent()
		if component != nil {
			pool.pool = append(pool.pool, component)
			pool.available = append(pool.available, component)
		}
	}

	return pool
}

// createComponent creates a new component instance based on the component type
func (p *MemoryPool) createComponent() ecs.Component {
	switch p.componentType {
	case ecs.ComponentTypeTransform:
		return components.NewTransformComponent()
	case ecs.ComponentTypeSprite:
		return components.NewSpriteComponent()
	case ecs.ComponentTypePhysics:
		return components.NewPhysicsComponent()
	case ecs.ComponentTypeHealth:
		return components.NewHealthComponent(100) // Default max health
	case ecs.ComponentTypeAI:
		return components.NewAIComponent()
	default:
		return nil
	}
}

// GetComponentType returns the component type this pool manages
func (p *MemoryPool) GetComponentType() ecs.ComponentType {
	return p.componentType
}

// GetUsedCount returns the number of components currently in use
func (p *MemoryPool) GetUsedCount() int {
	return len(p.used)
}

// GetCapacity returns the maximum capacity of the pool
func (p *MemoryPool) GetCapacity() int {
	return p.capacity
}

// GetAvailableCount returns the number of available components
func (p *MemoryPool) GetAvailableCount() int {
	return len(p.available)
}

// Acquire gets a component from the pool
func (p *MemoryPool) Acquire() (ecs.Component, error) {
	if len(p.available) == 0 {
		return nil, fmt.Errorf("pool capacity exceeded: no available components for type %s", p.componentType)
	}

	// Get component from available list
	component := p.available[len(p.available)-1]
	p.available = p.available[:len(p.available)-1]

	// Mark as used
	p.used[component] = true

	return component, nil
}

// Release returns a component to the pool
func (p *MemoryPool) Release(component ecs.Component) error {
	if component == nil {
		return errors.New("cannot release nil component")
	}

	if component.GetType() != p.componentType {
		return fmt.Errorf("component does not belong to this pool: expected %s, got %s",
			p.componentType, component.GetType())
	}

	// Check if component is currently acquired
	if !p.used[component] {
		return errors.New("component not currently acquired from this pool")
	}

	// Remove from used set
	delete(p.used, component)

	// Return to available list
	p.available = append(p.available, component)

	return nil
}

// Clear resets the pool, marking all components as available
func (p *MemoryPool) Clear() {
	p.used = make(map[ecs.Component]bool)
	p.available = make([]ecs.Component, len(p.pool))
	copy(p.available, p.pool)
}

// Resize changes the capacity of the pool
func (p *MemoryPool) Resize(newCapacity int) error {
	if newCapacity < len(p.used) {
		return fmt.Errorf("new capacity %d cannot be smaller than used count %d",
			newCapacity, len(p.used))
	}

	if newCapacity > p.capacity {
		// Expand pool
		additionalCapacity := newCapacity - p.capacity
		for i := 0; i < additionalCapacity; i++ {
			component := p.createComponent()
			if component != nil {
				p.pool = append(p.pool, component)
				p.available = append(p.available, component)
			}
		}
	} else if newCapacity < p.capacity {
		// Shrink pool - remove excess available components
		excessCount := p.capacity - newCapacity
		if len(p.available) >= excessCount {
			p.available = p.available[:len(p.available)-excessCount]
			p.pool = p.pool[:newCapacity]
		}
	}

	p.capacity = newCapacity
	return nil
}

// GetStatistics returns detailed statistics about the pool
func (p *MemoryPool) GetStatistics() PoolStatistics {
	usedCount := len(p.used)
	availableCount := len(p.available)
	usageRatio := float64(usedCount) / float64(p.capacity)

	// Estimate memory usage
	var memoryUsed, memoryAllocated int64
	if len(p.pool) > 0 {
		componentSize := int64(p.pool[0].Size())
		memoryUsed = int64(usedCount) * componentSize
		memoryAllocated = int64(p.capacity) * componentSize
	}

	return PoolStatistics{
		ComponentType:   p.componentType,
		UsedCount:       usedCount,
		Capacity:        p.capacity,
		AvailableCount:  availableCount,
		UsageRatio:      usageRatio,
		MemoryUsed:      memoryUsed,
		MemoryAllocated: memoryAllocated,
	}
}
