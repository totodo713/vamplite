# ECSフレームワーク データフロー設計

## システム全体データフロー

### 高レベルアーキテクチャフロー
```mermaid
flowchart TD
    A[Game Initialize] --> B[ECS World Creation]
    B --> C[Component Registration]
    C --> D[System Registration]
    D --> E[Game Loop Start]
    
    E --> F[Frame Begin]
    F --> G[Input Processing]
    G --> H[System Updates]
    H --> I[Rendering]
    I --> J[Frame End]
    J --> K{Game Running?}
    K -->|Yes| F
    K -->|No| L[Cleanup & Exit]
    
    subgraph "ECS Core"
        M[Entity Manager]
        N[Component Store]
        O[System Manager]
        P[Query Engine]
    end
    
    H --> M
    H --> N
    H --> O
    H --> P
```

## エンティティライフサイクルフロー

### エンティティ作成・削除フロー
```mermaid
sequenceDiagram
    participant G as Game Code
    participant EM as EntityManager
    participant CS as ComponentStore
    participant Q as QueryEngine
    participant S as Systems
    
    Note over G,S: エンティティ作成フロー
    G->>EM: CreateEntity()
    EM->>EM: Generate EntityID
    EM->>EM: Add to active entities
    EM-->>G: Return EntityID
    
    G->>CS: AddComponent(entityId, component)
    CS->>CS: Store component data
    CS->>Q: Update query indices
    Q->>Q: Add to matching queries
    
    Note over G,S: エンティティ削除フロー
    G->>EM: DestroyEntity(entityId)
    EM->>CS: RemoveAllComponents(entityId)
    CS->>Q: Update query indices
    Q->>Q: Remove from all queries
    EM->>EM: Mark ID for reuse
    EM->>S: Notify entity destroyed
```

## コンポーネント管理データフロー

### コンポーネントストレージフロー
```mermaid
flowchart TD
    A[Component Request] --> B{Component Type}
    B -->|Transform| C[TransformStore]
    B -->|Sprite| D[SpriteStore]
    B -->|Health| E[HealthStore]
    B -->|Physics| F[PhysicsStore]
    
    C --> G[Dense Array Storage]
    D --> G
    E --> G
    F --> G
    
    G --> H[Sparse Set Mapping]
    H --> I[EntityID → Index]
    I --> J[Fast Component Access]
    
    subgraph "Memory Layout"
        K[Entity 0: Transform, Sprite]
        L[Entity 1: Transform, Health]
        M[Entity 2: Sprite, Physics]
        N[Entity N: ...]
    end
    
    G --> K
    G --> L
    G --> M
    G --> N
```

### コンポーネントアクセスパターン
```mermaid
sequenceDiagram
    participant S as System
    participant Q as QueryEngine
    participant CS as ComponentStore
    participant M as Memory
    
    Note over S,M: 高速クエリアクセス
    S->>Q: Query(Transform, Sprite)
    Q->>Q: Find matching entities
    Q-->>S: EntityID list [0,2,5,7,...]
    
    loop For each EntityID
        S->>CS: GetComponent<Transform>(entityId)
        CS->>M: Access dense array[index]
        M-->>CS: Component data
        CS-->>S: Transform component
        
        S->>CS: GetComponent<Sprite>(entityId)
        CS->>M: Access dense array[index]
        M-->>CS: Component data
        CS-->>S: Sprite component
        
        S->>S: Process component data
    end
```

## システム実行データフロー

