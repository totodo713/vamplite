package mod

import (
	"time"
)

// SecurityLevel セキュリティレベル
type SecurityLevel int

const (
	SecurityLevelUnrestricted SecurityLevel = iota
	SecurityLevelRestricted
	SecurityLevelStrict
)

// ViolationType セキュリティ違反の種類
type ViolationType string

const (
	ViolationTypeCommandInjection          ViolationType = "command_injection"
	ViolationTypePathTraversal             ViolationType = "path_traversal"
	ViolationTypeUnauthorizedNetworkAccess ViolationType = "unauthorized_network"
	ViolationTypeDangerousImport           ViolationType = "dangerous_import"
	ViolationTypeSQLInjection              ViolationType = "sql_injection"
)

// SeverityLevel 深刻度レベル
type SeverityLevel int

const (
	SeverityLow SeverityLevel = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// Resource リソースタイプ
type Resource string

const (
	ResourceEntity    Resource = "entity"
	ResourceComponent Resource = "component"
	ResourceSystem    Resource = "system"
	ResourceFile      Resource = "file"
	ResourceNetwork   Resource = "network"
)

// Action アクションタイプ
type Action string

const (
	ActionCreate     Action = "create"
	ActionRead       Action = "read"
	ActionUpdate     Action = "update"
	ActionDelete     Action = "delete"
	ActionSystemCall Action = "system_call"
)

// Permission 権限タイプ
type Permission string

const (
	PermissionFileRead  Permission = "file_read"
	PermissionFileWrite Permission = "file_write"
	PermissionNetwork   Permission = "network"
)

// ModSecurityValidator MODセキュリティ検証器のメインインターフェース
type ModSecurityValidator interface {
	// 静的解析
	AnalyzeCode(code string) (*SecurityAnalysisResult, error)
	ValidateImports(imports []string) error
	DetectDangerousPatterns(ast interface{}) []SecurityViolation

	// 権限管理
	SetPermissionPolicy(modID string, policy PermissionPolicy) error
	CheckPermission(modID string, resource Resource, action Action) bool
	RequestPermissionElevation(modID string, permission Permission) (*ElevationToken, error)

	// 実行時検証
	ValidateRuntimeOperation(op Operation) error
	MonitorResourceUsage(modID string) *ResourceUsage
	DetectAnomalies(behavior []BehaviorEvent) []Anomaly

	// 監査
	LogSecurityEvent(event ValidatorSecurityEvent) error
	GenerateSecurityReport(modID string, period time.Duration) *SecurityReport
	GetAuditTrail(filter AuditFilter) []AuditEntry
}

// PermissionPolicy MODの権限ポリシー
type PermissionPolicy struct {
	Level            SecurityLevel
	AllowedResources []Resource
	DeniedActions    []Action
	TimeRestrictions TimeWindow
	RateLimits       map[Action]RateLimit
}

// TimeWindow 時間制限
type TimeWindow struct {
	Start time.Time
	End   time.Time
}

// RateLimit レート制限
type RateLimit struct {
	Count  int
	Window time.Duration
}

// SecurityAnalysisResult 静的解析結果
type SecurityAnalysisResult struct {
	Safe        bool
	Violations  []SecurityViolation
	RiskScore   int
	Suggestions []SecuritySuggestion
}

// SecurityViolation セキュリティ違反
type SecurityViolation struct {
	Type        ViolationType
	Severity    SeverityLevel
	Location    CodeLocation
	Description string
	Remediation string
}

// CodeLocation コード位置
type CodeLocation struct {
	File   string
	Line   int
	Column int
}

// SecuritySuggestion セキュリティ提案
type SecuritySuggestion struct {
	Issue       string
	Suggestion  string
	Alternative string
}

// ElevationToken 権限昇格トークン
type ElevationToken struct {
	Token      string
	Permission Permission
	ExpiresAt  time.Time
	ModID      string
}

// Operation 実行時操作
type Operation struct {
	Type   OperationType
	Target string
	ModID  string
}

// OperationType 操作タイプ
type OperationType string

const (
	OpFileRead       OperationType = "file_read"
	OpFileWrite      OperationType = "file_write"
	OpNetworkConnect OperationType = "network_connect"
)

// ResourceUsage リソース使用状況
type ResourceUsage struct {
	Memory     int64
	CPU        float64
	Goroutines int
	Timestamp  time.Time
}

// BehaviorEvent 動作イベント
type BehaviorEvent struct {
	Type     EventType
	Count    int
	Duration time.Duration
	Target   string
}

// EventType イベントタイプ
type EventType string

const (
	EventEntityCreate     EventType = "entity_create"
	EventFileAccess       EventType = "file_access"
	EventNetworkConnect   EventType = "network_connect"
	EventViolation        EventType = "violation"
	EventPermissionDenied EventType = "permission_denied"
)

// Anomaly 異常
type Anomaly struct {
	Type        AnomalyType
	Severity    SeverityLevel
	Description string
	Action      RecommendedAction
}

// AnomalyType 異常タイプ
type AnomalyType string

const (
	AnomalyHighResourceUsage AnomalyType = "high_resource_usage"
	AnomalyUnusualPattern    AnomalyType = "unusual_pattern"
	AnomalySuspiciousAccess  AnomalyType = "suspicious_access"
)

// RecommendedAction 推奨アクション
type RecommendedAction string

const (
	ActionIsolate   RecommendedAction = "isolate"
	ActionTerminate RecommendedAction = "terminate"
	ActionAlert     RecommendedAction = "alert"
	ActionLog       RecommendedAction = "log"
)

// ValidatorSecurityEvent セキュリティ検証イベント
type ValidatorSecurityEvent struct {
	Type      EventType
	ModID     string
	Details   string
	Timestamp time.Time
	Severity  SeverityLevel
}

// SecurityReport セキュリティレポート
type SecurityReport struct {
	ModID           string
	Period          time.Duration
	ViolationCount  int
	RiskScore       int
	Violations      []SecurityViolation
	ResourceUsage   []ResourceUsage
	Recommendations []string
}

// AuditFilter 監査フィルター
type AuditFilter struct {
	ModID      string
	StartTime  time.Time
	EndTime    time.Time
	EventTypes []EventType
	Severity   *SeverityLevel
}

// AuditEntry 監査エントリ
type AuditEntry struct {
	ID        string
	Timestamp time.Time
	ModID     string
	Event     ValidatorSecurityEvent
	Action    string
	Result    string
}
