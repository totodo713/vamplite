package ecs

import (
	"errors"
	"time"
)

// ==============================================
// Event System Types
// ==============================================

// EventTypeID defines event type identifiers (different from existing EventType)
type EventTypeID uint32

// Event type constants for EventBus system
const (
	EventTypeIDEntityCreated EventTypeID = iota
	EventTypeIDEntityDestroyed
	EventTypeIDComponentAdded
	EventTypeIDComponentRemoved
	EventTypeIDComponentUpdated
	EventTypeIDSystemStarted
	EventTypeIDSystemStopped
	EventTypeIDPlayerDamaged
	EventTypeIDEnemyDefeated
	EventTypeIDItemCollected
)

// EventPriority defines event processing priority
type EventPriority uint8

const (
	EventPriorityLow EventPriority = iota
	EventPriorityNormal
	EventPriorityHigh
	EventPriorityCritical
)

// EventBusEvent defines the interface for all EventBus events (different from existing Event)
type EventBusEvent interface {
	GetType() EventTypeID
	GetEntityID() EntityID
	GetTimestamp() time.Time
	GetPriority() EventPriority
	Validate() error
}

// EventBusEventBase provides common event fields for EventBus events
type EventBusEventBase struct {
	Type      EventTypeID
	EntityID  EntityID
	Timestamp time.Time
	Priority  EventPriority
}

func (e EventBusEventBase) GetType() EventTypeID       { return e.Type }
func (e EventBusEventBase) GetEntityID() EntityID      { return e.EntityID }
func (e EventBusEventBase) GetTimestamp() time.Time    { return e.Timestamp }
func (e EventBusEventBase) GetPriority() EventPriority { return e.Priority }
func (e EventBusEventBase) Validate() error            { return nil }

// Specific event implementations for EventBus

type EntityCreatedBusEvent struct {
	EventBusEventBase
	Components []ComponentType
}

type EntityDestroyedBusEvent struct {
	EventBusEventBase
}

type ComponentAddedBusEvent struct {
	EventBusEventBase
	ComponentType ComponentType
	ComponentData interface{}
}

type ComponentRemovedBusEvent struct {
	EventBusEventBase
	ComponentType ComponentType
}

type ComponentUpdatedBusEvent struct {
	EventBusEventBase
	ComponentType ComponentType
	OldData       interface{}
	NewData       interface{}
}

type SystemStartedBusEvent struct {
	EventBusEventBase
	SystemType SystemType
}

type SystemStoppedBusEvent struct {
	EventBusEventBase
	SystemType SystemType
}

type PlayerDamagedBusEvent struct {
	EventBusEventBase
	Damage       float64
	DamageSource EntityID
}

type EnemyDefeatedBusEvent struct {
	EventBusEventBase
	Reward int
}

type ItemCollectedBusEvent struct {
	EventBusEventBase
	ItemType string
	Amount   int
}

// Generic test event for testing purposes
type GenericTestBusEvent struct {
	EventBusEventBase
	TestData string
}

// Handler types
type HandlerID string
type SubscriptionID uint64

// EventBusHandler defines the interface for EventBus event handlers
type EventBusHandler interface {
	Handle(event EventBusEvent) error
	GetHandlerID() HandlerID
	GetSupportedEventTypes() []EventTypeID
}

// EventBusFilter defines the interface for EventBus event filtering
type EventBusFilter interface {
	Filter(event EventBusEvent) bool
}

// EventBusFilterFunc allows functions to implement EventBusFilter interface
type EventBusFilterFunc func(EventBusEvent) bool

func (f EventBusFilterFunc) Filter(event EventBusEvent) bool {
	return f(event)
}

// EventBusStats provides statistics about event bus performance
type EventBusStats struct {
	EventsPublished    uint64
	EventsProcessed    uint64
	EventsDropped      uint64
	HandlerErrors      uint64
	HandlerPanics      uint64
	AvgLatencyNanos    int64
	TotalSubscriptions int
	QueueSize          int
	WorkerCount        int
}

// EventBusConfig defines configuration for EventBus
type EventBusConfig struct {
	BufferSize     int
	NumWorkers     int
	EnableMetrics  bool
	EnablePriority bool
	MaxHandlers    int
}

// DefaultEventBusConfig returns a default configuration optimized for game development
func DefaultEventBusConfig() *EventBusConfig {
	return &EventBusConfig{
		BufferSize:     1000,
		NumWorkers:     4,
		EnableMetrics:  true,
		EnablePriority: false,
		MaxHandlers:    1000,
	}
}

// EventBus defines the main interface for event bus
type EventBus interface {
	// Lifecycle
	Start() error
	Stop() error
	IsRunning() bool
	Flush() error

	// Event publishing
	Publish(eventType EventTypeID, event EventBusEvent) error
	PublishAsync(eventType EventTypeID, event EventBusEvent) error

	// Subscription management
	Subscribe(eventType EventTypeID, handler EventBusHandler) (SubscriptionID, error)
	Unsubscribe(subscriptionID SubscriptionID) error
	SubscribeWithFilter(eventType EventTypeID, filter EventBusFilter, handler EventBusHandler) (SubscriptionID, error)

	// Information
	GetStats() EventBusStats
	GetSubscriptions() map[SubscriptionID]*EventBusSubscription
}

// EventBusSubscription represents an active EventBus subscription
type EventBusSubscription struct {
	ID      SubscriptionID
	Type    EventTypeID
	Handler EventBusHandler
	Filter  EventBusFilter
	Created time.Time
	Active  bool
}

// Event system errors
var (
	ErrEventBusNotStarted   = errors.New("event bus is not started")
	ErrEventBusStopped      = errors.New("event bus is stopped")
	ErrInvalidEventType     = errors.New("invalid event type")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrHandlerPanic         = errors.New("event handler panic")
	ErrQueueFull            = errors.New("event queue is full")
	ErrEventNil             = errors.New("event cannot be nil")
)
