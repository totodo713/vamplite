package ecs

import (
	"runtime"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_ObjectPool_Creation tests pool creation and basic properties
func Test_ObjectPool_Creation(t *testing.T) {
	// Given: pool parameters
	poolName := "TestPool"
	objectSize := 64
	initialCapacity := 100

	// When: create pool
	pool := NewObjectPool(poolName, objectSize, initialCapacity)

	// Then: properties are correctly set
	assert.NotNil(t, pool)
	assert.Equal(t, objectSize, pool.ObjectSize())
	assert.Equal(t, initialCapacity, pool.Capacity())
	assert.Equal(t, 0, pool.Size()) // initially no objects in use
}

// Test_ObjectPool_GetPut tests object acquisition and return
func Test_ObjectPool_GetPut(t *testing.T) {
	// Given: initialized pool
	pool := NewObjectPool("TestPool", 64, 10)

	// When: get an object
	ptr1, err := pool.Get()
	assert.NoError(t, err)
	assert.NotNil(t, ptr1)
	assert.Equal(t, 1, pool.Size())

	// When: get another object
	ptr2, err := pool.Get()
	assert.NoError(t, err)
	assert.NotNil(t, ptr2)
	assert.NotEqual(t, ptr1, ptr2) // different addresses
	assert.Equal(t, 2, pool.Size())

	// When: return an object
	err = pool.Put(ptr1)
	assert.NoError(t, err)
	assert.Equal(t, 1, pool.Size())

	// When: get object again
	ptr3, err := pool.Get()
	assert.NoError(t, err)
	assert.Equal(t, ptr1, ptr3) // should be reused
}

// Test_ObjectPool_Overflow tests automatic expansion when capacity is exceeded
func Test_ObjectPool_Overflow(t *testing.T) {
	// Given: small capacity pool
	pool := NewObjectPool("TestPool", 32, 2)

	// When: get objects exceeding capacity
	ptrs := make([]unsafe.Pointer, 5)
	for i := 0; i < 5; i++ {
		ptr, err := pool.Get()
		assert.NoError(t, err)
		assert.NotNil(t, ptr)
		ptrs[i] = ptr
	}

	// Then: pool auto-expands
	assert.True(t, pool.Capacity() >= 5)
	assert.Equal(t, 5, pool.Size())
}

// Test_ObjectPool_Concurrent tests thread-safe operations
func Test_ObjectPool_Concurrent(t *testing.T) {
	pool := NewObjectPool("ConcurrentPool", 64, 100)

	var wg sync.WaitGroup
	iterations := 1000
	goroutines := 10

	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				ptr, err := pool.Get()
				if err == nil {
					time.Sleep(time.Microsecond)
					pool.Put(ptr)
				}
			}
		}()
	}

	wg.Wait()

	// Pool should be consistent after concurrent operations
	assert.True(t, pool.Size() >= 0)
	assert.True(t, pool.Size() <= pool.Capacity())
}

