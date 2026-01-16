package ui

import (
	"testing"
	"time"
)

func TestNewMetricsTimeSeries(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	if mts == nil {
		t.Fatal("NewMetricsTimeSeries returned nil")
	}

	if mts.MaxSize != 12 {
		t.Errorf("MaxSize = %d, want 12", mts.MaxSize)
	}

	if mts.LastSnapshot != nil {
		t.Errorf("LastSnapshot should be nil initially")
	}

	if len(mts.DataPoints) != 0 {
		t.Errorf("DataPoints length = %d, want 0", len(mts.DataPoints))
	}
}

func TestMetricsTimeSeries_AddSnapshot_FirstSnapshot(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	snapshot := &MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  1000,
		SearchTotal: 500,
	}

	added := mts.AddSnapshot(snapshot)

	// First snapshot should not add a data point
	if added {
		t.Error("First snapshot should return false")
	}

	if len(mts.DataPoints) != 0 {
		t.Errorf("DataPoints length = %d, want 0 after first snapshot", len(mts.DataPoints))
	}

	if mts.LastSnapshot != snapshot {
		t.Error("LastSnapshot should be set to the first snapshot")
	}
}

func TestMetricsTimeSeries_AddSnapshot_NormalOperation(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	// First snapshot
	snapshot1 := &MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  1000,
		SearchTotal: 500,
	}
	mts.AddSnapshot(snapshot1)

	// Second snapshot (5 seconds later, 500 more inserts, 200 more searches)
	snapshot2 := &MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  1500,
		SearchTotal: 700,
	}
	added := mts.AddSnapshot(snapshot2)

	if !added {
		t.Error("Second snapshot should return true")
	}

	if len(mts.DataPoints) != 1 {
		t.Fatalf("DataPoints length = %d, want 1", len(mts.DataPoints))
	}

	dp := mts.DataPoints[0]

	// Expected rates: 500/5 = 100 inserts/sec, 200/5 = 40 searches/sec
	expectedInsertRate := 100.0
	expectedSearchRate := 40.0

	if dp.InsertRate != expectedInsertRate {
		t.Errorf("InsertRate = %f, want %f (calculation: 500 inserts / 5 seconds)", dp.InsertRate, expectedInsertRate)
	}

	if dp.SearchRate != expectedSearchRate {
		t.Errorf("SearchRate = %f, want %f (calculation: 200 searches / 5 seconds)", dp.SearchRate, expectedSearchRate)
	}

	if !dp.Timestamp.Equal(snapshot2.Timestamp) {
		t.Errorf("Timestamp = %v, want %v (should match second snapshot)", dp.Timestamp, snapshot2.Timestamp)
	}
}

func TestMetricsTimeSeries_AddSnapshot_CounterReset(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	// First snapshot
	snapshot1 := &MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  1000,
		SearchTotal: 500,
	}
	mts.AddSnapshot(snapshot1)

	// Second snapshot with lower totals (cluster restart)
	snapshot2 := &MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  100,  // Reset to lower value
		SearchTotal: 50,   // Reset to lower value
	}
	added := mts.AddSnapshot(snapshot2)

	if !added {
		t.Error("Should add data point even with counter reset")
	}

	if len(mts.DataPoints) != 1 {
		t.Fatalf("DataPoints length = %d, want 1", len(mts.DataPoints))
	}

	dp := mts.DataPoints[0]

	// Negative rates should be converted to 0
	if dp.InsertRate != 0 {
		t.Errorf("InsertRate = %f, want 0 (counter reset)", dp.InsertRate)
	}

	if dp.SearchRate != 0 {
		t.Errorf("SearchRate = %f, want 0 (counter reset)", dp.SearchRate)
	}
}

func TestMetricsTimeSeries_AddSnapshot_ZeroTimeDelta(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	// First snapshot
	snapshot1 := &MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  1000,
		SearchTotal: 500,
	}
	mts.AddSnapshot(snapshot1)

	// Second snapshot with same timestamp
	snapshot2 := &MetricsSnapshot{
		Timestamp:   time.Unix(0, 0), // Same time
		IndexTotal:  1500,
		SearchTotal: 700,
	}
	added := mts.AddSnapshot(snapshot2)

	if added {
		t.Error("Should not add data point with zero time delta")
	}

	if len(mts.DataPoints) != 0 {
		t.Errorf("DataPoints length = %d, want 0 (zero time delta)", len(mts.DataPoints))
	}
}

