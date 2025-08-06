# TASK-103: SystemManager実装 - テストケース仕様

## テストケース概要

SystemManagerの全機能を検証するため、単体テスト・統合テスト・パフォーマンステストを段階的に実装します。各テストケースはBDD形式で記述し、明確な検証ポイントを定義します。

## 1. システム登録・ライフサイクル管理テスト

### TC-SM-001: システム登録機能テスト

#### TC-SM-001-01: 正常なシステム登録
```go
func TestSystemManager_RegisterSystem_Success(t *testing.T) {
    // Given: 新しいSystemManagerとMockSystem
    // When: システムを登録する
    // Then: システムが正常に登録され、GetSystemで取得できる
}
```

#### TC-SM-001-02: 重複システム登録拒否
```go
func TestSystemManager_RegisterSystem_DuplicateError(t *testing.T) {
    // Given: 既に登録済みのシステムがあるSystemManager
    // When: 同じ型のシステムを再度登録しようとする
    // Then: エラーが返され、元のシステムが保持される
}
```

#### TC-SM-001-03: 優先度付きシステム登録
```go
func TestSystemManager_RegisterSystemWithPriority_Success(t *testing.T) {
    // Given: SystemManagerと異なる優先度のシステム
    // When: 優先度付きでシステムを登録する
    // Then: 優先度順でシステムが管理される
}
```

#### TC-SM-001-04: nilシステム登録エラー
```go
func TestSystemManager_RegisterSystem_NilSystemError(t *testing.T) {
    // Given: SystemManager
    // When: nilシステムを登録しようとする
    // Then: エラーが返される
}
```

#### TC-SM-001-05: システム登録解除
```go
func TestSystemManager_UnregisterSystem_Success(t *testing.T) {
    // Given: 登録済みシステムがあるSystemManager
    // When: システムを登録解除する
    // Then: システムが削除され、GetSystemでエラーが返される
}
```

### TC-SM-002: システム状態管理テスト

#### TC-SM-002-01: システム有効化・無効化
```go
func TestSystemManager_EnableDisableSystem_Success(t *testing.T) {
    // Given: 登録済みシステムがあるSystemManager
    // When: システムを無効化・有効化する
    // Then: IsSystemEnabledが正しい状態を返す
}
```

#### TC-SM-002-02: 有効・無効システム一覧取得
```go
func TestSystemManager_GetEnabledDisabledSystems_Success(t *testing.T) {
    // Given: 有効・無効システムが混在するSystemManager
    // When: 有効・無効システム一覧を取得する
    // Then: 正しいシステム一覧が返される
}
```

## 2. 依存関係管理テスト

### TC-SM-003: 依存関係設定・検証テスト

#### TC-SM-003-01: 依存関係設定
```go
func TestSystemManager_SetSystemDependency_Success(t *testing.T) {
    // Given: 2つの登録済みシステムがあるSystemManager
    // When: システムAがシステムBに依存するよう設定する
    // Then: 依存関係が設定され、GetSystemDependenciesで取得できる
}
```

#### TC-SM-003-02: 循環依存検出・拒否
```go
func TestSystemManager_SetSystemDependency_CyclicError(t *testing.T) {
    // Given: A→B依存が設定済みのSystemManager
    // When: B→A依存を設定しようとする
    // Then: 循環依存エラーが返される
}
```

#### TC-SM-003-03: 複雑な循環依存検出
```go
func TestSystemManager_SetSystemDependency_ComplexCyclicError(t *testing.T) {
    // Given: A→B→C依存が設定済みのSystemManager
    // When: C→A依存を設定しようとする
    // Then: 循環依存エラーが返される
}
```

#### TC-SM-003-04: 依存関係削除
```go
func TestSystemManager_RemoveSystemDependency_Success(t *testing.T) {
    // Given: 依存関係が設定済みのSystemManager
    // When: 依存関係を削除する
    // Then: 依存関係が削除され、実行順序が更新される
}
```

### TC-SM-004: 実行順序計算テスト

#### TC-SM-004-01: 基本実行順序計算
```go
func TestSystemManager_GetExecutionOrder_BasicOrder(t *testing.T) {
    // Given: A→B→C依存があるSystemManager
    // When: 実行順序を取得する
    // Then: [A, B, C]の順序が返される
}
```

