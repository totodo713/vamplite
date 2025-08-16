// Package ecs provides the core Entity Component System framework for Muscle Dreamer.
package ecs

import "time"

// ==============================================
// Core ECS Types - 基本型定義
// ==============================================

// EntityID represents a unique entity identifier.
type EntityID uint64

// ComponentType represents a component type identifier.
type ComponentType uint16

// ArchetypeID represents an archetype identifier for grouping similar entities.
type ArchetypeID uint32

// SystemID represents a system identifier.
type SystemID uint16

// EventType represents an event type identifier.
type EventType uint16

// ==============================================
// Component Interface - コンポーネント基底インターフェース
// ==============================================

// Component is the base interface that all components must implement.
type Component interface {
	// GetType returns the component type identifier
	GetType() ComponentType
	
	// Clone creates a deep copy of the component
	Clone() Component
	
	// Reset resets the component to its default state for object pooling
	Reset()
}

// ==============================================
// Event Interface - イベント基底インターフェース  
// ==============================================

// Event represents a game event that can be published through the event bus.
type Event interface {
	// GetType returns the event type identifier
	GetType() EventType
	
	// GetTimestamp returns when the event was created
	GetTimestamp() time.Time
	
	// GetEntityID returns the entity associated with this event (if any)
	GetEntityID() EntityID
}

// ==============================================
// Query Interface - エンティティクエリ
// ==============================================

// Query represents a query for finding entities with specific component combinations.
type Query interface {
	// Execute runs the query and returns matching entity IDs
	Execute() []EntityID
	
	// With adds a required component type to the query
	With(ComponentType) Query
	
	// Without adds an excluded component type to the query  
	Without(ComponentType) Query
	
	// Count returns the number of matching entities without allocating
	Count() int
	
	// ForEach iterates over matching entities with a callback
	ForEach(func(EntityID))
}

// ==============================================
// Constants - 定数定義
// ==============================================

const (
	// InvalidEntityID represents an invalid entity ID
	InvalidEntityID EntityID = 0
	
	// InvalidComponentType represents an invalid component type
	InvalidComponentType ComponentType = 0
	
	// InvalidArchetypeID represents an invalid archetype ID
	InvalidArchetypeID ArchetypeID = 0
	
	// InvalidSystemID represents an invalid system ID
	InvalidSystemID SystemID = 0
	
	// InvalidEventType represents an invalid event type
	InvalidEventType EventType = 0
)

// ==============================================
// Result Types - 結果型
// ==============================================

// QueryResult represents the result of a query operation.
type QueryResult struct {
	Entities []EntityID
	Count    int
}

// ComponentResult represents the result of a component operation.
type ComponentResult struct {
	Component Component
	Found     bool
}