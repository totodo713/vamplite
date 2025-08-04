// Package ecs provides the core Entity Component System framework for Muscle Dreamer.
package ecs

import (
	"time"
)

// ==============================================
// QueryEngine Interface - 高速エンティティクエリシステム
// ==============================================

// QueryEngine provides high-performance entity querying with caching and optimization.
// It uses bitsets and spatial indices for sub-millisecond query execution.
type QueryEngine interface {
	// Query creation and execution
	CreateQuery(QueryBuilder) QueryResult
	CacheQuery(string, QueryBuilder) QueryResult
	Execute(QueryBuilder) QueryResult
	ExecuteCached(string) QueryResult
	ExecuteRaw(string) QueryResult

	// Query optimization and caching
	OptimizeQueries() error
	OptimizeQuery(string) error
	ClearQueryCache() error
	ClearQuery(string) error
	PrewarmCache([]string) error

	// Query cache management
	SetCacheSize(int)
	GetCacheSize() int
	GetCacheHitRate() float64
	GetCachedQueries() []string
	IsCached(string) bool

	// Real-time query updates
	UpdateQueryCache(EntityID, ComponentType, bool) error
	InvalidateQueriesForEntity(EntityID) error
	InvalidateQueriesForComponent(ComponentType) error
	RefreshAllQueries() error

	// Query statistics and monitoring
	GetQueryStats() []QueryStats
	GetQueryStatsForQuery(string) (*QueryStats, error)
	GetEngineStats() *QueryEngineStats
	ResetStats() error

	// Advanced querying
	CreateSpatialQuery(SpatialQueryBuilder) SpatialQueryResult
	CreateTemporalQuery(TemporalQueryBuilder) TemporalQueryResult
	ExecuteComplexQuery(ComplexQueryBuilder) ComplexQueryResult

	// Query subscription for reactive systems
	SubscribeToQuery(string, func(QueryUpdateEvent)) error
	UnsubscribeFromQuery(string, func(QueryUpdateEvent)) error
	GetQuerySubscribers(string) int

	// Performance tuning
	SetQueryTimeout(time.Duration)
	GetQueryTimeout() time.Duration
	EnableQueryProfiling(bool)
	IsQueryProfilingEnabled() bool

	// Batch operations
	ExecuteQueries([]QueryBuilder) []QueryResult
	ExecuteCachedQueries([]string) []QueryResult
	InvalidateQueries([]string) error

	// Serialization
	SerializeQueryCache() ([]byte, error)
	DeserializeQueryCache([]byte) error
	ExportQueries(format string) ([]byte, error)

	// Debug and validation
	ValidateQuery(QueryBuilder) error
	GetDebugInfo() *QueryEngineDebugInfo
	DumpQueryCache() string
}

// ==============================================
// Query Builder Interface
// ==============================================

// QueryBuilder provides a fluent interface for constructing complex entity queries.
type QueryBuilder interface {
	// Component requirements
	With(ComponentType) QueryBuilder
	Without(ComponentType) QueryBuilder
	WithAll([]ComponentType) QueryBuilder
	WithAny([]ComponentType) QueryBuilder
	WithNone([]ComponentType) QueryBuilder

	// Custom filtering
	Where(func(EntityID, []Component) bool) QueryBuilder
	WhereComponent(ComponentType, func(Component) bool) QueryBuilder
	WhereEntity(func(EntityID) bool) QueryBuilder

	// Result modification
	Limit(int) QueryBuilder
	Offset(int) QueryBuilder
	OrderBy(func(EntityID, EntityID) bool) QueryBuilder
	OrderByComponent(ComponentType, func(Component, Component) bool) QueryBuilder

	// Performance optimization
	Cache(string) QueryBuilder
	CacheFor(time.Duration) QueryBuilder
	UseBitset(bool) QueryBuilder
	UseIndex(string) QueryBuilder

	// Spatial queries
	WithinRadius(Vector2, float64) QueryBuilder
	WithinBounds(AABB) QueryBuilder
	Intersects(AABB) QueryBuilder
	Nearest(Vector2, int) QueryBuilder

	// Hierarchical queries
	Children(EntityID) QueryBuilder
	Descendants(EntityID) QueryBuilder
	Ancestors(EntityID) QueryBuilder
	Siblings(EntityID) QueryBuilder

	// Temporal queries
	CreatedAfter(time.Time) QueryBuilder
	ModifiedSince(time.Time) QueryBuilder
	OlderThan(time.Duration) QueryBuilder
	InTimeRange(time.Time, time.Time) QueryBuilder

	// Grouping and aggregation
	GroupBy(ComponentType) QueryBuilder
	Aggregate(func([]Component) interface{}) QueryBuilder
	Count() QueryBuilder
	Distinct(ComponentType) QueryBuilder

	// Execution
	Execute() QueryResult
	ExecuteAsync() <-chan QueryResult
	Stream() <-chan EntityID
	ExecuteWithCallback(func(EntityID, []Component)) error

	// Query serialization
	ToString() string
	ToHash() string
	GetSignature() string
	Clone() QueryBuilder
}

