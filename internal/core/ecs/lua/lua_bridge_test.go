package lua

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

// TestLuaBridge_CreateDestroyVM - VM作成・削除の基本テスト
func TestLuaBridge_CreateDestroyVM(t *testing.T) {
	bridge := NewLuaBridge()

	config := &LuaVMConfig{
		SandboxEnabled: true,
		ResourceLimits: &ResourceLimits{
			MaxExecutionTime: 100 * time.Millisecond,
			MaxMemoryUsage:   10 * 1024 * 1024, // 10MB
		},
	}

	// VM作成テスト
	vm, err := bridge.CreateVM(config)
	require.NoError(t, err, "VM creation should succeed")
	require.NotNil(t, vm, "Created VM should not be nil")
	require.NotNil(t, vm.state, "VM state should not be nil")

	// VM削除テスト
	err = bridge.DestroyVM(vm)
	require.NoError(t, err, "VM destruction should succeed")
}

// TestDataConversion_GoToLua_BasicTypes - 基本型変換テスト
func TestDataConversion_GoToLua_BasicTypes(t *testing.T) {
	bridge := NewLuaBridge()
	vm := setupTestVM(t, bridge)
	defer bridge.DestroyVM(vm)

	testCases := []struct {
		name     string
		input    interface{}
		expected lua.LValueType
		value    interface{}
	}{
		{"string", "hello world", lua.LTString, "hello world"},
		{"int", 42, lua.LTNumber, float64(42)},
		{"float64", 3.14159, lua.LTNumber, 3.14159},
		{"bool_true", true, lua.LTBool, true},
		{"bool_false", false, lua.LTBool, false},
		{"empty_string", "", lua.LTString, ""},
		{"zero_int", 0, lua.LTNumber, float64(0)},
		{"negative_int", -100, lua.LTNumber, float64(-100)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// GoからLuaに変換
			luaVal, err := bridge.GoToLua(vm, tc.input)
			require.NoError(t, err, "GoToLua conversion should succeed for %s", tc.name)
			assert.Equal(t, tc.expected, luaVal.Type(), "Lua type should match expected")

			// 値の検証
			switch tc.expected {
			case lua.LTString:
				assert.Equal(t, tc.value, luaVal.String())
			case lua.LTNumber:
				assert.Equal(t, tc.value, float64(lua.LVAsNumber(luaVal)))
			case lua.LTBool:
				assert.Equal(t, tc.value, lua.LVAsBool(luaVal))
			}
		})
	}
}

// TestDataConversion_LuaToGo_BasicTypes - Lua→Go基本型変換テスト
func TestDataConversion_LuaToGo_BasicTypes(t *testing.T) {
	bridge := NewLuaBridge()
	vm := setupTestVM(t, bridge)
	defer bridge.DestroyVM(vm)

	testCases := []struct {
		name      string
		luaValue  lua.LValue
		targetPtr interface{}
		expected  interface{}
	}{
		{"string", lua.LString("test string"), new(string), "test string"},
		{"int", lua.LNumber(123), new(int), 123},
		{"float64", lua.LNumber(2.71828), new(float64), 2.71828},
		{"bool_true", lua.LTrue, new(bool), true},
		{"bool_false", lua.LFalse, new(bool), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := bridge.LuaToGo(vm, tc.luaValue, tc.targetPtr)
			require.NoError(t, err, "LuaToGo conversion should succeed for %s", tc.name)

			// 値の検証（ポインタから値を取得）
			switch ptr := tc.targetPtr.(type) {
			case *string:
				assert.Equal(t, tc.expected, *ptr)
			case *int:
				assert.Equal(t, tc.expected, *ptr)
			case *float64:
				assert.Equal(t, tc.expected, *ptr)
			case *bool:
				assert.Equal(t, tc.expected, *ptr)
			}
		})
	}
}

// TestDataConversion_GoSliceToLuaTable - Goスライス→Luaテーブル変換テスト
func TestDataConversion_GoSliceToLuaTable(t *testing.T) {
	bridge := NewLuaBridge()
	vm := setupTestVM(t, bridge)
	defer bridge.DestroyVM(vm)

	testSlice := []string{"apple", "banana", "cherry"}

	luaVal, err := bridge.GoToLua(vm, testSlice)
	require.NoError(t, err, "Slice conversion should succeed")
	require.Equal(t, lua.LTTable, luaVal.Type(), "Result should be Lua table")

	luaTable := luaVal.(*lua.LTable)

	// Luaテーブルは1-indexedなので注意
	assert.Equal(t, "apple", luaTable.RawGetInt(1).String())
	assert.Equal(t, "banana", luaTable.RawGetInt(2).String())
	assert.Equal(t, "cherry", luaTable.RawGetInt(3).String())
	assert.Equal(t, 3, luaTable.Len())
}

// setupTestVM - テスト用VM作成ヘルパー
func setupTestVM(t *testing.T, bridge LuaBridge) *LuaVM {
	config := &LuaVMConfig{
		SandboxEnabled: false, // テスト用は無効
		ResourceLimits: &ResourceLimits{
			MaxExecutionTime: 1 * time.Second,
			MaxMemoryUsage:   50 * 1024 * 1024, // 50MB
		},
	}

	vm, err := bridge.CreateVM(config)
	require.NoError(t, err)
	return vm
}