// Package lintdesign contains the interface definitions and data structures
// for the lint compliance system design.
// This file serves as a design specification and should not be compiled directly.
package lintdesign

import (
	"context"
	"io"
	"regexp"
	"time"
)

// =====================================================================================
// Core Lint Engine Interfaces
// =====================================================================================

// LintEngine is the main interface for the lint processing system
type LintEngine interface {
	// Initialize sets up the lint engine with configuration
	Initialize(ctx context.Context, config *LintConfig) error
	
	// RunLint executes lint analysis on specified files
	RunLint(ctx context.Context, files []string, opts *LintOptions) (*LintResult, error)
	
	// RunIncremental performs incremental lint analysis on changed files only
	RunIncremental(ctx context.Context, changedFiles []string, opts *LintOptions) (*LintResult, error)
	
	// GetSupportedRules returns list of all supported lint rules
	GetSupportedRules() []LintRule
	
	// Shutdown gracefully shuts down the lint engine
	Shutdown(ctx context.Context) error
}

// LintRunner defines interface for individual linter execution
type LintRunner interface {
	// Name returns the linter name (e.g., "golangci-lint", "gosec")
	Name() string
	
	// Version returns the linter version
	Version() string
	
	// Run executes the linter on specified files
	Run(ctx context.Context, files []string) (*LintRunnerResult, error)
	
	// SupportsIncrementalMode indicates if incremental analysis is supported
	SupportsIncrementalMode() bool
	
	// GetDefaultConfig returns default configuration for this runner
	GetDefaultConfig() map[string]interface{}
}

// =====================================================================================
// Configuration Management Interfaces
// =====================================================================================

// ConfigManager handles lint configuration management
type ConfigManager interface {
	// LoadConfig loads configuration from file or default
	LoadConfig(configPath string) (*LintConfig, error)
	
	// ValidateConfig validates configuration correctness
	ValidateConfig(config *LintConfig) error
	
	// GetEnvironmentConfig returns environment-specific config overrides
	GetEnvironmentConfig(env Environment) (*ConfigOverride, error)
	
	// GetECSOptimizations returns ECS-specific optimizations
	GetECSOptimizations() (*ECSConfig, error)
	
	// SaveConfig saves configuration to file
	SaveConfig(config *LintConfig, path string) error
}

// ConfigValidator validates configuration files and rules
type ConfigValidator interface {
	// ValidateYAML validates YAML configuration syntax
	ValidateYAML(data []byte) error
	
	// ValidateRules validates lint rule definitions
	ValidateRules(rules []LintRule) error
	
	// ValidateExclusions validates exclusion patterns
	ValidateExclusions(exclusions []ExclusionRule) error
}

// =====================================================================================
// Quality Gate Interfaces
// =====================================================================================

// QualityGate controls quality standards and decisions
type QualityGate interface {
	// Evaluate assesses lint results against quality criteria
	Evaluate(ctx context.Context, result *LintResult) (*QualityAssessment, error)
	
	// GetRules returns all quality rules
	GetRules() []QualityRule
	
	// SetThreshold sets quality threshold for specific metric
	SetThreshold(metric string, threshold interface{}) error
	
	// IsBlocking determines if quality gate should block pipeline
	IsBlocking(assessment *QualityAssessment) bool
}

// QualityRule defines a specific quality evaluation rule
type QualityRule interface {
	// Evaluate evaluates a single rule against lint results
	Evaluate(result *LintResult) (bool, error)
	
	// Severity returns rule severity level
	Severity() Severity
	
	// Message returns human-readable rule description
	Message() string
	
	// Category returns rule category (security, performance, etc.)
	Category() RuleCategory
}

// =====================================================================================
// Performance Monitoring Interfaces
// =====================================================================================

// PerformanceMonitor tracks lint execution performance
type PerformanceMonitor interface {
	// StartMonitoring begins performance tracking
	StartMonitoring(ctx context.Context) error
	
	// RecordMetric records a specific performance metric
	RecordMetric(name string, value float64, labels map[string]string)
	
	// GetCurrentStats returns current performance statistics
	GetCurrentStats() *PerformanceStats
	
	// StopMonitoring stops performance tracking and returns final stats
	StopMonitoring() (*PerformanceStats, error)
}

