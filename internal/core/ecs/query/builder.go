package query

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"muscle-dreamer/internal/core/ecs"
)

// SpatialFilterType represents types of spatial filtering
type SpatialFilterType int

const (
	SpatialFilterNone SpatialFilterType = iota
	SpatialFilterRadius
	SpatialFilterBounds
	SpatialFilterIntersects
	SpatialFilterNearest
)

// HierarchyFilterType represents types of hierarchical filtering
type HierarchyFilterType int

const (
	HierarchyFilterNone HierarchyFilterType = iota
	HierarchyFilterChildren
	HierarchyFilterDescendants
	HierarchyFilterAncestors
	HierarchyFilterSiblings
)

// QueryBuilderImpl implements the QueryBuilder interface
type QueryBuilderImpl struct {
	// Component constraints
	requiredComponents ComponentBitSet
	excludedComponents ComponentBitSet
	optionalComponents ComponentBitSet // For WithAny

	// Custom filters
	customFilter     func(ecs.EntityID, []ecs.Component) bool
	componentFilters map[ecs.ComponentType]func(ecs.Component) bool
	entityFilter     func(ecs.EntityID) bool

	// Result modifiers
	limit                int
	offset               int
	orderByFunc          func(ecs.EntityID, ecs.EntityID) bool
	orderByComponentFunc func(ecs.Component, ecs.Component) bool
	orderByComponentType ecs.ComponentType

	// Performance options
	cacheKey      string
	cacheDuration time.Duration
	useBitset     bool
	indexHint     string

	// Spatial query options
	spatialFilter SpatialFilterType
	spatialCenter ecs.Vector2
	spatialRadius float64
	spatialBounds ecs.AABB
	nearestCount  int

	// Hierarchical query options
	hierarchyFilter HierarchyFilterType
	hierarchyRoot   ecs.EntityID

	// Temporal query options
	createdAfter   time.Time
	modifiedSince  time.Time
	olderThan      time.Duration
	timeRangeStart time.Time
	timeRangeEnd   time.Time

	// Grouping and aggregation
	groupByComponent  ecs.ComponentType
	aggregateFunc     func([]ecs.Component) interface{}
	countOnly         bool
	distinctComponent ecs.ComponentType
}

// NewQueryBuilder creates a new query builder instance
func NewQueryBuilder() ecs.QueryBuilder {
	return &QueryBuilderImpl{
		requiredComponents: NewComponentBitSet(),
		excludedComponents: NewComponentBitSet(),
		optionalComponents: NewComponentBitSet(),
		componentFilters:   make(map[ecs.ComponentType]func(ecs.Component) bool),
		limit:              -1, // No limit by default
		offset:             0,
		useBitset:          true, // Use bitset optimization by default
	}
}

// Component requirements

func (qb *QueryBuilderImpl) With(componentType ecs.ComponentType) ecs.QueryBuilder {
	qb.requiredComponents = qb.requiredComponents.Set(componentType)
	return qb
}

func (qb *QueryBuilderImpl) Without(componentType ecs.ComponentType) ecs.QueryBuilder {
	qb.excludedComponents = qb.excludedComponents.Set(componentType)
	return qb
}

func (qb *QueryBuilderImpl) WithAll(componentTypes []ecs.ComponentType) ecs.QueryBuilder {
	for _, ct := range componentTypes {
		qb.requiredComponents = qb.requiredComponents.Set(ct)
	}
	return qb
}

func (qb *QueryBuilderImpl) WithAny(componentTypes []ecs.ComponentType) ecs.QueryBuilder {
	for _, ct := range componentTypes {
		qb.optionalComponents = qb.optionalComponents.Set(ct)
	}
	return qb
}

func (qb *QueryBuilderImpl) WithNone(componentTypes []ecs.ComponentType) ecs.QueryBuilder {
	for _, ct := range componentTypes {
		qb.excludedComponents = qb.excludedComponents.Set(ct)
	}
	return qb
}

// Custom filtering

func (qb *QueryBuilderImpl) Where(filter func(ecs.EntityID, []ecs.Component) bool) ecs.QueryBuilder {
	qb.customFilter = filter
	return qb
}

func (qb *QueryBuilderImpl) WhereComponent(componentType ecs.ComponentType, filter func(ecs.Component) bool) ecs.QueryBuilder {
	qb.componentFilters[componentType] = filter
	return qb
}

func (qb *QueryBuilderImpl) WhereEntity(filter func(ecs.EntityID) bool) ecs.QueryBuilder {
	qb.entityFilter = filter
	return qb
}

// Result modification

