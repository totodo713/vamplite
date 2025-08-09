package components

import (
	"encoding/json"
	"errors"
	"math"
	"time"

	"muscle-dreamer/internal/core/ecs"
)

// AIComponent handles NPC artificial intelligence
type AIComponent struct {
	State           AIState       `json:"state"`
	Target          ecs.EntityID  `json:"target"`
	PatrolPoints    []ecs.Vector2 `json:"patrolPoints"`
	DetectionRadius float64       `json:"detectionRadius"`
	AttackRange     float64       `json:"attackRange"`
	Speed           float64       `json:"speed"`
	Behavior        AIBehavior    `json:"behavior"`
	LastStateChange time.Time     `json:"lastStateChange"`

	// Internal state
	currentPatrolIndex int       `json:"currentPatrolIndex"`
	stateHistory       []AIState `json:"-"`
}

// NewAIComponent creates a new AI component with default values
func NewAIComponent() *AIComponent {
	return &AIComponent{
		State:              AIStateIdle,
		Target:             ecs.InvalidEntityID,
		PatrolPoints:       make([]ecs.Vector2, 0),
		DetectionRadius:    50.0,
		AttackRange:        10.0,
		Speed:              100.0,
		Behavior:           AIBehaviorNeutral,
		LastStateChange:    time.Time{},
		currentPatrolIndex: 0,
		stateHistory:       make([]AIState, 0),
	}
}

// GetType returns the component type
func (a *AIComponent) GetType() ecs.ComponentType {
	return ecs.ComponentTypeAI
}

// SetState changes the AI state
func (a *AIComponent) SetState(state AIState) {
	if a.State != state {
		a.State = state
		a.stateHistory = append(a.stateHistory, state)
		a.LastStateChange = time.Now()
	}
}

// SetTarget sets the target entity
func (a *AIComponent) SetTarget(target ecs.EntityID) {
	a.Target = target
}

// ClearTarget clears the current target
func (a *AIComponent) ClearTarget() {
	a.Target = ecs.InvalidEntityID
}

// SetPatrolPoints sets the patrol points
func (a *AIComponent) SetPatrolPoints(points []ecs.Vector2) {
	a.PatrolPoints = make([]ecs.Vector2, len(points))
	copy(a.PatrolPoints, points)
	a.currentPatrolIndex = 0
}

// GetNextPatrolPoint gets the next patrol point and advances the index
func (a *AIComponent) GetNextPatrolPoint() ecs.Vector2 {
	if len(a.PatrolPoints) == 0 {
		return ecs.Vector2{X: 0, Y: 0}
	}

	point := a.PatrolPoints[a.currentPatrolIndex]
	a.currentPatrolIndex = (a.currentPatrolIndex + 1) % len(a.PatrolPoints)
	return point
}

// SetBehavior sets the AI behavior
func (a *AIComponent) SetBehavior(behavior AIBehavior) {
	a.Behavior = behavior
}

// IsInDetectionRange checks if a target is within detection range
func (a *AIComponent) IsInDetectionRange(aiPosition, targetPosition ecs.Vector2) bool {
	distance := a.calculateDistance(aiPosition, targetPosition)
	return distance <= a.DetectionRadius
}

// IsInAttackRange checks if a target is within attack range
func (a *AIComponent) IsInAttackRange(aiPosition, targetPosition ecs.Vector2) bool {
	distance := a.calculateDistance(aiPosition, targetPosition)
	return distance <= a.AttackRange
}

// GetStateHistory returns the state history
func (a *AIComponent) GetStateHistory() []AIState {
	history := make([]AIState, len(a.stateHistory))
	copy(history, a.stateHistory)
	return history
}

// calculateDistance calculates the distance between two points
func (a *AIComponent) calculateDistance(pos1, pos2 ecs.Vector2) float64 {
	dx := pos2.X - pos1.X
	dy := pos2.Y - pos1.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// Clone creates a deep copy of the component
func (a *AIComponent) Clone() ecs.Component {
	patrolPoints := make([]ecs.Vector2, len(a.PatrolPoints))
	copy(patrolPoints, a.PatrolPoints)

	return &AIComponent{
		State:              a.State,
		Target:             a.Target,
		PatrolPoints:       patrolPoints,
		DetectionRadius:    a.DetectionRadius,
		AttackRange:        a.AttackRange,
		Speed:              a.Speed,
		Behavior:           a.Behavior,
		LastStateChange:    a.LastStateChange,
		currentPatrolIndex: a.currentPatrolIndex,
		stateHistory:       make([]AIState, 0),
	}
}

// Validate ensures the component data is valid
func (a *AIComponent) Validate() error {
	if a.DetectionRadius < 0 {
		return errors.New("detection radius cannot be negative")
	}
	if a.AttackRange < 0 {
		return errors.New("attack range cannot be negative")
	}
	if a.Speed < 0 {
		return errors.New("speed cannot be negative")
	}
	return nil
}

// Size returns the memory size of the component in bytes
func (a *AIComponent) Size() int {
	// Approximate size calculation
	baseSize := 72                               // Basic fields
	patrolPointsSize := len(a.PatrolPoints) * 16 // Vector2 size
	return baseSize + patrolPointsSize
}

// Serialize converts the component to bytes
func (a *AIComponent) Serialize() ([]byte, error) {
	// Create a serializable version without internal state
	data := struct {
		State              AIState       `json:"state"`
		Target             ecs.EntityID  `json:"target"`
		PatrolPoints       []ecs.Vector2 `json:"patrolPoints"`
		DetectionRadius    float64       `json:"detectionRadius"`
		AttackRange        float64       `json:"attackRange"`
		Speed              float64       `json:"speed"`
		Behavior           AIBehavior    `json:"behavior"`
		LastStateChange    time.Time     `json:"lastStateChange"`
		CurrentPatrolIndex int           `json:"currentPatrolIndex"`
	}{
		State:              a.State,
		Target:             a.Target,
		PatrolPoints:       a.PatrolPoints,
		DetectionRadius:    a.DetectionRadius,
		AttackRange:        a.AttackRange,
		Speed:              a.Speed,
		Behavior:           a.Behavior,
		LastStateChange:    a.LastStateChange,
		CurrentPatrolIndex: a.currentPatrolIndex,
	}
	return json.Marshal(data)
}

// Deserialize loads component data from bytes
func (a *AIComponent) Deserialize(data []byte) error {
	var serialData struct {
		State              AIState       `json:"state"`
		Target             ecs.EntityID  `json:"target"`
		PatrolPoints       []ecs.Vector2 `json:"patrolPoints"`
		DetectionRadius    float64       `json:"detectionRadius"`
		AttackRange        float64       `json:"attackRange"`
		Speed              float64       `json:"speed"`
		Behavior           AIBehavior    `json:"behavior"`
		LastStateChange    time.Time     `json:"lastStateChange"`
		CurrentPatrolIndex int           `json:"currentPatrolIndex"`
	}

	if err := json.Unmarshal(data, &serialData); err != nil {
		return err
	}

	a.State = serialData.State
	a.Target = serialData.Target
	a.PatrolPoints = serialData.PatrolPoints
	a.DetectionRadius = serialData.DetectionRadius
	a.AttackRange = serialData.AttackRange
	a.Speed = serialData.Speed
	a.Behavior = serialData.Behavior
	a.LastStateChange = serialData.LastStateChange
	a.currentPatrolIndex = serialData.CurrentPatrolIndex
	a.stateHistory = make([]AIState, 0)

	return nil
}
