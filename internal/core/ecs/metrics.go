package ecs

import (
	"math"
	"sort"
	"sync"
	"time"
)

// AlertLevel represents the severity of an alert
type AlertLevel int

const (
	AlertLevelWarning AlertLevel = iota
	AlertLevelError
	AlertLevelCritical
)

// MetricType represents the type of metric
type MetricType int

const (
	MetricTypeCounter MetricType = iota
	MetricTypeGauge
	MetricTypeHistogram
)

// MetricsSummary contains aggregated metrics data
type MetricsSummary struct {
	Name      string
	Count     int64
	Sum       float64
	Mean      float64
	Min       float64
	Max       float64
	StdDev    float64
	P50       float64
	P90       float64
	P95       float64
	P99       float64
	Timestamp time.Time
}

// Alert represents a threshold violation alert
type Alert struct {
	MetricName string
	Level      AlertLevel
	Value      float64
	Threshold  float64
	Message    string
	Timestamp  time.Time
}

// MetricsCollector interface for collecting and managing metrics
type MetricsCollector interface {
	RecordCounter(name string, value float64, tags ...string)
	RecordGauge(name string, value float64, tags ...string)
	RecordHistogram(name string, value float64, tags ...string)
	GetMetrics(name string, window time.Duration) *MetricsSummary
	GetAllMetrics() map[string]*MetricsSummary
	SetThreshold(name string, level AlertLevel, value float64)
	GetAlerts() []Alert
	ClearAlerts()
	Start() error
	Stop() error
}

// metricPoint represents a single metric data point
type metricPoint struct {
	value      float64
	timestamp  time.Time
	metricType MetricType
}

// threshold represents a metric threshold configuration
type threshold struct {
	level     AlertLevel
	value     float64
	lastAlert time.Time
}

// metricsCollectorImpl is the concrete implementation of MetricsCollector
type metricsCollectorImpl struct {
	mu         sync.RWMutex
	metrics    map[string][]metricPoint
	thresholds map[string]map[AlertLevel]*threshold
	alerts     []Alert
	alertMu    sync.RWMutex
	running    bool
	stopCh     chan struct{}

	// Rate limiting for alerts (1 alert per metric per minute)
	alertRateLimit time.Duration
}

// NewMetricsCollector creates a new metrics collector instance
func NewMetricsCollector() MetricsCollector {
	return &metricsCollectorImpl{
		metrics:        make(map[string][]metricPoint),
		thresholds:     make(map[string]map[AlertLevel]*threshold),
		alerts:         make([]Alert, 0),
		stopCh:         make(chan struct{}),
		alertRateLimit: 1 * time.Minute,
	}
}

// Start begins the metrics collection
func (m *metricsCollectorImpl) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return nil
	}

	m.running = true

	// Start background cleanup goroutine
	go m.cleanupOldMetrics()

	return nil
}

// Stop halts the metrics collection
func (m *metricsCollectorImpl) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.running = false
	close(m.stopCh)

	return nil
}

// RecordCounter records a counter metric (cumulative)
func (m *metricsCollectorImpl) RecordCounter(name string, value float64, tags ...string) {
	m.recordMetric(name, value, MetricTypeCounter)
}

// RecordGauge records a gauge metric (point-in-time value)
func (m *metricsCollectorImpl) RecordGauge(name string, value float64, tags ...string) {
	m.recordMetric(name, value, MetricTypeGauge)
	m.checkThresholds(name, value)
}

// RecordHistogram records a histogram metric (distribution)
func (m *metricsCollectorImpl) RecordHistogram(name string, value float64, tags ...string) {
	m.recordMetric(name, value, MetricTypeHistogram)
}

// recordMetric is the internal method for recording metrics
func (m *metricsCollectorImpl) recordMetric(name string, value float64, metricType MetricType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	point := metricPoint{
		value:      value,
		timestamp:  time.Now(),
		metricType: metricType,
	}

	if _, exists := m.metrics[name]; !exists {
		m.metrics[name] = make([]metricPoint, 0, 1000)
	}

	m.metrics[name] = append(m.metrics[name], point)

	// Keep only last 10000 points per metric to prevent unbounded growth
	if len(m.metrics[name]) > 10000 {
		m.metrics[name] = m.metrics[name][len(m.metrics[name])-10000:]
	}
}

