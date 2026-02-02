package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bratyabasu07/deflot/internal/config"
	appCtx "github.com/bratyabasu07/deflot/internal/context"
)

type URLScan struct {
	domain string
	apiKey string
}

func NewURLScan(domain string, cfg config.Config) *URLScan {
	return &URLScan{
		domain: domain,
		apiKey: cfg.ApiKeys.URLScan,
	}
}

func (s *URLScan) Name() string {
	return "urlscan"
}

func (s *URLScan) NeedsKey() bool {
	return true
}

type urlScanResponse struct {
	Results []struct {
		Page struct {
			URL string `json:"url"`
		} `json:"page"`
	} `json:"results"`
	Total int `json:"total"`
}

func (s *URLScan) Run(ctx context.Context, results chan<- appCtx.ScanRecord) {
	if s.apiKey == "" {
		return
	}

	// Basic search, size 1000 to get a good chunk. Not full pagination for brevity in this task,
	// but good enough for production "quick" recon.
	apiURL := fmt.Sprintf("https://urlscan.io/api/v1/search/?q=domain:%s&size=1000", s.domain)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return
	}
	req.Header.Set("API-Key", s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	err = WithRetry(func() error {
		var e error
		resp, e = client.Do(req)
		return e
	}, 3)

	if err != nil {
		fmt.Printf("[!] URLScan Request Failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("[!] URLScan Status: %d\n", resp.StatusCode)
		return
	}

	var data urlScanResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return
	}

	for _, res := range data.Results {
		select {
		case <-ctx.Done():
			return
		case results <- appCtx.ScanRecord{
			URL:      res.Page.URL,
			Source:   "urlscan",
			Category: "none",
		}:
		}
	}
}
