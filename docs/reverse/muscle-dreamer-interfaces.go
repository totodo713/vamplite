// ==============================================
// Muscle Dreamer - 型定義・インターフェース集約
// 既存コードから逆生成 (2025-08-03)
// ==============================================

package interfaces

import (
	"image/color"
	"time"
	
	"github.com/hajimehoshi/ebiten/v2"
)

// ==============================================
// 1. ゲームコア型定義
// ==============================================

// GameEngine - ゲームエンジンの基本インターフェース
type GameEngine interface {
	Update() error
	Draw(screen *ebiten.Image)
	Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int)
	Run() error
}

// Game - 実装済みゲーム構造体
type Game struct {
	// 将来的にECSコンポーネントマネージャーを格納
	entities EntityManager
	systems  SystemManager

	// ゲーム状態管理
	state  GameState
	config *GameConfig

	// アセット管理
	assets AssetManager
	themes ThemeManager
	mods   ModManager

	// 描画・オーディオ
	renderer Renderer
	audio    AudioManager

	// 入力管理
	input InputManager

	// パフォーマンス監視
	metrics *PerformanceMetrics
}

// GameState - ゲーム状態列挙
type GameState int

const (
	GameStateMenu GameState = iota
	GameStateLoading
	GameStatePlaying
	GameStatePaused
	GameStateGameOver
	GameStateSettings
)

// ==============================================
// 2. 設定・コンフィグ型定義
// ==============================================

// GameConfig - game.yamlから読み込まれる設定
type GameConfig struct {
	Game     GameSettings     `yaml:"game"`
	Graphics GraphicsSettings `yaml:"graphics"`
	Audio    AudioSettings    `yaml:"audio"`
	Input    InputSettings    `yaml:"input"`
}

// GameSettings - ゲーム基本設定
type GameSettings struct {
	Title   string `yaml:"title"`
	Version string `yaml:"version"`
}

// GraphicsSettings - 描画設定
type GraphicsSettings struct {
	Width      int  `yaml:"width"`
	Height     int  `yaml:"height"`
	Fullscreen bool `yaml:"fullscreen"`
	VSync      bool `yaml:"vsync"`
}

// AudioSettings - オーディオ設定
type AudioSettings struct {
	MasterVolume float64 `yaml:"master_volume"`
	BGMVolume    float64 `yaml:"bgm_volume"`
	SFXVolume    float64 `yaml:"sfx_volume"`
}

// InputSettings - 入力設定
type InputSettings struct {
	KeyboardEnabled bool `yaml:"keyboard_enabled"`
	MouseEnabled    bool `yaml:"mouse_enabled"`
	GamepadEnabled  bool `yaml:"gamepad_enabled"`
}

// ==============================================
// 3. ECS (Entity Component System) インターフェース
// ==============================================

// EntityManager - エンティティ管理インターフェース
type EntityManager interface {
	CreateEntity() EntityID
	DestroyEntity(EntityID)
	GetComponent(EntityID, ComponentType) Component
	AddComponent(EntityID, Component)
	RemoveComponent(EntityID, ComponentType)
	HasComponent(EntityID, ComponentType) bool
	GetEntitiesWith(ComponentType) []EntityID
}

// SystemManager - システム管理インターフェース
type SystemManager interface {
	RegisterSystem(System)
	UnregisterSystem(SystemType)
	UpdateSystems(deltaTime float64) error
	RenderSystems(screen *ebiten.Image)
}

// System - ゲームシステム基底インターフェース
type System interface {
	Update(deltaTime float64, entities EntityManager) error
	GetType() SystemType
}

// Component - コンポーネント基底インターフェース
type Component interface {
	GetType() ComponentType
}

// EntityID - エンティティ識別子
type EntityID uint64

// ComponentType - コンポーネント種別
type ComponentType string

// SystemType - システム種別
type SystemType string

