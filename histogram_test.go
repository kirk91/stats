package stats

// func TestHistogramStatistic(t *testing.T) {
// hist := buckethist.New()
// for i := 2; i < 100; i++ {
// hist.RecordValue(uint64(i))
// }
// hstats := newHistogramStatistics(hist)
// assert.Equal(t, defaultSupportedQuantiles, hstats.SupportedQuantiles())
// assert.Equal(t, len(defaultSupportedQuantiles), len(hstats.ComputedQuantiles()))
// assert.Equal(t, hist.ApproxMax(), hstats.Max())
// assert.Equal(t, hist.ApproxMin(), hstats.Min())
// assert.Equal(t, hist.ApproxSum(), hstats.Sum())
// assert.Equal(t, hist.Count(), hstats.Count())

// buckets := hist.Buckets()
// bucketVals := make([]uint64, len(buckets))
// for i := 0; i < len(buckets); i++ {
// bucketVals[i] = buckets[i].Count()
// }
// assert.Equal(t, bucketVals, hstats.BucketVals())
// }

// func TestHistogram(t *testing.T) {
// h := NewHistogram(nil, "foo.bar", "foo", nil)
// assert.Contains(t, h.Summary(), "No recorded values")

// istats := h.IntervalStatistics()
// assert.Equal(t, uint64(0), istats.Count())

// h.Record(3)
// istats = h.IntervalStatistics()
// assert.Equal(t, uint64(1), istats.Count())

// // refresh
// h.refreshIntervalStatistic()
// istats = h.IntervalStatistics()
// assert.Equal(t, uint64(1), istats.Count())

// // refresh twice
// h.refreshIntervalStatistic()
// istats = h.IntervalStatistics()
// assert.Equal(t, uint64(0), istats.Count())

// assert.Contains(t, h.Summary(), "P0")
// assert.Contains(t, h.Summary(), "P25")
// assert.Contains(t, h.Summary(), "P50")
// assert.Contains(t, h.Summary(), "P95")
// assert.Contains(t, h.Summary(), "P100")
// }