// MetricsCollector collects and aggregates metrics
type MetricsCollector interface {
	// Collect gathers metrics from various sources
	Collect(ctx context.Context) error
	
	// Export exports collected metrics to external systems
	Export(ctx context.Context, exporter MetricsExporter) error
	
	// GetSummary returns metrics summary
	GetSummary() *MetricsSummary
}

// MetricsExporter defines interface for exporting metrics
type MetricsExporter interface {
	// Export exports metrics data
	Export(ctx context.Context, data *MetricsData) error
	
	// SupportedFormats returns supported export formats
	SupportedFormats() []string
}

// =====================================================================================
// Error Handling and Recovery Interfaces
// =====================================================================================

// ErrorHandler manages error handling and recovery
type ErrorHandler interface {
	// HandleError processes and potentially recovers from errors
	HandleError(ctx context.Context, err error) (*ErrorRecovery, error)
	
	// IsRetryable determines if error is retryable
	IsRetryable(err error) bool
	
	// GetFallbackAction returns fallback action for error
	GetFallbackAction(err error) FallbackAction
	
	// RecordError records error for analysis
	RecordError(err error, context map[string]interface{})
}

// RetryPolicy defines retry behavior
type RetryPolicy interface {
	// ShouldRetry determines if operation should be retried
	ShouldRetry(attempt int, err error) bool
	
	// NextDelay returns delay before next retry
	NextDelay(attempt int) time.Duration
	
	// MaxAttempts returns maximum retry attempts
	MaxAttempts() int
}

// =====================================================================================
// Data Structures
// =====================================================================================

// LintConfig represents the main lint configuration
type LintConfig struct {
	// Version of config format
	Version string `yaml:"version"`
	
	// Enabled linters
	Linters *LintersConfig `yaml:"linters"`
	
	// Linter-specific settings
	LintersSettings map[string]interface{} `yaml:"linters-settings"`
	
	// Run configuration
	Run *RunConfig `yaml:"run"`
	
	// Issues configuration  
	Issues *IssuesConfig `yaml:"issues"`
	
	// Output configuration
	Output *OutputConfig `yaml:"output"`
	
	// ECS-specific settings
	ECS *ECSConfig `yaml:"ecs,omitempty"`
	
	// Environment-specific overrides
	Environments map[string]*ConfigOverride `yaml:"environments,omitempty"`
}

// LintersConfig configures enabled/disabled linters
type LintersConfig struct {
	// Disable all linters by default
	DisableAll bool `yaml:"disable-all"`
	
	// Enable specific linters
	Enable []string `yaml:"enable"`
	
	// Disable specific linters
	Disable []string `yaml:"disable"`
	
	// Fast mode (only fast linters)
	Fast bool `yaml:"fast"`
}

// RunConfig configures lint execution
type RunConfig struct {
	// Timeout for the whole lint process
	Timeout time.Duration `yaml:"timeout"`
	
	// Number of parallel workers
	Concurrency int `yaml:"concurrency"`
	
	// Go version to target
	GoVersion string `yaml:"go"`
	
	// Directories to skip
	SkipDirs []string `yaml:"skip-dirs"`
	
	// Files to skip (patterns)
	SkipFiles []string `yaml:"skip-files"`
	
	// Allow parallel execution
	AllowParallelRunners bool `yaml:"allow-parallel-runners"`
	
	// Module download mode
	ModulesDownloadMode string `yaml:"modules-download-mode"`
}

// IssuesConfig configures issue handling
type IssuesConfig struct {
	// Exclude patterns
	ExcludeRules []ExclusionRule `yaml:"exclude-rules"`
	
	// Include results that match these patterns
	IncludeFiles []string `yaml:"include-files"`
	
	// Exclude results that match these patterns  
	ExcludeFiles []string `yaml:"exclude-files"`
	
	// Maximum issues per linter
	MaxIssuesPerLinter int `yaml:"max-issues-per-linter"`
	
	// Maximum same issues
	MaxSameIssues int `yaml:"max-same-issues"`
	
	// Fix issues automatically (if supported)
	Fix bool `yaml:"fix"`
}

