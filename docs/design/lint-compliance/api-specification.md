# Lint対応 API仕様書

## 概要

Muscle DreamerプロジェクトのLint対応システムは、内部API、CLI API、CI/CD統合API の3層構造で設計されています。この文書では、各APIレイヤーの詳細な仕様を定義します。

## API アーキテクチャ

```
┌─────────────────────┐
│    External APIs    │  ← CLI, CI/CD, IDE統合
├─────────────────────┤
│   Application APIs  │  ← アプリケーション層
├─────────────────────┤
│    Internal APIs    │  ← 内部コンポーネント間通信
└─────────────────────┘
```

## 1. CLI API 仕様

### 1.1 基本コマンド

#### `make lint`
基本的なlint実行コマンド

**実行例:**
```bash
make lint
```

**内部処理フロー:**
```bash
golangci-lint run --config=.golangci.yml --timeout=5m
```

**戻り値:**
- `0`: 成功（lint エラーなし）
- `1`: 失敗（lint エラーあり、または実行エラー）
- `2`: 設定エラー

#### `make format`
コードフォーマット実行コマンド

**実行例:**
```bash
make format
```

**内部処理フロー:**
```bash
gofumpt -w $(find . -name "*.go" -not -path "./docs/*")
goimports -w $(find . -name "*.go" -not -path "./docs/*")
```

### 1.2 拡張オプション

#### 環境変数による制御

```bash
# 詳細モード
LINT_VERBOSE=1 make lint

# 並列度指定
LINT_CONCURRENCY=8 make lint

# 特定ディレクトリのみ
LINT_PATH=internal/core make lint

# ドライランモード
LINT_DRY_RUN=1 make lint

# 自動修正モード
LINT_FIX=1 make lint

# CI/CDモード（簡潔な出力）
CI=1 make lint
```

#### カスタムコマンド拡張

```bash
# ECS固有のlintチェック
make lint-ecs

# セキュリティ特化チェック
make lint-security

# パフォーマンスチェック
make lint-performance

# 増分lint（変更ファイルのみ）
make lint-incremental

# lint結果のレポート生成
make lint-report
```

## 2. 内部Go API 仕様

### 2.1 Lint Engine API

#### LintEngine インターフェース

```go
package lint

import (
    "context"
    "time"
)

// LintEngine は lint処理の中核インターフェース
type LintEngine interface {
    // Run は指定されたオプションでlintを実行
    Run(ctx context.Context, opts *Options) (*Result, error)
    
    // RunIncremental は増分lintを実行
    RunIncremental(ctx context.Context, changedFiles []string, opts *Options) (*Result, error)
    
    // Validate は設定の妥当性を検証
    Validate(config *Config) error
    
    // GetSupportedLinters は対応リンター一覧を取得
    GetSupportedLinters() []LinterInfo
    
    // Close はリソースをクリーンアップ
    Close() error
}

// Options はlint実行オプション
type Options struct {
    ConfigPath     string
    WorkingDir     string
    Timeout        time.Duration
    Concurrency    int
    Fix            bool
    Format         bool
    OrganizeImports bool
    Environment    Environment
    OutputFormat   OutputFormat
    ProgressCallback func(*Progress)
}

// Result はlint実行結果
type Result struct {
    Summary    *ExecutionSummary
    Issues     []Issue
    Performance *PerformanceStats
    Metrics    *Metrics
    Success    bool
    ExitCode   int
}

// Issue は個別のlint問題
type Issue struct {
    ID          string
    Severity    Severity
    Rule        string
    Linter      string
    Category    Category
    File        string
    Line        int
    Column      int
    EndLine     int
    EndColumn   int
    Message     string
    Snippet     string
    Fix         *SuggestedFix
    Context     map[string]interface{}
}

// ExecutionSummary は実行サマリー
type ExecutionSummary struct {
    StartTime        time.Time
    EndTime          time.Time
    Duration         time.Duration
    FilesProcessed   int
    LinesProcessed   int
    IssuesFound      int
    IssuesBySeverity map[Severity]int
    LintersExecuted  []string
    Success          bool
    Error            string
}
```

