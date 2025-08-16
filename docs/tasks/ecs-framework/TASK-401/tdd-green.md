# TASK-401: パフォーマンス最適化 - 最小実装 (Green段階)

## 概要

TDDのGreen段階として、失敗したテストが通るための最小限の実装を行います。過度な機能実装を避け、テストが通る必要最小限のコードのみを追加します。

## Red段階で確認された失敗テスト

### ❌ TestOptimizedComponentStore_SoALayout
```
Error: Not equal: expected: 1000, actual: 0
```
**原因**: `GetTransformArray()`が`nil`を返している（空実装）

### ✅ TestOptimizedComponentStore_Prefetch  
**状況**: PASSしているが、実装が空のため実際の機能はない

## Green段階の実装方針

### 1. 必要最小限の実装のみ
- テストが通るための最小コードのみ追加
- パフォーマンス最適化は後のRefactor段階で実装
- 複雑なSIMDやキャッシュ最適化は含めない

### 2. 実装優先順位
1. **高優先度**: 失敗しているテストを通すための実装
2. **中優先度**: PASSしているが機能がないテストの基本実装  
3. **低優先度**: まだ作成していないテストの準備

---

## 最小実装

### OptimizedComponentStore の最小実装

```go
package cache

import (
	. "muscle-dreamer/internal/core/ecs/optimizations"
)

// OptimizedComponentStore はコンポーネントストアの最小実装
type OptimizedComponentStore struct {
	// 最小限のデータストレージ
	transforms map[EntityID]TransformComponent
	sprites    map[EntityID]SpriteComponent
	
	// SoA配列（テスト用）
	transformArray []TransformComponent
}

// NewOptimizedComponentStore creates a new optimized component store
func NewOptimizedComponentStore() *OptimizedComponentStore {
	return &OptimizedComponentStore{
		transforms:     make(map[EntityID]TransformComponent),
		sprites:        make(map[EntityID]SpriteComponent),
		transformArray: make([]TransformComponent, 0),
	}
}

// AddTransform adds a transform component
func (cs *OptimizedComponentStore) AddTransform(entityID EntityID, component TransformComponent) {
	cs.transforms[entityID] = component
	cs.transformArray = append(cs.transformArray, component)
}

// GetTransform gets a transform component
func (cs *OptimizedComponentStore) GetTransform(entityID EntityID) *TransformComponent {
	if component, exists := cs.transforms[entityID]; exists {
		return &component
	}
	return nil
}

// GetTransformArray returns the transform array for SoA access
func (cs *OptimizedComponentStore) GetTransformArray() []TransformComponent {
	return cs.transformArray
}

// PrefetchComponents prefetches components (minimal implementation)
func (cs *OptimizedComponentStore) PrefetchComponents(entities []EntityID) {
	// 最小実装: 実際のプリフェッチはせず、メモリアクセスのみ
	for _, entityID := range entities {
		_ = cs.transforms[entityID] // メモリアクセス
	}
}

// RemoveTransform removes a transform component
func (cs *OptimizedComponentStore) RemoveTransform(entityID EntityID) {
	delete(cs.transforms, entityID)
	
	// transformArray からも削除（簡易実装）
	cs.rebuildTransformArray()
}

// AddSprite adds a sprite component
func (cs *OptimizedComponentStore) AddSprite(entityID EntityID, component SpriteComponent) {
	cs.sprites[entityID] = component
}

// RemoveSprite removes a sprite component
func (cs *OptimizedComponentStore) RemoveSprite(entityID EntityID) {
	delete(cs.sprites, entityID)
}

// rebuildTransformArray rebuilds the transform array after removal
func (cs *OptimizedComponentStore) rebuildTransformArray() {
	cs.transformArray = cs.transformArray[:0] // クリア
	for _, transform := range cs.transforms {
		cs.transformArray = append(cs.transformArray, transform)
	}
}
```

## 実装のポイント

### ✅ 最小限の要件を満たす設計
1. **テスト通過優先**: 失敗したテストが通るための最小コード
2. **シンプルなデータ構造**: map + slice の組み合わせ
3. **基本機能のみ**: 高度な最適化は含めない

### 🔄 後のRefactor段階で追加予定
1. **キャッシュライン整列**: 64バイト境界整列
2. **メモリプリフェッチ**: CPU命令活用
3. **SoA最適化**: 真のStructure of Arrays実装
4. **SIMD演算**: ベクトル処理最適化

---

## 実装実行

### ファイル更新
```bash
# 最小実装でテストを通す
go test ./internal/core/ecs/optimizations/cache -v
```

### 期待される結果
```
=== RUN   TestOptimizedComponentStore_SoALayout
--- PASS: TestOptimizedComponentStore_SoALayout (0.00s)
=== RUN   TestOptimizedComponentStore_Prefetch
--- PASS: TestOptimizedComponentStore_Prefetch (0.00s)
PASS
```

### テスト通過後の確認事項
1. **機能動作確認**: 基本的なCRUD操作が動作する
2. **メモリ安全性**: ポインタアクセスでクラッシュしない
3. **データ整合性**: map と slice の整合性が保たれる

---

## Green段階完了条件

### 必須条件
- [ ] 全テストがPASSする
- [ ] 基本的なコンポーネント操作が動作する  
- [ ] メモリ安全性が確保されている
- [ ] コンパイルエラーがない

### 確認事項
- [ ] テスト実行: `go test ./internal/core/ecs/optimizations/cache -v`
- [ ] 基本動作確認: 手動でのCRUD操作テスト
- [ ] エッジケース: 存在しないエンティティへのアクセス

---

## 次のステップ (Refactor段階)

Green段階完了後、**Refactor段階**で以下を実装：

### パフォーマンス最適化
1. **メモリレイアウト最適化**: 真のSoA実装
2. **キャッシュ効率向上**: アライメント、プリフェッチ
3. **SIMD演算統合**: ベクトル処理実装
4. **並列処理対応**: 並行安全性

### コード品質向上
1. **エラーハンドリング強化**
2. **パフォーマンス測定追加**
3. **メモリプロファイリング**
4. **ドキュメント整備**

---

**実装ステータス**: 🟢 Green段階 - 最小実装でテストを通す  
**次のフェーズ**: 🔄 Refactor段階 - パフォーマンス最適化とコード改善  
**作成日**: 2025-08-11  
**最終更新**: 2025-08-11