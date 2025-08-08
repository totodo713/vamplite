package ecs

import (
	"sync"
	"testing"
	"time"
)

// TestRecordCounter tests counter metric recording
func TestRecordCounter(t *testing.T) {
	collector := NewMetricsCollector()
	err := collector.Start()
	if err != nil {
		t.Fatalf("Failed to start collector: %v", err)
	}
	defer collector.Stop()

	// Record counter metrics
	collector.RecordCounter("entities.created", 10, "system:test")
	collector.RecordCounter("entities.created", 5, "system:test")

	// Get metrics summary
	time.Sleep(10 * time.Millisecond) // Allow time for processing
	summary := collector.GetMetrics("entities.created", 1*time.Second)

	if summary == nil {
		t.Fatal("Expected metrics summary, got nil")
	}

	if summary.Sum != 15 {
		t.Errorf("Expected sum 15, got %f", summary.Sum)
	}

	if summary.Count != 2 {
		t.Errorf("Expected count 2, got %d", summary.Count)
	}
}

// TestRecordGauge tests gauge metric recording
func TestRecordGauge(t *testing.T) {
	collector := NewMetricsCollector()
	err := collector.Start()
	if err != nil {
		t.Fatalf("Failed to start collector: %v", err)
	}
	defer collector.Stop()

	// Record gauge metrics (only last value should be kept)
	collector.RecordGauge("memory.usage", 100.0, "type:heap")
	collector.RecordGauge("memory.usage", 150.0, "type:heap")
	collector.RecordGauge("memory.usage", 120.0, "type:heap")

	// Get metrics summary
	time.Sleep(10 * time.Millisecond)
	summary := collector.GetMetrics("memory.usage", 1*time.Second)

	if summary == nil {
		t.Fatal("Expected metrics summary, got nil")
	}

	// For gauge, the last value should be the current value
	if summary.Mean != 120.0 {
		t.Errorf("Expected current value 120.0, got %f", summary.Mean)
	}
}

// TestRecordHistogram tests histogram metric recording
func TestRecordHistogram(t *testing.T) {
	collector := NewMetricsCollector()
	err := collector.Start()
	if err != nil {
		t.Fatalf("Failed to start collector: %v", err)
	}
	defer collector.Stop()

	// Record histogram values
	values := []float64{1, 2, 3, 4, 5}
	for _, v := range values {
		collector.RecordHistogram("frame.time", v, "fps:60")
	}

	// Get metrics summary
	time.Sleep(10 * time.Millisecond)
	summary := collector.GetMetrics("frame.time", 1*time.Second)

	if summary == nil {
		t.Fatal("Expected metrics summary, got nil")
	}

	// Check basic statistics
	expectedMean := 3.0
	if summary.Mean != expectedMean {
		t.Errorf("Expected mean %f, got %f", expectedMean, summary.Mean)
	}

	if summary.Min != 1.0 {
		t.Errorf("Expected min 1.0, got %f", summary.Min)
	}

	if summary.Max != 5.0 {
		t.Errorf("Expected max 5.0, got %f", summary.Max)
	}

	if summary.Count != 5 {
		t.Errorf("Expected count 5, got %d", summary.Count)
	}
}

// TestCalculatePercentiles tests percentile calculations
func TestCalculatePercentiles(t *testing.T) {
	collector := NewMetricsCollector()
	err := collector.Start()
	if err != nil {
		t.Fatalf("Failed to start collector: %v", err)
	}
	defer collector.Stop()

	// Generate 100 values from 1 to 100
	for i := 1; i <= 100; i++ {
		collector.RecordHistogram("latency", float64(i))
	}

	time.Sleep(10 * time.Millisecond)
	summary := collector.GetMetrics("latency", 1*time.Second)

	if summary == nil {
		t.Fatal("Expected metrics summary, got nil")
	}

	// Check percentiles (allowing small margin for interpolation)
	tolerance := 1.0

	if summary.P50 < 50-tolerance || summary.P50 > 50+tolerance {
		t.Errorf("Expected P50 around 50, got %f", summary.P50)
	}

	if summary.P90 < 90-tolerance || summary.P90 > 90+tolerance {
		t.Errorf("Expected P90 around 90, got %f", summary.P90)
	}

	if summary.P95 < 95-tolerance || summary.P95 > 95+tolerance {
		t.Errorf("Expected P95 around 95, got %f", summary.P95)
	}

	if summary.P99 < 99-tolerance || summary.P99 > 99+tolerance {
		t.Errorf("Expected P99 around 99, got %f", summary.P99)
	}
}

// TestThresholdExceeded tests threshold monitoring and alerting
func TestThresholdExceeded(t *testing.T) {
	collector := NewMetricsCollector()
	err := collector.Start()
	if err != nil {
		t.Fatalf("Failed to start collector: %v", err)
	}
	defer collector.Stop()

	// Set thresholds
	collector.SetThreshold("cpu.usage", AlertLevelWarning, 80.0)
	collector.SetThreshold("cpu.usage", AlertLevelError, 90.0)

	// Record value that exceeds error threshold
	collector.RecordGauge("cpu.usage", 95.0)

	// Allow time for alert processing
	time.Sleep(50 * time.Millisecond)

	// Check alerts
	alerts := collector.GetAlerts()
	if len(alerts) == 0 {
		t.Fatal("Expected alerts to be generated")
	}

	// Find the error level alert
	found := false
	for _, alert := range alerts {
		if alert.MetricName == "cpu.usage" && alert.Level == AlertLevelError {
			found = true
			if alert.Value != 95.0 {
				t.Errorf("Expected alert value 95.0, got %f", alert.Value)
			}
			if alert.Threshold != 90.0 {
				t.Errorf("Expected threshold 90.0, got %f", alert.Threshold)
			}
		}
	}

	if !found {
		t.Error("Expected Error level alert for cpu.usage")
	}
}

