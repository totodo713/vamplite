# ECSフレームワーク メモリストレージ設計

## 概要

ECSフレームワークのメモリストレージは、高性能なゲーム実行に最適化された構造を提供します。データベースではなく、メモリ内での効率的なデータ構造と管理戦略を採用し、10,000以上のエンティティを60FPSで処理する性能を実現します。

## ストレージアーキテクチャ

### データ構造の選択

#### 1. Entity Storage (スパースセット + ジェネレーション)

```
Entity Storage Structure:
┌─────────────────────────────────────────────────────────────┐
│ EntityManager                                               │
├─────────────────────────────────────────────────────────────┤
│ SparseArray: [EntityID] → DenseIndex                      │
│ ┌─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┐    │
│ │  0  │  -  │  2  │  -  │  1  │  -  │  3  │  -  │  -  │    │
│ └─────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┘    │
│                                                             │
│ DenseArray: EntityData                                     │
│ ┌─────┬─────┬─────┬─────┐                                  │
│ │ E0  │ E4  │ E2  │ E6  │                                  │
│ └─────┴─────┴─────┴─────┘                                  │
│                                                             │
│ Generation: [uint32] - Entity recycling management        │
│ ┌─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┐    │
│ │  1  │  0  │  2  │  0  │  3  │  0  │  1  │  0  │  0  │    │
│ └─────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┘    │
└─────────────────────────────────────────────────────────────┘

Performance Characteristics:
- Entity Creation: O(1)
- Entity Deletion: O(1) 
- Entity Lookup: O(1)
- Memory Overhead: 16 bytes per entity slot
```

#### 2. Component Storage (Structure of Arrays)

```
Component Storage Structure:
┌─────────────────────────────────────────────────────────────┐
│ ComponentPool<TransformComponent>                           │
├─────────────────────────────────────────────────────────────┤
│ EntityMap: EntityID → ComponentIndex                       │
│ ┌─────────────┬─────────────┬─────────────┬─────────────┐   │
│ │   E1→0      │   E3→1      │   E7→2      │   E9→3      │   │
│ └─────────────┴─────────────┴─────────────┴─────────────┘   │
│                                                             │
│ ComponentArray: [TransformComponent]                       │
│ ┌─────────────┬─────────────┬─────────────┬─────────────┐   │
│ │ Transform1  │ Transform3  │ Transform7  │ Transform9  │   │
│ │ X:100,Y:50  │ X:200,Y:75  │ X:150,Y:30  │ X:300,Y:90  │   │
│ └─────────────┴─────────────┴─────────────┴─────────────┘   │
│                                                             │
│ FreeIndices: [ComponentIndex] - Recycling pool            │
│ ┌─────┬─────┬─────┬─────┐                                  │
│ │  4  │  8  │ 12  │ 15  │                                  │
│ └─────┴─────┴─────┴─────┘                                  │
└─────────────────────────────────────────────────────────────┘

Memory Layout Benefits:
- Sequential memory access for cache efficiency
- Minimal memory fragmentation
- Fast component iteration
- Type-specific memory pools
```

#### 3. Query Index (Bitset + Archetype)

```
Query Index Structure:
┌─────────────────────────────────────────────────────────────┐
│ QueryEngine                                                 │
├─────────────────────────────────────────────────────────────┤
│ EntityComponentMask: EntityID → ComponentMask              │
│ ┌─────────────┬─────────────┬─────────────┬─────────────┐   │
│ │ E1: 0011010 │ E3: 0010110 │ E7: 0011000 │ E9: 0001110 │   │
│ └─────────────┴─────────────┴─────────────┴─────────────┘   │
│                                                             │
│ ArchetypeGroups: ComponentMask → [EntityID]               │
│ ┌─────────────┬─────────────────────────────────────────┐   │
│ │  0011010    │ [E1, E15, E23, E44, ...]                │   │
│ │  0010110    │ [E3, E18, E31, ...]                     │   │
│ │  0011000    │ [E7, E22, E39, ...]                     │   │
│ │  0001110    │ [E9, E12, E27, ...]                     │   │
│ └─────────────┴─────────────────────────────────────────┘   │
│                                                             │
│ QueryCache: QuerySignature → [EntityID]                   │
│ ┌─────────────┬─────────────────────────────────────────┐   │
│ │"Mov+Rend"   │ [E1, E7, E15, E23, ...]  (cached)       │   │
│ │"Phy+Col"    │ [E3, E9, E18, E31, ...]  (cached)       │   │
│ └─────────────┴─────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘

Query Performance:
- Component mask check: O(1) bitwise operation
- Archetype lookup: O(1) hash table access
- Cached query: O(1) direct array access
- Cache hit rate target: >95%
```