// 基本コンポーネント型定義
const (
	ComponentTransform ComponentType = "transform"
	ComponentSprite    ComponentType = "sprite"
	ComponentRigidBody ComponentType = "rigidbody"
	ComponentHealth    ComponentType = "health"
	ComponentInventory ComponentType = "inventory"
	ComponentAI        ComponentType = "ai"
)

// 基本システム型定義
const (
	SystemMovement  SystemType = "movement"
	SystemRendering SystemType = "rendering"
	SystemPhysics   SystemType = "physics"
	SystemAudio     SystemType = "audio"
	SystemInput     SystemType = "input"
	SystemGameplay  SystemType = "gameplay"
)

// ==============================================
// 4. 具体的コンポーネント型定義
// ==============================================

// TransformComponent - 位置・回転・スケール
type TransformComponent struct {
	X, Y     float64
	Rotation float64
	ScaleX   float64
	ScaleY   float64
}

func (t TransformComponent) GetType() ComponentType {
	return ComponentTransform
}

// SpriteComponent - スプライト描画
type SpriteComponent struct {
	Image  *ebiten.Image
	Width  int
	Height int
	Color  color.Color
}

func (s SpriteComponent) GetType() ComponentType {
	return ComponentSprite
}

// HealthComponent - HP管理
type HealthComponent struct {
	Current int
	Maximum int
}

func (h HealthComponent) GetType() ComponentType {
	return ComponentHealth
}

// ==============================================
// 5. アセット管理型定義
// ==============================================

// AssetManager - アセット管理インターフェース
type AssetManager interface {
	LoadImage(path string) (*ebiten.Image, error)
	LoadAudio(path string) (AudioClip, error)
	LoadFont(path string) (Font, error)
	UnloadAsset(path string)
	GetLoadedAssets() map[string]Asset
}

// Asset - アセット基底インターフェース
type Asset interface {
	GetPath() string
	GetSize() int64
	IsLoaded() bool
}

// AudioClip - オーディオアセット
type AudioClip interface {
	Asset
	Play() error
	Stop() error
	SetVolume(float64)
	GetDuration() time.Duration
}

// Font - フォントアセット
type Font interface {
	Asset
	RenderText(text string, size int) *ebiten.Image
}

// ==============================================
// 6. テーマシステム型定義
// ==============================================

// ThemeManager - テーマ管理インターフェース
type ThemeManager interface {
	LoadTheme(themeName string) (*Theme, error)
	GetCurrentTheme() *Theme
	SetTheme(themeName string) error
	GetAvailableThemes() []string
}

// Theme - テーマ定義
type Theme struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Description  string            `yaml:"description"`
	Author       string            `yaml:"author"`
	Assets       map[string]string `yaml:"assets"`
	Localization map[string]string `yaml:"localization"`
	Scripts      []string          `yaml:"scripts"`
}

// ==============================================
// 7. MODシステム型定義
// ==============================================

// ModManager - MOD管理インターフェース
type ModManager interface {
	LoadMod(modPath string) (*Mod, error)
	EnableMod(modName string) error
	DisableMod(modName string) error
	GetEnabledMods() []*Mod
	ValidateMod(modPath string) error
}

// Mod - MOD定義
type Mod struct {
	Name         string         `yaml:"name"`
	Version      string         `yaml:"version"`
	Description  string         `yaml:"description"`
	Author       string         `yaml:"author"`
	Dependencies []string       `yaml:"dependencies"`
	Scripts      []string       `yaml:"scripts"`
	Permissions  ModPermissions `yaml:"permissions"`
}

// ModPermissions - MODセキュリティ権限
type ModPermissions struct {
	FileAccess    []string `yaml:"file_access"`
	NetworkAccess bool     `yaml:"network_access"`
	SystemAccess  bool     `yaml:"system_access"`
}

// ==============================================
// 8. 描画・レンダリング型定義
// ==============================================

// Renderer - 描画インターフェース
type Renderer interface {
	DrawSprite(sprite *SpriteComponent, transform *TransformComponent, screen *ebiten.Image)
	DrawText(text string, x, y int, font Font, screen *ebiten.Image)
	DrawDebugInfo(info string, screen *ebiten.Image)
	SetBackgroundColor(color color.Color)
	Clear(screen *ebiten.Image)
}

