package components

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"muscle-dreamer/internal/core/ecs"
)

func Test_AIComponent_CreateAndInitialize(t *testing.T) {
	// Arrange & Act
	ai := NewAIComponent()

	// Assert
	assert.Equal(t, ecs.ComponentTypeAI, ai.GetType())
	assert.Equal(t, AIStateIdle, ai.State)
	assert.Equal(t, ecs.InvalidEntityID, ai.Target)
	assert.Empty(t, ai.PatrolPoints)
	assert.Equal(t, 50.0, ai.DetectionRadius)
	assert.Equal(t, 10.0, ai.AttackRange)
	assert.Equal(t, 100.0, ai.Speed)
	assert.Equal(t, AIBehaviorNeutral, ai.Behavior)
	assert.Zero(t, ai.LastStateChange)
}

func Test_AIComponent_StateTransition(t *testing.T) {
	// Arrange
	ai := NewAIComponent()
	assert.Equal(t, AIStateIdle, ai.State)

	// Act
	ai.SetState(AIStatePatrol)

	// Assert
	assert.Equal(t, AIStatePatrol, ai.State)
	assert.NotZero(t, ai.LastStateChange)
}

func Test_AIComponent_SetTarget(t *testing.T) {
	// Arrange
	ai := NewAIComponent()
	targetID := ecs.EntityID(12345)

	// Act
	ai.SetTarget(targetID)

	// Assert
	assert.Equal(t, targetID, ai.Target)
}

func Test_AIComponent_PatrolBehavior(t *testing.T) {
	// Arrange
	ai := NewAIComponent()
	patrolPoints := []ecs.Vector2{
		{X: 0, Y: 0}, {X: 100, Y: 0}, {X: 100, Y: 100}, {X: 0, Y: 100},
	}
	ai.SetPatrolPoints(patrolPoints)

	// Act
	ai.SetState(AIStatePatrol)
	nextPoint := ai.GetNextPatrolPoint()

	// Assert
	assert.Equal(t, AIStatePatrol, ai.State)
	assert.Equal(t, patrolPoints[0], nextPoint)
}

func Test_AIComponent_PatrolPointsCycle(t *testing.T) {
	// Arrange
	ai := NewAIComponent()
	patrolPoints := []ecs.Vector2{
		{X: 0, Y: 0}, {X: 100, Y: 0}, {X: 100, Y: 100},
	}
	ai.SetPatrolPoints(patrolPoints)

	// Act & Assert - cycle through all points
	assert.Equal(t, patrolPoints[0], ai.GetNextPatrolPoint())
	assert.Equal(t, patrolPoints[1], ai.GetNextPatrolPoint())
	assert.Equal(t, patrolPoints[2], ai.GetNextPatrolPoint())
	assert.Equal(t, patrolPoints[0], ai.GetNextPatrolPoint()) // Cycles back to start
}

func Test_AIComponent_IsInDetectionRange(t *testing.T) {
	// Arrange
	ai := NewAIComponent()
	ai.DetectionRadius = 50.0
	aiPosition := ecs.Vector2{X: 0, Y: 0}

	// Test cases
	testCases := []struct {
		targetPos ecs.Vector2
		expected  bool
	}{
		{ecs.Vector2{X: 30, Y: 0}, true},   // Within range
		{ecs.Vector2{X: 0, Y: 40}, true},   // Within range
		{ecs.Vector2{X: 60, Y: 0}, false},  // Outside range
		{ecs.Vector2{X: 40, Y: 30}, true},  // Within range (diagonal)
		{ecs.Vector2{X: 50, Y: 50}, false}, // Outside range (diagonal)
	}

	for _, tc := range testCases {
		// Act
		inRange := ai.IsInDetectionRange(aiPosition, tc.targetPos)

		// Assert
		assert.Equal(t, tc.expected, inRange,
			"Position %v should be %v for detection", tc.targetPos, tc.expected)
	}
}

func Test_AIComponent_IsInAttackRange(t *testing.T) {
	// Arrange
	ai := NewAIComponent()
	ai.AttackRange = 15.0
	aiPosition := ecs.Vector2{X: 0, Y: 0}

	// Test cases
	testCases := []struct {
		targetPos ecs.Vector2
		expected  bool
	}{
		{ecs.Vector2{X: 10, Y: 0}, true},   // Within range
		{ecs.Vector2{X: 20, Y: 0}, false},  // Outside range
		{ecs.Vector2{X: 10, Y: 10}, true},  // Within range (diagonal)
		{ecs.Vector2{X: 15, Y: 15}, false}, // Outside range (diagonal)
	}

	for _, tc := range testCases {
		// Act
		inRange := ai.IsInAttackRange(aiPosition, tc.targetPos)

		// Assert
		assert.Equal(t, tc.expected, inRange,
			"Position %v should be %v for attack", tc.targetPos, tc.expected)
	}
}

