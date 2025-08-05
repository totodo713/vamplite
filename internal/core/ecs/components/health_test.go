package components

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"muscle-dreamer/internal/core/ecs"
)

func Test_HealthComponent_CreateAndInitialize(t *testing.T) {
	// Arrange & Act
	health := NewHealthComponent(100)

	// Assert
	assert.Equal(t, ecs.ComponentTypeHealth, health.GetType())
	assert.Equal(t, 100, health.CurrentHealth)
	assert.Equal(t, 100, health.MaxHealth)
	assert.Equal(t, 0, health.Shield)
	assert.False(t, health.IsInvincible)
	assert.Zero(t, health.LastDamageTime)
	assert.Equal(t, 0.0, health.RegenerationRate)
	assert.Empty(t, health.StatusEffects)
}

func Test_HealthComponent_TakeDamage(t *testing.T) {
	// Arrange
	health := NewHealthComponent(100)
	damage := 25

	// Act
	actualDamage := health.TakeDamage(damage)

	// Assert
	assert.Equal(t, 75, health.CurrentHealth)
	assert.Equal(t, 25, actualDamage)
	assert.NotZero(t, health.LastDamageTime)
}

func Test_HealthComponent_TakeDamageWithShield(t *testing.T) {
	// Arrange
	health := NewHealthComponent(100)
	health.Shield = 30
	damage := 50

	// Act
	actualDamage := health.TakeDamage(damage)

	// Assert
	assert.Equal(t, 80, health.CurrentHealth) // 50 - 30 shield = 20 damage
	assert.Equal(t, 0, health.Shield)
	assert.Equal(t, 20, actualDamage)
}

func Test_HealthComponent_TakeDamageInvincible(t *testing.T) {
	// Arrange
	health := NewHealthComponent(100)
	health.IsInvincible = true
	damage := 50

	// Act
	actualDamage := health.TakeDamage(damage)

	// Assert
	assert.Equal(t, 100, health.CurrentHealth) // No damage taken
	assert.Equal(t, 0, actualDamage)
}

func Test_HealthComponent_TakeDamageExceedsHealth(t *testing.T) {
	// Arrange
	health := NewHealthComponent(100)
	health.CurrentHealth = 30
	damage := 50

	// Act
	actualDamage := health.TakeDamage(damage)

	// Assert
	assert.Equal(t, 0, health.CurrentHealth) // Clamped to 0
	assert.Equal(t, 30, actualDamage)        // Only damaged remaining health
}

func Test_HealthComponent_Heal(t *testing.T) {
	// Arrange
	health := NewHealthComponent(100)
	health.CurrentHealth = 50
	healAmount := 30

	// Act
	actualHeal := health.Heal(healAmount)

	// Assert
	assert.Equal(t, 80, health.CurrentHealth)
	assert.Equal(t, 30, actualHeal)
}

func Test_HealthComponent_HealExceedsMax(t *testing.T) {
	// Arrange
	health := NewHealthComponent(100)
	health.CurrentHealth = 90
	healAmount := 30

	// Act
	actualHeal := health.Heal(healAmount)

	// Assert
	assert.Equal(t, 100, health.CurrentHealth) // Clamped to max
	assert.Equal(t, 10, actualHeal)            // Only healed remaining amount
}

func Test_HealthComponent_Regeneration(t *testing.T) {
	// Arrange
	health := NewHealthComponent(100)
	health.CurrentHealth = 50
	health.RegenerationRate = 5.0 // 5 HP per second
	deltaTime := 0.5              // 0.5 seconds

	// Act
	health.UpdateRegeneration(deltaTime)

	// Assert
	// Expected: 50 + (5.0 * 0.5) = 52.5, but truncated to int = 52
	expectedHealth := 52
	assert.Equal(t, expectedHealth, health.CurrentHealth)
}

func Test_HealthComponent_IsDead(t *testing.T) {
	// Arrange
	health := NewHealthComponent(100)

	// Assert - alive
	assert.False(t, health.IsDead())

	// Act - take fatal damage
	health.TakeDamage(100)

	// Assert - dead
	assert.True(t, health.IsDead())
}

func Test_HealthComponent_AddStatusEffect(t *testing.T) {
	// Arrange
	health := NewHealthComponent(100)
	effect := StatusEffect{
		Type:     StatusTypePoison,
		Duration: 5.0,
		Strength: 2.0,
	}

	// Act
	health.AddStatusEffect(effect)

	// Assert
	assert.Len(t, health.StatusEffects, 1)
	addedEffect := health.StatusEffects[0]
	assert.Equal(t, effect.Type, addedEffect.Type)
	assert.Equal(t, effect.Duration, addedEffect.Duration)
	assert.Equal(t, effect.Strength, addedEffect.Strength)
	assert.False(t, addedEffect.StartTime.IsZero()) // StartTime should be set
}