// ==============================================
// Query Result Interface
// ==============================================

// QueryResult contains the results of an entity query with various access methods.
type QueryResult interface {
	// Entity access
	GetEntities() []EntityID
	GetEntityCount() int
	GetFirst() (EntityID, bool)
	GetLast() (EntityID, bool)
	GetAt(int) (EntityID, bool)
	IsEmpty() bool

	// Component access
	GetComponents(ComponentType) []Component
	GetComponentsFor(EntityID) []Component
	GetComponentsForEntities([]EntityID) map[EntityID][]Component
	GetComponentsByType() map[ComponentType][]Component

	// Iteration
	ForEach(func(EntityID, []Component))
	ForEachEntity(func(EntityID))
	ForEachComponent(ComponentType, func(EntityID, Component))
	Map(func(EntityID, []Component) interface{}) []interface{}

	// Filtering and transformation
	Filter(func(EntityID, []Component) bool) QueryResult
	Transform(func(EntityID, []Component) (EntityID, []Component)) QueryResult
	Take(int) QueryResult
	Skip(int) QueryResult

	// Set operations
	Union(QueryResult) QueryResult
	Intersection(QueryResult) QueryResult
	Difference(QueryResult) QueryResult
	SymmetricDifference(QueryResult) QueryResult

	// Grouping and aggregation
	GroupBy(func(EntityID, []Component) string) map[string]QueryResult
	Aggregate(func([]EntityID) interface{}) interface{}
	Reduce(func(interface{}, EntityID, []Component) interface{}, interface{}) interface{}

	// Sorting
	Sort(func(EntityID, EntityID) bool) QueryResult
	SortByComponent(ComponentType, func(Component, Component) bool) QueryResult

	// Data conversion
	ToSlice() []EntityID
	ToMap() map[EntityID][]Component
	ToJSON() ([]byte, error)
	ToCSV() ([]byte, error)

	// Metadata
	GetQueryTime() time.Duration
	GetCacheHit() bool
	GetResultHash() string
	GetTimestamp() time.Time
	GetQuerySignature() string

	// Subscriptions
	Subscribe(func(QueryUpdateEvent)) error
	Unsubscribe(func(QueryUpdateEvent)) error
	OnUpdate(func(QueryResult)) error
}

// ==============================================
// Spatial Query Interface
// ==============================================

// SpatialQueryBuilder provides spatial query capabilities for position-based components.
type SpatialQueryBuilder interface {
	// Spatial constraints
	WithinRadius(center Vector2, radius float64) SpatialQueryBuilder
	WithinBounds(bounds AABB) SpatialQueryBuilder
	IntersectsBounds(bounds AABB) SpatialQueryBuilder
	ContainsPoint(point Vector2) SpatialQueryBuilder

	// Distance-based queries
	Nearest(point Vector2, count int) SpatialQueryBuilder
	Farthest(point Vector2, count int) SpatialQueryBuilder
	WithinDistance(point Vector2, minDistance, maxDistance float64) SpatialQueryBuilder

	// Area-based queries
	InPolygon(vertices []Vector2) SpatialQueryBuilder
	InCircle(center Vector2, radius float64) SpatialQueryBuilder
	InEllipse(center Vector2, radiusX, radiusY float64) SpatialQueryBuilder

	// Ray casting
	RayIntersection(origin Vector2, direction Vector2, maxDistance float64) SpatialQueryBuilder
	LineIntersection(start Vector2, end Vector2) SpatialQueryBuilder

	// Grid-based queries
	InGrid(gridX, gridY, cellSize int) SpatialQueryBuilder
	InGridCell(x, y int) SpatialQueryBuilder
	InGridRange(startX, startY, endX, endY int) SpatialQueryBuilder

	// Hierarchical spatial queries
	InQuadrant(quadrant int) SpatialQueryBuilder
	InSpatialHash(hash uint64) SpatialQueryBuilder

	// Execution
	Execute() SpatialQueryResult
	ExecuteWithDetails() SpatialQueryResultDetailed
}

