package mod

import (
	"fmt"
	"regexp"
	"sync"
	"time"
)

// SecuritySeverity セキュリティ重要度
type SecuritySeverity int

const (
	SecurityInfo SecuritySeverity = iota
	SecurityWarning
	SecurityCritical
)

// SecurityEvent セキュリティイベント
type SecurityEvent struct {
	Timestamp time.Time
	ModID     string
	Operation string
	Details   string
	Severity  SecuritySeverity
}

// SecurityAuditLogger セキュリティ監査ログ
type SecurityAuditLogger interface {
	LogViolation(modID, operation, details string)
	LogSuspiciousActivity(modID, activity string)
	GetViolationHistory(modID string) []SecurityEvent
}

// SecurityAuditLoggerImpl セキュリティ監査ログの実装
type SecurityAuditLoggerImpl struct {
	mu     sync.RWMutex
	events map[string][]SecurityEvent
}

// NewSecurityAuditLogger 新しいセキュリティ監査ログを作成
func NewSecurityAuditLogger() SecurityAuditLogger {
	return &SecurityAuditLoggerImpl{
		events: make(map[string][]SecurityEvent),
	}
}

func (s *SecurityAuditLoggerImpl) LogViolation(modID, operation, details string) {
	s.logEvent(modID, operation, details, SecurityCritical)
}

func (s *SecurityAuditLoggerImpl) LogSuspiciousActivity(modID, activity string) {
	s.logEvent(modID, "suspicious_activity", activity, SecurityWarning)
}

func (s *SecurityAuditLoggerImpl) logEvent(modID, operation, details string, severity SecuritySeverity) {
	s.mu.Lock()
	defer s.mu.Unlock()

	event := SecurityEvent{
		Timestamp: time.Now(),
		ModID:     modID,
		Operation: operation,
		Details:   details,
		Severity:  severity,
	}

	s.events[modID] = append(s.events[modID], event)
}

func (s *SecurityAuditLoggerImpl) GetViolationHistory(modID string) []SecurityEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.events[modID]
}

// AdvancedSecurityValidator 高度なセキュリティ検証
type AdvancedSecurityValidator struct {
	modID             string
	dangerousPatterns []*regexp.Regexp
	auditLogger       SecurityAuditLogger
	violationCount    int
	maxViolations     int
}

// NewAdvancedSecurityValidator 高度なセキュリティ検証器を作成
func NewAdvancedSecurityValidator(modID string, logger SecurityAuditLogger) *AdvancedSecurityValidator {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\.\.+/`),                    // パストラバーサル
		regexp.MustCompile(`(rm|del|delete).*(-r|-rf)`), // 削除コマンド
		regexp.MustCompile(`^(exec|cmd)$`),              // システム実行（完全一致）
		regexp.MustCompile(`(http|tcp|udp)://`),         // ネットワーク
		regexp.MustCompile(`/etc/(passwd|shadow)`),      // システムファイル
		regexp.MustCompile(`\.(ssh|config)`),            // 設定ファイル
	}

	return &AdvancedSecurityValidator{
		modID:             modID,
		dangerousPatterns: patterns,
		auditLogger:       logger,
		maxViolations:     5, // 5回違反で停止
	}
}

// ValidateSystemID システムIDの安全性を検証
func (s *AdvancedSecurityValidator) ValidateSystemID(systemID string) error {
	for _, pattern := range s.dangerousPatterns {
		if pattern.MatchString(systemID) {
			s.violationCount++
			s.auditLogger.LogViolation(s.modID, "system_id", fmt.Sprintf("pattern match: %s", pattern.String()))

			if s.violationCount >= s.maxViolations {
				return &SecurityError{
					ModID:     s.modID,
					Operation: "system_register",
					Reason:    fmt.Sprintf("too many violations (%d)", s.violationCount),
				}
			}

			return &SecurityError{
				ModID:     s.modID,
				Operation: "system_register",
				Reason:    fmt.Sprintf("pattern match: %s", pattern.String()),
			}
		}
	}
	return nil
}

// ValidateEntityTag エンティティタグの安全性を検証
func (s *AdvancedSecurityValidator) ValidateEntityTag(tag string) error {
	// タグ内の危険パターンをチェック
	for _, pattern := range s.dangerousPatterns {
		if pattern.MatchString(tag) {
			s.auditLogger.LogSuspiciousActivity(s.modID, fmt.Sprintf("suspicious tag: %s", tag))
			// タグは警告のみで拒否しない（実際のファイルアクセスはブロック）
		}
	}
	return nil
}

// GetViolationCount 違反回数を取得
func (s *AdvancedSecurityValidator) GetViolationCount() int {
	return s.violationCount
}

// ResetViolations 違反回数をリセット
func (s *AdvancedSecurityValidator) ResetViolations() {
	s.violationCount = 0
}

// ResourceMonitor リソース監視
type ResourceMonitor struct {
	mu              sync.RWMutex
	memoryThreshold int64
	entityThreshold int
	observers       []ResourceObserver
}

// ResourceObserver リソース変更の観察者
type ResourceObserver interface {
	OnMemoryUsageChanged(modID string, oldUsage, newUsage int64)
	OnEntityCountChanged(modID string, oldCount, newCount int)
	OnThresholdExceeded(modID string, resource string, current, threshold interface{})
}

// NewResourceMonitor 新しいリソース監視を作成
func NewResourceMonitor(memoryThreshold int64, entityThreshold int) *ResourceMonitor {
	return &ResourceMonitor{
		memoryThreshold: memoryThreshold,
		entityThreshold: entityThreshold,
		observers:       make([]ResourceObserver, 0),
	}
}

// AddObserver 観察者を追加
func (r *ResourceMonitor) AddObserver(observer ResourceObserver) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.observers = append(r.observers, observer)
}

// NotifyMemoryUsageChanged メモリ使用量変更を通知
func (r *ResourceMonitor) NotifyMemoryUsageChanged(modID string, oldUsage, newUsage int64) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, observer := range r.observers {
		observer.OnMemoryUsageChanged(modID, oldUsage, newUsage)
	}

	if newUsage >= r.memoryThreshold {
		for _, observer := range r.observers {
			observer.OnThresholdExceeded(modID, "memory", newUsage, r.memoryThreshold)
		}
	}
}

// NotifyEntityCountChanged エンティティ数変更を通知
func (r *ResourceMonitor) NotifyEntityCountChanged(modID string, oldCount, newCount int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, observer := range r.observers {
		observer.OnEntityCountChanged(modID, oldCount, newCount)
	}

	if newCount >= r.entityThreshold {
		for _, observer := range r.observers {
			observer.OnThresholdExceeded(modID, "entities", newCount, r.entityThreshold)
		}
	}
}
