package mod

import (
	"time"

	"muscle-dreamer/internal/core/ecs"
)

// ModECSAPIImpl はModECSAPIの基本実装
type ModECSAPIImpl struct {
	modID        string
	context      *ModContext
	entityAPI    ModEntityAPI
	componentAPI ModComponentAPI
	queryAPI     ModQueryAPI
	systemAPI    ModSystemAPI
}

// NewModECSAPI 新しいModECSAPIインスタンスを作成
func NewModECSAPI(modID string, config ModConfig) ModECSAPI {
	ctx := &ModContext{
		ModID:             modID,
		MaxEntities:       config.MaxEntities,
		MaxMemory:         config.MaxMemory,
		MaxExecutionTime:  config.MaxExecutionTime,
		AllowedComponents: config.AllowedComponents,
		CreatedEntities:   make([]ecs.EntityID, 0),
		MemoryUsage:       0,
		ExecutionTime:     0,
		QueryCount:        0,
		MaxQueryCount:     config.MaxQueryCount,
	}

	impl := &ModECSAPIImpl{
		modID:   modID,
		context: ctx,
	}

	impl.entityAPI = &ModEntityAPIImpl{
		api:           impl,
		entities:      make(map[ecs.EntityID][]string),
		ownedEntities: make(map[ecs.EntityID]bool),
		monitor:       NewPerformanceMonitor(),
	}
	impl.componentAPI = &ModComponentAPIImpl{
		api:        impl,
		components: make(map[ecs.EntityID]map[ecs.ComponentType]ecs.Component),
		cache:      NewComponentCache(),
	}
	impl.queryAPI = &ModQueryAPIImpl{api: impl}
	impl.systemAPI = &ModSystemAPIImpl{
		api:               impl,
		systems:           make(map[string]ModSystem),
		securityValidator: NewAdvancedSecurityValidator(modID, NewSecurityAuditLogger()),
	}

	return impl
}

func (m *ModECSAPIImpl) Entities() ModEntityAPI {
	return m.entityAPI
}

func (m *ModECSAPIImpl) Components() ModComponentAPI {
	return m.componentAPI
}

func (m *ModECSAPIImpl) Queries() ModQueryAPI {
	return m.queryAPI
}

func (m *ModECSAPIImpl) Systems() ModSystemAPI {
	return m.systemAPI
}

func (m *ModECSAPIImpl) GetContext() *ModContext {
	return m.context
}

// ModEntityAPIImpl はModEntityAPIの基本実装
type ModEntityAPIImpl struct {
	api           *ModECSAPIImpl
	nextID        ecs.EntityID
	entities      map[ecs.EntityID][]string // エンティティIDとタグのマップ
	ownedEntities map[ecs.EntityID]bool     // 高速所有権チェック用
	monitor       *PerformanceMonitor
}

func (m *ModEntityAPIImpl) Create(tags ...string) (ecs.EntityID, error) {
	start := time.Now()
	defer func() {
		m.monitor.RecordAPICall("entity_create", time.Since(start))
	}()

	if len(m.api.context.CreatedEntities) >= m.api.context.MaxEntities {
		return ecs.InvalidEntityID, ErrEntityLimitExceeded
	}

	// メモリ制限チェック
	if m.api.context.MemoryUsage+64 > m.api.context.MaxMemory {
		return ecs.InvalidEntityID, ErrMemoryLimitExceeded
	}

	// 効率的なIDを割り当て
	m.nextID++
	entityID := m.nextID

	// MODタグを自動追加
	allTags := append([]string{"mod:" + m.api.modID}, tags...)

	m.entities[entityID] = allTags
	m.ownedEntities[entityID] = true

	m.api.context.CreatedEntities = append(m.api.context.CreatedEntities, entityID)
	m.api.context.MemoryUsage += 64

	// パフォーマンス監視
	m.monitor.RecordMemorySnapshot(m.api.context.MemoryUsage)

	return entityID, nil
}

func (m *ModEntityAPIImpl) Delete(id ecs.EntityID) error {
	start := time.Now()
	defer func() {
		m.monitor.RecordAPICall("entity_delete", time.Since(start))
	}()

	// 効率的な権限チェック
	if !m.isOwnedEntity(id) {
		// システムエンティティかどうか簡易判定
		if id < 1000 { // 1000未満をシステムエンティティと仮定
			return ErrSystemEntityAccess
		}
		return ErrEntityPermissionDenied
	}

	// エンティティ削除（効率化）
	delete(m.entities, id)
	delete(m.ownedEntities, id)

	// CreatedEntitiesから効率的に削除
	for i, entityID := range m.api.context.CreatedEntities {
		if entityID == id {
			m.api.context.CreatedEntities = append(
				m.api.context.CreatedEntities[:i],
				m.api.context.CreatedEntities[i+1:]...)
			break
		}
	}

	m.api.context.MemoryUsage -= 64 // メモリ使用量を減算
	m.monitor.RecordMemorySnapshot(m.api.context.MemoryUsage)

	return nil
}

