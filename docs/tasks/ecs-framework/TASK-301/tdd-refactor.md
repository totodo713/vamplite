# TASK-301: ModECSAPI実装 - Refactor段階（品質向上）

## リファクタリング概要

Green段階で実装した最小機能を基に、コードの品質向上、パフォーマンス最適化、セキュリティ強化を行います。

## リファクタリング対象と改善内容

### 1. パフォーマンス最適化

#### 1.1 メモリ効率の改善
**対象**: `ModEntityAPIImpl`, `ModComponentAPIImpl`  
**改善内容**:
- エンティティ管理の効率化（スライスからマップへ）
- メモリプールの導入
- ガベージコレクション負荷軽減

**実装**:
```go
// 改善前: スライスでの線形検索
func (m *ModEntityAPIImpl) isOwnedEntity(id ecs.EntityID) bool {
    for _, entityID := range m.api.context.CreatedEntities {
        if entityID == id {
            return true
        }
    }
    return false
}

// 改善後: マップでのO(1)検索
type ModEntityAPIImpl struct {
    api           *ModECSAPIImpl
    nextID        ecs.EntityID
    entities      map[ecs.EntityID][]string
    ownedEntities map[ecs.EntityID]bool  // 高速所有権チェック用
}

func (m *ModEntityAPIImpl) isOwnedEntity(id ecs.EntityID) bool {
    return m.ownedEntities[id]
}
```

#### 1.2 クエリ実行の最適化
**対象**: `ModQueryAPIImpl`  
**改善内容**:
- クエリ結果キャッシュ
- 不要なクエリ実行回避
- バッチ処理対応

### 2. セキュリティ強化

#### 2.1 高度な脅威検出
**対象**: `ModSystemAPIImpl`, セキュリティチェック  
**改善内容**:
- より詳細なパストラバーサル検出
- 危険APIパターンの包括的チェック
- セキュリティ監査ログ

**実装**:
```go
// SecurityValidator セキュリティ検証器
type SecurityValidator struct {
    modID            string
    dangerousPatterns []string
    auditLogger      SecurityAuditLogger
}

func (s *SecurityValidator) ValidateSystemID(systemID string) error {
    // 高度なパターンマッチング
    patterns := []string{
        `\.\.+/`,                    // パストラバーサル
        `(rm|del|delete).*(-r|-rf)`, // 削除コマンド
        `(exec|system|cmd)`,         // システム実行
        `(http|tcp|udp)://`,         // ネットワーク
    }
    
    for _, pattern := range patterns {
        if matched, _ := regexp.MatchString(pattern, systemID); matched {
            s.auditLogger.LogViolation(s.modID, "system_id", pattern)
            return &SecurityError{
                ModID:     s.modID,
                Operation: "system_register",
                Reason:    fmt.Sprintf("pattern match: %s", pattern),
            }
        }
    }
    return nil
}
```

#### 2.2 リソース監視強化
**対象**: `ModContext`, リソース制限  
**改善内容**:
- リアルタイムリソース監視
- 異常検出・アラート
- 詳細な使用量統計

### 3. エラーハンドリング改善

#### 3.1 詳細なエラー情報
**対象**: 全エラー処理  
**改善内容**:
- エラー発生時のコンテキスト情報
- スタックトレース（開発時）
- 復旧可能性の判定

**実装**:
```go
// EnhancedError 拡張エラー情報
type EnhancedError struct {
    BaseError    error
    ModID        string
    Operation    string
    Context      map[string]interface{}
    Timestamp    time.Time
    Recoverable  bool
}

func (e *EnhancedError) Error() string {
    return fmt.Sprintf("[%s] %s in %s: %v (context: %+v)", 
        e.Timestamp.Format(time.RFC3339), 
        e.Operation, 
        e.ModID, 
        e.BaseError, 
        e.Context)
}
```

### 4. 設計パターン適用

#### 4.1 Builder Pattern適用
**対象**: `ModConfig`作成  
**改善内容**:
- 設定の柔軟性向上
- デフォルト値管理
- 検証ロジック統合

#### 4.2 Observer Pattern適用
**対象**: リソース監視  
**改善内容**:
- リソース使用量変更の通知
- しきい値監視
- イベント駆動型制限

### 5. コード品質向上

#### 5.1 インターフェース分離
**対象**: 大きなインターフェース  
**改善内容**:
- 単一責任原則適用
- インターフェース分割
- 疎結合設計

#### 5.2 テスタビリティ向上
**対象**: テストコード  
**改善内容**:
- モック実装の追加
- テスト用ファクトリー
- より詳細なアサーション

## リファクタリング実装手順

### Phase 1: パフォーマンス最適化

実装先: `internal/core/ecs/mod/performance.go`
```go
package mod

import (
    "sync"
    "time"
)

// PerformanceMonitor パフォーマンス監視
type PerformanceMonitor struct {
    mu                sync.RWMutex
    apiCallDurations  map[string][]time.Duration
    memorySnapshots   []int64
    queryFrequency    map[string]int
}

// EntityPool エンティティ用オブジェクトプール
type EntityPool struct {
    pool chan *ModEntityAPIImpl
}

func NewEntityPool(size int) *EntityPool {
    return &EntityPool{
        pool: make(chan *ModEntityAPIImpl, size),
    }
}

func (p *EntityPool) Get() *ModEntityAPIImpl {
    select {
    case entity := <-p.pool:
        return entity
    default:
        return &ModEntityAPIImpl{}
    }
}