func TestMetricsTimeSeries_AddSnapshot_NilSnapshot(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	added := mts.AddSnapshot(nil)

	if added {
		t.Error("Should not add nil snapshot")
	}

	if len(mts.DataPoints) != 0 {
		t.Errorf("DataPoints length = %d, want 0", len(mts.DataPoints))
	}
}

func TestMetricsTimeSeries_AddSnapshot_RingBufferWraparound(t *testing.T) {
	mts := NewMetricsTimeSeries(3) // Small buffer for testing

	baseTime := time.Unix(0, 0)

	// Add initial snapshot
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   baseTime,
		IndexTotal:  0,
		SearchTotal: 0,
	})

	// Add 5 more snapshots (should wrap around at 3)
	for i := 1; i <= 5; i++ {
		snapshot := &MetricsSnapshot{
			Timestamp:   baseTime.Add(time.Duration(i*5) * time.Second),
			IndexTotal:  int64(i * 100),
			SearchTotal: int64(i * 50),
		}
		mts.AddSnapshot(snapshot)
	}

	// Should only have last 3 data points
	if len(mts.DataPoints) != 3 {
		t.Errorf("DataPoints length = %d, want 3 (ring buffer limit)", len(mts.DataPoints))
	}

	// Verify oldest data point is from snapshot 3 (not 1 or 2)
	// After adding 5 snapshots to buffer size 3, first 2 should be dropped
	oldestTime := baseTime.Add(15 * time.Second) // i=3
	if !mts.DataPoints[0].Timestamp.Equal(oldestTime) {
		t.Errorf("Oldest timestamp = %v, want %v (ring buffer should have dropped first 2 snapshots)",
			mts.DataPoints[0].Timestamp, oldestTime)
	}

	// Verify newest data point is from snapshot 5
	newestTime := baseTime.Add(25 * time.Second) // i=5
	if !mts.DataPoints[2].Timestamp.Equal(newestTime) {
		t.Errorf("Newest timestamp = %v, want %v (most recent snapshot should be retained)",
			mts.DataPoints[2].Timestamp, newestTime)
	}

	// Verify the buffer contains exactly snapshots 3, 4, 5 in order
	expectedTimes := []time.Time{
		baseTime.Add(15 * time.Second), // i=3
		baseTime.Add(20 * time.Second), // i=4
		baseTime.Add(25 * time.Second), // i=5
	}
	for i, expectedTime := range expectedTimes {
		if !mts.DataPoints[i].Timestamp.Equal(expectedTime) {
			t.Errorf("DataPoint[%d] timestamp = %v, want %v (ring buffer order)",
				i, mts.DataPoints[i].Timestamp, expectedTime)
		}
	}
}

func TestMetricsTimeSeries_CalculateSummary_EmptyBuffer(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	summary := mts.CalculateSummary("insert")

	if summary.Current != 0 || summary.Average != 0 || summary.Peak != 0 || summary.Min != 0 {
		t.Errorf("Summary for empty buffer should be all zeros, got %+v", summary)
	}
}

func TestMetricsTimeSeries_CalculateSummary_SingleDataPoint(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	// Add two snapshots to get one data point
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  0,
		SearchTotal: 0,
	})
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  500,
		SearchTotal: 200,
	})

	summary := mts.CalculateSummary("insert")

	expectedRate := 100.0 // 500/5

	if summary.Current != expectedRate {
		t.Errorf("Current = %f, want %f", summary.Current, expectedRate)
	}
	if summary.Average != expectedRate {
		t.Errorf("Average = %f, want %f", summary.Average, expectedRate)
	}
	if summary.Peak != expectedRate {
		t.Errorf("Peak = %f, want %f", summary.Peak, expectedRate)
	}
	if summary.Min != expectedRate {
		t.Errorf("Min = %f, want %f", summary.Min, expectedRate)
	}
}

