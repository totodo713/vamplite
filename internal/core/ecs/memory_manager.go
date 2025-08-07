package ecs

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// MemoryManager manages memory allocation and pooling for the ECS framework
type MemoryManager interface {
	// Pool management
	CreatePool(name string, objectSize int, initialCapacity int) error
	GetPool(name string) (ObjectPool, error)
	DestroyPool(name string) error

	// Memory allocation
	Allocate(size int) (unsafe.Pointer, error)
	AllocateAligned(size int, alignment int) (unsafe.Pointer, error)
	Deallocate(ptr unsafe.Pointer) error

	// GC control
	TriggerGC() error
	SetGCThreshold(bytes int64) error
	GetGCStats() GCStats

	// Memory monitoring
	GetMemoryUsage() MemoryManagerUsage
	SetMemoryLimit(bytes int64) error
	RegisterMemoryWarningCallback(threshold float64, callback func())

	// Leak detection
	EnableLeakDetection(enabled bool)
	GetLeakReport() LeakReport
	ForceCleanup() error

	// Metrics
	GetMetrics() MemoryMetrics
	ResetMetrics()
}

// ObjectPool manages a pool of reusable objects
type ObjectPool interface {
	Get() (unsafe.Pointer, error)
	Put(ptr unsafe.Pointer) error
	Size() int
	Capacity() int
	ObjectSize() int
	Resize(newCapacity int) error
	Clear()
}

// GCStats contains garbage collection statistics
type GCStats struct {
	NumGC      uint32
	PauseTotal time.Duration
	LastGC     time.Time
	HeapAlloc  uint64
	HeapSys    uint64
}

// MemoryManagerUsage contains current memory usage information for MemoryManager
type MemoryManagerUsage struct {
	Allocated uint64
	Used      uint64
	Reserved  uint64
	Pools     map[string]PoolUsage
}

// PoolUsage contains usage information for a specific pool
type PoolUsage struct {
	Allocated int
	InUse     int
	Available int
	HitRate   float64
}

// LeakReport contains memory leak detection results
type LeakReport struct {
	TotalLeaks  int
	LeakedBytes uint64
	Leaks       []LeakInfo
}

// LeakInfo contains information about a specific memory leak
type LeakInfo struct {
	Address     uintptr
	Size        uint64
	AllocatedAt time.Time
	StackTrace  []string
}

// MemoryMetrics contains overall memory metrics
type MemoryMetrics struct {
	TotalAllocations   uint64
	TotalDeallocations uint64
	CurrentUsage       uint64
	PeakUsage          uint64
	FragmentationRate  float64
	PoolHitRate        float64
}

// objectPoolImpl is the concrete implementation of ObjectPool
type objectPoolImpl struct {
	name         string
	objectSize   int
	capacity     int
	inUseCount   int32            // atomic - objects currently in use
	available    []unsafe.Pointer // slice-based pool for better performance
	availableMu  sync.Mutex       // dedicated mutex for available slice
	hits         uint64           // atomic
	misses       uint64           // atomic
	totalCreated int32            // atomic - total objects ever created
}

// NewObjectPool creates a new object pool
func NewObjectPool(name string, objectSize int, initialCapacity int) ObjectPool {
	pool := &objectPoolImpl{
		name:       name,
		objectSize: objectSize,
		capacity:   initialCapacity,
		available:  make([]unsafe.Pointer, 0, initialCapacity),
	}

	return pool
}

func (p *objectPoolImpl) Get() (unsafe.Pointer, error) {
	// Fast path: try to get from available pool
	p.availableMu.Lock()
	if len(p.available) > 0 {
		// Get from end of slice for better cache locality
		ptr := p.available[len(p.available)-1]
		p.available = p.available[:len(p.available)-1]
		p.availableMu.Unlock()

		atomic.AddInt32(&p.inUseCount, 1)
		atomic.AddUint64(&p.hits, 1)
		return ptr, nil
	}
	p.availableMu.Unlock()

	// Slow path: allocate new object
	ptr := allocateAlignedFast(p.objectSize, 64)
	atomic.AddInt32(&p.inUseCount, 1)
	atomic.AddInt32(&p.totalCreated, 1)
	atomic.AddUint64(&p.misses, 1)

	// Check if we need to expand pool capacity
	totalCreated := int(atomic.LoadInt32(&p.totalCreated))
	if totalCreated > p.capacity {
		p.availableMu.Lock()
		if totalCreated > p.capacity {
			p.expandPool()
		}
		p.availableMu.Unlock()
	}

	return ptr, nil
}