// SpatialQueryResult contains results of spatial queries.
type SpatialQueryResult interface {
	QueryResult

	// Spatial-specific methods
	GetDistances() []float64
	GetPositions() []Vector2
	GetBounds() AABB
	GetCenter() Vector2
	GetNearestDistance() float64
	GetFarthestDistance() float64

	// Spatial sorting
	SortByDistance(Vector2) SpatialQueryResult
	SortByX() SpatialQueryResult
	SortByY() SpatialQueryResult

	// Spatial filtering
	FilterByDistance(Vector2, float64, float64) SpatialQueryResult
	FilterByBounds(AABB) SpatialQueryResult
}

// SpatialQueryResultDetailed contains detailed spatial query results.
type SpatialQueryResultDetailed interface {
	SpatialQueryResult

	GetIntersectionDetails() []IntersectionDetail
	GetDistanceDetails() []DistanceDetail
	GetRaycastResults() []RaycastResult
}

// IntersectionDetail contains detailed information about spatial intersections.
type IntersectionDetail struct {
	EntityID          EntityID `json:"entity_id"`
	IntersectionPoint Vector2  `json:"intersection_point"`
	IntersectionArea  float64  `json:"intersection_area"`
	OverlapPercentage float64  `json:"overlap_percentage"`
}

// DistanceDetail contains detailed distance information.
type DistanceDetail struct {
	EntityID  EntityID `json:"entity_id"`
	Distance  float64  `json:"distance"`
	Position  Vector2  `json:"position"`
	Direction Vector2  `json:"direction"`
}

// RaycastResult contains results of ray casting queries.
type RaycastResult struct {
	EntityID      EntityID `json:"entity_id"`
	HitPoint      Vector2  `json:"hit_point"`
	Distance      float64  `json:"distance"`
	Normal        Vector2  `json:"normal"`
	HitFromInside bool     `json:"hit_from_inside"`
}

// ==============================================
// Temporal Query Interface
// ==============================================

// TemporalQueryBuilder provides time-based query capabilities.
type TemporalQueryBuilder interface {
	// Time constraints
	CreatedAfter(time.Time) TemporalQueryBuilder
	CreatedBefore(time.Time) TemporalQueryBuilder
	CreatedBetween(time.Time, time.Time) TemporalQueryBuilder
	ModifiedSince(time.Time) TemporalQueryBuilder
	ModifiedBefore(time.Time) TemporalQueryBuilder

	// Age-based queries
	OlderThan(time.Duration) TemporalQueryBuilder
	YoungerThan(time.Duration) TemporalQueryBuilder
	AgeBetween(time.Duration, time.Duration) TemporalQueryBuilder

	// Lifecycle queries
	ActiveSince(time.Time) TemporalQueryBuilder
	InactiveFor(time.Duration) TemporalQueryBuilder
	RecentlyChanged(time.Duration) TemporalQueryBuilder

	// Component history
	ComponentChangedSince(ComponentType, time.Time) TemporalQueryBuilder
	ComponentStable(ComponentType, time.Duration) TemporalQueryBuilder

	// Execution
	Execute() TemporalQueryResult
}

// TemporalQueryResult contains results of temporal queries.
type TemporalQueryResult interface {
	QueryResult

	// Temporal-specific methods
	GetCreationTimes() []time.Time
	GetModificationTimes() []time.Time
	GetAges() []time.Duration
	GetOldest() (EntityID, time.Time, bool)
	GetNewest() (EntityID, time.Time, bool)

	// Temporal sorting
	SortByAge() TemporalQueryResult
	SortByCreationTime() TemporalQueryResult
	SortByModificationTime() TemporalQueryResult

	// Temporal filtering
	FilterByAge(time.Duration, time.Duration) TemporalQueryResult
	FilterByCreationTime(time.Time, time.Time) TemporalQueryResult
}

// ==============================================
// Complex Query Interface
// ==============================================

