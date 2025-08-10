package tests

import (
	"crypto/rand"
	"fmt"
	"math/big"
	mathRand "math/rand"
	"sync"
	"time"

	"muscle-dreamer/internal/core/ecs"
	"muscle-dreamer/internal/core/ecs/components"
)

const (
	testScreenWidth  = 800
	testScreenHeight = 600
	fullCircleRad    = 6.28
	mockHash         = "mock_hash"
)

// cryptoRandFloat64 generates a cryptographically secure random float64 between 0 and 1.
func cryptoRandFloat64() float64 {
	max := big.NewInt(1 << 53)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		// Fallback to zero in case of error
		return 0.0
	}
	return float64(n.Int64()) / float64(max.Int64())
}

// MockWorld is a test implementation of the World interface
type MockWorld struct {
	mu            sync.RWMutex
	entities      map[ecs.EntityID]bool
	components    map[ecs.EntityID]map[ecs.ComponentType]ecs.Component
	systems       map[ecs.SystemType]ecs.System
	systemEnabled map[ecs.SystemType]bool
	nextEntityID  ecs.EntityID
	config        *ecs.WorldConfig
	metrics       *ecs.PerformanceMetrics
	events        []interface{}
}

// NewMockWorld creates a new mock world for testing
func NewMockWorld() *MockWorld {
	return &MockWorld{
		entities:      make(map[ecs.EntityID]bool),
		components:    make(map[ecs.EntityID]map[ecs.ComponentType]ecs.Component),
		systems:       make(map[ecs.SystemType]ecs.System),
		systemEnabled: make(map[ecs.SystemType]bool),
		nextEntityID:  1,
		config: &ecs.WorldConfig{
			MaxEntities:       10000,
			ComponentPoolSize: 1000,
			EnableMetrics:     true,
			EnableEvents:      true,
			ThreadPoolSize:    4,
			QueryCacheSize:    100,
			GCInterval:        time.Minute,
		},
		metrics: &ecs.PerformanceMetrics{
			EntityCount:    0,
			ComponentCount: 0,
			SystemCount:    0,
			MemoryUsage:    0,
			UpdateTime:     0,
			RenderTime:     0,
		},
		events: make([]interface{}, 0),
	}
}

// Entity management
func (w *MockWorld) CreateEntity() ecs.EntityID {
	w.mu.Lock()
	defer w.mu.Unlock()

	id := w.nextEntityID
	w.nextEntityID++
	w.entities[id] = true
	w.components[id] = make(map[ecs.ComponentType]ecs.Component)
	w.metrics.EntityCount++
	return id
}

func (w *MockWorld) DestroyEntity(entity ecs.EntityID) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.entities[entity] {
		return fmt.Errorf("entity %d does not exist", entity)
	}

	delete(w.entities, entity)
	delete(w.components, entity)
	w.metrics.EntityCount--
	return nil
}

func (w *MockWorld) IsEntityValid(entity ecs.EntityID) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.entities[entity]
}

func (w *MockWorld) GetEntityCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.entities)
}

func (w *MockWorld) GetActiveEntities() []ecs.EntityID {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make([]ecs.EntityID, 0, len(w.entities))
	for id := range w.entities {
		result = append(result, id)
	}
	return result
}

// Component management
func (w *MockWorld) AddComponent(entity ecs.EntityID, component ecs.Component) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.entities[entity] {
		return fmt.Errorf("entity %d does not exist", entity)
	}

	if w.components[entity] == nil {
		w.components[entity] = make(map[ecs.ComponentType]ecs.Component)
	}

	w.components[entity][component.GetType()] = component
	w.metrics.ComponentCount++
	return nil
}

func (w *MockWorld) RemoveComponent(entity ecs.EntityID, componentType ecs.ComponentType) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.entities[entity] {
		return fmt.Errorf("entity %d does not exist", entity)
	}

	if w.components[entity] != nil {
		delete(w.components[entity], componentType)
		w.metrics.ComponentCount--
	}
	return nil
}

