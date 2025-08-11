# TASK-303: Lua Bridge実装 - 詳細要件定義

## 要件概要

**要件ID**: REQ-303  
**要件名**: Luaスクリプティング API統合  
**要件分類**: MOD統合・スクリプティング  
**優先度**: 高  
**依存要件**: REQ-301 (ModECSAPI), REQ-302 (ModSecurityValidator)

## 背景・目的

### 背景
- Muscle DreamerのMODシステムでは、安全でパフォーマンスの良いスクリプティング環境が必要
- Luaは軽量で安全なサンドボックス実行が可能で、ゲーム業界で広く使用されている
- Go言語との相互運用性を確保し、ECSシステムへの安全なアクセスを提供する必要がある

### 目的
1. **Go ↔ Lua データ相互変換**: 型安全で効率的なデータ変換機能を提供
2. **Lua向けECS API**: 制限されたECSインターフェースをLuaから操作可能にする
3. **安全なスクリプト実行**: サンドボックス環境での安全なLuaコード実行
4. **動的ロード・アンロード**: MODの動的な読み込み・削除機能
5. **エラーハンドリング**: Luaエラーの適切な捕捉・処理・デバッグ支援

## 機能要件

### FR-303-01: Go ↔ Lua データ変換機能

**詳細要件**:
- Go の基本型（string, int, float, bool）とLua型の相互変換
- Goスライス・マップとLuaテーブルの相互変換  
- Goの構造体とLuaテーブルの相互変換（reflection利用）
- エラー安全な変換（型不一致時の適切なエラー処理）

**受け入れ基準**:
- [ ] Go基本型→Lua型変換が正常動作する
- [ ] Lua型→Go基本型変換が正常動作する
- [ ] Goスライス→Luaテーブル変換が正常動作する
- [ ] Luaテーブル→Goスライス変換が正常動作する
- [ ] Go構造体→Luaテーブル変換が正常動作する
- [ ] Luaテーブル→Go構造体変換が正常動作する
- [ ] 型不一致時に適切なエラーが返される
- [ ] 変換オーバーヘッドが最小化されている（<1ms）

### FR-303-02: Lua向けECS APIラッパー

**詳細要件**:
- EntityManager操作のLuaラッパー（CreateEntity, DestroyEntity, GetEntity）
- ComponentStore操作のLuaラッパー（AddComponent, RemoveComponent, GetComponent）  
- SystemManager操作のLuaラッパー（RegisterSystem, ExecuteSystemなど）
- QueryEngine操作のLuaラッパー（QueryEntities, FilterComponentsなど）
- API権限制御（MODごとの許可・禁止機能管理）

**受け入れ基準**:
- [ ] LuaからEntityManagerの全基本操作が実行可能
- [ ] LuaからComponentStoreの全基本操作が実行可能  
- [ ] LuaからSystemManagerの制限された操作が実行可能
- [ ] LuaからQueryEngineの基本操作が実行可能
- [ ] API権限制御が正常に動作する
- [ ] 不正なAPI呼び出し時に適切なエラーが返される
- [ ] APIドキュメントが自動生成される

### FR-303-03: Luaスクリプト実行環境

**詳細要件**:
- 独立したLua VMの作成・管理・削除
- スクリプトのサンドボックス実行（ファイルアクセス・ネットワークアクセス制限）
- スクリプト実行時間・メモリ使用量制限
- Luaグローバル環境の制御（危険な標準ライブラリの削除）
- 複数スクリプトの並列実行サポート

**受け入れ基準**:
- [ ] Lua VMの作成・削除が正常動作する
- [ ] サンドボックス環境が適切に制限される
- [ ] ファイルアクセスが制限される
- [ ] ネットワークアクセスが制限される  
- [ ] スクリプト実行時間制限が動作する（デフォルト: 100ms）
- [ ] スクリプトメモリ使用量制限が動作する（デフォルト: 10MB）
- [ ] 危険なライブラリ（os, io, debug等）がアクセス不可能
- [ ] 複数スクリプトの並列実行が安全に動作する

