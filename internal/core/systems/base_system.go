// Package systems provides the core game systems for the ECS framework.
//
// This package implements the basic systems required for game functionality:
// Movement, Rendering, Physics, and Audio systems. All systems implement
// the System interface and provide thread-safe, high-performance operations.
package systems

import (
	"sync"
	"time"

	"muscle-dreamer/internal/core/ecs"
)

// BaseSystem provides common functionality for all systems.
// This includes metrics collection, state management, and error handling.
type BaseSystem struct {
	systemType SystemType
	priority   ecs.Priority
	enabled    bool
	metrics    *ecs.SystemMetrics
	mutex      sync.RWMutex

	// Error handling
	errorHandler func(error)
	lastError    error
}

// SystemType represents the type of a system (alias for clarity)
type SystemType = ecs.SystemType

// System priority constants for basic systems
const (
	InputSystemPriority     ecs.Priority = 100 // Highest - process input first
	MovementSystemPriority  ecs.Priority = 90  // High - process movement
	PhysicsSystemPriority   ecs.Priority = 80  // High - physics simulation
	RenderingSystemPriority ecs.Priority = 20  // Low - render after logic
	AudioSystemPriority     ecs.Priority = 30  // Low - audio after physics
)

// System type constants
const (
	MovementSystemType  SystemType = ecs.SystemTypeMovement
	RenderingSystemType SystemType = ecs.SystemTypeRendering
	PhysicsSystemType   SystemType = ecs.SystemTypePhysics
	AudioSystemType     SystemType = ecs.SystemTypeAudio
)

// NewBaseSystem creates a new base system with the given type and priority.
func NewBaseSystem(systemType SystemType, priority ecs.Priority) *BaseSystem {
	return &BaseSystem{
		systemType: systemType,
		priority:   priority,
		enabled:    true,
		metrics: &ecs.SystemMetrics{
			SystemType:       systemType,
			ExecutionCount:   0,
			TotalTime:        0,
			AverageTime:      0,
			MaxTime:          0,
			MinTime:          0,
			ErrorCount:       0,
			LastExecution:    time.Now().UnixNano(),
			EntitysProcessed: 0,
			MemoryAllocated:  0,
		},
	}
}

// GetType returns the system type for identification.
func (bs *BaseSystem) GetType() ecs.SystemType {
	return bs.systemType
}

// GetPriority returns the execution priority.
func (bs *BaseSystem) GetPriority() ecs.Priority {
	return bs.priority
}

// SetPriority sets the execution priority.
func (bs *BaseSystem) SetPriority(priority ecs.Priority) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	bs.priority = priority
}

// GetDependencies returns an empty slice (no dependencies by default).
func (bs *BaseSystem) GetDependencies() []ecs.SystemType {
	return []ecs.SystemType{}
}

// GetRequiredComponents returns an empty slice (to be overridden).
func (bs *BaseSystem) GetRequiredComponents() []ecs.ComponentType {
	return []ecs.ComponentType{}
}

// Initialize sets up the system (empty implementation).
func (bs *BaseSystem) Initialize(world ecs.World) error {
	return nil
}

// Update processes entities (empty implementation).
func (bs *BaseSystem) Update(world ecs.World, deltaTime float64) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	start := time.Now()
	defer func() {
		elapsed := time.Since(start).Nanoseconds()
		bs.metrics.ExecutionCount++
		bs.metrics.TotalTime += elapsed
		bs.metrics.LastExecution = start.UnixNano()

		if bs.metrics.ExecutionCount > 0 {
			bs.metrics.AverageTime = bs.metrics.TotalTime / bs.metrics.ExecutionCount
		}

		if elapsed > bs.metrics.MaxTime {
			bs.metrics.MaxTime = elapsed
		}
		if bs.metrics.MinTime == 0 || elapsed < bs.metrics.MinTime {
			bs.metrics.MinTime = elapsed
		}
	}()

	return nil
}

// Render draws entities (empty implementation).
func (bs *BaseSystem) Render(world ecs.World, renderer interface{}) error {
	// Rendering systems use the same metrics as Update
	return bs.Update(world, 0) // deltaTime not used in render
}

// Shutdown cleans up system resources (empty implementation).
func (bs *BaseSystem) Shutdown() error {
	return nil
}

// IsEnabled returns whether the system is currently active.
func (bs *BaseSystem) IsEnabled() bool {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	return bs.enabled
}

// SetEnabled controls system execution.
func (bs *BaseSystem) SetEnabled(enabled bool) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	bs.enabled = enabled
}

// GetMetrics returns system performance data.
func (bs *BaseSystem) GetMetrics() *ecs.SystemMetrics {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	// Return a copy to prevent concurrent access
	metrics := *bs.metrics
	return &metrics
}

// GetThreadSafety returns the thread safety level.
func (bs *BaseSystem) GetThreadSafety() ecs.ThreadSafetyLevel {
	return ecs.ThreadSafetyFull // BaseSystem is fully thread-safe
}

// CanRunInParallel returns true if system can run concurrently.
func (bs *BaseSystem) CanRunInParallel() bool {
	return true // BaseSystem supports parallel execution
}

// SetErrorHandler sets a custom error handler for the system.
func (bs *BaseSystem) SetErrorHandler(handler func(error)) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	bs.errorHandler = handler
}

// handleError processes an error through the error handler.
func (bs *BaseSystem) handleError(err error) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	bs.lastError = err
	bs.metrics.ErrorCount++

	if bs.errorHandler != nil {
		bs.errorHandler(err)
	}
}

// GetLastError returns the last error that occurred.
func (bs *BaseSystem) GetLastError() error {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	return bs.lastError
}

// ResetMetrics clears all metrics data.
func (bs *BaseSystem) ResetMetrics() {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	bs.metrics.ExecutionCount = 0
	bs.metrics.TotalTime = 0
	bs.metrics.AverageTime = 0
	bs.metrics.MaxTime = 0
	bs.metrics.MinTime = 0
	bs.metrics.ErrorCount = 0
	bs.metrics.LastExecution = time.Now().UnixNano()
	bs.metrics.EntitysProcessed = 0
	bs.metrics.MemoryAllocated = 0
}