func (w *MockWorld) GetComponent(entity ecs.EntityID, componentType ecs.ComponentType) (ecs.Component, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.entities[entity] {
		return nil, fmt.Errorf("entity %d does not exist", entity)
	}

	if comps, ok := w.components[entity]; ok {
		if comp, exists := comps[componentType]; exists {
			return comp, nil
		}
	}
	return nil, fmt.Errorf("component %s not found on entity %d", componentType, entity)
}

func (w *MockWorld) HasComponent(entity ecs.EntityID, componentType ecs.ComponentType) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if comps, ok := w.components[entity]; ok {
		_, exists := comps[componentType]
		return exists
	}
	return false
}

func (w *MockWorld) GetComponents(entity ecs.EntityID) []ecs.Component {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if comps, ok := w.components[entity]; ok {
		result := make([]ecs.Component, 0, len(comps))
		for _, comp := range comps {
			result = append(result, comp)
		}
		return result
	}
	return nil
}

// System management
func (w *MockWorld) RegisterSystem(system ecs.System) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.systems[system.GetType()] = system
	w.systemEnabled[system.GetType()] = true
	w.metrics.SystemCount++
	return nil
}

func (w *MockWorld) UnregisterSystem(systemType ecs.SystemType) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.systems, systemType)
	delete(w.systemEnabled, systemType)
	w.metrics.SystemCount--
	return nil
}

func (w *MockWorld) GetSystem(systemType ecs.SystemType) (ecs.System, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if sys, exists := w.systems[systemType]; exists {
		return sys, nil
	}
	return nil, fmt.Errorf("system %s not found", systemType)
}

func (w *MockWorld) GetAllSystems() []ecs.System {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make([]ecs.System, 0, len(w.systems))
	for _, sys := range w.systems {
		result = append(result, sys)
	}
	return result
}

func (w *MockWorld) EnableSystem(systemType ecs.SystemType) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.systems[systemType]; !exists {
		return fmt.Errorf("system %s not registered", systemType)
	}
	w.systemEnabled[systemType] = true
	return nil
}

func (w *MockWorld) DisableSystem(systemType ecs.SystemType) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.systems[systemType]; !exists {
		return fmt.Errorf("system %s not registered", systemType)
	}
	w.systemEnabled[systemType] = false
	return nil
}

func (w *MockWorld) IsSystemEnabled(systemType ecs.SystemType) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.systemEnabled[systemType]
}

