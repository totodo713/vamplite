# ECS Framework API Endpoints

## 概要

Muscle Dreamer ECSフレームワークのREST API仕様です。主にデバッグ、監視、MOD管理、セーブデータ管理に使用されます。WebAssembly版でのブラウザ統合やリモート管理ツールからのアクセスを想定しています。

## 認証・セキュリティ

### 認証方式
- **ローカル開発**: 認証なし（localhost限定）
- **デバッグモード**: API Key ヘッダー認証
- **プロダクション**: 無効化推奨

### セキュリティヘッダー
```http
X-API-Key: your-debug-api-key
Content-Type: application/json
Accept: application/json
```

## エラーレスポンス形式

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": {
      "field": "Additional error details"
    }
  },
  "timestamp": "2025-08-03T12:00:00Z"
}
```

## 成功レスポンス形式

```json
{
  "success": true,
  "data": {
    // Response data
  },
  "meta": {
    "total": 100,
    "page": 1,
    "limit": 20
  },
  "timestamp": "2025-08-03T12:00:00Z"
}
```

---

## 1. World Management API

### GET /api/v1/world/status
ワールドの現在状態を取得

**レスポンス:**
```json
{
  "success": true,
  "data": {
    "world_id": "world_001",
    "state": "running",
    "entity_count": 1250,
    "component_count": 4500,
    "system_count": 12,
    "memory_usage": {
      "allocated": 67108864,
      "used": 45232128,
      "peak": 78643200
    },
    "performance": {
      "fps": 59.8,
      "frame_time_ms": 16.7,
      "update_time_ms": 8.2,
      "render_time_ms": 6.1
    },
    "uptime_seconds": 3600
  }
}
```

### POST /api/v1/world/pause
ワールドの実行を一時停止

**リクエスト:**
```json
{
  "reason": "debug_inspection"
}
```

### POST /api/v1/world/resume
ワールドの実行を再開

### GET /api/v1/world/metrics
詳細なパフォーマンスメトリクスを取得

**クエリパラメータ:**
- `from`: 開始時刻 (ISO 8601)
- `to`: 終了時刻 (ISO 8601)
- `interval`: サンプリング間隔 (seconds)

**レスポンス:**
```json
{
  "success": true,
  "data": {
    "metrics": [
      {
        "timestamp": "2025-08-03T12:00:00Z",
        "entity_count": 1250,
        "frame_time_us": 16700,
        "memory_used": 45232128,
        "systems": [
          {
            "type": "physics",
            "execution_time_us": 3200,
            "entities_processed": 450
          }
        ]
      }
    ],
    "summary": {
      "avg_fps": 59.2,
      "avg_frame_time_ms": 16.9,
      "peak_memory_mb": 75.1
    }
  }
}
```

---

## 2. Entity Management API

### GET /api/v1/entities
エンティティ一覧を取得

**クエリパラメータ:**
- `page`: ページ番号 (default: 1)
- `limit`: 1ページの件数 (default: 20, max: 100)
- `components`: フィルタするコンポーネント型 (カンマ区切り)
- `tag`: エンティティタグでフィルタ

**レスポンス:**
```json
{
  "success": true,
  "data": {
    "entities": [
      {
        "id": 12345,
        "components": ["transform", "sprite", "physics"],
        "tag": "player",
        "created_at": "2025-08-03T12:00:00Z"
      }
    ]
  },
  "meta": {
    "total": 1250,
    "page": 1,
    "limit": 20,
    "pages": 63
  }
}
```

### GET /api/v1/entities/{entityId}
特定エンティティの詳細を取得

**レスポンス:**
```json
{
  "success": true,
  "data": {
    "id": 12345,
    "tag": "player",
    "components": {
      "transform": {
        "position": {"x": 100.5, "y": 200.0},
        "rotation": 0.0,
        "scale": {"x": 1.0, "y": 1.0}
      },
      "sprite": {
        "image_id": "player.png",
        "color": {"r": 255, "g": 255, "b": 255, "a": 255},
        "visible": true
      }
    },
    "parent": null,
    "children": [12346, 12347],
    "created_at": "2025-08-03T12:00:00Z"
  }
}
```

### POST /api/v1/entities
新しいエンティティを作成

**リクエスト:**
```json
{
  "tag": "debug_entity",
  "components": {
    "transform": {
      "position": {"x": 0, "y": 0},
      "rotation": 0,
      "scale": {"x": 1, "y": 1}
    }
  }
}
```

### PUT /api/v1/entities/{entityId}/components/{componentType}
エンティティのコンポーネントを更新

**リクエスト:**
```json
{
  "position": {"x": 150.0, "y": 250.0},
  "rotation": 1.57
}
```

### DELETE /api/v1/entities/{entityId}
エンティティを削除

---

## 3. System Management API

### GET /api/v1/systems
システム一覧と状態を取得

**レスポンス:**
```json
{
  "success": true,
  "data": {
    "systems": [
      {
        "type": "physics",
        "name": "Physics System",
        "enabled": true,
        "priority": 70,
        "dependencies": [],
        "performance": {
          "avg_execution_time_us": 3200,
          "max_execution_time_us": 8500,
          "entities_processed": 450,
          "update_count": 108000
        }
      }
    ]
  }
}
```

### PUT /api/v1/systems/{systemType}/enabled
システムの有効/無効を切り替え

**リクエスト:**
```json
{
  "enabled": false,
  "reason": "debugging"
}
```

### GET /api/v1/systems/{systemType}/performance
システムの詳細パフォーマンス情報

**レスポンス:**
```json
{
  "success": true,
  "data": {
    "system_type": "physics",
    "performance_history": [
      {
        "timestamp": "2025-08-03T12:00:00Z",
        "execution_time_us": 3200,
        "entities_processed": 450
      }
    ],
    "statistics": {
      "avg_execution_time_us": 3200,
      "p95_execution_time_us": 5800,
      "p99_execution_time_us": 8500,
      "total_entities_processed": 48600000
    }
  }
}
```

---

## 4. Query API

### POST /api/v1/queries/execute
エンティティクエリを実行

**リクエスト:**
```json
{
  "query": {
    "with": ["transform", "sprite"],
    "without": ["physics"],
    "tags": ["enemy"],
    "limit": 50
  }
}
```

**レスポンス:**
```json
{
  "success": true,
  "data": {
    "entities": [12350, 12351, 12352],
    "execution_time_us": 120,
    "cache_hit": false
  }
}
```

### GET /api/v1/queries/performance
クエリパフォーマンス統計

**レスポンス:**
```json
{
  "success": true,
  "data": {
    "queries": [
      {
        "signature": "with:transform,sprite",
        "execution_count": 45000,
        "avg_execution_time_us": 85,
        "cache_hit_rate": 0.87,
        "last_executed": "2025-08-03T12:00:00Z"
      }
    ],
    "cache_stats": {
      "total_cached": 156,
      "hit_rate": 0.82,
      "evictions": 23
    }
  }
}
```

---

## 5. Save/Load API

### GET /api/v1/saves/slots
セーブスロット一覧を取得

**レスポンス:**
```json
{
  "success": true,
  "data": {
    "slots": [
      {
        "slot_number": 1,
        "save_name": "Main Game",
        "description": "Level 5 - Boss Fight",
        "thumbnail_url": "/api/v1/saves/1/thumbnail",
        "created_at": "2025-08-03T10:00:00Z",
        "updated_at": "2025-08-03T12:00:00Z",
        "play_time_seconds": 7200,
        "entity_count": 1250,
        "file_size": 524288
      }
    ]
  }
}
```

### POST /api/v1/saves/slots/{slotNumber}/save
現在のワールド状態をセーブ

**リクエスト:**
```json
{
  "save_name": "Boss Fight Checkpoint",
  "description": "Right before the final boss",
  "create_thumbnail": true,
  "compression": "gzip"
}
```

### POST /api/v1/saves/slots/{slotNumber}/load
セーブデータからワールド状態をロード

**レスポンス:**
```json
{
  "success": true,
  "data": {
    "load_time_ms": 850,
    "entities_loaded": 1250,
    "components_loaded": 4500,
    "world_state": "loaded"
  }
}
```

### DELETE /api/v1/saves/slots/{slotNumber}
セーブデータを削除

---

## 6. MOD Management API

### GET /api/v1/mods
インストール済みMOD一覧

**レスポンス:**
```json
{
  "success": true,
  "data": {
    "mods": [
      {
        "mod_id": "zombie_apocalypse",
        "name": "Zombie Apocalypse Mod",
        "version": "1.2.0",
        "author": "ModCreator",
        "enabled": true,
        "validated": true,
        "permissions": ["create_entities", "modify_ai"],
        "resource_usage": {
          "memory_mb": 12.5,
          "entity_count": 45,
          "cpu_time_ms": 2.3
        }
      }
    ]
  }
}
```

### PUT /api/v1/mods/{modId}/enabled
MODの有効/無効を切り替え

**リクエスト:**
```json
{
  "enabled": true,
  "validate_dependencies": true
}
```

### GET /api/v1/mods/{modId}/status
MODの詳細ステータス

**レスポンス:**
```json
{
  "success": true,
  "data": {
    "mod_id": "zombie_apocalypse",
    "status": "running",
    "validation_result": {
      "is_valid": true,
      "warnings": [],
      "errors": []
    },
    "security_checks": {
      "permissions_valid": true,
      "resource_limits_ok": true,
      "sandboxed": true
    },
    "performance": {
      "systems_registered": 3,
      "entities_created": 45,
      "avg_update_time_us": 1200
    }
  }
}
```

### POST /api/v1/mods/install
新しいMODをインストール

**リクエスト:**
```json
{
  "mod_package_url": "https://example.com/mod.zip",
  "validate_signature": true,
  "auto_enable": false
}
```

---

## 7. Debug API

### GET /api/v1/debug/memory
メモリ使用状況の詳細

**レスポンス:**
```json
{
  "success": true,
  "data": {
    "total_allocated": 67108864,
    "total_used": 45232128,
    "fragmentation": 0.05,
    "gc_stats": {
      "gc_count": 156,
      "last_gc_time": "2025-08-03T11:58:30Z",
      "gc_time_total_ms": 450
    },
    "pools": [
      {
        "component_type": "transform",
        "block_size": 64,
        "total_blocks": 10000,
        "used_blocks": 1250,
        "utilization": 0.125
      }
    ]
  }
}
```

### GET /api/v1/debug/components/{componentType}/stats
コンポーネント種別の統計情報

**レスポンス:**
```json
{
  "success": true,
  "data": {
    "component_type": "transform",
    "total_instances": 1250,
    "memory_used": 80000,
    "memory_reserved": 640000,
    "fragmentation": 0.02,
    "access_patterns": {
      "reads_per_frame": 3750,
      "writes_per_frame": 450,
      "cache_hit_rate": 0.95
    }
  }
}
```

### POST /api/v1/debug/snapshot
デバッグスナップショットを作成

**リクエスト:**
```json
{
  "include_entities": true,
  "include_components": true,
  "include_memory": true,
  "include_performance": true
}
```

### GET /api/v1/debug/logs
デバッグログを取得

**クエリパラメータ:**
- `level`: ログレベル (debug, info, warn, error)
- `component`: コンポーネント名でフィルタ
- `lines`: 取得行数 (default: 100)

---

## 8. WebSocket API

### WebSocket: /ws/v1/events
リアルタイムイベントストリーム

**接続時送信:**
```json
{
  "action": "subscribe",
  "events": ["entity_created", "entity_destroyed", "performance_alert"]
}
```

**受信イベント例:**
```json
{
  "event": "entity_created",
  "data": {
    "entity_id": 12500,
    "components": ["transform", "sprite"],
    "timestamp": "2025-08-03T12:00:00Z"
  }
}
```

### WebSocket: /ws/v1/metrics
リアルタイムメトリクスストリーム

**接続時設定:**
```json
{
  "action": "configure",
  "interval_ms": 1000,
  "metrics": ["fps", "memory", "entity_count"]
}
```

---

## エラーコード一覧

| コード | 説明 |
|--------|------|
| `WORLD_NOT_FOUND` | ワールドが見つからない |
| `ENTITY_NOT_FOUND` | エンティティが見つからない |
| `COMPONENT_NOT_FOUND` | コンポーネントが見つからない |
| `SYSTEM_NOT_FOUND` | システムが見つからない |
| `INVALID_QUERY` | 無効なクエリ |
| `SAVE_SLOT_NOT_FOUND` | セーブスロットが見つからない |
| `MOD_NOT_FOUND` | MODが見つからない |
| `PERMISSION_DENIED` | 権限不足 |
| `RESOURCE_EXHAUSTED` | リソース不足 |
| `VALIDATION_ERROR` | バリデーションエラー |
| `INTERNAL_ERROR` | 内部エラー |

---

このAPI仕様により、ECSフレームワークの状態監視、デバッグ、管理が効率的に行えます。特にWebAssembly環境でのブラウザ統合や、MODシステムの安全な管理において重要な役割を果たします。