func Test_HealthComponent_RemoveStatusEffect(t *testing.T) {
	// Arrange
	health := NewHealthComponent(100)
	effect1 := StatusEffect{Type: StatusTypePoison, Duration: 5.0}
	effect2 := StatusEffect{Type: StatusTypeBurn, Duration: 3.0}
	health.AddStatusEffect(effect1)
	health.AddStatusEffect(effect2)

	// Act
	health.RemoveStatusEffect(StatusTypePoison)

	// Assert
	assert.Len(t, health.StatusEffects, 1)
	assert.Equal(t, StatusTypeBurn, health.StatusEffects[0].Type)
}

func Test_HealthComponent_UpdateStatusEffects(t *testing.T) {
	// Arrange
	health := NewHealthComponent(100)
	effect := StatusEffect{
		Type:     StatusTypePoison,
		Duration: 1.0,
		Strength: 5.0,
	}
	health.AddStatusEffect(effect)
	deltaTime := 0.5

	// Act
	health.UpdateStatusEffects(deltaTime)

	// Assert
	assert.Len(t, health.StatusEffects, 1)
	assert.InDelta(t, 0.5, health.StatusEffects[0].Duration, 0.001)
}

func Test_HealthComponent_Serialization(t *testing.T) {
	// Arrange
	health := NewHealthComponent(100)
	health.CurrentHealth = 75
	health.Shield = 25
	health.RegenerationRate = 2.5
	health.IsInvincible = true

	// Act
	data, err := health.Serialize()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Create new component and deserialize
	newHealth := NewHealthComponent(100)
	err = newHealth.Deserialize(data)
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, health.CurrentHealth, newHealth.CurrentHealth)
	assert.Equal(t, health.MaxHealth, newHealth.MaxHealth)
	assert.Equal(t, health.Shield, newHealth.Shield)
	assert.Equal(t, health.RegenerationRate, newHealth.RegenerationRate)
	assert.Equal(t, health.IsInvincible, newHealth.IsInvincible)
}

func Test_HealthComponent_Clone(t *testing.T) {
	// Arrange
	original := NewHealthComponent(100)
	original.CurrentHealth = 80
	original.Shield = 20

	// Act
	cloned := original.Clone()

	// Assert
	assert.NotSame(t, original, cloned)
	clonedHealth := cloned.(*HealthComponent)
	assert.Equal(t, original.CurrentHealth, clonedHealth.CurrentHealth)
	assert.Equal(t, original.MaxHealth, clonedHealth.MaxHealth)
	assert.Equal(t, original.Shield, clonedHealth.Shield)
}

func Test_HealthComponent_Validate(t *testing.T) {
	// Arrange
	health := NewHealthComponent(100)

	// Act & Assert - valid state
	err := health.Validate()
	assert.NoError(t, err)

	// Invalid current health (negative)
	health.CurrentHealth = -10
	err = health.Validate()
	assert.Error(t, err)

	// Invalid max health (zero)
	health.CurrentHealth = 50
	health.MaxHealth = 0
	err = health.Validate()
	assert.Error(t, err)
}

func Test_HealthComponent_HasStatusEffect(t *testing.T) {
	// Arrange
	health := NewHealthComponent(100)
	poisonEffect := StatusEffect{
		Type:     StatusTypePoison,
		Duration: 5.0,
		Strength: 10.0,
	}

	// Act & Assert - No effect initially
	assert.False(t, health.HasStatusEffect(StatusTypePoison))

	// Add effect
	health.AddStatusEffect(poisonEffect)
	assert.True(t, health.HasStatusEffect(StatusTypePoison))
	assert.False(t, health.HasStatusEffect(StatusTypeBurn))
}

func Test_HealthComponent_Size(t *testing.T) {
	// Arrange
	health := NewHealthComponent(100)

	// Act
	size := health.Size()

	// Assert
	assert.LessOrEqual(t, size, 72, "HealthComponent size should be <= 72 bytes")
	assert.Greater(t, size, 0, "HealthComponent size should be > 0")
}

func Benchmark_HealthComponent_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewHealthComponent(100)
	}
}

func Benchmark_HealthComponent_TakeDamage(b *testing.B) {
	health := NewHealthComponent(1000000) // Large health for sustained testing

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		health.TakeDamage(1)
	}
}