### FR-303-04: 動的ロード・アンロード機能

**詳細要件**:
- Luaスクリプトファイルの動的読み込み
- スクリプトのホットリロード（実行中の変更反映）
- スクリプトのアンロード・リソース解放
- スクリプト依存関係管理
- ロード・アンロード時のエラー処理

**受け入れ基準**:
- [ ] Luaスクリプトファイルの動的読み込みが動作する
- [ ] スクリプトのホットリロードが動作する
- [ ] スクリプトのアンロードが完全にリソースを解放する
- [ ] スクリプト間の依存関係が適切に管理される
- [ ] ロード・アンロード時のエラーが適切に処理される
- [ ] メモリリークが発生しない（24時間テスト）

### FR-303-05: エラーハンドリング・デバッグ支援

**詳細要件**:
- Luaランタイムエラーの捕捉・Go例外変換
- スタックトレース情報の提供
- デバッグ情報の出力（行番号、関数名等）
- エラーログの構造化出力
- デバッガー統合のためのフック機能

**受け入れ基準**:
- [ ] Luaランタイムエラーが適切にGoエラーに変換される
- [ ] スタックトレース情報が正確に取得できる
- [ ] エラー発生行番号・関数名が特定できる
- [ ] エラーログが構造化されている（JSON形式）
- [ ] デバッガー統合のためのフック機能が動作する
- [ ] パニックによるプロセス終了が発生しない

## 非機能要件

### NFR-303-01: パフォーマンス要件

**要求仕様**:
- Go ↔ Lua データ変換時間: <1ms（基本型・小構造体）
- Luaスクリプト実行オーバーヘッド: <10%（Goネイティブ比）
- 複数スクリプト同時実行: 最大100並列
- メモリ使用量: スクリプト1つあたり<10MB

**測定方法**:
- ベンチマークテストによる定量的測定
- プロファイリングツールによる詳細解析
- 負荷テストによる並列実行性能確認

### NFR-303-02: セキュリティ要件

**要求仕様**:
- サンドボックス脱出攻撃100%防御
- ファイルシステムアクセス制限100%実施
- ネットワークアクセス制限100%実施
- メモリ使用量攻撃（メモリボム）防御

**検証方法**:
- ペネトレーションテスト実施
- 既知攻撃パターンでの検証
- セキュリティスキャンツール実行

### NFR-303-03: 可用性・安定性要件

**要求仕様**:
- Luaスクリプトエラー時のシステム継続動作
- 24時間連続実行でのメモリリーク<1MB
- 異常スクリプト時のリソース適切解放
- フォルトトレラント設計（1つのスクリプトエラーが他に影響しない）

**検証方法**:
- 長期実行安定性テスト
- メモリプロファイリング継続監視
- 故障注入テスト実行

## インターフェース設計

### Lua Bridge Core インターフェース

```go
// LuaBridge - メインのLua統合インターフェース
type LuaBridge interface {
    // Lua VM管理
    CreateVM(config *LuaVMConfig) (*LuaVM, error)
    DestroyVM(vm *LuaVM) error
    
    // スクリプト実行
    LoadScript(vm *LuaVM, scriptPath string) (*LuaScript, error)
    UnloadScript(vm *LuaVM, script *LuaScript) error
    ExecuteScript(vm *LuaVM, script *LuaScript) error
    
    // データ変換
    GoToLua(vm *LuaVM, value interface{}) (lua.LValue, error)
    LuaToGo(vm *LuaVM, value lua.LValue, target interface{}) error
    
    // API登録
    RegisterECSAPI(vm *LuaVM, ecsAPI *ModECSAPI) error
    SetPermissions(vm *LuaVM, permissions *APIPermissions) error
}

// LuaVM - Lua仮想マシンラッパー
type LuaVM struct {
    state        *lua.LState
    sandbox      *Sandbox
    permissions  *APIPermissions
    resources    *ResourceLimits
    errorHandler ErrorHandler
}

// LuaScript - Luaスクリプト管理
type LuaScript struct {
    path       string
    content    []byte
    loaded     bool
    metadata   *ScriptMetadata
}

// APIPermissions - API権限管理
type APIPermissions struct {
    AllowedAPIs    []string
    ForbiddenAPIs  []string
    ResourceLimits *ResourceLimits
}

// ResourceLimits - リソース制限
type ResourceLimits struct {
    MaxExecutionTime time.Duration // デフォルト: 100ms
    MaxMemoryUsage   int64         // デフォルト: 10MB
    MaxFileAccess    bool          // デフォルト: false
    MaxNetworkAccess bool          // デフォルト: false
}
```

