package ui

import "time"

// NewMetricsTimeSeries creates a new MetricsTimeSeries with specified buffer size
func NewMetricsTimeSeries(maxSize int) *MetricsTimeSeries {
	// Handle edge case of zero or negative maxSize
	capacity := maxSize
	if capacity < 0 {
		capacity = 0
	}

	return &MetricsTimeSeries{
		DataPoints:   make([]MetricsDataPoint, 0, capacity),
		MaxSize:      maxSize,
		LastSnapshot: nil,
	}
}

// AddSnapshot calculates rates from the snapshot and adds a new data point
// to the time series buffer. Returns true if a data point was added, false otherwise.
func (mts *MetricsTimeSeries) AddSnapshot(snapshot *MetricsSnapshot) bool {
	if snapshot == nil {
		return false
	}

	// Skip first snapshot (no previous data to compare)
	if mts.LastSnapshot == nil {
		mts.LastSnapshot = snapshot
		return false
	}

	// Calculate time delta
	timeDelta := snapshot.Timestamp.Sub(mts.LastSnapshot.Timestamp).Seconds()

	// Avoid division by zero
	if timeDelta == 0 {
		return false
	}

	// Calculate deltas
	indexDelta := float64(snapshot.IndexTotal - mts.LastSnapshot.IndexTotal)
	searchDelta := float64(snapshot.SearchTotal - mts.LastSnapshot.SearchTotal)

	// Calculate rates (per second)
	insertRate := indexDelta / timeDelta
	searchRate := searchDelta / timeDelta

	// Handle negative values (cluster restart, counter reset)
	if insertRate < 0 {
		insertRate = 0
	}
	if searchRate < 0 {
		searchRate = 0
	}

	// Create data point
	dataPoint := MetricsDataPoint{
		Timestamp:  snapshot.Timestamp,
		InsertRate: insertRate,
		SearchRate: searchRate,
	}

	// Add to ring buffer
	// Skip if MaxSize is zero or negative (degenerate case)
	if mts.MaxSize <= 0 {
		return false
	}

	if len(mts.DataPoints) < mts.MaxSize {
		// Buffer not full yet, append
		mts.DataPoints = append(mts.DataPoints, dataPoint)
	} else {
		// Buffer full, shift left and add at end (simple ring buffer)
		mts.DataPoints = append(mts.DataPoints[1:], dataPoint)
	}

	// Update last snapshot
	mts.LastSnapshot = snapshot

	return true
}

// CalculateSummary computes aggregate statistics for a specific metric type
func (mts *MetricsTimeSeries) CalculateSummary(metricType string) MetricsSummary {
	if len(mts.DataPoints) == 0 {
		return MetricsSummary{
			Current: 0,
			Average: 0,
			Peak:    0,
			Min:     0,
		}
	}

	// Extract values based on metric type
	var values []float64
	for _, dp := range mts.DataPoints {
		var val float64
		switch metricType {
		case "insert":
			val = dp.InsertRate
		case "search":
			val = dp.SearchRate
		default:
			// Unknown metric type
			return MetricsSummary{}
		}
		values = append(values, val)
	}

	// Calculate statistics
	var sum, min, max float64
	min = values[0]
	max = values[0]

	for _, v := range values {
		sum += v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	avg := sum / float64(len(values))
	current := values[len(values)-1]

	return MetricsSummary{
		Current: current,
		Average: avg,
		Peak:    max,
		Min:     min,
	}
}

// GetDataPoints returns a copy of the current data points
func (mts *MetricsTimeSeries) GetDataPoints() []MetricsDataPoint {
	points := make([]MetricsDataPoint, len(mts.DataPoints))
	copy(points, mts.DataPoints)
	return points
}

// Clear resets the time series, clearing all data points and snapshot
func (mts *MetricsTimeSeries) Clear() {
	mts.DataPoints = make([]MetricsDataPoint, 0, mts.MaxSize)
	mts.LastSnapshot = nil
}

// Size returns the current number of data points in the buffer
func (mts *MetricsTimeSeries) Size() int {
	return len(mts.DataPoints)
}

// IsFull returns whether the buffer has reached max capacity
func (mts *MetricsTimeSeries) IsFull() bool {
	return len(mts.DataPoints) >= mts.MaxSize
}

// GetTimeRange returns the time span covered by current data points
// Returns zero duration if there are fewer than 2 data points
func (mts *MetricsTimeSeries) GetTimeRange() time.Duration {
	if len(mts.DataPoints) < 2 {
		return 0
	}

	oldest := mts.DataPoints[0].Timestamp
	newest := mts.DataPoints[len(mts.DataPoints)-1].Timestamp

	return newest.Sub(oldest)
}
