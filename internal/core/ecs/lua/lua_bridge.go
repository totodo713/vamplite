package lua

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/yuin/gopher-lua"
)

// LuaBridgeImpl - LuaBridgeインターフェースの実装
type LuaBridgeImpl struct{}

// NewLuaBridge - LuaBridge実装のコンストラクタ
func NewLuaBridge() LuaBridge {
	return &LuaBridgeImpl{}
}

// CreateVM - Lua VM作成
func (lb *LuaBridgeImpl) CreateVM(config *LuaVMConfig) (*LuaVM, error) {
	if config == nil {
		config = &LuaVMConfig{
			SandboxEnabled: false,
			ResourceLimits: &ResourceLimits{
				MaxExecutionTime: 100 * time.Millisecond,
				MaxMemoryUsage:   10 * 1024 * 1024, // 10MB
			},
		}
	}

	// Lua State作成
	state := lua.NewState()
	if state == nil {
		return nil, errors.New("failed to create Lua state")
	}

	// サンドボックス設定
	var sandbox *Sandbox
	if config.SandboxEnabled {
		sandbox = &Sandbox{
			FileSystemRestricted: true,
			NetworkRestricted:    true,
			OSCommandsBlocked:    true,
		}

		// サンドボックス適用
		err := applySandbox(state, sandbox)
		if err != nil {
			state.Close()
			return nil, fmt.Errorf("failed to apply sandbox: %w", err)
		}
	}

	vm := &LuaVM{
		state:       state,
		sandbox:     sandbox,
		permissions: config.Permissions,
		resources:   config.ResourceLimits,
	}

	return vm, nil
}

// DestroyVM - Lua VM削除
func (lb *LuaBridgeImpl) DestroyVM(vm *LuaVM) error {
	if vm == nil {
		return errors.New("vm is nil")
	}
	if vm.state == nil {
		return errors.New("vm state is nil")
	}

	vm.state.Close()
	return nil
}

// LoadScript - Luaスクリプト読み込み
func (lb *LuaBridgeImpl) LoadScript(vm *LuaVM, scriptPath string) (*LuaScript, error) {
	// 最小実装: 基本的なスクリプト情報のみ作成
	script := &LuaScript{
		path:   scriptPath,
		loaded: false,
		metadata: &ScriptMetadata{
			Name:       scriptPath,
			Version:    "1.0.0",
			APIVersion: "1.0.0",
		},
	}

	return script, nil
}

// UnloadScript - Luaスクリプトアンロード
func (lb *LuaBridgeImpl) UnloadScript(vm *LuaVM, script *LuaScript) error {
	if script == nil {
		return errors.New("script is nil")
	}

	script.loaded = false
	return nil
}

// ExecuteScript - Luaスクリプト実行
func (lb *LuaBridgeImpl) ExecuteScript(vm *LuaVM, script *LuaScript) error {
	if vm == nil || vm.state == nil {
		return errors.New("vm or vm state is nil")
	}
	if script == nil {
		return errors.New("script is nil")
	}

	// 最小実装: 空のスクリプト実行のみ対応
	err := vm.state.DoString("-- empty script")
	if err != nil {
		return fmt.Errorf("script execution failed: %w", err)
	}

	script.loaded = true
	return nil
}

// GoToLua - Go値をLua値に変換
func (lb *LuaBridgeImpl) GoToLua(vm *LuaVM, value interface{}) (lua.LValue, error) {
	if vm == nil || vm.state == nil {
		return nil, errors.New("vm or vm state is nil")
	}

	return convertGoToLua(vm.state, value)
}

// LuaToGo - Lua値をGo値に変換
func (lb *LuaBridgeImpl) LuaToGo(vm *LuaVM, value lua.LValue, target interface{}) error {
	if vm == nil || vm.state == nil {
		return errors.New("vm or vm state is nil")
	}

	return convertLuaToGo(value, target)
}

// RegisterECSAPI - ECS APIをLua VMに登録
func (lb *LuaBridgeImpl) RegisterECSAPI(vm *LuaVM, ecsAPI *ModECSAPI) error {
	if vm == nil || vm.state == nil {
		return errors.New("vm or vm state is nil")
	}
	if ecsAPI == nil {
		return errors.New("ecsAPI is nil")
	}

	// 最小実装: 基本的なダミー関数のみ登録
	ecsTable := vm.state.NewTable()
	vm.state.SetGlobal("ecs", ecsTable)

	return nil
}

// SetPermissions - API権限設定
func (lb *LuaBridgeImpl) SetPermissions(vm *LuaVM, permissions *APIPermissions) error {
	if vm == nil {
		return errors.New("vm is nil")
	}

	vm.permissions = permissions
	return nil
}