func Test_AIComponent_BehaviorTypes(t *testing.T) {
	// Arrange
	ai := NewAIComponent()

	// Test all behavior types
	behaviors := []AIBehavior{
		AIBehaviorNeutral,
		AIBehaviorAggressive,
		AIBehaviorDefensive,
		AIBehaviorFriendly,
		AIBehaviorCoward,
	}

	for _, behavior := range behaviors {
		// Act
		ai.SetBehavior(behavior)

		// Assert
		assert.Equal(t, behavior, ai.Behavior)
	}
}

func Test_AIComponent_StateHistory(t *testing.T) {
	// Arrange
	ai := NewAIComponent()
	states := []AIState{AIStatePatrol, AIStateChase, AIStateAttack}

	// Act
	for _, state := range states {
		ai.SetState(state)
	}

	// Assert
	history := ai.GetStateHistory()
	assert.Len(t, history, len(states))
	for i, state := range states {
		assert.Equal(t, state, history[i])
	}
}

func Test_AIComponent_ClearTarget(t *testing.T) {
	// Arrange
	ai := NewAIComponent()
	ai.SetTarget(ecs.EntityID(12345))
	assert.NotEqual(t, ecs.InvalidEntityID, ai.Target)

	// Act
	ai.ClearTarget()

	// Assert
	assert.Equal(t, ecs.InvalidEntityID, ai.Target)
}

func Test_AIComponent_Serialization(t *testing.T) {
	// Arrange
	ai := NewAIComponent()
	ai.State = AIStateChase
	ai.Target = ecs.EntityID(54321)
	ai.DetectionRadius = 75.0
	ai.AttackRange = 20.0
	ai.Speed = 150.0
	ai.Behavior = AIBehaviorAggressive
	ai.SetPatrolPoints([]ecs.Vector2{{X: 10, Y: 20}, {X: 30, Y: 40}})

	// Act
	data, err := ai.Serialize()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Debug: Print the serialized data
	t.Logf("Serialized data: %s", string(data))

	// Create new component and deserialize
	newAI := NewAIComponent()
	err = newAI.Deserialize(data)
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, ai.State, newAI.State)
	assert.Equal(t, ai.Target, newAI.Target)
	assert.Equal(t, ai.DetectionRadius, newAI.DetectionRadius)
	assert.Equal(t, ai.AttackRange, newAI.AttackRange)
	assert.Equal(t, ai.Speed, newAI.Speed)
	assert.Equal(t, ai.Behavior, newAI.Behavior)
	assert.Equal(t, ai.PatrolPoints, newAI.PatrolPoints)
}

func Test_AIComponent_Clone(t *testing.T) {
	// Arrange
	original := NewAIComponent()
	original.State = AIStatePatrol
	original.DetectionRadius = 80.0
	original.Behavior = AIBehaviorDefensive

	// Act
	cloned := original.Clone()

	// Assert
	assert.NotSame(t, original, cloned)
	clonedAI := cloned.(*AIComponent)
	assert.Equal(t, original.State, clonedAI.State)
	assert.Equal(t, original.DetectionRadius, clonedAI.DetectionRadius)
	assert.Equal(t, original.Behavior, clonedAI.Behavior)
}

func Test_AIComponent_Validate(t *testing.T) {
	// Arrange
	ai := NewAIComponent()

	// Act & Assert - valid state
	err := ai.Validate()
	assert.NoError(t, err)

	// Invalid detection radius (negative)
	ai.DetectionRadius = -10.0
	err = ai.Validate()
	assert.Error(t, err)

	// Invalid attack range (negative)
	ai.DetectionRadius = 50.0
	ai.AttackRange = -5.0
	err = ai.Validate()
	assert.Error(t, err)

	// Invalid speed (negative)
	ai.AttackRange = 10.0
	ai.Speed = -50.0
	err = ai.Validate()
	assert.Error(t, err)
}

func Test_AIComponent_Size(t *testing.T) {
	// Arrange
	ai := NewAIComponent()

	// Act
	size := ai.Size()

	// Assert
	assert.LessOrEqual(t, size, 96, "AIComponent size should be <= 96 bytes")
	assert.Greater(t, size, 0, "AIComponent size should be > 0")
}

func Benchmark_AIComponent_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewAIComponent()
	}
}

func Benchmark_AIComponent_StateTransition(b *testing.B) {
	ai := NewAIComponent()
	states := []AIState{AIStateIdle, AIStatePatrol, AIStateChase, AIStateAttack}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ai.SetState(states[i%len(states)])
	}
}

func Benchmark_AIComponent_RangeCheck(b *testing.B) {
	ai := NewAIComponent()
	aiPos := ecs.Vector2{X: 0, Y: 0}
	targetPos := ecs.Vector2{X: 30, Y: 40}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ai.IsInDetectionRange(aiPos, targetPos)
	}
}