### Lua側 ECS API

```lua
-- Entity管理API
local entity = ecs.create_entity()
ecs.destroy_entity(entity)
local exists = ecs.entity_exists(entity)

-- Component管理API  
ecs.add_component(entity, "Transform", {x=10, y=20, z=0})
local transform = ecs.get_component(entity, "Transform")
ecs.remove_component(entity, "Transform")
local has = ecs.has_component(entity, "Transform")

-- Query API
local entities = ecs.query()
    :with("Transform")
    :with("Sprite")
    :without("Health")
    :execute()

for _, entity in pairs(entities) do
    local transform = ecs.get_component(entity, "Transform")
    print("Entity " .. entity .. " at " .. transform.x .. "," .. transform.y)
end

-- Event API
ecs.fire_event("PlayerDeath", {player_id = 123, cause = "zombie"})
ecs.subscribe("PlayerDeath", function(event_data)
    print("Player " .. event_data.player_id .. " died: " .. event_data.cause)
end)
```

## データモデル

### Script Metadata
```go
type ScriptMetadata struct {
    Name         string            `json:"name"`
    Version      string            `json:"version"`
    Author       string            `json:"author"`
    Description  string            `json:"description"`
    Dependencies []string          `json:"dependencies"`
    Permissions  []string          `json:"permissions"`
    API_Version  string            `json:"api_version"`
    Entry_Point  string            `json:"entry_point"`
}
```

### Runtime Statistics
```go
type RuntimeStats struct {
    ScriptPath       string        `json:"script_path"`
    ExecutionCount   int64         `json:"execution_count"`
    TotalExecTime    time.Duration `json:"total_execution_time"`
    AverageExecTime  time.Duration `json:"average_execution_time"`
    MemoryUsage      int64         `json:"memory_usage"`
    ErrorCount       int64         `json:"error_count"`
    LastExecuted     time.Time     `json:"last_executed"`
}
```

## エラー処理戦略

### エラー分類
1. **Lua構文エラー**: スクリプト読み込み時の構文チェック
2. **Lua実行時エラー**: スクリプト実行中のランタイムエラー  
3. **API権限エラー**: 許可されていないAPI呼び出し
4. **リソース制限エラー**: メモリ・実行時間制限超過
5. **データ変換エラー**: Go ↔ Lua型変換失敗

### エラーハンドリング方針
- **フォルトイソレーション**: 1つのスクリプトエラーが他に影響しない
- **グレースフルデグラデーション**: エラー時も基本機能は継続
- **詳細なエラー情報**: デバッグ可能な詳細情報提供
- **自動回復**: 可能な場合は自動的に回復処理実行

## セキュリティ考慮事項

### サンドボックス実装
- **ファイルアクセス制限**: 指定ディレクトリ外へのアクセス禁止
- **ネットワークアクセス禁止**: socket、http等のライブラリ無効化
- **システムコマンド禁止**: os.execute、io.popen等の無効化
- **リソース制限**: CPU時間・メモリ使用量の上限設定