// ==============================================
// 9. オーディオ管理型定義
// ==============================================

// AudioManager - オーディオ管理インターフェース
type AudioManager interface {
	PlayBGM(clip AudioClip) error
	PlaySFX(clip AudioClip) error
	StopBGM() error
	StopAllSFX() error
	SetMasterVolume(volume float64)
	SetBGMVolume(volume float64)
	SetSFXVolume(volume float64)
}

// ==============================================
// 10. 入力管理型定義
// ==============================================

// InputManager - 入力管理インターフェース
type InputManager interface {
	IsKeyPressed(key ebiten.Key) bool
	IsKeyJustPressed(key ebiten.Key) bool
	IsMousePressed(button ebiten.MouseButton) bool
	GetMousePosition() (int, int)
	IsGamepadConnected(id ebiten.GamepadID) bool
	GetGamepadAxisValue(id ebiten.GamepadID, axis int) float64
}

// InputState - 入力状態
type InputState struct {
	Keys     map[ebiten.Key]bool
	Mouse    MouseState
	Gamepads map[ebiten.GamepadID]GamepadState
}

// MouseState - マウス状態
type MouseState struct {
	X, Y    int
	Buttons map[ebiten.MouseButton]bool
}

// GamepadState - ゲームパッド状態
type GamepadState struct {
	Buttons map[ebiten.StandardGamepadButton]bool
	Axes    map[int]float64
}

// ==============================================
// 11. セーブ・ロード型定義
// ==============================================

// SaveManager - セーブデータ管理インターフェース
type SaveManager interface {
	SaveGame(slot int, data *SaveData) error
	LoadGame(slot int) (*SaveData, error)
	DeleteSave(slot int) error
	GetSaveSlots() []SaveSlotInfo
}

// SaveData - セーブデータ
type SaveData struct {
	Version    string                 `json:"version"`
	Timestamp  time.Time              `json:"timestamp"`
	PlayerData PlayerData             `json:"player_data"`
	WorldState WorldState             `json:"world_state"`
	Settings   GameConfig             `json:"settings"`
	CustomData map[string]interface{} `json:"custom_data"`
}

// SaveSlotInfo - セーブスロット情報
type SaveSlotInfo struct {
	Slot      int       `json:"slot"`
	Exists    bool      `json:"exists"`
	Timestamp time.Time `json:"timestamp"`
	Preview   string    `json:"preview"`
}

// PlayerData - プレイヤーデータ
type PlayerData struct {
	Name       string  `json:"name"`
	Level      int     `json:"level"`
	Experience int64   `json:"experience"`
	Health     int     `json:"health"`
	Position   Vector2 `json:"position"`
}

// WorldState - ワールド状態
type WorldState struct {
	CurrentLevel    string                 `json:"current_level"`
	CompletedLevels []string               `json:"completed_levels"`
	Inventory       []InventoryItem        `json:"inventory"`
	Flags           map[string]interface{} `json:"flags"`
}

// InventoryItem - インベントリアイテム
type InventoryItem struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
}

// ==============================================
// 12. パフォーマンス監視型定義
// ==============================================

// PerformanceMetrics - パフォーマンス指標
type PerformanceMetrics struct {
	FPS              float64       `json:"fps"`
	FrameTime        time.Duration `json:"frame_time"`
	UpdateTime       time.Duration `json:"update_time"`
	DrawTime         time.Duration `json:"draw_time"`
	MemoryUsage      int64         `json:"memory_usage"`
	EntityCount      int           `json:"entity_count"`
	SystemCount      int           `json:"system_count"`
	ActiveComponents int           `json:"active_components"`
}

// ==============================================
// 13. ユーティリティ型定義
// ==============================================

// Vector2 - 2D座標
type Vector2 struct {
	X, Y float64
}

// Rectangle - 矩形
type Rectangle struct {
	X, Y, Width, Height float64
}

// Circle - 円形
type Circle struct {
	X, Y, Radius float64
}

