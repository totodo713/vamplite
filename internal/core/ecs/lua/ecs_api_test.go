package lua

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

// TestLuaAPI_EntityManager - EntityManager Lua APIテスト
func TestLuaAPI_EntityManager(t *testing.T) {
	bridge := NewLuaBridge()
	vm := setupTestVM(t, bridge)
	defer bridge.DestroyVM(vm)

	// Mock ECS APIを登録
	mockAPI := NewMockECSAPI()
	modAPI := ModECSAPI(mockAPI)
	err := bridge.RegisterECSAPI(vm, &modAPI)
	require.NoError(t, err, "ECS API registration should succeed")

	// ecs.create_entity() テスト
	luaCode := `
        local entity = ecs.create_entity()
        assert(entity ~= nil, "Created entity should not be nil")
        assert(type(entity) == "number", "Entity ID should be number")
        return entity
    `

	err = vm.state.DoString(luaCode)
	require.NoError(t, err, "Lua entity creation should succeed")

	result := vm.state.Get(-1)
	assert.Equal(t, lua.LTNumber, result.Type())
	entityID := EntityID(lua.LVAsNumber(result))
	assert.True(t, entityID > 0, "Entity ID should be positive")

	// ecs.entity_exists() テスト
	existsCode := `
        local exists = ecs.entity_exists(` + lua.LVAsString(result) + `)
        assert(exists == true, "Created entity should exist")
        return exists
    `

	err = vm.state.DoString(existsCode)
	require.NoError(t, err, "Entity existence check should succeed")

	existsResult := vm.state.Get(-1)
	assert.Equal(t, true, lua.LVAsBool(existsResult))
}

// TestLuaAPI_ComponentStore - ComponentStore Lua APIテスト
func TestLuaAPI_ComponentStore(t *testing.T) {
	bridge := NewLuaBridge()
	vm := setupTestVM(t, bridge)
	defer bridge.DestroyVM(vm)

	mockAPI := NewMockECSAPI()
	modAPI := ModECSAPI(mockAPI)
	err := bridge.RegisterECSAPI(vm, &modAPI)
	require.NoError(t, err)

	// コンポーネント追加・取得・削除のテスト
	componentCode := `
        local entity = ecs.create_entity()
        
        -- Transform コンポーネント追加
        local success = ecs.add_component(entity, "Transform", {
            x = 10.5,
            y = 20.5,
            z = 0.0
        })
        assert(success == true, "Component addition should succeed")
        
        -- コンポーネント存在確認
        local has = ecs.has_component(entity, "Transform")
        assert(has == true, "Entity should have Transform component")
        
        -- コンポーネント取得
        local transform = ecs.get_component(entity, "Transform")
        assert(transform ~= nil, "Component should be retrievable")
        assert(transform.x == 10.5, "Component data should match")
        assert(transform.y == 20.5, "Component data should match")
        assert(transform.z == 0.0, "Component data should match")
        
        -- コンポーネント削除
        local removed = ecs.remove_component(entity, "Transform")
        assert(removed == true, "Component removal should succeed")
        
        -- 削除後の存在確認
        local has_after = ecs.has_component(entity, "Transform")
        assert(has_after == false, "Entity should not have component after removal")
        
        return {entity = entity, success = success}
    `

	err = vm.state.DoString(componentCode)
	require.NoError(t, err, "Component operations should succeed")

	result := vm.state.Get(-1)
	assert.Equal(t, lua.LTTable, result.Type())
}

// TestLuaAPI_QueryEngine - Query Engine Lua APIテスト
func TestLuaAPI_QueryEngine(t *testing.T) {
	bridge := NewLuaBridge()
	vm := setupTestVM(t, bridge)
	defer bridge.DestroyVM(vm)

	mockAPI := NewMockECSAPI()
	modAPI := ModECSAPI(mockAPI)
	err := bridge.RegisterECSAPI(vm, &modAPI)
	require.NoError(t, err)

	queryCode := `
        -- テストエンティティ作成
        local entity1 = ecs.create_entity()
        ecs.add_component(entity1, "Transform", {x=10, y=10, z=0})
        ecs.add_component(entity1, "Sprite", {texture="player.png"})
        
        local entity2 = ecs.create_entity()
        ecs.add_component(entity2, "Transform", {x=20, y=20, z=0})
        
        -- Transform を持つエンティティをクエリ
        local entities = ecs.query()
            :with("Transform")
            :execute()
            
        assert(#entities >= 2, "Should find at least 2 entities")
        
        -- Transform AND Sprite を持つエンティティをクエリ
        local sprite_entities = ecs.query()
            :with("Transform")
            :with("Sprite")
            :execute()
            
        assert(#sprite_entities >= 1, "Should find at least 1 entity with both components")
        
        return {
            all_transform = #entities,
            with_sprite = #sprite_entities
        }
    `

	err = vm.state.DoString(queryCode)
	require.NoError(t, err, "Query operations should succeed")

	result := vm.state.Get(-1)
	assert.Equal(t, lua.LTTable, result.Type())
}