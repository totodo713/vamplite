# Lint対応 データフロー図

## システム全体データフロー

### 概要フロー図

```mermaid
flowchart TB
    Dev[👨‍💻 Developer] --> IDE[💻 IDE/Editor]
    Dev --> CLI[🖥️ Command Line]
    
    IDE --> Local[📁 Local Files]
    CLI --> Local
    
    Local --> Git[🗂️ Git Repository]
    Git --> GHA[⚙️ GitHub Actions]
    
    GHA --> LintEngine[🔧 Lint Engine]
    LintEngine --> QualityGate[🚪 Quality Gate]
    
    QualityGate --> Pass[✅ Pass]
    QualityGate --> Fail[❌ Fail]
    
    Pass --> Merge[🔀 Merge Ready]
    Fail --> Feedback[📝 Feedback]
    
    Feedback --> Dev
    
    classDef developer fill:#e1f5fe
    classDef system fill:#f3e5f5  
    classDef process fill:#e8f5e8
    classDef decision fill:#fff3e0
    classDef result fill:#fce4ec
    
    class Dev developer
    class IDE,CLI,Local,Git system
    class LintEngine,GHA process
    class QualityGate decision
    class Pass,Fail,Merge,Feedback result
```

## 詳細処理フロー

### 1. 開発時リアルタイムlint処理

```mermaid
sequenceDiagram
    participant D as Developer
    participant IDE as IDE/Editor
    participant FS as File System
    participant LE as Lint Engine
    participant QG as Quality Gate
    participant FB as Feedback UI
    
    D->>IDE: ファイル編集
    IDE->>FS: ファイル保存
    FS->>LE: ファイル変更検知
    
    Note over LE: インクリメンタル処理
    LE->>LE: 変更ファイル特定
    LE->>LE: golangci-lint実行
    LE->>LE: gofumpt/goimports実行
    
    LE->>QG: lint結果送信
    QG->>QG: 品質判定
    
    alt 品質OK
        QG->>FB: ✅ 成功メッセージ
        FB->>IDE: 緑色インジケータ
    else 品質NG  
        QG->>FB: ❌ エラー詳細
        FB->>IDE: 赤色インジケータ + エラーリスト
        FB->>D: 修正提案表示
    end
    
    Note over D,FB: 即座のフィードバック
```

### 2. コマンドライン実行フロー

```mermaid
flowchart TD
    MakeLint[make lint] --> CheckTool{golangci-lint<br/>インストール済み?}
    
    CheckTool -->|No| InstallTool[golangci-lint<br/>自動インストール]
    CheckTool -->|Yes| LoadConfig[設定ファイル読込]
    InstallTool --> LoadConfig
    
    LoadConfig --> ConfigValidation{設定検証}
    ConfigValidation -->|Invalid| ConfigError[❌ 設定エラー]
    ConfigValidation -->|Valid| FileDiscovery[対象ファイル探索]
    
    FileDiscovery --> FileFilter[ファイルフィルタ]
    FileFilter --> ParallelLint[並列lint実行]
    
    ParallelLint --> CollectResults[結果収集]
    CollectResults --> FormatOutput[出力フォーマット]
    
    FormatOutput --> QualityCheck{品質判定}
    QualityCheck -->|Pass| Success[✅ 成功終了]
    QualityCheck -->|Fail| DetailedReport[詳細レポート出力]
    
    DetailedReport --> FixSuggestion[修正提案生成]
    FixSuggestion --> ErrorExit[❌ エラー終了]
    
    Success --> ExitCode0[終了コード: 0]
    ErrorExit --> ExitCodeError[終了コード: 1]
    
    classDef success fill:#c8e6c9
    classDef error fill:#ffcdd2
    classDef process fill:#e1f5fe
    classDef decision fill:#fff3e0
    
    class Success,ExitCode0 success
    class ConfigError,DetailedReport,ErrorExit,ExitCodeError error
    class LoadConfig,FileDiscovery,ParallelLint,CollectResults process
    class CheckTool,ConfigValidation,QualityCheck decision
```

### 3. CI/CDパイプライン統合フロー

