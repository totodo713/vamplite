# TASK-101: EntityManager実装 - テストケース仕様書

## 概要

EntityManagerの全機能に対する包括的なテストケースを定義します。TDD形式で実装するため、先にテストケースを明確に定義し、その後に実装を行います。

## テストカテゴリー

### 1. 基本エンティティ操作テスト

#### 1.1 エンティティ作成テスト
```go
func TestEntityManager_CreateEntity(t *testing.T)
```
- **TC001**: 新しいエンティティを作成できること
- **TC002**: 作成されたエンティティIDが有効であること (> 0)
- **TC003**: 連続で作成したエンティティのIDが一意であること
- **TC004**: エンティティカウントが正しく増加すること

#### 1.2 指定IDでのエンティティ作成テスト
```go
func TestEntityManager_CreateEntityWithID(t *testing.T)
```
- **TC005**: 指定したIDでエンティティを作成できること
- **TC006**: 既存IDでの作成時にエラーが返されること
- **TC007**: 無効ID(0)での作成時にエラーが返されること

#### 1.3 エンティティ削除テスト
```go
func TestEntityManager_DestroyEntity(t *testing.T)
```
- **TC008**: 有効なエンティティを削除できること
- **TC009**: 削除後、エンティティが無効になること
- **TC010**: 無効なエンティティの削除時にエラーが返されること
- **TC011**: エンティティカウントが正しく減少すること

#### 1.4 エンティティ有効性確認テスト
```go
func TestEntityManager_IsValid(t *testing.T)
```
- **TC012**: 作成直後のエンティティが有効と判定されること
- **TC013**: 削除後のエンティティが無効と判定されること
- **TC014**: 存在しないエンティティが無効と判定されること
- **TC015**: EntityID(0)が無効と判定されること

### 2. エンティティリサイクルテスト

#### 2.1 エンティティリサイクル機能テスト
```go
func TestEntityManager_RecycleEntity(t *testing.T)
```
- **TC016**: エンティティをリサイクルプールに追加できること
- **TC017**: リサイクル後、新規作成時に同じIDが再利用されること
- **TC018**: リサイクル数が正しくカウントされること
- **TC019**: 無効エンティティのリサイクル時にエラーが返されること

#### 2.2 リサイクルプール管理テスト
```go
func TestEntityManager_ClearRecycled(t *testing.T)
```
- **TC020**: リサイクルプールをクリアできること
- **TC021**: クリア後、リサイクル数が0になること
- **TC022**: クリア後、新規作成時に新しいIDが生成されること

### 3. エンティティ関係管理テスト

#### 3.1 親子関係設定テスト
```go
func TestEntityManager_SetParent(t *testing.T)
```
- **TC023**: 有効な親子関係を設定できること
- **TC024**: 親を取得できること
- **TC025**: 子一覧を取得できること
- **TC026**: 循環参照設定時にエラーが返されること
- **TC027**: 無効エンティティでの関係設定時にエラーが返されること

#### 3.2 階層関係取得テスト
```go
func TestEntityManager_GetHierarchy(t *testing.T)
```
- **TC028**: 子孫エンティティを全て取得できること
- **TC029**: 祖先エンティティを全て取得できること
- **TC030**: 深い階層(10レベル)の関係を正しく処理できること
- **TC031**: 祖先関係を正しく判定できること

#### 3.3 関係解除テスト
```go
func TestEntityManager_RemoveFromParent(t *testing.T)
```
- **TC032**: 親子関係を解除できること
- **TC033**: 解除後、親が取得できないこと
- **TC034**: 解除後、子一覧から削除されること

### 4. エンティティメタデータテスト

#### 4.1 タグ機能テスト
```go
func TestEntityManager_Tags(t *testing.T)
```
- **TC035**: エンティティにタグを設定できること
- **TC036**: 設定されたタグを取得できること
- **TC037**: タグを削除できること
- **TC038**: タグでエンティティを検索できること
- **TC039**: 存在しないタグで検索時に空配列が返されること
- **TC040**: 全タグ一覧を取得できること

#### 4.2 グループ機能テスト
```go
func TestEntityManager_Groups(t *testing.T)
```
- **TC041**: グループを作成できること
- **TC042**: エンティティをグループに追加できること
- **TC043**: グループからエンティティを削除できること
- **TC044**: グループ内エンティティ一覧を取得できること
- **TC045**: エンティティが属するグループ一覧を取得できること
- **TC046**: グループを削除できること
- **TC047**: 存在しないグループ操作時にエラーが返されること

### 5. バッチ操作テスト