## メモリ管理戦略

### 1. Memory Pool Management

```go
// Memory Pool Architecture
type MemoryPool struct {
    // Fixed-size block pools for different component sizes
    SmallBlocks   *BlockPool // 1-64 bytes
    MediumBlocks  *BlockPool // 65-512 bytes  
    LargeBlocks   *BlockPool // 513-4096 bytes
    HugeBlocks    *BlockPool // >4096 bytes
    
    // Allocation tracking
    AllocatedBytes   int64
    AvailableBytes   int64
    FragmentedBytes  int64
    
    // Performance metrics
    AllocationCount  int64
    DeallocationCount int64
    PoolHitRate      float64
}

type BlockPool struct {
    BlockSize      int
    BlocksPerChunk int
    FreeBlocks     []unsafe.Pointer
    UsedBlocks     []unsafe.Pointer
    ChunkList      []*MemoryChunk
}

type MemoryChunk struct {
    Memory     []byte
    BlockSize  int
    FreeCount  int
    UsedCount  int
    FreeList   []int
}
```

### 2. Garbage Collection Optimization

```
GC Optimization Strategy:
┌─────────────────────────────────────────────────────────────┐
│ Memory Allocation Patterns                                  │
├─────────────────────────────────────────────────────────────┤
│ 1. Object Pool Pattern                                     │
│    - Pre-allocate component pools                          │
│    - Reuse objects instead of new allocations              │
│    - Minimal GC pressure                                   │
│                                                             │
│ 2. Zero-Allocation Operations                              │
│    - Slice reuse for query results                         │
│    - Temporary buffer pools                                │
│    - In-place data modifications                           │
│                                                             │
│ 3. Batch Allocation                                        │
│    - Allocate large chunks, subdivide internally           │
│    - Reduce allocation frequency                           │
│    - Improve memory locality                               │
│                                                             │
│ 4. Lazy Cleanup                                            │
│    - Defer memory release to low-activity frames           │
│    - Batch deallocation operations                         │
│    - Background cleanup goroutines                         │
└─────────────────────────────────────────────────────────────┘

Target GC Performance:
- GC pause time: <1ms
- GC frequency: <10 times per second
- Memory growth rate: <1MB per minute during steady state
```

### 3. Cache-Friendly Data Layout

```
CPU Cache Optimization:
┌─────────────────────────────────────────────────────────────┐
│ Cache Line Alignment (64-byte boundaries)                  │
├─────────────────────────────────────────────────────────────┤
│ TransformComponent Array:                                   │
│ ┌──────────────────────────────────────────────────────────┐ │
│ │ Cache Line 1: T1  T2  T3  T4                           │ │
│ │ Cache Line 2: T5  T6  T7  T8                           │ │
│ │ Cache Line 3: T9  T10 T11 T12                          │ │
│ └──────────────────────────────────────────────────────────┘ │
│                                                             │
│ VelocityComponent Array:                                    │
│ ┌──────────────────────────────────────────────────────────┐ │
│ │ Cache Line 1: V1  V2  V3  V4                           │ │
│ │ Cache Line 2: V5  V6  V7  V8                           │ │
│ │ Cache Line 3: V9  V10 V11 V12                          │ │
│ └──────────────────────────────────────────────────────────┘ │
│                                                             │
│ System Processing Pattern:                                  │
│ MovementSystem.Update():                                    │
│   - Sequential read: Transform[0..n]                       │
│   - Sequential read: Velocity[0..n]                        │
│   - Sequential write: Transform[0..n]                      │
│   - Cache miss rate: <5%                                   │
└─────────────────────────────────────────────────────────────┘
```

