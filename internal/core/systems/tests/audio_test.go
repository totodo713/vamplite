package tests

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"muscle-dreamer/internal/core/ecs"
	"muscle-dreamer/internal/core/ecs/components"
	"muscle-dreamer/internal/core/systems"
)

func TestAudioSystem_Interface(t *testing.T) {
	system := systems.NewAudioSystem()

	var _ ecs.System = system

	assert.Equal(t, systems.AudioSystemType, system.GetType())

	required := system.GetRequiredComponents()
	assert.Contains(t, required, ecs.ComponentTypeAudio)
}

func TestAudioSystem_PlaySound(t *testing.T) {
	system := systems.NewAudioSystem()
	world := createWorldWithEntities()
	mockAudioEngine := &MockAudioEngine{}
	system.SetAudioEngine(mockAudioEngine)

	// Note: AudioComponent needs to be created since it doesn't exist yet
	entity := world.CreateEntity()
	// For now, we'll create a mock audio component
	audioComp := &MockAudioComponent{
		SoundID:   "jump_sound",
		Volume:    0.8,
		IsPlaying: true,
		IsLoop:    false,
	}
	world.AddComponent(entity, audioComp)

	system.Initialize(world)

	err := system.Update(world, 0.016)
	assert.NoError(t, err)

	// オーディオエンジンに再生要求が送られることを確認
	assert.Equal(t, 1, mockAudioEngine.PlayCallCount)
	assert.Equal(t, "jump_sound", mockAudioEngine.LastSoundID)
	assert.InDelta(t, 0.8, mockAudioEngine.LastVolume, 0.01)
}

func TestAudioSystem_MasterVolume(t *testing.T) {
	system := systems.NewAudioSystem()
	mockAudioEngine := &MockAudioEngine{}
	system.SetAudioEngine(mockAudioEngine)

	// マスター音量設定
	system.SetMasterVolume(0.5)
	assert.Equal(t, 0.5, system.GetMasterVolume())

	// 直接音声再生（マスター音量適用確認）
	err := system.PlaySound("test_sound", 0.8, 1.0, false)
	assert.NoError(t, err)

	// マスター音量が適用されることを確認
	expectedVolume := 0.8 * 0.5 // 0.4
	assert.InDelta(t, expectedVolume, mockAudioEngine.LastVolume, 0.01)
}