// ComplexQueryBuilder provides advanced query capabilities combining multiple criteria.
type ComplexQueryBuilder interface {
	// Sub-query combination
	And(QueryBuilder) ComplexQueryBuilder
	Or(QueryBuilder) ComplexQueryBuilder
	Not(QueryBuilder) ComplexQueryBuilder
	Xor(QueryBuilder) ComplexQueryBuilder

	// Nested queries
	Exists(QueryBuilder) ComplexQueryBuilder
	NotExists(QueryBuilder) ComplexQueryBuilder
	ForAll(QueryBuilder, func(EntityID) bool) ComplexQueryBuilder
	ForAny(QueryBuilder, func(EntityID) bool) ComplexQueryBuilder

	// Join operations
	Join(QueryBuilder, func(EntityID, EntityID) bool) ComplexQueryBuilder
	LeftJoin(QueryBuilder, func(EntityID, EntityID) bool) ComplexQueryBuilder
	RightJoin(QueryBuilder, func(EntityID, EntityID) bool) ComplexQueryBuilder

	// Window functions
	WithWindow(WindowFunction) ComplexQueryBuilder
	WithPartition(func(EntityID) string) ComplexQueryBuilder
	WithRowNumber() ComplexQueryBuilder
	WithRank() ComplexQueryBuilder

	// Execution
	Execute() ComplexQueryResult
}

// ComplexQueryResult contains results of complex queries.
type ComplexQueryResult interface {
	QueryResult

	// Complex-specific methods
	GetSubResults() map[string]QueryResult
	GetJoinResults() []JoinResult
	GetWindowResults() []WindowResult
	GetPartitions() map[string]QueryResult
}

// JoinResult contains results of join operations.
type JoinResult struct {
	LeftEntity  EntityID `json:"left_entity"`
	RightEntity EntityID `json:"right_entity"`
	JoinKey     string   `json:"join_key"`
	MatchScore  float64  `json:"match_score"`
}

// WindowResult contains results of window functions.
type WindowResult struct {
	EntityID    EntityID    `json:"entity_id"`
	RowNumber   int64       `json:"row_number"`
	Rank        int64       `json:"rank"`
	DenseRank   int64       `json:"dense_rank"`
	Percentile  float64     `json:"percentile"`
	WindowValue interface{} `json:"window_value"`
}

// WindowFunction represents different types of window functions.
type WindowFunction int

const (
	WindowRowNumber WindowFunction = iota
	WindowRank
	WindowDenseRank
	WindowPercentRank
	WindowCumeDist
	WindowLag
	WindowLead
	WindowFirstValue
	WindowLastValue
)

// ==============================================
// Query Events and Subscriptions
// ==============================================

// QueryUpdateEvent represents a change to query results.
type QueryUpdateEvent struct {
	QueryName     string         `json:"query_name"`
	EventType     QueryEventType `json:"event_type"`
	EntityID      EntityID       `json:"entity_id,omitempty"`
	ComponentType ComponentType  `json:"component_type,omitempty"`
	Timestamp     time.Time      `json:"timestamp"`
	OldResult     QueryResult    `json:"old_result,omitempty"`
	NewResult     QueryResult    `json:"new_result,omitempty"`
}

// QueryEventType represents different types of query events.
type QueryEventType int

const (
	QueryEventEntityAdded QueryEventType = iota
	QueryEventEntityRemoved
	QueryEventEntityModified
	QueryEventResultChanged
	QueryEventCacheInvalidated
	QueryEventOptimized
)

// QuerySubscription manages subscriptions to query changes.
type QuerySubscription interface {
	// Subscription management
	Subscribe(queryName string, callback func(QueryUpdateEvent)) error
	Unsubscribe(queryName string, callback func(QueryUpdateEvent)) error
	UnsubscribeAll(queryName string) error

	// Event filtering
	SubscribeFiltered(queryName string, filter func(QueryUpdateEvent) bool, callback func(QueryUpdateEvent)) error
	SubscribeToEntityChanges(EntityID, func(QueryUpdateEvent)) error
	SubscribeToComponentChanges(ComponentType, func(QueryUpdateEvent)) error

	// Batch subscriptions
	SubscribeToQueries([]string, func(QueryUpdateEvent)) error
	SubscribeToPattern(string, func(QueryUpdateEvent)) error

	// Subscription info
	GetSubscribers(queryName string) int
	GetSubscribedQueries() []string
	IsSubscribed(queryName string) bool

	// Event history
	GetEventHistory(queryName string) []QueryUpdateEvent
	ClearEventHistory(queryName string) error
}