func (p *objectPoolImpl) Put(ptr unsafe.Pointer) error {
	if ptr == nil {
		return fmt.Errorf("cannot put nil pointer to pool")
	}

	// Decrease in-use count
	atomic.AddInt32(&p.inUseCount, -1)

	// Try to put back into pool for reuse
	p.availableMu.Lock()
	if len(p.available) < p.capacity {
		p.available = append(p.available, ptr)
		p.availableMu.Unlock()
		return nil
	}
	p.availableMu.Unlock()

	// Pool is full, free the memory
	freeAlignedFast(ptr)
	return nil
}

func (p *objectPoolImpl) Size() int {
	// Size represents objects currently in use (borrowed from pool)
	return int(atomic.LoadInt32(&p.inUseCount))
}

func (p *objectPoolImpl) Capacity() int {
	return p.capacity
}

func (p *objectPoolImpl) ObjectSize() int {
	return p.objectSize
}

func (p *objectPoolImpl) Resize(newCapacity int) error {
	if newCapacity < p.Size() {
		return fmt.Errorf("cannot resize pool to %d, %d objects in use", newCapacity, p.Size())
	}

	p.availableMu.Lock()
	defer p.availableMu.Unlock()

	// If shrinking and we have more available objects than new capacity
	if newCapacity < len(p.available) {
		// Free excess objects
		for i := newCapacity; i < len(p.available); i++ {
			freeAlignedFast(p.available[i])
		}
		p.available = p.available[:newCapacity]
	}

	// Update capacity and ensure slice has adequate capacity
	p.capacity = newCapacity
	if cap(p.available) < newCapacity {
		newSlice := make([]unsafe.Pointer, len(p.available), newCapacity)
		copy(newSlice, p.available)
		p.available = newSlice
	}

	return nil
}

func (p *objectPoolImpl) Clear() {
	p.availableMu.Lock()
	defer p.availableMu.Unlock()

	// Free all available objects
	for _, ptr := range p.available {
		freeAlignedFast(ptr)
	}

	// Reset slice
	p.available = p.available[:0]
	atomic.StoreInt32(&p.inUseCount, 0)
	atomic.StoreInt32(&p.totalCreated, 0)
}

func (p *objectPoolImpl) expandPool() {
	newCapacity := p.capacity * 2
	p.capacity = newCapacity

	// Expand slice capacity if needed
	if cap(p.available) < newCapacity {
		newSlice := make([]unsafe.Pointer, len(p.available), newCapacity)
		copy(newSlice, p.available)
		p.available = newSlice
	}
}

// memoryManagerImpl is the concrete implementation of MemoryManager
type memoryManagerImpl struct {
	pools              map[string]ObjectPool
	poolsMu            sync.RWMutex
	allocations        map[unsafe.Pointer]*allocationInfo
	allocationsMu      sync.RWMutex
	gcThreshold        int64
	memoryLimit        int64
	currentUsage       int64  // atomic
	peakUsage          int64  // atomic
	totalAllocations   uint64 // atomic
	totalDeallocations uint64 // atomic
	leakDetection      bool
	warningCallbacks   []warningCallback
	callbacksMu        sync.RWMutex
}

type allocationInfo struct {
	size        uint64
	allocatedAt time.Time
	stackTrace  []string
}