## パフォーマンス最適化仕様

### 1. Memory Access Patterns

| 操作 | データ構造 | 複雑度 | キャッシュ効率 | 最適化手法 |
|------|------------|--------|----------------|------------|
| Entity作成 | SparseSet | O(1) | 高 | Index recycling |
| Component追加 | ComponentPool | O(1) | 高 | Array append |
| Component検索 | HashMap | O(1) | 中 | Hash pre-computation |
| Entity反復 | DenseArray | O(n) | 最高 | Sequential access |
| Query実行 | ArchetypeGroups | O(1) | 高 | Pre-computed groups |

### 2. Memory Footprint Targets

```
Memory Usage Breakdown (per 10,000 entities):
┌─────────────────────────────────────────────────────────────┐
│ Component Type          │ Size/Entity │ Total Memory      │   │
├─────────────────────────┼─────────────┼───────────────────┤   │
│ EntityID + Metadata     │ 16 bytes    │ 160 KB           │   │
│ Transform Component     │ 32 bytes    │ 320 KB           │   │
│ Sprite Component        │ 48 bytes    │ 480 KB           │   │
│ Velocity Component      │ 24 bytes    │ 240 KB           │   │
│ Health Component        │ 16 bytes    │ 160 KB           │   │
│ Collision Component     │ 40 bytes    │ 400 KB           │   │
│ Index Structures        │ 8 bytes     │ 80 KB            │   │
│ Query Cache            │ 4 bytes     │ 40 KB            │   │
├─────────────────────────┼─────────────┼───────────────────┤   │
│ Total Base Memory       │ 188 bytes   │ 1.88 MB          │   │
│ Additional Overhead     │ 12 bytes    │ 120 KB           │   │
├─────────────────────────┼─────────────┼───────────────────┤   │
│ Total Memory Usage      │ 200 bytes   │ 2.0 MB           │   │
└─────────────────────────────────────────────────────────────┘

Target: <100 bytes per entity overhead = EXCEEDED
Optimization needed: Bit packing, struct alignment
```

### 3. Concurrent Access Management

```go
// Thread-Safe Component Access
type ComponentStore struct {
    // Reader-Writer locks for component types
    ComponentLocks map[ComponentType]*sync.RWMutex
    
    // Lock-free read paths for common queries
    AtomicReaders map[ComponentType]*AtomicComponentReader
    
    // Write batching for performance
    WriteBatch    *ComponentWriteBatch
    BatchMutex    sync.Mutex
}

// Lock-free query execution
type LockFreeQueryEngine struct {
    // Immutable query cache
    QueryCache atomic.Value // map[string][]EntityID
    
    // Copy-on-write archetype updates
    ArchetypeMap atomic.Value // map[ComponentMask][]EntityID
    
    // RCU (Read-Copy-Update) for hot path queries
    CurrentGeneration uint64
    OldGenerations    []QuerySnapshot
}

// Memory ordering and synchronization
type MemoryBarrier struct {
    // Sequential consistency for component updates
    UpdateSeqNo uint64
    
    // Memory fences for cross-thread visibility
    WriteFence  sync.Mutex
    ReadFence   sync.RWMutex
}
```

## ストレージ実装仕様

### 1. Entity Manager Implementation

```go
type EntityManager struct {
    // Core storage
    entities    []EntityData
    freeIndices []uint32
    generation  []uint32
    
    // Sparse mapping
    sparseTodense map[uint32]uint32
    denseToSparse []uint32
    
    // Performance counters
    createCount uint64
    destroyCount uint64
    activeCount uint32
    
    // Memory management
    entityPool *sync.Pool
    mutex      sync.RWMutex
}

type EntityData struct {
    ID         EntityID
    Generation uint32
    Active     bool
    ComponentMask ComponentMask
}
```

