# Lint対応 アーキテクチャ設計

## システム概要

Muscle DreamerプロジェクトのGoコードに対する静的解析とコード品質管理システム。golangci-lintを中核とした包括的な品質保証パイプラインを提供し、50以上のlinter規則への準拠とECSフレームワーク特有のパフォーマンス要件を両立させる。

## アーキテクチャパターン

- **パターン**: Pipeline Architecture with Quality Gates
- **理由**: 
  - 段階的な品質チェック処理が可能
  - 各段階での品質ゲート設定により早期フィードバック実現
  - CI/CDパイプラインとの自然な統合
  - 拡張性とメンテナンス性の確保

## システム全体構成

```
┌─────────────────────────────────────────────────────────────┐
│                    開発環境                                  │
├─────────────────────────────────────────────────────────────┤
│  IDE/Editor     │  Local Tools      │  Development Server   │
│  - VS Code      │  - golangci-lint  │  - make dev           │
│  - GoLand       │  - gofumpt        │  - Auto-reload        │
│  - vim/neovim   │  - goimports      │  - Debug mode         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  Git Repository                             │
├─────────────────────────────────────────────────────────────┤
│  Configuration  │  Source Code      │  Build System         │
│  - .golangci.yml│  - internal/      │  - Makefile          │
│  - .github/     │  - cmd/           │  - Docker compose    │
│  - CLAUDE.md    │  - *.go files     │  - mise.toml         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 CI/CD Pipeline (GitHub Actions)            │
├─────────────────────────────────────────────────────────────┤
│  Trigger Stage  │  Quality Stage    │  Integration Stage    │
│  - Push/PR      │  - lint check     │  - Build test        │
│  - File filter  │  - Format check   │  - Performance test   │
│  - Path filter  │  - Security check │  - Integration test   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Quality Feedback Loop                    │
├─────────────────────────────────────────────────────────────┤
│  Report Stage   │  Notification     │  Metrics Collection   │
│  - Error detail │  - PR comments    │  - Quality metrics    │
│  - Fix guidance │  - Status checks  │  - Trend analysis     │
│  - Performance │  - Team alerts    │  - Dashboard update   │
└─────────────────────────────────────────────────────────────┘
```

## コンポーネント詳細設計

### 1. Lint Processing Engine

**責務**: 静的解析処理の中核エンジン
**技術**: golangci-lint + カスタムラッパー

```go
type LintEngine struct {
    Config     *LintConfig
    Runners    []LintRunner
    Formatter  OutputFormatter
    Metrics    MetricsCollector
}

type LintRunner interface {
    Name() string
    Run(ctx context.Context, files []string) (*LintResult, error)
    SupportsIncrementalMode() bool
}
```

### 2. Configuration Manager

**責務**: lint設定の管理と動的調整
**技術**: YAML + Go struct validation

```go
type ConfigManager struct {
    BaseConfig      *BaseConfig
    EnvironmentOverrides map[string]*ConfigOverride
    ECSOptimizations    *ECSConfig
}

type ECSConfig struct {
    HotPathExclusions   []string
    PerformanceRules    map[string]interface{}
    EntitySystemPatterns []string
}
```

### 3. Quality Gate Controller

**責務**: 品質基準の判定と制御
**技術**: Rule-based decision engine

```go
type QualityGate struct {
    Rules       []QualityRule
    Thresholds  map[string]int
    Actions     map[string]Action
}

type QualityRule interface {
    Evaluate(result *LintResult) (bool, error)
    Severity() Severity
    Message() string
}
```

### 4. Performance Monitor

**責務**: lint処理パフォーマンスの監視
**技術**: Go metrics + Prometheus compatible

```go
type PerformanceMonitor struct {
    StartTime      time.Time
    MemoryTracker  *MemoryTracker  
    ExecutionTimer *Timer
    Profiler       *Profiler
}
```

## データフロー設計

### メインフロー: 開発時lint実行

```
Developer Action → File Change Detection → Incremental Lint → Quality Gate → Feedback Display
     │                    │                      │              │              │
     └── make lint ────────┴── Changed Files ────┴── Results ───┴── Pass/Fail ──┴── IDE/Terminal
```

### サブフロー: CI/CD統合

```
GitHub Event → Trigger Analysis → Full Lint Suite → Quality Assessment → Status Update
     │              │                   │                    │               │
     └── PR/Push ────┴── All Go Files ──┴── Complete Report ─┴── Gate Decision ─┴── PR Status
```

## 品質アーキテクチャ

### 1. 階層化品質チェック

```
┌─────────────────┐ ← 最高優先度
│  Security Layer │   gosec, G101-G602
├─────────────────┤
│  Correctness    │   errcheck, govet, staticcheck  
├─────────────────┤
│  Performance    │   ECS optimizations, prealloc
├─────────────────┤
│  Maintainability│   gocyclo, dupl, unused
├─────────────────┤
│  Style Layer    │   gofmt, goimports, revive
└─────────────────┘ ← 基本品質
```

### 2. ECS Framework特化設計

```go
type ECSOptimizer struct {
    EntityPatterns    []string  // Entity lifecycle patterns
    ComponentRules    []Rule    // Component memory layout rules  
    SystemConstraints []Rule    // System performance constraints
    QueryOptimizer    QueryOpt  // Query performance rules
}

// パフォーマンス臨界パス除外設定
var ECSHotPaths = []string{
    "internal/core/ecs/entity.go",
    "internal/core/ecs/component.go", 
    "internal/core/ecs/system.go",
    "internal/core/ecs/query.go",
}
```

