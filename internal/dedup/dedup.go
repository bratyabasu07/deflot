package dedup

import (
	"net/url"
	"strings"
	"sync"
)

// CheckResult indicates if a URL should proceed.
type CheckResult bool

const (
	Pass CheckResult = true
	Drop CheckResult = false
)

// Dedup handles duplication checking and scope enforcement.
type Dedup struct {
	seen         sync.Map
	targetDomain string
	wildcard     bool
	disableDedup bool
}

// New creates a new deduplicator.
func New(targetDomain string, wildcard bool, disableDedup bool) *Dedup {
	return &Dedup{
		targetDomain: strings.ToLower(targetDomain),
		wildcard:     wildcard,
		disableDedup: disableDedup,
	}
}

// Check returns Pass if the URL is valid and unseen, Drop otherwise.
func (d *Dedup) Check(rawURL string) CheckResult {
	// 1. Scope Check (Wildcard Logic)
	// We parse again here because we need the Host.
	// (Optimization: Pass cached parsed URL if possible later)
	u, err := url.Parse(rawURL)
	if err != nil {
		return Drop
	}

	if !d.inScope(u.Host) {
		return Drop
	}

	// 2. Dedup Check
	if d.disableDedup {
		return Pass
	}

	// We use the full rawURL as the key.
	// Since it's already normalized by the pipeline, this is safe.
	if _, loaded := d.seen.LoadOrStore(rawURL, true); loaded {
		// Already seen
		return Drop
	}

	return Pass
}

// inScope checks if the host matches the target rules.
func (d *Dedup) inScope(host string) bool {
	host = strings.ToLower(host)

	// If no domain was given (e.g. input file mode without domain constraint), pass everything?
	// Architecture says "REQUIRED: -d". So we assume strict mode.
	if d.targetDomain == "" {
		return true
	}

	if host == d.targetDomain {
		return true
	}

	if d.wildcard {
		// Check for .targetDomain suffix
		suffix := "." + d.targetDomain
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}

	return false
}