### システム更新パイプライン
```mermaid
flowchart TD
    A[Frame Update Start] --> B[System Dependency Check]
    B --> C{Dependency Graph}
    
    C --> D[Phase 1: Input Systems]
    C --> E[Phase 2: Logic Systems]
    C --> F[Phase 3: Physics Systems]
    C --> G[Phase 4: Rendering Systems]
    
    subgraph "Phase 1 - Input"
        D1[Input System]
        D2[Controller System]
    end
    
    subgraph "Phase 2 - Logic"
        E1[AI System]
        E2[Gameplay System]
        E3[Animation System]
    end
    
    subgraph "Phase 3 - Physics"
        F1[Movement System]
        F2[Collision System]
        F3[Physics System]
    end
    
    subgraph "Phase 4 - Rendering"
        G1[Sprite Rendering]
        G2[UI Rendering]
        G3[Debug Rendering]
    end
    
    D --> D1
    D --> D2
    D1 --> E
    D2 --> E
    
    E --> E1
    E --> E2
    E --> E3
    E1 --> F
    E2 --> F
    E3 --> F
    
    F --> F1
    F --> F2
    F --> F3
    F1 --> G
    F2 --> G
    F3 --> G
    
    G --> G1
    G --> G2
    G --> G3
    G1 --> H[Frame Complete]
    G2 --> H
    G3 --> H
```

### システム並列実行フロー
```mermaid
flowchart TD
    A[System Manager] --> B{Dependency Analysis}
    B --> C[Independent Systems Group 1]
    B --> D[Independent Systems Group 2]
    B --> E[Independent Systems Group 3]
    
    subgraph "Parallel Execution"
        C --> C1[AI System]
        C --> C2[Audio System]
        
        D --> D1[Animation System]
        D --> D2[Particle System]
        
        E --> E1[Background System]
        E --> E2[UI Logic System]
    end
    
    C1 --> F[Synchronization Point]
    C2 --> F
    D1 --> F
    D2 --> F
    E1 --> F
    E2 --> F
    
    F --> G[Next Phase Systems]
```

## メモリ管理データフロー

### メモリプール管理フロー
```mermaid
sequenceDiagram
    participant A as Application
    participant MP as MemoryPool
    participant CS as ComponentStore
    participant GC as GarbageCollector
    
    Note over A,GC: 初期化フェーズ
    A->>MP: Initialize pools
    MP->>MP: Pre-allocate blocks
    
    Note over A,GC: 実行時フェーズ
    A->>CS: AddComponent()
    CS->>MP: RequestBlock()
    MP->>MP: Find free block
    MP-->>CS: Memory block
    CS->>CS: Store component
    
    Note over A,GC: クリーンアップフェーズ
    A->>CS: RemoveComponent()
    CS->>MP: ReturnBlock()
    MP->>MP: Mark as free
    
    Note over A,GC: ガベージコレクション
    MP->>GC: Check fragmentation
    GC->>GC: Compact memory
    GC->>CS: Update pointers
```

### メモリレイアウト最適化
```mermaid
flowchart LR
    subgraph "Inefficient AoS (Array of Structures)"
        A1[Entity1: X,Y,Sprite,Health]
        A2[Entity2: X,Y,Sprite,Health]
        A3[Entity3: X,Y,Sprite,Health]
        A4[...]
        
        A1 --> A2 --> A3 --> A4
    end
    
    subgraph "Efficient SoA (Structure of Arrays)"
        B1[X Array: X1,X2,X3,...]
        B2[Y Array: Y1,Y2,Y3,...]
        B3[Sprite Array: S1,S2,S3,...]
        B4[Health Array: H1,H2,H3,...]
        
        B1 --> B2 --> B3 --> B4
    end
    
    C[Cache Line] --> B1
    C --> B2
```

## クエリエンジンデータフロー

### 高速エンティティ検索フロー
```mermaid
flowchart TD
    A[System Query Request] --> B[Query Builder]
    B --> C{Query Type}
    
    C -->|Simple Query| D[Single Component]
    C -->|Complex Query| E[Multiple Components]
    C -->|Filtered Query| F[With/Without Filters]
    
    D --> G[Bitset Lookup]
    E --> H[Bitset Intersection]
    F --> I[Bitset Operations]
    
    G --> J[Entity List]
    H --> J
    I --> J
    
    J --> K[Cache Result]
    K --> L[Return to System]
    
    subgraph "Bitset Operations"
        M[Entity 0: 101010...]
        N[Entity 1: 110100...]
        O[Entity 2: 001110...]
        P[Result:   100000...]
    end
    
    H --> M
    H --> N
    H --> O
    H --> P
```