// OutputConfig configures lint output formatting
type OutputConfig struct {
	// Output formats
	Formats []OutputFormat `yaml:"formats"`
	
	// Print issued lines
	PrintIssuedLines bool `yaml:"print-issued-lines"`
	
	// Print linter name
	PrintLinterName bool `yaml:"print-linter-name"`
	
	// Sort results
	SortResults bool `yaml:"sort-results"`
	
	// Path prefix to strip
	PathPrefix string `yaml:"path-prefix"`
}

// ECSConfig contains ECS framework specific configuration
type ECSConfig struct {
	// Hot path exclusions for performance-critical code
	HotPathExclusions []string `yaml:"hot-path-exclusions"`
	
	// Performance rules specific to ECS
	PerformanceRules map[string]interface{} `yaml:"performance-rules"`
	
	// Entity naming patterns
	EntityPatterns []string `yaml:"entity-patterns"`
	
	// Component naming patterns
	ComponentPatterns []string `yaml:"component-patterns"`
	
	// System naming patterns  
	SystemPatterns []string `yaml:"system-patterns"`
	
	// Memory optimization rules
	MemoryOptimization *MemoryOptimizationConfig `yaml:"memory-optimization"`
	
	// Query optimization settings
	QueryOptimization *QueryOptimizationConfig `yaml:"query-optimization"`
}

// MemoryOptimizationConfig configures memory-related optimizations
type MemoryOptimizationConfig struct {
	// Maximum component size in bytes
	MaxComponentSizeBytes int `yaml:"max-component-size-bytes"`
	
	// Check memory alignment
	CheckAlignment bool `yaml:"check-alignment"`
	
	// Prefer value types over pointers
	PreferValueTypes bool `yaml:"prefer-value-types"`
	
	// Pool usage patterns
	PoolUsagePatterns []string `yaml:"pool-usage-patterns"`
}

// QueryOptimizationConfig configures query-related optimizations
type QueryOptimizationConfig struct {
	// Maximum entities per query per frame
	MaxEntitiesPerFrame int `yaml:"max-entities-per-frame"`
	
	// Query complexity limits
	MaxQueryComplexity int `yaml:"max-query-complexity"`
	
	// Cache query results
	CacheResults bool `yaml:"cache-results"`
	
	// Batch processing patterns
	BatchProcessingPatterns []string `yaml:"batch-processing-patterns"`
}

// ConfigOverride provides environment-specific configuration overrides
type ConfigOverride struct {
	// Environment name
	Environment string `yaml:"environment"`
	
	// Override timeout
	Timeout *time.Duration `yaml:"timeout,omitempty"`
	
	// Override concurrency
	Concurrency *int `yaml:"concurrency,omitempty"`
	
	// Additional enabled linters
	AdditionalLinters []string `yaml:"additional-linters,omitempty"`
	
	// Disabled linters for this environment
	DisabledLinters []string `yaml:"disabled-linters,omitempty"`
	
	// Environment-specific exclusions
	ExclusionRules []ExclusionRule `yaml:"exclusion-rules,omitempty"`
}

// LintOptions provides runtime options for lint execution
type LintOptions struct {
	// Working directory
	WorkingDir string
	
	// Configuration file path
	ConfigFile string
	
	// Environment (development, ci, production)
	Environment Environment
	
	// Incremental mode
	Incremental bool
	
	// Fix issues automatically
	Fix bool
	
	// Format code
	Format bool
	
	// Organize imports
	OrganizeImports bool
	
	// Output writer
	Output io.Writer
	
	// Progress callback
	ProgressCallback func(progress *Progress)
	
	// Additional context
	Context map[string]interface{}
}

// LintResult contains the results of lint analysis
type LintResult struct {
	// Execution summary
	Summary *ExecutionSummary `json:"summary"`
	
	// Issues found
	Issues []LintIssue `json:"issues"`
	
	// Performance statistics  
	Performance *PerformanceStats `json:"performance"`
	
	// Metrics collected
	Metrics *LintMetrics `json:"metrics"`
	
	// Warnings (non-blocking issues)
	Warnings []LintWarning `json:"warnings"`
	
	// Files processed
	FilesProcessed []string `json:"files_processed"`
	
	// Configuration used
	ConfigUsed *LintConfig `json:"config_used,omitempty"`
}

