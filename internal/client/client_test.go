package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestBuildURLWithFilters(t *testing.T) {
	t.Parallel()

	u, err := BuildURL("https://api.openf1.org/v1", "laps", map[string]string{
		"session_key":   "latest",
		"driver_number": "44",
		"limit":         "5",
	}, []string{"speed>=300", "throttle<95"})
	if err != nil {
		t.Fatalf("BuildURL returned error: %v", err)
	}

	if !strings.HasPrefix(u, "https://api.openf1.org/v1/laps?") {
		t.Fatalf("unexpected URL prefix: %s", u)
	}
	for _, expected := range []string{
		"session_key=latest",
		"driver_number=44",
		"limit=5",
		"speed%3E%3D300",
		"throttle%3C95",
	} {
		if !strings.Contains(u, expected) {
			t.Fatalf("URL missing %q: %s", expected, u)
		}
	}
}

func TestClientQueryRetries429(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := calls.Add(1)
		if count == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"slow down"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"driver_number":44,"name_acronym":"HAM"}]`))
	}))
	defer srv.Close()

	c := NewWithConfig(Config{
		BaseURL:           srv.URL,
		HTTPClient:        srv.Client(),
		RequestsPerSecond: 1000,
		MaxRetries:        1,
	})
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	records, err := c.Query(ctx, "drivers", nil, nil)
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if calls.Load() != 2 {
		t.Fatalf("expected 2 calls (retry), got %d", calls.Load())
	}
}

func TestQueryObjectResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"session_key":9999}`))
	}))
	defer srv.Close()

	c := NewWithConfig(Config{
		BaseURL:           srv.URL,
		HTTPClient:        srv.Client(),
		RequestsPerSecond: 1000,
		MaxRetries:        0,
	})
	defer c.Close()

	records, err := c.Query(context.Background(), "sessions", nil, nil)
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
}
