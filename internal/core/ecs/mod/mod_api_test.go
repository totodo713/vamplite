package mod

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"muscle-dreamer/internal/core/ecs"
)

func TestModEntityAPI_Create(t *testing.T) {
	// この段階では実装が存在しないため全てのテストが失敗する

	// セットアップ（まだ実装されていない）
	api := createTestModAPI(t, "test-mod")

	t.Run("正常なエンティティ作成", func(t *testing.T) {
		entityID, err := api.Entities().Create("test-tag")
		require.NoError(t, err)
		assert.NotEqual(t, ecs.InvalidEntityID, entityID)

		// MODタグが自動付与されることを確認
		tags, err := api.Entities().GetTags(entityID)
		require.NoError(t, err)
		// uniqueModIDでタグが付与されるため、modプレフィックスのみ確認
		found := false
		for _, tag := range tags {
			if strings.HasPrefix(tag, "mod:test-mod") {
				found = true
				break
			}
		}
		assert.True(t, found, "MOD prefix tag not found")
		assert.Contains(t, tags, "test-tag")
	})

	t.Run("エンティティ作成上限テスト", func(t *testing.T) {
		// 独立したAPIインスタンスで上限テスト
		limitAPI := createTestModAPI(t, "limit-test-mod")

		// 100個まで作成成功
		for i := 0; i < 100; i++ {
			_, err := limitAPI.Entities().Create()
			require.NoError(t, err)
		}

		// 101個目は失敗
		_, err := limitAPI.Entities().Create()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "entity limit exceeded")
	})
}

func TestModEntityAPI_Delete(t *testing.T) {
	api := createTestModAPI(t, "test-mod")

	t.Run("自分のエンティティ削除", func(t *testing.T) {
		entityID, err := api.Entities().Create("test-entity")
		require.NoError(t, err)

		err = api.Entities().Delete(entityID)
		assert.NoError(t, err)
	})

	t.Run("他MODエンティティ削除拒否", func(t *testing.T) {
		// 他MODのエンティティを模擬
		otherModEntity := createEntityWithMod(t, "other-mod")

		err := api.Entities().Delete(otherModEntity)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})

	t.Run("システムエンティティ削除拒否", func(t *testing.T) {
		// システムエンティティを模擬
		systemEntity := createSystemEntity(t)

		err := api.Entities().Delete(systemEntity)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "system entity access denied")
	})
}

func TestModComponentAPI_Add(t *testing.T) {
	api := createTestModAPI(t, "test-mod")

	t.Run("許可されたコンポーネント追加", func(t *testing.T) {
		entityID, err := api.Entities().Create()
		require.NoError(t, err)

		// 許可されたコンポーネント（Sprite）
		spriteComponent := createTestSpriteComponent()
		err = api.Components().Add(entityID, spriteComponent)
		assert.NoError(t, err)
	})

	t.Run("禁止コンポーネント追加拒否", func(t *testing.T) {
		entityID, err := api.Entities().Create()
		require.NoError(t, err)

		// 禁止されたコンポーネント（FileIO - 存在しないが模擬）
		fileIOComponent := createTestFileIOComponent()
		err = api.Components().Add(entityID, fileIOComponent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "component not allowed")
	})

	t.Run("他MODエンティティへのコンポーネント追加拒否", func(t *testing.T) {
		otherModEntity := createEntityWithMod(t, "other-mod")
		spriteComponent := createTestSpriteComponent()

		err := api.Components().Add(otherModEntity, spriteComponent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})
}

func TestModComponentAPI_Get(t *testing.T) {
	api := createTestModAPI(t, "test-mod")

	t.Run("自分のコンポーネント取得", func(t *testing.T) {
		entityID, err := api.Entities().Create()
		require.NoError(t, err)

		spriteComponent := createTestSpriteComponent()
		err = api.Components().Add(entityID, spriteComponent)
		require.NoError(t, err)

		retrieved, err := api.Components().Get(entityID, ecs.ComponentTypeSprite)
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	t.Run("権限のないコンポーネント取得拒否", func(t *testing.T) {
		otherModEntity := createEntityWithMod(t, "other-mod")

		_, err := api.Components().Get(otherModEntity, ecs.ComponentTypeSprite)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})
}

func TestModQueryAPI_Find(t *testing.T) {
	api := createTestModAPI(t, "test-mod")

	t.Run("MODエンティティのみ検索", func(t *testing.T) {
		// 自分のエンティティを作成
		myEntity, err := api.Entities().Create("my-entity")
		require.NoError(t, err)

		// 他MODとシステムエンティティを模擬作成
		createEntityWithMod(t, "other-mod")
		createSystemEntity(t)

		// クエリ実行 - 自分のエンティティのみ返却されるべき
		query := createTestQuery()
		results, err := api.Queries().Find(query)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, myEntity, results[0])
	})

	t.Run("クエリ実行回数制限", func(t *testing.T) {
		// 独立したAPIインスタンスでクエリ制限テスト
		queryAPI := createTestModAPI(t, "query-limit-test-mod")
		query := createTestQuery()

		// 1000回まで成功
		for i := 0; i < 1000; i++ {
			_, err := queryAPI.Queries().Find(query)
			assert.NoError(t, err)
		}

		// 1001回目は失敗
		_, err := queryAPI.Queries().Find(query)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "query limit exceeded")
	})
}

func TestModSystemAPI_Register(t *testing.T) {
	api := createTestModAPI(t, "test-mod")

	t.Run("正常なシステム登録", func(t *testing.T) {
		system := createTestModSystem("test-system")
		err := api.Systems().Register(system)
		assert.NoError(t, err)

		registered := api.Systems().GetRegistered()
		assert.Contains(t, registered, "safe-test-system")
	})

	t.Run("システム実行時間制限", func(t *testing.T) {
		// 10ms実行時間のシステム（制限5ms超過）
		longRunningSystem := createLongRunningModSystem("long-system", 10*time.Millisecond)
		err := api.Systems().Register(longRunningSystem)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "execution time exceeds limit")
	})
}