```mermaid
sequenceDiagram
    participant GH as GitHub
    participant GHA as GitHub Actions
    participant Runner as CI Runner
    participant LintSuite as Lint Suite
    participant Reporter as Report Generator
    participant PR as Pull Request
    
    GH->>GHA: PR/Push Event
    GHA->>Runner: Workflow開始
    
    Runner->>Runner: Go環境セットアップ
    Runner->>Runner: 依存関係インストール
    Runner->>Runner: golangci-lint取得
    
    Runner->>LintSuite: 全体lint実行開始
    
    Note over LintSuite: 段階的品質チェック
    LintSuite->>LintSuite: セキュリティチェック (gosec)
    LintSuite->>LintSuite: 正確性チェック (errcheck, govet)
    LintSuite->>LintSuite: パフォーマンスチェック (ECS最適化)
    LintSuite->>LintSuite: 保守性チェック (cyclo, dupl)
    LintSuite->>LintSuite: スタイルチェック (fmt, imports)
    
    LintSuite->>Reporter: 結果集計
    Reporter->>Reporter: レポート生成
    
    alt 全チェックPASS
        Reporter->>PR: ✅ 品質チェック通過
        Reporter->>GHA: 成功ステータス
    else 品質問題あり
        Reporter->>PR: ❌ 詳細エラーレポート
        Reporter->>GHA: 失敗ステータス
        Reporter->>GH: ブロック状態設定
    end
```

### 4. ECS Framework特化処理フロー

```mermaid
flowchart LR
    GoFiles[*.go ファイル] --> TypeDetect{ファイル種別判定}
    
    TypeDetect -->|Entity| EntityRules[Entity特化Rules]
    TypeDetect -->|Component| ComponentRules[Component特化Rules]  
    TypeDetect -->|System| SystemRules[System特化Rules]
    TypeDetect -->|Query| QueryRules[Query特化Rules]
    TypeDetect -->|General| GeneralRules[一般Rules]
    
    EntityRules --> EntityOptim[Entity最適化チェック]
    ComponentRules --> ComponentOptim[メモリレイアウト最適化]
    SystemRules --> SystemOptim[パフォーマンス最適化]
    QueryRules --> QueryOptim[クエリ効率チェック]
    GeneralRules --> GeneralOptim[一般品質チェック]
    
    EntityOptim --> PerformanceGate{性能要件チェック}
    ComponentOptim --> PerformanceGate
    SystemOptim --> PerformanceGate  
    QueryOptim --> PerformanceGate
    GeneralOptim --> QualityGate{品質要件チェック}
    
    PerformanceGate -->|OK| QualityGate
    PerformanceGate -->|NG| PerformanceAlert[⚠️ 性能劣化警告]
    
    QualityGate -->|OK| ECSCompliant[✅ ECS準拠]
    QualityGate -->|NG| QualityIssue[❌ 品質問題]
    
    PerformanceAlert --> FixRecommend[修正推奨提示]
    QualityIssue --> FixRecommend
    
    classDef ecsfile fill:#e8f5e8
    classDef rules fill:#e1f5fe
    classDef optim fill:#fff3e0
    classDef gate fill:#f3e5f5
    classDef result fill:#fce4ec
    
    class GoFiles ecsfile
    class EntityRules,ComponentRules,SystemRules,QueryRules,GeneralRules rules
    class EntityOptim,ComponentOptim,SystemOptim,QueryOptim,GeneralOptim optim
    class PerformanceGate,QualityGate gate
    class ECSCompliant,QualityIssue,PerformanceAlert result
```

## データ変換フロー

### 5. lint結果データ変換パイプライン

