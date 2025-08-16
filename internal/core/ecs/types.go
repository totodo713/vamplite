// Package ecs provides the core Entity Component System framework for Muscle Dreamer.
package ecs

// ==============================================
// Core ECS Types - 基本型定義 (不足している型のみ)
// ==============================================

// EntityID represents a unique entity identifier.
type EntityID uint64

// ComponentType represents a component type identifier.
type ComponentType uint16

// SystemType represents a system type identifier.
type SystemType uint16

// SystemID represents a system instance identifier.
type SystemID uint16

// Priority represents execution priority for systems.
type Priority int

// ThreadSafetyLevel represents thread safety requirements.
type ThreadSafetyLevel int

// ==============================================
// Constants - 定数定義
// ==============================================

const (
	// InvalidEntityID represents an invalid entity ID
	InvalidEntityID EntityID = 0
	
	// InvalidComponentType represents an invalid component type
	InvalidComponentType ComponentType = 0
	
	// InvalidSystemType represents an invalid system type
	InvalidSystemType SystemType = 0
	
	// InvalidSystemID represents an invalid system ID
	InvalidSystemID SystemID = 0
)

// Priority constants
const (
	PriorityLowest  Priority = 0
	PriorityLow     Priority = 25
	PriorityNormal  Priority = 50
	PriorityHigh    Priority = 75
	PriorityHighest Priority = 100
)

// Thread safety levels
const (
	ThreadSafetyNone ThreadSafetyLevel = iota
	ThreadSafetyRead
	ThreadSafetyWrite
	ThreadSafetyFull
)

// ==============================================
// Missing Support Types - 不足している補助型
// ==============================================

// StorageStats represents storage performance metrics.
type StorageStats struct {
	TotalEntities     int     `json:"totalEntities"`
	ComponentCount    int     `json:"componentCount"`
	MemoryUsage       int64   `json:"memoryUsage"`
	AverageAccessTime float64 `json:"averageAccessTime"`
}

// SystemMetrics represents system performance data.
type SystemMetrics struct {
	ExecutionTime    float64 `json:"executionTime"`
	EntitiesProcessed int    `json:"entitiesProcessed"`
	MemoryUsage      int64   `json:"memoryUsage"`
	ErrorCount       int     `json:"errorCount"`
}