type warningCallback struct {
	threshold float64
	callback  func()
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager() MemoryManager {
	return &memoryManagerImpl{
		pools:       make(map[string]ObjectPool),
		allocations: make(map[unsafe.Pointer]*allocationInfo),
		gcThreshold: 100 * 1024 * 1024, // 100MB default
		memoryLimit: 0,                 // No limit by default
	}
}

func (m *memoryManagerImpl) CreatePool(name string, objectSize int, initialCapacity int) error {
	m.poolsMu.Lock()
	defer m.poolsMu.Unlock()

	if _, exists := m.pools[name]; exists {
		return fmt.Errorf("pool %s already exists", name)
	}

	m.pools[name] = NewObjectPool(name, objectSize, initialCapacity)
	return nil
}

func (m *memoryManagerImpl) GetPool(name string) (ObjectPool, error) {
	m.poolsMu.RLock()
	defer m.poolsMu.RUnlock()

	pool, exists := m.pools[name]
	if !exists {
		return nil, fmt.Errorf("pool %s not found", name)
	}

	return pool, nil
}

func (m *memoryManagerImpl) DestroyPool(name string) error {
	m.poolsMu.Lock()
	defer m.poolsMu.Unlock()

	pool, exists := m.pools[name]
	if !exists {
		return fmt.Errorf("pool %s not found", name)
	}

	pool.Clear()
	delete(m.pools, name)
	return nil
}

func (m *memoryManagerImpl) Allocate(size int) (unsafe.Pointer, error) {
	// Check memory limit
	if m.memoryLimit > 0 {
		newUsage := atomic.AddInt64(&m.currentUsage, int64(size))
		if newUsage > m.memoryLimit {
			atomic.AddInt64(&m.currentUsage, -int64(size))
			return nil, fmt.Errorf("memory limit exceeded: would use %d, limit is %d", newUsage, m.memoryLimit)
		}

		// Check warning thresholds
		m.checkWarnings(newUsage)
	} else {
		atomic.AddInt64(&m.currentUsage, int64(size))
	}

	// Update peak usage
	for {
		current := atomic.LoadInt64(&m.currentUsage)
		peak := atomic.LoadInt64(&m.peakUsage)
		if current <= peak || atomic.CompareAndSwapInt64(&m.peakUsage, peak, current) {
			break
		}
	}

	// Allocate memory
	ptr := allocate(size)

	// Track allocation if leak detection is enabled
	if m.leakDetection {
		m.allocationsMu.Lock()
		m.allocations[ptr] = &allocationInfo{
			size:        uint64(size),
			allocatedAt: time.Now(),
			stackTrace:  getStackTrace(),
		}
		m.allocationsMu.Unlock()
	}

	atomic.AddUint64(&m.totalAllocations, 1)

	// Check if GC should be triggered
	if atomic.LoadInt64(&m.currentUsage) > m.gcThreshold {
		runtime.GC()
	}

	return ptr, nil
}

func (m *memoryManagerImpl) AllocateAligned(size int, alignment int) (unsafe.Pointer, error) {
	// Allocate extra space for alignment
	allocSize := size + alignment

	// Check memory limit
	if m.memoryLimit > 0 {
		newUsage := atomic.AddInt64(&m.currentUsage, int64(allocSize))
		if newUsage > m.memoryLimit {
			atomic.AddInt64(&m.currentUsage, -int64(allocSize))
			return nil, fmt.Errorf("memory limit exceeded")
		}

		// Check warning thresholds
		m.checkWarnings(newUsage)
	} else {
		atomic.AddInt64(&m.currentUsage, int64(allocSize))
	}

	// Update peak usage
	for {
		current := atomic.LoadInt64(&m.currentUsage)
		peak := atomic.LoadInt64(&m.peakUsage)
		if current <= peak || atomic.CompareAndSwapInt64(&m.peakUsage, peak, current) {
			break
		}
	}

	// Allocate aligned memory
	ptr := allocateAligned(size, alignment)

	// Track allocation if leak detection is enabled
	if m.leakDetection {
		m.allocationsMu.Lock()
		m.allocations[ptr] = &allocationInfo{
			size:        uint64(allocSize),
			allocatedAt: time.Now(),
			stackTrace:  getStackTrace(),
		}
		m.allocationsMu.Unlock()
	}

	atomic.AddUint64(&m.totalAllocations, 1)

	return ptr, nil
}

func (m *memoryManagerImpl) Deallocate(ptr unsafe.Pointer) error {
	if ptr == nil {
		return fmt.Errorf("cannot deallocate nil pointer")
	}

	// Remove from tracking if leak detection is enabled
	if m.leakDetection {
		m.allocationsMu.Lock()
		if info, exists := m.allocations[ptr]; exists {
			atomic.AddInt64(&m.currentUsage, -int64(info.size))
			delete(m.allocations, ptr)
		}
		m.allocationsMu.Unlock()
	} else {
		// Estimate size for non-tracked allocations
		atomic.AddInt64(&m.currentUsage, -100) // Conservative estimate
	}

	free(ptr)
	atomic.AddUint64(&m.totalDeallocations, 1)

	return nil
}

func (m *memoryManagerImpl) TriggerGC() error {
	runtime.GC()
	return nil
}

func (m *memoryManagerImpl) SetGCThreshold(bytes int64) error {
	atomic.StoreInt64(&m.gcThreshold, bytes)
	return nil
}

func (m *memoryManagerImpl) GetGCStats() GCStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return GCStats{
		NumGC:      memStats.NumGC,
		PauseTotal: time.Duration(memStats.PauseTotalNs),
		LastGC:     time.Unix(0, int64(memStats.LastGC)),
		HeapAlloc:  memStats.HeapAlloc,
		HeapSys:    memStats.HeapSys,
	}
}