#### 使用例

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "muscle-dreamer/internal/lint"
)

func main() {
    engine := lint.NewEngine()
    defer engine.Close()
    
    opts := &lint.Options{
        ConfigPath:  ".golangci.yml",
        WorkingDir:  ".",
        Timeout:     5 * time.Minute,
        Concurrency: 0, // auto-detect
        Fix:         false,
        Format:      true,
        OrganizeImports: true,
        Environment: lint.EnvironmentDevelopment,
        OutputFormat: lint.OutputFormatColoredLine,
        ProgressCallback: func(p *lint.Progress) {
            fmt.Printf("Progress: %d/%d files processed\n", 
                p.FilesProcessed, p.TotalFiles)
        },
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    result, err := engine.Run(ctx, opts)
    if err != nil {
        fmt.Printf("Lint execution failed: %v\n", err)
        return
    }
    
    fmt.Printf("Lint completed successfully\n")
    fmt.Printf("Files processed: %d\n", result.Summary.FilesProcessed)
    fmt.Printf("Issues found: %d\n", result.Summary.IssuesFound)
    fmt.Printf("Execution time: %v\n", result.Summary.Duration)
    
    for _, issue := range result.Issues {
        fmt.Printf("%s:%d:%d: %s [%s/%s]\n",
            issue.File, issue.Line, issue.Column,
            issue.Message, issue.Linter, issue.Rule)
    }
}
```

### 2.2 Configuration API

#### ConfigManager インターフェース

```go
package config

// ConfigManager は設定管理インターフェース
type ConfigManager interface {
    // Load は設定ファイルを読み込み
    Load(path string) (*Config, error)
    
    // LoadWithEnvironment は環境固有の設定を読み込み
    LoadWithEnvironment(path string, env Environment) (*Config, error)
    
    // Validate は設定の妥当性を検証
    Validate(config *Config) error
    
    // Save は設定をファイルに保存
    Save(config *Config, path string) error
    
    // GetDefault はデフォルト設定を取得
    GetDefault() *Config
    
    // Merge は複数の設定をマージ
    Merge(base *Config, overrides ...*ConfigOverride) *Config
}

// Config API エンドポイント
type ConfigAPI struct {
    manager ConfigManager
}

// REST風API設計
func (api *ConfigAPI) GetConfig(ctx context.Context, req *GetConfigRequest) (*GetConfigResponse, error)
func (api *ConfigAPI) UpdateConfig(ctx context.Context, req *UpdateConfigRequest) (*UpdateConfigResponse, error)  
func (api *ConfigAPI) ValidateConfig(ctx context.Context, req *ValidateConfigRequest) (*ValidateConfigResponse, error)
func (api *ConfigAPI) ResetConfig(ctx context.Context, req *ResetConfigRequest) (*ResetConfigResponse, error)
```

### 2.3 Quality Gate API

#### QualityGate インターフェース

```go
package qualitygate

// QualityGate は品質ゲートインターフェース  
type QualityGate interface {
    // Evaluate は品質評価を実行
    Evaluate(ctx context.Context, result *lint.Result) (*Assessment, error)
    
    // GetRules は品質ルール一覧を取得
    GetRules() []Rule
    
    // AddRule は品質ルールを追加
    AddRule(rule Rule) error
    
    // RemoveRule は品質ルールを削除
    RemoveRule(ruleID string) error
    
    // SetThreshold は閾値を設定
    SetThreshold(metric string, threshold interface{}) error
}

// Assessment は品質評価結果
type Assessment struct {
    Passed       bool
    Score        float64
    RuleResults  []RuleResult
    BlockingIssues []lint.Issue
    Recommendations []string
    Timestamp    time.Time
}

// Rule は品質ルール
type Rule struct {
    ID          string
    Name        string
    Description string
    Metric      MetricType
    Threshold   interface{}
    Operator    Operator
    Severity    RuleSeverity
    Weight      float64
    Enabled     bool
}

// RuleResult はルール評価結果
type RuleResult struct {
    RuleID       string
    Passed       bool
    ActualValue  interface{}
    Threshold    interface{}
    Weight       float64
    Message      string
}
```

### 2.4 Metrics Collection API

#### MetricsCollector インターフェース

```go
package metrics

// MetricsCollector はメトリクス収集インターフェース
type MetricsCollector interface {
    // Record はメトリクスを記録
    Record(name string, value float64, labels map[string]string)
    
    // RecordDuration は実行時間を記録
    RecordDuration(name string, duration time.Duration, labels map[string]string)
    
    // RecordCounter はカウンターをインクリメント
    RecordCounter(name string, labels map[string]string)
    
    // GetSummary は集計結果を取得
    GetSummary(timeRange TimeRange) (*Summary, error)
    
    // Export はメトリクスをエクスポート
    Export(ctx context.Context, exporter Exporter) error
    
    // Reset はメトリクスをリセット
    Reset()
}

// Exporter はメトリクスエクスポーター
type Exporter interface {
    Export(ctx context.Context, data *MetricsData) error
    SupportedFormats() []string
}

// 標準メトリクス定義
const (
    MetricLintExecutionTime    = "lint_execution_time_seconds"
    MetricLintIssuesTotal      = "lint_issues_total"
    MetricLintFilesProcessed   = "lint_files_processed"
    MetricLintMemoryUsage      = "lint_memory_usage_bytes"
    MetricLintErrorsTotal      = "lint_errors_total"
    MetricQualityScore         = "quality_score"
)
```

## 3. CI/CD統合API

### 3.1 GitHub Actions 統合

#### ワークフロー定義

```yaml
name: Lint Quality Check

on:
  pull_request:
    branches: [main, develop]
  push:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'
          
      - name: Cache dependencies  
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          
      - name: Install golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --config=.golangci.yml --timeout=5m
          
      - name: Run Lint Analysis
        run: make lint
        env:
          CI: "1"
          LINT_OUTPUT_FORMAT: "github-actions"
          
      - name: Quality Gate Check
        run: make lint-quality-gate
        
      - name: Upload Lint Results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: lint-results
          path: |
            lint-results.json
            lint-results.xml
            coverage.html
```

#### カスタムアクション定義

```yaml
# .github/actions/muscle-dreamer-lint/action.yml
name: 'Muscle Dreamer Lint'
description: 'Run comprehensive lint analysis for Muscle Dreamer'

inputs:
  config-path:
    description: 'Path to lint configuration file'
    required: false
    default: '.golangci.yml'
  
  timeout:
    description: 'Lint execution timeout'
    required: false
    default: '5m'
  
  fix-issues:
    description: 'Automatically fix issues when possible'
    required: false
    default: 'false'
  
  quality-gate:
    description: 'Enable quality gate checks'
    required: false
    default: 'true'

outputs:
  issues-count:
    description: 'Number of issues found'
    value: ${{ steps.lint.outputs.issues-count }}
  
  quality-score:
    description: 'Overall quality score'
    value: ${{ steps.lint.outputs.quality-score }}
  
  passed:
    description: 'Whether all checks passed'
    value: ${{ steps.lint.outputs.passed }}

runs:
  using: 'composite'
  steps:
    - name: Run Lint Analysis
      id: lint
      shell: bash
      run: |
        # Lint実行とメトリクス収集
        make lint LINT_CONFIG=${{ inputs.config-path }} LINT_TIMEOUT=${{ inputs.timeout }}
        
        # 結果の解析とアウトプット設定
        echo "issues-count=$(jq '.summary.issues_found' lint-results.json)" >> $GITHUB_OUTPUT
        echo "quality-score=$(jq '.quality_assessment.score' lint-results.json)" >> $GITHUB_OUTPUT
        echo "passed=$(jq '.quality_assessment.passed' lint-results.json)" >> $GITHUB_OUTPUT
```

### 3.2 HTTP API エンドポイント

#### REST API 仕様

```go
package api

import (
    "net/http"
    "github.com/gorilla/mux"
)

// Server はHTTP APIサーバー
type Server struct {
    engine LintEngine
    config ConfigManager
    metrics MetricsCollector
}

// Routes はAPIルートを定義
func (s *Server) Routes() http.Handler {
    r := mux.NewRouter()
    
    // Health Check
    r.HandleFunc("/health", s.HealthCheck).Methods("GET")
    
    // Configuration API
    r.HandleFunc("/api/v1/config", s.GetConfig).Methods("GET")
    r.HandleFunc("/api/v1/config", s.UpdateConfig).Methods("PUT")
    r.HandleFunc("/api/v1/config/validate", s.ValidateConfig).Methods("POST")
    
    // Lint Execution API
    r.HandleFunc("/api/v1/lint", s.RunLint).Methods("POST")
    r.HandleFunc("/api/v1/lint/incremental", s.RunIncrementalLint).Methods("POST")
    r.HandleFunc("/api/v1/lint/status/{jobId}", s.GetLintStatus).Methods("GET")
    
    // Quality Gate API
    r.HandleFunc("/api/v1/quality-gates", s.GetQualityGates).Methods("GET")
    r.HandleFunc("/api/v1/quality-gates/{id}/evaluate", s.EvaluateQualityGate).Methods("POST")
    
    // Metrics API
    r.HandleFunc("/api/v1/metrics", s.GetMetrics).Methods("GET")
    r.HandleFunc("/api/v1/metrics/export", s.ExportMetrics).Methods("POST")
    
    // Reporting API
    r.HandleFunc("/api/v1/reports/summary", s.GetSummaryReport).Methods("GET")
    r.HandleFunc("/api/v1/reports/detailed", s.GetDetailedReport).Methods("GET")
    
    return r
}
```

#### エンドポイント詳細仕様

##### POST /api/v1/lint
Lint分析の実行

**Request:**
```json
{
  "config_path": ".golangci.yml",
  "working_dir": ".",
  "timeout": "5m",
  "concurrency": 4,
  "fix": false,
  "format": true,
  "organize_imports": true,
  "environment": "development",
  "output_format": "json",
  "files": ["internal/core/ecs/*.go"],
  "incremental": false
}
```

**Response (202 Accepted):**
```json
{
  "job_id": "lint-job-12345",
  "status": "queued",
  "created_at": "2024-08-10T12:00:00Z",
  "estimated_completion": "2024-08-10T12:05:00Z"
}
```

##### GET /api/v1/lint/status/{jobId}
Lint分析の状態確認

**Response:**
```json
{
  "job_id": "lint-job-12345",
  "status": "completed",
  "progress": {
    "current_step": "analysis",
    "completed_steps": 4,
    "total_steps": 4,
    "percentage": 100,
    "files_processed": 125,
    "total_files": 125
  },
  "created_at": "2024-08-10T12:00:00Z",
  "started_at": "2024-08-10T12:00:05Z",
  "completed_at": "2024-08-10T12:03:42Z",
  "result": {
    "summary": {
      "duration": "3m37s",
      "files_processed": 125,
      "lines_processed": 15420,
      "issues_found": 23,
      "issues_by_severity": {
        "error": 2,
        "warning": 15,
        "info": 6
      },
      "success": false
    },
    "issues": [...],
    "performance": {...},
    "quality_assessment": {
      "passed": false,
      "score": 0.72,
      "blocking_issues": [...]
    }
  }
}
```

##### GET /api/v1/metrics
メトリクス取得

**Query Parameters:**
- `from`: 開始日時 (RFC3339)
- `to`: 終了日時 (RFC3339)  
- `granularity`: 粒度 (minute, hour, day)
- `metrics`: 取得するメトリクス名 (カンマ区切り)

**Response:**
```json
{
  "time_range": {
    "from": "2024-08-10T00:00:00Z",
    "to": "2024-08-10T23:59:59Z"
  },
  "granularity": "hour",
  "metrics": {
    "lint_execution_time_seconds": [
      {"timestamp": "2024-08-10T12:00:00Z", "value": 218.5},
      {"timestamp": "2024-08-10T13:00:00Z", "value": 195.2}
    ],
    "lint_issues_total": [
      {"timestamp": "2024-08-10T12:00:00Z", "value": 23},
      {"timestamp": "2024-08-10T13:00:00Z", "value": 18}
    ],
    "quality_score": [
      {"timestamp": "2024-08-10T12:00:00Z", "value": 0.72},
      {"timestamp": "2024-08-10T13:00:00Z", "value": 0.78}
    ]
  },
  "aggregates": {
    "lint_execution_time_seconds": {
      "min": 180.1,
      "max": 245.8,
      "avg": 205.3,
      "p95": 238.9
    },
    "lint_runs_total": 156,
    "successful_runs": 134,
    "success_rate": 0.859
  }
}
```

## 4. IDE統合API

### 4.1 Language Server Protocol (LSP) 統合

#### LSP Server Implementation

```go
package lsp

import (
    "context"
    "encoding/json"
)

// LintLanguageServer はLSP対応のlintサーバー
type LintLanguageServer struct {
    engine    LintEngine
    workspace string
    client    Client
}

// LSP Methods Implementation
func (s *LintLanguageServer) Initialize(ctx context.Context, params *InitializeParams) (*InitializeResult, error) {
    return &InitializeResult{
        Capabilities: ServerCapabilities{
            TextDocumentSync: &TextDocumentSyncOptions{
                OpenClose: true,
                Change:    TextDocumentSyncKindIncremental,
                Save:      &SaveOptions{IncludeText: false},
            },
            DiagnosticProvider: &DiagnosticOptions{
                Identifier:            "muscle-dreamer-lint",
                InterFileDependencies: true,
                WorkspaceDiagnostics:  true,
            },
            CodeActionProvider: &CodeActionOptions{
                CodeActionKinds: []CodeActionKind{
                    CodeActionQuickFix,
                    CodeActionSourceFixAll,
                },
                ResolveProvider: true,
            },
            ExecuteCommandProvider: &ExecuteCommandOptions{
                Commands: []string{
                    "muscle-dreamer.lint.run",
                    "muscle-dreamer.lint.fix",
                    "muscle-dreamer.lint.format",
                },
            },
        },
    }, nil
}

func (s *LintLanguageServer) DidSave(ctx context.Context, params *DidSaveTextDocumentParams) error {
    // ファイル保存時の自動lint実行
    go s.runLintForFile(params.TextDocument.URI)
    return nil
}

func (s *LintLanguageServer) Diagnostic(ctx context.Context, params *DocumentDiagnosticParams) (*DocumentDiagnosticReport, error) {
    // ファイル固有の診断情報を返す
    diagnostics, err := s.getDiagnosticsForFile(params.TextDocument.URI)
    if err != nil {
        return nil, err
    }
    
    return &DocumentDiagnosticReport{
        Kind:  DocumentDiagnosticReportKindFull,
        Items: diagnostics,
    }, nil
}

func (s *LintLanguageServer) CodeAction(ctx context.Context, params *CodeActionParams) ([]CodeAction, error) {
    // コード修正アクションを提供
    actions := make([]CodeAction, 0)
    
    for _, diagnostic := range params.Context.Diagnostics {
        if fix := s.getSuggestedFix(diagnostic); fix != nil {
            actions = append(actions, CodeAction{
                Title: fmt.Sprintf("Fix: %s", fix.Description),
                Kind:  CodeActionQuickFix,
                Edit: &WorkspaceEdit{
                    DocumentChanges: []TextDocumentEdit{
                        {
                            TextDocument: OptionalVersionedTextDocumentIdentifier{
                                URI: params.TextDocument.URI,
                            },
                            Edits: []TextEdit{
                                {
                                    Range:   diagnostic.Range,
                                    NewText: fix.Replacement,
                                },
                            },
                        },
                    },
                },
            })
        }
    }
    
    return actions, nil
}
```

### 4.2 VS Code Extension API

#### Extension Manifest

```json
{
  "name": "muscle-dreamer-lint",
  "displayName": "Muscle Dreamer Lint",
  "description": "Advanced Go linting for Muscle Dreamer ECS framework",
  "version": "1.0.0",
  "publisher": "muscle-dreamer",
  "engines": {
    "vscode": "^1.80.0"
  },
  "categories": ["Linters", "Other"],
  "activationEvents": [
    "onLanguage:go",
    "workspaceContains:**/.golangci.yml",
    "workspaceContains:**/CLAUDE.md"
  ],
  "contributes": {
    "commands": [
      {
        "command": "muscle-dreamer.lint.run",
        "title": "Run Lint Analysis",
        "category": "Muscle Dreamer"
      },
      {
        "command": "muscle-dreamer.lint.fix",
        "title": "Fix All Issues",
        "category": "Muscle Dreamer"
      },
      {
        "command": "muscle-dreamer.lint.configure",
        "title": "Configure Lint Settings",
        "category": "Muscle Dreamer"
      }
    ],
    "configuration": {
      "type": "object",
      "title": "Muscle Dreamer Lint",
      "properties": {
        "muscle-dreamer-lint.enabled": {
          "type": "boolean",
          "default": true,
          "description": "Enable Muscle Dreamer lint analysis"
        },
        "muscle-dreamer-lint.runOnSave": {
          "type": "boolean", 
          "default": true,
          "description": "Run lint analysis on file save"
        },
        "muscle-dreamer-lint.configPath": {
          "type": "string",
          "default": ".golangci.yml",
          "description": "Path to lint configuration file"
        }
      }
    }
  }
}
```

## 5. エラーハンドリングAPI

### 5.1 エラー定義

```go
package errors

import "fmt"

// エラーカテゴリー定義
type ErrorCategory string

const (
    CategoryConfiguration ErrorCategory = "configuration"
    CategoryExecution     ErrorCategory = "execution" 
    CategoryValidation    ErrorCategory = "validation"
    CategoryTimeout       ErrorCategory = "timeout"
    CategoryPermission    ErrorCategory = "permission"
    CategoryDependency    ErrorCategory = "dependency"
)

// LintError は統一されたlintエラー型
type LintError struct {
    Category    ErrorCategory
    Code        string
    Message     string
    Details     string
    Cause       error
    Recoverable bool
    Context     map[string]interface{}
}

func (e *LintError) Error() string {
    return fmt.Sprintf("[%s:%s] %s", e.Category, e.Code, e.Message)
}

// 事前定義エラー
var (
    ErrConfigNotFound = &LintError{
        Category:    CategoryConfiguration,
        Code:        "CONFIG_NOT_FOUND",
        Message:     "Lint configuration file not found",
        Recoverable: true,
    }
    
    ErrGolangciLintNotInstalled = &LintError{
        Category:    CategoryDependency,
        Code:        "GOLANGCI_NOT_INSTALLED", 
        Message:     "golangci-lint is not installed",
        Recoverable: true,
    }
    
    ErrExecutionTimeout = &LintError{
        Category:    CategoryTimeout,
        Code:        "EXECUTION_TIMEOUT",
        Message:     "Lint execution timeout",
        Recoverable: false,
    }
    
    ErrQualityGateFailed = &LintError{
        Category:    CategoryValidation,
        Code:        "QUALITY_GATE_FAILED",
        Message:     "Quality gate validation failed",
        Recoverable: false,
    }
)
```

### 5.2 Recovery API

```go
package recovery

// RecoveryStrategy はエラー回復戦略
type RecoveryStrategy interface {
    // CanRecover は回復可能かどうかを判定
    CanRecover(err error) bool
    
    // Recover はエラーからの回復を試行
    Recover(ctx context.Context, err error) error
    
    // GetFallback はフォールバック動作を取得
    GetFallback(err error) FallbackAction
}

// AutoRecovery は自動回復システム
type AutoRecovery struct {
    strategies []RecoveryStrategy
    retryPolicy RetryPolicy
    logger     Logger
}

func (ar *AutoRecovery) HandleError(ctx context.Context, err error) error {
    for _, strategy := range ar.strategies {
        if strategy.CanRecover(err) {
            return ar.retryWithStrategy(ctx, err, strategy)
        }
    }
    
    // 回復不可能な場合はフォールバック
    return ar.fallback(err)
}
```

## 6. パフォーマンス監視API

### 6.1 メトリクスAPI

```go
package monitoring

// PerformanceMonitor はパフォーマンス監視インターフェース
type PerformanceMonitor interface {
    // Start は監視を開始
    Start(ctx context.Context) error
    
    // Record はメトリクスを記録
    Record(metric Metric) error
    
    // GetStats は統計情報を取得
    GetStats(timeRange TimeRange) (*Stats, error)
    
    // SetAlert はアラート条件を設定
    SetAlert(condition AlertCondition) error
    
    // Stop は監視を停止
    Stop() error
}

// リアルタイム監視
type RealTimeMonitor struct {
    collectors []MetricCollector
    alerting   AlertManager
    storage    MetricStorage
}

func (m *RealTimeMonitor) collectMetrics(ctx context.Context) {
    ticker := time.NewTicker(time.Second * 10)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            for _, collector := range m.collectors {
                metrics, err := collector.Collect(ctx)
                if err != nil {
                    continue
                }
                
                for _, metric := range metrics {
                    if err := m.storage.Store(metric); err != nil {
                        continue
                    }
                    
                    // アラートチェック
                    if m.alerting.ShouldAlert(metric) {
                        m.alerting.TriggerAlert(metric)
                    }
                }
            }
        }
    }
}
```

### 6.2 アラートAPI

```go
package alerting

