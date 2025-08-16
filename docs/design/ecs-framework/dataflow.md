# ECSフレームワーク データフロー図

## システム全体のデータフロー

### 1. ECS アーキテクチャ概要

```mermaid
flowchart TD
    A[Game Loop] --> B[System Manager]
    B --> C[Update Systems]
    B --> D[Render Systems]
    
    C --> E[Entity Manager]
    C --> F[Component Store]
    
    E --> G[Entity Creation/Deletion]
    E --> H[Entity Queries]
    
    F --> I[Component Add/Remove]
    F --> J[Component Data Access]
    
    H --> K[Query Engine]
    K --> L[Bit Mask Filtering]
    L --> M[Entity Collections]
    
    style A fill:#e1f5fe
    style B fill:#f3e5f5
    style K fill:#fff3e0
```

### 2. エンティティライフサイクル

```mermaid
sequenceDiagram
    participant G as Game Loop
    participant SM as System Manager
    participant EM as Entity Manager
    participant CS as Component Store
    participant QE as Query Engine
    
    G->>SM: Update Frame
    SM->>EM: Create Entity
    EM->>EM: Generate EntityID
    EM-->>SM: Return EntityID
    
    SM->>CS: Add Components
    CS->>CS: Store Component Data
    CS->>QE: Update Entity Queries
    QE-->>CS: Query Index Updated
    
    SM->>SM: Execute Systems
    SM->>QE: Query Entities
    QE-->>SM: Entity Collections
    
    SM->>CS: Access/Modify Components
    CS-->>SM: Component Data
    
    Note over G,QE: Frame Complete
```

### 3. システム実行フロー

```mermaid
flowchart TD
    Start([Frame Start]) --> A[System Manager]
    A --> B{Dependency Check}
    B -->|Independent| C[Parallel Execution]
    B -->|Dependent| D[Sequential Execution]
    
    C --> E[Movement System]
    C --> F[AI System]
    C --> G[Physics System]
    
    D --> H[Input System]
    H --> I[Game Logic System]
    I --> J[Audio System]
    
    E --> K[Update Components]
    F --> K
    G --> K
    H --> K
    I --> K
    J --> K
    
    K --> L{All Systems Complete?}
    L -->|No| M[Wait for Completion]
    M --> L
    L -->|Yes| N[Render Systems]
    
    N --> O[Sprite Render]
    N --> P[UI Render]
    N --> Q[Effect Render]
    
    O --> End([Frame End])
    P --> End
    Q --> End
    
    style Start fill:#c8e6c9
    style End fill:#ffcdd2
    style C fill:#fff3e0
    style D fill:#f3e5f5
```

### 4. コンポーネントクエリ処理

```mermaid
flowchart LR
    A[System Query Request] --> B[Query Engine]
    B --> C{Cache Hit?}
    C -->|Yes| D[Return Cached Results]
    C -->|No| E[Build Bit Mask]
    
    E --> F[Entity Iteration]
    F --> G{Component Match?}
    G -->|Yes| H[Add to Results]
    G -->|No| I[Skip Entity]
    
    H --> J[Continue Iteration]
    I --> J
    J --> K{More Entities?}
    K -->|Yes| F
    K -->|No| L[Cache Results]
    
    L --> M[Return Entity Collection]
    D --> M
    
    style A fill:#e1f5fe
    style B fill:#fff3e0
    style M fill:#c8e6c9
```

### 5. メモリ管理データフロー

```mermaid
flowchart TD
    A[Component Request] --> B[Memory Manager]
    B --> C{Pool Available?}
    C -->|Yes| D[Allocate from Pool]
    C -->|No| E[Create New Pool]
    
    E --> F[Allocate Memory Block]
    F --> D
    
    D --> G[Return Component Slot]
    
    H[Component Deletion] --> I[Memory Manager]
    I --> J[Mark Slot as Free]
    J --> K[Return to Pool]
    
    K --> L{Pool Threshold?}
    L -->|Above| M[Keep in Pool]
    L -->|Below| N[Release to GC]
    
    style A fill:#e1f5fe
    style H fill:#ffcdd2
    style B fill:#fff3e0
    style I fill:#fff3e0
```

## 詳細データフロー仕様

### エンティティ作成フロー

```mermaid
sequenceDiagram
    participant C as Client Code
    participant EM as Entity Manager
    participant ID as ID Generator
    participant CS as Component Store
    participant QE as Query Engine
    participant MP as Memory Pool
    
    C->>EM: CreateEntity()
    EM->>ID: GenerateID()
    ID->>ID: Increment Counter + Generation
    ID-->>EM: EntityID{Index, Generation}
    
    EM->>EM: Mark Entity as Active
    EM-->>C: Return EntityID
    
    C->>CS: AddComponent(EntityID, Component)
    CS->>MP: AllocateComponent()
    MP-->>CS: ComponentSlot
    CS->>CS: Store Component Data
    
    CS->>QE: UpdateEntityMask(EntityID)
    QE->>QE: Set Component Bits
    QE-->>CS: Index Updated
    
    Note over C,MP: Entity fully initialized
```

