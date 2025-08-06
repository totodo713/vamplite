package systems

import (
	"math"

	"muscle-dreamer/internal/core/ecs"
)

// AudioSystem handles 3D positional audio, sound effects, and background music.
// It processes entities with AudioComponent to play sounds with proper
// volume and positioning based on the listener's location.
type AudioSystem struct {
	*BaseSystem

	// Audio parameters
	listenerPosition ecs.Vector2
	masterVolume     float64
	audioEngine      AudioEngine
	activeSounds     map[string]*ActiveSound
}

// AudioEngine interface for audio playback abstraction.
type AudioEngine interface {
	PlaySound(soundID string, volume, pitch float64, loop bool) error
	StopSound(soundID string) error
	SetVolume(soundID string, volume float64) error
	IsPlaying(soundID string) bool
	LoadSound(soundID string, filePath string) error
	UnloadSound(soundID string) error
	SetListenerPosition(position ecs.Vector2) error
}

// ActiveSound represents a currently playing sound.
type ActiveSound struct {
	SoundID     string
	EntityID    ecs.EntityID
	Volume      float64
	Pitch       float64
	IsLoop      bool
	StartTime   int64
	Is3D        bool
	Position    ecs.Vector2
	MaxDistance float64
}

// NewAudioSystem creates a new audio system.
func NewAudioSystem() *AudioSystem {
	return &AudioSystem{
		BaseSystem:       NewBaseSystem(AudioSystemType, AudioSystemPriority),
		listenerPosition: ecs.Vector2{X: 0, Y: 0},
		masterVolume:     1.0,
		activeSounds:     make(map[string]*ActiveSound),
	}
}

// GetRequiredComponents returns the components this system operates on.
func (as *AudioSystem) GetRequiredComponents() []ecs.ComponentType {
	return []ecs.ComponentType{ecs.ComponentTypeAudio}
}

// Initialize sets up the audio system.
func (as *AudioSystem) Initialize(world ecs.World) error {
	// TODO: Implement initialization
	return as.BaseSystem.Initialize(world)
}

// Update processes audio entities and updates sound playback.
func (as *AudioSystem) Update(world ecs.World, deltaTime float64) error {
	// TODO: Implement audio processing
	return as.BaseSystem.Update(world, deltaTime)
}

// SetAudioEngine sets the audio engine implementation.
func (as *AudioSystem) SetAudioEngine(engine AudioEngine) {
	as.audioEngine = engine
}

// GetAudioEngine returns the current audio engine.
func (as *AudioSystem) GetAudioEngine() AudioEngine {
	return as.audioEngine
}

// SetListener sets the audio listener position (usually the player).
func (as *AudioSystem) SetListener(position ecs.Vector2) {
	as.listenerPosition = position
	if as.audioEngine != nil {
		as.audioEngine.SetListenerPosition(position)
	}
}

// GetListener returns the current listener position.
func (as *AudioSystem) GetListener() ecs.Vector2 {
	return as.listenerPosition
}

// SetMasterVolume sets the global volume multiplier.
func (as *AudioSystem) SetMasterVolume(volume float64) {
	as.masterVolume = math.Max(0.0, math.Min(1.0, volume))
}

// GetMasterVolume returns the current master volume.
func (as *AudioSystem) GetMasterVolume() float64 {
	return as.masterVolume
}

// PlaySound immediately plays a sound with given parameters.
func (as *AudioSystem) PlaySound(soundID string, volume, pitch float64, loop bool) error {
	if as.audioEngine == nil {
		return nil // No audio engine available
	}

	finalVolume := volume * as.masterVolume
	return as.audioEngine.PlaySound(soundID, finalVolume, pitch, loop)
}

// StopSound stops a currently playing sound.
func (as *AudioSystem) StopSound(soundID string) error {
	if as.audioEngine == nil {
		return nil
	}

	delete(as.activeSounds, soundID)
	return as.audioEngine.StopSound(soundID)
}

// GetActiveSounds returns all currently playing sounds.
func (as *AudioSystem) GetActiveSounds() map[string]*ActiveSound {
	// Return a copy to prevent external modification
	sounds := make(map[string]*ActiveSound)
	for k, v := range as.activeSounds {
		soundCopy := *v
		sounds[k] = &soundCopy
	}
	return sounds
}

// calculate3DVolume computes volume based on distance from listener.
func (as *AudioSystem) calculate3DVolume(audioPos ecs.Vector2, baseVolume, maxDistance float64) float64 {
	distance := math.Sqrt(
		math.Pow(audioPos.X-as.listenerPosition.X, 2) +
			math.Pow(audioPos.Y-as.listenerPosition.Y, 2),
	)

	if distance >= maxDistance {
		return 0.0 // Silent if too far
	}

	// Linear falloff (could be logarithmic for more realistic effect)
	distanceRatio := 1.0 - (distance / maxDistance)
	return baseVolume * distanceRatio * as.masterVolume
}

// calculateDopplerPitch computes pitch based on relative velocity (Doppler effect).
func (as *AudioSystem) calculateDopplerPitch(velocity ecs.Vector2, basePitch float64) float64 {
	// Simplified Doppler effect calculation
	// In a real implementation, this would need more sophisticated physics
	speedOfSound := 343.0                       // m/s
	relativeVelocity := velocity.X + velocity.Y // Simplified

	if math.Abs(relativeVelocity) < 0.1 {
		return basePitch // No significant relative motion
	}

	pitchShift := 1.0 + (relativeVelocity / speedOfSound * 0.1)
	return basePitch * pitchShift
}
