// Package ecs provides the core Entity Component System framework for Muscle Dreamer.
//
// This package implements a high-performance ECS architecture designed for
// 2D game development, supporting 10,000+ entities at 60 FPS with memory-efficient
// component storage and parallel system execution.
package ecs

import (
	"sync"
	"time"
)

// ==============================================
// Basic Types
// ==============================================

// EntityID represents a unique entity identifier.
// Uses uint64 for maximum performance and large entity capacity.
type EntityID uint64

// ComponentType represents the type of a component.
// String-based for human readability and debugging ease.
type ComponentType string

// SystemType represents the type of a system.
// String-based for clear system identification and dependency management.
type SystemType string

// Priority defines execution priority for systems.
// Higher values execute first, allowing fine-grained control over execution order.
type Priority int

// Priority constants for common system execution order
const (
	PriorityLowest  Priority = 0   // Background/cleanup systems
	PriorityLow     Priority = 25  // Non-critical systems
	PriorityNormal  Priority = 50  // Default priority
	PriorityHigh    Priority = 75  // Important game logic
	PriorityHighest Priority = 100 // Critical input/physics systems
)

// ==============================================
// Performance Metrics Types
// ==============================================

// PerformanceMetrics contains real-time performance data for the ECS framework.
type PerformanceMetrics struct {
	EntityCount    int           `json:"entity_count"`    // Current active entities
	ComponentCount int           `json:"component_count"` // Total component instances
	SystemCount    int           `json:"system_count"`    // Registered systems
	MemoryUsage    int64         `json:"memory_usage"`    // Memory usage in bytes
	FrameTime      time.Duration `json:"frame_time"`      // Last frame processing time
	UpdateTime     time.Duration `json:"update_time"`     // System update time
	RenderTime     time.Duration `json:"render_time"`     // Rendering time
	QueryTime      time.Duration `json:"query_time"`      // Entity query time
	GCTime         time.Duration `json:"gc_time"`         // Garbage collection time
	Timestamp      time.Time     `json:"timestamp"`       // Measurement timestamp

	// Performance targets tracking
	TargetFPS          float64 `json:"target_fps"`           // Target FPS (60)
	ActualFPS          float64 `json:"actual_fps"`           // Measured FPS
	MemoryLimitBytes   int64   `json:"memory_limit_bytes"`   // Memory limit (256MB)
	MemoryUsagePercent float64 `json:"memory_usage_percent"` // Memory usage percentage
}

// StorageStats contains component storage statistics for memory optimization.
type StorageStats struct {
	ComponentType  ComponentType `json:"component_type"`  // Component type name
	ComponentCount int           `json:"component_count"` // Number of instances
	MemoryUsed     int64         `json:"memory_used"`     // Actual memory used
	MemoryReserved int64         `json:"memory_reserved"` // Reserved memory
	Fragmentation  float64       `json:"fragmentation"`   // Memory fragmentation ratio
	AccessCount    int64         `json:"access_count"`    // Access frequency
	CacheHitRate   float64       `json:"cache_hit_rate"`  // Cache efficiency
}

// Note: QueryStats is defined in query.go

// MemoryUsage contains detailed memory usage information.
type MemoryUsage struct {
	TotalAllocated int64     `json:"total_allocated"` // Total allocated memory
	TotalReserved  int64     `json:"total_reserved"`  // Reserved but unused
	TotalFree      int64     `json:"total_free"`      // Available memory
	Fragmentation  float64   `json:"fragmentation"`   // Memory fragmentation
	GCCount        int64     `json:"gc_count"`        // GC cycles count
	LastGCTime     time.Time `json:"last_gc_time"`    // Last GC execution
}

// ==============================================
// Threading and Synchronization Types
// ==============================================

// ThreadSafetyLevel defines the thread safety level of a system.
type ThreadSafetyLevel int

const (
	ThreadSafetyNone  ThreadSafetyLevel = iota // No thread safety - single thread only
	ThreadSafetyRead                           // Read-only operations are thread safe
	ThreadSafetyWrite                          // Limited write operations safe
	ThreadSafetyFull                           // Fully thread safe
)

// ==============================================
// Configuration Types
// ==============================================

// WorldConfig contains world initialization parameters.
type WorldConfig struct {
	MaxEntities    int           `json:"max_entities"`     // Maximum entities (10,000)
	MemoryLimit    int64         `json:"memory_limit"`     // Memory limit in bytes (256MB)
	EnableMetrics  bool          `json:"enable_metrics"`   // Enable performance monitoring
	EnableEvents   bool          `json:"enable_events"`    // Enable event system
	ThreadPoolSize int           `json:"thread_pool_size"` // Parallel system threads
	QueryCacheSize int           `json:"query_cache_size"` // Query cache capacity
	GCInterval     time.Duration `json:"gc_interval"`      // GC frequency

	// Performance tuning
	ComponentPoolSize int `json:"component_pool_size"` // Component memory pool size
	EntityPoolSize    int `json:"entity_pool_size"`    // Entity ID pool size
	SystemBatchSize   int `json:"system_batch_size"`   // System batch processing size
	CacheLineSize     int `json:"cache_line_size"`     // CPU cache line size (64)

	// Debug and development
	EnableDebugMode bool `json:"enable_debug_mode"` // Debug information
	EnableProfiling bool `json:"enable_profiling"`  // Performance profiling
	LogLevel        int  `json:"log_level"`         // Logging verbosity (0-4)
}