// ExecutionSummary summarizes lint execution
type ExecutionSummary struct {
	// Start time
	StartTime time.Time `json:"start_time"`
	
	// End time
	EndTime time.Time `json:"end_time"`
	
	// Total duration
	Duration time.Duration `json:"duration"`
	
	// Success status
	Success bool `json:"success"`
	
	// Error message (if failed)
	Error string `json:"error,omitempty"`
	
	// Total files analyzed
	TotalFiles int `json:"total_files"`
	
	// Total lines analyzed
	TotalLines int `json:"total_lines"`
	
	// Issues count by severity
	IssuesBySeverity map[Severity]int `json:"issues_by_severity"`
	
	// Linters executed
	LintersExecuted []string `json:"linters_executed"`
}

// LintIssue represents a single lint issue
type LintIssue struct {
	// Unique issue identifier
	ID string `json:"id"`
	
	// Severity level
	Severity Severity `json:"severity"`
	
	// Rule that generated this issue
	Rule string `json:"rule"`
	
	// Linter name
	Linter string `json:"linter"`
	
	// Category of the issue
	Category RuleCategory `json:"category"`
	
	// File path
	FilePath string `json:"file_path"`
	
	// Line number
	Line int `json:"line"`
	
	// Column number
	Column int `json:"column"`
	
	// End line (for multi-line issues)
	EndLine int `json:"end_line,omitempty"`
	
	// End column
	EndColumn int `json:"end_column,omitempty"`
	
	// Issue message
	Message string `json:"message"`
	
	// Code snippet
	Snippet string `json:"snippet,omitempty"`
	
	// Suggested fix
	SuggestedFix *SuggestedFix `json:"suggested_fix,omitempty"`
	
	// Additional context
	Context map[string]interface{} `json:"context,omitempty"`
}

// SuggestedFix contains automated fix suggestions
type SuggestedFix struct {
	// Fix description
	Description string `json:"description"`
	
	// Original code
	Original string `json:"original"`
	
	// Replacement code
	Replacement string `json:"replacement"`
	
	// Confidence level (0.0 to 1.0)
	Confidence float64 `json:"confidence"`
	
	// Whether fix can be applied automatically
	AutoApplicable bool `json:"auto_applicable"`
}

// LintWarning represents non-critical warnings
type LintWarning struct {
	// Warning type
	Type string `json:"type"`
	
	// Warning message
	Message string `json:"message"`
	
	// File path (optional)
	FilePath string `json:"file_path,omitempty"`
	
	// Additional context
	Context map[string]interface{} `json:"context,omitempty"`
}

// LintMetrics contains performance and quality metrics
type LintMetrics struct {
	// Execution time metrics
	ExecutionTime *TimeMetrics `json:"execution_time"`
	
	// Memory usage metrics
	MemoryUsage *MemoryMetrics `json:"memory_usage"`
	
	// Issue count metrics
	IssueCounts map[string]int `json:"issue_counts"`
	
	// Linter performance metrics
	LinterPerformance map[string]*LinterMetrics `json:"linter_performance"`
	
	// Code quality metrics
	QualityMetrics *QualityMetrics `json:"quality_metrics"`
}

// TimeMetrics contains timing-related metrics
type TimeMetrics struct {
	// Total execution time
	Total time.Duration `json:"total"`
	
	// Setup time
	Setup time.Duration `json:"setup"`
	
	// Analysis time
	Analysis time.Duration `json:"analysis"`
	
	// Formatting time
	Formatting time.Duration `json:"formatting"`
	
	// Cleanup time
	Cleanup time.Duration `json:"cleanup"`
}

// MemoryMetrics contains memory usage metrics
type MemoryMetrics struct {
	// Peak memory usage in bytes
	Peak int64 `json:"peak"`
	
	// Average memory usage in bytes
	Average int64 `json:"average"`
	
	// Memory allocations
	Allocations int64 `json:"allocations"`
	
	// Garbage collections
	GCCount int `json:"gc_count"`
}

