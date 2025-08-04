// Package ecs provides the core Entity Component System framework for Muscle Dreamer.
package ecs

import (
	"context"
	"time"
)

// ==============================================
// SystemManager Interface - システム登録・実行管理
// ==============================================

// SystemManager manages system registration, execution order, and lifecycle.
// It handles system dependencies, parallel execution, and performance monitoring.
type SystemManager interface {
	// System registration and lifecycle
	RegisterSystem(System) error
	RegisterSystemWithPriority(System, Priority) error
	UnregisterSystem(SystemType) error
	GetSystem(SystemType) (System, error)
	GetAllSystems() []System
	GetSystemCount() int

	// System execution control
	UpdateSystems(World, float64) error
	RenderSystems(World, interface{}) error
	InitializeSystems(World) error
	ShutdownSystems() error

	// System state management
	EnableSystem(SystemType) error
	DisableSystem(SystemType) error
	IsSystemEnabled(SystemType) bool
	GetEnabledSystems() []SystemType
	GetDisabledSystems() []SystemType

	// Execution order and dependencies
	SetSystemDependency(dependent SystemType, dependency SystemType) error
	RemoveSystemDependency(dependent SystemType, dependency SystemType) error
	GetSystemDependencies(SystemType) []SystemType
	GetSystemDependents(SystemType) []SystemType
	GetExecutionOrder() []SystemType
	ValidateExecutionOrder() error
	RecomputeExecutionOrder() error

	// Parallel execution management
	SetParallelExecution(bool)
	IsParallelExecutionEnabled() bool
	GetParallelGroups() [][]SystemType
	SetMaxParallelSystems(int)
	GetMaxParallelSystems() int

	// Performance monitoring
	GetSystemMetrics(SystemType) (*SystemMetrics, error)
	GetAllSystemMetrics() map[SystemType]*SystemMetrics
	ResetSystemMetrics() error
	EnableProfiling(bool)
	IsProfilingEnabled() bool

	// System filtering and queries
	GetSystemsByComponent(ComponentType) []System
	GetSystemsByPriority(Priority) []System
	GetSystemsByThreadSafety(ThreadSafetyLevel) []System
	FindSystemsByPredicate(func(System) bool) []System

	// Batch operations
	RegisterSystems([]System) error
	UnregisterSystems([]SystemType) error
	EnableSystems([]SystemType) error
	DisableSystems([]SystemType) error

	// Error handling and recovery
	SetErrorHandler(func(SystemType, error) error)
	GetSystemErrors() map[SystemType][]error
	ClearSystemErrors(SystemType) error
	GetFailedSystems() []SystemType

	// Configuration and tuning
	SetSystemTimeout(SystemType, time.Duration) error
	GetSystemTimeout(SystemType) time.Duration
	SetGlobalTimeout(time.Duration)
	GetGlobalTimeout() time.Duration

	// Serialization and persistence
	SerializeSystemState() ([]byte, error)
	DeserializeSystemState([]byte) error
	SaveSystemConfiguration(string) error
	LoadSystemConfiguration(string) error

	// Thread safety
	Lock()
	RLock()
	Unlock()
	RUnlock()

	// Debug and diagnostics
	ValidateIntegrity() error
	GetDebugInfo() *SystemManagerDebugInfo
	DumpExecutionOrder() string
	GetDependencyGraph() *DependencyGraph
}

// ==============================================
// System Execution Pipeline
// ==============================================

// SystemExecutor handles the actual execution of systems with monitoring and error handling.
type SystemExecutor interface {
	// Single system execution
	ExecuteSystem(System, World, float64) error
	ExecuteSystemWithTimeout(System, World, float64, time.Duration) error
	ExecuteSystemAsync(System, World, float64) <-chan error

	// Batch system execution
	ExecuteSystems([]System, World, float64) error
	ExecuteSystemsParallel([]System, World, float64) error
	ExecuteSystemsWithDependencies([]System, World, float64) error

	// Execution control
	PauseExecution()
	ResumeExecution()
	IsPaused() bool
	StopExecution()
	IsRunning() bool

	// Performance monitoring
	GetExecutionMetrics() *ExecutionMetrics
	GetSystemExecutionTime(SystemType) time.Duration
	GetLastExecutionErrors() []ExecutionError

	// Configuration
	SetWorkerPoolSize(int)
	GetWorkerPoolSize() int
	SetExecutionTimeout(time.Duration)
	GetExecutionTimeout() time.Duration
}

