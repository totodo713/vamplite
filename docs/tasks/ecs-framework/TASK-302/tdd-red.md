# TASK-302: ModSecurityValidator TDD Red段階

## 実装状況

TDD Red段階として、失敗するテストを作成しました。

## 作成ファイル

1. **security_validator.go**
   - ModSecurityValidatorインターフェース定義
   - 各種型定義（SecurityLevel, ViolationType, Resource, Action等）
   - データ構造の定義（PermissionPolicy, SecurityAnalysisResult等）

2. **security_validator_test.go**
   - 14個のテストケース実装
   - 静的解析、権限管理、実行時検証、監査機能のテスト

## テストケース一覧

### 静的解析テスト
- ✅ TestSecurityValidator_AnalyzeCode_DangerousCommands
- ✅ TestSecurityValidator_AnalyzeCode_PathTraversal
- ✅ TestSecurityValidator_AnalyzeCode_UnauthorizedNetwork
- ✅ TestSecurityValidator_AnalyzeCode_SafeCode
- ✅ TestSecurityValidator_ValidateImports
- ✅ TestSecurityValidator_SQLInjection

### 権限管理テスト
- ✅ TestSecurityValidator_CheckPermission
- ✅ TestSecurityValidator_RequestPermissionElevation

### 実行時検証テスト
- ✅ TestSecurityValidator_MonitorResourceUsage
- ✅ TestSecurityValidator_DetectAnomalies
- ✅ TestSecurityValidator_ValidateRuntimeOperation

### 監査テスト
- ✅ TestSecurityValidator_LogSecurityEvent
- ✅ TestSecurityValidator_GenerateSecurityReport
- ✅ TestSecurityValidator_GetAuditTrail

### パフォーマンステスト
- ✅ TestSecurityValidator_Performance_AnalyzeSpeed

## テスト実行結果

```bash
$ go test ./internal/core/ecs/mod -run TestSecurityValidator -v

=== RUN   TestSecurityValidator_AnalyzeCode_DangerousCommands
--- FAIL: TestSecurityValidator_AnalyzeCode_DangerousCommands (0.00s)
panic: runtime error: invalid memory address or nil pointer dereference
```

すべてのテストが失敗しています（期待通り）。

## 既存コードとの競合解決

- SecurityEvent型が既存のsecurity.goと競合
  - → ValidatorSecurityEventに名前変更して解決

## 次のステップ

TDD Green段階に進み、テストが通る最小実装を行います。

### 実装優先順位

1. **Phase 1: 基本的な静的解析**
   - 正規表現ベースの危険パターン検出
   - 基本的なインポート検証

2. **Phase 2: 権限管理**
   - ポリシー設定と権限チェック
   - 権限昇格メカニズム

3. **Phase 3: 実行時検証**
   - リソース監視
   - 異常検知

4. **Phase 4: 監査機能**
   - イベントログ
   - レポート生成

## Red段階の完了確認

- [x] インターフェース定義完了
- [x] 型定義完了
- [x] テストケース作成完了
- [x] すべてのテストが失敗（nil pointer dereference）
- [x] TDD Red段階の要件を満たす