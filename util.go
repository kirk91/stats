package stats

import (
	"context"
	"runtime"
	"syscall"
	"time"
)

// CollectRuntimeMetrics collects the runtime metrics.
func CollectRuntimeMetrics(ctx context.Context, store *Store, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	scope := store.CreateScope("runtime.")
	defer store.DeleteScope(scope)

	var lastNumGC uint32
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		var usage syscall.Rusage
		syscall.Getrusage(syscall.RUSAGE_SELF, &usage)
		stime := usage.Stime.Sec + int64(usage.Stime.Usec)/1e6
		utime := usage.Utime.Sec + int64(usage.Utime.Usec)/1e6
		scope.Gauge("cpu_utime").Set(uint64(utime))
		scope.Gauge("cpu_stime").Set(uint64(stime))

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		numGC := memStats.NumGC
		if numGC < lastNumGC {
			lastNumGC = 0
		}
		scope.Gauge("num_gc").Set(uint64(numGC))
		scope.Gauge("num_gc_persecond").Set(uint64(numGC - lastNumGC))

		if numGC-lastNumGC >= 256 {
			lastNumGC = numGC - 255
		}
		for i := lastNumGC; i < numGC; i++ {
			pause := memStats.PauseNs[i%256] / uint64(time.Millisecond)
			scope.Histogram("gc_pause_ms").Record(pause)
		}
		lastNumGC = numGC

		scope.Gauge("num_goroutines").Set(uint64(runtime.NumGoroutine()))
		scope.Gauge("gc_pause_total_ms").Set(memStats.PauseTotalNs / uint64(time.Millisecond))
		scope.Gauge("alloc_bytes").Set(memStats.Alloc)
		scope.Gauge("total_alloc_bytes").Set(memStats.TotalAlloc)
		scope.Gauge("sys_bytes").Set(memStats.Sys)
		scope.Gauge("heap_alloc_bytes").Set(memStats.HeapAlloc)
		scope.Gauge("heap_sys_bytes").Set(memStats.HeapSys)
		scope.Gauge("heap_idle_bytes").Set(memStats.HeapIdle)
		scope.Gauge("heap_inuse_bytes").Set(memStats.HeapInuse)
		scope.Gauge("heap_released_bytes").Set(memStats.HeapReleased)
		scope.Gauge("heap_objects").Set(memStats.HeapObjects)
		scope.Gauge("stack_inuse_bytes").Set(memStats.StackInuse)
		scope.Gauge("stack_sys_bytes").Set(memStats.StackSys)
		scope.Gauge("lookups").Set(memStats.Lookups)
		scope.Gauge("mallocs").Set(memStats.Mallocs)
		scope.Gauge("frees").Set(memStats.Frees)
	}
}