// applySandbox - サンドボックス制限を適用
func applySandbox(state *lua.LState, sandbox *Sandbox) error {
	if sandbox == nil {
		return nil
	}

	// 危険な関数・ライブラリを無効化
	if sandbox.FileSystemRestricted {
		// io ライブラリを制限
		state.SetGlobal("io", lua.LNil)
		state.SetGlobal("dofile", lua.LNil)
		state.SetGlobal("loadfile", lua.LNil)
	}

	if sandbox.OSCommandsBlocked {
		// os ライブラリを制限
		state.SetGlobal("os", lua.LNil)
	}

	// debug ライブラリを無効化
	state.SetGlobal("debug", lua.LNil)

	// package ライブラリを制限
	state.SetGlobal("package", lua.LNil)
	state.SetGlobal("require", lua.LNil)

	return nil
}

// convertGoToLua - Go値をLua値に変換
func convertGoToLua(state *lua.LState, value interface{}) (lua.LValue, error) {
	if value == nil {
		return lua.LNil, nil
	}

	switch v := value.(type) {
	case string:
		return lua.LString(v), nil
	case int:
		return lua.LNumber(float64(v)), nil
	case int64:
		return lua.LNumber(float64(v)), nil
	case float32:
		return lua.LNumber(float64(v)), nil
	case float64:
		return lua.LNumber(v), nil
	case bool:
		return lua.LBool(v), nil
	case []string:
		table := state.NewTable()
		for i, item := range v {
			table.RawSetInt(i+1, lua.LString(item)) // 1-indexed
		}
		return table, nil
	case []int:
		table := state.NewTable()
		for i, item := range v {
			table.RawSetInt(i+1, lua.LNumber(float64(item)))
		}
		return table, nil
	case map[string]interface{}:
		table := state.NewTable()
		for key, val := range v {
			luaVal, err := convertGoToLua(state, val)
			if err != nil {
				return nil, err
			}
			table.RawSetString(key, luaVal)
		}
		return table, nil
	default:
		// reflectionを使用してstructを変換
		return convertStructToLua(state, value)
	}
}

// convertStructToLua - 構造体をLuaテーブルに変換（reflection使用）
func convertStructToLua(state *lua.LState, value interface{}) (lua.LValue, error) {
	v := reflect.ValueOf(value)
	t := reflect.TypeOf(value)

	// ポインタの場合は実体を取得
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("unsupported type: %T", value)
	}

	table := state.NewTable()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 非公開フィールドはスキップ
		if !field.CanInterface() {
			continue
		}

		// JSONタグからフィールド名を取得
		fieldName := fieldType.Name
		if tag := fieldType.Tag.Get("json"); tag != "" && tag != "-" {
			fieldName = tag
		}

		// フィールド値をLua値に変換
		luaVal, err := convertGoToLua(state, field.Interface())
		if err != nil {
			return nil, fmt.Errorf("failed to convert field %s: %w", fieldName, err)
		}

		table.RawSetString(fieldName, luaVal)
	}

	return table, nil
}

// convertLuaToGo - Lua値をGo値に変換
func convertLuaToGo(value lua.LValue, target interface{}) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return errors.New("target must be a pointer")
	}

	targetElem := targetValue.Elem()

	switch value.Type() {
	case lua.LTString:
		if targetElem.Kind() == reflect.String {
			targetElem.SetString(string(value.(lua.LString)))
			return nil
		}
	case lua.LTNumber:
		num := float64(value.(lua.LNumber))
		switch targetElem.Kind() {
		case reflect.Int:
			targetElem.SetInt(int64(num))
			return nil
		case reflect.Float64:
			targetElem.SetFloat(num)
			return nil
		}
	case lua.LTBool:
		if targetElem.Kind() == reflect.Bool {
			targetElem.SetBool(bool(value.(lua.LBool)))
			return nil
		}
	case lua.LTTable:
		// スライス変換をサポート
		if targetElem.Kind() == reflect.Slice {
			return convertLuaTableToSlice(value.(*lua.LTable), target)
		}
	case lua.LTNil:
		// nilの場合はゼロ値を設定
		targetElem.Set(reflect.Zero(targetElem.Type()))
		return nil
	}

	return fmt.Errorf("cannot convert Lua %s to Go %s", value.Type(), targetElem.Kind())
}

// convertLuaTableToSlice - LuaテーブルをGoスライスに変換
func convertLuaTableToSlice(table *lua.LTable, target interface{}) error {
	targetValue := reflect.ValueOf(target).Elem()
	elemType := targetValue.Type().Elem()

	var slice reflect.Value

	// テーブルを配列形式で処理（1-indexed）
	table.ForEach(func(key, value lua.LValue) {
		if !slice.IsValid() {
			slice = reflect.MakeSlice(targetValue.Type(), 0, 0)
		}

		elem := reflect.New(elemType).Elem()

		switch elemType.Kind() {
		case reflect.String:
			if value.Type() == lua.LTString {
				elem.SetString(string(value.(lua.LString)))
			}
		case reflect.Int:
			if value.Type() == lua.LTNumber {
				elem.SetInt(int64(float64(value.(lua.LNumber))))
			}
		case reflect.Float64:
			if value.Type() == lua.LTNumber {
				elem.SetFloat(float64(value.(lua.LNumber)))
			}
		}

		slice = reflect.Append(slice, elem)
	})

	if slice.IsValid() {
		targetValue.Set(slice)
	}

	return nil
}