### 2. Component Store Implementation

```go
type ComponentStore struct {
    // Type-specific storage
    componentPools map[ComponentType]*ComponentPool
    
    // Entity to component mapping
    entityComponents map[EntityID]map[ComponentType]ComponentIndex
    
    // Memory pools
    memoryManager *MemoryManager
    
    // Concurrent access
    locks map[ComponentType]*sync.RWMutex
}

type ComponentPool struct {
    ComponentType ComponentType
    Components    []Component
    EntityMap     map[EntityID]ComponentIndex
    FreeIndices   []ComponentIndex
    
    // Memory layout optimization
    ElementSize   int
    AlignedSize   int
    CacheLineSize int
}
```

### 3. Query Engine Implementation

```go
type QueryEngine struct {
    // Query cache
    cachedQueries map[string]*CachedQuery
    cacheHits     uint64
    cacheMisses   uint64
    
    // Archetype management
    archetypes map[ComponentMask]*ArchetypeGroup
    
    // Hot path optimization
    commonQueries []*PrecomputedQuery
    
    // Background maintenance
    cleanupTicker *time.Ticker
    rebuildChan   chan ComponentMask
}

type ArchetypeGroup struct {
    Mask     ComponentMask
    Entities []EntityID
    Capacity int
    
    // Memory management
    GrowthFactor float64
    MaxCapacity  int
}

type CachedQuery struct {
    Signature string
    Results   []EntityID
    LastUsed  time.Time
    HitCount  uint64
    
    // Invalidation tracking
    DependentTypes []ComponentType
    Generation     uint64
}
```

## メモリ監視とデバッグ

### 1. Memory Profiling Integration

```go
type MemoryProfiler struct {
    // Allocation tracking
    AllocationMap map[uintptr]*AllocationInfo
    
    // Performance metrics
    TotalAllocated   int64
    TotalDeallocated int64
    PeakMemoryUsage  int64
    
    // Leak detection
    LeakThreshold    time.Duration
    SuspiciousAllocs []*AllocationInfo
    
    // Profiling hooks
    OnAllocation   func(*AllocationInfo)
    OnDeallocation func(*AllocationInfo)
}

type AllocationInfo struct {
    Pointer     uintptr
    Size        int64
    AllocTime   time.Time
    StackTrace  []uintptr
    ComponentType ComponentType
    EntityID    EntityID
}
```

### 2. Performance Monitoring

```go
type PerformanceMonitor struct {
    // Frame timing
    FrameTime       MovingAverage
    SystemTimes     map[SystemType]MovingAverage
    
    // Memory metrics
    MemoryUsage     MovingAverage
    GCPressure      MovingAverage
    
    // Cache performance
    CacheHitRate    MovingAverage
    CacheMissRate   MovingAverage
    
    // Entity statistics
    EntityCount     MovingAverage
    ComponentCount  map[ComponentType]MovingAverage
}

type MovingAverage struct {
    WindowSize int
    Values     []float64
    Index      int
    Sum        float64
    Count      int
}
```

## データ永続化戦略

### 1. Save/Load System

```go
type WorldSerializer struct {
    // Serialization format
    Version    uint32
    Compression bool
    Encryption  bool
    
    // Component serializers
    ComponentSerializers map[ComponentType]ComponentSerializer
    
    // Incremental saves
    LastSaveChecksum [32]byte
    ChangedEntities  []EntityID
    
    // Background saving
    SaveQueue    chan SaveRequest
    SaveWorker   *sync.WaitGroup
}

type SaveData struct {
    Header      SaveHeader
    Entities    []SerializedEntity
    Systems     []SerializedSystem
    Metadata    map[string]interface{}
    Checksum    [32]byte
}

type SerializedEntity struct {
    EntityID   EntityID
    Components []SerializedComponent
    Active     bool
    Generation uint32
}
```

このメモリストレージ設計により、ECSフレームワークは目標性能を達成し、スケーラブルなゲーム開発を支援します。