// LinterMetrics contains per-linter performance metrics
type LinterMetrics struct {
	// Execution time
	Duration time.Duration `json:"duration"`
	
	// Memory usage
	MemoryUsage int64 `json:"memory_usage"`
	
	// Issues found
	IssuesFound int `json:"issues_found"`
	
	// Files processed
	FilesProcessed int `json:"files_processed"`
	
	// Success status
	Success bool `json:"success"`
	
	// Error message
	Error string `json:"error,omitempty"`
}

// QualityMetrics contains code quality measurements
type QualityMetrics struct {
	// Cyclomatic complexity statistics
	CyclomaticComplexity *ComplexityStats `json:"cyclomatic_complexity"`
	
	// Code duplication statistics
	Duplication *DuplicationStats `json:"duplication"`
	
	// Test coverage information
	TestCoverage *CoverageStats `json:"test_coverage,omitempty"`
	
	// Technical debt estimation
	TechnicalDebt *TechnicalDebtStats `json:"technical_debt"`
}

// ComplexityStats contains complexity analysis results
type ComplexityStats struct {
	// Average complexity
	Average float64 `json:"average"`
	
	// Maximum complexity
	Maximum int `json:"maximum"`
	
	// Functions above threshold
	AboveThreshold int `json:"above_threshold"`
	
	// Complexity distribution
	Distribution map[string]int `json:"distribution"`
}

// DuplicationStats contains code duplication analysis
type DuplicationStats struct {
	// Percentage of duplicated lines
	Percentage float64 `json:"percentage"`
	
	// Number of duplicated blocks
	Blocks int `json:"blocks"`
	
	// Lines of duplicated code
	Lines int `json:"lines"`
	
	// Files with duplication
	FilesAffected int `json:"files_affected"`
}

// CoverageStats contains test coverage information
type CoverageStats struct {
	// Line coverage percentage
	LinesCovered float64 `json:"lines_covered"`
	
	// Function coverage percentage
	FunctionsCovered float64 `json:"functions_covered"`
	
	// Branch coverage percentage
	BranchesCovered float64 `json:"branches_covered"`
	
	// Uncovered lines
	UncoveredLines []FileLine `json:"uncovered_lines"`
}

// TechnicalDebtStats estimates technical debt
type TechnicalDebtStats struct {
	// Estimated hours to fix all issues
	EstimatedHours float64 `json:"estimated_hours"`
	
	// Debt ratio (debt / total code size)
	DebtRatio float64 `json:"debt_ratio"`
	
	// Issues by category
	IssuesByCategory map[RuleCategory]int `json:"issues_by_category"`
	
	// Highest priority fixes
	HighPriorityFixes []string `json:"high_priority_fixes"`
}

// PerformanceStats contains detailed performance statistics
type PerformanceStats struct {
	// CPU usage statistics
	CPUUsage *CPUStats `json:"cpu_usage"`
	
	// Memory statistics
	Memory *MemoryMetrics `json:"memory"`
	
	// I/O statistics
	IOStats *IOStats `json:"io_stats"`
	
	// Timing breakdown
	Timing *TimeMetrics `json:"timing"`
	
	// Parallel processing statistics
	Concurrency *ConcurrencyStats `json:"concurrency"`
}

// CPUStats contains CPU usage information
type CPUStats struct {
	// Average CPU usage percentage
	Average float64 `json:"average"`
	
	// Peak CPU usage percentage
	Peak float64 `json:"peak"`
	
	// CPU time spent in user mode
	UserTime time.Duration `json:"user_time"`
	
	// CPU time spent in system mode
	SystemTime time.Duration `json:"system_time"`
}

// IOStats contains I/O performance information
type IOStats struct {
	// Files read
	FilesRead int64 `json:"files_read"`
	
	// Bytes read
	BytesRead int64 `json:"bytes_read"`
	
	// Files written
	FilesWritten int64 `json:"files_written"`
	
	// Bytes written
	BytesWritten int64 `json:"bytes_written"`
	
	// I/O operations per second
	IOPS float64 `json:"iops"`
}

