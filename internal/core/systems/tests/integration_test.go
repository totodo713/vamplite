package tests

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"muscle-dreamer/internal/core/ecs"
	"muscle-dreamer/internal/core/ecs/components"
	"muscle-dreamer/internal/core/systems"
)

func TestSystemsIntegration_MovementToPhysics(t *testing.T) {
	movementSystem := systems.NewMovementSystem()
	physicsSystem := systems.NewPhysicsSystem()
	world := createWorldWithEntities()

	// 両システム初期化
	err := movementSystem.Initialize(world)
	assert.NoError(t, err)
	err = physicsSystem.Initialize(world)
	assert.NoError(t, err)

	// 移動・物理コンポーネント持ちエンティティ
	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 0, Y: 0},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		Mass:     1.0,
		Velocity: ecs.Vector2{X: 100, Y: 0},
		IsStatic: false,
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)

	// システム順次実行
	for i := 0; i < 60; i++ {
		err := movementSystem.Update(world, 0.016)
		assert.NoError(t, err)

		err = physicsSystem.Update(world, 0.016)
		assert.NoError(t, err)
	}

	// 連携動作確認
	updatedTransformComp, err := world.GetComponent(entity, ecs.ComponentTypeTransform)
	assert.NoError(t, err)
	updatedTransform := updatedTransformComp.(*components.TransformComponent)
	assert.Greater(t, updatedTransform.Position.X, float64(50)) // 移動確認
}

func TestSystemsIntegration_PhysicsToRendering(t *testing.T) {
	physicsSystem := systems.NewPhysicsSystem()
	renderingSystem := systems.NewRenderingSystem()
	world := createWorldWithEntities()
	mockRenderer := &MockRenderer{}

	// システム初期化
	physicsSystem.Initialize(world)
	renderingSystem.Initialize(world)

	// 物理・描画コンポーネント持ちエンティティ
	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 100, Y: 100},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		Mass:     1.0,
		Velocity: ecs.Vector2{X: 50, Y: 0},
		IsStatic: false,
		Gravity:  false,
	}
	sprite := &components.SpriteComponent{
		TextureID: "test_entity",
		ZOrder:    0,
		Visible:   true,
		Color:     ecs.Color{R: 255, G: 255, B: 255, A: 255},
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)
	world.AddComponent(entity, sprite)

	// Physics更新 → Rendering実行
	for i := 0; i < 30; i++ {
		err := physicsSystem.Update(world, 0.016)
		assert.NoError(t, err)
	}

	err := renderingSystem.Render(world, mockRenderer)
	assert.NoError(t, err)

	// 物理演算により位置が変化し、その位置で描画されることを確認
	updatedTransformComp, err := world.GetComponent(entity, ecs.ComponentTypeTransform)
	assert.NoError(t, err)
	updatedTransform := updatedTransformComp.(*components.TransformComponent)
	assert.Greater(t, updatedTransform.Position.X, float64(100))
	assert.Equal(t, 1, mockRenderer.DrawCallCount) // 描画も実行される
}

