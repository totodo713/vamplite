// ========================================================
// Asset Management Interface Test Suite
// アセット管理インターフェーステスト
// ========================================================

package interfaces_test

import (
	"errors"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"image"
	"image/color"
	"muscle-dreamer/docs/reverse/interfaces"
	"testing"
	"time"
)

// ========================================================
// Mock Asset Implementations
// ========================================================

// MockAsset - Asset のモック実装
type MockAsset struct {
	mock.Mock
	path   string
	size   int64
	loaded bool
}

func NewMockAsset(path string, size int64) *MockAsset {
	return &MockAsset{
		path:   path,
		size:   size,
		loaded: true,
	}
}

func (m *MockAsset) GetPath() string {
	args := m.Called()
	if args.Get(0) != nil {
		return args.String(0)
	}
	return m.path
}

func (m *MockAsset) GetSize() int64 {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(int64)
	}
	return m.size
}

func (m *MockAsset) IsLoaded() bool {
	args := m.Called()
	if len(args) > 0 {
		return args.Bool(0)
	}
	return m.loaded
}

// MockAudioClip - AudioClip のモック実装
type MockAudioClip struct {
	MockAsset
	isPlaying bool
	volume    float64
	duration  time.Duration
}

func NewMockAudioClip(path string, duration time.Duration) *MockAudioClip {
	return &MockAudioClip{
		MockAsset: MockAsset{path: path, size: 1024, loaded: true},
		duration:  duration,
		volume:    1.0,
	}
}

func (m *MockAudioClip) Play() error {
	args := m.Called()
	m.isPlaying = true
	return args.Error(0)
}

func (m *MockAudioClip) Stop() error {
	args := m.Called()
	m.isPlaying = false
	return args.Error(0)
}

func (m *MockAudioClip) SetVolume(volume float64) {
	m.Called(volume)
	m.volume = volume
}

func (m *MockAudioClip) GetDuration() time.Duration {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(time.Duration)
	}
	return m.duration
}

// MockFont - Font のモック実装
type MockFont struct {
	MockAsset
	name string
}

func NewMockFont(path, name string) *MockFont {
	return &MockFont{
		MockAsset: MockAsset{path: path, size: 512, loaded: true},
		name:      name,
	}
}

func (m *MockFont) RenderText(text string, size int) *ebiten.Image {
	args := m.Called(text, size)
	
	if args.Get(0) != nil {
		return args.Get(0).(*ebiten.Image)
	}
	
	img := ebiten.NewImage(len(text)*size/2, size)
	img.Fill(color.White)
	
	return img
}

// MockAssetManager - AssetManager のモック実装
type MockAssetManager struct {
	mock.Mock
	assets map[string]interfaces.Asset
}

func NewMockAssetManager() *MockAssetManager {
	return &MockAssetManager{
		assets: make(map[string]interfaces.Asset),
	}
}

func (m *MockAssetManager) LoadImage(path string) (*ebiten.Image, error) {
	args := m.Called(path)
	
	if args.Get(0) != nil {
		return args.Get(0).(*ebiten.Image), args.Error(1)
	}

	img := ebiten.NewImage(64, 64)
	img.Fill(color.RGBA{255, 0, 0, 255})
	
	return img, nil
}

func (m *MockAssetManager) LoadAudio(path string) (interfaces.AudioClip, error) {
	args := m.Called(path)
	
	if args.Get(0) != nil {
		return args.Get(0).(interfaces.AudioClip), args.Error(1)
	}

	audioClip := NewMockAudioClip(path, time.Second)
	return audioClip, nil
}

func (m *MockAssetManager) LoadFont(path string) (interfaces.Font, error) {
	args := m.Called(path)
	

	

		return args.Get(0).(interfaces.Font), args.Error(1)
	}
	

	return font, nil
}

func (m *MockAssetManager) UnloadAsset(path string) {
	m.Called(path)
	delete(m.assets, path)
}

func (m *MockAssetManager) GetLoadedAssets() map[string]interfaces.Asset {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(map[string]interfaces.Asset)
	}
	return m.assets
}

// ========================================================
// Asset Interface Tests
// ========================================================

// TestAssetInterface - Asset インターフェース契約テスト
func TestAssetInterface(t *testing.T) {
	t.Run("BasicAssetProperties", func(t *testing.T) {
		asset := NewMockAsset("test/asset.png", 1024)
		

		asset.On("GetSize").Return(int64(1024))
		asset.On("IsLoaded").Return(true)
		

		assert.Equal(t, int64(1024), asset.GetSize())
		assert.True(t, asset.IsLoaded())
		

	})
	

		validPaths := []string{
			"assets/sprites/player.png",
			"assets/audio/bgm.ogg",
			"assets/fonts/ui_font.ttf",
		}
		

			asset := NewMockAsset(path, 512)
			assert.Equal(t, path, asset.GetPath())
			assert.Greater(t, len(asset.GetPath()), 0)
		}
	})
}