// ConcurrencyStats contains parallel processing statistics
type ConcurrencyStats struct {
	// Number of workers used
	Workers int `json:"workers"`
	
	// Average queue length
	AverageQueueLength float64 `json:"average_queue_length"`
	
	// Worker utilization percentage
	WorkerUtilization float64 `json:"worker_utilization"`
	
	// Synchronization overhead
	SyncOverhead time.Duration `json:"sync_overhead"`
}

// QualityAssessment contains quality gate evaluation results
type QualityAssessment struct {
	// Overall assessment result
	Passed bool `json:"passed"`
	
	// Assessment score (0.0 to 1.0)
	Score float64 `json:"score"`
	
	// Rule evaluations
	RuleEvaluations []RuleEvaluation `json:"rule_evaluations"`
	
	// Blocking issues
	BlockingIssues []LintIssue `json:"blocking_issues"`
	
	// Recommendations
	Recommendations []string `json:"recommendations"`
	
	// Assessment timestamp
	Timestamp time.Time `json:"timestamp"`
}

// RuleEvaluation contains individual rule evaluation result
type RuleEvaluation struct {
	// Rule name
	Rule string `json:"rule"`
	
	// Evaluation passed
	Passed bool `json:"passed"`
	
	// Actual value
	ActualValue interface{} `json:"actual_value"`
	
	// Expected threshold
	Threshold interface{} `json:"threshold"`
	
	// Rule weight in overall assessment
	Weight float64 `json:"weight"`
	
	// Rule message
	Message string `json:"message"`
}

// =====================================================================================
// Utility Types and Interfaces
// =====================================================================================

// LintRule represents a lint rule definition
type LintRule struct {
	// Rule identifier
	ID string `json:"id"`
	
	// Rule name
	Name string `json:"name"`
	
	// Rule description
	Description string `json:"description"`
	
	// Rule category
	Category RuleCategory `json:"category"`
	
	// Rule severity
	Severity Severity `json:"severity"`
	
	// Rule pattern (regex)
	Pattern *regexp.Regexp `json:"-"`
	
	// Rule configuration
	Config map[string]interface{} `json:"config"`
	
	// Rule tags
	Tags []string `json:"tags"`
}

// ExclusionRule defines patterns to exclude from lint analysis
type ExclusionRule struct {
	// Path pattern to exclude
	Path string `yaml:"path"`
	
	// Linter to exclude
	Linters []string `yaml:"linters,omitempty"`
	
	// Text pattern to exclude
	Text string `yaml:"text,omitempty"`
	
	// File patterns to exclude
	Source string `yaml:"source,omitempty"`
}

// OutputFormat defines output format configuration
type OutputFormat struct {
	// Format name (json, checkstyle, junit, etc.)
	Format string `yaml:"format"`
	
	// Output path
	Path string `yaml:"path,omitempty"`
}

// Progress contains progress information
type Progress struct {
	// Current step
	CurrentStep string `json:"current_step"`
	
	// Total steps
	TotalSteps int `json:"total_steps"`
	
	// Completed steps
	CompletedSteps int `json:"completed_steps"`
	
	// Progress percentage
	Percentage float64 `json:"percentage"`
	
	// Current file being processed
	CurrentFile string `json:"current_file,omitempty"`
	
	// Files processed so far
	FilesProcessed int `json:"files_processed"`
	
	// Total files to process
	TotalFiles int `json:"total_files"`
}

// FileLine represents a file location
type FileLine struct {
	// File path
	File string `json:"file"`
	
	// Line number
	Line int `json:"line"`
	
	// Column number (optional)
	Column int `json:"column,omitempty"`
}

// LintRunnerResult contains results from individual linter execution
type LintRunnerResult struct {
	// Runner name
	Runner string `json:"runner"`
	
	// Execution success
	Success bool `json:"success"`
	
	// Issues found
	Issues []LintIssue `json:"issues"`
	
	// Execution time
	Duration time.Duration `json:"duration"`
	
	// Memory usage
	MemoryUsage int64 `json:"memory_usage"`
	
	// Error message (if failed)  
	Error string `json:"error,omitempty"`
}