// Update and rendering
func (w *MockWorld) Update(dt float64) error {
	start := time.Now()
	defer func() {
		w.metrics.UpdateTime = time.Since(start)
	}()

	for _, system := range w.systems {
		if w.systemEnabled[system.GetType()] {
			if err := system.Update(w, dt); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *MockWorld) Render(screen interface{}) error {
	start := time.Now()
	defer func() {
		w.metrics.RenderTime = time.Since(start)
	}()

	for _, system := range w.systems {
		if w.systemEnabled[system.GetType()] {
			if err := system.Render(w, screen); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *MockWorld) Shutdown() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.entities = make(map[ecs.EntityID]bool)
	w.components = make(map[ecs.EntityID]map[ecs.ComponentType]ecs.Component)
	w.systems = make(map[ecs.SystemType]ecs.System)
	w.systemEnabled = make(map[ecs.SystemType]bool)
	return nil
}

// Metrics and statistics
func (w *MockWorld) GetMetrics() *ecs.PerformanceMetrics {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.metrics
}

func (w *MockWorld) GetMemoryUsage() *ecs.MemoryUsage {
	usage := int64(len(w.entities)*32 + len(w.components)*64)
	return &ecs.MemoryUsage{
		TotalAllocated: usage,
		TotalReserved:  usage * 2,
		TotalFree:      1024*1024*256 - usage, // 256MB limit
		Fragmentation:  0.1,
		GCCount:        0,
		LastGCTime:     time.Now(),
	}
}

func (w *MockWorld) GetStorageStats() []ecs.StorageStats {
	usage := int64(len(w.entities)*32 + len(w.components)*64)
	return []ecs.StorageStats{
		{
			ComponentType:  "mock",
			ComponentCount: w.metrics.ComponentCount,
			MemoryUsed:     usage,
			MemoryReserved: usage * 2,
			Fragmentation:  0.1,
		},
	}
}

func (w *MockWorld) GetQueryStats() []ecs.QueryStats {
	return []ecs.QueryStats{
		{
			QuerySignature: "mock_query",
			ExecutionCount: 0,
			CacheHitRate:   0.0,
			TotalTime:      0,
		},
	}
}

// Configuration
func (w *MockWorld) GetConfig() *ecs.WorldConfig {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.config
}

func (w *MockWorld) UpdateConfig(config *ecs.WorldConfig) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.config = config
	return nil
}

// Events
func (w *MockWorld) EmitEvent(event ecs.Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.events = append(w.events, event)
	return nil
}

func (w *MockWorld) Subscribe(eventType ecs.EventType, handler ecs.EventHandler) error {
	// Mock implementation - not needed for basic tests
	return nil
}

func (w *MockWorld) Unsubscribe(eventType ecs.EventType, handler ecs.EventHandler) error {
	// Mock implementation - not needed for basic tests
	return nil
}

// Queries
func (w *MockWorld) Query() ecs.QueryBuilder {
	// Mock implementation - return a simple query builder
	return &MockQueryBuilder{world: w}
}

func (w *MockWorld) CreateQuery(builder ecs.QueryBuilder) ecs.QueryResult {
	// Mock implementation - convert to MockQueryBuilder if possible
	if mockBuilder, ok := builder.(*MockQueryBuilder); ok {
		return &MockQueryResult{world: w, builder: mockBuilder}
	}
	// Fallback for other QueryBuilder implementations
	return &MockQueryResult{world: w, builder: &MockQueryBuilder{world: w}}
}

func (w *MockWorld) ExecuteQuery(builder ecs.QueryBuilder) ecs.QueryResult {
	return w.CreateQuery(builder)
}

// Batch operations
func (w *MockWorld) CreateEntities(count int) []ecs.EntityID {
	result := make([]ecs.EntityID, count)
	for i := 0; i < count; i++ {
		result[i] = w.CreateEntity()
	}
	return result
}

func (w *MockWorld) DestroyEntities(entities []ecs.EntityID) error {
	for _, entity := range entities {
		if err := w.DestroyEntity(entity); err != nil {
			return err
		}
	}
	return nil
}

func (w *MockWorld) AddComponents(entity ecs.EntityID, components []ecs.Component) error {
	for _, comp := range components {
		if err := w.AddComponent(entity, comp); err != nil {
			return err
		}
	}
	return nil
}

func (w *MockWorld) RemoveComponents(entity ecs.EntityID, componentTypes []ecs.ComponentType) error {
	for _, compType := range componentTypes {
		if err := w.RemoveComponent(entity, compType); err != nil {
			return err
		}
	}
	return nil
}

// Serialization (mock implementation)
func (w *MockWorld) SerializeEntity(entity ecs.EntityID) ([]byte, error) {
	return []byte(fmt.Sprintf("entity_%d", entity)), nil
}

func (w *MockWorld) DeserializeEntity(data []byte) (ecs.EntityID, error) {
	entity := w.CreateEntity()
	return entity, nil
}

func (w *MockWorld) SerializeWorld() ([]byte, error) {
	return []byte("world_data"), nil
}

func (w *MockWorld) DeserializeWorld(data []byte) error {
	return nil
}

// Thread safety
func (w *MockWorld) Lock() {
	w.mu.Lock()
}

func (w *MockWorld) RLock() {
	w.mu.RLock()
}

func (w *MockWorld) Unlock() {
	w.mu.Unlock()
}

func (w *MockWorld) RUnlock() {
	w.mu.RUnlock()
}

// Helper method for query filtering
func (w *MockWorld) matchesFilter(entity ecs.EntityID, filter interface{}) bool {
	// Simple implementation for testing
	// In real implementation, would check component requirements
	return w.entities[entity]
}

// MockQueryBuilder implements ecs.QueryBuilder
type MockQueryBuilder struct {
	world      *MockWorld
	components []ecs.ComponentType
	excluded   []ecs.ComponentType
}

func (qb *MockQueryBuilder) With(componentType ecs.ComponentType) ecs.QueryBuilder {
	qb.components = append(qb.components, componentType)
	return qb
}

func (qb *MockQueryBuilder) Without(componentType ecs.ComponentType) ecs.QueryBuilder {
	qb.excluded = append(qb.excluded, componentType)
	return qb
}

func (qb *MockQueryBuilder) Build() ecs.QueryResult {
	return &MockQueryResult{world: qb.world, builder: qb}
}

// Required methods for ecs.QueryBuilder interface
func (qb *MockQueryBuilder) WithAll(types []ecs.ComponentType) ecs.QueryBuilder  { return qb }
func (qb *MockQueryBuilder) WithAny(types []ecs.ComponentType) ecs.QueryBuilder  { return qb }
func (qb *MockQueryBuilder) WithNone(types []ecs.ComponentType) ecs.QueryBuilder { return qb }
func (qb *MockQueryBuilder) Where(predicate func(ecs.EntityID, []ecs.Component) bool) ecs.QueryBuilder {
	return qb
}

func (qb *MockQueryBuilder) WhereComponent(componentType ecs.ComponentType, predicate func(ecs.Component) bool) ecs.QueryBuilder {
	return qb
}

func (qb *MockQueryBuilder) WhereEntity(predicate func(ecs.EntityID) bool) ecs.QueryBuilder {
	return qb
}
func (qb *MockQueryBuilder) Limit(n int) ecs.QueryBuilder  { return qb }
func (qb *MockQueryBuilder) Offset(n int) ecs.QueryBuilder { return qb }
func (qb *MockQueryBuilder) OrderBy(compareFn func(ecs.EntityID, ecs.EntityID) bool) ecs.QueryBuilder {
	return qb
}

func (qb *MockQueryBuilder) OrderByComponent(componentType ecs.ComponentType, compareFn func(ecs.Component, ecs.Component) bool) ecs.QueryBuilder {
	return qb
}
func (qb *MockQueryBuilder) Cache(key string) ecs.QueryBuilder                { return qb }
func (qb *MockQueryBuilder) CacheFor(duration time.Duration) ecs.QueryBuilder { return qb }
func (qb *MockQueryBuilder) UseBitset(enable bool) ecs.QueryBuilder           { return qb }
func (qb *MockQueryBuilder) UseIndex(indexName string) ecs.QueryBuilder       { return qb }
func (qb *MockQueryBuilder) WithinRadius(center ecs.Vector2, radius float64) ecs.QueryBuilder {
	return qb
}
func (qb *MockQueryBuilder) WithinBounds(bounds ecs.AABB) ecs.QueryBuilder            { return qb }
func (qb *MockQueryBuilder) Intersects(bounds ecs.AABB) ecs.QueryBuilder              { return qb }
func (qb *MockQueryBuilder) Nearest(point ecs.Vector2, n int) ecs.QueryBuilder        { return qb }
func (qb *MockQueryBuilder) Children(parent ecs.EntityID) ecs.QueryBuilder            { return qb }
func (qb *MockQueryBuilder) Descendants(ancestor ecs.EntityID) ecs.QueryBuilder       { return qb }
func (qb *MockQueryBuilder) Ancestors(descendant ecs.EntityID) ecs.QueryBuilder       { return qb }
func (qb *MockQueryBuilder) Siblings(entity ecs.EntityID) ecs.QueryBuilder            { return qb }
func (qb *MockQueryBuilder) CreatedAfter(timestamp time.Time) ecs.QueryBuilder        { return qb }
func (qb *MockQueryBuilder) ModifiedSince(timestamp time.Time) ecs.QueryBuilder       { return qb }
func (qb *MockQueryBuilder) OlderThan(duration time.Duration) ecs.QueryBuilder        { return qb }
func (qb *MockQueryBuilder) InTimeRange(start, end time.Time) ecs.QueryBuilder        { return qb }
func (qb *MockQueryBuilder) GroupBy(componentType ecs.ComponentType) ecs.QueryBuilder { return qb }
func (qb *MockQueryBuilder) Aggregate(fn func([]ecs.Component) interface{}) ecs.QueryBuilder {
	return qb
}
func (qb *MockQueryBuilder) Count() ecs.QueryBuilder                                   { return qb }
func (qb *MockQueryBuilder) Distinct(componentType ecs.ComponentType) ecs.QueryBuilder { return qb }
func (qb *MockQueryBuilder) Execute() ecs.QueryResult                                  { return qb.Build() }
func (qb *MockQueryBuilder) ExecuteAsync() <-chan ecs.QueryResult {
	ch := make(chan ecs.QueryResult, 1)
	ch <- qb.Build()
	close(ch)
	return ch
}

func (qb *MockQueryBuilder) Stream() <-chan ecs.EntityID {
	ch := make(chan ecs.EntityID)
	go func() {
		defer close(ch)
		for _, entity := range qb.Build().GetEntities() {
			ch <- entity
		}
	}()
	return ch
}

func (qb *MockQueryBuilder) ExecuteWithCallback(callback func(ecs.EntityID, []ecs.Component)) error {
	qb.Build().ForEach(callback)
	return nil
}
func (qb *MockQueryBuilder) ToString() string     { return "mock_query" }
func (qb *MockQueryBuilder) ToHash() string       { return mockHash }
func (qb *MockQueryBuilder) GetSignature() string { return "mock_signature" }
func (qb *MockQueryBuilder) Clone() ecs.QueryBuilder {
	clone := *qb
	return &clone
}

// MockQueryResult implements ecs.QueryResult
type MockQueryResult struct {
	world   *MockWorld
	builder *MockQueryBuilder
}

func (qr *MockQueryResult) GetEntities() []ecs.EntityID {
	result := make([]ecs.EntityID, 0)
	for entity := range qr.world.entities {
		if qr.matchesQuery(entity) {
			result = append(result, entity)
		}
	}
	return result
}

func (qr *MockQueryResult) GetCount() int {
	return len(qr.GetEntities())
}

func (qr *MockQueryResult) ForEach(fn func(ecs.EntityID, []ecs.Component)) {
	for _, entity := range qr.GetEntities() {
		components := qr.world.GetComponents(entity)
		fn(entity, components)
	}
}

func (qr *MockQueryResult) matchesQuery(entity ecs.EntityID) bool {
	if qr.builder == nil {
		return true
	}

	// Check required components
	for _, compType := range qr.builder.components {
		if !qr.world.HasComponent(entity, compType) {
			return false
		}
	}

	// Check excluded components
	for _, compType := range qr.builder.excluded {
		if qr.world.HasComponent(entity, compType) {
			return false
		}
	}

	return true
}

// Required methods for ecs.QueryResult interface
func (qr *MockQueryResult) GetEntityCount() int { return qr.GetCount() }

func (qr *MockQueryResult) GetFirst() (ecs.EntityID, bool) {
	entities := qr.GetEntities()
	if len(entities) > 0 {
		return entities[0], true
	}
	return 0, false
}

func (qr *MockQueryResult) GetLast() (ecs.EntityID, bool) {
	entities := qr.GetEntities()
	if len(entities) > 0 {
		return entities[len(entities)-1], true
	}
	return 0, false
}

func (qr *MockQueryResult) GetAt(index int) (ecs.EntityID, bool) {
	entities := qr.GetEntities()
	if index >= 0 && index < len(entities) {
		return entities[index], true
	}
	return 0, false
}
func (qr *MockQueryResult) IsEmpty() bool { return qr.GetCount() == 0 }
func (qr *MockQueryResult) GetComponents(componentType ecs.ComponentType) []ecs.Component {
	result := make([]ecs.Component, 0)
	for _, entity := range qr.GetEntities() {
		if comp, err := qr.world.GetComponent(entity, componentType); err == nil {
			result = append(result, comp)
		}
	}
	return result
}

func (qr *MockQueryResult) GetComponentsFor(entity ecs.EntityID) []ecs.Component {
	return qr.world.GetComponents(entity)
}

func (qr *MockQueryResult) GetComponentsForEntities(entities []ecs.EntityID) map[ecs.EntityID][]ecs.Component {
	result := make(map[ecs.EntityID][]ecs.Component)
	for _, entity := range entities {
		result[entity] = qr.world.GetComponents(entity)
	}
	return result
}

func (qr *MockQueryResult) GetComponentsByType() map[ecs.ComponentType][]ecs.Component {
	result := make(map[ecs.ComponentType][]ecs.Component)
	for _, entity := range qr.GetEntities() {
		components := qr.world.GetComponents(entity)
		for _, comp := range components {
			compType := comp.GetType()
			if result[compType] == nil {
				result[compType] = make([]ecs.Component, 0)
			}
			result[compType] = append(result[compType], comp)
		}
	}
	return result
}

func (qr *MockQueryResult) ForEachEntity(fn func(ecs.EntityID)) {
	for _, entity := range qr.GetEntities() {
		fn(entity)
	}
}

func (qr *MockQueryResult) ForEachComponent(componentType ecs.ComponentType, fn func(ecs.EntityID, ecs.Component)) {
	for _, entity := range qr.GetEntities() {
		if comp, err := qr.world.GetComponent(entity, componentType); err == nil {
			fn(entity, comp)
		}
	}
}

func (qr *MockQueryResult) Map(fn func(ecs.EntityID, []ecs.Component) interface{}) []interface{} {
	result := make([]interface{}, 0)
	for _, entity := range qr.GetEntities() {
		components := qr.world.GetComponents(entity)
		result = append(result, fn(entity, components))
	}
	return result
}

func (qr *MockQueryResult) Filter(predicate func(ecs.EntityID, []ecs.Component) bool) ecs.QueryResult {
	// Return a new MockQueryResult with filtered entities
	return qr
}

func (qr *MockQueryResult) Transform(fn func(ecs.EntityID, []ecs.Component) (ecs.EntityID, []ecs.Component)) ecs.QueryResult {
	return qr
}
func (qr *MockQueryResult) Take(n int) ecs.QueryResult                                { return qr }
func (qr *MockQueryResult) Skip(n int) ecs.QueryResult                                { return qr }
func (qr *MockQueryResult) Union(other ecs.QueryResult) ecs.QueryResult               { return qr }
func (qr *MockQueryResult) Intersection(other ecs.QueryResult) ecs.QueryResult        { return qr }
func (qr *MockQueryResult) Difference(other ecs.QueryResult) ecs.QueryResult          { return qr }
func (qr *MockQueryResult) SymmetricDifference(other ecs.QueryResult) ecs.QueryResult { return qr }
func (qr *MockQueryResult) GroupBy(fn func(ecs.EntityID, []ecs.Component) string) map[string]ecs.QueryResult {
	result := make(map[string]ecs.QueryResult)
	for _, entity := range qr.GetEntities() {
		components := qr.world.GetComponents(entity)
		key := fn(entity, components)
		if result[key] == nil {
			result[key] = &MockQueryResult{world: qr.world, builder: qr.builder}
		}
	}
	return result
}

func (qr *MockQueryResult) Aggregate(fn func([]ecs.EntityID) interface{}) interface{} {
	return fn(qr.GetEntities())
}

func (qr *MockQueryResult) Reduce(fn func(interface{}, ecs.EntityID, []ecs.Component) interface{}, initial interface{}) interface{} {
	result := initial
	for _, entity := range qr.GetEntities() {
		components := qr.world.GetComponents(entity)
		result = fn(result, entity, components)
	}
	return result
}
func (qr *MockQueryResult) Sort(fn func(ecs.EntityID, ecs.EntityID) bool) ecs.QueryResult { return qr }
func (qr *MockQueryResult) SortByComponent(componentType ecs.ComponentType, compareFn func(ecs.Component, ecs.Component) bool) ecs.QueryResult {
	return qr
}
func (qr *MockQueryResult) ToSlice() []ecs.EntityID { return qr.GetEntities() }
func (qr *MockQueryResult) ToMap() map[ecs.EntityID][]ecs.Component {
	result := make(map[ecs.EntityID][]ecs.Component)
	for _, entity := range qr.GetEntities() {
		result[entity] = qr.world.GetComponents(entity)
	}
	return result
}
func (qr *MockQueryResult) ToJSON() ([]byte, error)                               { return []byte("[]"), nil }
func (qr *MockQueryResult) ToCSV() ([]byte, error)                                { return []byte(""), nil }
func (qr *MockQueryResult) GetQueryTime() time.Duration                           { return 0 }
func (qr *MockQueryResult) GetCacheHit() bool                                     { return false }
func (qr *MockQueryResult) GetResultHash() string                                 { return mockHash }
func (qr *MockQueryResult) GetTimestamp() time.Time                               { return time.Now() }
func (qr *MockQueryResult) GetQuerySignature() string                             { return "mock" }
func (qr *MockQueryResult) Subscribe(callback func(ecs.QueryUpdateEvent)) error   { return nil }
func (qr *MockQueryResult) Unsubscribe(callback func(ecs.QueryUpdateEvent)) error { return nil }
func (qr *MockQueryResult) OnUpdate(callback func(ecs.QueryResult)) error         { return nil }

// Test helper functions
func CreateTestWorld() *MockWorld {
	return NewMockWorld()
}

func CreateTestEntity(world *MockWorld) ecs.EntityID {
	return world.CreateEntity()
}

func CreateTestTransform(x, y float64) *components.TransformComponent {
	return &components.TransformComponent{
		Position: ecs.Vector2{X: x, Y: y},
		Rotation: 0,
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
}

func CreateTestPhysics(vx, vy float64) *components.PhysicsComponent {
	return &components.PhysicsComponent{
		Velocity:     ecs.Vector2{X: vx, Y: vy},
		Acceleration: ecs.Vector2{X: 0, Y: 0},
		Mass:         1.0,
		Friction:     0.0,
		IsStatic:     false,
	}
}

func CreateTestSprite(textureID string) *components.SpriteComponent {
	return &components.SpriteComponent{
		TextureID: textureID,
		SourceRect: ecs.AABB{
			Min: ecs.Vector2{X: 0, Y: 0},
			Max: ecs.Vector2{X: 32, Y: 32},
		},
		Visible: true,
		ZOrder:  0,
	}
}

// MockRenderer for testing rendering system
type MockRenderer struct {
	DrawCalls      []DrawCall
	DrawCallCount  int
	LastTexture    string
	DrawOrder      []string
	ClearCallCount int
}

type DrawCall struct {
	TextureID string
	Position  ecs.Vector2
	Rotation  float64
	Scale     ecs.Vector2
	ZOrder    int
}

func (r *MockRenderer) Clear() {
	r.ClearCallCount++
}

func (r *MockRenderer) DrawSprite(textureID string, position, scale ecs.Vector2, rotation float64, zOrder int) {
	r.DrawCallCount++
	r.LastTexture = textureID
	r.DrawOrder = append(r.DrawOrder, textureID)
	r.DrawCalls = append(r.DrawCalls, DrawCall{
		TextureID: textureID,
		Position:  position,
		Rotation:  rotation,
		Scale:     scale,
		ZOrder:    zOrder,
	})
}

func (r *MockRenderer) Present() {
	// No-op for mock
}

func (r *MockRenderer) Reset() {
	r.DrawCalls = nil
	r.DrawCallCount = 0
	r.LastTexture = ""
	r.DrawOrder = nil
	r.ClearCallCount = 0
}

// MockAudioEngine for testing audio system
type MockAudioEngine struct {
	PlayCalls                 []PlayCall
	PlayCallCount             int
	LastSoundID               string
	LastVolume                float64
	VolumeHistory             []float64
	StopCalls                 []string
	SetListenerPositionCalled bool
	ListenerPosition          ecs.Vector2
	PlayingSounds             map[string]bool
}

type PlayCall struct {
	SoundID  string
	Volume   float64
	Pitch    float64
	Loop     bool
	Position ecs.Vector2
}

func NewMockAudioEngine() *MockAudioEngine {
	return &MockAudioEngine{
		PlayingSounds: make(map[string]bool),
	}
}

func (e *MockAudioEngine) Play(soundID string, volume, pitch float64, loop bool) {
	e.PlayCallCount++
	e.LastSoundID = soundID
	e.LastVolume = volume
	e.VolumeHistory = append(e.VolumeHistory, volume)
	e.PlayCalls = append(e.PlayCalls, PlayCall{
		SoundID: soundID,
		Volume:  volume,
		Pitch:   pitch,
		Loop:    loop,
	})
	e.PlayingSounds[soundID] = true
}

func (e *MockAudioEngine) Play3D(soundID string, position ecs.Vector2, volume, pitch float64, loop bool) {
	e.PlayCallCount++
	e.LastSoundID = soundID
	e.LastVolume = volume
	e.VolumeHistory = append(e.VolumeHistory, volume)
	e.PlayCalls = append(e.PlayCalls, PlayCall{
		SoundID:  soundID,
		Volume:   volume,
		Pitch:    pitch,
		Loop:     loop,
		Position: position,
	})
	e.PlayingSounds[soundID] = true
}

func (e *MockAudioEngine) Stop(soundID string) {
	e.StopCalls = append(e.StopCalls, soundID)
	e.PlayingSounds[soundID] = false
}

func (e *MockAudioEngine) IsPlaying(soundID string) bool {
	return e.PlayingSounds[soundID]
}

func (e *MockAudioEngine) SetListenerPosition(pos ecs.Vector2) error {
	e.SetListenerPositionCalled = true
	e.ListenerPosition = pos
	return nil
}

func (e *MockAudioEngine) LoadSound(soundID, filePath string) error {
	return nil // Mock implementation
}

func (e *MockAudioEngine) PlaySound(soundID string, volume, pitch float64, loop bool) error {
	e.Play(soundID, volume, pitch, loop)
	return nil
}

func (e *MockAudioEngine) SetMasterVolume(volume float64) {
	// No-op for mock
}

func (e *MockAudioEngine) SetVolume(soundID string, volume float64) error {
	return nil // Mock implementation
}

func (e *MockAudioEngine) StopSound(soundID string) error {
	e.Stop(soundID)
	return nil
}

func (e *MockAudioEngine) UnloadSound(soundID string) error {
	return nil // Mock implementation
}

func (e *MockAudioEngine) Reset() {
	e.PlayCalls = nil
	e.PlayCallCount = 0
	e.LastSoundID = ""
	e.LastVolume = 0
	e.VolumeHistory = nil
	e.StopCalls = nil
	e.SetListenerPositionCalled = false
	e.ListenerPosition = ecs.Vector2{}
	e.PlayingSounds = make(map[string]bool)
}

// Random entity creation helpers
func CreateRandomEntities(world *MockWorld, count int) []ecs.EntityID {
	entities := make([]ecs.EntityID, count)
	for i := 0; i < count; i++ {
		entity := world.CreateEntity()

		// Add random transform
		transform := &components.TransformComponent{
			Position: ecs.Vector2{
				X: cryptoRandFloat64() * testScreenWidth,
				Y: cryptoRandFloat64() * testScreenHeight,
			},
			Rotation: cryptoRandFloat64() * fullCircleRad,
			Scale:    ecs.Vector2{X: 1, Y: 1},
		}
		world.AddComponent(entity, transform)

		// 50% chance to add physics
		if mathRand.Float32() < 0.5 {
			physics := &components.PhysicsComponent{
				Velocity: ecs.Vector2{
					X: (mathRand.Float64() - 0.5) * 200,
					Y: (mathRand.Float64() - 0.5) * 200,
				},
				Mass:     mathRand.Float64()*10 + 1,
				IsStatic: mathRand.Float32() < 0.2,
			}
			world.AddComponent(entity, physics)
		}

		// 70% chance to add sprite
		if mathRand.Float32() < 0.7 {
			sprite := &components.SpriteComponent{
				TextureID: fmt.Sprintf("texture_%d", mathRand.Intn(10)),
				SourceRect: ecs.AABB{
					Min: ecs.Vector2{X: 0, Y: 0},
					Max: ecs.Vector2{X: 32, Y: 32},
				},
				Visible: true,
				ZOrder:  mathRand.Intn(10),
			}
			world.AddComponent(entity, sprite)
		}

		entities[i] = entity
	}
	return entities
}
