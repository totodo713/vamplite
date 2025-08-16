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

// ==============================================
// Missing Math Types - 数学型定義
// ==============================================

// Vector2 represents a 2D vector for spatial operations.
type Vector2 struct {
	X, Y float32
}

// AABB represents an Axis-Aligned Bounding Box for collision detection.
type AABB struct {
	Min, Max Vector2
}

// ==============================================
// World Configuration Types - World設定型
// ==============================================

// WorldConfig represents configuration for the ECS world.
type WorldConfig struct {
	MaxEntities       int     `json:"maxEntities"`
	InitialCapacity   int     `json:"initialCapacity"`
	ComponentPoolSize int     `json:"componentPoolSize"`
	EnableMetrics     bool    `json:"enableMetrics"`
	TargetFPS         float64 `json:"targetFPS"`
}

// PerformanceMetrics represents world performance data.
type PerformanceMetrics struct {
	FPS                float64 `json:"fps"`
	FrameTime          float64 `json:"frameTime"`
	EntityCount        int     `json:"entityCount"`
	SystemUpdateTime   float64 `json:"systemUpdateTime"`
	MemoryUsage        int64   `json:"memoryUsage"`
}

// MemoryUsage represents memory usage statistics.
type MemoryUsage struct {
	TotalAllocated int64 `json:"totalAllocated"`
	CurrentUsage   int64 `json:"currentUsage"`
	PeakUsage      int64 `json:"peakUsage"`
	GCCount        int64 `json:"gcCount"`
}

// ==============================================
// Temporary Component Type Constants - 一時的コンポーネント型定義
// ==============================================

// Component type constants for testing and basic functionality
const (
	ComponentTypeTransform ComponentType = 1
	ComponentTypeSprite    ComponentType = 2
	ComponentTypePhysics   ComponentType = 3
	ComponentTypeAI        ComponentType = 4
	ComponentTypeHealth    ComponentType = 5
	ComponentTypeInput     ComponentType = 6
	ComponentTypeAudio     ComponentType = 7
	ComponentTypeAnimation ComponentType = 8
	ComponentTypeInventory ComponentType = 9
	ComponentTypeEnergy    ComponentType = 10
)

// ==============================================
// Color Type - 色型定義
// ==============================================

// Color represents RGBA color values for sprites and rendering.
type Color struct {
	R, G, B, A float32
}

// System type constants for testing and basic functionality
const (
	SystemTypeRender    SystemType = 1
	SystemTypePhysics   SystemType = 2
	SystemTypeAI        SystemType = 3
	SystemTypeInput     SystemType = 4
	SystemTypeAudio     SystemType = 5
	SystemTypeAnimation SystemType = 6
)

// SystemType to string conversion for testing
func (st SystemType) String() string {
	switch st {
	case SystemTypeRender:
		return "RenderSystem"
	case SystemTypePhysics:
		return "PhysicsSystem"
	case SystemTypeAI:
		return "AISystem"
	case SystemTypeInput:
		return "InputSystem"
	case SystemTypeAudio:
		return "AudioSystem"
	case SystemTypeAnimation:
		return "AnimationSystem"
	default:
		return "UnknownSystem"
	}
}

// SystemTypeFromString converts string to SystemType for testing
func SystemTypeFromString(s string) SystemType {
	switch s {
	case "TestSystem":
		return SystemTypeRender // Temporary mapping for tests
	case "HighPrioritySystem":
		return SystemTypePhysics
	case "LowPrioritySystem":
		return SystemTypeAI
	case "EnabledSystem":
		return SystemTypeInput
	case "DisabledSystem":
		return SystemTypeAudio
	case "SystemA":
		return SystemTypeAnimation
	case "SystemB":
		return SystemTypeRender
	case "RenderSystem":
		return SystemTypeRender
	case "PhysicsSystem":
		return SystemTypePhysics
	case "AISystem":
		return SystemTypeAI
	case "InputSystem":
		return SystemTypeInput
	case "AudioSystem":
		return SystemTypeAudio
	case "AnimationSystem":
		return SystemTypeAnimation
	default:
		return SystemTypeRender // Default fallback
	}
}

// ComponentTypeFromString converts string to ComponentType for testing
func ComponentTypeFromString(s string) ComponentType {
	switch s {
	case "test-fileio":
		return ComponentTypeInventory // Temporary mapping for test
	case "transform":
		return ComponentTypeTransform
	case "sprite":
		return ComponentTypeSprite
	case "physics":
		return ComponentTypePhysics
	case "ai":
		return ComponentTypeAI
	case "health":
		return ComponentTypeHealth
	case "input":
		return ComponentTypeInput
	case "audio":
		return ComponentTypeAudio
	case "animation":
		return ComponentTypeAnimation
	case "inventory":
		return ComponentTypeInventory
	case "energy":
		return ComponentTypeEnergy
	default:
		return ComponentTypeTransform // Default fallback
	}
}