// AlertManager はアラート管理システム
type AlertManager interface {
    // AddRule はアラートルールを追加
    AddRule(rule AlertRule) error
    
    // TriggerAlert はアラートをトリガー
    TriggerAlert(metric Metric) error
    
    // GetActiveAlerts はアクティブなアラートを取得
    GetActiveAlerts() ([]Alert, error)
    
    // ResolveAlert はアラートを解決
    ResolveAlert(alertID string) error
}

// AlertRule はアラートルール
type AlertRule struct {
    ID          string
    Name        string
    Metric      string
    Condition   string
    Threshold   float64
    Duration    time.Duration
    Severity    AlertSeverity
    Actions     []AlertAction
    Enabled     bool
}

// アラート例
var DefaultAlerts = []AlertRule{
    {
        ID:        "lint-execution-timeout",
        Name:      "Lint Execution Timeout",
        Metric:    "lint_execution_time_seconds",
        Condition: ">",
        Threshold: 300, // 5 minutes
        Duration:  time.Second * 30,
        Severity:  SeverityWarning,
        Actions: []AlertAction{
            {Type: ActionTypeLog, Config: map[string]interface{}{"level": "warn"}},
            {Type: ActionTypeSlack, Config: map[string]interface{}{"channel": "#dev-alerts"}},
        },
        Enabled: true,
    },
    {
        ID:        "quality-score-degradation",
        Name:      "Quality Score Degradation",
        Metric:    "quality_score",
        Condition: "<",
        Threshold: 0.7,
        Duration:  time.Minute * 5,
        Severity:  SeverityError,
        Actions: []AlertAction{
            {Type: ActionTypeLog, Config: map[string]interface{}{"level": "error"}},
            {Type: ActionTypeEmail, Config: map[string]interface{}{"to": "team@muscle-dreamer.com"}},
        },
        Enabled: true,
    },
}
```

## 7. セキュリティAPI

### 7.1 認証・認可

```go
package security