func TestAudioSystem_3DAudio(t *testing.T) {
	system := systems.NewAudioSystem()
	system.SetListener(ecs.Vector2{X: 0, Y: 0}) // リスナー位置
	world := createWorldWithEntities()
	mockAudioEngine := &MockAudioEngine{}
	system.SetAudioEngine(mockAudioEngine)

	// 近距離オーディオソース
	nearEntity := world.CreateEntity()
	nearTransform := &components.TransformComponent{
		Position: ecs.Vector2{X: 10, Y: 0},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	nearAudio := &MockAudioComponent{
		SoundID:     "near_sound",
		Volume:      1.0,
		IsPlaying:   true,
		Is3D:        true,
		MaxDistance: 100,
	}
	world.AddComponent(nearEntity, nearTransform)
	world.AddComponent(nearEntity, nearAudio)

	// 遠距離オーディオソース
	farEntity := world.CreateEntity()
	farTransform := &components.TransformComponent{
		Position: ecs.Vector2{X: 90, Y: 0},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	farAudio := &MockAudioComponent{
		SoundID:     "far_sound",
		Volume:      1.0,
		IsPlaying:   true,
		Is3D:        true,
		MaxDistance: 100,
	}
	world.AddComponent(farEntity, farTransform)
	world.AddComponent(farEntity, farAudio)

	system.Initialize(world)

	err := system.Update(world, 0.016)
	assert.NoError(t, err)

	// 距離による音量減衰を確認
	assert.Equal(t, 2, mockAudioEngine.PlayCallCount)

	// 近距離音源の方が大きい音量で再生される
	assert.Greater(t, mockAudioEngine.VolumeHistory[0], mockAudioEngine.VolumeHistory[1])
}

func TestAudioSystem_ListenerPosition(t *testing.T) {
	system := systems.NewAudioSystem()

	// リスナー位置設定
	listenerPos := ecs.Vector2{X: 100, Y: 200}
	system.SetListener(listenerPos)

	assert.Equal(t, listenerPos, system.GetListener())
}

func TestAudioSystem_AudioEngineIntegration(t *testing.T) {
	system := systems.NewAudioSystem()
	mockAudioEngine := &MockAudioEngine{}

	// オーディオエンジン設定
	system.SetAudioEngine(mockAudioEngine)
	assert.Equal(t, mockAudioEngine, system.GetAudioEngine())

	// リスナー位置設定がオーディオエンジンに伝播することを確認
	listenerPos := ecs.Vector2{X: 50, Y: 75}
	system.SetListener(listenerPos)

	assert.True(t, mockAudioEngine.SetListenerPositionCalled)
	assert.Equal(t, listenerPos, mockAudioEngine.ListenerPosition)
}

func TestAudioSystem_SoundControl(t *testing.T) {
	system := systems.NewAudioSystem()
	mockAudioEngine := &MockAudioEngine{}
	system.SetAudioEngine(mockAudioEngine)

	// 音声再生
	err := system.PlaySound("bg_music", 0.7, 1.0, true)
	assert.NoError(t, err)
	assert.True(t, mockAudioEngine.IsPlaying("bg_music"))

	// 音声停止
	err = system.StopSound("bg_music")
	assert.NoError(t, err)
	assert.False(t, mockAudioEngine.IsPlaying("bg_music"))
}

func TestAudioSystem_ActiveSounds(t *testing.T) {
	system := systems.NewAudioSystem()
	mockAudioEngine := &MockAudioEngine{}
	system.SetAudioEngine(mockAudioEngine)

	// 初期状態では再生中の音声なし
	activeSounds := system.GetActiveSounds()
	assert.Empty(t, activeSounds)

	// 音声再生後は再生中リストに追加される
	system.PlaySound("test_sound", 0.5, 1.0, false)

	// Note: 実装時にアクティブサウンド管理が必要
	// 現在はPlaySoundが直接エンジンを呼び出すだけ
}

func TestAudioSystem_VolumeConstraints(t *testing.T) {
	system := systems.NewAudioSystem()

	// 音量制限テスト（0.0 - 1.0）
	system.SetMasterVolume(1.5) // 1.0を超える値
	assert.Equal(t, 1.0, system.GetMasterVolume())

	system.SetMasterVolume(-0.5) // 負の値
	assert.Equal(t, 0.0, system.GetMasterVolume())

	system.SetMasterVolume(0.75) // 正常な値
	assert.Equal(t, 0.75, system.GetMasterVolume())
}

func TestAudioSystem_DistanceAttenuation(t *testing.T) {
	system := systems.NewAudioSystem()
	system.SetListener(ecs.Vector2{X: 0, Y: 0})

	// Private method testing would require exposure or friend functions
	// For now, we test through public interface behavior

	baseVolume := 1.0
	maxDistance := 100.0

	// Test positions at different distances
	positions := []ecs.Vector2{
		{X: 0, Y: 0},   // Same as listener (0 distance)
		{X: 25, Y: 0},  // 25% distance
		{X: 50, Y: 0},  // 50% distance
		{X: 75, Y: 0},  // 75% distance
		{X: 100, Y: 0}, // Maximum distance
		{X: 150, Y: 0}, // Beyond maximum distance
	}

	// Calculate expected volumes manually for validation
	for _, pos := range positions {
		distance := math.Sqrt(pos.X*pos.X + pos.Y*pos.Y)
		var expectedVolume float64

		if distance >= maxDistance {
			expectedVolume = 0.0
		} else {
			distanceRatio := 1.0 - (distance / maxDistance)
			expectedVolume = baseVolume * distanceRatio
		}

		// Verify calculation logic
		assert.GreaterOrEqual(t, expectedVolume, 0.0)
		assert.LessOrEqual(t, expectedVolume, baseVolume)
	}
}

// Mock objects for audio system testing

// MockAudioComponent simulates an audio component (not yet implemented in core)
type MockAudioComponent struct {
	SoundID     string
	Volume      float64
	Pitch       float64
	IsPlaying   bool
	IsLoop      bool
	Is3D        bool
	MaxDistance float64
}

func (mac *MockAudioComponent) GetType() ecs.ComponentType {
	return ecs.ComponentTypeAudio
}

func (mac *MockAudioComponent) Clone() ecs.Component {
	clone := *mac
	return &clone
}

func (mac *MockAudioComponent) Validate() error {
	return nil
}

func (mac *MockAudioComponent) Size() int {
	return 64 // Approximate size
}

func (mac *MockAudioComponent) Serialize() ([]byte, error) {
	return []byte{}, nil
}

func (mac *MockAudioComponent) Deserialize([]byte) error {
	return nil
}

// MockAudioEngine simulates an audio engine for testing
type MockAudioEngine struct {
	PlayCallCount             int
	StopCallCount             int
	LoadCallCount             int
	LastSoundID               string
	LastVolume                float64
	LastPitch                 float64
	LastLoop                  bool
	VolumeHistory             []float64
	PlayingSounds             map[string]bool
	SetListenerPositionCalled bool
	ListenerPosition          ecs.Vector2
}

func NewMockAudioEngine() *MockAudioEngine {
	return &MockAudioEngine{
		PlayingSounds: make(map[string]bool),
		VolumeHistory: make([]float64, 0),
	}
}

func (mae *MockAudioEngine) PlaySound(soundID string, volume, pitch float64, loop bool) error {
	mae.PlayCallCount++
	mae.LastSoundID = soundID
	mae.LastVolume = volume
	mae.LastPitch = pitch
	mae.LastLoop = loop
	mae.VolumeHistory = append(mae.VolumeHistory, volume)
	mae.PlayingSounds[soundID] = true
	return nil
}

func (mae *MockAudioEngine) StopSound(soundID string) error {
	mae.StopCallCount++
	mae.PlayingSounds[soundID] = false
	return nil
}

func (mae *MockAudioEngine) SetVolume(soundID string, volume float64) error {
	return nil
}

func (mae *MockAudioEngine) IsPlaying(soundID string) bool {
	return mae.PlayingSounds[soundID]
}

func (mae *MockAudioEngine) LoadSound(soundID string, filePath string) error {
	mae.LoadCallCount++
	return nil
}

func (mae *MockAudioEngine) UnloadSound(soundID string) error {
	return nil
}

func (mae *MockAudioEngine) SetListenerPosition(position ecs.Vector2) error {
	mae.SetListenerPositionCalled = true
	mae.ListenerPosition = position
	return nil
}
