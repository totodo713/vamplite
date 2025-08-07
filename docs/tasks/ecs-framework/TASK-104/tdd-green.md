# TASK-104: 基本システム実装 - Green段階（最小実装）

## 実装方針

TDD の Green 段階として、Red段階で作成した失敗テストを1つずつ通すように最小限の実装を行います。この段階では：
1. テストが成功するための最小限の機能実装
2. 過度な設計や最適化は避ける
3. 1つのテストが通ったら次のテストに進む
4. すべての基本機能が動作する状態を目指す

## 修正が必要な問題

### 1. GetComponentメソッドの戻り値修正

現在のECSフレームワークでは`GetComponent`が`(Component, error)`を返すため、テストコードを修正する必要があります。

### 2. システム実装の完成

Red段階では空実装のため、実際の機能実装が必要です。

## 実装ステップ

### Step 1: テストコード修正

まず、コンパイルエラーを解決するためにテストコードを修正します。

### Step 2: MovementSystem実装

最もシンプルなMovementSystemから実装を開始します。

### Step 3: RenderingSystem実装

描画機能を実装します。

### Step 4: PhysicsSystem実装

物理演算機能を実装します。

### Step 5: AudioSystem実装

音声機能を実装します。

### Step 6: 統合テスト確認

全システムが連携して動作することを確認します。

## 実装内容

### Step 1: テストコード修正

GetComponentの戻り値を適切に処理するようにテストを修正します。

```go
// 修正前
updatedTransform := world.GetComponent(entity, ecs.ComponentTypeTransform).(*components.TransformComponent)

// 修正後
component, err := world.GetComponent(entity, ecs.ComponentTypeTransform)
assert.NoError(t, err)
updatedTransform := component.(*components.TransformComponent)
```

### Step 2: MovementSystem最小実装

基本的な位置更新機能を実装します：

```go
func (ms *MovementSystem) Update(world ecs.World, deltaTime float64) error {
	entities := world.GetEntitiesWithComponents(ms.GetRequiredComponents()...)
	
	for _, entity := range entities {
		transformComp, err := world.GetComponent(entity, ecs.ComponentTypeTransform)
		if err != nil {
			continue
		}
		
		transform := transformComp.(*components.TransformComponent)
		
		// 基本的な位置更新: position += velocity * deltaTime
		transform.Position.X += transform.Velocity.X * deltaTime
		transform.Position.Y += transform.Velocity.Y * deltaTime
		
		// 回転更新
		transform.Rotation += transform.AngularVelocity * deltaTime
	}
	
	return nil
}
```

### Step 3: RenderingSystem最小実装

基本的な描画機能を実装します：

```go
func (rs *RenderingSystem) Render(world ecs.World, renderer Renderer) error {
	entities := world.GetEntitiesWithComponents(rs.GetRequiredComponents()...)
	
	// Z-Orderでソート（後で実装）
	for _, entity := range entities {
		transformComp, err := world.GetComponent(entity, ecs.ComponentTypeTransform)
		if err != nil {
			continue
		}
		spriteComp, err := world.GetComponent(entity, ecs.ComponentTypeSprite)
		if err != nil {
			continue
		}
		
		transform := transformComp.(*components.TransformComponent)
		sprite := spriteComp.(*components.SpriteComponent)
		
		// 基本描画呼び出し
		renderer.DrawSprite(sprite.TextureID, transform.Position.X, transform.Position.Y, 
			transform.Scale.X, transform.Scale.Y, transform.Rotation)
	}
	
	return nil
}
```

### Step 4: PhysicsSystem最小実装

基本的な物理演算を実装します：

```go
func (ps *PhysicsSystem) Update(world ecs.World, deltaTime float64) error {
	entities := world.GetEntitiesWithComponents(ps.GetRequiredComponents()...)
	
	for _, entity := range entities {
		physicsComp, err := world.GetComponent(entity, ecs.ComponentTypePhysics)
		if err != nil {
			continue
		}
		
		physics := physicsComp.(*components.PhysicsComponent)
		
		// 重力適用
		if ps.gravity.Y != 0 {
			physics.Velocity.Y += ps.gravity.Y * deltaTime
		}
		
		// 基本的な速度制限
		if ps.maxSpeed > 0 {
			speed := math.Sqrt(physics.Velocity.X*physics.Velocity.X + physics.Velocity.Y*physics.Velocity.Y)
			if speed > ps.maxSpeed {
				scale := ps.maxSpeed / speed
				physics.Velocity.X *= scale
				physics.Velocity.Y *= scale
			}
		}
	}
	
	return nil
}
```

### Step 5: AudioSystem最小実装

基本的な音声機能を実装します：

```go
func (as *AudioSystem) Update(world ecs.World, deltaTime float64) error {
	entities := world.GetEntitiesWithComponents(as.GetRequiredComponents()...)
	
	for _, entity := range entities {
		audioComp, err := world.GetComponent(entity, ecs.ComponentTypeAudio)
		if err != nil {
			continue
		}
		
		audio := audioComp.(*components.AudioComponent)
		
		if audio.IsPlaying && as.audioEngine != nil {
			// 基本的な音声再生
			as.audioEngine.PlaySound(audio.SoundID, audio.Volume)
		}
	}
	
	return nil
}
```

## 実装完了基準

以下のテストが成功すればGreen段階完了：

### 基本機能テスト
- [ ] MovementSystem: 位置・回転更新
- [ ] RenderingSystem: 基本描画
- [ ] PhysicsSystem: 重力・速度制限
- [ ] AudioSystem: 音声再生

### インターフェーステスト
- [ ] 全システムのSystem インターフェース実装確認
- [ ] GetType, GetPriority, GetRequiredComponents実装確認

### エラーハンドリング
- [ ] 不正なコンポーネントアクセス時のエラー処理
- [ ] nil チェック・範囲外チェック

## 次のステップ

Green段階完了後、Refactor段階で：
1. パフォーマンス最適化
2. エラーハンドリング強化
3. 高度な機能実装（カリング、衝突検出、3Dオーディオ等）
4. 統合テスト・負荷テスト実装

---

## Green段階実装開始

この方針に基づいて、段階的にシステム実装を完成させていきます。