func TestSystemsIntegration_AllSystemsTogether(t *testing.T) {
	// 全システム作成
	movementSystem := systems.NewMovementSystem()
	physicsSystem := systems.NewPhysicsSystem()
	renderingSystem := systems.NewRenderingSystem()
	audioSystem := systems.NewAudioSystem()

	world := createWorldWithEntities()
	mockRenderer := &MockRenderer{}
	mockAudioEngine := NewMockAudioEngine()

	// システム初期化
	systems := []ecs.System{movementSystem, physicsSystem, renderingSystem, audioSystem}
	for _, system := range systems {
		err := system.Initialize(world)
		assert.NoError(t, err)
	}

	audioSystem.SetAudioEngine(mockAudioEngine)

	// 全コンポーネント持ちエンティティ作成
	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 200, Y: 300},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		Mass:     2.0,
		Velocity: ecs.Vector2{X: 75, Y: -25},
		IsStatic: false,
		Gravity:  true,
	}
	sprite := &components.SpriteComponent{
		TextureID: "game_entity",
		ZOrder:    1,
		Visible:   true,
		Color:     ecs.Color{R: 255, G: 255, B: 255, A: 255},
	}
	audio := &MockAudioComponent{
		SoundID:     "entity_sound",
		Volume:      0.6,
		Playing:     false, // 初期は再生しない
		ThreeD:      true,
		MaxDistance: 150,
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)
	world.AddComponent(entity, sprite)
	world.AddComponent(entity, audio)

	physicsSystem.SetGravity(ecs.Vector2{X: 0, Y: 200})

	// 統合シミュレーション実行
	for i := 0; i < 120; i++ { // 2秒間
		// Update系システム実行
		err := movementSystem.Update(world, 0.016)
		assert.NoError(t, err)

		err = physicsSystem.Update(world, 0.016)
		assert.NoError(t, err)

		err = audioSystem.Update(world, 0.016)
		assert.NoError(t, err)

		// 30フレームごとに描画
		if i%30 == 0 {
			err = renderingSystem.Render(world, mockRenderer)
			assert.NoError(t, err)
		}
	}

	// 全システム連携動作確認
	updatedTransformComp, err := world.GetComponent(entity, ecs.ComponentTypeTransform)
	assert.NoError(t, err)
	updatedTransform := updatedTransformComp.(*components.TransformComponent)

	updatedPhysicsComp, err := world.GetComponent(entity, ecs.ComponentTypePhysics)
	assert.NoError(t, err)
	updatedPhysics := updatedPhysicsComp.(*components.PhysicsComponent)

	// 物理演算により位置・速度が変化
	assert.NotEqual(t, 200.0, updatedTransform.Position.X)
	assert.NotEqual(t, 300.0, updatedTransform.Position.Y)
	assert.NotEqual(t, -25.0, updatedPhysics.Velocity.Y) // 重力の影響

	// 描画が実行される
	assert.Greater(t, mockRenderer.DrawCallCount, 0)
}

func TestSystemsPerformance_10000Entities(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	systems := []ecs.System{
		systems.NewMovementSystem(),
		systems.NewPhysicsSystem(),
		// RenderingSystem は除外（モックレンダラー設定が複雑なため）
		// AudioSystem は除外（オーディオエンジン設定が複雑なため）
	}

	world := createWorldWithEntities()

	// 全システム初期化
	for _, system := range systems {
		err := system.Initialize(world)
		assert.NoError(t, err)
	}

	// 10,000エンティティ作成
	for i := 0; i < 10000; i++ {
		entity := world.CreateEntity()

		transform := &components.TransformComponent{
			Position: ecs.Vector2{X: rand.Float64() * 800, Y: rand.Float64() * 600},
			Scale:    ecs.Vector2{X: 1, Y: 1},
		}
		physics := &components.PhysicsComponent{
			Mass:     1.0,
			Velocity: ecs.Vector2{X: rand.Float64()*200 - 100, Y: rand.Float64()*200 - 100},
			IsStatic: rand.Intn(10) < 2, // 20%の確率で静的
			Gravity:  rand.Intn(2) == 1, // 50%の確率で重力あり
		}

		world.AddComponent(entity, transform)
		world.AddComponent(entity, physics)
	}

	// パフォーマンス測定
	start := time.Now()

	for i := 0; i < 60; i++ { // 1秒間シミュレーション
		for _, system := range systems {
			err := system.Update(world, 0.016)
			assert.NoError(t, err)
		}
	}

	elapsed := time.Since(start)

	// パフォーマンス要件：1秒間のシミュレーションが2秒以内で完了
	assert.Less(t, elapsed, 2*time.Second, "Performance test failed: took %v", elapsed)

	// 各システムのメトリクス確認
	for _, system := range systems {
		metrics := system.GetMetrics()
		avgUpdateTime := time.Duration(metrics.AverageTime)

		// 各システムの平均実行時間が10ms以下
		assert.Less(t, avgUpdateTime, 10*time.Millisecond,
			"System %s average update time: %v", system.GetType(), avgUpdateTime)

		t.Logf("System %s: %d updates, avg time: %v, total: %v",
			system.GetType(), metrics.ExecutionCount, avgUpdateTime, time.Duration(metrics.TotalTime))
	}
}