// Test_MemoryManager_CreatePool tests pool creation via memory manager
func Test_MemoryManager_CreatePool(t *testing.T) {
	// Given: memory manager
	mm := NewMemoryManager()

	// When: create pool
	err := mm.CreatePool("EntityPool", 128, 1000)
	assert.NoError(t, err)

	// Then: pool can be retrieved
	pool, err := mm.GetPool("EntityPool")
	assert.NoError(t, err)
	assert.NotNil(t, pool)

	// When: create pool with same name
	err = mm.CreatePool("EntityPool", 64, 500)

	// Then: error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// Test_MemoryManager_DestroyPool tests pool destruction
func Test_MemoryManager_DestroyPool(t *testing.T) {
	mm := NewMemoryManager()

	// Create and destroy pool
	err := mm.CreatePool("TempPool", 64, 100)
	require.NoError(t, err)

	err = mm.DestroyPool("TempPool")
	assert.NoError(t, err)

	// Pool should no longer exist
	_, err = mm.GetPool("TempPool")
	assert.Error(t, err)
}

// Test_MemoryManager_Allocate tests direct memory allocation
func Test_MemoryManager_Allocate(t *testing.T) {
	// Given: memory manager
	mm := NewMemoryManager()

	// When: allocate memory
	ptr, err := mm.Allocate(256)
	assert.NoError(t, err)
	assert.NotNil(t, ptr)

	// Then: memory usage increases
	usage := mm.GetMemoryUsage()
	assert.True(t, usage.Allocated >= 256)

	// When: deallocate memory
	err = mm.Deallocate(ptr)
	assert.NoError(t, err)
}

// Test_MemoryManager_AllocateAligned tests aligned memory allocation
func Test_MemoryManager_AllocateAligned(t *testing.T) {
	// Given: memory manager
	mm := NewMemoryManager()

	// When: allocate with 64-byte alignment
	ptr, err := mm.AllocateAligned(100, 64)
	assert.NoError(t, err)
	assert.NotNil(t, ptr)

	// Then: address is multiple of 64
	address := uintptr(ptr)
	assert.Equal(t, uintptr(0), address%64)

	// Cleanup
	err = mm.Deallocate(ptr)
	assert.NoError(t, err)
}

// Test_MemoryManager_GCControl tests GC control functionality
func Test_MemoryManager_GCControl(t *testing.T) {
	// Given: memory manager
	mm := NewMemoryManager()

	// When: set GC threshold
	err := mm.SetGCThreshold(10 * 1024 * 1024) // 10MB
	assert.NoError(t, err)

	// When: allocate large amount of memory
	ptrs := make([]unsafe.Pointer, 100)
	for i := 0; i < 100; i++ {
		ptr, _ := mm.Allocate(1024 * 100) // 100KB each
		ptrs[i] = ptr
	}

	// When: manually trigger GC
	initialStats := mm.GetGCStats()
	err = mm.TriggerGC()
	assert.NoError(t, err)

	// Then: GC stats are updated
	newStats := mm.GetGCStats()
	assert.True(t, newStats.NumGC > initialStats.NumGC)

	// Cleanup
	for _, ptr := range ptrs {
		if ptr != nil {
			mm.Deallocate(ptr)
		}
	}
}

// Test_MemoryManager_UsageTracking tests memory usage tracking
func Test_MemoryManager_UsageTracking(t *testing.T) {
	// Given: memory manager
	mm := NewMemoryManager()

	// When: check initial state
	usage := mm.GetMemoryUsage()
	initialAllocated := usage.Allocated

	// When: allocate memory
	ptrs := make([]unsafe.Pointer, 10)
	for i := 0; i < 10; i++ {
		ptr, _ := mm.Allocate(1024)
		ptrs[i] = ptr
	}

	// Then: usage increases
	usage = mm.GetMemoryUsage()
	assert.True(t, usage.Allocated > initialAllocated)
	assert.True(t, usage.Used >= 10*1024)

	// When: deallocate memory
	for _, ptr := range ptrs {
		if ptr != nil {
			mm.Deallocate(ptr)
		}
	}

	// Then: usage decreases
	usage = mm.GetMemoryUsage()
	assert.True(t, usage.Used < 10*1024)
}

// Test_MemoryManager_MemoryLimit tests memory limit enforcement
func Test_MemoryManager_MemoryLimit(t *testing.T) {
	// Given: manager with memory limit
	mm := NewMemoryManager()
	err := mm.SetMemoryLimit(1024 * 1024) // 1MB
	assert.NoError(t, err)

	// When: allocate within limit
	ptr1, err := mm.Allocate(512 * 1024) // 512KB
	assert.NoError(t, err)
	assert.NotNil(t, ptr1)

	ptr2, err := mm.Allocate(256 * 1024) // 256KB
	assert.NoError(t, err)
	assert.NotNil(t, ptr2)

	// When: allocate exceeding limit
	ptr3, err := mm.Allocate(512 * 1024) // 512KB (total would be 1.25MB)

	// Then: error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "memory limit exceeded")
	assert.Nil(t, ptr3)

	// Cleanup
	mm.Deallocate(ptr1)
	mm.Deallocate(ptr2)
}

// Test_MemoryManager_WarningCallback tests memory warning callbacks
func Test_MemoryManager_WarningCallback(t *testing.T) {
	// Given: memory manager with warning callback
	mm := NewMemoryManager()
	mm.SetMemoryLimit(1024 * 1024) // 1MB

	warningTriggered := false
	mm.RegisterMemoryWarningCallback(0.8, func() {
		warningTriggered = true
	})

	// When: use more than 80% of memory
	ptr, err := mm.Allocate(850 * 1024) // 850KB (> 80% of 1MB)
	assert.NoError(t, err)
	assert.NotNil(t, ptr)

	// Then: warning callback is triggered
	assert.True(t, warningTriggered)

	// Cleanup
	mm.Deallocate(ptr)
}

// Test_MemoryManager_LeakDetection tests memory leak detection
func Test_MemoryManager_LeakDetection(t *testing.T) {
	// Given: manager with leak detection enabled
	mm := NewMemoryManager()
	mm.EnableLeakDetection(true)

	// When: allocate memory and don't free some
	ptr1, _ := mm.Allocate(100)
	ptr2, _ := mm.Allocate(200)
	ptr3, _ := mm.Allocate(300)

	mm.Deallocate(ptr2) // don't free ptr1 and ptr3

	// When: get leak report
	report := mm.GetLeakReport()

	// Then: leaks are detected
	assert.Equal(t, 2, report.TotalLeaks)
	assert.Equal(t, uint64(400), report.LeakedBytes) // 100 + 300
	assert.Len(t, report.Leaks, 2)

	// Cleanup
	mm.Deallocate(ptr1)
	mm.Deallocate(ptr3)
}

// Test_MemoryManager_Metrics tests metrics collection
func Test_MemoryManager_Metrics(t *testing.T) {
	mm := NewMemoryManager()

	// Initial metrics
	metrics := mm.GetMetrics()
	assert.Equal(t, uint64(0), metrics.TotalAllocations)
	assert.Equal(t, uint64(0), metrics.TotalDeallocations)

	// Perform operations
	ptr1, _ := mm.Allocate(100)
	ptr2, _ := mm.Allocate(200)
	mm.Deallocate(ptr1)

	// Check updated metrics
	metrics = mm.GetMetrics()
	assert.Equal(t, uint64(2), metrics.TotalAllocations)
	assert.Equal(t, uint64(1), metrics.TotalDeallocations)
	assert.True(t, metrics.CurrentUsage > 0)
	assert.True(t, metrics.PeakUsage >= metrics.CurrentUsage)

	// Cleanup
	mm.Deallocate(ptr2)
}

// Test_MemoryManager_ForceCleanup tests forced cleanup
func Test_MemoryManager_ForceCleanup(t *testing.T) {
	mm := NewMemoryManager()
	mm.EnableLeakDetection(true)

	// Create some allocations
	ptr1, _ := mm.Allocate(100)
	_, _ = mm.Allocate(200) // ptr2 will be leaked for testing
	mm.Deallocate(ptr1)     // Only deallocate ptr1

	// Force cleanup
	err := mm.ForceCleanup()
	assert.NoError(t, err)

	// Check no leaks remain
	report := mm.GetLeakReport()
	assert.Equal(t, 0, report.TotalLeaks)
}

// Benchmark_ObjectPool_GetPut benchmarks pool operations
func Benchmark_ObjectPool_GetPut(b *testing.B) {
	pool := NewObjectPool("BenchPool", 64, 1000)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ptr, _ := pool.Get()
			pool.Put(ptr)
		}
	})
}