#### 5.1 バッチ作成テスト
```go
func TestEntityManager_CreateEntities(t *testing.T)
```
- **TC048**: 複数エンティティを一括作成できること
- **TC049**: 指定した数だけエンティティが作成されること
- **TC050**: 1000個の大量作成が1フレーム(16.67ms)以内で完了すること

#### 5.2 バッチ削除テスト
```go
func TestEntityManager_DestroyEntities(t *testing.T)
```
- **TC051**: 複数エンティティを一括削除できること
- **TC052**: 削除後、全エンティティが無効になること
- **TC053**: 無効エンティティが含まれた配列でも部分的に削除できること
- **TC054**: 1000個の大量削除が1フレーム(16.67ms)以内で完了すること

#### 5.3 バッチ検証テスト
```go
func TestEntityManager_ValidateEntities(t *testing.T)
```
- **TC055**: 有効なエンティティのみが返されること
- **TC056**: 無効エンティティが除外されること
- **TC057**: 空配列入力時に空配列が返されること

### 6. イベントシステムテスト

#### 6.1 作成イベントテスト
```go
func TestEntityManager_OnEntityCreated(t *testing.T)
```
- **TC058**: エンティティ作成時にコールバックが呼ばれること
- **TC059**: 複数コールバックが全て呼ばれること
- **TC060**: 正しいエンティティIDがコールバックに渡されること

#### 6.2 削除イベントテスト
```go
func TestEntityManager_OnEntityDestroyed(t *testing.T)
```
- **TC061**: エンティティ削除時にコールバックが呼ばれること
- **TC062**: バッチ削除時に各エンティティでコールバックが呼ばれること

#### 6.3 関係変更イベントテスト
```go
func TestEntityManager_OnParentChanged(t *testing.T)
```
- **TC063**: 親子関係設定時にコールバックが呼ばれること
- **TC064**: 関係解除時にコールバックが呼ばれること
- **TC065**: 正しい引数(子, 旧親, 新親)が渡されること

### 7. アーキタイプ管理テスト

#### 7.1 アーキタイプ取得テスト
```go
func TestEntityManager_Archetype(t *testing.T)
```
- **TC066**: エンティティのアーキタイプIDを取得できること
- **TC067**: 同じコンポーネント構成のエンティティが同じアーキタイプIDを持つこと
- **TC068**: アーキタイプ別エンティティ一覧を取得できること
- **TC069**: アーキタイプ数を取得できること

### 8. メモリ・パフォーマンス管理テスト

#### 8.1 メモリ管理テスト
```go
func TestEntityManager_Memory(t *testing.T)
```
- **TC070**: メモリコンパクション処理が実行できること
- **TC071**: 断片化率を取得できること
- **TC072**: メモリ使用量を取得できること
- **TC073**: エンティティあたりのメモリ使用量が100B以下であること

#### 8.2 プール統計テスト
```go
func TestEntityManager_PoolStats(t *testing.T)
```
- **TC074**: プール統計情報を取得できること
- **TC075**: 統計値が実際の状態と一致すること
- **TC076**: ヒット率が正しく計算されること

### 9. シリアライゼーションテスト

#### 9.1 単体シリアライゼーションテスト
```go
func TestEntityManager_Serialize(t *testing.T)
```
- **TC077**: エンティティをシリアライズできること
- **TC078**: シリアライズされたデータからエンティティを復元できること
- **TC079**: 復元されたエンティティが元と同じ状態であること
- **TC080**: 関係性も正しく復元されること

#### 9.2 バッチシリアライゼーションテスト
```go
func TestEntityManager_SerializeBatch(t *testing.T)
```
- **TC081**: 複数エンティティを一括シリアライズできること
- **TC082**: 一括復元時に全エンティティが正しく復元されること
- **TC083**: 大量データ(1000個)のシリアライゼーションが高速であること

### 10. スレッドセーフティテスト

#### 10.1 並行アクセステスト
```go
func TestEntityManager_Concurrent(t *testing.T)
```
- **TC084**: 複数ゴルーチンからの同時作成が安全であること
- **TC085**: 読み取り専用操作の並行実行が正常に動作すること
- **TC086**: 書き込み操作のロックが正しく動作すること
- **TC087**: 10並行での1000エンティティ操作がデータ競合なしで完了すること

#### 10.2 ロック機能テスト
```go
func TestEntityManager_Locking(t *testing.T)
```
- **TC088**: 排他ロック・読み取りロックが正常に動作すること
- **TC089**: デッドロックが発生しないこと
- **TC090**: ロック解放が正しく動作すること

### 11. エラーハンドリングテスト