#### TC-SM-004-02: 複雑な実行順序計算
```go
func TestSystemManager_GetExecutionOrder_ComplexOrder(t *testing.T) {
    // Given: 複雑な依存関係があるSystemManager (A→C, B→C, C→D)
    // When: 実行順序を取得する
    // Then: 依存関係を満たす正しい順序が返される
}
```

#### TC-SM-004-03: 独立システムの順序
```go
func TestSystemManager_GetExecutionOrder_IndependentSystems(t *testing.T) {
    // Given: 独立したシステムA、B、CがあるSystemManager
    // When: 実行順序を取得する
    // Then: すべてのシステムが含まれ、依存関係制約を満たす
}
```

## 3. システム実行制御テスト

### TC-SM-005: システム実行テスト

#### TC-SM-005-01: Update実行テスト
```go
func TestSystemManager_UpdateSystems_Success(t *testing.T) {
    // Given: 複数の登録済みシステムがあるSystemManager
    // When: UpdateSystemsを呼び出す
    // Then: 全システムのUpdateが実行順序通りに呼び出される
}
```

#### TC-SM-005-02: Render実行テスト
```go
func TestSystemManager_RenderSystems_Success(t *testing.T) {
    // Given: レンダリングシステムが登録されたSystemManager
    // When: RenderSystemsを呼び出す
    // Then: 全システムのRenderが実行順序通りに呼び出される
}
```

#### TC-SM-005-03: システム初期化・終了テスト
```go
func TestSystemManager_InitializeShutdownSystems_Success(t *testing.T) {
    // Given: 複数システムが登録されたSystemManager
    // When: InitializeSystems、ShutdownSystemsを呼び出す
    // Then: 全システムが正しい順序で初期化・終了される
}
```

#### TC-SM-005-04: 無効システムスキップテスト
```go
func TestSystemManager_UpdateSystems_SkipDisabled(t *testing.T) {
    // Given: 有効・無効システムが混在するSystemManager
    // When: UpdateSystemsを呼び出す
    // Then: 有効システムのみが実行される
}
```

### TC-SM-006: エラー処理・隔離テスト

#### TC-SM-006-01: システム実行エラー隔離
```go
func TestSystemManager_UpdateSystems_ErrorIsolation(t *testing.T) {
    // Given: 一つがエラーを発生するシステムが含まれるSystemManager
    // When: UpdateSystemsを呼び出す
    // Then: エラーシステムが隔離され、他システムは継続実行される
}
```

#### TC-SM-006-02: カスタムエラーハンドラー
```go
func TestSystemManager_UpdateSystems_CustomErrorHandler(t *testing.T) {
    // Given: カスタムエラーハンドラーが設定されたSystemManager
    // When: システム実行エラーが発生する
    // Then: カスタムエラーハンドラーが呼び出される
}
```

#### TC-SM-006-03: エラー履歴記録
```go
func TestSystemManager_GetSystemErrors_History(t *testing.T) {
    // Given: エラー発生履歴があるSystemManager
    // When: GetSystemErrorsを呼び出す
    // Then: 正しいエラー履歴が返される
}
```

## 4. 並列実行制御テスト

### TC-SM-007: 並列実行テスト

#### TC-SM-007-01: 独立システム並列実行
```go
func TestSystemManager_UpdateSystems_ParallelExecution(t *testing.T) {
    // Given: 並列実行が有効で、独立システムがあるSystemManager
    // When: UpdateSystemsを呼び出す
    // Then: 独立システムが並列実行される
}
```

#### TC-SM-007-02: 依存関係と並列実行の両立
```go
func TestSystemManager_UpdateSystems_ParallelWithDependencies(t *testing.T) {
    // Given: 依存関係と独立システムが混在するSystemManager
    // When: 並列実行でUpdateSystemsを呼び出す
    // Then: 依存関係を保ちつつ、可能な部分が並列実行される
}
```

#### TC-SM-007-03: 並列度制限
```go
func TestSystemManager_UpdateSystems_ParallelLimit(t *testing.T) {
    // Given: 並列度制限が設定されたSystemManager
    // When: 大量の独立システムを実行する
    // Then: 並列度制限が正しく適用される
}
```