// TestConcurrentRecording tests concurrent metric recording
func TestConcurrentRecording(t *testing.T) {
	collector := NewMetricsCollector()
	err := collector.Start()
	if err != nil {
		t.Fatalf("Failed to start collector: %v", err)
	}
	defer collector.Stop()

	var wg sync.WaitGroup
	numGoroutines := 100
	recordsPerGoroutine := 100

	// Start multiple goroutines recording metrics
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < recordsPerGoroutine; j++ {
				collector.RecordCounter("concurrent.test", 1.0, "goroutine:test")
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond) // Allow time for processing

	// Check that all metrics were recorded
	summary := collector.GetMetrics("concurrent.test", 1*time.Second)
	if summary == nil {
		t.Fatal("Expected metrics summary, got nil")
	}

	expectedSum := float64(numGoroutines * recordsPerGoroutine)
	if summary.Sum != expectedSum {
		t.Errorf("Expected sum %f, got %f", expectedSum, summary.Sum)
	}
}

// TestTimeWindowAggregation tests time-based metric aggregation
func TestTimeWindowAggregation(t *testing.T) {
	collector := NewMetricsCollector()
	err := collector.Start()
	if err != nil {
		t.Fatalf("Failed to start collector: %v", err)
	}
	defer collector.Stop()

	// Record metrics at different times
	collector.RecordCounter("window.test", 10.0)
	time.Sleep(500 * time.Millisecond)

	collector.RecordCounter("window.test", 20.0)
	time.Sleep(600 * time.Millisecond) // Total elapsed: 1.1s

	collector.RecordCounter("window.test", 30.0)

	// Get metrics for last 1 second (should include last two values: 20.0 and 30.0)
	summary := collector.GetMetrics("window.test", 1*time.Second)
	if summary == nil {
		t.Fatal("Expected metrics summary, got nil")
	}

	// Should include values from the last second (20.0 + 30.0 = 50.0)
	if summary.Sum != 50.0 {
		t.Errorf("Expected sum 50.0 for 1s window, got %f", summary.Sum)
	}

	// Get metrics for last 2 seconds (should include all)
	summary = collector.GetMetrics("window.test", 2*time.Second)
	if summary.Sum != 60.0 {
		t.Errorf("Expected sum 60.0 for 2s window, got %f", summary.Sum)
	}
}

// TestAlertRateLimit tests alert rate limiting
func TestAlertRateLimit(t *testing.T) {
	collector := NewMetricsCollector()
	err := collector.Start()
	if err != nil {
		t.Fatalf("Failed to start collector: %v", err)
	}
	defer collector.Stop()

	// Set threshold and rate limit
	collector.SetThreshold("rate.test", AlertLevelWarning, 50.0)

	// Trigger multiple threshold violations quickly
	for i := 0; i < 10; i++ {
		collector.RecordGauge("rate.test", 60.0)
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(100 * time.Millisecond)

	// Check that alerts are rate-limited
	alerts := collector.GetAlerts()

	// Should have limited number of alerts (not 10)
	if len(alerts) >= 10 {
		t.Errorf("Expected rate-limited alerts, got %d alerts", len(alerts))
	}
}

// TestGetAllMetrics tests retrieving all metrics
func TestGetAllMetrics(t *testing.T) {
	collector := NewMetricsCollector()
	err := collector.Start()
	if err != nil {
		t.Fatalf("Failed to start collector: %v", err)
	}
	defer collector.Stop()

	// Record different types of metrics
	collector.RecordCounter("metric.a", 10.0)
	collector.RecordGauge("metric.b", 20.0)
	collector.RecordHistogram("metric.c", 30.0)

	time.Sleep(10 * time.Millisecond)

	// Get all metrics
	allMetrics := collector.GetAllMetrics()

	if len(allMetrics) != 3 {
		t.Errorf("Expected 3 metrics, got %d", len(allMetrics))
	}

	// Check that all metrics are present
	if _, ok := allMetrics["metric.a"]; !ok {
		t.Error("metric.a not found")
	}
	if _, ok := allMetrics["metric.b"]; !ok {
		t.Error("metric.b not found")
	}
	if _, ok := allMetrics["metric.c"]; !ok {
		t.Error("metric.c not found")
	}
}

// BenchmarkMetricsOverhead benchmarks the overhead of metrics collection
func BenchmarkMetricsOverhead(b *testing.B) {
	collector := NewMetricsCollector()
	collector.Start()
	defer collector.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.RecordCounter("bench.counter", 1.0)
	}
}

// BenchmarkHighThroughput benchmarks high-throughput metric recording
func BenchmarkHighThroughput(b *testing.B) {
	collector := NewMetricsCollector()
	collector.Start()
	defer collector.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			collector.RecordHistogram("bench.histogram", float64(time.Now().UnixNano()%1000))
		}
	})
}
