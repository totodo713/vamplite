# TASK-101: EntityManager実装 - 要件定義書

## 概要

EntityManagerは、ECSフレームワークにおけるエンティティのライフサイクル管理を担当する中核コンポーネントです。エンティティの作成・削除・リサイクル・関係管理・タグ管理・グループ管理等の機能を提供し、高性能でスレッドセーフな操作を実現します。

## 機能要件

### 1. エンティティライフサイクル管理 (REQ-001)

#### 1.1 基本操作
- **CreateEntity()**: 新しいエンティティIDを生成・作成
- **CreateEntityWithID(EntityID)**: 指定IDでエンティティを作成
- **DestroyEntity(EntityID)**: エンティティを削除
- **IsValid(EntityID)**: エンティティの有効性を確認
- **GetActiveEntities()**: アクティブなエンティティ一覧を取得
- **GetEntityCount()**: 現在のエンティティ数を取得
- **GetMaxEntityCount()**: 最大エンティティ数を取得

#### 1.2 エンティティリサイクル
- **RecycleEntity(EntityID)**: エンティティをリサイクルプールに追加
- **GetRecycledCount()**: リサイクル済みエンティティ数を取得
- **ClearRecycled()**: リサイクルプールをクリア

### 2. エンティティ関係管理 (REQ-101)

#### 2.1 親子関係
- **SetParent(child, parent)**: 親子関係を設定
- **GetParent(EntityID)**: 親エンティティを取得
- **GetChildren(EntityID)**: 子エンティティ一覧を取得
- **GetDescendants(EntityID)**: 全子孫エンティティを取得
- **GetAncestors(EntityID)**: 全祖先エンティティを取得
- **RemoveFromParent(EntityID)**: 親子関係を解除
- **IsAncestor(ancestor, descendant)**: 祖先関係を確認

#### 2.2 循環参照防止
- 親子関係設定時の循環参照検出・防止
- 無限ループの防止機能

### 3. エンティティメタデータ管理 (REQ-401)

#### 3.1 タグ機能
- **SetTag(EntityID, string)**: エンティティに文字列タグを設定
- **GetTag(EntityID)**: エンティティのタグを取得
- **RemoveTag(EntityID)**: エンティティのタグを削除
- **FindByTag(string)**: タグでエンティティを検索
- **GetAllTags()**: 全タグ一覧を取得

#### 3.2 グループ機能
- **CreateGroup(string)**: エンティティグループを作成
- **AddToGroup(EntityID, string)**: エンティティをグループに追加
- **RemoveFromGroup(EntityID, string)**: エンティティをグループから削除
- **GetGroup(string)**: グループ内のエンティティ一覧を取得
- **GetEntityGroups(EntityID)**: エンティティが属するグループ一覧を取得
- **DestroyGroup(string)**: グループを削除

### 4. バッチ操作 (パフォーマンス要件)

#### 4.1 大量処理
- **CreateEntities(count)**: 複数エンティティを一括作成
- **DestroyEntities([]EntityID)**: 複数エンティティを一括削除
- **ValidateEntities([]EntityID)**: 複数エンティティの有効性を一括確認

### 5. イベントシステム (REQ-201)

#### 5.1 ライフサイクルイベント
- **OnEntityCreated(callback)**: エンティティ作成時のコールバック登録
- **OnEntityDestroyed(callback)**: エンティティ削除時のコールバック登録
- **OnParentChanged(callback)**: 親子関係変更時のコールバック登録

### 6. アーキタイプ管理 (パフォーマンス最適化)

#### 6.1 アーキタイプ操作
- **GetArchetype(EntityID)**: エンティティのアーキタイプIDを取得
- **GetEntitiesByArchetype(ArchetypeID)**: アーキタイプ別エンティティ一覧を取得
- **GetArchetypeCount()**: アーキタイプ数を取得

### 7. メモリ・パフォーマンス管理

#### 7.1 メモリ最適化
- **Compact()**: メモリの断片化を解消
- **GetFragmentation()**: メモリ断片化率を取得
- **GetMemoryUsage()**: メモリ使用量を取得
- **GetPoolStats()**: エンティティプール統計を取得

### 8. シリアライゼーション

#### 8.1 保存・復元
- **SerializeEntity(EntityID)**: エンティティをシリアライズ
- **DeserializeEntity(*EntityData)**: エンティティをデシリアライズ
- **SerializeBatch([]EntityID)**: 複数エンティティを一括シリアライズ
- **DeserializeBatch([]*EntityData)**: 複数エンティティを一括デシリアライズ

### 9. スレッドセーフティ

#### 9.1 ロック機能
- **Lock()**: 排他ロック取得
- **RLock()**: 読み取り専用ロック取得
- **Unlock()**: 排他ロック解放
- **RUnlock()**: 読み取り専用ロック解放

### 10. デバッグ・検証

#### 10.1 デバッグ機能
- **ValidateIntegrity()**: データ整合性を検証
- **GetDebugInfo()**: デバッグ情報を取得

## 非機能要件

### パフォーマンス要件 (NFR-001, NFR-002)

1. **作成性能**: エンティティ作成 - 1000個/フレーム
2. **削除性能**: エンティティ削除 - 1000個/フレーム  
3. **検索性能**: タグ・グループ検索 - <1ms
4. **メモリ効率**: <100B/エンティティ
5. **同時実行**: スレッドセーフな操作保証