// Color - RGBA色定義
type Color struct {
	R, G, B, A uint8
}

// ==============================================
// 14. エラー型定義
// ==============================================

// GameError - ゲームエラーインターフェース
type GameError interface {
	error
	GetCode() string
	GetSeverity() ErrorSeverity
}

// ErrorSeverity - エラー重要度
type ErrorSeverity int

const (
	ErrorSeverityInfo ErrorSeverity = iota
	ErrorSeverityWarning
	ErrorSeverityError
	ErrorSeverityFatal
)

// ConfigError - 設定エラー
type ConfigError struct {
	Code    string
	Message string
	Field   string
}

func (e ConfigError) Error() string {
	return e.Message
}

func (e ConfigError) GetCode() string {
	return e.Code
}

func (e ConfigError) GetSeverity() ErrorSeverity {
	return ErrorSeverityError
}

// AssetError - アセットエラー
type AssetError struct {
	Code      string
	Message   string
	AssetPath string
}

func (e AssetError) Error() string {
	return e.Message
}

func (e AssetError) GetCode() string {
	return e.Code
}

func (e AssetError) GetSeverity() ErrorSeverity {
	return ErrorSeverityWarning
}

// ==============================================
// 15. WebAssembly専用型定義
// ==============================================

// WebAssemblyBridge - WebAssembly連携インターフェース
type WebAssemblyBridge interface {
	CallJavaScript(functionName string, args ...interface{}) (interface{}, error)
	RegisterCallback(name string, callback func(...interface{}) interface{})
	GetBrowserInfo() BrowserInfo
}

// BrowserInfo - ブラウザ情報
type BrowserInfo struct {
	UserAgent    string
	Language     string
	Platform     string
	ScreenWidth  int
	ScreenHeight int
	Supports     BrowserSupport
}

// BrowserSupport - ブラウザサポート機能
type BrowserSupport struct {
	WebAssembly  bool
	WebGL        bool
	AudioContext bool
	Fullscreen   bool
	LocalStorage bool
	IndexedDB    bool
}

// ==============================================
// 16. 定数定義
// ==============================================

const (
	// ゲーム定数
	DefaultScreenWidth  = 1280
	DefaultScreenHeight = 720
	TargetFPS           = 60
	MaxEntities         = 10000

	// ファイルパス
	ConfigPath = "config/game.yaml"
	AssetPath  = "assets/"
	ThemePath  = "themes/"
	ModPath    = "mods/"
	SavePath   = "saves/"

	// Web関連
	WebPort             = 3000
	HealthCheckEndpoint = "/health"
	GameWasmFile        = "game.wasm"
	WasmExecFile        = "wasm_exec.js"
)

// ==============================================
// 17. 初期化関数型定義
// ==============================================

// InitOptions - 初期化オプション
type InitOptions struct {
	ConfigPath    string
	AssetPath     string
	EnableDebug   bool
	EnableMetrics bool
	WebMode       bool
}

// GameFactory - ゲームファクトリ関数型
type GameFactory func(options InitOptions) (GameEngine, error)

// ==============================================
// 実装例コメント
// ==============================================

/*
使用例:

// ゲーム初期化
game := &Game{
    entities: NewEntityManager(),
    systems:  NewSystemManager(),
    state:    GameStateMenu,
    config:   LoadConfig("config/game.yaml"),
}

// ECS使用例
playerEntity := game.entities.CreateEntity()
game.entities.AddComponent(playerEntity, &TransformComponent{
    X: 100, Y: 100, Rotation: 0, ScaleX: 1, ScaleY: 1,
})
game.entities.AddComponent(playerEntity, &SpriteComponent{
    Image: assets.LoadImage("player.png"),
    Width: 32, Height: 32,
})

// システム登録
game.systems.RegisterSystem(&MovementSystem{})
game.systems.RegisterSystem(&RenderingSystem{})

// ゲームループ
for {
    game.systems.UpdateSystems(deltaTime)
    game.systems.RenderSystems(screen)
}
*/
