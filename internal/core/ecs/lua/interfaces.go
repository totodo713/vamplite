package lua

import (
	"time"

	lua "github.com/yuin/gopher-lua"
)

// LuaBridge - メインのLua統合インターフェース
type LuaBridge interface {
	// Lua VM管理
	CreateVM(config *LuaVMConfig) (*LuaVM, error)
	DestroyVM(vm *LuaVM) error

	// スクリプト実行
	LoadScript(vm *LuaVM, scriptPath string) (*LuaScript, error)
	UnloadScript(vm *LuaVM, script *LuaScript) error
	ExecuteScript(vm *LuaVM, script *LuaScript) error

	// データ変換
	GoToLua(vm *LuaVM, value interface{}) (lua.LValue, error)
	LuaToGo(vm *LuaVM, value lua.LValue, target interface{}) error

	// API登録
	RegisterECSAPI(vm *LuaVM, ecsAPI *ModECSAPI) error
	SetPermissions(vm *LuaVM, permissions *APIPermissions) error
}

// LuaVM - Lua仮想マシンラッパー
type LuaVM struct {
	state        *lua.LState
	sandbox      *Sandbox
	permissions  *APIPermissions
	resources    *ResourceLimits
	errorHandler ErrorHandler
}

// LuaVMConfig - Lua VM設定
type LuaVMConfig struct {
	SandboxEnabled bool
	ResourceLimits *ResourceLimits
	Permissions    *APIPermissions
}

// LuaScript - Luaスクリプト管理
type LuaScript struct {
	path     string
	content  []byte
	loaded   bool
	metadata *ScriptMetadata
}

// APIPermissions - API権限管理
type APIPermissions struct {
	AllowedAPIs    []string
	ForbiddenAPIs  []string
	ResourceLimits *ResourceLimits
}

// ResourceLimits - リソース制限
type ResourceLimits struct {
	MaxExecutionTime time.Duration // デフォルト: 100ms
	MaxMemoryUsage   int64         // デフォルト: 10MB
	MaxFileAccess    bool          // デフォルト: false
	MaxNetworkAccess bool          // デフォルト: false
}

// Sandbox - サンドボックス制御
type Sandbox struct {
	FileSystemRestricted bool
	NetworkRestricted    bool
	OSCommandsBlocked    bool
}

// ScriptMetadata - スクリプトメタデータ
type ScriptMetadata struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Author       string   `json:"author"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"`
	Permissions  []string `json:"permissions"`
	APIVersion   string   `json:"api_version"`
	EntryPoint   string   `json:"entry_point"`
}

// ModECSAPI - MOD向けECS API制限インターフェース
type ModECSAPI interface {
	// EntityManager操作（制限版）
	CreateEntity() (EntityID, error)
	DestroyEntity(id EntityID) error
	EntityExists(id EntityID) bool

	// ComponentStore操作（制限版）
	AddComponent(entityID EntityID, componentType string, data interface{}) error
	RemoveComponent(entityID EntityID, componentType string) error
	GetComponent(entityID EntityID, componentType string) (interface{}, error)
	HasComponent(entityID EntityID, componentType string) bool

	// Query操作（制限版）
	QueryEntities() QueryBuilder

	// Event操作（制限版）
	FireEvent(eventType string, data interface{}) error
	SubscribeEvent(eventType string, callback func(interface{})) error
}

// EntityID - エンティティID型定義
type EntityID uint64

// QueryBuilder - クエリビルダーインターフェース
type QueryBuilder interface {
	With(componentType string) QueryBuilder
	Without(componentType string) QueryBuilder
	Execute() ([]EntityID, error)
}

// ErrorHandler - エラーハンドラー関数型
type ErrorHandler func(error) error