// TestAudioClipInterface - AudioClip インターフェース契約テスト
func TestAudioClipInterface(t *testing.T) {
	t.Run("AudioPlayback", func(t *testing.T) {
		audioClip := NewMockAudioClip("assets/audio/test.ogg", 5*time.Second)
		

		audioClip.On("Stop").Return(nil)
		audioClip.On("SetVolume", 0.5).Return()
		audioClip.On("GetDuration").Return(5 * time.Second)
		

		err := audioClip.Play()
		assert.NoError(t, err)
		assert.True(t, audioClip.isPlaying)
		

		audioClip.SetVolume(0.5)
		assert.Equal(t, 0.5, audioClip.volume)
		

		duration := audioClip.GetDuration()
		assert.Equal(t, 5*time.Second, duration)
		

		err = audioClip.Stop()
		assert.NoError(t, err)
		assert.False(t, audioClip.isPlaying)
		

	})
	

		audioClip := NewMockAudioClip("test.ogg", time.Second)
		

		

			audioClip.On("SetVolume", vol).Return()
			audioClip.SetVolume(vol)
			assert.Equal(t, vol, audioClip.volume)
		}
	})
	

		audioClip := NewMockAudioClip("corrupted.ogg", 0)
		

		audioClip.On("Stop").Return(errors.New("audio not playing"))
		

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "audio format not supported")
		

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "audio not playing")
		

	})
}

// TestFontInterface - Font インターフェース契約テスト
func TestFontInterface(t *testing.T) {
	t.Run("TextRendering", func(t *testing.T) {
		font := NewMockFont("assets/fonts/ui.ttf", "UIFont")
		

		

		assert.NotNil(t, textImage)
		

		assert.Greater(t, bounds.Dx(), 0)
		assert.Greater(t, bounds.Dy(), 0)
		

	})
	

		font := NewMockFont("test.ttf", "TestFont")
		

		

			font.On("RenderText", "Test", size).Return((*ebiten.Image)(nil))
			

			assert.NotNil(t, img)
			

			assert.Equal(t, size, bounds.Dy()) // 高さがサイズと一致
		}
	})
	

		font := NewMockFont("test.ttf", "TestFont")
		

		

		assert.NotNil(t, img)
		

		bounds := img.Bounds()
		assert.GreaterOrEqual(t, bounds.Dx(), 0)
	})
}

// ========================================================
// AssetManager Interface Tests
// ========================================================

// TestAssetManagerInterface - AssetManager インターフェース契約テスト
func TestAssetManagerInterface(t *testing.T) {
	am := NewMockAssetManager()
	

		imagePath := "assets/sprites/player.png"
		

		am.On("GetLoadedAssets").Return((map[string]interfaces.Asset)(nil))
		

		assert.NoError(t, err)
		assert.NotNil(t, img)
		

		loadedAssets := am.GetLoadedAssets()
		assert.Contains(t, loadedAssets, imagePath)
		

	})
	

		audioPath := "assets/audio/bgm.ogg"
		

		

		assert.NoError(t, err)
		assert.NotNil(t, audio)
		

		assert.Implements(t, (*interfaces.AudioClip)(nil), audio)
		

	})
	

		fontPath := "assets/fonts/ui.ttf"
		

		

		assert.NoError(t, err)
		assert.NotNil(t, font)
		

		assert.Implements(t, (*interfaces.Font)(nil), font)
		

	})
	

		assetPath := "assets/temp/temp.png"
		

		am.On("LoadImage", assetPath).Return((*ebiten.Image)(nil), nil)
		am.On("UnloadAsset", assetPath).Return()
		am.On("GetLoadedAssets").Return((map[string]interfaces.Asset)(nil))
		

		assert.NoError(t, err)
		

		am.UnloadAsset(assetPath)
		

		loadedAssets := am.GetLoadedAssets()
		assert.NotContains(t, loadedAssets, assetPath)
		

	})
}

// ========================================================
// Error Handling Tests
// ========================================================

// TestAssetManagerErrorHandling - AssetManager エラーハンドリングテスト
func TestAssetManagerErrorHandling(t *testing.T) {
	am := NewMockAssetManager()
	

		invalidPaths := []string{
			"nonexistent.png",
			"../../../etc/passwd",
			"",
			"assets/images/\x00malicious.png",
		}
		

			am.On("LoadImage", path).Return((*ebiten.Image)(nil), errors.New("invalid path"))
			

			assert.Error(t, err)
		}
	})
	

		unsupportedPath := "assets/audio/unsupported.mp3"
		

		

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
		

	})
	

		largePath := "assets/huge_image.png"
		

		

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too large")
		

	})
}