// Benchmark_MemoryManager_Allocate benchmarks direct allocation
func Benchmark_MemoryManager_Allocate(b *testing.B) {
	mm := NewMemoryManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptr, _ := mm.Allocate(128)
		mm.Deallocate(ptr)
	}
}

// Benchmark_MemoryManager_WithPool benchmarks pool-based allocation
func Benchmark_MemoryManager_WithPool(b *testing.B) {
	mm := NewMemoryManager()
	mm.CreatePool("BenchPool", 128, 10000)
	pool, _ := mm.GetPool("BenchPool")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ptr, _ := pool.Get()
			pool.Put(ptr)
		}
	})
}

// Benchmark_MemoryManager_AllocateAligned benchmarks aligned allocation
func Benchmark_MemoryManager_AllocateAligned(b *testing.B) {
	mm := NewMemoryManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptr, _ := mm.AllocateAligned(128, 64)
		mm.Deallocate(ptr)
	}
}

// Test_MemoryManager_StressTest performs stress testing
func Test_MemoryManager_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	mm := NewMemoryManager()
	mm.EnableLeakDetection(true)
	mm.SetMemoryLimit(100 * 1024 * 1024) // 100MB limit

	// Run for 10 seconds
	duration := 10 * time.Second
	start := time.Now()

	var allocations []unsafe.Pointer
	allocCount := 0
	deallocCount := 0

	for time.Since(start) < duration {
		// Random allocation size
		size := (allocCount%10 + 1) * 100

		// Randomly allocate or deallocate
		if allocCount == deallocCount || (allocCount-deallocCount < 1000 && allocCount%2 == 0) {
			// Allocate
			ptr, err := mm.Allocate(size)
			if err == nil {
				allocations = append(allocations, ptr)
				allocCount++
			}
		} else if len(allocations) > 0 {
			// Deallocate
			idx := deallocCount % len(allocations)
			if allocations[idx] != nil {
				mm.Deallocate(allocations[idx])
				allocations[idx] = nil
				deallocCount++
			}
		}

		// Check metrics periodically
		if allocCount%1000 == 0 {
			metrics := mm.GetMetrics()
			assert.True(t, metrics.FragmentationRate < 0.1) // < 10% fragmentation

			// Force GC occasionally
			if allocCount%5000 == 0 {
				runtime.GC()
			}
		}
	}

	// Cleanup remaining allocations
	for _, ptr := range allocations {
		if ptr != nil {
			mm.Deallocate(ptr)
		}
	}

	// Final checks
	report := mm.GetLeakReport()
	assert.Equal(t, 0, report.TotalLeaks)

	metrics := mm.GetMetrics()
	assert.Equal(t, uint64(allocCount), metrics.TotalAllocations)
	assert.Equal(t, uint64(allocCount), metrics.TotalDeallocations)
}