func (m *ModEntityAPIImpl) GetTags(id ecs.EntityID) ([]string, error) {
	if tags, exists := m.entities[id]; exists {
		return tags, nil
	}
	return nil, ErrEntityPermissionDenied
}

func (m *ModEntityAPIImpl) GetOwned() ([]ecs.EntityID, error) {
	return m.api.context.CreatedEntities, nil
}

func (m *ModEntityAPIImpl) isOwnedEntity(id ecs.EntityID) bool {
	return m.ownedEntities[id]
}

// ModComponentAPIImpl はModComponentAPIの基本実装
type ModComponentAPIImpl struct {
	api        *ModECSAPIImpl
	components map[ecs.EntityID]map[ecs.ComponentType]ecs.Component
	cache      *ComponentCache
}

func (m *ModComponentAPIImpl) Add(entity ecs.EntityID, component ecs.Component) error {
	// 権限チェック
	entityAPI := m.api.entityAPI.(*ModEntityAPIImpl)
	if !entityAPI.isOwnedEntity(entity) {
		return ErrComponentPermissionDenied
	}

	// コンポーネント型許可チェック
	if !m.IsAllowed(component.GetType()) {
		return ErrComponentNotAllowed
	}

	// 最小実装：メモリマップに保存
	if m.components == nil {
		m.components = make(map[ecs.EntityID]map[ecs.ComponentType]ecs.Component)
	}
	if m.components[entity] == nil {
		m.components[entity] = make(map[ecs.ComponentType]ecs.Component)
	}

	m.components[entity][component.GetType()] = component
	return nil
}

func (m *ModComponentAPIImpl) Get(entity ecs.EntityID, componentType ecs.ComponentType) (ecs.Component, error) {
	// 権限チェック
	entityAPI := m.api.entityAPI.(*ModEntityAPIImpl)
	if !entityAPI.isOwnedEntity(entity) {
		return nil, ErrComponentPermissionDenied
	}

	if m.components == nil {
		return nil, nil
	}

	if entityComponents, exists := m.components[entity]; exists {
		if component, exists := entityComponents[componentType]; exists {
			return component, nil
		}
	}

	return nil, nil
}

func (m *ModComponentAPIImpl) Remove(entity ecs.EntityID, componentType ecs.ComponentType) error {
	// 権限チェック
	entityAPI := m.api.entityAPI.(*ModEntityAPIImpl)
	if !entityAPI.isOwnedEntity(entity) {
		return ErrComponentPermissionDenied
	}

	if m.components != nil && m.components[entity] != nil {
		delete(m.components[entity], componentType)
	}

	return nil
}

func (m *ModComponentAPIImpl) IsAllowed(componentType ecs.ComponentType) bool {
	for _, allowed := range m.api.context.AllowedComponents {
		if componentType == allowed {
			return true
		}
	}
	return false
}

// ModQueryAPIImpl はModQueryAPIの基本実装
type ModQueryAPIImpl struct {
	api *ModECSAPIImpl
}

func (m *ModQueryAPIImpl) Find(query ecs.QueryBuilder) ([]ecs.EntityID, error) {
	// クエリ実行回数制限チェック
	if m.api.context.QueryCount >= m.api.context.MaxQueryCount {
		return nil, ErrQueryLimitExceeded
	}

	m.api.context.QueryCount++

	// 最小実装：自分が作成したエンティティのみ返却
	return m.api.context.CreatedEntities, nil
}

func (m *ModQueryAPIImpl) Count(query ecs.QueryBuilder) (int, error) {
	entities, err := m.Find(query)
	if err != nil {
		return 0, err
	}
	return len(entities), nil
}

func (m *ModQueryAPIImpl) GetExecutionCount() int {
	return m.api.context.QueryCount
}

func (m *ModQueryAPIImpl) ResetExecutionCount() {
	m.api.context.QueryCount = 0
}

// ModSystemAPIImpl はModSystemAPIの基本実装
type ModSystemAPIImpl struct {
	api               *ModECSAPIImpl
	systems           map[string]ModSystem
	securityValidator *AdvancedSecurityValidator
}

func (m *ModSystemAPIImpl) Register(system ModSystem) error {
	// 実行時間制限チェック
	if system.GetMaxExecutionTime() > m.api.context.MaxExecutionTime {
		return ErrSystemExecutionTimeExceeded
	}

	// 高度なセキュリティチェック
	if err := m.securityValidator.ValidateSystemID(system.GetID()); err != nil {
		return err
	}

	m.systems[system.GetID()] = system
	return nil
}

func (m *ModSystemAPIImpl) Unregister(systemID string) error {
	if m.systems != nil {
		delete(m.systems, systemID)
	}
	return nil
}

func (m *ModSystemAPIImpl) GetRegistered() []string {
	if m.systems == nil {
		return []string{}
	}

	result := make([]string, 0, len(m.systems))
	for id := range m.systems {
		result = append(result, id)
	}
	return result
}