// AuthProvider は認証プロバイダー
type AuthProvider interface {
    // Authenticate はユーザー認証
    Authenticate(ctx context.Context, token string) (*User, error)
    
    // Authorize は操作認可
    Authorize(ctx context.Context, user *User, operation Operation) error
    
    // GetPermissions は権限取得
    GetPermissions(ctx context.Context, user *User) ([]Permission, error)
}

// Operation は操作定義
type Operation struct {
    Resource string // "lint", "config", "metrics"
    Action   string // "read", "write", "execute"
    Context  map[string]interface{}
}

// Permission は権限定義
type Permission struct {
    Resource string
    Actions  []string
    Conditions map[string]interface{}
}
```

## 8. テストAPI

### 8.1 テスト用モック

```go
package testutil

// MockLintEngine はテスト用のモックエンジン
type MockLintEngine struct {
    RunFunc func(ctx context.Context, opts *Options) (*Result, error)
    ValidateFunc func(config *Config) error
    GetSupportedLintersFunc func() []LinterInfo
    CloseFunc func() error
}

func (m *MockLintEngine) Run(ctx context.Context, opts *Options) (*Result, error) {
    if m.RunFunc != nil {
        return m.RunFunc(ctx, opts)
    }
    return &Result{Success: true}, nil
}

