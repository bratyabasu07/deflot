package pipeline

import (
	"context"
	appCtx "github.com/elliot/deflot/internal/context"
	"github.com/elliot/deflot/internal/dedup"
	"github.com/elliot/deflot/internal/filters"
	"github.com/elliot/deflot/internal/integrations/jssecrethunter"
	"github.com/elliot/deflot/internal/normalize"
	"github.com/elliot/deflot/internal/output"
	"github.com/elliot/deflot/internal/status"
	"github.com/elliot/deflot/internal/summary"
	"strings"
	"sync"
	"time"
)

// Pipeline orchestrates the flow of data.
type Pipeline struct {
	appCtx  *appCtx.AppContext
	dedup   *dedup.Dedup
	checker *status.Checker
	filter  *filters.Engine
	writer  *output.Writer
	stats   *summary.Stats

	jsScanner *jssecrethunter.Scanner
	scannerWg sync.WaitGroup

	notify func(string)
}

// New creates a new pipeline instance.
func New(ctx *appCtx.AppContext, d *dedup.Dedup, c *status.Checker, f *filters.Engine, w *output.Writer, s *summary.Stats, notify func(string), js *jssecrethunter.Scanner) *Pipeline {
	return &Pipeline{
		appCtx:    ctx,
		dedup:     d,
		checker:   c,
		filter:    f,
		writer:    w,
		stats:     s,
		notify:    notify,
		jsScanner: js,
	}
}

// Start initiates the processing pipeline.
// It accepts an input channel (from sources) and returns a completion channel.
func (p *Pipeline) Start(ctx context.Context, input <-chan appCtx.ScanRecord) <-chan struct{} {
	done := make(chan struct{})

	// 1. Worker Pool for Processing
	// We do NOT want to process line-by-line sequentially if we have network IO (like HTTP checks).
	// However, normalization and filtering are CPU bound and fast.
	// The HTTP check (Gate) is the bottleneck. The architecture says "Affected by: -w, --workers".
	// So we fan-out here.

	var wg sync.WaitGroup
	workers := p.appCtx.Workers
	if workers < 1 {
		workers = 1
	}

	// We use a buffered channel for intermediate steps if needed, but here we can just
	// fan out consumers from the single input channel.

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			p.worker(ctx, input)
		}(i)
	}

	// Waiter routine
	go func() {
		wg.Wait()          // Wait for all workers to finish processing stream
		p.scannerWg.Wait() // Wait for any async background scanners
		close(done)
	}()

	return done
}

// worker allows concurrent processing of the stream.
func (p *Pipeline) worker(ctx context.Context, input <-chan appCtx.ScanRecord) {
	// Respect delay if set (rate limiting)
	// Note: Global delay across all workers vs per-worker?
	// Usually per-request. If delay is set, we sleep.
	delay := time.Duration(p.appCtx.Delay) * time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return
		case record, ok := <-input:
			if !ok {
				return
			}

			// Pre-check: empty validation
			record.URL = strings.TrimSpace(record.URL)
			if record.URL == "" {
				continue
			}

			if delay > 0 {
				time.Sleep(delay)
			}

			p.processRecord(record)
		}
	}
}

// processRecord pushes the item through the logical steps.
func (p *Pipeline) processRecord(record appCtx.ScanRecord) {
	p.stats.IncTotal()

	// 1. Normalize
	validatedURL, err := normalize.Normalize(record.URL)
	if err != nil {
		return // Skip invalid
	}
	record.URL = validatedURL

	// 2. Dedup Gate
	if p.dedup.Check(record.URL) == dedup.Drop {
		return
	}
	p.stats.IncDedup()

	// 3. Status Gate (if enabled checks)
	passed, code := p.checker.Check(record.URL)
	if !passed {
		return
	}
	record.StatusCode = code
	p.stats.IncStatus()

	// 4. Filter Classification
	category := p.filter.Classify(record.URL)
	if category != "none" {
		p.stats.IncCategory(category)
		if p.notify != nil {
			p.notify(category)
		}

		// Trigger JS Scan if enabled
		if category == filters.CatJS && p.jsScanner != nil {
			p.scannerWg.Add(1)
			go func(u string) {
				defer p.scannerWg.Done()
				// TODO: Add context with timeout?
				_ = p.jsScanner.Scan(context.Background(), u, p.appCtx.OutputDir)
			}(record.URL)
		}
	}
	record.Category = category

	// 5. Output
	// If category is none, but we passed all gates, we output to default list logic inside writer
	if err := p.writer.Write(record); err != nil {
		// Log error?
	}
}