## MODシステム統合データフロー

### MOD ECS API フロー
```mermaid
sequenceDiagram
    participant M as MOD (Lua)
    participant B as API Bridge
    participant V as Validator
    participant E as ECS Core
    participant S as Security
    
    Note over M,S: MOD API呼び出し
    M->>B: CreateEntity()
    B->>V: Validate request
    V->>S: Check permissions
    S->>S: Verify resource limits
    S-->>V: Permission granted
    V-->>B: Validation passed
    B->>E: Forward to ECS
    E->>E: Create entity
    E-->>B: EntityID
    B-->>M: Return EntityID
    
    Note over M,S: コンポーネント追加
    M->>B: AddComponent(id, component)
    B->>V: Validate component type
    V->>S: Check allowed components
    alt Component allowed
        S-->>V: Allowed
        V-->>B: Proceed
        B->>E: Add component
        E-->>B: Success
        B-->>M: Success
    else Component forbidden
        S-->>V: Forbidden
        V-->>B: Reject
        B-->>M: Error: Permission denied
    end
```

### MODセキュリティフロー
```mermaid
flowchart TD
    A[MOD System Registration] --> B[Security Check]
    B --> C{Permission Valid?}
    
    C -->|Yes| D[Sandbox Creation]
    C -->|No| E[Reject Registration]
    
    D --> F[Resource Limit Setup]
    F --> G[API Proxy Creation]
    G --> H[System Registration]
    
    H --> I[Runtime Monitoring]
    I --> J{Resource Usage OK?}
    
    J -->|Yes| K[Continue Execution]
    J -->|No| L[Throttle/Suspend]
    
    K --> I
    L --> M[Log Security Event]
    M --> N[Notify Admin]
    
    subgraph "Security Layers"
        O[Memory Limit]
        P[CPU Time Limit]
        Q[API Access Control]
        R[Entity Count Limit]
    end
    
    F --> O
    F --> P
    F --> Q
    F --> R
```

## パフォーマンス監視データフロー

### メトリクス収集フロー
```mermaid
flowchart TD
    A[ECS Operations] --> B[Metrics Collector]
    B --> C[Performance Counters]
    
    C --> D[Entity Count]
    C --> E[System Execution Time]
    C --> F[Memory Usage]
    C --> G[Frame Time]
    
    D --> H[Metrics Aggregator]
    E --> H
    F --> H
    G --> H
    
    H --> I{Threshold Check}
    I -->|OK| J[Store Metrics]
    I -->|Warning| K[Performance Alert]
    
    J --> L[Performance Dashboard]
    K --> M[Auto-optimization]
    K --> N[Developer Notification]
    
    subgraph "Auto-optimization"
        O[Entity Pooling]
        P[System Prioritization]
        Q[Memory Compaction]
        R[Query Caching]
    end
    
    M --> O
    M --> P
    M --> Q
    M --> R
```

## エラーハンドリングフロー

### エラー処理・復旧フロー
```mermaid
sequenceDiagram
    participant S as System
    participant E as ECS Core
    participant EH as ErrorHandler
    participant L as Logger
    participant R as Recovery
    
    Note over S,R: エラー発生・処理フロー
    S->>E: System.Update()
    E->>E: Process entities
    E--xS: Error occurred
    
    S->>EH: HandleError(error)
    EH->>EH: Classify error severity
    EH->>L: Log error details
    
    alt Recoverable Error
        EH->>R: Attempt recovery
        R->>R: Restore system state
        R-->>EH: Recovery successful
        EH-->>S: Continue execution
    else Fatal Error
        EH->>R: Graceful shutdown
        R->>R: Save game state
        R->>R: Clean up resources
        R-->>EH: Shutdown complete
        EH-->>S: Terminate system
    end
```

---

このデータフロー設計により、ECSフレームワークの各コンポーネント間の相互作用が明確になり、高性能・高信頼性・高セキュリティを実現する実装指針が提供されます。