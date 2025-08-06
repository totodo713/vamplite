package components

import (
	"muscle-dreamer/internal/core/ecs"
	"time"
)

// AudioComponent handles 3D positional audio and sound effects for entities.
// It supports distance-based volume attenuation and directional audio.
type AudioComponent struct {
	// Sound identification
	SoundID string `json:"sound_id"`

	// Playback control
	Volume    float64 `json:"volume"` // 0.0 to 1.0
	Pitch     float64 `json:"pitch"`  // Pitch multiplier (1.0 = normal)
	IsPlaying bool    `json:"is_playing"`
	IsLoop    bool    `json:"is_loop"`
	IsPaused  bool    `json:"is_paused"`

	// 3D audio properties
	Is3D        bool    `json:"is_3d"`        // Enable 3D positional audio
	MaxDistance float64 `json:"max_distance"` // Maximum audible distance
	MinDistance float64 `json:"min_distance"` // Minimum distance (full volume)
	Rolloff     float64 `json:"rolloff"`      // Distance attenuation factor

	// Audio effects
	LowPassFilter  float64 `json:"low_pass_filter"`  // Low-pass filter frequency
	HighPassFilter float64 `json:"high_pass_filter"` // High-pass filter frequency
	ReverbLevel    float64 `json:"reverb_level"`     // Reverb effect level

	// Runtime state
	PlaybackPosition float64 `json:"playback_position"` // Current playback position in seconds
	LastPlayTime     int64   `json:"last_play_time"`    // Last play time in nanoseconds
	FadeIn           float64 `json:"fade_in"`           // Fade in duration
	FadeOut          float64 `json:"fade_out"`          // Fade out duration

	// Priority and grouping
	Priority   int    `json:"priority"`    // Playback priority (higher = more important)
	AudioGroup string `json:"audio_group"` // Audio group for volume control (e.g., "sfx", "music", "voice")
}

// NewAudioComponent creates a new audio component with default values.
func NewAudioComponent(soundID string) *AudioComponent {
	return &AudioComponent{
		SoundID:     soundID,
		Volume:      1.0,
		Pitch:       1.0,
		IsPlaying:   false,
		IsLoop:      false,
		IsPaused:    false,
		Is3D:        false,
		MaxDistance: 100.0,
		MinDistance: 1.0,
		Rolloff:     1.0,
		Priority:    0,
		AudioGroup:  "sfx",
	}
}

// GetType returns the component type for identification.
func (ac *AudioComponent) GetType() ecs.ComponentType {
	return ecs.ComponentTypeAudio
}

// Clone creates a deep copy of the component.
func (ac *AudioComponent) Clone() ecs.Component {
	clone := *ac
	return &clone
}

// Validate ensures the component data is valid.
func (ac *AudioComponent) Validate() error {
	if ac.SoundID == "" {
		return &ecs.ECSError{
			Code:      "VALIDATION_ERROR",
			Message:   "AudioComponent: SoundID cannot be empty",
			Component: string(ecs.ComponentTypeAudio),
			Timestamp: time.Now(),
		}
	}

	if ac.Volume < 0.0 || ac.Volume > 1.0 {
		return &ecs.ECSError{
			Code:      "VALIDATION_ERROR",
			Message:   "AudioComponent: Volume must be between 0.0 and 1.0",
			Component: string(ecs.ComponentTypeAudio),
			Timestamp: time.Now(),
		}
	}

	if ac.Pitch <= 0.0 {
		return &ecs.ECSError{
			Code:      "VALIDATION_ERROR",
			Message:   "AudioComponent: Pitch must be greater than 0.0",
			Component: string(ecs.ComponentTypeAudio),
			Timestamp: time.Now(),
		}
	}

	if ac.MaxDistance <= 0.0 {
		return &ecs.ECSError{
			Code:      "VALIDATION_ERROR",
			Message:   "AudioComponent: MaxDistance must be greater than 0.0",
			Component: string(ecs.ComponentTypeAudio),
			Timestamp: time.Now(),
		}
	}

	if ac.MinDistance < 0.0 || ac.MinDistance > ac.MaxDistance {
		return &ecs.ECSError{
			Code:      "VALIDATION_ERROR",
			Message:   "AudioComponent: MinDistance must be between 0.0 and MaxDistance",
			Component: string(ecs.ComponentTypeAudio),
			Timestamp: time.Now(),
		}
	}

	return nil
}

// Size returns the memory size of the component in bytes.
func (ac *AudioComponent) Size() int {
	return 96 + len(ac.SoundID) + len(ac.AudioGroup) // Approximate size
}

// Serialize converts the component to bytes for persistence.
func (ac *AudioComponent) Serialize() ([]byte, error) {
	// TODO: Implement proper serialization
	return []byte{}, nil
}

// Deserialize loads component data from bytes.
func (ac *AudioComponent) Deserialize(data []byte) error {
	// TODO: Implement proper deserialization
	return nil
}

// Play starts audio playback.
func (ac *AudioComponent) Play() {
	ac.IsPlaying = true
	ac.IsPaused = false
}

// Stop stops audio playback.
func (ac *AudioComponent) Stop() {
	ac.IsPlaying = false
	ac.IsPaused = false
	ac.PlaybackPosition = 0
}

// Pause pauses audio playback.
func (ac *AudioComponent) Pause() {
	ac.IsPaused = true
}

// Resume resumes paused audio playback.
func (ac *AudioComponent) Resume() {
	if ac.IsPaused {
		ac.IsPaused = false
	}
}

// SetVolume sets the playback volume with validation.
func (ac *AudioComponent) SetVolume(volume float64) {
	if volume < 0.0 {
		ac.Volume = 0.0
	} else if volume > 1.0 {
		ac.Volume = 1.0
	} else {
		ac.Volume = volume
	}
}

// SetPitch sets the playback pitch with validation.
func (ac *AudioComponent) SetPitch(pitch float64) {
	if pitch > 0.0 {
		ac.Pitch = pitch
	}
}

// Set3D enables or disables 3D positional audio.
func (ac *AudioComponent) Set3D(enable bool, maxDistance, minDistance, rolloff float64) {
	ac.Is3D = enable
	if enable {
		ac.MaxDistance = maxDistance
		ac.MinDistance = minDistance
		ac.Rolloff = rolloff
	}
}

// IsActive returns true if the audio component is actively playing.
func (ac *AudioComponent) IsActive() bool {
	return ac.IsPlaying && !ac.IsPaused
}

// GetEffectiveVolume returns the effective volume considering fade effects.
func (ac *AudioComponent) GetEffectiveVolume(currentTime float64) float64 {
	volume := ac.Volume

	// Apply fade in/out effects
	if ac.FadeIn > 0 && currentTime < ac.FadeIn {
		fadeInRatio := currentTime / ac.FadeIn
		volume *= fadeInRatio
	}

	if ac.FadeOut > 0 {
		// This would need playback duration information
		// For now, just return the base volume
	}

	return volume
}
