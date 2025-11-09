package util

import (
	"fmt"
	"runtime"
	"time"
)

// MemStats represents memory statistics
type MemStats struct {
	Alloc      uint64 // Currently allocated memory in bytes
	TotalAlloc uint64 // Total allocated memory in bytes
	Sys        uint64 // Total memory obtained from OS in bytes
	NumGC      uint32 // Number of garbage collections
	Goroutines int    // Number of goroutines
	HeapAlloc  uint64 // Heap allocated in bytes
	HeapSys    uint64 // Heap memory obtained from OS
	HeapIdle   uint64 // Heap idle memory
	HeapInuse  uint64 // Heap in-use memory
}

// GetMemStats returns current memory statistics
func GetMemStats() *MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &MemStats{
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		NumGC:      m.NumGC,
		Goroutines: runtime.NumGoroutine(),
		HeapAlloc:  m.HeapAlloc,
		HeapSys:    m.HeapSys,
		HeapIdle:   m.HeapInuse,
		HeapInuse:  m.HeapInuse,
	}
}

// FormatMemStats returns formatted memory statistics
func FormatMemStats(stats *MemStats) string {
	return fmt.Sprintf(`Memory Statistics:
  Allocated: %s
  Total Allocated: %s
  System Memory: %s
  Heap Allocated: %s
  Heap System: %s
  Heap In-Use: %s
  Heap Idle: %s
  Goroutines: %d
  GC Runs: %d
`,
		formatBytes(stats.Alloc),
		formatBytes(stats.TotalAlloc),
		formatBytes(stats.Sys),
		formatBytes(stats.HeapAlloc),
		formatBytes(stats.HeapSys),
		formatBytes(stats.HeapInuse),
		formatBytes(stats.HeapIdle),
		stats.Goroutines,
		stats.NumGC,
	)
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// StartMemoryMonitor starts a background goroutine that logs memory stats periodically
func StartMemoryMonitor(interval time.Duration, logger func(string)) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			stats := GetMemStats()
			logger(FormatMemStats(stats))
		}
	}()
}

// CPUStats represents CPU statistics
type CPUStats struct {
	NumCPU       int
	NumGoroutine int
}

// GetCPUStats returns current CPU-related statistics
func GetCPUStats() *CPUStats {
	return &CPUStats{
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
	}
}