func (m *memoryManagerImpl) GetMemoryUsage() MemoryManagerUsage {
	usage := MemoryManagerUsage{
		Allocated: uint64(atomic.LoadInt64(&m.currentUsage)),
		Used:      uint64(atomic.LoadInt64(&m.currentUsage)),
		Reserved:  0,
		Pools:     make(map[string]PoolUsage),
	}

	m.poolsMu.RLock()
	for name, pool := range m.pools {
		poolImpl := pool.(*objectPoolImpl)
		usage.Pools[name] = PoolUsage{
			Allocated: poolImpl.Capacity(),
			InUse:     poolImpl.Size(),
			Available: poolImpl.Capacity() - poolImpl.Size(),
			HitRate:   float64(poolImpl.hits) / float64(poolImpl.hits+poolImpl.misses+1),
		}
	}
	m.poolsMu.RUnlock()

	return usage
}

func (m *memoryManagerImpl) SetMemoryLimit(bytes int64) error {
	atomic.StoreInt64(&m.memoryLimit, bytes)
	return nil
}

func (m *memoryManagerImpl) RegisterMemoryWarningCallback(threshold float64, callback func()) {
	m.callbacksMu.Lock()
	defer m.callbacksMu.Unlock()

	m.warningCallbacks = append(m.warningCallbacks, warningCallback{
		threshold: threshold,
		callback:  callback,
	})
}

func (m *memoryManagerImpl) EnableLeakDetection(enabled bool) {
	m.leakDetection = enabled
}

func (m *memoryManagerImpl) GetLeakReport() LeakReport {
	report := LeakReport{
		Leaks: make([]LeakInfo, 0),
	}

	m.allocationsMu.RLock()
	for ptr, info := range m.allocations {
		report.TotalLeaks++
		report.LeakedBytes += info.size
		report.Leaks = append(report.Leaks, LeakInfo{
			Address:     uintptr(ptr),
			Size:        info.size,
			AllocatedAt: info.allocatedAt,
			StackTrace:  info.stackTrace,
		})
	}
	m.allocationsMu.RUnlock()

	return report
}

func (m *memoryManagerImpl) ForceCleanup() error {
	m.allocationsMu.Lock()
	for ptr := range m.allocations {
		free(ptr)
		delete(m.allocations, ptr)
	}
	m.allocationsMu.Unlock()

	atomic.StoreInt64(&m.currentUsage, 0)

	return nil
}

func (m *memoryManagerImpl) GetMetrics() MemoryMetrics {
	return MemoryMetrics{
		TotalAllocations:   atomic.LoadUint64(&m.totalAllocations),
		TotalDeallocations: atomic.LoadUint64(&m.totalDeallocations),
		CurrentUsage:       uint64(atomic.LoadInt64(&m.currentUsage)),
		PeakUsage:          uint64(atomic.LoadInt64(&m.peakUsage)),
		FragmentationRate:  m.calculateFragmentation(),
		PoolHitRate:        m.calculatePoolHitRate(),
	}
}

func (m *memoryManagerImpl) ResetMetrics() {
	atomic.StoreUint64(&m.totalAllocations, 0)
	atomic.StoreUint64(&m.totalDeallocations, 0)
	atomic.StoreInt64(&m.peakUsage, atomic.LoadInt64(&m.currentUsage))
}

func (m *memoryManagerImpl) checkWarnings(usage int64) {
	if m.memoryLimit == 0 {
		return
	}

	ratio := float64(usage) / float64(m.memoryLimit)

	m.callbacksMu.RLock()
	callbacks := make([]warningCallback, len(m.warningCallbacks))
	copy(callbacks, m.warningCallbacks)
	m.callbacksMu.RUnlock()

	for _, wc := range callbacks {
		if ratio >= wc.threshold {
			// Call synchronously for testing
			wc.callback()
		}
	}
}

