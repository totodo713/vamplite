package mod

import (
	"time"

	"muscle-dreamer/internal/core/ecs"
)

// ModECSAPI はMOD向けの制限されたECS APIのメインインターフェース
type ModECSAPI interface {
	Entities() ModEntityAPI
	Components() ModComponentAPI
	Queries() ModQueryAPI
	Systems() ModSystemAPI
	GetContext() *ModContext
}

// ModEntityAPI はMOD向けの制限されたエンティティ操作API
type ModEntityAPI interface {
	Create(tags ...string) (ecs.EntityID, error)
	Delete(id ecs.EntityID) error
	GetTags(id ecs.EntityID) ([]string, error)
	GetOwned() ([]ecs.EntityID, error)
}

// ModComponentAPI はMOD向けの制限されたコンポーネント操作API
type ModComponentAPI interface {
	Add(entity ecs.EntityID, component ecs.Component) error
	Get(entity ecs.EntityID, componentType ecs.ComponentType) (ecs.Component, error)
	Remove(entity ecs.EntityID, componentType ecs.ComponentType) error
	IsAllowed(componentType ecs.ComponentType) bool
}

// ModQueryAPI はMOD向けの制限されたクエリ操作API
type ModQueryAPI interface {
	Find(query ecs.QueryBuilder) ([]ecs.EntityID, error)
	Count(query ecs.QueryBuilder) (int, error)
	GetExecutionCount() int
	ResetExecutionCount()
}

// ModSystemAPI はMOD向けの制限されたシステム操作API
type ModSystemAPI interface {
	Register(system ModSystem) error
	Unregister(systemID string) error
	GetRegistered() []string
}

// ModSystem はMOD向けシステムインターフェース
type ModSystem interface {
	GetID() string
	Update(ctx *ModContext, deltaTime time.Duration) error
	GetMaxExecutionTime() time.Duration
}

// ModContext はMOD実行コンテキスト
type ModContext struct {
	ModID             string
	MaxEntities       int
	MaxMemory         int64
	MaxExecutionTime  time.Duration
	AllowedComponents []ecs.ComponentType
	CreatedEntities   []ecs.EntityID
	MemoryUsage       int64
	ExecutionTime     time.Duration
	QueryCount        int
	MaxQueryCount     int
}

// ModECSAPIFactory はModECSAPIの作成ファクトリー
type ModECSAPIFactory interface {
	Create(modID string, config ModConfig) (ModECSAPI, error)
	Destroy(modID string) error
}

// ModConfig はMOD設定
type ModConfig struct {
	MaxEntities       int
	MaxMemory         int64
	MaxExecutionTime  time.Duration
	AllowedComponents []ecs.ComponentType
	MaxQueryCount     int
}

// DefaultModConfig デフォルトMOD設定
func DefaultModConfig() ModConfig {
	return ModConfig{
		MaxEntities:      100,
		MaxMemory:        10 * 1024 * 1024, // 10MB
		MaxExecutionTime: 5 * time.Millisecond,
		AllowedComponents: []ecs.ComponentType{
			ecs.ComponentTypeSprite,
			ecs.ComponentTypePhysics,
			ecs.ComponentTypeHealth,
			ecs.ComponentTypeAI,
			ecs.ComponentTypeInventory,
			ecs.ComponentTypeEnergy,
		},
		MaxQueryCount: 1000,
	}
}