func TestMetricsTimeSeries_CalculateSummary_MultipleDataPoints(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	// Add initial snapshot
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  0,
		SearchTotal: 0,
	})

	// Add varying rates - using cumulative totals
	// Rates we want: 100, 200, 150, 300, 250
	// Cumulative totals at 5s intervals:
	cumulativeTotals := []int64{
		500,  // 100/s * 5s = 500
		1500, // 500 + (200/s * 5s) = 500 + 1000 = 1500
		2250, // 1500 + (150/s * 5s) = 1500 + 750 = 2250
		3750, // 2250 + (300/s * 5s) = 2250 + 1500 = 3750
		5000, // 3750 + (250/s * 5s) = 3750 + 1250 = 5000
	}

	for i, total := range cumulativeTotals {
		snapshot := &MetricsSnapshot{
			Timestamp:   time.Unix(int64((i+1)*5), 0),
			IndexTotal:  total,
			SearchTotal: 0,
		}
		mts.AddSnapshot(snapshot)
	}

	summary := mts.CalculateSummary("insert")

	// Current should be last rate
	if summary.Current != 250.0 {
		t.Errorf("Current = %f, want 250.0", summary.Current)
	}

	// Peak should be max rate
	if summary.Peak != 300.0 {
		t.Errorf("Peak = %f, want 300.0", summary.Peak)
	}

	// Min should be min rate
	if summary.Min != 100.0 {
		t.Errorf("Min = %f, want 100.0", summary.Min)
	}

	// Average should be (100+200+150+300+250)/5 = 200
	expectedAvg := 200.0
	if summary.Average != expectedAvg {
		t.Errorf("Average = %f, want %f", summary.Average, expectedAvg)
	}
}

func TestMetricsTimeSeries_CalculateSummary_SearchMetric(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  0,
		SearchTotal: 0,
	})
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  500,
		SearchTotal: 250,
	})

	summary := mts.CalculateSummary("search")

	expectedRate := 50.0 // 250/5

	if summary.Current != expectedRate {
		t.Errorf("Current = %f, want %f for search metric", summary.Current, expectedRate)
	}
}

func TestMetricsTimeSeries_CalculateSummary_UnknownMetric(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  0,
		SearchTotal: 0,
	})
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  500,
		SearchTotal: 250,
	})

	summary := mts.CalculateSummary("unknown")

	// Unknown metric should return zero summary
	if summary.Current != 0 || summary.Average != 0 {
		t.Errorf("Unknown metric should return zero summary, got %+v", summary)
	}
}

func TestMetricsTimeSeries_GetDataPoints(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  0,
		SearchTotal: 0,
	})
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  500,
		SearchTotal: 250,
	})

	points := mts.GetDataPoints()

	if len(points) != 1 {
		t.Errorf("GetDataPoints returned %d points, want 1", len(points))
	}

	// Modify the copy
	points[0].InsertRate = 999

	// Original should be unchanged
	if mts.DataPoints[0].InsertRate == 999 {
		t.Error("GetDataPoints should return a copy, not the original slice")
	}
}

func TestMetricsTimeSeries_Clear(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  0,
		SearchTotal: 0,
	})
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  500,
		SearchTotal: 250,
	})

	if len(mts.DataPoints) != 1 || mts.LastSnapshot == nil {
		t.Fatal("Setup failed")
	}

	mts.Clear()

	if len(mts.DataPoints) != 0 {
		t.Errorf("After Clear, DataPoints length = %d, want 0", len(mts.DataPoints))
	}

	if mts.LastSnapshot != nil {
		t.Error("After Clear, LastSnapshot should be nil")
	}
}

func TestMetricsTimeSeries_Size(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	if mts.Size() != 0 {
		t.Errorf("Initial Size = %d, want 0", mts.Size())
	}

	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  0,
		SearchTotal: 0,
	})
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  500,
		SearchTotal: 250,
	})

	if mts.Size() != 1 {
		t.Errorf("Size = %d, want 1", mts.Size())
	}
}

func TestMetricsTimeSeries_IsFull(t *testing.T) {
	mts := NewMetricsTimeSeries(2) // Small buffer

	if mts.IsFull() {
		t.Error("Empty buffer should not be full")
	}

	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  0,
		SearchTotal: 0,
	})
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  500,
		SearchTotal: 250,
	})

	if mts.IsFull() {
		t.Error("Buffer with 1 data point should not be full (max=2)")
	}

	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(10, 0),
		IndexTotal:  1000,
		SearchTotal: 500,
	})

	if !mts.IsFull() {
		t.Error("Buffer with 2 data points should be full (max=2)")
	}
}

func TestMetricsTimeSeries_GetTimeRange(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	// No data points
	if mts.GetTimeRange() != 0 {
		t.Error("TimeRange should be 0 for empty buffer")
	}

	// One data point
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  0,
		SearchTotal: 0,
	})
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  500,
		SearchTotal: 250,
	})

	if mts.GetTimeRange() != 0 {
		t.Error("TimeRange should be 0 for single data point")
	}

	// Multiple data points
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(10, 0),
		IndexTotal:  1000,
		SearchTotal: 500,
	})

	expectedRange := 5 * time.Second // From 5s to 10s
	if mts.GetTimeRange() != expectedRange {
		t.Errorf("TimeRange = %v, want %v", mts.GetTimeRange(), expectedRange)
	}
}

