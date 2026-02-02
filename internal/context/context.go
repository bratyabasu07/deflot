package context

import (
	"errors"
	"strings"
)

// AppContext holds the immutable runtime state of the application.
// It is passed down to all modules (Sources, Pipeline, Filters, Output).
type AppContext struct {
	// Targets
	Domain    string
	InputFile string
	Wildcard  bool

	// Sources
	Sources []string // List of enabled sources (empty = all)

	// Pipeline Controls
	Workers int
	Delay   int // ms
	Timeout int // seconds
	NoDedup bool
	Match   []string // status codes to match (e.g., "200", "403")

	// Output Mode
	JSON   bool
	Stdout bool

	// Filters (Enabled/Disabled)
	Filters FilterConfig

	// Output
	OutputDir string
}

// FilterConfig defines which filters are active.
type FilterConfig struct {
	SensitiveUrls bool
	Params        bool
	JS            bool
	ExcludeLibs   bool
	PDF           bool
	Log           bool
	Config        bool
}

// ScanRecord represents a single unit of work in the pipeline.
type ScanRecord struct {
	URL        string `json:"normalized_url"`
	Source     string `json:"source"`
	StatusCode int    `json:"http_status,omitempty"`
	Category   string `json:"category"`
}

// New creates a new AppContext.
func New(domain, input string, wildcard bool, output string, sources string,
	workers, delay, timeout int, noDedup bool, mc string, jsonMode, stdoutMode bool, filters FilterConfig) (*AppContext, error) {

	// Basic Validation
	if domain == "" && input == "" {
		return nil, errors.New("context: must provide domain or input file")
	}

	// Parse Sources (comma-separated to slice)
	var sourceList []string
	if sources != "" {
		parts := strings.Split(sources, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				sourceList = append(sourceList, trimmed)
			}
		}
	}

	// Parse Match Codes
	var matchCodes []string
	if mc != "" {
		parts := strings.Split(mc, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				matchCodes = append(matchCodes, trimmed)
			}
		}
	}

	return &AppContext{
		Domain:    domain,
		InputFile: input,
		Wildcard:  wildcard,
		OutputDir: output,
		Sources:   sourceList,
		Workers:   workers,
		Delay:     delay,
		Timeout:   timeout,
		NoDedup:   noDedup,
		Match:     matchCodes,
		JSON:      jsonMode,
		Stdout:    stdoutMode,
		Filters:   filters,
	}, nil
}