#### 11.1 入力検証テスト
```go
func TestEntityManager_ErrorHandling(t *testing.T)
```
- **TC091**: 無効なEntityIDでの各種操作時に適切なエラーが返されること
- **TC092**: nil引数での各種操作時に適切なエラーが返されること
- **TC093**: 循環参照設定時にErrCircularReferenceが返されること
- **TC094**: 存在しないタグ・グループアクセス時に適切なエラーが返されること

#### 11.2 境界値テスト
```go
func TestEntityManager_BoundaryValues(t *testing.T)
```
- **TC095**: 最大エンティティ数(100,000)での動作が正常であること
- **TC096**: 最大階層深度(100)での動作が正常であること
- **TC097**: 空文字列タグ設定時にエラーが返されること
- **TC098**: 長すぎるタグ名(>256文字)設定時にエラーが返されること

### 12. 統合テスト

#### 12.1 実ゲームシナリオテスト
```go
func TestEntityManager_GameScenario(t *testing.T)
```
- **TC099**: プレイヤー・敵・アイテムエンティティの作成・管理ができること
- **TC100**: 複雑な階層構造(武器→アタッチメント→パーツ)が正しく管理されること
- **TC101**: ゲームオブジェクトのタグ・グループによる分類が正常動作すること

#### 12.2 長期実行テスト
```go
func TestEntityManager_LongRunning(t *testing.T)
```
- **TC102**: 1時間の連続実行でメモリリークがないこと
- **TC103**: 10万回の作成・削除サイクルで性能劣化がないこと
- **TC104**: 長期実行後もデータ整合性が保たれること

### 13. パフォーマンステスト

#### 13.1 作成・削除性能テスト
```go
func BenchmarkEntityManager_Creation(b *testing.B)
func BenchmarkEntityManager_Destruction(b *testing.B)
```
- **PC001**: 1000エンティティ作成時間 < 16.67ms
- **PC002**: 1000エンティティ削除時間 < 16.67ms
- **PC003**: メモリ使用量 < 100B/エンティティ

#### 13.2 検索性能テスト
```go
func BenchmarkEntityManager_Search(b *testing.B)
```
- **PC004**: タグ検索時間 < 1ms (10,000エンティティ中)
- **PC005**: グループ検索時間 < 1ms
- **PC006**: 階層検索時間 < 1ms

#### 13.3 スケーラビリティテスト
```go
func BenchmarkEntityManager_Scale(b *testing.B)
```
- **PC007**: 100,000エンティティでの基本操作が正常動作すること
- **PC008**: 10,000タグでの検索が<1msで完了すること
- **PC009**: 1,000グループでの管理が正常動作すること

## テスト実装戦略

### Phase 1: 基本機能テスト (TC001-TC030)
- エンティティの基本ライフサイクル
- 関係管理の基本機能
- エラーハンドリングの基本

### Phase 2: 高度機能テスト (TC031-TC070)
- メタデータ管理
- バッチ操作
- イベントシステム
- アーキタイプ管理

### Phase 3: システム統合テスト (TC071-TC090)
- メモリ管理
- シリアライゼーション
- スレッドセーフティ

### Phase 4: 品質保証テスト (TC091-PC009)
- エラーハンドリング
- 境界値テスト
- パフォーマンステスト
- 長期実行テスト

## テストデータ設計

### テスト用エンティティ構造
```go
// テスト用の基本エンティティパターン
type TestEntityPattern struct {
    Name        string
    EntityCount int
    Hierarchy   [][]EntityID  // [parent][children]
    Tags        map[EntityID]string
    Groups      map[string][]EntityID
}

// 標準テストパターン
var StandardTestPatterns = []TestEntityPattern{
    {
        Name: "Simple",
        EntityCount: 10,
        // 単純な構造
    },
    {
        Name: "Complex",
        EntityCount: 1000,
        // 複雑な階層・関係
    },
    {
        Name: "Large",
        EntityCount: 10000,
        // 大規模データ
    },
}
```

## モック・テストダブル

### インターフェースモック
```go
type MockEntityEventBus struct {
    Events []EntityEvent
}

type MockArchetypeManager struct {
    Archetypes map[EntityID]ArchetypeID
}
```

## テスト実行環境

### 実行条件
- **Go version**: 1.22+
- **並行数**: CPU core数
- **メモリ制限**: 1GB
- **実行時間制限**: 各テスト30秒

### CI/CD環境テスト
- **Linux**: Ubuntu 22.04
- **Windows**: Windows Server 2022
- **macOS**: macOS 13

---

これらのテストケースに基づいて、次ステップでTDD Red段階（失敗するテスト実装）を開始します。