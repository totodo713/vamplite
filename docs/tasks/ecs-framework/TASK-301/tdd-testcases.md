# TASK-301: ModECSAPI実装 - テストケース仕様

## テストケース概要

MOECSAPIの各機能に対する包括的なテストケースを定義します。セキュリティ、パフォーマンス、機能性の3つの観点から検証します。

## 単体テストケース

### TC-301-001: ModEntityAPI基本機能テスト

#### TC-301-001-001: エンティティ作成テスト
```go
func TestModEntityAPI_Create(t *testing.T) {
    // 正常ケース: エンティティ作成成功
    // 制限ケース: 作成上限（100個）に達した場合のエラー
    // タグ付与: mod:{mod_name}タグが自動付与される
}
```

#### TC-301-001-002: エンティティ削除テスト  
```go
func TestModEntityAPI_Delete(t *testing.T) {
    // 正常ケース: 自分が作成したエンティティの削除
    // 拒否ケース: 他MODエンティティの削除試行
    // 拒否ケース: システムエンティティの削除試行
}
```

#### TC-301-001-003: エンティティタグ取得テスト
```go
func TestModEntityAPI_GetTags(t *testing.T) {
    // 正常ケース: 自分のエンティティのタグ取得
    // 制限ケース: 他エンティティのタグ取得時の制限情報
}
```

### TC-301-002: ModComponentAPI機能テスト

#### TC-301-002-001: コンポーネント追加テスト
```go
func TestModComponentAPI_Add(t *testing.T) {
    // 正常ケース: 許可されたコンポーネント追加
    // 拒否ケース: 禁止コンポーネント追加試行（FileIOComponent等）
    // 拒否ケース: 他MODエンティティへのコンポーネント追加試行
}
```

#### TC-301-002-002: コンポーネント取得テスト
```go
func TestModComponentAPI_Get(t *testing.T) {
    // 正常ケース: 自分のコンポーネント取得
    // 制限ケース: 読み取り専用コンポーネント取得
    // 拒否ケース: アクセス権限のないコンポーネント取得試行
}
```

#### TC-301-002-003: コンポーネント削除テスト
```go
func TestModComponentAPI_Remove(t *testing.T) {
    // 正常ケース: 自分が追加したコンポーネント削除
    // 拒否ケース: システムコンポーネント削除試行
    // 拒否ケース: 他MODコンポーネント削除試行
}
```

### TC-301-003: ModQueryAPI機能テスト

#### TC-301-003-001: 基本クエリテスト
```go
func TestModQueryAPI_Find(t *testing.T) {
    // 正常ケース: MODエンティティのみの検索結果
    // フィルタリング: システムエンティティが結果から除外される
    // フィルタリング: 他MODエンティティが結果から除外される
}
```

#### TC-301-003-002: クエリ実行制限テスト
```go
func TestModQueryAPI_ExecutionLimits(t *testing.T) {
    // 制限ケース: 1000回/フレーム上限での拒否
    // 制限ケース: 10ms時間制限での中断
    // エラーハンドリング: 制限超過時の適切なエラー返却
}
```

### TC-301-004: ModSystemAPI機能テスト

#### TC-301-004-001: システム登録テスト
```go
func TestModSystemAPI_Register(t *testing.T) {
    // 正常ケース: 有効なMODシステム登録
    // 拒否ケース: 危険な操作を含むシステム登録試行
}
```

#### TC-301-004-002: システム実行制限テスト
```go
func TestModSystemAPI_ExecutionLimits(t *testing.T) {
    // 制限ケース: 5ms/フレーム実行時間制限
    // 制限ケース: 10MBメモリ使用制限
    // 隔離ケース: システム実行エラーの隔離
}
```

## セキュリティテストケース

### TC-301-SEC-001: パストラバーサル攻撃防御
```go
func TestSecurityPathTraversal(t *testing.T) {
    // 攻撃パターン: ../../../etc/passwd
    // 攻撃パターン: ..\\..\\..\\windows\\system32
    // 検証: 全てのパストラバーサル試行が確実にブロックされる
}
```

### TC-301-SEC-002: システムコマンド実行防止
```go
func TestSecuritySystemCommands(t *testing.T) {
    // 攻撃パターン: rm -rf /
    // 攻撃パターン: exec.Command() 使用試行
    // 検証: システムコマンド実行が100%ブロックされる
}
```

### TC-301-SEC-003: メモリ破壊攻撃防御
```go
func TestSecurityMemoryProtection(t *testing.T) {
    // 攻撃パターン: バッファオーバーフロー試行
    // 攻撃パターン: 不正メモリアクセス試行
    // 検証: メモリ保護機能が正常動作する
}
```

### TC-301-SEC-004: DoS攻撃防御
```go
func TestSecurityDoSProtection(t *testing.T) {
    // 攻撃パターン: 無限ループ実行
    // 攻撃パターン: 大量メモリ確保試行
    // 検証: リソース制限によってDoSが軽減される
}
```