### システム並列実行フロー

```mermaid
flowchart TD
    A[Frame Start] --> B[System Manager]
    B --> C[Dependency Analysis]
    C --> D[Create Execution Groups]
    
    D --> E[Group 1: Independent Systems]
    D --> F[Group 2: Physics Dependencies]
    D --> G[Group 3: Rendering Dependencies]
    
    E --> H[Goroutine Pool]
    H --> I[Movement System]
    H --> J[AI System]
    H --> K[Input System]
    
    I --> L[Sync Point 1]
    J --> L
    K --> L
    
    L --> F
    F --> M[Physics System]
    F --> N[Collision System]
    
    M --> O[Sync Point 2]
    N --> O
    
    O --> G
    G --> P[Render System]
    G --> Q[UI System]
    
    P --> R[Frame Complete]
    Q --> R
    
    style E fill:#c8e6c9
    style F fill:#fff3e0
    style G fill:#f3e5f5
```

### 高性能クエリシステム

```mermaid
flowchart LR
    A[System Query] --> B[Component Mask]
    B --> C[Mask: 0011010]
    
    C --> D[Entity Iteration]
    D --> E[Entity 1: 0011010]
    D --> F[Entity 2: 0010010]
    D --> G[Entity 3: 0011110]
    
    E --> H{Mask & Entity == Mask?}
    F --> I{Mask & Entity == Mask?}
    G --> J{Mask & Entity == Mask?}
    
    H -->|Yes: 0011010| K[Include Entity 1]
    I -->|No: 0010010| L[Skip Entity 2]
    J -->|Yes: 0011110| M[Include Entity 3]
    
    K --> N[Result Collection]
    M --> N
    
    N --> O[Return [Entity1, Entity3]]
    
    style C fill:#fff3e0
    style H fill:#c8e6c9
    style I fill:#ffcdd2
    style J fill:#c8e6c9
```

## パフォーマンス最適化フロー

### キャッシュフレンドリーなデータアクセス

```mermaid
flowchart TD
    A[System Update Request] --> B[Component Query]
    B --> C[Sequential Memory Access]
    
    C --> D[Transform Array]
    C --> E[Velocity Array]
    C --> F[Position Array]
    
    D --> G[Cache Line 1: T1,T2,T3,T4]
    E --> H[Cache Line 2: V1,V2,V3,V4]
    F --> I[Cache Line 3: P1,P2,P3,P4]
    
    G --> J[CPU Cache Hit]
    H --> J
    I --> J
    
    J --> K[High Performance Processing]
    
    style C fill:#c8e6c9
    style J fill:#fff3e0
    style K fill:#e1f5fe
```

### メモリプール効率化

```mermaid
sequenceDiagram
    participant S as System
    participant MP as Memory Pool
    participant GC as Garbage Collector
    
    Note over S,GC: Component Creation Phase
    S->>MP: Request Component Slot
    MP->>MP: Check Free Slots
    
    alt Pool has free slots
        MP-->>S: Return recycled slot
    else Pool empty
        MP->>MP: Allocate new block
        MP-->>S: Return new slot
    end
    
    Note over S,GC: Component Deletion Phase
    S->>MP: Release Component
    MP->>MP: Mark slot as free
    MP->>MP: Add to free list
    
    Note over S,GC: GC Optimization
    MP->>MP: Batch allocations
    MP->>GC: Minimal allocation calls
    GC-->>MP: Reduced GC pressure
```

## エラーハンドリングフロー

### 防御的プログラミング

```mermaid
flowchart TD
    A[API Call] --> B{Input Validation}
    B -->|Valid| C[Execute Operation]
    B -->|Invalid| D[Return Error]
    
    C --> E{Operation Success?}
    E -->|Success| F[Return Result]
    E -->|Error| G[Log Error Details]
    
    G --> H[Attempt Recovery]
    H --> I{Recovery Success?}
    I -->|Success| J[Continue Operation]
    I -->|Failed| K[Graceful Degradation]
    
    D --> L[Error Response]
    K --> L
    F --> M[Success Response]
    J --> M
    
    style B fill:#fff3e0
    style G fill:#ffcdd2
    style H fill:#f3e5f5
    style F fill:#c8e6c9
```

## MODセキュリティデータフロー

```mermaid
sequenceDiagram
    participant M as MOD System
    participant API as ECS API Gateway
    participant V as Validator
    participant ECS as ECS Core
    
    M->>API: Component Request
    API->>V: Validate Permissions
    V->>V: Check Access Rights
    
    alt Permission granted
        V-->>API: Validation Success
        API->>ECS: Forward Request
        ECS-->>API: Return Data
        API-->>M: Filtered Response
    else Permission denied
        V-->>API: Validation Failed
        API-->>M: Access Denied Error
    end
    
    Note over M,ECS: All MOD access is mediated
```

このデータフロー設計により、ECSフレームワークは高性能・安全性・拡張性を同時に実現します。