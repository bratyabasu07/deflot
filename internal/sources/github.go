package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/bratyabasu07/deflot/internal/config"
	appCtx "github.com/bratyabasu07/deflot/internal/context"
)

type GitHub struct {
	domain string
	apiKey string
}

func NewGitHub(domain string, cfg config.Config) *GitHub {
	return &GitHub{
		domain: domain,
		apiKey: cfg.ApiKeys.GitHub,
	}
}

func (s *GitHub) Name() string {
	return "github"
}

func (s *GitHub) NeedsKey() bool {
	return true
}

type githubResponse struct {
	Items []struct {
		HTMLURL string `json:"html_url"`
	} `json:"items"`
	IncompleteResults bool `json:"incomplete_results"`
}

func (s *GitHub) Run(ctx context.Context, results chan<- appCtx.ScanRecord) {
	if s.apiKey == "" {
		return
	}

	query := fmt.Sprintf("\"%s\"", s.domain)
	encodedQuery := url.QueryEscape(query)

	page := 1
	client := &http.Client{Timeout: 30 * time.Second}

	for {
		// Rate limit safe? GitHub search is 30 req/min (~2s delay)
		// We should respect that or we get 403.
		// "Respect rate limits... Handle retries internally"

		apiURL := fmt.Sprintf("https://api.github.com/search/code?q=%s&per_page=100&page=%d", encodedQuery, page)
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return
		}
		req.Header.Set("Authorization", "token "+s.apiKey)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		var resp *http.Response
		err = WithRetry(func() error {
			var e error
			resp, e = client.Do(req)
			return e
		}, 3)

		if err != nil {
			fmt.Printf("[!] GitHub Request Failed: %v\n", err)
			return
		}

		if resp.StatusCode == 403 || resp.StatusCode == 429 {
			// Rate limit
			resp.Body.Close()
			time.Sleep(2 * time.Second) // Basic
			// For robust impl, read X-RateLimit-Reset. But simplest is sleep and retry or stop.
			// Given "streaming", blocking 60s is bad.
			// We'll stop on heavy rate limit.
			break
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			break
		}

		var data githubResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			resp.Body.Close()
			break
		}
		resp.Body.Close()

		for _, item := range data.Items {
			select {
			case <-ctx.Done():
				return
			case results <- appCtx.ScanRecord{
				URL:      item.HTMLURL,
				Source:   "github",
				Category: "none",
			}:
			}
		}

		if len(data.Items) == 0 || page >= 5 { // simple cap to avoid deep paging
			break
		}
		page++

		// GitHub search API limit is strict, let's sleep a bit between pages
		time.Sleep(2 * time.Second)
	}
}
