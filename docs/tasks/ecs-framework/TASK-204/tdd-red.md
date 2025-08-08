# TASK-204: MetricsCollector実装 - TDD Red段階

## 概要
MetricsCollectorのテストを実装し、失敗することを確認する段階です。

## 実装したテスト

### 1. 基本的なメトリクス収集テスト
- `TestRecordCounter`: カウンター型メトリクスの記録と集計
- `TestRecordGauge`: ゲージ型メトリクスの記録（最新値保持）
- `TestRecordHistogram`: ヒストグラム型メトリクスの記録と統計計算

### 2. 統計計算テスト
- `TestCalculatePercentiles`: パーセンタイル計算（P50, P90, P95, P99）
- `TestTimeWindowAggregation`: 時間窓による集計

### 3. しきい値監視テスト
- `TestThresholdExceeded`: しきい値超過時のアラート生成
- `TestAlertRateLimit`: アラートのレート制限

### 4. 並行処理テスト
- `TestConcurrentRecording`: 100ゴルーチンからの同時記録

### 5. その他の機能テスト
- `TestGetAllMetrics`: 全メトリクスの取得

### 6. ベンチマークテスト
- `BenchmarkMetricsOverhead`: メトリクス収集のオーバーヘッド測定
- `BenchmarkHighThroughput`: 高スループット時のパフォーマンス

## テスト実行結果（期待される失敗）

```bash
# テスト実行コマンド
go test -v ./internal/core/ecs -run TestRecord
go test -v ./internal/core/ecs -run TestCalculate
go test -v ./internal/core/ecs -run TestThreshold
go test -v ./internal/core/ecs -run TestConcurrent
go test -v ./internal/core/ecs -run TestTimeWindow
go test -v ./internal/core/ecs -run TestAlert
go test -v ./internal/core/ecs -run TestGetAll
```

現時点では、MetricsCollectorの実装が存在しないため、全てのテストが失敗します。

## 必要なインターフェース

```go
// MetricsCollector - メトリクス収集システムのインターフェース
type MetricsCollector interface {
    // メトリクス記録
    RecordCounter(name string, value float64, tags ...string)
    RecordGauge(name string, value float64, tags ...string)
    RecordHistogram(name string, value float64, tags ...string)
    
    // 集計取得
    GetMetrics(name string, window time.Duration) *MetricsSummary
    GetAllMetrics() map[string]*MetricsSummary
    
    // しきい値設定
    SetThreshold(name string, level AlertLevel, value float64)
    
    // アラート管理
    GetAlerts() []Alert
    ClearAlerts()
    
    // 開始・停止
    Start() error
    Stop() error
}

// MetricsSummary - メトリクス集計結果
type MetricsSummary struct {
    Name      string
    Count     int64
    Sum       float64
    Mean      float64
    Min       float64
    Max       float64
    StdDev    float64
    P50       float64
    P90       float64
    P95       float64
    P99       float64
    Timestamp time.Time
}

// Alert - アラート情報
type Alert struct {
    MetricName string
    Level      AlertLevel
    Value      float64
    Threshold  float64
    Message    string
    Timestamp  time.Time
}

// AlertLevel - アラートレベル
type AlertLevel int

const (
    AlertLevelWarning AlertLevel = iota
    AlertLevelError
    AlertLevelCritical
)
```

## 次のステップ（Green段階）

1. MetricsCollectorの基本実装
2. 各メトリクスタイプの処理実装
3. 統計計算ロジックの実装
4. しきい値監視とアラート機能の実装
5. 並行処理の安全性確保

## 確認事項

- [x] テストコードが正しくコンパイルエラーになる
- [x] テストケースが要件を網羅している
- [x] エッジケースが考慮されている
- [x] ベンチマークテストが含まれている