// DefaultWorldConfig returns a default configuration optimized for game development.
func DefaultWorldConfig() WorldConfig {
	return WorldConfig{
		MaxEntities:       10000,
		MemoryLimit:       256 * 1024 * 1024, // 256MB
		EnableMetrics:     true,
		EnableEvents:      true,
		ThreadPoolSize:    4,
		QueryCacheSize:    1000,
		GCInterval:        30 * time.Second,
		ComponentPoolSize: 1000,
		EntityPoolSize:    1000,
		SystemBatchSize:   100,
		CacheLineSize:     64,
		EnableDebugMode:   false,
		EnableProfiling:   false,
		LogLevel:          2, // Info level
	}
}

// ==============================================
// Utility Types
// ==============================================

// Vector2 represents a 2D vector for positions, velocities, etc.
type Vector2 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// AABB (Axis-Aligned Bounding Box) for collision detection.
type AABB struct {
	Min Vector2 `json:"min"` // Minimum point
	Max Vector2 `json:"max"` // Maximum point
}

// Color represents RGBA color values.
type Color struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
	A uint8 `json:"a"`
}

// TransformMatrix represents a 3x3 2D transformation matrix in column-major order.
type TransformMatrix [9]float64

// ==============================================
// Thread-Safe Collections
// ==============================================

// SafeEntitySet provides thread-safe entity collections.
type SafeEntitySet struct {
	entities map[EntityID]bool
	mutex    sync.RWMutex
}

// NewSafeEntitySet creates a new thread-safe entity set.
func NewSafeEntitySet() *SafeEntitySet {
	return &SafeEntitySet{
		entities: make(map[EntityID]bool),
	}
}

// Add adds an entity to the set.
func (s *SafeEntitySet) Add(entity EntityID) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.entities[entity] = true
}

// Remove removes an entity from the set.
func (s *SafeEntitySet) Remove(entity EntityID) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.entities, entity)
}

// Contains checks if an entity exists in the set.
func (s *SafeEntitySet) Contains(entity EntityID) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.entities[entity]
}

// ToSlice returns all entities as a slice.
func (s *SafeEntitySet) ToSlice() []EntityID {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entities := make([]EntityID, 0, len(s.entities))
	for entity := range s.entities {
		entities = append(entities, entity)
	}
	return entities
}

// Len returns the number of entities in the set.
func (s *SafeEntitySet) Len() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.entities)
}

// Clear removes all entities from the set.
func (s *SafeEntitySet) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.entities = make(map[EntityID]bool)
}

// ==============================================
// Constants
// ==============================================

const (
	// Performance targets
	TargetFPS              = 60                    // Target frames per second
	MaxEntityCount         = 10000                 // Maximum supported entities
	MaxComponentsPerFrame  = 50000                 // Max component operations per frame
	MaxSystemExecutionTime = 10 * time.Millisecond // Max system execution time
	MaxQueryTime           = 1 * time.Millisecond  // Max query execution time
	MaxMemoryPerEntity     = 100                   // Max memory bytes per entity

	// Memory management
	DefaultMemoryLimit = 256 * 1024 * 1024 // 256MB default limit
	ComponentAlignment = 64                // Memory alignment for components
	CacheLineSize      = 64                // CPU cache line size

	// Threading
	DefaultThreadPoolSize = 4 // Default parallel threads
	MaxConcurrentSystems  = 8 // Max systems running in parallel

	// Cache and pools
	DefaultQueryCacheSize = 1000 // Default query cache capacity
	DefaultPoolSize       = 1000 // Default object pool size

	// Invalid values
	InvalidEntityID      EntityID      = 0  // Invalid entity identifier
	InvalidComponentType ComponentType = "" // Invalid component type
	InvalidSystemType    SystemType    = "" // Invalid system type
)

// Component type constants for built-in components
const (
	ComponentTypeTransform ComponentType = "transform"
	ComponentTypeSprite    ComponentType = "sprite"
	ComponentTypePhysics   ComponentType = "physics"
	ComponentTypeHealth    ComponentType = "health"
	ComponentTypeAI        ComponentType = "ai"
	ComponentTypeInventory ComponentType = "inventory"
	ComponentTypeAudio     ComponentType = "audio"
	ComponentTypeInput     ComponentType = "input"
)

// System type constants for built-in systems
const (
	SystemTypeInput     SystemType = "input"
	SystemTypeAI        SystemType = "ai"
	SystemTypePhysics   SystemType = "physics"
	SystemTypeMovement  SystemType = "movement"
	SystemTypeCollision SystemType = "collision"
	SystemTypeAnimation SystemType = "animation"
	SystemTypeAudio     SystemType = "audio"
	SystemTypeRendering SystemType = "rendering"
	SystemTypeUI        SystemType = "ui"
	SystemTypeDebug     SystemType = "debug"
)