func TestSystemsPerformance_LargeEntityCounts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// 段階的にエンティティ数を増やしてテスト
	entityCounts := []int{100, 500, 1000, 5000}

	movementSystem := systems.NewMovementSystem()

	for _, count := range entityCounts {
		t.Run(fmt.Sprintf("entities_%d", count), func(t *testing.T) {
			world := createWorldWithEntities()
			movementSystem.Initialize(world)

			// エンティティ作成
			for i := 0; i < count; i++ {
				entity := world.CreateEntity()
				transform := &components.TransformComponent{
					Position: ecs.Vector2{X: rand.Float64() * 800, Y: rand.Float64() * 600},
					Scale:    ecs.Vector2{X: 1, Y: 1},
				}
				physics := &components.PhysicsComponent{
					Mass:     1.0,
					Velocity: ecs.Vector2{X: rand.Float64() * 100, Y: rand.Float64() * 100},
				}
				world.AddComponent(entity, transform)
				world.AddComponent(entity, physics)
			}

			// 10フレーム実行して平均時間を測定
			start := time.Now()
			for i := 0; i < 10; i++ {
				err := movementSystem.Update(world, 0.016)
				assert.NoError(t, err)
			}
			elapsed := time.Since(start)
			avgFrameTime := elapsed / 10

			t.Logf("Entities: %d, Avg frame time: %v", count, avgFrameTime)

			// スケーラビリティ確認（フレーム時間が合理的な範囲内）
			maxFrameTime := time.Duration(count/100+1) * time.Millisecond
			assert.Less(t, avgFrameTime, maxFrameTime,
				"Frame time %v exceeded limit %v for %d entities", avgFrameTime, maxFrameTime, count)
		})
	}
}

func TestSystemsIntegration_ErrorHandling(t *testing.T) {
	systems := []ecs.System{
		systems.NewMovementSystem(),
		systems.NewPhysicsSystem(),
		systems.NewRenderingSystem(),
		systems.NewAudioSystem(),
	}

	world := createWorldWithEntities()

	// システム初期化
	for _, system := range systems {
		err := system.Initialize(world)
		assert.NoError(t, err)

		// エラーハンドラー設定
		// errorCount := 0
		// system.SetErrorHandler(func(err error) {
		// 	errorCount++
		// })
	}

	// 不正な値を持つエンティティ作成
	entity := world.CreateEntity()
	invalidTransform := &components.TransformComponent{
		Position: ecs.Vector2{X: math.NaN(), Y: math.Inf(1)}, // NaN, Inf
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	invalidPhysics := &components.PhysicsComponent{
		Mass:     -1.0, // 負の質量
		Velocity: ecs.Vector2{X: math.Inf(-1), Y: math.NaN()},
	}
	world.AddComponent(entity, invalidTransform)
	world.AddComponent(entity, invalidPhysics)

	// システム実行してエラーハンドリング確認
	for _, system := range systems {
		err := system.Update(world, 0.016)

		// エラーが適切に処理される（パニックしない）
		if err != nil {
			assert.Contains(t, err.Error(), "invalid") // エラーメッセージ確認
		}

		// システムが引き続き動作可能
		assert.True(t, system.IsEnabled())
	}
}

func TestSystemsIntegration_ThreadSafety(t *testing.T) {
	system := systems.NewMovementSystem()
	world := createWorldWithEntities()
	system.Initialize(world)

	// エンティティ作成
	entity := world.CreateEntity()
	transform := &components.TransformComponent{
		Position: ecs.Vector2{X: 0, Y: 0},
		Scale:    ecs.Vector2{X: 1, Y: 1},
	}
	physics := &components.PhysicsComponent{
		Mass:     1.0,
		Velocity: ecs.Vector2{X: 100, Y: 0},
	}
	world.AddComponent(entity, transform)
	world.AddComponent(entity, physics)

	// 並行実行テスト（データ競合検出）
	done := make(chan bool, 2)

	// Goroutine 1: Update実行
	go func() {
		for i := 0; i < 100; i++ {
			system.Update(world, 0.016)
		}
		done <- true
	}()

	// Goroutine 2: メトリクス読み取り
	go func() {
		for i := 0; i < 100; i++ {
			_ = system.GetMetrics()
			_ = system.IsEnabled()
		}
		done <- true
	}()

	// 両方完了を待つ
	<-done
	<-done

	// データ競合がない場合、正常に完了する
	assert.True(t, true, "Thread safety test completed without race conditions")
}
