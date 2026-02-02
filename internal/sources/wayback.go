package sources

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"time"

	appCtx "github.com/bratyabasu07/deflot/internal/context"
)

type Wayback struct {
	domain string
}

func NewWayback(domain string) *Wayback {
	return &Wayback{domain: domain}
}

func (s *Wayback) Name() string {
	return "wayback"
}

func (s *Wayback) NeedsKey() bool {
	return false
}

func (s *Wayback) Run(ctx context.Context, results chan<- appCtx.ScanRecord) {
	// Wayback CDX API
	// Using output=txt for stream-friendly processing
	apiURL := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=*.%s/*&output=txt&fl=original&collapse=urlkey", s.domain)

	client := &http.Client{
		Timeout: 45 * time.Second, // Increased timeout for stability
	}

	err := WithRetry(func() error {
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return err // Context error or bad URL
		}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("bad status: %d", resp.StatusCode)
		}

		// Stream results directly inside the retry block?
		// No, if we stream valid data and THEN fail, we might duplicate on retry?
		// But CDX failure usually happens at connection start.
		// For a streaming API, 'WithRetry' is tricky.
		// If we read 50% lines and fail, retrying might duplicate.
		// However, for Recon, better to retry and maybe have duplicates (Deduplicator handles it) than missing data.

		scanner := bufio.NewScanner(resp.Body)
		// Increase buffer size for long URLs
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			raw := scanner.Text()
			if raw == "" {
				continue
			}

			select {
			case <-ctx.Done():
				return nil // Stop processing
			case results <- appCtx.ScanRecord{URL: raw, Source: "wayback", Category: "none"}:
			}
		}

		return scanner.Err()
	}, 3)

	if err != nil {
		fmt.Printf("[!] Wayback Failed after retries: %v\n", err)
	}
}
