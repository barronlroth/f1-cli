package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultBaseURL    = "https://api.openf1.org/v1"
	defaultTimeout    = 20 * time.Second
	defaultRPS        = 3
	defaultMaxRetries = 2
)

type Config struct {
	BaseURL           string
	HTTPClient        *http.Client
	RequestsPerSecond int
	MaxRetries        int
}

type Response struct {
	Body       []byte
	Header     http.Header
	StatusCode int
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	limiter    <-chan time.Time
	ticker     *time.Ticker
	maxRetries int
}

func New(baseURL string, httpClient *http.Client) *Client {
	cfg := Config{
		BaseURL:           baseURL,
		HTTPClient:        httpClient,
		RequestsPerSecond: defaultRPS,
		MaxRetries:        defaultMaxRetries,
	}
	return NewWithConfig(cfg)
}

func NewWithConfig(cfg Config) *Client {
	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}

	rps := cfg.RequestsPerSecond
	if rps <= 0 {
		rps = defaultRPS
	}

	maxRetries := cfg.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	interval := time.Second / time.Duration(rps)
	if interval <= 0 {
		interval = time.Millisecond
	}
	ticker := time.NewTicker(interval)

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		limiter:    ticker.C,
		ticker:     ticker,
		maxRetries: maxRetries,
	}
}

func (c *Client) Close() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
}

func (c *Client) Query(ctx context.Context, endpoint string, params map[string]string, filters []string) ([]map[string]any, error) {
	resp, err := c.GetRaw(ctx, endpoint, params, filters)
	if err != nil {
		return nil, err
	}

	var payload any
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return nil, fmt.Errorf("decode JSON response: %w", err)
	}

	records, ok := toRecordSlice(payload)
	if !ok {
		return nil, errors.New("expected JSON array or object response")
	}

	return records, nil
}

func (c *Client) GetRaw(ctx context.Context, endpoint string, params map[string]string, filters []string) (*Response, error) {
	requestURL, err := BuildURL(c.baseURL, endpoint, params, filters)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if err := c.waitForRateLimit(ctx); err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt < c.maxRetries {
			lastErr = fmt.Errorf("openf1 %s: status 429", endpoint)
			if err := sleepWithContext(ctx, retryDelay(resp.Header.Get("Retry-After"), attempt)); err != nil {
				return nil, err
			}
			continue
		}

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			return nil, fmt.Errorf("openf1 %s: status %d: %s", endpoint, resp.StatusCode, truncateBody(string(body)))
		}

		return &Response{
			Body:       body,
			Header:     resp.Header.Clone(),
			StatusCode: resp.StatusCode,
		}, nil
	}

	return nil, lastErr
}

func (c *Client) waitForRateLimit(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.limiter:
		return nil
	}
}

func toRecordSlice(payload any) ([]map[string]any, bool) {
	switch typed := payload.(type) {
	case []any:
		out := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			record, ok := item.(map[string]any)
			if !ok {
				return nil, false
			}
			out = append(out, record)
		}
		return out, true
	case map[string]any:
		return []map[string]any{typed}, true
	default:
		return nil, false
	}
}

func retryDelay(retryAfterHeader string, attempt int) time.Duration {
	retryAfterHeader = strings.TrimSpace(retryAfterHeader)
	if retryAfterHeader != "" {
		if seconds, err := strconv.Atoi(retryAfterHeader); err == nil && seconds >= 0 {
			return time.Duration(seconds) * time.Second
		}

		if ts, err := http.ParseTime(retryAfterHeader); err == nil {
			delay := time.Until(ts)
			if delay > 0 {
				return delay
			}
		}
	}

	// Basic exponential backoff: 1s, 2s, 4s...
	return time.Second << attempt
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func truncateBody(body string) string {
	const max = 256
	body = strings.TrimSpace(body)
	if len(body) <= max {
		return body
	}
	return body[:max] + "..."
}