// ========================================================
// Performance Tests
// ========================================================

// TestAssetManagerPerformance - AssetManager パフォーマンステスト
func TestAssetManagerPerformance(t *testing.T) {
	t.Run("BulkAssetLoading", func(t *testing.T) {
		am := NewMockAssetManager()
		

		

		for i := 0; i < assetCount; i++ {
			path := fmt.Sprintf("assets/test_%d.png", i)
			am.On("LoadImage", path).Return((*ebiten.Image)(nil), nil)
		}
		

		for i := 0; i < assetCount; i++ {
			path := fmt.Sprintf("assets/test_%d.png", i)
			_, err := am.LoadImage(path)
			assert.NoError(t, err)
		}
		elapsed := time.Since(start)
		

		assert.Less(t, elapsed, time.Second)
		

		

	})
	

		am := NewMockAssetManager()
		

		

		am.On("LoadImage", assetPath).Return((*ebiten.Image)(nil), nil).Once()
		am.On("GetLoadedAssets").Return((map[string]interfaces.Asset)(nil))
		

		firstLoadStart := time.Now()
		_, err := am.LoadImage(assetPath)
		firstLoadTime := time.Since(firstLoadStart)
		assert.NoError(t, err)
		

		loadedAssets := am.GetLoadedAssets()
		assert.Contains(t, loadedAssets, assetPath)
		

		

	})
}

// ========================================================
// Integration Tests
// ========================================================

// TestAssetIntegration - アセット統合テスト
func TestAssetIntegration(t *testing.T) {
	t.Run("GameAssetWorkflow", func(t *testing.T) {
		am := NewMockAssetManager()
		

		assets := map[string]string{
			"player_sprite": "assets/sprites/player.png",
			"bgm":          "assets/audio/background.ogg",
			"bgm":           "assets/audio/background.ogg",
			"ui_font":       "assets/fonts/ui.ttf",
		

		

		am.On("LoadImage", assets["player_sprite"]).Return((*ebiten.Image)(nil), nil)
		img, err := am.LoadImage(assets["player_sprite"])
		assert.NoError(t, err)
		loadedAssets["player_sprite"] = img
		

		am.On("LoadAudio", assets["bgm"]).Return((interfaces.AudioClip)(nil), nil)
		audio, err := am.LoadAudio(assets["bgm"])
		assert.NoError(t, err)
		loadedAssets["bgm"] = audio
		

		am.On("LoadFont", assets["ui_font"]).Return((interfaces.Font)(nil), nil)
		font, err := am.LoadFont(assets["ui_font"])
		assert.NoError(t, err)
		loadedAssets["ui_font"] = font
		

		assert.Len(t, loadedAssets, 3)
		assert.NotNil(t, loadedAssets["player_sprite"])
		assert.NotNil(t, loadedAssets["bgm"])
		assert.NotNil(t, loadedAssets["ui_font"])
		

		assert.IsType(t, &ebiten.Image{}, loadedAssets["player_sprite"])
		assert.Implements(t, (*interfaces.AudioClip)(nil), loadedAssets["bgm"])
		assert.Implements(t, (*interfaces.Font)(nil), loadedAssets["ui_font"])
		

	})
}

// ========================================================
// Benchmark Tests
// ========================================================

// BenchmarkAssetLoading - アセット読み込みベンチマーク
func BenchmarkAssetLoading(b *testing.B) {
	am := NewMockAssetManager()
	

	

	b.ReportAllocs()
	

		path := fmt.Sprintf("assets/benchmark_%d.png", i%100)
		am.LoadImage(path)
	}
}

// BenchmarkAudioPlayback - オーディオ再生ベンチマーク
func BenchmarkAudioPlayback(b *testing.B) {
	audioClip := NewMockAudioClip("benchmark.ogg", time.Second)
	

	audioClip.On("Stop").Return(nil)
	

	b.ReportAllocs()
	

		audioClip.Play()
		audioClip.Stop()
	}
}

// BenchmarkTextRendering - テキスト描画ベンチマーク
func BenchmarkTextRendering(b *testing.B) {
	font := NewMockFont("benchmark.ttf", "BenchmarkFont")
	

	

	b.ReportAllocs()
	

		text := fmt.Sprintf("Benchmark text %d", i)
		font.RenderText(text, 16)
	}
}
