package sources

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bratyabasu07/deflot/internal/config"
	appCtx "github.com/bratyabasu07/deflot/internal/context"
)

// Source is the interface that all modular sources must implement.
type Source interface {
	Name() string
	// Run starts the crawling process.
	// It pushes discovered URLs to the results channel.
	// It must respect the context for cancellation.
	Run(ctx context.Context, results chan<- appCtx.ScanRecord)
	// NeedsKey returns true if this source requires an API key.
	NeedsKey() bool
}

// Manager handles the lifecycle of all sources.
type Manager struct {
	appCtx  *appCtx.AppContext
	config  config.Config
	sources []Source

	activeSources int64
}

// NewManager creates a new source manager.
func NewManager(ctx *appCtx.AppContext, cfg config.Config) *Manager {
	return &Manager{
		appCtx: ctx,
		config: cfg,
	}
}

// Register adds a source to the manager.
func (m *Manager) Register(s Source) {
	m.sources = append(m.sources, s)
}

// ActiveCount returns the number of currently running sources.
func (m *Manager) ActiveCount() int64 {
	return atomic.LoadInt64(&m.activeSources)
}

// StartAll runs all registered and enabled sources concurrently.
// It returns a single channel that aggregates all URLs.
func (m *Manager) StartAll(ctx context.Context) <-chan appCtx.ScanRecord {
	out := make(chan appCtx.ScanRecord, 100) // Buffered to prevent blocking sources on small hiccups
	var wg sync.WaitGroup

	// Get API keys for checking availability
	keys := config.GetAPIKeys()

	for _, src := range m.sources {
		// 1. Check if source is enabled by user
		if !m.isSourceEnabled(src.Name()) {
			continue
		}

		// 2. Check if API key is present if required
		if src.NeedsKey() {
			if !m.hasKey(src.Name(), keys) {
				fmt.Printf("[!] Skipping %s: Missing API Key\n", src.Name())
				continue
			}
		}

		wg.Add(1)
		atomic.AddInt64(&m.activeSources, 1)

		// Start Indicator
		if !m.appCtx.JSON && !m.appCtx.Stdout && isTTY() {
			fmt.Printf("\r→ %s\n", src.Name())
		}

		go func(s Source) {
			defer wg.Done()
			defer atomic.AddInt64(&m.activeSources, -1)

			// End Indicator (defer triggers on exit)
			defer func() {
				if !m.appCtx.JSON && !m.appCtx.Stdout && isTTY() {
					fmt.Printf("\r← %s\n", s.Name())
				}
			}()

			s.Run(ctx, out)
		}(src)
	}

	// Closer routine
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// isSourceEnabled checks if the user allowed this specific source.
// If AppCtx.Sources is empty, ALL sources are enabled.
func (m *Manager) isSourceEnabled(name string) bool {
	if len(m.appCtx.Sources) == 0 {
		return true
	}
	for _, s := range m.appCtx.Sources {
		if s == name {
			return true
		}
	}
	return false
}

// hasKey checks availability of specific keys.
// This is a simple helper mapping source names to key struct fields.
func (m *Manager) hasKey(name string, keys config.ApiKeys) bool {
	switch name {
	case "virustotal":
		return keys.VirusTotal != ""
	case "urlscan":
		return keys.URLScan != ""
	default:
		return true // If it needs checks but isn't listed, assume open or handled inside
	}
}

// Helper: Retry with exponential backoff
// Usage: sources should call this inside their Run loop if they implement HTTP requests manually.
func WithRetry(operation func() error, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if err = operation(); err == nil {
			return nil
		}

		// Exponential backoff: 1s, 2s, 4s...
		duration := time.Duration(1<<i) * time.Second
		time.Sleep(duration)
	}
	return fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}

// isTTY matches the logic in internal/ui/intro.go but duplicated to avoid import cycles.
func isTTY() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
