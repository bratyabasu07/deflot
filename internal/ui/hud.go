package ui

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elliot/deflot/internal/sources"
	"github.com/elliot/deflot/internal/summary"
)

// StartHUD starts the minimal status display.
func StartHUD(ctx context.Context, stats *summary.Stats, mgr *sources.Manager, isJSON, isStdout bool) func() {
	if isJSON || isStdout || !isTTY() {
		return func() {}
	}

	// WaitGroup to ensure we clean up propery
	var wg sync.WaitGroup
	wg.Add(1)

	// Cancellation for the internal loop
	hudCtx, cancel := context.WithCancel(ctx)

	go func() {
		defer wg.Done()
		runHUD(hudCtx, stats, mgr)
	}()

	return func() {
		cancel()
		wg.Wait()
		// Final newline
		fmt.Println()
	}
}

func runHUD(ctx context.Context, stats *summary.Stats, mgr *sources.Manager) {
	ticker := time.NewTicker(900 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Print one last final state?
			// Usually easier to just stop. The cleanup function prints a newline.
			return
		case <-ticker.C:
			renderLine(stats, mgr)
		}
	}
}

func renderLine(stats *summary.Stats, mgr *sources.Manager) {
	activeSrc := mgr.ActiveCount()
	totalURLs := atomic.LoadUint64(&stats.TotalURLs)

	phase := "FILTER"
	if activeSrc > 0 {
		phase = "COLLECT"
	}

	// DEFLOT | SRC:3 | URL:12431 | PHASE:FILTER
	line := fmt.Sprintf("DEFLOT | SRC:%d | URL:%d | PHASE:%s", activeSrc, totalURLs, phase)

	// Overwrite line
	fmt.Printf("\r\033[K%s", line)
}
