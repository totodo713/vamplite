# TASK-204: MetricsCollector実装 - テストケース仕様

## テストケース一覧

### 1. 基本的なメトリクス収集テスト

#### TC-001: カウンター型メトリクスの記録
```go
func TestRecordCounter(t *testing.T) {
    // Given: MetricsCollectorのインスタンス
    // When: カウンターメトリクスを記録
    // Then: メトリクスが正しく保存される
}
```

#### TC-002: ゲージ型メトリクスの記録
```go
func TestRecordGauge(t *testing.T) {
    // Given: MetricsCollectorのインスタンス
    // When: ゲージメトリクスを記録
    // Then: 最新の値が保持される
}
```

#### TC-003: ヒストグラム型メトリクスの記録
```go
func TestRecordHistogram(t *testing.T) {
    // Given: MetricsCollectorのインスタンス
    // When: 複数の値を記録
    // Then: 分布データが正しく保存される
}
```

### 2. メトリクス集計テスト

#### TC-004: 基本統計の計算
```go
func TestCalculateBasicStats(t *testing.T) {
    // Given: 複数のメトリクス値
    // When: 統計計算を実行
    // Then: 平均、最小、最大、標準偏差が正しく計算される
    
    // テストデータ: [1, 2, 3, 4, 5]
    // 期待値:
    // - Mean: 3.0
    // - Min: 1.0
    // - Max: 5.0
    // - StdDev: 1.414
}
```

#### TC-005: パーセンタイル計算
```go
func TestCalculatePercentiles(t *testing.T) {
    // Given: 100個のサンプルデータ
    // When: パーセンタイル計算を実行
    // Then: P50, P90, P95, P99が正しく計算される
}
```

#### TC-006: 時間窓による集計
```go
func TestTimeWindowAggregation(t *testing.T) {
    // Given: 異なる時刻のメトリクス
    // When: 1秒窓で集計
    // Then: 指定時間内のメトリクスのみ集計される
}
```

### 3. しきい値監視とアラートテスト

#### TC-007: しきい値超過検出
```go
func TestThresholdExceeded(t *testing.T) {
    // Given: しきい値設定（Warning: 80, Error: 90）
    // When: 値95を記録
    // Then: Errorアラートが生成される
}
```

#### TC-008: アラート履歴管理
```go
func TestAlertHistory(t *testing.T) {
    // Given: 複数のアラート発生
    // When: アラート履歴を取得
    // Then: 全アラートが時系列順に取得できる
}
```

#### TC-009: アラートレート制限
```go
func TestAlertRateLimit(t *testing.T) {
    // Given: レート制限設定（1アラート/分）
    // When: 短時間に複数のしきい値超過
    // Then: 最初のアラートのみ生成される
}
```

### 4. 並行処理テスト

#### TC-010: 並行メトリクス記録
```go
func TestConcurrentRecording(t *testing.T) {
    // Given: 100個のゴルーチン
    // When: 同時にメトリクスを記録
    // Then: データ競合なく全メトリクスが記録される
}
```

#### TC-011: 並行読み書き
```go
func TestConcurrentReadWrite(t *testing.T) {
    // Given: 書き込みと読み込みのゴルーチン
    // When: 同時実行
    // Then: データ整合性が保たれる
}
```

### 5. パフォーマンステスト

#### TC-012: 監視オーバーヘッド測定
```go
func BenchmarkMetricsOverhead(b *testing.B) {
    // Given: 通常のシステム実行
    // When: メトリクス収集あり/なしで比較
    // Then: オーバーヘッド < 1%
}
```

#### TC-013: 大量メトリクス処理
```go
func BenchmarkHighThroughput(b *testing.B) {
    // Given: 10,000メトリクス/秒の負荷
    // When: 継続的に記録
    // Then: 処理遅延なし
}
```

#### TC-014: メモリ使用量測定
```go
func TestMemoryUsage(t *testing.T) {
    // Given: 1時間分のメトリクス
    // When: 継続的に記録
    // Then: メモリ使用量 < 10MB
}
```

### 6. エラーハンドリングテスト

#### TC-015: 無効なメトリクス名
```go
func TestInvalidMetricName(t *testing.T) {
    // Given: 空文字列や特殊文字を含む名前
    // When: メトリクス記録を試行
    // Then: エラーが返される
}
```

#### TC-016: メモリ不足時の処理
```go
func TestMemoryPressure(t *testing.T) {
    // Given: メモリ制限に近い状態
    // When: 新規メトリクス記録
    // Then: 古いデータが削除される
}
```

### 7. 統合テスト

#### TC-017: ECSシステムとの統合
```go
func TestECSIntegration(t *testing.T) {
    // Given: 実行中のECSシステム
    // When: システムメトリクスを自動収集
    // Then: フレーム時間、エンティティ数等が正しく記録される
}
```

#### TC-018: 複数メトリクスタイプの混在
```go
func TestMixedMetricTypes(t *testing.T) {
    // Given: カウンター、ゲージ、ヒストグラムの混在
    // When: 同時に記録と集計
    // Then: 各タイプが正しく処理される
}
```

## エッジケース

### EC-001: ゼロ除算の回避
- 空のメトリクスセットでの統計計算
- 期待動作: デフォルト値またはエラーを返す

### EC-002: タイムスタンプの巻き戻り
- システム時刻が過去に戻る場合
- 期待動作: 単調増加時刻を使用

### EC-003: 極端な値の処理
- NaN、Inf、極大値/極小値
- 期待動作: 適切にフィルタリングまたは制限

### EC-004: 高頻度アラート
- 値が境界値付近で振動
- 期待動作: ヒステリシスまたはデバウンス処理

## テスト実行計画

### フェーズ1: 単体テスト
1. 基本的なメトリクス収集（TC-001〜TC-003）
2. 統計計算（TC-004〜TC-006）
3. エラーハンドリング（TC-015〜TC-016）

### フェーズ2: 機能テスト
1. しきい値監視（TC-007〜TC-009）
2. 並行処理（TC-010〜TC-011）

### フェーズ3: 非機能テスト
1. パフォーマンス（TC-012〜TC-014）
2. 統合テスト（TC-017〜TC-018）

## 期待される品質基準

- **テストカバレッジ**: > 90%
- **パフォーマンス**: 全ベンチマークでターゲット達成
- **並行処理**: データ競合ゼロ（-race フラグでパス）
- **メモリリーク**: ゼロ（pprof で確認）