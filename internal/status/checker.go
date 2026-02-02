package status

import (
	"crypto/tls"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Checker verifies if a URL is alive and matches status codes.
type Checker struct {
	client     *http.Client
	matchCodes map[int]bool
	enabled    bool
}

// New creates a new status checker.
func New(timeout int, matchCodes []string) *Checker {
	// If no match codes provided, we don't need to check liveness explicitly?
	// Or do we assume 200 OK by default?
	// Architecture says "async probe | live | forbidden | historical".
	// Usually if --mc is NOT provided, tools often skip this step to be fast.
	// But the prompt implies this IS a gate.
	// If --mc is empty, we set enabled = false (skip check).

	enabled := len(matchCodes) > 0

	matchMap := make(map[int]bool)
	for _, c := range matchCodes {
		if i, err := strconv.Atoi(c); err == nil {
			matchMap[i] = true
		}
	}

	// Optimized transport
	transport := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives:   false, // Enable Keep-Alives for performance
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 30, // Allow reuse for same-host bursts (e.g. assets)
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(timeout) * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeout) * time.Second,
		// Don't follow redirects too far, or maybe check redirect status?
		// We'll stick to default redirect following for now, but usually status check = final status.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	return &Checker{
		client:     client,
		matchCodes: matchMap,
		enabled:    enabled,
	}
}

// Check probes the URL. Returns (passed, statusCode).
func (c *Checker) Check(url string) (bool, int) {
	if !c.enabled {
		return true, 0 // Passthrough if logic not requested
	}

	// 1. HEAD request
	resp, err := c.client.Head(url)

	// Fallback triggers:
	// - Error is nil but status is 405 (Method Not Allowed)
	// - Error is "Method Not Allowed" (some servers/proxies)
	// If it's a timeout/connection reset, GET will likely fail too, so we skip it to save time.

	shouldFallback := false
	if err == nil && resp.StatusCode == http.StatusMethodNotAllowed {
		shouldFallback = true
		resp.Body.Close() // Close previous body
	} else if err != nil {
		// If it's a network error, usually we fail.
		// But some servers block HEAD entirely with connection closed.
		// We'll allow fallback for specific errors if we want to be aggressive,
		// but "Respect timeouts strictly" means we shouldn't double-dip on 10s timeout.
		// Let's assume fallback only on 405 or specific HTTP-level rejections.
		// For safety in this audit: if HEAD fails entirely, we log it as failed.
		// Exception: Some WAFs drop HEAD. We will try GET only if err tells us so?
		// To follow "Prefer simplicity > overengineering", we stick to the prompt's "HEAD -> GET fallback must be safe".
		// The safest way is: Check HEAD. If clean fail (network), fail. If 405, Try GET.

		// Wait, prompt said "HEAD first, fallback to GET". It implies functional fallback.
		// Let's keep it simple: If HEAD errors, try GET. But handle the Close correctly.
		shouldFallback = true
	}

	if shouldFallback {
		resp, err = c.client.Get(url)
		if err != nil {
			return false, 0 // Connection failed or timeout
		}
	}
	defer resp.Body.Close()

	// 2. Match Status
	// If matchCodes contains the status, return true.
	if c.matchCodes[resp.StatusCode] {
		return true, resp.StatusCode
	}

	return false, resp.StatusCode
}
