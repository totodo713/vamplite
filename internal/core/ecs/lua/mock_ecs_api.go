package lua

import (
	"errors"
)

// MockECSAPI - テスト用のMock ECS API実装
type MockECSAPI struct {
	entities   map[EntityID]bool
	components map[EntityID]map[string]interface{}
	nextID     EntityID
}

// NewMockECSAPI - Mock ECS API作成
func NewMockECSAPI() *MockECSAPI {
	return &MockECSAPI{
		entities:   make(map[EntityID]bool),
		components: make(map[EntityID]map[string]interface{}),
		nextID:     1,
	}
}

func (m *MockECSAPI) CreateEntity() (EntityID, error) {
	id := m.nextID
	m.nextID++
	m.entities[id] = true
	m.components[id] = make(map[string]interface{})
	return id, nil
}

func (m *MockECSAPI) DestroyEntity(id EntityID) error {
	if !m.entities[id] {
		return errors.New("entity not found")
	}
	delete(m.entities, id)
	delete(m.components, id)
	return nil
}

func (m *MockECSAPI) EntityExists(id EntityID) bool {
	return m.entities[id]
}

func (m *MockECSAPI) AddComponent(entityID EntityID, componentType string, data interface{}) error {
	if !m.entities[entityID] {
		return errors.New("entity not found")
	}
	m.components[entityID][componentType] = data
	return nil
}

func (m *MockECSAPI) RemoveComponent(entityID EntityID, componentType string) error {
	if !m.entities[entityID] {
		return errors.New("entity not found")
	}
	delete(m.components[entityID], componentType)
	return nil
}

func (m *MockECSAPI) GetComponent(entityID EntityID, componentType string) (interface{}, error) {
	if !m.entities[entityID] {
		return nil, errors.New("entity not found")
	}
	component, exists := m.components[entityID][componentType]
	if !exists {
		return nil, nil
	}
	return component, nil
}

func (m *MockECSAPI) HasComponent(entityID EntityID, componentType string) bool {
	if !m.entities[entityID] {
		return false
	}
	_, exists := m.components[entityID][componentType]
	return exists
}

func (m *MockECSAPI) QueryEntities() QueryBuilder {
	return NewMockQueryBuilder(m)
}

func (m *MockECSAPI) FireEvent(eventType string, data interface{}) error {
	// Mock implementation - 実際には何もしない
	return nil
}

func (m *MockECSAPI) SubscribeEvent(eventType string, callback func(interface{})) error {
	// Mock implementation - 実際には何もしない
	return nil
}

// MockQueryBuilder - Mock Query Builder実装
type MockQueryBuilder struct {
	api          *MockECSAPI
	withTypes    []string
	withoutTypes []string
}

func NewMockQueryBuilder(api *MockECSAPI) *MockQueryBuilder {
	return &MockQueryBuilder{api: api}
}

func (m *MockQueryBuilder) With(componentType string) QueryBuilder {
	m.withTypes = append(m.withTypes, componentType)
	return m
}

func (m *MockQueryBuilder) Without(componentType string) QueryBuilder {
	m.withoutTypes = append(m.withoutTypes, componentType)
	return m
}

func (m *MockQueryBuilder) Execute() ([]EntityID, error) {
	var results []EntityID

	for entityID := range m.api.entities {
		matches := true

		// WITH条件チェック
		for _, componentType := range m.withTypes {
			if !m.api.HasComponent(entityID, componentType) {
				matches = false
				break
			}
		}

		// WITHOUT条件チェック
		if matches {
			for _, componentType := range m.withoutTypes {
				if m.api.HasComponent(entityID, componentType) {
					matches = false
					break
				}
			}
		}

		if matches {
			results = append(results, entityID)
		}
	}

	return results, nil
}