func (p *EntityPool) Put(entity *ModEntityAPIImpl) {
    // リセット処理
    entity.entities = make(map[ecs.EntityID][]string)
    entity.ownedEntities = make(map[ecs.EntityID]bool)
    
    select {
    case p.pool <- entity:
    default:
        // プール満杯時は破棄
    }
}
```

### Phase 2: セキュリティ強化

実装先: `internal/core/ecs/mod/security.go`
```go
package mod

import (
    "regexp"
    "time"
)

// SecurityAuditLogger セキュリティ監査ログ
type SecurityAuditLogger interface {
    LogViolation(modID, operation, details string)
    LogSuspiciousActivity(modID, activity string)
    GetViolationHistory(modID string) []SecurityEvent
}

// SecurityEvent セキュリティイベント
type SecurityEvent struct {
    Timestamp time.Time
    ModID     string
    Operation string
    Details   string
    Severity  SecuritySeverity
}

// SecuritySeverity セキュリティ重要度
type SecuritySeverity int

const (
    SecurityInfo SecuritySeverity = iota
    SecurityWarning
    SecurityCritical
)

// AdvancedSecurityValidator 高度なセキュリティ検証
type AdvancedSecurityValidator struct {
    modID                string
    dangerousPatterns    []*regexp.Regexp
    auditLogger         SecurityAuditLogger
    violationCount      int
    maxViolations       int
}

func NewAdvancedSecurityValidator(modID string, logger SecurityAuditLogger) *AdvancedSecurityValidator {
    patterns := []*regexp.Regexp{
        regexp.MustCompile(`\.\.+/`),                    // パストラバーサル
        regexp.MustCompile(`(rm|del|delete).*(-r|-rf)`), // 削除コマンド
        regexp.MustCompile(`(exec|system|cmd)`),         // システム実行
        regexp.MustCompile(`(http|tcp|udp)://`),         // ネットワーク
        regexp.MustCompile(`/etc/(passwd|shadow)`),      // システムファイル
        regexp.MustCompile(`\..(ssh|config)`),           // 設定ファイル
    }

    return &AdvancedSecurityValidator{
        modID:             modID,
        dangerousPatterns: patterns,
        auditLogger:      logger,
        maxViolations:    5, // 5回違反で停止
    }
}
```

### Phase 3: 設計パターン適用

実装先: `internal/core/ecs/mod/builder.go`
```go
package mod

// ModConfigBuilder MOD設定のビルダー
type ModConfigBuilder struct {
    config ModConfig
}

// NewModConfigBuilder 新しいビルダーを作成
func NewModConfigBuilder() *ModConfigBuilder {
    return &ModConfigBuilder{
        config: DefaultModConfig(),
    }
}

func (b *ModConfigBuilder) WithMaxEntities(max int) *ModConfigBuilder {
    b.config.MaxEntities = max
    return b
}

func (b *ModConfigBuilder) WithMaxMemory(max int64) *ModConfigBuilder {
    b.config.MaxMemory = max
    return b
}

func (b *ModConfigBuilder) WithAllowedComponents(components ...ecs.ComponentType) *ModConfigBuilder {
    b.config.AllowedComponents = components
    return b
}

func (b *ModConfigBuilder) Build() (ModConfig, error) {
    // 設定値検証
    if b.config.MaxEntities <= 0 {
        return ModConfig{}, errors.New("max entities must be positive")
    }
    if b.config.MaxMemory <= 0 {
        return ModConfig{}, errors.New("max memory must be positive")
    }
    
    return b.config, nil
}
```

## リファクタリング実行手順

### 1. パフォーマンス改善実装
```bash
# パフォーマンス監視機能追加
touch internal/core/ecs/mod/performance.go

# 既存実装の最適化
# - ModEntityAPIImpl の効率化
# - メモリプール導入
```

### 2. セキュリティ強化実装
```bash
# セキュリティ機能追加
touch internal/core/ecs/mod/security.go

# 既存セキュリティチェックの強化
```

### 3. 設計パターン適用
```bash
# ビルダーパターン実装
touch internal/core/ecs/mod/builder.go

# Observer パターンでリソース監視
touch internal/core/ecs/mod/observer.go
```

### 4. テスト品質向上
```bash
# モック実装追加
touch internal/core/ecs/mod/mocks.go

# 追加テストケース実装
# - 高負荷テスト
# - セキュリティテスト強化
# - エラー境界テスト
```

### 5. ドキュメント更新
```bash
# API仕様書作成
touch internal/core/ecs/mod/README.md

# 使用例・サンプル追加
touch internal/core/ecs/mod/examples_test.go
```

## 品質メトリクス目標

### パフォーマンス目標
- [ ] API呼び出しオーバーヘッド: <50μs（改善前<100μs）
- [ ] メモリ使用効率: 20%向上
- [ ] ガベージコレクション負荷: 50%削減

### セキュリティ目標
- [ ] 脅威検出率: >99%
- [ ] 誤検出率: <1%
- [ ] セキュリティ監査完全性: 100%

### コード品質目標
- [ ] 循環的複雑度: <8
- [ ] テストカバレッジ: >95%
- [ ] コードデュプリケーション: <2%

## リファクタリング後の期待効果

### 技術的効果
- **実行効率向上**: 20-30%のパフォーマンス改善
- **メモリ使用量削減**: 15-25%のメモリ効率化
- **セキュリティ強化**: より包括的な脅威防御

### 保守性効果
- **コード可読性向上**: 明確な責任分離
- **テスタビリティ向上**: モック・スタブ対応
- **拡張性向上**: 新機能追加の容易性

---

**作成日時**: 2025-08-08  
**段階**: TDD Refactor  
**目標**: 品質向上・最適化 🎯