// GetMetrics returns aggregated metrics for a specific time window
func (m *metricsCollectorImpl) GetMetrics(name string, window time.Duration) *MetricsSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	points, exists := m.metrics[name]
	if !exists || len(points) == 0 {
		return nil
	}

	now := time.Now()
	cutoff := now.Add(-window)

	// Filter points within the time window
	var filteredPoints []metricPoint
	for _, p := range points {
		if p.timestamp.After(cutoff) {
			filteredPoints = append(filteredPoints, p)
		}
	}

	if len(filteredPoints) == 0 {
		return nil
	}

	// Calculate statistics
	summary := &MetricsSummary{
		Name:      name,
		Timestamp: now,
		Count:     int64(len(filteredPoints)),
	}

	// Extract values
	values := make([]float64, len(filteredPoints))
	for i, p := range filteredPoints {
		values[i] = p.value
	}

	// For gauge metrics, use only the last value for mean
	if len(filteredPoints) > 0 && filteredPoints[0].metricType == MetricTypeGauge {
		summary.Mean = filteredPoints[len(filteredPoints)-1].value
		summary.Sum = summary.Mean
		summary.Min = summary.Mean
		summary.Max = summary.Mean
		return summary
	}

	// Calculate basic statistics
	summary.Sum = 0
	summary.Min = values[0]
	summary.Max = values[0]

	for _, v := range values {
		summary.Sum += v
		if v < summary.Min {
			summary.Min = v
		}
		if v > summary.Max {
			summary.Max = v
		}
	}

	summary.Mean = summary.Sum / float64(len(values))

	// Calculate standard deviation
	var variance float64
	for _, v := range values {
		diff := v - summary.Mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	summary.StdDev = math.Sqrt(variance)

	// Calculate percentiles
	if len(values) > 0 {
		sort.Float64s(values)
		summary.P50 = percentile(values, 0.50)
		summary.P90 = percentile(values, 0.90)
		summary.P95 = percentile(values, 0.95)
		summary.P99 = percentile(values, 0.99)
	}

	return summary
}

// percentile calculates the percentile value from a sorted slice
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	index := p * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// GetAllMetrics returns summaries for all metrics
func (m *metricsCollectorImpl) GetAllMetrics() map[string]*MetricsSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*MetricsSummary)

	for name := range m.metrics {
		if summary := m.GetMetrics(name, 1*time.Minute); summary != nil {
			result[name] = summary
		}
	}

	return result
}

// SetThreshold sets an alert threshold for a metric
func (m *metricsCollectorImpl) SetThreshold(name string, level AlertLevel, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.thresholds[name]; !exists {
		m.thresholds[name] = make(map[AlertLevel]*threshold)
	}

	m.thresholds[name][level] = &threshold{
		level: level,
		value: value,
	}
}

// checkThresholds checks if a metric value exceeds any thresholds
func (m *metricsCollectorImpl) checkThresholds(name string, value float64) {
	m.mu.RLock()
	thresholds, exists := m.thresholds[name]
	m.mu.RUnlock()

	if !exists {
		return
	}

	now := time.Now()

	for level, thresh := range thresholds {
		if value > thresh.value {
			// Check rate limiting
			if now.Sub(thresh.lastAlert) < m.alertRateLimit {
				continue
			}

			// Create alert
			alert := Alert{
				MetricName: name,
				Level:      level,
				Value:      value,
				Threshold:  thresh.value,
				Message:    getAlertMessage(name, level, value, thresh.value),
				Timestamp:  now,
			}

			// Update last alert time
			thresh.lastAlert = now

			// Store alert
			m.alertMu.Lock()
			m.alerts = append(m.alerts, alert)

			// Keep only last 1000 alerts
			if len(m.alerts) > 1000 {
				m.alerts = m.alerts[len(m.alerts)-1000:]
			}
			m.alertMu.Unlock()
		}
	}
}

// getAlertMessage generates an alert message
func getAlertMessage(metric string, level AlertLevel, value, threshold float64) string {
	levelStr := "WARNING"
	switch level {
	case AlertLevelError:
		levelStr = "ERROR"
	case AlertLevelCritical:
		levelStr = "CRITICAL"
	}

	return levelStr + ": " + metric + " exceeded threshold"
}

// GetAlerts returns all current alerts
func (m *metricsCollectorImpl) GetAlerts() []Alert {
	m.alertMu.RLock()
	defer m.alertMu.RUnlock()

	result := make([]Alert, len(m.alerts))
	copy(result, m.alerts)
	return result
}

// ClearAlerts removes all alerts
func (m *metricsCollectorImpl) ClearAlerts() {
	m.alertMu.Lock()
	defer m.alertMu.Unlock()

	m.alerts = m.alerts[:0]
}

// cleanupOldMetrics periodically removes old metric data
func (m *metricsCollectorImpl) cleanupOldMetrics() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			cutoff := time.Now().Add(-5 * time.Minute)

			for name, points := range m.metrics {
				// Find the first point that's newer than cutoff
				i := 0
				for i < len(points) && points[i].timestamp.Before(cutoff) {
					i++
				}

				if i > 0 {
					m.metrics[name] = points[i:]
				}

				// Remove metric entirely if no recent data
				if len(m.metrics[name]) == 0 {
					delete(m.metrics, name)
				}
			}
			m.mu.Unlock()

		case <-m.stopCh:
			return
		}
	}
}
