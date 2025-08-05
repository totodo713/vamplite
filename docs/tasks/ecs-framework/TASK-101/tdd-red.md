# TASK-101: EntityManager実装 - TDD Red段階

## 概要

TDD Red段階では、要件仕様に基づいて失敗するテストを先に実装します。この段階では、EntityManagerの実装は存在しないため、全てのテストが失敗することを確認します。

## Red段階の実装戦略

### Phase 1: 基本機能テスト実装
最も重要な基本機能のテストから実装を開始し、段階的に高度な機能のテストを追加していきます。

1. **基本エンティティ操作**: CreateEntity, DestroyEntity, IsValid
2. **エンティティリサイクル**: RecycleEntity, GetRecycledCount
3. **親子関係管理**: SetParent, GetParent, GetChildren
4. **メタデータ管理**: Tags, Groups
5. **バッチ操作**: CreateEntities, DestroyEntities

## 実装結果

### 構造体とインターフェース実装
- `DefaultEntityManager` 構造体を作成
- 全メソッドをスタブ実装（`panic("not implemented")`）
- 必要なフィールドを定義（次のGreen段階で使用予定）

### テスト実装
- 104個のテストケースを実装
- 基本機能からパフォーマンステストまで包括的にカバー
- エラーハンドリングのテストも含む

### テスト実行結果
```bash
$ go test ./internal/core/ecs -v -run TestEntityManager_CreateEntity
=== RUN   TestEntityManager_CreateEntity
=== RUN   TestEntityManager_CreateEntity/TC001:_Create_new_entity
--- FAIL: TestEntityManager_CreateEntity (0.00s)
    --- FAIL: TestEntityManager_CreateEntity/TC001:_Create_new_entity (0.00s)
panic: not implemented
```

✅ **Red段階完了**: 期待通り全テストが失敗することを確認

## エラー統合の課題と解決

### 問題
既存の`errors.go`ファイルとEntityManagerのエラー定義が競合

### 解決策
- 既存のECSエラーフレームワークを活用
- `ECSError`構造体と定数を使用
- テストケースでエラータイプチェックを適切に修正

## 次のステップ
Green段階で実際の実装を行い、テストが通るようにします。
