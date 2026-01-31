package ui

import (
	"time"
)

// ThreadPoolTimeSeries manages time-series data for thread pool metrics
type ThreadPoolTimeSeries struct {
	dataPoints    []ThreadPoolDataPoint
	maxDataPoints int
	lastSnapshot  *ThreadPoolSnapshot
}

// NewThreadPoolTimeSeries creates a new thread pool time series tracker
func NewThreadPoolTimeSeries(maxDataPoints int) *ThreadPoolTimeSeries {
	return &ThreadPoolTimeSeries{
		dataPoints:    make([]ThreadPoolDataPoint, 0, maxDataPoints),
		maxDataPoints: maxDataPoints,
	}
}

// AddSnapshot adds a new snapshot and calculates rejection rates
func (ts *ThreadPoolTimeSeries) AddSnapshot(snapshot *ThreadPoolSnapshot) {
	if ts.lastSnapshot == nil {
		// First snapshot - use as baseline only
		ts.lastSnapshot = snapshot
		return
	}

	// Calculate time delta
	timeDelta := snapshot.Timestamp.Sub(ts.lastSnapshot.Timestamp).Seconds()
	if timeDelta <= 0 {
		// Skip if no time has passed
		return
	}

	// Create new data point
	dataPoint := ThreadPoolDataPoint{
		Timestamp: snapshot.Timestamp,
		Pools:     make(map[string]ThreadPoolMetrics),
	}

	// Calculate metrics for each pool
	for poolName, currentStats := range snapshot.Pools {
		metrics := ThreadPoolMetrics{
			QueueDepth: float64(currentStats.QueueDepth),
		}

		// Calculate rejection rate
		if lastStats, ok := ts.lastSnapshot.Pools[poolName]; ok {
			rejectedDelta := currentStats.RejectedTotal - lastStats.RejectedTotal
			if rejectedDelta < 0 {
				// Counter reset (cluster restart) - set to 0
				rejectedDelta = 0
			}
			metrics.RejectionRate = float64(rejectedDelta) / timeDelta
		}

		dataPoint.Pools[poolName] = metrics
	}

	// Add to ring buffer
	if len(ts.dataPoints) >= ts.maxDataPoints {
		// Remove oldest
		ts.dataPoints = ts.dataPoints[1:]
	}
	ts.dataPoints = append(ts.dataPoints, dataPoint)

	// Update last snapshot
	ts.lastSnapshot = snapshot
}

// ThreadPoolSummary contains summary statistics for a thread pool
type ThreadPoolSummary struct {
	CurrentQueue      float64
	AverageQueue      float64
	PeakQueue         float64
	MinQueue          float64
	CurrentRejections float64
	AverageRejections float64
	PeakRejections    float64
}

// CalculateSummary computes summary statistics for each pool
func (ts *ThreadPoolTimeSeries) CalculateSummary() map[string]ThreadPoolSummary {
	summary := make(map[string]ThreadPoolSummary)

	if len(ts.dataPoints) == 0 {
		return summary
	}

	// Track values for each pool
	poolQueues := make(map[string][]float64)
	poolRejections := make(map[string][]float64)

	for _, dp := range ts.dataPoints {
		for poolName, metrics := range dp.Pools {
			poolQueues[poolName] = append(poolQueues[poolName], metrics.QueueDepth)
			poolRejections[poolName] = append(poolRejections[poolName], metrics.RejectionRate)
		}
	}

	// Calculate summary for each pool
	for poolName := range poolQueues {
		queues := poolQueues[poolName]
		rejections := poolRejections[poolName]

		s := ThreadPoolSummary{}

		// Queue depth stats
		if len(queues) > 0 {
			s.CurrentQueue = queues[len(queues)-1]
			s.MinQueue = queues[0]
			s.PeakQueue = queues[0]
			sum := 0.0
			for _, v := range queues {
				sum += v
				if v < s.MinQueue {
					s.MinQueue = v
				}
				if v > s.PeakQueue {
					s.PeakQueue = v
				}
			}
			s.AverageQueue = sum / float64(len(queues))
		}

		// Rejection rate stats
		if len(rejections) > 0 {
			s.CurrentRejections = rejections[len(rejections)-1]
			sum := 0.0
			for _, v := range rejections {
				sum += v
				if v > s.PeakRejections {
					s.PeakRejections = v
				}
			}
			s.AverageRejections = sum / float64(len(rejections))
		}

		summary[poolName] = s
	}

	return summary
}

// GetDataPoints returns all data points
func (ts *ThreadPoolTimeSeries) GetDataPoints() []ThreadPoolDataPoint {
	return ts.dataPoints
}

// Size returns the number of data points
func (ts *ThreadPoolTimeSeries) Size() int {
	return len(ts.dataPoints)
}

// GetTimeRange returns the time range of the data
func (ts *ThreadPoolTimeSeries) GetTimeRange() (time.Time, time.Time) {
	if len(ts.dataPoints) == 0 {
		return time.Time{}, time.Time{}
	}
	return ts.dataPoints[0].Timestamp, ts.dataPoints[len(ts.dataPoints)-1].Timestamp
}

// Clear removes all data points
func (ts *ThreadPoolTimeSeries) Clear() {
	ts.dataPoints = make([]ThreadPoolDataPoint, 0, ts.maxDataPoints)
	ts.lastSnapshot = nil
}