// ==============================================
// Query Cache Management
// ==============================================

// QueryCache manages cached query results for performance optimization.
type QueryCache interface {
	// Cache operations
	Get(key string) (QueryResult, bool)
	Set(key string, result QueryResult) error
	SetWithTTL(key string, result QueryResult, ttl time.Duration) error
	Delete(key string) error
	Clear() error

	// Cache info
	Size() int
	MaxSize() int
	SetMaxSize(int)
	GetHitRate() float64
	GetHitCount() int64
	GetMissCount() int64

	// Cache policies
	SetEvictionPolicy(EvictionPolicy)
	GetEvictionPolicy() EvictionPolicy
	SetTTL(time.Duration)
	GetTTL() time.Duration

	// Cache optimization
	Optimize() error
	Compact() error
	GetFragmentation() float64
	PrewarmCache([]string) error

	// Statistics
	GetStats() *QueryCacheStats
	GetKeyStats(string) *QueryCacheKeyStats
	ResetStats()

	// Serialization
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}

// EvictionPolicy represents different cache eviction policies.
type EvictionPolicy int

const (
	EvictionLRU EvictionPolicy = iota
	EvictionLFU
	EvictionFIFO
	EvictionTTL
	EvictionRandom
)

// QueryCacheStats contains statistics about query cache performance.
type QueryCacheStats struct {
	Size             int       `json:"size"`
	MaxSize          int       `json:"max_size"`
	HitCount         int64     `json:"hit_count"`
	MissCount        int64     `json:"miss_count"`
	HitRate          float64   `json:"hit_rate"`
	EvictionCount    int64     `json:"eviction_count"`
	MemoryUsage      int64     `json:"memory_usage_bytes"`
	AverageKeySize   int64     `json:"average_key_size_bytes"`
	AverageValueSize int64     `json:"average_value_size_bytes"`
	OldestEntry      time.Time `json:"oldest_entry"`
	NewestEntry      time.Time `json:"newest_entry"`
	Fragmentation    float64   `json:"fragmentation"`
}

// QueryCacheKeyStats contains statistics about a specific cache key.
type QueryCacheKeyStats struct {
	Key          string        `json:"key"`
	HitCount     int64         `json:"hit_count"`
	AccessCount  int64         `json:"access_count"`
	CreatedAt    time.Time     `json:"created_at"`
	LastAccessed time.Time     `json:"last_accessed"`
	TTL          time.Duration `json:"ttl"`
	Size         int64         `json:"size_bytes"`
	IsExpired    bool          `json:"is_expired"`
}

// ==============================================
// Query Performance and Statistics
// ==============================================

// QueryStats contains performance statistics for individual queries.
type QueryStats struct {
	QuerySignature     string        `json:"query_signature"`
	QueryName          string        `json:"query_name,omitempty"`
	ExecutionCount     int64         `json:"execution_count"`
	TotalTime          time.Duration `json:"total_time"`
	AverageTime        time.Duration `json:"average_time"`
	MinTime            time.Duration `json:"min_time"`
	MaxTime            time.Duration `json:"max_time"`
	LastExecuted       time.Time     `json:"last_executed"`
	CacheHitRate       float64       `json:"cache_hit_rate"`
	CacheHitCount      int64         `json:"cache_hit_count"`
	CacheMissCount     int64         `json:"cache_miss_count"`
	ResultCount        int64         `json:"average_result_count"`
	EntitysScanned     int64         `json:"entities_scanned"`
	ComponentsAccessed int64         `json:"components_accessed"`
	IndexUsage         []string      `json:"indices_used"`
	OptimizationLevel  int           `json:"optimization_level"`
	MemoryUsage        int64         `json:"memory_usage_bytes"`
	ErrorCount         int64         `json:"error_count"`
	TimeoutCount       int64         `json:"timeout_count"`
}

