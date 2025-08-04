package components

import (
	"time"
)

// AI state types
type AIState int

const (
	AIStateIdle AIState = iota
	AIStatePatrol
	AIStateChase
	AIStateAttack
	AIStateFlee
	AIStateDead
)

// AI behavior types
type AIBehavior int

const (
	AIBehaviorNeutral AIBehavior = iota
	AIBehaviorAggressive
	AIBehaviorDefensive
	AIBehaviorFriendly
	AIBehaviorCoward
)

// Status effect types
type StatusType int

const (
	StatusTypePoison StatusType = iota
	StatusTypeBurn
	StatusTypeFreeze
	StatusTypeStun
	StatusTypeShield
	StatusTypeRegen
)

// StatusEffect represents a temporary effect on an entity
type StatusEffect struct {
	Type      StatusType `json:"type"`
	Duration  float64    `json:"duration"`
	Strength  float64    `json:"strength"`
	StartTime time.Time  `json:"start_time"`
}

// TransformMatrix represents a 3x3 transformation matrix
type TransformMatrix [9]float64
