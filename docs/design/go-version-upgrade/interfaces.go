// Package upgrade provides interfaces and types for Go version upgrade management
package upgrade

import (
	"time"
)

// VersionInfo represents Go version information
type VersionInfo struct {
	Current   string    `json:"current"`
	Target    string    `json:"target"`
	Timestamp time.Time `json:"timestamp"`
}

// BuildConfig represents build configuration for different platforms
type BuildConfig struct {
	Platform     string            `json:"platform"`
	Architecture string            `json:"architecture"`
	GOOS         string            `json:"goos"`
	GOARCH       string            `json:"goarch"`
	BuildFlags   []string          `json:"build_flags"`
	LDFlags      []string          `json:"ld_flags"`
	Tags         []string          `json:"tags"`
	Environment  map[string]string `json:"environment"`
}

// DependencyInfo represents a Go module dependency
type DependencyInfo struct {
	Module   string `json:"module"`
	Version  string `json:"version"`
	Indirect bool   `json:"indirect"`
	Replace  string `json:"replace,omitempty"`
}

// CompatibilityCheck represents the result of a compatibility check
type CompatibilityCheck struct {
	Module      string   `json:"module"`
	Compatible  bool     `json:"compatible"`
	Issues      []string `json:"issues,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// TestResult represents test execution results
type TestResult struct {
	Package  string        `json:"package"`
	Tests    int           `json:"tests"`
	Passed   int           `json:"passed"`
	Failed   int           `json:"failed"`
	Skipped  int           `json:"skipped"`
	Duration time.Duration `json:"duration"`
	Coverage float64       `json:"coverage"`
	Errors   []string      `json:"errors,omitempty"`
}

// BenchmarkResult represents benchmark execution results
type BenchmarkResult struct {
	Name        string    `json:"name"`
	Operations  int       `json:"operations"`
	NsPerOp     float64   `json:"ns_per_op"`
	AllocsPerOp int       `json:"allocs_per_op"`
	BytesPerOp  int       `json:"bytes_per_op"`
	Timestamp   time.Time `json:"timestamp"`
}

// PerformanceMetrics represents performance measurements
type PerformanceMetrics struct {
	BuildTime   time.Duration `json:"build_time"`
	BinarySize  int64         `json:"binary_size"`
	StartupTime time.Duration `json:"startup_time"`
	MemoryUsage int64         `json:"memory_usage"`
	FPS         float64       `json:"fps"`
	CPUUsage    float64       `json:"cpu_usage"`
	MeasuredAt  time.Time     `json:"measured_at"`
	GoVersion   string        `json:"go_version"`
}

// MigrationStep represents a single step in the migration process
type MigrationStep struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // pending, in_progress, completed, failed
	StartedAt   time.Time `json:"started_at,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// MigrationPlan represents the complete migration plan
type MigrationPlan struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	FromVersion string          `json:"from_version"`
	ToVersion   string          `json:"to_version"`
	Steps       []MigrationStep `json:"steps"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Status      string          `json:"status"`
}

// RollbackInfo represents rollback information
type RollbackInfo struct {
	Reason      string    `json:"reason"`
	FromVersion string    `json:"from_version"`
	ToVersion   string    `json:"to_version"`
	Timestamp   time.Time `json:"timestamp"`
	Success     bool      `json:"success"`
	Details     string    `json:"details,omitempty"`
}

// UpgradeManager interface for managing the upgrade process
type UpgradeManager interface {
	// CheckCurrentVersion returns the current Go version
	CheckCurrentVersion() (*VersionInfo, error)

	// ValidateTargetVersion validates if the target version is suitable
	ValidateTargetVersion(target string) error

	// CheckDependencies checks all dependencies for compatibility
	CheckDependencies() ([]CompatibilityCheck, error)

	// CreateMigrationPlan creates a migration plan
	CreateMigrationPlan(from, to string) (*MigrationPlan, error)

	// ExecuteStep executes a single migration step
	ExecuteStep(stepID string) error

	// Rollback performs a rollback to previous version
	Rollback(reason string) (*RollbackInfo, error)

	// GetMetrics returns current performance metrics
	GetMetrics() (*PerformanceMetrics, error)
}

// BuildManager interface for managing builds
type BuildManager interface {
	// Build performs a build for the specified configuration
	Build(config *BuildConfig) error

	// BuildAll builds for all configured platforms
	BuildAll() error

	// Clean removes build artifacts
	Clean() error

	// GetBuildInfo returns information about a build
	GetBuildInfo(platform string) (*BuildConfig, error)
}

// TestManager interface for managing tests
type TestManager interface {
	// RunTests runs all tests
	RunTests() ([]TestResult, error)

	// RunBenchmarks runs benchmark tests
	RunBenchmarks() ([]BenchmarkResult, error)

	// RunIntegrationTests runs integration tests
	RunIntegrationTests() ([]TestResult, error)

	// GetCoverage returns test coverage information
	GetCoverage() (float64, error)
}

// DependencyManager interface for managing dependencies
type DependencyManager interface {
	// ListDependencies lists all dependencies
	ListDependencies() ([]DependencyInfo, error)

	// UpdateDependencies updates all dependencies
	UpdateDependencies() error

	// CheckVulnerabilities checks for known vulnerabilities
	CheckVulnerabilities() ([]string, error)

	// Tidy runs go mod tidy
	Tidy() error
}

// MonitorManager interface for monitoring the system
type MonitorManager interface {
	// StartMonitoring starts the monitoring process
	StartMonitoring() error

	// StopMonitoring stops the monitoring process
	StopMonitoring() error

	// GetCurrentMetrics returns current metrics
	GetCurrentMetrics() (*PerformanceMetrics, error)

	// CompareMetrics compares metrics between versions
	CompareMetrics(before, after *PerformanceMetrics) (map[string]float64, error)

	// SetThresholds sets performance thresholds
	SetThresholds(metrics *PerformanceMetrics) error

	// CheckThresholds checks if current metrics meet thresholds
	CheckThresholds() (bool, []string, error)
}

// ValidationResult represents validation results
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Warnings []string `json:"warnings,omitempty"`
	Errors   []string `json:"errors,omitempty"`
}

// Validator interface for validation operations
type Validator interface {
	// ValidateGoVersion validates Go version compatibility
	ValidateGoVersion(version string) (*ValidationResult, error)

	// ValidateBuild validates build output
	ValidateBuild(platform string) (*ValidationResult, error)

	// ValidateTests validates test results
	ValidateTests(results []TestResult) (*ValidationResult, error)

	// ValidatePerformance validates performance metrics
	ValidatePerformance(metrics *PerformanceMetrics) (*ValidationResult, error)
}

// NotificationLevel represents the severity of a notification
type NotificationLevel string

const (
	NotificationInfo    NotificationLevel = "info"
	NotificationWarning NotificationLevel = "warning"
	NotificationError   NotificationLevel = "error"
	NotificationSuccess NotificationLevel = "success"
)

// Notification represents a system notification
type Notification struct {
	Level     NotificationLevel `json:"level"`
	Message   string            `json:"message"`
	Details   string            `json:"details,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// Notifier interface for sending notifications
type Notifier interface {
	// Notify sends a notification
	Notify(notification *Notification) error

	// NotifyProgress sends a progress update
	NotifyProgress(step *MigrationStep) error

	// NotifyCompletion sends a completion notification
	NotifyCompletion(plan *MigrationPlan) error

	// NotifyError sends an error notification
	NotifyError(err error) error
}