// QueryEngineStats contains overall statistics about the query engine.
type QueryEngineStats struct {
	TotalQueries        int64                 `json:"total_queries"`
	CachedQueries       int                   `json:"cached_queries"`
	ActiveSubscriptions int                   `json:"active_subscriptions"`
	AverageQueryTime    time.Duration         `json:"average_query_time"`
	TotalQueryTime      time.Duration         `json:"total_query_time"`
	CacheHitRate        float64               `json:"cache_hit_rate"`
	MemoryUsage         int64                 `json:"memory_usage_bytes"`
	OptimizedQueries    int64                 `json:"optimized_queries"`
	IndexUsage          map[string]int64      `json:"index_usage"`
	EntityIndexSize     int                   `json:"entity_index_size"`
	ComponentIndexSize  map[ComponentType]int `json:"component_index_size"`
	SpatialIndexSize    int                   `json:"spatial_index_size"`
	TemporalIndexSize   int                   `json:"temporal_index_size"`
	LastOptimization    time.Time             `json:"last_optimization"`
	OptimizationCount   int64                 `json:"optimization_count"`
}

// QueryEngineDebugInfo provides debugging information about the query engine.
type QueryEngineDebugInfo struct {
	CachedQueries           []string                      `json:"cached_queries"`
	ActiveSubscriptions     map[string]int                `json:"active_subscriptions"`
	IndexStatus             map[string]IndexStatus        `json:"index_status"`
	QueryStats              []QueryStats                  `json:"query_stats"`
	EngineStats             *QueryEngineStats             `json:"engine_stats"`
	CacheStats              *QueryCacheStats              `json:"cache_stats"`
	RecentErrors            []QueryError                  `json:"recent_errors"`
	PerformanceIssues       []QueryPerformanceIssue       `json:"performance_issues"`
	OptimizationSuggestions []QueryOptimizationSuggestion `json:"optimization_suggestions"`
}

// IndexStatus contains information about query indices.
type IndexStatus struct {
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	Size          int       `json:"size"`
	MemoryUsage   int64     `json:"memory_usage_bytes"`
	LastUpdated   time.Time `json:"last_updated"`
	UsageCount    int64     `json:"usage_count"`
	IsOptimal     bool      `json:"is_optimal"`
	Fragmentation float64   `json:"fragmentation"`
}

// QueryError represents an error that occurred during query execution.
type QueryError struct {
	QuerySignature string        `json:"query_signature"`
	ErrorMessage   string        `json:"error_message"`
	ErrorType      string        `json:"error_type"`
	Timestamp      time.Time     `json:"timestamp"`
	Duration       time.Duration `json:"duration"`
	StackTrace     string        `json:"stack_trace,omitempty"`
}

// QueryPerformanceIssue represents a detected performance issue with queries.
type QueryPerformanceIssue struct {
	QuerySignature string             `json:"query_signature"`
	IssueType      QueryIssueType     `json:"issue_type"`
	Severity       QueryIssueSeverity `json:"severity"`
	Description    string             `json:"description"`
	Impact         string             `json:"impact"`
	Recommendation string             `json:"recommendation"`
	DetectedAt     time.Time          `json:"detected_at"`
	Metric         string             `json:"metric"`
	Value          float64            `json:"value"`
	Threshold      float64            `json:"threshold"`
}

// QueryIssueType represents different types of query performance issues.
type QueryIssueType int

const (
	IssueSlowExecution QueryIssueType = iota
	IssueLowCacheHitRate
	IssueHighMemoryUsage
	IssueFrequentTimeouts
	IssueMissingIndex
	IssueIneffientFilter
	IssueExcessiveScanning
)

// QueryIssueSeverity represents the severity of query performance issues.
type QueryIssueSeverity int

const (
	IssueSeverityLow QueryIssueSeverity = iota
	IssueSeverityMedium
	IssueSeverityHigh
	IssueSeverityCritical
)

// QueryOptimizationSuggestion represents a suggestion for optimizing query performance.
type QueryOptimizationSuggestion struct {
	QuerySignature      string                `json:"query_signature"`
	OptimizationType    QueryOptimizationType `json:"optimization_type"`
	Title               string                `json:"title"`
	Description         string                `json:"description"`
	ExpectedImprovement float64               `json:"expected_improvement_percent"`
	ImplementationCost  QueryOptimizationCost `json:"implementation_cost"`
	Priority            int                   `json:"priority"`
	CodeExample         string                `json:"code_example,omitempty"`
}

// QueryOptimizationType represents different types of query optimizations.
type QueryOptimizationType int

const (
	OptimizationAddIndex QueryOptimizationType = iota
	OptimizationRewriteQuery
	OptimizationCacheQuery
	OptimizationUseFiltering
	OptimizationBatchQueries
	OptimizationReduceScope
)

// QueryOptimizationCost represents the cost of implementing an optimization.
type QueryOptimizationCost int