// ExecutionMetrics contains performance data for system execution.
type ExecutionMetrics struct {
	TotalExecutions      int64         `json:"total_executions"`
	SuccessfulExecutions int64         `json:"successful_executions"`
	FailedExecutions     int64         `json:"failed_executions"`
	TotalExecutionTime   time.Duration `json:"total_execution_time"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	LastExecutionTime    time.Duration `json:"last_execution_time"`
	PeakExecutionTime    time.Duration `json:"peak_execution_time"`
	SystemExecutions     int64         `json:"system_executions"`
	ParallelExecutions   int64         `json:"parallel_executions"`
	TimeoutCount         int64         `json:"timeout_count"`
}

// ExecutionError represents an error that occurred during system execution.
type ExecutionError struct {
	SystemType SystemType    `json:"system_type"`
	Error      string        `json:"error"`
	Timestamp  time.Time     `json:"timestamp"`
	Duration   time.Duration `json:"duration"`
	Context    string        `json:"context,omitempty"`
}

// ==============================================
// System Dependency Management
// ==============================================

// DependencyGraph represents the dependency relationships between systems.
type DependencyGraph struct {
	Nodes         []SystemType                `json:"nodes"`
	Edges         map[SystemType][]SystemType `json:"edges"`
	Levels        [][]SystemType              `json:"levels"`
	HasCycles     bool                        `json:"has_cycles"`
	CyclicSystems []SystemType                `json:"cyclic_systems,omitempty"`
}

// DependencyManager manages system dependencies and execution order.
type DependencyManager interface {
	// Dependency operations
	AddDependency(dependent SystemType, dependency SystemType) error
	RemoveDependency(dependent SystemType, dependency SystemType) error
	HasDependency(dependent SystemType, dependency SystemType) bool
	GetDependencies(SystemType) []SystemType
	GetDependents(SystemType) []SystemType

	// Dependency validation
	ValidateDependencies() error
	DetectCycles() ([][]SystemType, error)
	HasCycles() bool
	GetCyclicSystems() []SystemType

	// Execution order computation
	ComputeExecutionOrder() ([]SystemType, error)
	ComputeParallelGroups() ([][]SystemType, error)
	GetTopologicalSort() ([]SystemType, error)

	// Dependency analysis
	GetDependencyDepth(SystemType) int
	GetMaxDependencyDepth() int
	GetDependencyChain(SystemType) []SystemType
	GetCriticalPath() []SystemType

	// Graph operations
	GetGraph() *DependencyGraph
	GetGraphMetrics() *DependencyGraphMetrics
	ExportGraph(format string) ([]byte, error)
	ImportGraph([]byte) error

	// Optimization
	OptimizeDependencies() error
	SuggestOptimizations() []DependencyOptimization
}

// DependencyGraphMetrics contains statistics about the dependency graph.
type DependencyGraphMetrics struct {
	NodeCount          int     `json:"node_count"`
	EdgeCount          int     `json:"edge_count"`
	MaxDepth           int     `json:"max_depth"`
	AverageDepth       float64 `json:"average_depth"`
	CyclicNodes        int     `json:"cyclic_nodes"`
	IsolatedNodes      int     `json:"isolated_nodes"`
	CriticalPathLength int     `json:"critical_path_length"`
	ParallelismFactor  float64 `json:"parallelism_factor"`
}

// DependencyOptimization represents a suggested optimization for dependencies.
type DependencyOptimization struct {
	Type        OptimizationType   `json:"type"`
	Description string             `json:"description"`
	Systems     []SystemType       `json:"systems"`
	Impact      OptimizationImpact `json:"impact"`
	Effort      OptimizationEffort `json:"effort"`
}

// OptimizationType represents different types of dependency optimizations.
type OptimizationType int

const (
	OptimizationRemoveRedundant OptimizationType = iota
	OptimizationMergeSequential
	OptimizationParallelizeIndependent
	OptimizationReorderForCache
	OptimizationBreakCycle
)

// OptimizationImpact represents the expected impact of an optimization.
type OptimizationImpact int

const (
	ImpactLow OptimizationImpact = iota
	ImpactMedium
	ImpactHigh
	ImpactCritical
)

// OptimizationEffort represents the effort required to implement an optimization.
type OptimizationEffort int

const (
	EffortLow OptimizationEffort = iota
	EffortMedium
	EffortHigh
	EffortCritical
)

// ==============================================
// System Scheduling and Execution
// ==============================================

// SystemScheduler manages when and how systems are executed.
type SystemScheduler interface {
	// Scheduling
	ScheduleSystem(SystemType, ScheduleMode) error
	UnscheduleSystem(SystemType) error
	IsScheduled(SystemType) bool
	GetScheduledSystems() []SystemType

	// Execution timing
	SetSystemFrequency(SystemType, float64) error
	GetSystemFrequency(SystemType) float64
	SetSystemPriority(SystemType, Priority) error
	GetSystemPriority(SystemType) Priority

	// Frame-based scheduling
	ScheduleEveryFrame(SystemType) error
	ScheduleEveryNFrames(SystemType, int) error
	ScheduleAtFPS(SystemType, float64) error

	// Time-based scheduling
	ScheduleAtInterval(SystemType, time.Duration) error
	ScheduleOnce(SystemType, time.Time) error
	ScheduleAfterDelay(SystemType, time.Duration) error

	// Conditional scheduling
	ScheduleWhen(SystemType, func() bool) error
	ScheduleOnEvent(SystemType, EventType) error

	// Schedule management
	PauseSchedule()
	ResumeSchedule()
	ClearSchedule()
	GetScheduleState() ScheduleState

	// Statistics
	GetSchedulingStats() *SchedulingStats
}

// ScheduleMode defines how a system should be scheduled.
type ScheduleMode int

const (
	ScheduleModeEveryFrame ScheduleMode = iota
	ScheduleModeInterval
	ScheduleModeConditional
	ScheduleModeEvent
	ScheduleModeOnce
)

// ScheduleState represents the current state of the scheduler.
type ScheduleState int

const (
	ScheduleStateRunning ScheduleState = iota
	ScheduleStatePaused
	ScheduleStateStopped
)

// SchedulingStats contains statistics about system scheduling.
type SchedulingStats struct {
	ScheduledSystems    int                          `json:"scheduled_systems"`
	ActiveSystems       int                          `json:"active_systems"`
	PausedSystems       int                          `json:"paused_systems"`
	TotalExecutions     int64                        `json:"total_executions"`
	MissedExecutions    int64                        `json:"missed_executions"`
	AverageFrameTime    time.Duration                `json:"average_frame_time"`
	SystemExecutionTime map[SystemType]time.Duration `json:"system_execution_time"`
	LastScheduleUpdate  time.Time                    `json:"last_schedule_update"`
}

// ==============================================
// System Performance Monitoring
// ==============================================

// SystemProfiler provides detailed performance profiling for systems.
type SystemProfiler interface {
	// Profiling control
	StartProfiling(SystemType) error
	StopProfiling(SystemType) error
	IsProfilingEnabled(SystemType) bool
	EnableGlobalProfiling(bool)

	// Performance data collection
	CollectMetrics(SystemType, time.Duration) error
	GetPerformanceProfile(SystemType) *PerformanceProfile
	GetProfilingData(SystemType, time.Time, time.Time) *ProfilingData

	// Performance analysis
	AnalyzePerformance(SystemType) *PerformanceAnalysis
	GetBottlenecks() []SystemPerformanceBottleneck
	GetPerformanceRecommendations(SystemType) []SystemPerformanceRecommendation

	// Memory profiling
	ProfileMemoryUsage(SystemType) *SystemMemoryProfile
	GetMemoryLeaks() []SystemMemoryLeak
	AnalyzeMemoryPattern(SystemType) *SystemMemoryAnalysis

	// Report generation
	GeneratePerformanceReport() ([]byte, error)
	ExportProfilingData(format string) ([]byte, error)
}

// PerformanceProfile contains detailed performance data for a system.
type PerformanceProfile struct {
	SystemType        SystemType    `json:"system_type"`
	SampleCount       int64         `json:"sample_count"`
	TotalTime         time.Duration `json:"total_time"`
	MinTime           time.Duration `json:"min_time"`
	MaxTime           time.Duration `json:"max_time"`
	AverageTime       time.Duration `json:"average_time"`
	MedianTime        time.Duration `json:"median_time"`
	Percentile95Time  time.Duration `json:"percentile_95_time"`
	Percentile99Time  time.Duration `json:"percentile_99_time"`
	StandardDeviation float64       `json:"standard_deviation"`
	CPUUsage          float64       `json:"cpu_usage_percent"`
	MemoryUsage       int64         `json:"memory_usage_bytes"`
	CacheHitRate      float64       `json:"cache_hit_rate"`
}

// ProfilingData contains raw profiling measurements.
type ProfilingData struct {
	SystemType   SystemType      `json:"system_type"`
	StartTime    time.Time       `json:"start_time"`
	EndTime      time.Time       `json:"end_time"`
	Samples      []ProfileSample `json:"samples"`
	TotalSamples int64           `json:"total_samples"`
}

// ProfileSample represents a single performance measurement.
type ProfileSample struct {
	Timestamp         time.Time     `json:"timestamp"`
	ExecutionTime     time.Duration `json:"execution_time"`
	MemoryUsage       int64         `json:"memory_usage_bytes"`
	CPUUsage          float64       `json:"cpu_usage_percent"`
	EntitiesProcessed int           `json:"entities_processed"`
	Context           string        `json:"context,omitempty"`
}

// PerformanceAnalysis contains analysis results for system performance.
type PerformanceAnalysis struct {
	SystemType              SystemType               `json:"system_type"`
	OverallRating           PerformanceRating        `json:"overall_rating"`
	PerformanceIssues       []PerformanceIssue       `json:"performance_issues"`
	OptimizationSuggestions []OptimizationSuggestion `json:"optimization_suggestions"`
	TrendAnalysis           *PerformanceTrend        `json:"trend_analysis"`
	ComparisonToBenchmark   *BenchmarkComparison     `json:"benchmark_comparison"`
}

// PerformanceRating represents the overall performance rating of a system.
type PerformanceRating int

const (
	PerformanceExcellent PerformanceRating = iota
	PerformanceGood
	PerformanceFair
	PerformancePoor
	PerformanceCritical
)

// PerformanceIssue represents a detected performance problem.
type PerformanceIssue struct {
	Type        PerformanceIssueType     `json:"type"`
	Severity    PerformanceIssueSeverity `json:"severity"`
	Description string                   `json:"description"`
	Impact      string                   `json:"impact"`
	Suggestion  string                   `json:"suggestion"`
	Metric      string                   `json:"metric"`
	Value       float64                  `json:"value"`
	Threshold   float64                  `json:"threshold"`
}

// PerformanceIssueType represents different types of performance issues.
type PerformanceIssueType int

const (
	IssueHighExecutionTime PerformanceIssueType = iota
	IssueMemoryLeak
	IssueCacheMiss
	IssueExcessiveAllocations
	IssueIneffientAlgorithm
	IssueBottleneck
)

// PerformanceIssueSeverity represents the severity of a performance issue.
type PerformanceIssueSeverity int

const (
	PerfSeverityInfo PerformanceIssueSeverity = iota
	PerfSeverityMinor
	PerfSeverityMajor
	PerfSeverityCritical
)

// OptimizationSuggestion represents a suggestion for improving system performance.
type OptimizationSuggestion struct {
	Title                string               `json:"title"`
	Description          string               `json:"description"`
	ExpectedImpact       OptimizationImpact   `json:"expected_impact"`
	ImplementationEffort OptimizationEffort   `json:"implementation_effort"`
	Category             OptimizationCategory `json:"category"`
	Priority             int                  `json:"priority"`
	CodeExample          string               `json:"code_example,omitempty"`
}

// OptimizationCategory represents different categories of optimizations.
type OptimizationCategory int

const (
	CategoryAlgorithmic OptimizationCategory = iota
	CategoryMemory
	CategoryCaching
	CategoryParallelization
	CategoryDataStructure
	CategoryIO
)

// PerformanceTrend contains trend analysis for system performance over time.
type PerformanceTrend struct {
	Direction         TrendDirection `json:"direction"`
	Strength          TrendStrength  `json:"strength"`
	DurationAnalyzed  time.Duration  `json:"duration_analyzed"`
	AverageChange     float64        `json:"average_change_percent"`
	PeakPerformance   time.Time      `json:"peak_performance"`
	WorstPerformance  time.Time      `json:"worst_performance"`
	PredictedNextWeek float64        `json:"predicted_next_week"`
}

// TrendDirection represents the direction of a performance trend.
type TrendDirection int

const (
	TrendImproving TrendDirection = iota
	TrendStable
	TrendDegrading
	TrendUnknown
)

// TrendStrength represents how strong a performance trend is.
type TrendStrength int

const (
	TrendWeak TrendStrength = iota
	TrendModerate
	TrendStrong
	TrendVeryStrong
)

// BenchmarkComparison contains comparison results against performance benchmarks.
type BenchmarkComparison struct {
	BenchmarkName    string          `json:"benchmark_name"`
	BenchmarkVersion string          `json:"benchmark_version"`
	ComparisonTime   time.Time       `json:"comparison_time"`
	PerformanceRatio float64         `json:"performance_ratio"` // Current/Benchmark
	Status           BenchmarkStatus `json:"status"`
	Details          string          `json:"details"`
}

// BenchmarkStatus represents how current performance compares to benchmarks.
type BenchmarkStatus int

const (
	BenchmarkExceeds BenchmarkStatus = iota
	BenchmarkMeets
	BenchmarkNear
	BenchmarkBelow
	BenchmarkFails
)

// ==============================================
// System Manager Debug Information
// ==============================================

// SystemManagerDebugInfo provides comprehensive debugging information.
type SystemManagerDebugInfo struct {
	RegisteredSystems  int                           `json:"registered_systems"`
	EnabledSystems     int                           `json:"enabled_systems"`
	DisabledSystems    int                           `json:"disabled_systems"`
	DependencyCount    int                           `json:"dependency_count"`
	ExecutionOrder     []SystemType                  `json:"execution_order"`
	ParallelGroups     [][]SystemType                `json:"parallel_groups"`
	SystemMetrics      map[SystemType]*SystemMetrics `json:"system_metrics"`
	DependencyGraph    *DependencyGraph              `json:"dependency_graph"`
	FailedSystems      []SystemType                  `json:"failed_systems"`
	SystemErrors       map[SystemType][]string       `json:"system_errors"`
	ParallelExecution  bool                          `json:"parallel_execution_enabled"`
	MaxParallelSystems int                           `json:"max_parallel_systems"`
	GlobalTimeout      time.Duration                 `json:"global_timeout"`
	ProfilingEnabled   bool                          `json:"profiling_enabled"`
	LastExecutionTime  time.Time                     `json:"last_execution_time"`
	TotalExecutions    int64                         `json:"total_executions"`
}

// ==============================================
// System Worker Pool
// ==============================================

// WorkerPool manages a pool of workers for parallel system execution.
type WorkerPool interface {
	// Pool management
	Start() error
	Stop() error
	Resize(int) error
	GetSize() int
	GetActiveWorkers() int

	// Task execution
	SubmitTask(Task) error
	SubmitTaskWithTimeout(Task, time.Duration) error
	SubmitTaskAsync(Task) <-chan TaskResult

	// Batch execution
	SubmitTasks([]Task) error
	WaitForCompletion() error
	WaitForCompletionWithTimeout(time.Duration) error

	// Pool statistics
	GetStats() *WorkerPoolStats
	GetWorkerStats() []WorkerStats
}

// Task represents a unit of work for the worker pool.
type Task interface {
	Execute(context.Context) error
	GetID() string
	GetPriority() Priority
	GetTimeout() time.Duration
}

// TaskResult contains the result of task execution.
type TaskResult struct {
	TaskID      string        `json:"task_id"`
	Success     bool          `json:"success"`
	Error       error         `json:"error,omitempty"`
	Duration    time.Duration `json:"duration"`
	CompletedAt time.Time     `json:"completed_at"`
}

// WorkerPoolStats contains statistics about the worker pool.
type WorkerPoolStats struct {
	PoolSize        int           `json:"pool_size"`
	ActiveWorkers   int           `json:"active_workers"`
	IdleWorkers     int           `json:"idle_workers"`
	QueuedTasks     int           `json:"queued_tasks"`
	CompletedTasks  int64         `json:"completed_tasks"`
	FailedTasks     int64         `json:"failed_tasks"`
	AverageTaskTime time.Duration `json:"average_task_time"`
	TotalTaskTime   time.Duration `json:"total_task_time"`
	Throughput      float64       `json:"tasks_per_second"`
}

// WorkerStats contains statistics about an individual worker.
type WorkerStats struct {
	WorkerID        int           `json:"worker_id"`
	IsActive        bool          `json:"is_active"`
	TasksCompleted  int64         `json:"tasks_completed"`
	TasksFailed     int64         `json:"tasks_failed"`
	TotalWorkTime   time.Duration `json:"total_work_time"`
	AverageTaskTime time.Duration `json:"average_task_time"`
	LastTaskTime    time.Time     `json:"last_task_time"`
}

// ==============================================
// Thread Safety and Concurrency
// ==============================================

// SystemConcurrencyManager manages concurrent system execution safely.
type SystemConcurrencyManager interface {
	// Concurrency control
	AcquireSystemLock(SystemType) error
	ReleaseSystemLock(SystemType) error
	TryAcquireSystemLock(SystemType) bool
	IsSystemLocked(SystemType) bool

	// Resource management
	AcquireResource(string) error
	ReleaseResource(string) error
	IsResourceAvailable(string) bool
	GetResourceOwner(string) SystemType

	// Deadlock prevention
	SetLockTimeout(time.Duration)
	GetLockTimeout() time.Duration
	DetectDeadlocks() []DeadlockInfo
	ResolveDeadlock(DeadlockInfo) error

	// Synchronization primitives
	CreateBarrier(string, int) error
	WaitAtBarrier(string) error
	ReleaseBarrier(string) error

	// Thread safety validation
	ValidateThreadSafety() error
	GetConcurrencyViolations() []ConcurrencyViolation
}

// DeadlockInfo contains information about a detected deadlock.
type DeadlockInfo struct {
	Systems    []SystemType     `json:"systems"`
	Resources  []string         `json:"resources"`
	DetectedAt time.Time        `json:"detected_at"`
	Severity   DeadlockSeverity `json:"severity"`
}

// DeadlockSeverity represents the severity of a deadlock.
type DeadlockSeverity int

const (
	DeadlockMild DeadlockSeverity = iota
	DeadlockModerate
	DeadlockSevere
	DeadlockCritical
)

// ConcurrencyViolation represents a thread safety violation.
type ConcurrencyViolation struct {
	SystemType    SystemType    `json:"system_type"`
	ViolationType ViolationType `json:"violation_type"`
	Description   string        `json:"description"`
	DetectedAt    time.Time     `json:"detected_at"`
	StackTrace    string        `json:"stack_trace,omitempty"`
}

// ViolationType represents different types of concurrency violations.
type ViolationType int

const (
	ViolationRaceCondition ViolationType = iota
	ViolationDataRace
	ViolationDeadlock
	ViolationLivelockk
	ViolationResourceLeak
)

// ==============================================
// Additional System Performance Types
// ==============================================

// SystemPerformanceBottleneck represents a performance bottleneck in a system.
type SystemPerformanceBottleneck struct {
	SystemType  SystemType               `json:"system_type"`
	Type        string                   `json:"type"`
	Description string                   `json:"description"`
	Impact      float64                  `json:"impact_percent"`
	Severity    PerformanceIssueSeverity `json:"severity"`
}

// SystemPerformanceRecommendation represents a performance improvement recommendation.
type SystemPerformanceRecommendation struct {
	SystemType   SystemType `json:"system_type"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	ExpectedGain float64    `json:"expected_gain_percent"`
	Difficulty   int        `json:"difficulty_level"`
	CodeExample  string     `json:"code_example,omitempty"`
}

// SystemMemoryProfile contains memory usage profile for a system.
type SystemMemoryProfile struct {
	SystemType      SystemType `json:"system_type"`
	AllocatedMemory int64      `json:"allocated_memory_bytes"`
	PeakMemory      int64      `json:"peak_memory_bytes"`
	Allocations     int64      `json:"total_allocations"`
	Deallocations   int64      `json:"total_deallocations"`
	ActiveObjects   int64      `json:"active_objects"`
}

// SystemMemoryLeak represents a detected memory leak in a system.
type SystemMemoryLeak struct {
	SystemType  SystemType `json:"system_type"`
	LeakRate    int64      `json:"leak_rate_bytes_per_second"`
	DetectedAt  time.Time  `json:"detected_at"`
	Description string     `json:"description"`
	StackTrace  string     `json:"stack_trace,omitempty"`
}

// SystemMemoryAnalysis contains memory usage analysis for a system.
type SystemMemoryAnalysis struct {
	SystemType      SystemType `json:"system_type"`
	MemoryPattern   string     `json:"memory_pattern"`
	OptimalMemory   int64      `json:"optimal_memory_bytes"`
	WastePercentage float64    `json:"waste_percentage"`
	Recommendations []string   `json:"recommendations"`
}
