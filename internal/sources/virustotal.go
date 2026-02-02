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

type VirusTotal struct {
	domain string
	apiKey string
}

func NewVirusTotal(domain string, cfg config.Config) *VirusTotal {
	return &VirusTotal{
		domain: domain,
		apiKey: cfg.ApiKeys.VirusTotal,
	}
}

func (s *VirusTotal) Name() string {
	return "virustotal"
}

func (s *VirusTotal) NeedsKey() bool {
	return true
}

type vtResponse struct {
	Data []struct {
		Id string `json:"id"` // The subdomain/url
	} `json:"data"`
	Links struct {
		Next string `json:"next"`
	} `json:"links"`
}

func (s *VirusTotal) Run(ctx context.Context, results chan<- appCtx.ScanRecord) {
	if s.apiKey == "" {
		return
	}

	// Fetch subdomains
	cursor := ""
	baseURL := fmt.Sprintf("https://www.virustotal.com/api/v3/domains/%s/subdomains?limit=40", s.domain)

	client := &http.Client{Timeout: 30 * time.Second}

	for {
		url := baseURL
		if cursor != "" {
			url = cursor // Next link is full URL
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			break
		}
		req.Header.Set("x-apikey", s.apiKey)

		var resp *http.Response
		err = WithRetry(func() error {
			var e error
			resp, e = client.Do(req)
			return e
		}, 3)

		if err != nil {
			fmt.Printf("[!] VT Request Failed: %v\n", err)
			break
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			// 401/403 means bad key, stop
			if resp.StatusCode == 401 || resp.StatusCode == 403 {
				fmt.Println("[!] VT Auth failed")
				break
			}
			// Rate limit?
			if resp.StatusCode == 429 {
				time.Sleep(5 * time.Second) // naive backoff
				// retry? loop continues, might process next?
				// actually we just consumed the body.
				// For now, break on error.
				break
			}
			break
		}

		var vtResp vtResponse
		if err := json.NewDecoder(resp.Body).Decode(&vtResp); err != nil {
			break
		}

		for _, item := range vtResp.Data {
			select {
			case <-ctx.Done():
				return
			case results <- appCtx.ScanRecord{
				URL:      item.Id,
				Source:   "virustotal",
				Category: "none",
			}:
			}
		}

		if vtResp.Links.Next == "" {
			break
		}
		cursor = vtResp.Links.Next
	}
}