const (
	CostLow QueryOptimizationCost = iota
	CostMedium
	CostHigh
	CostVeryHigh
)

// ==============================================
// Bitset Operations for High-Performance Queries
// ==============================================

// BitSet provides high-performance bit manipulation for entity filtering.
type BitSet interface {
	// Basic operations
	Set(EntityID) BitSet
	Clear(EntityID) BitSet
	Flip(EntityID) BitSet
	Test(EntityID) bool

	// Bulk operations
	SetAll() BitSet
	ClearAll() BitSet
	FlipAll() BitSet

	// Set operations
	And(BitSet) BitSet
	Or(BitSet) BitSet
	Xor(BitSet) BitSet
	Not() BitSet

	// Information
	Count() int
	Size() int
	IsEmpty() bool
	NextSet(EntityID) (EntityID, bool)
	PreviousSet(EntityID) (EntityID, bool)

	// Iteration
	ForEachSet(func(EntityID))
	ForEachClear(func(EntityID))
	ToSlice() []EntityID

	// Serialization
	Marshal() ([]byte, error)
	Unmarshal([]byte) error

	// Statistics
	GetDensity() float64
	GetMemoryUsage() int64
}

// ==============================================
// Query Index Management
// ==============================================

// QueryIndexManager manages indices for optimizing query performance.
type QueryIndexManager interface {
	// Index creation
	CreateComponentIndex(ComponentType) error
	CreateSpatialIndex(ComponentType) error
	CreateTemporalIndex() error
	CreateCompositeIndex([]ComponentType) error
	CreateCustomIndex(string, func(EntityID) []string) error

	// Index management
	DropIndex(string) error
	RebuildIndex(string) error
	OptimizeIndex(string) error
	GetIndexes() []string
	GetIndexInfo(string) *IndexInfo

	// Index usage
	GetOptimalIndex(QueryBuilder) string
	SuggestIndices([]QueryBuilder) []IndexSuggestion
	AnalyzeIndexUsage() *IndexUsageAnalysis

	// Maintenance
	CompactIndices() error
	ValidateIndices() error
	GetIndexFragmentation() map[string]float64
}

// IndexInfo contains information about a specific index.
type IndexInfo struct {
	Name            string          `json:"name"`
	Type            IndexType       `json:"type"`
	ComponentTypes  []ComponentType `json:"component_types"`
	Size            int             `json:"size"`
	MemoryUsage     int64           `json:"memory_usage_bytes"`
	CreatedAt       time.Time       `json:"created_at"`
	LastUpdated     time.Time       `json:"last_updated"`
	UsageCount      int64           `json:"usage_count"`
	Selectivity     float64         `json:"selectivity"`
	Fragmentation   float64         `json:"fragmentation"`
	IsUnique        bool            `json:"is_unique"`
	IsPrimary       bool            `json:"is_primary"`
	MaintenanceCost float64         `json:"maintenance_cost"`
}

// IndexType represents different types of query indices.
type IndexType int

const (
	IndexTypeComponent IndexType = iota
	IndexTypeSpatial
	IndexTypeTemporal
	IndexTypeComposite
	IndexTypeHash
	IndexTypeBTree
	IndexTypeCustom
)

// IndexSuggestion represents a suggestion for creating an index.
type IndexSuggestion struct {
	Name                string          `json:"name"`
	Type                IndexType       `json:"type"`
	ComponentTypes      []ComponentType `json:"component_types"`
	ExpectedImprovement float64         `json:"expected_improvement_percent"`
	MaintenanceCost     float64         `json:"maintenance_cost"`
	MemoryRequirement   int64           `json:"memory_requirement_bytes"`
	Justification       string          `json:"justification"`
	Priority            int             `json:"priority"`
}

// IndexUsageAnalysis contains analysis of index usage patterns.
type IndexUsageAnalysis struct {
	TotalQueries      int64              `json:"total_queries"`
	IndexedQueries    int64              `json:"indexed_queries"`
	IndexUsageRate    float64            `json:"index_usage_rate"`
	MostUsedIndices   []string           `json:"most_used_indices"`
	UnusedIndices     []string           `json:"unused_indices"`
	IndexEfficiency   map[string]float64 `json:"index_efficiency"`
	SuggestedIndices  []IndexSuggestion  `json:"suggested_indices"`
	OptimizationGains map[string]float64 `json:"optimization_gains"`
}
