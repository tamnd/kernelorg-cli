// Package kernelorg provides a client for the kernel.org releases JSON API.
// All data comes from a single endpoint: GET /releases.json.
// Filtering by moniker or version is performed in-memory after the fetch.
package kernelorg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Config holds the configuration for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Retries   int
	Timeout   time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://www.kernel.org",
		UserAgent: "kernel-cli/0.1 (+https://github.com/tamnd/kernelorg-cli)",
		Rate:      1 * time.Second,
		Retries:   3,
		Timeout:   15 * time.Second,
	}
}

// Client talks to kernel.org over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
	last time.Time
	mu   sync.Mutex
}

// NewClient returns a Client using the given Config.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// Releases fetches /releases.json and returns all releases.
// If moniker is non-empty, only releases matching that moniker are returned.
func (c *Client) Releases(ctx context.Context, moniker string) ([]Release, error) {
	resp, err := c.fetchReleases(ctx)
	if err != nil {
		return nil, err
	}
	var out []Release
	for _, w := range resp.Releases {
		if moniker != "" && w.Moniker != moniker {
			continue
		}
		out = append(out, toRelease(w))
	}
	return out, nil
}

// GetRelease fetches /releases.json and returns the release matching version exactly.
// Returns ErrNotFound if no such version exists.
func (c *Client) GetRelease(ctx context.Context, version string) (Release, error) {
	resp, err := c.fetchReleases(ctx)
	if err != nil {
		return Release{}, err
	}
	for _, w := range resp.Releases {
		if w.Version == version {
			return toRelease(w), nil
		}
	}
	return Release{}, ErrNotFound
}

// LatestStable returns the latest stable kernel version.
func (c *Client) LatestStable(ctx context.Context) (LatestStable, error) {
	resp, err := c.fetchReleases(ctx)
	if err != nil {
		return LatestStable{}, err
	}
	return LatestStable{Version: resp.LatestStable.Version}, nil
}

// fetchReleases performs the HTTP GET for /releases.json with retry logic.
func (c *Client) fetchReleases(ctx context.Context) (*wireResponse, error) {
	url := c.cfg.BaseURL + "/releases.json"
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			var wr wireResponse
			if err := json.Unmarshal(body, &wr); err != nil {
				return nil, fmt.Errorf("decode releases.json: %w", err)
			}
			return &wr, nil
		}
		lastErr = err
		if !retry {
			return nil, lastErr
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

// do performs a single HTTP GET, returning whether the caller should retry.
func (c *Client) do(ctx context.Context, url string) (body []byte, retry bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

// pace enforces the minimum inter-request gap.
func (c *Client) pace() {
	if c.cfg.Rate <= 0 {
		return
	}
	c.mu.Lock()
	wait := c.cfg.Rate - time.Since(c.last)
	c.mu.Unlock()
	if wait > 0 {
		time.Sleep(wait)
	}
	c.mu.Lock()
	c.last = time.Now()
	c.mu.Unlock()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
