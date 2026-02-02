package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/elliot/deflot/internal/config"
	appCtx "github.com/elliot/deflot/internal/context"
)

type AlienVault struct {
	domain string
	apiKey string
}

func NewAlienVault(domain string, cfg config.Config) *AlienVault {
	return &AlienVault{
		domain: domain,
		apiKey: cfg.ApiKeys.AlienVault,
	}
}

func (s *AlienVault) Name() string {
	return "otx"
}

func (s *AlienVault) NeedsKey() bool {
	return true
}

type otxResponse struct {
	URLList []struct {
		URL string `json:"url"`
	} `json:"url_list"`
	HasNext bool `json:"has_next"`
}

func (s *AlienVault) Run(ctx context.Context, results chan<- appCtx.ScanRecord) {
	if s.apiKey == "" {
		return
	}

	page := 1
	client := &http.Client{Timeout: 30 * time.Second}

	for {
		apiURL := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/url_list?limit=50&page=%d", s.domain, page)
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return
		}
		req.Header.Set("X-OTX-API-KEY", s.apiKey)

		var resp *http.Response
		err = WithRetry(func() error {
			var e error
			resp, e = client.Do(req)
			return e
		}, 3)

		if err != nil {
			fmt.Printf("[!] OTX Request Failed: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			// Stop on error
			break
		}

		var data otxResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			break
		}

		if len(data.URLList) == 0 {
			break
		}

		for _, item := range data.URLList {
			select {
			case <-ctx.Done():
				return
			case results <- appCtx.ScanRecord{
				URL:      item.URL,
				Source:   "otx",
				Category: "none",
			}:
			}
		}

		if !data.HasNext {
			break
		}
		page++
	}
}
