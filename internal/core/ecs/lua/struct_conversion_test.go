package lua

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

// TestStruct - テスト用構造体
type TestStruct struct {
	Name   string  `json:"name"`
	Age    int     `json:"age"`
	Score  float64 `json:"score"`
	Active bool    `json:"active"`
}

// TestDataConversion_GoStructToLuaTable - Go構造体→Luaテーブル変換テスト
func TestDataConversion_GoStructToLuaTable(t *testing.T) {
	bridge := NewLuaBridge()
	vm := setupTestVM(t, bridge)
	defer bridge.DestroyVM(vm)

	testStruct := TestStruct{
		Name:   "TestPlayer",
		Age:    25,
		Score:  88.5,
		Active: true,
	}

	luaVal, err := bridge.GoToLua(vm, testStruct)
	require.NoError(t, err, "Struct conversion should succeed")
	require.Equal(t, lua.LTTable, luaVal.Type(), "Result should be Lua table")

	luaTable := luaVal.(*lua.LTable)

	assert.Equal(t, "TestPlayer", luaTable.RawGetString("name").String())
	assert.Equal(t, float64(25), float64(lua.LVAsNumber(luaTable.RawGetString("age"))))
	assert.Equal(t, 88.5, float64(lua.LVAsNumber(luaTable.RawGetString("score"))))
	assert.Equal(t, true, lua.LVAsBool(luaTable.RawGetString("active")))
}

// TestDataConversion_GoMapToLuaTable - Goマップ→Luaテーブル変換テスト
func TestDataConversion_GoMapToLuaTable(t *testing.T) {
	bridge := NewLuaBridge()
	vm := setupTestVM(t, bridge)
	defer bridge.DestroyVM(vm)

	testMap := map[string]interface{}{
		"name":   "Player1",
		"level":  42,
		"health": 100.5,
		"alive":  true,
	}

	luaVal, err := bridge.GoToLua(vm, testMap)
	require.NoError(t, err, "Map conversion should succeed")
	require.Equal(t, lua.LTTable, luaVal.Type(), "Result should be Lua table")

	luaTable := luaVal.(*lua.LTable)

	assert.Equal(t, "Player1", luaTable.RawGetString("name").String())
	assert.Equal(t, float64(42), float64(lua.LVAsNumber(luaTable.RawGetString("level"))))
	assert.Equal(t, 100.5, float64(lua.LVAsNumber(luaTable.RawGetString("health"))))
	assert.Equal(t, true, lua.LVAsBool(luaTable.RawGetString("alive")))
}

// TestDataConversion_LuaTableToGoSlice - Luaテーブル→Goスライス変換テスト
func TestDataConversion_LuaTableToGoSlice(t *testing.T) {
	bridge := NewLuaBridge()
	vm := setupTestVM(t, bridge)
	defer bridge.DestroyVM(vm)

	// Luaテーブル作成
	luaTable := vm.state.NewTable()
	luaTable.RawSetInt(1, lua.LString("first"))
	luaTable.RawSetInt(2, lua.LString("second"))
	luaTable.RawSetInt(3, lua.LString("third"))

	var result []string
	err := bridge.LuaToGo(vm, luaTable, &result)
	require.NoError(t, err, "Table to slice conversion should succeed")

	expected := []string{"first", "second", "third"}
	assert.Equal(t, expected, result)
}

// TestDataConversion_TypeError - 型不一致エラーテスト
func TestDataConversion_TypeError(t *testing.T) {
	bridge := NewLuaBridge()
	vm := setupTestVM(t, bridge)
	defer bridge.DestroyVM(vm)

	// nil値変換試行
	_, err := bridge.GoToLua(vm, nil)
	require.NoError(t, err, "nil conversion should succeed (returns LNil)")

	// 未対応型変換試行
	unsupportedValue := make(chan int) // channel型は未対応
	_, err = bridge.GoToLua(vm, unsupportedValue)
	require.Error(t, err, "Unsupported type conversion should fail")
	assert.Contains(t, err.Error(), "unsupported type")

	// 型不一致でのLuaToGo変換
	luaNumber := lua.LNumber(42)
	var stringTarget string
	err = bridge.LuaToGo(vm, luaNumber, &stringTarget)
	require.Error(t, err, "Type mismatch conversion should fail")
}