func (qb *QueryBuilderImpl) Limit(n int) ecs.QueryBuilder {
	if n >= 0 {
		qb.limit = n
	}
	return qb
}

func (qb *QueryBuilderImpl) Offset(n int) ecs.QueryBuilder {
	if n >= 0 {
		qb.offset = n
	}
	return qb
}

func (qb *QueryBuilderImpl) OrderBy(sortFunc func(ecs.EntityID, ecs.EntityID) bool) ecs.QueryBuilder {
	qb.orderByFunc = sortFunc
	return qb
}

func (qb *QueryBuilderImpl) OrderByComponent(componentType ecs.ComponentType, sortFunc func(ecs.Component, ecs.Component) bool) ecs.QueryBuilder {
	qb.orderByComponentType = componentType
	qb.orderByComponentFunc = sortFunc
	return qb
}

// Performance optimization

func (qb *QueryBuilderImpl) Cache(key string) ecs.QueryBuilder {
	qb.cacheKey = key
	return qb
}

func (qb *QueryBuilderImpl) CacheFor(duration time.Duration) ecs.QueryBuilder {
	qb.cacheDuration = duration
	return qb
}

func (qb *QueryBuilderImpl) UseBitset(use bool) ecs.QueryBuilder {
	qb.useBitset = use
	return qb
}

func (qb *QueryBuilderImpl) UseIndex(indexName string) ecs.QueryBuilder {
	qb.indexHint = indexName
	return qb
}

// Spatial queries

func (qb *QueryBuilderImpl) WithinRadius(center ecs.Vector2, radius float64) ecs.QueryBuilder {
	qb.spatialFilter = SpatialFilterRadius
	qb.spatialCenter = center
	qb.spatialRadius = radius
	return qb
}

func (qb *QueryBuilderImpl) WithinBounds(bounds ecs.AABB) ecs.QueryBuilder {
	qb.spatialFilter = SpatialFilterBounds
	qb.spatialBounds = bounds
	return qb
}

func (qb *QueryBuilderImpl) Intersects(bounds ecs.AABB) ecs.QueryBuilder {
	qb.spatialFilter = SpatialFilterIntersects
	qb.spatialBounds = bounds
	return qb
}

func (qb *QueryBuilderImpl) Nearest(point ecs.Vector2, count int) ecs.QueryBuilder {
	qb.spatialFilter = SpatialFilterNearest
	qb.spatialCenter = point
	qb.nearestCount = count
	return qb
}

// Hierarchical queries

func (qb *QueryBuilderImpl) Children(parent ecs.EntityID) ecs.QueryBuilder {
	qb.hierarchyFilter = HierarchyFilterChildren
	qb.hierarchyRoot = parent
	return qb
}

func (qb *QueryBuilderImpl) Descendants(root ecs.EntityID) ecs.QueryBuilder {
	qb.hierarchyFilter = HierarchyFilterDescendants
	qb.hierarchyRoot = root
	return qb
}

func (qb *QueryBuilderImpl) Ancestors(child ecs.EntityID) ecs.QueryBuilder {
	qb.hierarchyFilter = HierarchyFilterAncestors
	qb.hierarchyRoot = child
	return qb
}

func (qb *QueryBuilderImpl) Siblings(entity ecs.EntityID) ecs.QueryBuilder {
	qb.hierarchyFilter = HierarchyFilterSiblings
	qb.hierarchyRoot = entity
	return qb
}

// Temporal queries

func (qb *QueryBuilderImpl) CreatedAfter(t time.Time) ecs.QueryBuilder {
	qb.createdAfter = t
	return qb
}

func (qb *QueryBuilderImpl) ModifiedSince(t time.Time) ecs.QueryBuilder {
	qb.modifiedSince = t
	return qb
}

func (qb *QueryBuilderImpl) OlderThan(duration time.Duration) ecs.QueryBuilder {
	qb.olderThan = duration
	return qb
}

func (qb *QueryBuilderImpl) InTimeRange(start, end time.Time) ecs.QueryBuilder {
	qb.timeRangeStart = start
	qb.timeRangeEnd = end
	return qb
}

// Grouping and aggregation

func (qb *QueryBuilderImpl) GroupBy(componentType ecs.ComponentType) ecs.QueryBuilder {
	qb.groupByComponent = componentType
	return qb
}

func (qb *QueryBuilderImpl) Aggregate(aggregator func([]ecs.Component) interface{}) ecs.QueryBuilder {
	qb.aggregateFunc = aggregator
	return qb
}

func (qb *QueryBuilderImpl) Count() ecs.QueryBuilder {
	qb.countOnly = true
	return qb
}