### 攻撃パターン対策
```lua
-- 以下のような攻撃パターンを検出・防御
-- 1. メモリボム攻撃
local huge_table = {}
for i=1,10000000 do huge_table[i] = string.rep("x", 1000) end

-- 2. 無限ループ攻撃  
while true do end

-- 3. ファイルアクセス攻撃
local file = io.open("/etc/passwd", "r")

-- 4. システムコマンド攻撃
os.execute("rm -rf /")
```

## テスト戦略

### 単体テスト項目
1. **データ変換テスト**: Go ↔ Lua 全型パターン
2. **API ラッパーテスト**: 各ECS API呼び出し
3. **サンドボックステスト**: 制限機能の動作確認
4. **エラーハンドリングテスト**: 各種エラーパターン
5. **リソース制限テスト**: メモリ・時間制限の動作確認

### 統合テスト項目
1. **MODスクリプト実行テスト**: 実際のMODシナリオ実行
2. **並列実行テスト**: 複数スクリプト同時実行
3. **ホットリロードテスト**: 実行中のスクリプト変更
4. **長期実行テスト**: 24時間メモリリーク確認
5. **セキュリティテスト**: 攻撃パターン防御確認

### パフォーマンステスト項目
1. **データ変換ベンチマーク**: 変換速度測定
2. **スクリプト実行ベンチマーク**: 実行オーバーヘッド測定  
3. **並列実行ベンチマーク**: スケーラビリティ測定
4. **メモリ使用量プロファイル**: メモリ効率測定

## 実装フェーズ

### Phase 1: Core Infrastructure (1日目)
- LuaBridge基本インターフェース実装
- LuaVM管理機能（作成・削除）
- 基本的なGo ↔ Lua データ変換

### Phase 2: ECS API Integration (2日目) 
- Lua向けECS APIラッパー実装
- API権限管理システム
- 基本的なComponent・Entity操作

### Phase 3: Security & Sandbox (3日目)
- サンドボックス環境実装
- リソース制限機能
- セキュリティ検証・テスト

### Phase 4: Advanced Features (4日目)
- 動的ロード・アンロード機能
- エラーハンドリング・デバッグ支援
- パフォーマンス最適化・統合テスト

## 成功基準

### 機能完成度
- [ ] 全機能要件100%実装完了
- [ ] 全受け入れ基準100%達成
- [ ] APIドキュメント完成

### 品質基準
- [ ] 単体テストカバレッジ95%以上
- [ ] 統合テスト全シナリオ通過
- [ ] セキュリティテスト100%通過

### パフォーマンス基準
- [ ] データ変換時間<1ms達成
- [ ] スクリプト実行オーバーヘッド<10%達成
- [ ] 100並列実行サポート確認

### セキュリティ基準
- [ ] 既知攻撃パターン100%防御確認
- [ ] ペネトレーションテスト通過
- [ ] セキュリティ監査完了

## 依存関係・制約事項

### 技術的依存関係
- **Go Lua Library**: github.com/yuin/gopher-lua使用
- **TASK-301**: ModECSAPI実装完了が前提
- **TASK-302**: ModSecurityValidator実装完了が前提

### 制約事項
- Lua 5.1互換性維持（gopher-lua制約）
- Goのreflection使用によるパフォーマンス影響
- サンドボックス機能の制限（完全分離の困難性）

### リスク要因
- **高**: Luaサンドボックス実装の複雑さ
- **中**: Go ↔ Lua データ変換のパフォーマンス
- **中**: 複数Lua VM並列実行時の安定性

## 参考資料

### 技術文書
- [gopher-lua Documentation](https://github.com/yuin/gopher-lua)
- [Lua 5.1 Reference Manual](https://www.lua.org/manual/5.1/)
- [Go Reflection Package](https://pkg.go.dev/reflect)

### 設計文書
- `docs/design/ecs-framework/interfaces.go` - ECSインターフェース定義
- `docs/design/mod-system/security.md` - MODセキュリティ設計  
- `docs/design/mod-system/scripting-api.md` - スクリプティングAPI設計

---

**要件定義完了**: TASK-303の詳細要件定義が完了しました。次にテストケース作成フェーズに進みます。