## 統合アーキテクチャ

### IDE統合

```yaml
ide_integration:
  vscode:
    extensions:
      - golang.go
      - golangci.golangci-lint
    settings:
      go.lintTool: golangci-lint
      go.lintFlags: [--config=.golangci.yml]
      
  goland:
    inspections:
      golangci_lint: enabled
      custom_rules: enabled
```

### CI/CD統合

```yaml
github_actions:
  triggers:
    - pull_request: [opened, synchronize]
    - push: [main, develop]
  
  quality_gates:
    lint_errors: 0
    security_issues: 0
    performance_regression: false
    coverage_threshold: 80%
```

## 拡張性設計

### プラグインアーキテクチャ

```go
type LintPlugin interface {
    Name() string
    Version() string
    Initialize(config map[string]interface{}) error
    Process(ctx context.Context, files []string) (*PluginResult, error)
    Cleanup() error
}

type PluginManager struct {
    plugins   map[string]LintPlugin
    loader    PluginLoader
    registry  PluginRegistry
}
```

### カスタムルール設定

```yaml
custom_rules:
  ecs_entity_naming:
    pattern: "^[A-Z][a-zA-Z0-9]*Entity$"
    severity: error
    
  component_memory_layout:
    check_alignment: true
    max_size_bytes: 64
    severity: warning
    
  system_performance:
    max_entities_per_frame: 10000
    max_execution_time_ms: 16
    severity: error
```

## セキュリティアーキテクチャ

### セキュリティ検証フロー

```
Input Validation → Static Analysis → Dynamic Analysis → Report Generation
       │                 │               │                    │
   File paths ────── gosec rules ── Runtime checks ──── Security report
```

### 機密情報検出

```go
type SecretScanner struct {
    Patterns    []SecretPattern
    Allowlist   []string  
    Severity    map[string]Level
}

type SecretPattern struct {
    Name        string
    Regex       *regexp.Regexp
    Description string
    Remediation string
}
```

## パフォーマンス設計

### 並列処理アーキテクチャ

```go
type ParallelProcessor struct {
    WorkerPool    *WorkerPool
    FileQueue     chan string
    ResultChannel chan *LintResult
    ErrorHandler  ErrorHandler
}

// 目標性能: 5分以内、メモリ<2GB、CPU効率利用
var PerformanceTargets = struct{
    MaxExecutionTime time.Duration // 5m
    MaxMemoryUsage   int64          // 2GB  
    MaxWorkers       int            // runtime.NumCPU()
}{
    MaxExecutionTime: 5 * time.Minute,
    MaxMemoryUsage:   2 << 30,
    MaxWorkers:       runtime.NumCPU(),
}
```

## 監視・ロギング設計

### メトリクス収集

```go
type LintMetrics struct {
    ExecutionTime    histogram
    ErrorCount       counter  
    FileCount        gauge
    RuleViolations   counter_vec  // by rule type
    PerformanceStats summary
}

// OpenTelemetry compatible metrics
var MetricDefinitions = []MetricDefinition{
    {Name: "lint_execution_duration_seconds", Type: "histogram"},
    {Name: "lint_errors_total", Type: "counter"},
    {Name: "lint_files_processed", Type: "gauge"},
    {Name: "lint_rule_violations", Type: "counter", Labels: []string{"rule", "severity"}},
}
```

## 災害復旧・可用性

### エラーハンドリング戦略

```go
type ErrorStrategy struct {
    RetryPolicy    RetryConfig
    Fallback       FallbackAction
    CircuitBreaker CircuitBreakerConfig
}

type RetryConfig struct {
    MaxAttempts   int           // 3
    BackoffDelay  time.Duration // exponential
    Timeout       time.Duration // per attempt  
}
```

### 品質保証継続性

```yaml
continuity_plan:
  primary_system: golangci-lint
  backup_systems:
    - go_vet: basic checks
    - staticcheck: advanced analysis
    - custom_scripts: project-specific rules
    
  degraded_mode:
    essential_checks_only: true
    performance_timeout: 2m
    error_tolerance: higher
```

## 技術選定根拠

### Core Technologies

| 技術 | 選定理由 | 代替案との比較 |
|-----|----------|---------------|
| golangci-lint | 包括的、高性能、豊富なlinter統合 | revive単体: 機能不足, custom solution: 開発コスト高 |
| gofumpt | gofmtより厳密、一貫性向上 | gofmt: 基本的すぎる, prettier: Go非対応 |
| GitHub Actions | GitHub統合、無料、豊富なエコシステム | Jenkins: 保守負荷, GitLab CI: GitHub以外 |

### Performance Rationale

- **並列処理**: CPU集約的タスクのためのワーカープール
- **インクリメンタル**: 変更ファイルのみ処理で高速化
- **キャッシュ**: 解析結果のキャッシュで再実行高速化
- **メモリ効率**: ストリーミング処理でメモリ使用量制御

## 制約・前提条件

### 技術制約
- Go 1.23+ 必須
- golangci-lint v1.55+ 必須  
- GitHub Actions環境での実行
- Linux/macOS/Windows対応

### パフォーマンス制約
- 実行時間: 5分以内
- メモリ使用量: 2GB以下
- ECS性能劣化: 5%以下
- CI/CD実行時間: 追加3分以下

### 互換性制約
- 既存API互換性維持
- 既存設定ファイル互換性
- IDE統合継続性
- 開発者ワークフロー継続性