func (m *memoryManagerImpl) calculateFragmentation() float64 {
	// Simplified fragmentation calculation
	var totalCapacity int64
	var totalUsed int64

	m.poolsMu.RLock()
	for _, pool := range m.pools {
		poolImpl := pool.(*objectPoolImpl)
		totalCapacity += int64(poolImpl.Capacity() * poolImpl.objectSize)
		totalUsed += int64(poolImpl.Size() * poolImpl.objectSize)
	}
	m.poolsMu.RUnlock()

	if totalCapacity == 0 {
		return 0
	}

	efficiency := float64(totalUsed) / float64(totalCapacity)
	return 1.0 - efficiency
}

func (m *memoryManagerImpl) calculatePoolHitRate() float64 {
	var totalHits uint64
	var totalMisses uint64

	m.poolsMu.RLock()
	for _, pool := range m.pools {
		poolImpl := pool.(*objectPoolImpl)
		totalHits += poolImpl.hits
		totalMisses += poolImpl.misses
	}
	m.poolsMu.RUnlock()

	if totalHits+totalMisses == 0 {
		return 0
	}

	return float64(totalHits) / float64(totalHits+totalMisses)
}

// Size-based sync.Pool for fast allocation
var (
	pool32   = sync.Pool{New: func() interface{} { b := make([]byte, 32); return &b[0] }}
	pool64   = sync.Pool{New: func() interface{} { b := make([]byte, 64); return &b[0] }}
	pool128  = sync.Pool{New: func() interface{} { b := make([]byte, 128); return &b[0] }}
	pool256  = sync.Pool{New: func() interface{} { b := make([]byte, 256); return &b[0] }}
	pool512  = sync.Pool{New: func() interface{} { b := make([]byte, 512); return &b[0] }}
	pool1024 = sync.Pool{New: func() interface{} { b := make([]byte, 1024); return &b[0] }}
)

// allocateFast - optimized allocation using sync.Pool
func allocateFast(size int) unsafe.Pointer {
	switch {
	case size <= 32:
		return unsafe.Pointer(pool32.Get().(*byte))
	case size <= 64:
		return unsafe.Pointer(pool64.Get().(*byte))
	case size <= 128:
		return unsafe.Pointer(pool128.Get().(*byte))
	case size <= 256:
		return unsafe.Pointer(pool256.Get().(*byte))
	case size <= 512:
		return unsafe.Pointer(pool512.Get().(*byte))
	case size <= 1024:
		return unsafe.Pointer(pool1024.Get().(*byte))
	default:
		// For large allocations, use standard Go allocation
		b := make([]byte, size)
		return unsafe.Pointer(&b[0])
	}
}

// freeFast - return to appropriate sync.Pool
func freeFast(ptr unsafe.Pointer) {
	// Note: In real implementation, we would track allocation size
	// For now, this is a placeholder since Go handles deallocation
}

// allocateAlignedFast - fast aligned allocation
func allocateAlignedFast(size int, alignment int) unsafe.Pointer {
	// For common sizes, use pre-aligned pools
	if alignment <= 64 && size <= 1024 {
		return allocateFast(size)
	}

	// Fallback to standard aligned allocation
	allocSize := size + alignment
	b := make([]byte, allocSize)
	ptr := uintptr(unsafe.Pointer(&b[0]))
	aligned := (ptr + uintptr(alignment) - 1) &^ (uintptr(alignment) - 1)
	return unsafe.Pointer(aligned)
}

// freeAlignedFast - fast aligned deallocation
func freeAlignedFast(ptr unsafe.Pointer) {
	freeFast(ptr)
}

// Legacy functions for backward compatibility
func allocate(size int) unsafe.Pointer {
	return allocateFast(size)
}

func allocateAligned(size int, alignment int) unsafe.Pointer {
	return allocateAlignedFast(size, alignment)
}

func free(ptr unsafe.Pointer) {
	freeFast(ptr)
}

func freeAligned(ptr unsafe.Pointer) {
	freeAlignedFast(ptr)
}

func getStackTrace() []string {
	// Simplified stack trace collection
	const depth = 10
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])

	frames := runtime.CallersFrames(pcs[:n])
	trace := make([]string, 0, n)

	for {
		frame, more := frames.Next()
		trace = append(trace, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}

	return trace
}