// テストヘルパー
func NewTestLintEngine() *MockLintEngine {
    return &MockLintEngine{
        RunFunc: func(ctx context.Context, opts *Options) (*Result, error) {
            return &Result{
                Summary: &ExecutionSummary{
                    Duration: time.Millisecond * 100,
                    FilesProcessed: 10,
                    IssuesFound: 0,
                    Success: true,
                },
                Issues: []Issue{},
                Success: true,
                ExitCode: 0,
            }, nil
        },
        ValidateFunc: func(config *Config) error {
            return nil
        },
        GetSupportedLintersFunc: func() []LinterInfo {
            return []LinterInfo{
                {Name: "errcheck", Version: "1.0.0"},
                {Name: "gosec", Version: "2.0.0"},
            }
        },
        CloseFunc: func() error {
            return nil
        },
    }
}
```

## 9. API統合例

### 9.1 フルスタック使用例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "muscle-dreamer/internal/lint"
    "muscle-dreamer/internal/config"
    "muscle-dreamer/internal/qualitygate"
    "muscle-dreamer/internal/metrics"
)

func main() {
    ctx := context.Background()
    
    // 設定管理
    configMgr := config.NewManager()
    lintConfig, err := configMgr.LoadWithEnvironment(".golangci.yml", config.EnvironmentDevelopment)
    if err != nil {
        log.Fatal(err)
    }
    
    // Lintエンジン初期化
    engine := lint.NewEngine()
    defer engine.Close()
    
    // 品質ゲート初期化
    qg := qualitygate.NewQualityGate()
    
    // メトリクス収集器初期化
    metrics := metrics.NewCollector()
    
    // Lint実行オプション
    opts := &lint.Options{
        ConfigPath:      ".golangci.yml",
        WorkingDir:      ".",
        Timeout:         5 * time.Minute,
        Environment:     lint.EnvironmentDevelopment,
        Fix:            false,
        Format:         true,
        OrganizeImports: true,
        ProgressCallback: func(p *lint.Progress) {
            fmt.Printf("進行状況: %d/%d ファイル処理済み (%.1f%%)\n",
                p.FilesProcessed, p.TotalFiles, p.Percentage)
        },
    }
    
    // Lint実行
    fmt.Println("Lint分析を開始します...")
    result, err := engine.Run(ctx, opts)
    if err != nil {
        log.Printf("Lint実行エラー: %v", err)
        return
    }
    
    // 結果の表示
    fmt.Printf("分析完了: %d ファイル処理, %d 問題発見\n",
        result.Summary.FilesProcessed,
        result.Summary.IssuesFound)
    
    // 品質ゲート評価
    assessment, err := qg.Evaluate(ctx, result)
    if err != nil {
        log.Printf("品質ゲート評価エラー: %v", err)
    } else {
        fmt.Printf("品質スコア: %.2f (合格: %v)\n",
            assessment.Score, assessment.Passed)
    }
    
    // メトリクス記録
    metrics.RecordDuration("lint_execution_time", result.Summary.Duration, nil)
    metrics.Record("lint_issues_count", float64(result.Summary.IssuesFound), nil)
    metrics.Record("quality_score", assessment.Score, nil)
    
    // 結果に基づく終了処理
    if !result.Success {
        fmt.Println("Lint分析が失敗しました")
        for _, issue := range result.Issues {
            fmt.Printf("%s:%d:%d: %s [%s/%s]\n",
                issue.File, issue.Line, issue.Column,
                issue.Message, issue.Linter, issue.Rule)
        }
    }
    
    if !assessment.Passed {
        fmt.Println("品質ゲートを通過できませんでした")
        for _, rec := range assessment.Recommendations {
            fmt.Printf("推奨: %s\n", rec)
        }
    }
}
```

この包括的なAPI設計により、Lint対応システムは柔軟性、拡張性、保守性を備えた堅牢なシステムとして構築されます。