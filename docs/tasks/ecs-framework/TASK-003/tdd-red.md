# TASK-003: 依存関係の更新と互換性確認 - Red Phase

## 現状分析と失敗テスト実装

### 現在の依存関係分析

まず、現在の依存関係の状況を把握し、どこに問題があるかを特定します。

#### 1. 現在の依存関係一覧

```bash
# 直接依存関係
go list -m -f '{{.Path}} {{.Version}}' all | grep -v muscle-dreamer
```

#### 2. Go 1.24 互換性チェック

```bash
# 基本互換性テスト（これが失敗することを確認）
go mod tidy
go mod download
go list -deps ./...
```

### テスト実装

#### TC-001: Go 1.24 互換性基本テスト（現状失敗を確認）

現在の状態では以下の問題が予想される：
1. 一部の依存関係がGo 1.24で非推奨のAPIを使用している可能性
2. 間接依存関係でGo 1.24との非互換性がある可能性

#### TC-101: 脆弱性スキャンテスト（失敗を確認）

```bash
# govulncheckがインストールされていない状態を確認
govulncheck ./...
# → command not found エラーが期待される
```

#### TC-201: 基本機能動作テスト（現状の問題確認）

既存のテストを実行して、現在の状態を確認：

```bash
# 現在のテスト実行状況
go test ./internal/core/...
go test ./internal/mod/...
```

### 実行記録

実際にテストを実行し、現在の失敗状況を記録しました。

#### 実行結果

1. **現在の依存関係の状況**:
   - ebiten/v2 v2.6.3
   - testify v1.10.0
   - 間接依存関係多数（golang.org/x/ パッケージ群）

2. **脆弱性スキャン**:
   ✅ `govulncheck ./...` → No vulnerabilities found

3. **テスト実行結果**:
   - ✅ `go test ./internal/core/ecs` → 成功 (0.022s)
   - ✅ `go test ./internal/core/ecs/components` → 成功 (0.006s)
   - ✅ `go test ./internal/core/ecs/storage` → 成功 (0.033s)
   - ❌ `go test -race ./...` → **FAILED**

4. **失敗の詳細**:
   ```
   # muscle-dreamer/docs/reverse/interface-tests
   docs/reverse/interface-tests/asset_interface_test.go:15:2: 
   package muscle-dreamer/docs/reverse/interfaces is not in std
   ```

5. **根本原因**:
   - `docs/reverse/interface-tests/asset_interface_test.go` で存在しないパッケージ `muscle-dreamer/docs/reverse/interfaces` をインポート
   - このパッケージが実際には存在しない
   - レース条件テスト（`go test -race ./...`）で全体テストが失敗

#### Red State確認完了 ✅

現在の状態では以下の問題が存在することを確認：
1. 不正なインポートによるテスト失敗
2. 存在しないパッケージへの依存
3. 全体テストの実行不能状態

これらの問題を次のGreen Phaseで解決します。