// Edge case tests

func TestNewMetricsTimeSeries_ZeroMaxSize(t *testing.T) {
	mts := NewMetricsTimeSeries(0)

	if mts == nil {
		t.Fatal("NewMetricsTimeSeries should not return nil even with zero maxSize")
	}

	if mts.MaxSize != 0 {
		t.Errorf("MaxSize = %d, want 0", mts.MaxSize)
	}

	// Should handle adding snapshots gracefully
	added1 := mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  1000,
		SearchTotal: 500,
	})

	// First snapshot should be stored even with zero maxSize (for rate calculation)
	if added1 {
		t.Error("First snapshot should return false (no data point added)")
	}

	added2 := mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  1500,
		SearchTotal: 700,
	})

	// Second snapshot should not be added when maxSize is 0
	if added2 {
		t.Error("With zero maxSize, AddSnapshot should return false (cannot add to buffer)")
	}

	// With zero maxSize, buffer should remain empty
	if len(mts.DataPoints) != 0 {
		t.Errorf("With zero maxSize, DataPoints should remain empty, got length %d", len(mts.DataPoints))
	}
}

func TestNewMetricsTimeSeries_NegativeMaxSize(t *testing.T) {
	mts := NewMetricsTimeSeries(-5)

	if mts == nil {
		t.Fatal("NewMetricsTimeSeries should not return nil even with negative maxSize")
	}

	// Negative size is unusual but should be handled
	if mts.MaxSize != -5 {
		t.Errorf("MaxSize = %d, want -5", mts.MaxSize)
	}

	// Should handle adding snapshots gracefully
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  1000,
		SearchTotal: 500,
	})

	added := mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  1500,
		SearchTotal: 700,
	})

	// With negative maxSize, should not add data points
	if added {
		t.Error("With negative maxSize, AddSnapshot should return false")
	}

	if len(mts.DataPoints) != 0 {
		t.Errorf("With negative maxSize, DataPoints should remain empty, got length %d", len(mts.DataPoints))
	}
}

func TestMetricsTimeSeries_AddSnapshot_FractionalSeconds(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	// First snapshot
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  0,
		SearchTotal: 0,
	})

	// Second snapshot 1.5 seconds later
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 1500000000), // 1.5 seconds in nanoseconds
		IndexTotal:  300,
		SearchTotal: 150,
	})

	if len(mts.DataPoints) != 1 {
		t.Fatalf("DataPoints length = %d, want 1", len(mts.DataPoints))
	}

	// Expected rates: 300/1.5 = 200 inserts/sec, 150/1.5 = 100 searches/sec
	expectedInsertRate := 200.0
	expectedSearchRate := 100.0

	if mts.DataPoints[0].InsertRate != expectedInsertRate {
		t.Errorf("InsertRate = %f, want %f", mts.DataPoints[0].InsertRate, expectedInsertRate)
	}

	if mts.DataPoints[0].SearchRate != expectedSearchRate {
		t.Errorf("SearchRate = %f, want %f", mts.DataPoints[0].SearchRate, expectedSearchRate)
	}
}

func TestMetricsTimeSeries_AddSnapshot_LargeTimeDelta(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	// First snapshot
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  0,
		SearchTotal: 0,
	})

	// Second snapshot 1 hour later
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(3600, 0),
		IndexTotal:  360000,
		SearchTotal: 180000,
	})

	if len(mts.DataPoints) != 1 {
		t.Fatalf("DataPoints length = %d, want 1", len(mts.DataPoints))
	}

	// Expected rates: 360000/3600 = 100 inserts/sec, 180000/3600 = 50 searches/sec
	expectedInsertRate := 100.0
	expectedSearchRate := 50.0

	if mts.DataPoints[0].InsertRate != expectedInsertRate {
		t.Errorf("InsertRate = %f, want %f", mts.DataPoints[0].InsertRate, expectedInsertRate)
	}

	if mts.DataPoints[0].SearchRate != expectedSearchRate {
		t.Errorf("SearchRate = %f, want %f", mts.DataPoints[0].SearchRate, expectedSearchRate)
	}
}