```mermaid
flowchart TD
    RawOutput[golangci-lint<br/>Raw Output] --> Parser[結果パーサー]
    
    Parser --> Normalizer[データ正規化]
    Normalizer --> Enricher[メタデータ付加]
    
    Enricher --> Categorizer[カテゴライザー]
    Categorizer --> SecurityIssues[🔒 セキュリティ課題]
    Categorizer --> PerformanceIssues[⚡ パフォーマンス課題] 
    Categorizer --> QualityIssues[📋 品質課題]
    Categorizer --> StyleIssues[🎨 スタイル課題]
    
    SecurityIssues --> PriorityAssigner[優先度割当]
    PerformanceIssues --> PriorityAssigner
    QualityIssues --> PriorityAssigner
    StyleIssues --> PriorityAssigner
    
    PriorityAssigner --> ContextEnricher[コンテキスト付加]
    ContextEnricher --> FixGenerator[修正提案生成]
    
    FixGenerator --> OutputFormatter[出力フォーマッタ]
    
    OutputFormatter --> ConsoleOutput[🖥️ コンソール出力]
    OutputFormatter --> IDEOutput[💻 IDE統合出力]
    OutputFormatter --> CIOutput[⚙️ CI/CD出力]
    OutputFormatter --> MetricsOutput[📊 メトリクス出力]
    
    classDef input fill:#e3f2fd
    classDef process fill:#f1f8e9
    classDef category fill:#fff8e1
    classDef output fill:#fce4ec
    
    class RawOutput input
    class Parser,Normalizer,Enricher,PriorityAssigner,ContextEnricher,FixGenerator,OutputFormatter process
    class SecurityIssues,PerformanceIssues,QualityIssues,StyleIssues category
    class ConsoleOutput,IDEOutput,CIOutput,MetricsOutput output
```

### 6. パフォーマンス監視データフロー

```mermaid
sequenceDiagram
    participant LintStart as Lint開始
    participant Monitor as Performance Monitor
    participant Collector as Metrics Collector
    participant Analyzer as Performance Analyzer
    participant Alert as Alert System
    participant Dashboard as Dashboard
    
    LintStart->>Monitor: 処理開始通知
    Monitor->>Monitor: タイマー開始
    Monitor->>Monitor: メモリ使用量測定開始
    
    loop Lint処理中
        Monitor->>Collector: リアルタイムメトリクス送信
        Collector->>Analyzer: データ蓄積・分析
        
        Note over Analyzer: 性能閾値チェック
        Analyzer->>Analyzer: 実行時間監視 (< 5分)
        Analyzer->>Analyzer: メモリ使用量監視 (< 2GB)
        Analyzer->>Analyzer: CPU使用率監視
        
        alt 閾値超過
            Analyzer->>Alert: ⚠️ 性能アラート
            Alert->>Dashboard: アラート表示
        end
    end
    
    LintStart->>Monitor: 処理完了通知
    Monitor->>Collector: 最終メトリクス送信
    Collector->>Dashboard: パフォーマンス報告更新
    
    Note over Dashboard: 実行時間、メモリ、エラー率等を可視化
```

## エラーハンドリングフロー

### 7. 例外・エラー処理フロー

```mermaid
flowchart TD
    LintExecution[Lint実行] --> ErrorDetection{エラー検出}
    
    ErrorDetection -->|システムエラー| SystemError[システムエラー]
    ErrorDetection -->|設定エラー| ConfigError[設定エラー]
    ErrorDetection -->|ツールエラー| ToolError[ツールエラー]
    ErrorDetection -->|タイムアウト| TimeoutError[タイムアウトエラー]
    ErrorDetection -->|正常| Success[正常終了]
    
    SystemError --> RetryLogic{リトライ可能?}
    ConfigError --> ConfigValidation[設定検証]
    ToolError --> ToolRecovery[ツール復旧処理]
    TimeoutError --> PartialResult[部分結果取得]
    
    RetryLogic -->|Yes| RetryExecution[リトライ実行]
    RetryLogic -->|No| FallbackMode[フォールバックモード]
    
    RetryExecution -->|成功| Success
    RetryExecution -->|失敗| FallbackMode
    
    ConfigValidation --> ConfigFix{自動修正可能?}
    ConfigFix -->|Yes| AutoFix[自動修正実行]
    ConfigFix -->|No| ManualFix[手動修正要求]
    
    ToolRecovery --> ToolReinstall[ツール再インストール]
    ToolReinstall --> RetryExecution
    
    PartialResult --> PartialReport[部分レポート生成]
    FallbackMode --> MinimalCheck[最小限チェック実行]
    
    AutoFix --> Success
    ManualFix --> ErrorReport[エラーレポート生成]
    PartialReport --> WarningExit[警告付き終了]
    MinimalCheck --> DegradedSuccess[機能縮退成功]
    
    classDef error fill:#ffcdd2
    classDef recovery fill:#fff3e0
    classDef success fill:#c8e6c9
    classDef warning fill:#ffecb3
    
    class SystemError,ConfigError,ToolError,TimeoutError,ErrorReport error
    class RetryLogic,ConfigValidation,ToolRecovery,RetryExecution,AutoFix recovery
    class Success,DegradedSuccess success  
    class PartialResult,PartialReport,WarningExit warning
```

