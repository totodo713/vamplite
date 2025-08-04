package components

import (
	"encoding/json"
	"errors"
	"time"

	"muscle-dreamer/internal/core/ecs"
)

// HealthComponent handles entity health and status effects
type HealthComponent struct {
	CurrentHealth    int            `json:"current_health"`
	MaxHealth        int            `json:"max_health"`
	Shield           int            `json:"shield"`
	IsInvincible     bool           `json:"is_invincible"`
	LastDamageTime   time.Time      `json:"last_damage_time"`
	RegenerationRate float64        `json:"regeneration_rate"`
	StatusEffects    []StatusEffect `json:"status_effects"`
}

// NewHealthComponent creates a new health component with the specified max health
func NewHealthComponent(maxHealth int) *HealthComponent {
	return &HealthComponent{
		CurrentHealth:    maxHealth,
		MaxHealth:        maxHealth,
		Shield:           0,
		IsInvincible:     false,
		LastDamageTime:   time.Time{},
		RegenerationRate: 0.0,
		StatusEffects:    make([]StatusEffect, 0),
	}
}

// GetType returns the component type
func (h *HealthComponent) GetType() ecs.ComponentType {
	return ecs.ComponentTypeHealth
}

// TakeDamage applies damage to the entity and returns actual damage dealt
func (h *HealthComponent) TakeDamage(damage int) int {
	if h.IsInvincible || damage <= 0 {
		return 0
	}

	actualDamage := damage

	// Apply shield first
	if h.Shield > 0 {
		if h.Shield >= damage {
			h.Shield -= damage
			return 0
		} else {
			actualDamage = damage - h.Shield
			h.Shield = 0
		}
	}

	// Apply remaining damage to health
	if h.CurrentHealth < actualDamage {
		actualDamage = h.CurrentHealth
	}

	h.CurrentHealth -= actualDamage
	if h.CurrentHealth < 0 {
		h.CurrentHealth = 0
	}

	h.LastDamageTime = time.Now()
	return actualDamage
}

// Heal restores health and returns actual amount healed
func (h *HealthComponent) Heal(amount int) int {
	if amount <= 0 {
		return 0
	}

	actualHeal := amount
	if h.CurrentHealth+amount > h.MaxHealth {
		actualHeal = h.MaxHealth - h.CurrentHealth
	}

	h.CurrentHealth += actualHeal
	return actualHeal
}

// UpdateRegeneration applies health regeneration over time
func (h *HealthComponent) UpdateRegeneration(deltaTime float64) {
	if h.RegenerationRate <= 0 || h.CurrentHealth >= h.MaxHealth {
		return
	}

	regenAmount := h.RegenerationRate * deltaTime
	newHealth := float64(h.CurrentHealth) + regenAmount

	if newHealth > float64(h.MaxHealth) {
		newHealth = float64(h.MaxHealth)
	}

	h.CurrentHealth = int(newHealth)
}

// IsDead returns true if current health is 0 or less
func (h *HealthComponent) IsDead() bool {
	return h.CurrentHealth <= 0
}

// AddStatusEffect adds a status effect
func (h *HealthComponent) AddStatusEffect(effect StatusEffect) {
	// Check if effect already exists and update it
	for i, existing := range h.StatusEffects {
		if existing.Type == effect.Type {
			h.StatusEffects[i] = effect
			return
		}
	}

	// Add new effect
	effect.StartTime = time.Now()
	h.StatusEffects = append(h.StatusEffects, effect)
}

// RemoveStatusEffect removes a status effect by type
func (h *HealthComponent) RemoveStatusEffect(effectType StatusType) {
	for i, effect := range h.StatusEffects {
		if effect.Type == effectType {
			h.StatusEffects = append(h.StatusEffects[:i], h.StatusEffects[i+1:]...)
			return
		}
	}
}

// UpdateStatusEffects updates all status effects and removes expired ones
func (h *HealthComponent) UpdateStatusEffects(deltaTime float64) {
	remaining := make([]StatusEffect, 0, len(h.StatusEffects))

	for _, effect := range h.StatusEffects {
		effect.Duration -= deltaTime
		if effect.Duration > 0 {
			remaining = append(remaining, effect)
		}
	}

	h.StatusEffects = remaining
}

// HasStatusEffect checks if a specific status effect is active
func (h *HealthComponent) HasStatusEffect(effectType StatusType) bool {
	for _, effect := range h.StatusEffects {
		if effect.Type == effectType {
			return true
		}
	}
	return false
}

// Clone creates a deep copy of the component
func (h *HealthComponent) Clone() ecs.Component {
	statusEffects := make([]StatusEffect, len(h.StatusEffects))
	copy(statusEffects, h.StatusEffects)

	return &HealthComponent{
		CurrentHealth:    h.CurrentHealth,
		MaxHealth:        h.MaxHealth,
		Shield:           h.Shield,
		IsInvincible:     h.IsInvincible,
		LastDamageTime:   h.LastDamageTime,
		RegenerationRate: h.RegenerationRate,
		StatusEffects:    statusEffects,
	}
}

// Validate ensures the component data is valid
func (h *HealthComponent) Validate() error {
	if h.CurrentHealth < 0 {
		return errors.New("current health cannot be negative")
	}
	if h.MaxHealth <= 0 {
		return errors.New("max health must be positive")
	}
	if h.Shield < 0 {
		return errors.New("shield cannot be negative")
	}
	if h.RegenerationRate < 0 {
		return errors.New("regeneration rate cannot be negative")
	}
	return nil
}

// Size returns the memory size of the component in bytes
func (h *HealthComponent) Size() int {
	// Approximate size calculation
	baseSize := 64                           // Basic fields
	effectsSize := len(h.StatusEffects) * 32 // Estimate per status effect
	return baseSize + effectsSize
}

// Serialize converts the component to bytes
func (h *HealthComponent) Serialize() ([]byte, error) {
	return json.Marshal(h)
}

// Deserialize loads component data from bytes
func (h *HealthComponent) Deserialize(data []byte) error {
	return json.Unmarshal(data, h)
}