#### TC-SM-007-04: 並列実行でのエラー処理
```go
func TestSystemManager_UpdateSystems_ParallelErrorHandling(t *testing.T) {
    // Given: 並列実行中にエラーが発生するシステム
    // When: 並列UpdateSystemsを実行する
    // Then: エラーが適切に処理され、他の並列実行に影響しない
}
```

### TC-SM-008: 並列グループ計算テスト

#### TC-SM-008-01: 基本並列グループ計算
```go
func TestSystemManager_GetParallelGroups_BasicGrouping(t *testing.T) {
    // Given: A→B、C→D依存があるSystemManager
    // When: GetParallelGroupsを呼び出す
    // Then: [[A,C], [B,D]]のグループが返される
}
```

#### TC-SM-008-02: 複雑並列グループ計算
```go
func TestSystemManager_GetParallelGroups_ComplexGrouping(t *testing.T) {
    // Given: 複雑な依存関係があるSystemManager
    // When: GetParallelGroupsを呼び出す
    // Then: 正しい並列グループが計算される
}
```

## 5. パフォーマンス監視テスト

### TC-SM-009: メトリクス収集テスト

#### TC-SM-009-01: システム実行時間測定
```go
func TestSystemManager_GetSystemMetrics_ExecutionTime(t *testing.T) {
    // Given: システム実行後のSystemManager
    // When: GetSystemMetricsを呼び出す
    // Then: 正確な実行時間が記録されている
}
```

#### TC-SM-009-02: メトリクス統計情報
```go
func TestSystemManager_GetSystemMetrics_Statistics(t *testing.T) {
    // Given: 複数回実行されたシステムがあるSystemManager
    // When: GetSystemMetricsを呼び出す
    // Then: 平均、最小、最大実行時間が正しく計算されている
}
```

#### TC-SM-009-03: メトリクスリセット
```go
func TestSystemManager_ResetSystemMetrics_Success(t *testing.T) {
    // Given: メトリクス履歴があるSystemManager
    // When: ResetSystemMetricsを呼び出す
    // Then: 全メトリクスがリセットされる
}
```

### TC-SM-010: パフォーマンス分析テスト

#### TC-SM-010-01: プロファイリング有効化・無効化
```go
func TestSystemManager_EnableProfiling_ToggleState(t *testing.T) {
    // Given: SystemManager
    // When: プロファイリングを有効化・無効化する
    // Then: IsProfilingEnabledが正しい状態を返す
}
```

#### TC-SM-010-02: プロファイリングデータ収集
```go
func TestSystemManager_Profiling_DataCollection(t *testing.T) {
    // Given: プロファイリングが有効なSystemManager
    // When: システムを実行する
    // Then: 詳細なプロファイリングデータが収集される
}
```

## 6. 設定・永続化テスト

### TC-SM-011: 設定管理テスト

#### TC-SM-011-01: システム設定保存・読み込み
```go
func TestSystemManager_SaveLoadConfiguration_Success(t *testing.T) {
    // Given: 設定されたSystemManager
    // When: 設定を保存し、新しいインスタンスで読み込む
    // Then: 設定が正しく復元される
}
```

#### TC-SM-011-02: 状態シリアライゼーション
```go
func TestSystemManager_SerializeDeserializeState_Success(t *testing.T) {
    // Given: 実行状態があるSystemManager
    // When: 状態をシリアライズ・デシリアライズする
    // Then: 状態が正しく復元される
}
```

## 7. スレッドセーフティテスト

### TC-SM-012: 並行アクセステスト

#### TC-SM-012-01: 並行システム登録
```go
func TestSystemManager_ConcurrentRegisterSystem_ThreadSafe(t *testing.T) {
    // Given: SystemManager
    // When: 複数ゴルーチンから同時にシステム登録する
    // Then: 競合状態なしで全システムが正しく登録される
}
```

#### TC-SM-012-02: 実行中の設定変更
```go
func TestSystemManager_ConcurrentConfigurationChange_ThreadSafe(t *testing.T) {
    // Given: 実行中のSystemManager
    // When: 別スレッドから設定変更する
    // Then: 競合状態なしで設定変更が適用される
}
```