## メトリクス収集・分析フロー

### 8. 品質メトリクス収集フロー

```mermaid
graph TB
    subgraph "データ収集層"
        LintResults[Lint実行結果] --> MetricExtractor[メトリクス抽出器]
        ExecutionStats[実行統計] --> MetricExtractor
        FileStats[ファイル統計] --> MetricExtractor
    end
    
    subgraph "データ処理層"
        MetricExtractor --> Aggregator[データ集約器]
        Aggregator --> Normalizer[正規化処理]
        Normalizer --> Calculator[指標計算器]
    end
    
    subgraph "分析層"
        Calculator --> TrendAnalyzer[トレンド分析]
        Calculator --> ThresholdChecker[閾値チェック]
        Calculator --> Comparator[比較分析器]
    end
    
    subgraph "出力層"
        TrendAnalyzer --> Dashboard[📊 ダッシュボード]
        ThresholdChecker --> AlertSystem[🚨 アラートシステム]
        Comparator --> Report[📋 品質レポート]
        
        Dashboard --> Visualization[可視化表示]
        AlertSystem --> Notification[通知送信]
        Report --> Documentation[ドキュメント生成]
    end
    
    classDef input fill:#e8f5e8
    classDef process fill:#e1f5fe
    classDef analysis fill:#fff3e0
    classDef output fill:#fce4ec
    
    class LintResults,ExecutionStats,FileStats input
    class MetricExtractor,Aggregator,Normalizer,Calculator process
    class TrendAnalyzer,ThresholdChecker,Comparator analysis
    class Dashboard,AlertSystem,Report,Visualization,Notification,Documentation output
```

## データ永続化・キャッシュフロー

### 9. キャッシュ戦略フロー

```mermaid
flowchart LR
    Request[lint実行要求] --> CacheCheck{キャッシュ確認}
    
    CacheCheck -->|HIT| CacheValid{キャッシュ有効?}
    CacheCheck -->|MISS| FullAnalysis[完全解析実行]
    
    CacheValid -->|Valid| CacheResult[キャッシュ結果返却]
    CacheValid -->|Invalid| IncrementalAnalysis[差分解析実行]
    
    FullAnalysis --> StoreCache[キャッシュ保存]
    IncrementalAnalysis --> UpdateCache[キャッシュ更新]
    
    StoreCache --> Result[結果返却]
    UpdateCache --> Result
    CacheResult --> Result
    
    subgraph "キャッシュ判定条件"
        FileModTime[ファイル更新時刻]
        ConfigHash[設定ファイルハッシュ]
        ToolVersion[ツールバージョン]
        RuleSet[適用ルールセット]
    end
    
    FileModTime --> CacheValid
    ConfigHash --> CacheValid
    ToolVersion --> CacheValid
    RuleSet --> CacheValid
    
    classDef cache fill:#e1f5fe
    classDef analysis fill:#f1f8e9
    classDef condition fill:#fff3e0
    classDef result fill:#e8f5e8
    
    class CacheCheck,CacheValid,CacheResult,StoreCache,UpdateCache cache
    class FullAnalysis,IncrementalAnalysis analysis
    class FileModTime,ConfigHash,ToolVersion,RuleSet condition
    class Request,Result result
```

この設計により、効率的で拡張性のあるlint処理システムが実現されます。各フローは独立性を保ちつつ、全体として統合されたエクスペリエンスを提供します。