// ErrorRecovery contains error recovery information
type ErrorRecovery struct {
	// Recovery successful
	Recovered bool `json:"recovered"`
	
	// Recovery action taken
	Action FallbackAction `json:"action"`
	
	// Recovery message
	Message string `json:"message"`
	
	// Additional context
	Context map[string]interface{} `json:"context"`
}

// MetricsData contains metrics for export
type MetricsData struct {
	// Timestamp
	Timestamp time.Time `json:"timestamp"`
	
	// Metrics by name
	Metrics map[string]interface{} `json:"metrics"`
	
	// Labels
	Labels map[string]string `json:"labels"`
	
	// Metadata
	Metadata map[string]interface{} `json:"metadata"`
}

// MetricsSummary contains aggregated metrics
type MetricsSummary struct {
	// Collection period
	Period time.Duration `json:"period"`
	
	// Total lint runs
	TotalRuns int `json:"total_runs"`
	
	// Successful runs
	SuccessfulRuns int `json:"successful_runs"`
	
	// Average execution time
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	
	// Average issues per run
	AverageIssuesPerRun float64 `json:"average_issues_per_run"`
	
	// Top issues by frequency
	TopIssues []IssueFrequency `json:"top_issues"`
}

// IssueFrequency contains issue frequency statistics
type IssueFrequency struct {
	// Rule name
	Rule string `json:"rule"`
	
	// Frequency count
	Count int `json:"count"`
	
	// Percentage of total issues
	Percentage float64 `json:"percentage"`
}

// =====================================================================================
// Enums and Constants
// =====================================================================================

// Severity represents issue severity levels
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
	SeverityHint    Severity = "hint"
)

// RuleCategory represents rule categories
type RuleCategory string

const (
	CategorySecurity      RuleCategory = "security"
	CategoryPerformance   RuleCategory = "performance" 
	CategoryReliability   RuleCategory = "reliability"
	CategoryMaintainability RuleCategory = "maintainability"
	CategoryStyle         RuleCategory = "style"
	CategoryComplexity    RuleCategory = "complexity"
	CategoryDuplication   RuleCategory = "duplication"
	CategoryTesting       RuleCategory = "testing"
	CategoryDocumentation RuleCategory = "documentation"
)

// Environment represents execution environment
type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentCI          Environment = "ci"
	EnvironmentProduction  Environment = "production"
	EnvironmentTesting     Environment = "testing"
)

// FallbackAction represents recovery actions
type FallbackAction string

const (
	FallbackNone        FallbackAction = "none"
	FallbackRetry       FallbackAction = "retry"
	FallbackDegradedMode FallbackAction = "degraded_mode"
	FallbackSkip        FallbackAction = "skip"
	FallbackFail        FallbackAction = "fail"
)

// =====================================================================================
// Helper Interfaces
// =====================================================================================

// Comparable allows comparison of lint results
type Comparable interface {
	// Compare compares with another result
	Compare(other interface{}) int
}

// Serializable allows serialization of data structures
type Serializable interface {
	// Serialize converts to byte representation
	Serialize() ([]byte, error)
	
	// Deserialize loads from byte representation
	Deserialize(data []byte) error
}

// Cacheable allows caching of results
type Cacheable interface {
	// CacheKey returns cache key for this item
	CacheKey() string
	
	// CacheExpiry returns cache expiration time
	CacheExpiry() time.Duration
	
	// IsValid checks if cached item is still valid
	IsValid() bool
}

// =====================================================================================
// Factory Interfaces
// =====================================================================================

// LintEngineFactory creates lint engine instances
type LintEngineFactory interface {
	// CreateEngine creates a new lint engine
	CreateEngine(config *LintConfig) (LintEngine, error)
	
	// CreateRunner creates a specific linter runner
	CreateRunner(name string, config map[string]interface{}) (LintRunner, error)
	
	// GetSupportedRunners returns list of supported runners
	GetSupportedRunners() []string
}

// ConfigManagerFactory creates configuration manager instances
type ConfigManagerFactory interface {
	// CreateManager creates a new configuration manager
	CreateManager() ConfigManager
	
	// CreateValidator creates a configuration validator
	CreateValidator() ConfigValidator
}

// This file serves as a comprehensive interface definition for the lint compliance system.
// It should be used as a reference for implementation and not compiled directly.