## 8. パフォーマンステスト

### TC-SM-013: スケーラビリティテスト

#### TC-SM-013-01: 大量システム実行テスト
```go
func TestSystemManager_UpdateSystems_LargeScale(t *testing.T) {
    // Given: 1000システムが登録されたSystemManager
    // When: UpdateSystemsを実行する
    // Then: 実行時間が要件（<100ms）を満たす
}
```

#### TC-SM-013-02: 複雑依存関係計算テスト
```go
func TestSystemManager_ComputeExecutionOrder_ComplexDependencies(t *testing.T) {
    // Given: 5000依存関係があるSystemManager
    // When: 実行順序を計算する
    // Then: 計算時間が要件（<5ms）を満たす
}
```

#### TC-SM-013-03: 並列実行効率テスト
```go
func TestSystemManager_ParallelExecution_Efficiency(t *testing.T) {
    // Given: 並列実行可能なシステム群
    // When: 並列実行する
    // Then: 理論値の80%以上の効率を達成する
}
```

### TC-SM-014: 長期実行安定性テスト

#### TC-SM-014-01: メモリリークテスト
```go
func TestSystemManager_LongRunning_NoMemoryLeak(t *testing.T) {
    // Given: SystemManager
    // When: 24時間継続実行する
    // Then: メモリリークが発生しない
}
```

#### TC-SM-014-02: パフォーマンス劣化テスト
```go
func TestSystemManager_LongRunning_PerformanceStability(t *testing.T) {
    // Given: SystemManager
    // When: 長期間実行する
    // Then: パフォーマンスが劣化しない
}
```

## 9. エラーケース・境界値テスト

### TC-SM-015: 境界値テスト

#### TC-SM-015-01: 最大システム数テスト
```go
func TestSystemManager_MaxSystems_BoundaryTest(t *testing.T) {
    // Given: SystemManager
    // When: 1000システムを登録する
    // Then: 正常に動作し、1001システム目は拒否される
}
```

#### TC-SM-015-02: ゼロシステムテスト
```go
func TestSystemManager_ZeroSystems_EmptyExecution(t *testing.T) {
    // Given: システムが登録されていないSystemManager
    // When: UpdateSystemsを呼び出す
    // Then: エラーなく正常終了する
}
```

### TC-SM-016: 異常ケーステスト

#### TC-SM-016-01: パニック発生システム処理
```go
func TestSystemManager_SystemPanic_Recovery(t *testing.T) {
    // Given: パニックを発生させるシステム
    // When: UpdateSystemsを実行する
    // Then: パニックが回復され、他システムは継続実行される
}
```

#### TC-SM-016-02: 無限ループシステム処理
```go
func TestSystemManager_SystemInfiniteLoop_Timeout(t *testing.T) {
    // Given: 無限ループするシステム
    // When: タイムアウト設定でUpdateSystemsを実行する
    // Then: タイムアウトで実行が終了される
}
```

## テスト実装戦略

### Phase 1: 基本機能テスト実装
1. **TC-SM-001～002**: システム登録・状態管理
2. **TC-SM-012**: 基本的なスレッドセーフティ
3. **TC-SM-015**: 基本境界値テスト

### Phase 2: 依存関係管理テスト実装
1. **TC-SM-003～004**: 依存関係設定・実行順序
2. **TC-SM-016**: 異常ケース処理

### Phase 3: 実行制御テスト実装
1. **TC-SM-005～006**: システム実行・エラー処理
2. **TC-SM-007～008**: 並列実行制御

### Phase 4: 高度機能テスト実装
1. **TC-SM-009～010**: パフォーマンス監視
2. **TC-SM-011**: 設定・永続化

### Phase 5: パフォーマンス・長期テスト実装
1. **TC-SM-013**: スケーラビリティテスト
2. **TC-SM-014**: 長期実行安定性テスト

## テスト品質基準

- **カバレッジ**: >95%（行カバレッジ）
- **実行時間**: 全テスト<30秒（パフォーマンステスト除く）
- **テスト安定性**: 100回実行で100%成功
- **テスト可読性**: BDD形式での明確な仕様記述

---

**次のステップ**: このテストケース仕様に基づいて、失敗するテストを実装し、TDD Red段階を開始します。