### 拡張性要件 (NFR-003)

1. **最大エンティティ数**: 100,000個まで対応
2. **最大階層深度**: 100レベルまで対応
3. **最大タグ数**: 10,000個まで対応
4. **最大グループ数**: 1,000個まで対応

### 安定性要件 (NFR-004)

1. **メモリリーク防止**: 24時間連続実行で増加<50MB
2. **エラーハンドリング**: 全API関数でエラー処理
3. **データ整合性**: 循環参照・不正状態の検出・防止

## データ構造仕様

### EntityID
```go
type EntityID uint32
```
- 32ビット符号なし整数
- 0は無効値として予約
- 1〜4294967295の範囲で使用

### EntityPoolStats
```go
type EntityPoolStats struct {
    TotalEntities    int     // 総エンティティ数
    ActiveEntities   int     // アクティブエンティティ数
    RecycledEntities int     // リサイクル済みエンティティ数
    PoolCapacity     int     // プール容量
    MemoryUsed       int64   // 使用メモリ(バイト)
    MemoryReserved   int64   // 予約メモリ(バイト)
    Fragmentation    float64 // 断片化率(0.0-1.0)
    HitRate          float64 // プールヒット率(0.0-1.0)
}
```

### EntityManagerDebugInfo
```go
type EntityManagerDebugInfo struct {
    EntityCount     int                 // エンティティ数
    MaxEntityID     EntityID            // 最大エンティティID
    RecycledCount   int                 // リサイクル済み数
    ArchetypeCount  int                 // アーキタイプ数
    TagCount        int                 // タグ数
    GroupCount      int                 // グループ数
    HierarchyDepth  int                 // 最大階層深度
    MemoryUsage     int64               // メモリ使用量
    PoolStats       *EntityPoolStats    // プール統計
    ArchetypeStats  map[ArchetypeID]int // アーキタイプ別統計
    TagDistribution map[string]int      // タグ別分布
    GroupSizes      map[string]int      // グループ別サイズ
}
```

## エラーハンドリング仕様

### エラータイプ
1. **ErrInvalidEntity**: 無効なエンティティIDアクセス
2. **ErrEntityNotFound**: 存在しないエンティティ操作
3. **ErrCircularReference**: 循環参照の検出
4. **ErrEntityAlreadyExists**: 既存エンティティIDでの作成試行
5. **ErrTagNotFound**: 存在しないタグアクセス
6. **ErrGroupNotFound**: 存在しないグループアクセス
7. **ErrMemoryLimitExceeded**: メモリ制限超過
8. **ErrConcurrentModification**: 同時変更の検出

### エラー処理方針
- 全API関数は適切なエラー型を返す
- パニックを起こさない防御的実装
- 部分的な失敗でもシステム全体を停止させない
- エラーログの詳細記録

## テスト戦略

### 単体テスト範囲
1. **基本操作**: 作成・削除・有効性確認
2. **関係管理**: 親子関係・循環参照防止
3. **メタデータ**: タグ・グループ機能
4. **パフォーマンス**: 大量データ処理
5. **エラーハンドリング**: 例外状況の処理
6. **スレッドセーフティ**: 並行アクセステスト

### 統合テスト範囲  
1. **システム連携**: ComponentStore・SystemManagerとの連携
2. **シリアライゼーション**: 保存・復元の整合性
3. **長期実行**: メモリリーク・安定性確認

### パフォーマンステスト
1. **作成性能**: 1000エンティティ/フレーム
2. **削除性能**: 1000エンティティ/フレーム
3. **検索性能**: タグ・グループ検索<1ms
4. **メモリ効率**: <100B/エンティティ
5. **スケーラビリティ**: 100,000エンティティ処理

## 受入基準

### 機能受入基準
- [ ] 全インターフェース機能が正常動作
- [ ] エラーハンドリングが適切に実装
- [ ] 循環参照防止機能が動作
- [ ] スレッドセーフな操作が保証

### パフォーマンス受入基準
- [ ] エンティティ作成: 1000個/フレーム達成
- [ ] エンティティ削除: 1000個/フレーム達成
- [ ] タグ検索: <1ms達成
- [ ] メモリ効率: <100B/エンティティ達成

### 品質受入基準
- [ ] 単体テストカバレッジ: >95%
- [ ] 統合テスト: 全シナリオ通過
- [ ] パフォーマンステスト: 全目標達成
- [ ] 長期実行テスト: 24時間安定動作

## 実装上の注意事項

### 設計原則
1. **単一責任**: EntityManagerは管理機能のみに特化
2. **インターフェース分離**: 機能別インターフェース設計
3. **依存性注入**: 外部依存性の最小化
4. **テスト容易性**: モック可能な設計

### パフォーマンス考慮事項
1. **メモリプール**: エンティティIDの効率的な再利用
2. **データ局所性**: キャッシュ効率を考慮したデータ配置
3. **ロック粒度**: 必要最小限のロック範囲
4. **遅延削除**: 削除処理の最適化

### セキュリティ考慮事項
1. **境界値チェック**: 配列・マップアクセスの安全性
2. **リソース制限**: メモリ・CPU使用量の制限
3. **入力検証**: API引数の妥当性確認

---

この要件定義書に基づいて、次ステップでテストケースを作成し、TDD形式でEntityManagerを実装していきます。