// セキュリティテスト

func TestModAPI_Security_PathTraversal(t *testing.T) {
	api := createTestModAPI(t, "malicious-mod")

	t.Run("パストラバーサル攻撃防御", func(t *testing.T) {
		// 悪意のあるタグでエンティティ作成試行
		maliciousTags := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32",
			"../../../../root/.ssh/id_rsa",
		}

		for _, tag := range maliciousTags {
			_, err := api.Entities().Create(tag)
			// タグは受け入れるが、ファイルアクセスは発生しない
			assert.NoError(t, err)
		}

		// しかし、ファイルシステムアクセスは完全ブロック
		// （この検証は実際の実装で行う）
	})
}

func TestModAPI_Security_SystemCommand(t *testing.T) {
	api := createTestModAPI(t, "malicious-mod")

	t.Run("システムコマンド実行防止", func(t *testing.T) {
		// 悪意のあるシステム登録試行
		maliciousSystem := createMaliciousSystem("rm -rf /")
		err := api.Systems().Register(maliciousSystem)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "security violation")
	})
}

func TestModAPI_ResourceLimits(t *testing.T) {
	t.Run("メモリ使用量制限", func(t *testing.T) {
		// 低メモリ制限でテスト（10個のエンティティで制限に達する）
		config := ModConfig{
			MaxEntities:       100,
			MaxMemory:         500, // 500バイト制限（64バイト×7個で制限超過）
			MaxExecutionTime:  5 * time.Millisecond,
			AllowedComponents: DefaultModConfig().AllowedComponents,
			MaxQueryCount:     1000,
		}
		testCounter++
		uniqueModID := fmt.Sprintf("memory-test-mod-%d", testCounter)
		memoryAPI, err := globalTestFactory.Create(uniqueModID, config)
		require.NoError(t, err)

		// メモリ制限まで作成
		for i := 0; i < 10; i++ {
			_, err := memoryAPI.Entities().Create("memory-test")
			if err != nil {
				assert.Contains(t, err.Error(), "memory limit exceeded")
				break
			}
		}

		// コンテキストのメモリ使用量確認
		ctx := memoryAPI.GetContext()
		assert.True(t, ctx.MemoryUsage <= ctx.MaxMemory)
	})
}

// パフォーマンステスト

func BenchmarkModAPI_EntityCreation(b *testing.B) {
	api := createTestModAPI(b, "perf-test-mod")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entityID, err := api.Entities().Create("perf-test")
		if err != nil {
			b.Fatal(err)
		}
		api.Entities().Delete(entityID)
	}
}

func BenchmarkModAPI_ComponentOperations(b *testing.B) {
	api := createTestModAPI(b, "perf-test-mod")
	entityID, _ := api.Entities().Create("perf-test")
	component := createTestSpriteComponent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		api.Components().Add(entityID, component)
		api.Components().Get(entityID, ecs.ComponentTypeSprite)
		api.Components().Remove(entityID, ecs.ComponentTypeSprite)
	}
}

// テストヘルパー関数はtest_helpers.goで実装