func TestMetricsTimeSeries_AddSnapshot_VeryLargeCounters(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	// Test with very large int64 values (but not at max to avoid overflow)
	largeValue := int64(1000000000000) // 1 trillion

	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  largeValue,
		SearchTotal: largeValue / 2,
	})

	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  largeValue + 5000,
		SearchTotal: (largeValue / 2) + 2500,
	})

	if len(mts.DataPoints) != 1 {
		t.Fatalf("DataPoints length = %d, want 1", len(mts.DataPoints))
	}

	// Expected rates: 5000/5 = 1000 inserts/sec, 2500/5 = 500 searches/sec
	expectedInsertRate := 1000.0
	expectedSearchRate := 500.0

	if mts.DataPoints[0].InsertRate != expectedInsertRate {
		t.Errorf("InsertRate = %f, want %f", mts.DataPoints[0].InsertRate, expectedInsertRate)
	}

	if mts.DataPoints[0].SearchRate != expectedSearchRate {
		t.Errorf("SearchRate = %f, want %f", mts.DataPoints[0].SearchRate, expectedSearchRate)
	}
}

func TestMetricsTimeSeries_CalculateSummary_AllIdenticalValues(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	// Add initial snapshot
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  0,
		SearchTotal: 0,
	})

	// Add multiple snapshots with identical rates (100/s)
	for i := 1; i <= 5; i++ {
		mts.AddSnapshot(&MetricsSnapshot{
			Timestamp:   time.Unix(int64(i*5), 0),
			IndexTotal:  int64(i * 500), // Always 100/s rate
			SearchTotal: int64(i * 250), // Always 50/s rate
		})
	}

	summary := mts.CalculateSummary("insert")

	// All values are 100.0
	if summary.Current != 100.0 {
		t.Errorf("Current = %f, want 100.0", summary.Current)
	}

	if summary.Average != 100.0 {
		t.Errorf("Average = %f, want 100.0", summary.Average)
	}

	if summary.Peak != 100.0 {
		t.Errorf("Peak = %f, want 100.0", summary.Peak)
	}

	if summary.Min != 100.0 {
		t.Errorf("Min = %f, want 100.0", summary.Min)
	}
}

func TestMetricsTimeSeries_AddSnapshot_MixedCounterResets(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	// First snapshot
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  1000,
		SearchTotal: 500,
	})

	// Normal increase
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  1500,
		SearchTotal: 700,
	})

	// Only search counter resets (mixed scenario)
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(10, 0),
		IndexTotal:  2000, // Still increasing
		SearchTotal: 50,   // Reset
	})

	if len(mts.DataPoints) != 2 {
		t.Fatalf("DataPoints length = %d, want 2", len(mts.DataPoints))
	}

	// First data point should have normal rates
	if mts.DataPoints[0].InsertRate != 100.0 {
		t.Errorf("First InsertRate = %f, want 100.0", mts.DataPoints[0].InsertRate)
	}

	if mts.DataPoints[0].SearchRate != 40.0 {
		t.Errorf("First SearchRate = %f, want 40.0", mts.DataPoints[0].SearchRate)
	}

	// Second data point should have normal insert rate but zero search rate
	if mts.DataPoints[1].InsertRate != 100.0 {
		t.Errorf("Second InsertRate = %f, want 100.0", mts.DataPoints[1].InsertRate)
	}

	if mts.DataPoints[1].SearchRate != 0.0 {
		t.Errorf("Second SearchRate = %f, want 0.0 (counter reset)", mts.DataPoints[1].SearchRate)
	}
}

func TestMetricsTimeSeries_Clear_PreservesMaxSize(t *testing.T) {
	originalMaxSize := 12
	mts := NewMetricsTimeSeries(originalMaxSize)

	// Add data
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  0,
		SearchTotal: 0,
	})
	mts.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  500,
		SearchTotal: 250,
	})

	// Clear
	mts.Clear()

	// MaxSize should be preserved
	if mts.MaxSize != originalMaxSize {
		t.Errorf("After Clear, MaxSize = %d, want %d", mts.MaxSize, originalMaxSize)
	}

	// Capacity should be preserved
	if cap(mts.DataPoints) != originalMaxSize {
		t.Errorf("After Clear, cap(DataPoints) = %d, want %d", cap(mts.DataPoints), originalMaxSize)
	}
}

func TestMetricsTimeSeries_GetDataPoints_EmptyBuffer(t *testing.T) {
	mts := NewMetricsTimeSeries(12)

	points := mts.GetDataPoints()

	if points == nil {
		t.Error("GetDataPoints should not return nil, even for empty buffer")
	}

	if len(points) != 0 {
		t.Errorf("GetDataPoints for empty buffer should return empty slice, got length %d", len(points))
	}
}