func (qb *QueryBuilderImpl) Distinct(componentType ecs.ComponentType) ecs.QueryBuilder {
	qb.distinctComponent = componentType
	return qb
}

// Execution

func (qb *QueryBuilderImpl) Execute() ecs.QueryResult {
	// This would be implemented by the QueryEngine
	// For now, return a stub
	return nil
}

func (qb *QueryBuilderImpl) ExecuteAsync() <-chan ecs.QueryResult {
	// This would be implemented by the QueryEngine
	// For now, return a stub channel
	ch := make(chan ecs.QueryResult, 1)
	close(ch)
	return ch
}

func (qb *QueryBuilderImpl) Stream() <-chan ecs.EntityID {
	// This would be implemented by the QueryEngine
	// For now, return a stub channel
	ch := make(chan ecs.EntityID)
	close(ch)
	return ch
}

func (qb *QueryBuilderImpl) ExecuteWithCallback(callback func(ecs.EntityID, []ecs.Component)) error {
	// This would be implemented by the QueryEngine
	return nil
}

// Query serialization

func (qb *QueryBuilderImpl) ToString() string {
	var parts []string

	// Required components
	if qb.requiredComponents.Count() > 0 {
		required := qb.requiredComponents.GetSetComponentTypes()
		parts = append(parts, fmt.Sprintf("required:[%s]", formatComponentTypes(required)))
	}

	// Excluded components
	if qb.excludedComponents.Count() > 0 {
		excluded := qb.excludedComponents.GetSetComponentTypes()
		parts = append(parts, fmt.Sprintf("excluded:[%s]", formatComponentTypes(excluded)))
	}

	// Optional components
	if qb.optionalComponents.Count() > 0 {
		optional := qb.optionalComponents.GetSetComponentTypes()
		parts = append(parts, fmt.Sprintf("optional:[%s]", formatComponentTypes(optional)))
	}

	// Modifiers
	if qb.limit >= 0 {
		parts = append(parts, fmt.Sprintf("limit:%d", qb.limit))
	}
	if qb.offset > 0 {
		parts = append(parts, fmt.Sprintf("offset:%d", qb.offset))
	}

	// Spatial constraints
	switch qb.spatialFilter {
	case SpatialFilterRadius:
		parts = append(parts, fmt.Sprintf("radius:%.2f@(%.2f,%.2f)",
			qb.spatialRadius, qb.spatialCenter.X, qb.spatialCenter.Y))
	case SpatialFilterBounds:
		parts = append(parts, fmt.Sprintf("bounds:[(%.2f,%.2f)-(%.2f,%.2f)]",
			qb.spatialBounds.Min.X, qb.spatialBounds.Min.Y,
			qb.spatialBounds.Max.X, qb.spatialBounds.Max.Y))
	case SpatialFilterNearest:
		parts = append(parts, fmt.Sprintf("nearest:%d@(%.2f,%.2f)",
			qb.nearestCount, qb.spatialCenter.X, qb.spatialCenter.Y))
	}

	// Hierarchy constraints
	if qb.hierarchyFilter != HierarchyFilterNone {
		parts = append(parts, fmt.Sprintf("hierarchy:%v[%d]", qb.hierarchyFilter, qb.hierarchyRoot))
	}

	// Cache settings
	if qb.cacheKey != "" {
		parts = append(parts, fmt.Sprintf("cache:%s", qb.cacheKey))
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

func (qb *QueryBuilderImpl) ToHash() string {
	// Generate a deterministic hash based on query parameters
	h := sha256.New()

	// Hash component constraints
	h.Write([]byte(fmt.Sprintf("req:%v", qb.requiredComponents)))
	h.Write([]byte(fmt.Sprintf("exc:%v", qb.excludedComponents)))
	h.Write([]byte(fmt.Sprintf("opt:%v", qb.optionalComponents)))

	// Hash modifiers
	h.Write([]byte(fmt.Sprintf("lim:%d", qb.limit)))
	h.Write([]byte(fmt.Sprintf("off:%d", qb.offset)))

	// Hash spatial constraints
	h.Write([]byte(fmt.Sprintf("spat:%v:%v:%v", qb.spatialFilter, qb.spatialCenter, qb.spatialRadius)))

	// Hash hierarchy constraints
	h.Write([]byte(fmt.Sprintf("hier:%v:%v", qb.hierarchyFilter, qb.hierarchyRoot)))

	return hex.EncodeToString(h.Sum(nil))[:16] // Return first 16 chars for brevity
}

func (qb *QueryBuilderImpl) GetSignature() string {
	// Generate a unique signature for this query configuration
	return fmt.Sprintf("Q_%s", qb.ToHash())
}

func (qb *QueryBuilderImpl) Clone() ecs.QueryBuilder {
	clone := &QueryBuilderImpl{
		requiredComponents: qb.requiredComponents,
		excludedComponents: qb.excludedComponents,
		optionalComponents: qb.optionalComponents,

		customFilter:     qb.customFilter,
		componentFilters: make(map[ecs.ComponentType]func(ecs.Component) bool),
		entityFilter:     qb.entityFilter,

		limit:                qb.limit,
		offset:               qb.offset,
		orderByFunc:          qb.orderByFunc,
		orderByComponentFunc: qb.orderByComponentFunc,
		orderByComponentType: qb.orderByComponentType,

		cacheKey:      qb.cacheKey,
		cacheDuration: qb.cacheDuration,
		useBitset:     qb.useBitset,
		indexHint:     qb.indexHint,

		spatialFilter: qb.spatialFilter,
		spatialCenter: qb.spatialCenter,
		spatialRadius: qb.spatialRadius,
		spatialBounds: qb.spatialBounds,
		nearestCount:  qb.nearestCount,

		hierarchyFilter: qb.hierarchyFilter,
		hierarchyRoot:   qb.hierarchyRoot,

		createdAfter:   qb.createdAfter,
		modifiedSince:  qb.modifiedSince,
		olderThan:      qb.olderThan,
		timeRangeStart: qb.timeRangeStart,
		timeRangeEnd:   qb.timeRangeEnd,

		groupByComponent:  qb.groupByComponent,
		aggregateFunc:     qb.aggregateFunc,
		countOnly:         qb.countOnly,
		distinctComponent: qb.distinctComponent,
	}

	// Deep copy component filters
	for k, v := range qb.componentFilters {
		clone.componentFilters[k] = v
	}

	return clone
}

// Helper function to format component types
func formatComponentTypes(types []ecs.ComponentType) string {
	strs := make([]string, len(types))
	for i, t := range types {
		strs[i] = string(t)
	}
	return strings.Join(strs, ",")
}

// GetRequiredComponents returns the required components bitset
func (qb *QueryBuilderImpl) GetRequiredComponents() ComponentBitSet {
	return qb.requiredComponents
}

// GetExcludedComponents returns the excluded components bitset
func (qb *QueryBuilderImpl) GetExcludedComponents() ComponentBitSet {
	return qb.excludedComponents
}

// GetOptionalComponents returns the optional components bitset
func (qb *QueryBuilderImpl) GetOptionalComponents() ComponentBitSet {
	return qb.optionalComponents
}

// HasSpatialConstraints returns true if the query has spatial constraints
func (qb *QueryBuilderImpl) HasSpatialConstraints() bool {
	return qb.spatialFilter != SpatialFilterNone
}

// HasHierarchicalConstraints returns true if the query has hierarchical constraints
func (qb *QueryBuilderImpl) HasHierarchicalConstraints() bool {
	return qb.hierarchyFilter != HierarchyFilterNone
}

// HasTemporalConstraints returns true if the query has temporal constraints
func (qb *QueryBuilderImpl) HasTemporalConstraints() bool {
	return !qb.createdAfter.IsZero() ||
		!qb.modifiedSince.IsZero() ||
		qb.olderThan > 0 ||
		!qb.timeRangeStart.IsZero()
}

// HasCustomFilters returns true if the query has custom filter functions
func (qb *QueryBuilderImpl) HasCustomFilters() bool {
	return qb.customFilter != nil ||
		qb.entityFilter != nil ||
		len(qb.componentFilters) > 0
}

// GetCacheKey returns the cache key if set
func (qb *QueryBuilderImpl) GetCacheKey() string {
	if qb.cacheKey != "" {
		return qb.cacheKey
	}
	// Generate automatic cache key based on query signature
	return qb.GetSignature()
}

// IsValid checks if the query configuration is valid
func (qb *QueryBuilderImpl) IsValid() bool {
	// Check for conflicting constraints
	if qb.requiredComponents.Intersects(qb.excludedComponents) {
		return false // Can't require and exclude the same component
	}

	// Check spatial constraints
	if qb.spatialFilter == SpatialFilterRadius && qb.spatialRadius <= 0 {
		return false
	}

	if qb.spatialFilter == SpatialFilterNearest && qb.nearestCount <= 0 {
		return false
	}

	// Check temporal constraints
	if !qb.timeRangeStart.IsZero() && !qb.timeRangeEnd.IsZero() {
		if qb.timeRangeEnd.Before(qb.timeRangeStart) {
			return false
		}
	}

	return true
}
