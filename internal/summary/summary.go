package summary

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Stats tracks the runtime metrics.
type Stats struct {
	StartTime time.Time

	TotalURLs    uint64
	PassedDedup  uint64
	PassedStatus uint64

	// Categories
	Secrets    uint64
	Sensitives uint64 // General bucket if needed, or breakdown
	Params     uint64
	JS         uint64

	mu sync.Mutex
}

// Global instance or per-pipeline?
// Architecture says "internal/summary".
// We can make it a singleton or pass it around. Passing is cleaner.

func New() *Stats {
	return &Stats{
		StartTime: time.Now(),
	}
}

func (s *Stats) IncTotal() {
	atomic.AddUint64(&s.TotalURLs, 1)
}

func (s *Stats) IncDedup() {
	atomic.AddUint64(&s.PassedDedup, 1)
}

func (s *Stats) IncStatus() {
	atomic.AddUint64(&s.PassedStatus, 1)
}

func (s *Stats) IncCategory(cat string) {
	// Atomic maps are hard, using mutex for map usually,
	// or specific counters. Given fixed categories, specific counters are faster.
	switch cat {
	case "secret":
		atomic.AddUint64(&s.Secrets, 1)
	case "config", "backup", "vcs", "cloud":
		atomic.AddUint64(&s.Sensitives, 1) // grouping other sensitives
	case "param":
		atomic.AddUint64(&s.Params, 1)
	case "js":
		atomic.AddUint64(&s.JS, 1)
	}
}

// PrintReport outputs the final summary to stdout.
func (s *Stats) PrintReport() {
	duration := time.Since(s.StartTime)

	fmt.Println("\n========================================")
	fmt.Println("             DEFLOT SUMMARY             ")
	fmt.Println("========================================")
	fmt.Printf("Duration      : %v\n", duration)
	fmt.Printf("Total URLs    : %d\n", atomic.LoadUint64(&s.TotalURLs))
	fmt.Printf("Unique (Dedup): %d\n", atomic.LoadUint64(&s.PassedDedup))
	fmt.Printf("Live (Status) : %d\n", atomic.LoadUint64(&s.PassedStatus))
	fmt.Println("----------------------------------------")
	fmt.Println("Classification:")
	fmt.Printf("  - Secrets   : %d\n", atomic.LoadUint64(&s.Secrets))
	fmt.Printf("  - Sensitive : %d\n", atomic.LoadUint64(&s.Sensitives))
	fmt.Printf("  - Params    : %d\n", atomic.LoadUint64(&s.Params))
	fmt.Printf("  - JS Files  : %d\n", atomic.LoadUint64(&s.JS))
	fmt.Println("========================================")
}