## パフォーマンステストケース

### TC-301-PERF-001: API呼び出しオーバーヘッド
```go
func BenchmarkModAPIOverhead(b *testing.B) {
    // 測定: Entity作成API呼び出し時間 (目標: <100μs)
    // 測定: Component操作API呼び出し時間 (目標: <100μs)
    // 測定: Query実行API呼び出し時間 (目標: <100μs)
}
```

### TC-301-PERF-002: サンドボックス実行オーバーヘッド
```go
func BenchmarkSandboxOverhead(b *testing.B) {
    // 測定: サンドボックス有無でのシステム実行時間差
    // 目標: オーバーヘッド<5%
}
```

### TC-301-PERF-003: 大量MOD同時実行
```go
func TestMassiveMODExecution(t *testing.T) {
    // シナリオ: 10個のMODが同時に複雑な操作実行
    // 検証: システム全体のパフォーマンス維持
    // 検証: メモリ使用量制限遵守
}
```

## 統合テストケース

### TC-301-INT-001: ECSコアシステム統合
```go
func TestECSCoreIntegration(t *testing.T) {
    // シナリオ: MODがEntityManager経由でエンティティ操作
    // シナリオ: MODがComponentStore経由でコンポーネント操作  
    // シナリオ: MODがSystemManager経由でシステム登録・実行
    // 検証: 制限付きアクセスが正常動作する
}
```

### TC-301-INT-002: EventBus統合
```go
func TestEventBusIntegration(t *testing.T) {
    // シナリオ: MODがイベント送信・受信
    // 制限: MOD間イベント分離
    // 制限: システムイベントへのアクセス制限
}
```

### TC-301-INT-003: MetricsCollector統合
```go
func TestMetricsIntegration(t *testing.T) {
    // シナリオ: MOD API使用状況の監視
    // シナリオ: セキュリティ違反の記録
    // 検証: メトリクス収集が正常動作する
}
```

## エラーケーステスト

### TC-301-ERR-001: 権限エラーハンドリング
```go
func TestPermissionErrors(t *testing.T) {
    // エラーケース: 権限のないエンティティアクセス試行
    // エラーケース: 禁止コンポーネント操作試行
    // 検証: 適切なエラーメッセージ返却
}
```

### TC-301-ERR-002: リソース制限エラー
```go
func TestResourceLimitErrors(t *testing.T) {
    // エラーケース: エンティティ作成上限超過
    // エラーケース: メモリ使用量制限超過  
    // エラーケース: 実行時間制限超過
    // 検証: 制限超過時の適切な処理
}
```

### TC-301-ERR-003: システムエラー分離
```go
func TestSystemErrorIsolation(t *testing.T) {
    // エラーケース: MODシステム実行時エラー
    // 検証: エラーがコアシステムに波及しない
    // 検証: ゲーム全体の継続動作
}
```

## 長期実行テストケース

### TC-301-LONG-001: 24時間安定性テスト
```go
func TestLongTermStability(t *testing.T) {
    // シナリオ: 10個のMODを24時間連続実行
    // 監視: メモリリーク発生状況
    // 監視: セキュリティ違反発生状況
    // 目標: メモリリーク<50MB/24h
}
```

## テスト実行手順

### 1. 単体テスト実行
```bash
go test ./internal/core/ecs/mod/ -v
go test ./internal/core/ecs/mod/ -race
go test ./internal/core/ecs/mod/ -cover
```

### 2. セキュリティテスト実行
```bash
go test ./internal/core/ecs/mod/ -v -run TestSecurity
```

### 3. パフォーマンステスト実行
```bash
go test ./internal/core/ecs/mod/ -bench=. -benchmem
```

### 4. 統合テスト実行
```bash
go test ./internal/core/ecs/mod/ -v -run TestIntegration
```

### 5. 長期テスト実行
```bash
go test ./internal/core/ecs/mod/ -v -run TestLongTerm -timeout=24h
```

## 受け入れ基準

### 機能テスト受け入れ基準
- [ ] 全単体テスト通過率: 100%
- [ ] テストカバレッジ: >95%
- [ ] 統合テスト通過率: 100%

### セキュリティテスト受け入れ基準  
- [ ] パストラバーサル攻撃防御率: 100%
- [ ] システムコマンド実行阻止率: 100%
- [ ] メモリ保護テスト通過率: 100%
- [ ] DoS攻撃軽減確認: 100%

### パフォーマンステスト受け入れ基準
- [ ] API呼び出しオーバーヘッド: <100μs
- [ ] サンドボックス実行オーバーヘッド: <5%
- [ ] 大量MOD実行時のシステム性能維持: 確認

### 長期実行テスト受け入れ基準
- [ ] 24時間安定性: メモリリーク<50MB
- [ ] エラー発生率: <0.01%
- [ ] システム可用性: >99.9%

---

**作成日時**: 2025-08-08  
**担当**: Claude Code  
**テスト項目数**: 20項目  
**カバレッジ目標**: >95%