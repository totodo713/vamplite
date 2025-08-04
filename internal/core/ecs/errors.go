package ecs

import (
	"fmt"
	"time"
)

// ==============================================
// Error Interface and Base Types
// ==============================================

// ECSError represents an error specific to the ECS framework.
// Provides detailed context for debugging and error handling.
type ECSError struct {
	Code      string    `json:"code"`                // Error code for programmatic handling
	Message   string    `json:"message"`             // Human-readable error message
	Component string    `json:"component,omitempty"` // Component involved in error
	Entity    EntityID  `json:"entity,omitempty"`    // Entity involved in error
	System    string    `json:"system,omitempty"`    // System that caused the error
	Timestamp time.Time `json:"timestamp"`           // When the error occurred
	Details   string    `json:"details,omitempty"`   // Additional error details
}

// Error implements the error interface.
func (e *ECSError) Error() string {
	if e.Entity != InvalidEntityID && e.Component != "" {
		return fmt.Sprintf("[%s] %s (Entity: %d, Component: %s)", e.Code, e.Message, e.Entity, e.Component)
	}
	if e.Entity != InvalidEntityID {
		return fmt.Sprintf("[%s] %s (Entity: %d)", e.Code, e.Message, e.Entity)
	}
	if e.Component != "" {
		return fmt.Sprintf("[%s] %s (Component: %s)", e.Code, e.Message, e.Component)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// String returns a detailed string representation for debugging.
func (e *ECSError) String() string {
	return fmt.Sprintf("ECSError{Code: %s, Message: %s, Entity: %d, Component: %s, System: %s, Time: %s}",
		e.Code, e.Message, e.Entity, e.Component, e.System, e.Timestamp.Format(time.RFC3339))
}

// IsRecoverable returns true if the error can be recovered from.
func (e *ECSError) IsRecoverable() bool {
	switch e.Code {
	case ErrEntityNotFound, ErrComponentNotFound, ErrSystemNotFound:
		return true // These are often temporary states
	case ErrMemoryLimit, ErrResourceExhausted:
		return false // Resource exhaustion is typically fatal
	case ErrCircularDependency, ErrInvalidEntityID:
		return false // Design/logic errors are not recoverable
	default:
		return true // Conservative approach - assume recoverable
	}
}

// GetSeverity returns the severity level of the error.
func (e *ECSError) GetSeverity() ErrorSeverity {
	switch e.Code {
	case ErrEntityNotFound, ErrComponentNotFound, ErrSystemNotFound:
		return SeverityWarning
	case ErrInvalidEntityID, ErrComponentExists, ErrSystemExists:
		return SeverityError
	case ErrCircularDependency, ErrMemoryLimit, ErrResourceExhausted, ErrPermissionDenied:
		return SeverityCritical
	default:
		return SeverityError
	}
}

// ==============================================
// Error Severity Levels
// ==============================================

// ErrorSeverity defines the severity level of errors.
type ErrorSeverity int

const (
	SeverityInfo     ErrorSeverity = iota // Informational messages
	SeverityWarning                       // Warning conditions
	SeverityError                         // Error conditions
	SeverityCritical                      // Critical conditions that may cause system failure
)

// String returns the string representation of error severity.
func (s ErrorSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// ==============================================
// Common Error Codes
// ==============================================

const (
	// Entity-related errors
	ErrEntityNotFound      = "ENTITY_NOT_FOUND"      // Entity does not exist
	ErrInvalidEntityID     = "INVALID_ENTITY_ID"     // EntityID is invalid (0 or corrupted)
	ErrEntityAlreadyExists = "ENTITY_ALREADY_EXISTS" // Entity already exists
	ErrEntityLimitReached  = "ENTITY_LIMIT_REACHED"  // Maximum entities reached

	// Component-related errors
	ErrComponentNotFound     = "COMPONENT_NOT_FOUND"     // Component not attached to entity
	ErrComponentExists       = "COMPONENT_EXISTS"        // Component already exists on entity
	ErrComponentTypeMismatch = "COMPONENT_TYPE_MISMATCH" // Component type doesn't match expected
	ErrInvalidComponentType  = "INVALID_COMPONENT_TYPE"  // ComponentType is invalid

	// System-related errors
	ErrSystemNotFound     = "SYSTEM_NOT_FOUND"    // System not registered
	ErrSystemExists       = "SYSTEM_EXISTS"       // System already registered
	ErrSystemDisabled     = "SYSTEM_DISABLED"     // System is currently disabled
	ErrCircularDependency = "CIRCULAR_DEPENDENCY" // Systems have circular dependencies
	ErrSystemTimeout      = "SYSTEM_TIMEOUT"      // System execution exceeded time limit

	// Memory and resource errors
	ErrMemoryLimit       = "MEMORY_LIMIT_EXCEEDED" // Memory usage exceeded limit
	ErrResourceExhausted = "RESOURCE_EXHAUSTED"    // System resources exhausted
	ErrAllocationFailed  = "ALLOCATION_FAILED"     // Memory allocation failed

	// Security and permission errors
	ErrPermissionDenied = "PERMISSION_DENIED" // Operation not permitted
	ErrSandboxViolation = "SANDBOX_VIOLATION" // MOD attempted restricted operation
	ErrInvalidOperation = "INVALID_OPERATION" // Operation not valid in current state

	// Query-related errors
	ErrInvalidQuery   = "INVALID_QUERY"    // Query syntax or logic error
	ErrQueryTimeout   = "QUERY_TIMEOUT"    // Query execution timeout
	ErrQueryCacheFull = "QUERY_CACHE_FULL" // Query cache capacity exceeded

	// Concurrency errors
	ErrConcurrencyViolation = "CONCURRENCY_VIOLATION" // Thread safety violation
	ErrDeadlock             = "DEADLOCK"              // Potential deadlock detected
	ErrRaceCondition        = "RACE_CONDITION"        // Race condition detected

	// Configuration errors
	ErrInvalidConfig = "INVALID_CONFIG" // Configuration is invalid
	ErrConfigMissing = "CONFIG_MISSING" // Required configuration missing

	// General errors
	ErrInitializationFailed = "INITIALIZATION_FAILED" // Component/system initialization failed
	ErrShutdownFailed       = "SHUTDOWN_FAILED"       // Clean shutdown failed
	ErrInternalError        = "INTERNAL_ERROR"        // Unexpected internal error
)

// ==============================================
// Error Factory Functions
// ==============================================

// NewECSError creates a new ECS error with the current timestamp.
func NewECSError(code, message string) *ECSError {
	return &ECSError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// NewEntityError creates an entity-specific error.
func NewEntityError(code, message string, entityID EntityID) *ECSError {
	return &ECSError{
		Code:      code,
		Message:   message,
		Entity:    entityID,
		Timestamp: time.Now(),
	}
}

// NewComponentError creates a component-specific error.
func NewComponentError(code, message string, entityID EntityID, componentType ComponentType) *ECSError {
	return &ECSError{
		Code:      code,
		Message:   message,
		Entity:    entityID,
		Component: string(componentType),
		Timestamp: time.Now(),
	}
}

// NewSystemError creates a system-specific error.
func NewSystemError(code, message string, systemType SystemType) *ECSError {
	return &ECSError{
		Code:      code,
		Message:   message,
		System:    string(systemType),
		Timestamp: time.Now(),
	}
}

// NewMemoryError creates a memory-related error with additional details.
func NewMemoryError(code, message string, details string) *ECSError {
	return &ECSError{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// ==============================================
// Error Helper Functions
// ==============================================

// IsEntityNotFound checks if an error is an entity not found error.
func IsEntityNotFound(err error) bool {
	if ecsErr, ok := err.(*ECSError); ok {
		return ecsErr.Code == ErrEntityNotFound
	}
	return false
}

// IsComponentNotFound checks if an error is a component not found error.
func IsComponentNotFound(err error) bool {
	if ecsErr, ok := err.(*ECSError); ok {
		return ecsErr.Code == ErrComponentNotFound
	}
	return false
}

// IsSystemError checks if an error is system-related.
func IsSystemError(err error) bool {
	if ecsErr, ok := err.(*ECSError); ok {
		return ecsErr.Code == ErrSystemNotFound ||
			ecsErr.Code == ErrSystemExists ||
			ecsErr.Code == ErrSystemDisabled ||
			ecsErr.Code == ErrCircularDependency ||
			ecsErr.Code == ErrSystemTimeout
	}
	return false
}

// IsMemoryError checks if an error is memory-related.
func IsMemoryError(err error) bool {
	if ecsErr, ok := err.(*ECSError); ok {
		return ecsErr.Code == ErrMemoryLimit ||
			ecsErr.Code == ErrResourceExhausted ||
			ecsErr.Code == ErrAllocationFailed
	}
	return false
}

// IsSecurityError checks if an error is security-related.
func IsSecurityError(err error) bool {
	if ecsErr, ok := err.(*ECSError); ok {
		return ecsErr.Code == ErrPermissionDenied ||
			ecsErr.Code == ErrSandboxViolation ||
			ecsErr.Code == ErrInvalidOperation
	}
	return false
}

// IsCriticalError checks if an error is critical and may require system shutdown.
func IsCriticalError(err error) bool {
	if ecsErr, ok := err.(*ECSError); ok {
		return ecsErr.GetSeverity() == SeverityCritical
	}
	return false
}

// ==============================================
// Error Wrapping and Context
// ==============================================

// WrapError wraps an existing error with ECS context.
func WrapError(err error, code, message string) *ECSError {
	return &ECSError{
		Code:      code,
		Message:   fmt.Sprintf("%s: %v", message, err),
		Timestamp: time.Now(),
	}
}

// WithEntity adds entity context to an existing ECS error.
func (e *ECSError) WithEntity(entityID EntityID) *ECSError {
	e.Entity = entityID
	return e
}

// WithComponent adds component context to an existing ECS error.
func (e *ECSError) WithComponent(componentType ComponentType) *ECSError {
	e.Component = string(componentType)
	return e
}

// WithSystem adds system context to an existing ECS error.
func (e *ECSError) WithSystem(systemType SystemType) *ECSError {
	e.System = string(systemType)
	return e
}

// WithDetails adds additional details to an existing ECS error.
func (e *ECSError) WithDetails(details string) *ECSError {
	e.Details = details
	return e
}

// ==============================================
// Predefined Common Errors
// ==============================================

// Common entity errors
var (
	EntityNotFoundErr = func(id EntityID) *ECSError {
		return NewEntityError(ErrEntityNotFound, fmt.Sprintf("Entity %d not found", id), id)
	}

	InvalidEntityIDErr = func(id EntityID) *ECSError {
		return NewEntityError(ErrInvalidEntityID, fmt.Sprintf("Invalid entity ID: %d", id), id)
	}

	EntityLimitReachedErr = func(limit int) *ECSError {
		return NewECSError(ErrEntityLimitReached, fmt.Sprintf("Entity limit of %d reached", limit))
	}
)

// Common component errors
var (
	ComponentNotFoundErr = func(entityID EntityID, componentType ComponentType) *ECSError {
		return NewComponentError(ErrComponentNotFound,
			fmt.Sprintf("Component %s not found on entity %d", componentType, entityID),
			entityID, componentType)
	}

	ComponentExistsErr = func(entityID EntityID, componentType ComponentType) *ECSError {
		return NewComponentError(ErrComponentExists,
			fmt.Sprintf("Component %s already exists on entity %d", componentType, entityID),
			entityID, componentType)
	}
)

// Common system errors
var (
	SystemNotFoundErr = func(systemType SystemType) *ECSError {
		return NewSystemError(ErrSystemNotFound,
			fmt.Sprintf("System %s not found", systemType), systemType)
	}

	CircularDependencyErr = func(systems []SystemType) *ECSError {
		return NewECSError(ErrCircularDependency,
			fmt.Sprintf("Circular dependency detected in systems: %v", systems))
	}
)

// Common memory errors
var (
	MemoryLimitErr = func(used, limit int64) *ECSError {
		return NewMemoryError(ErrMemoryLimit,
			fmt.Sprintf("Memory limit exceeded: %d bytes used, %d bytes limit", used, limit),
			fmt.Sprintf("Used: %d, Limit: %d", used, limit))
	}
)
