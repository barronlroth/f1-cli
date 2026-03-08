package driver

import (
	"context"
	"errors"
	"testing"
)

type mockQueryAPI struct {
	calls    int
	response []map[string]any
	err      error
}

func (m *mockQueryAPI) Query(ctx context.Context, endpoint string, params map[string]string, filters []string) ([]map[string]any, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func TestResolveDriverNumberPassthrough(t *testing.T) {
	t.Parallel()

	api := &mockQueryAPI{}
	resolver := NewResolver(api)

	got, err := resolver.Resolve(context.Background(), "44", "latest")
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if got != "44" {
		t.Fatalf("expected 44, got %s", got)
	}
	if api.calls != 0 {
		t.Fatalf("expected 0 API calls, got %d", api.calls)
	}
}

func TestResolveAcronymCachesResult(t *testing.T) {
	t.Parallel()

	api := &mockQueryAPI{
		response: []map[string]any{
			{"driver_number": float64(1), "name_acronym": "VER"},
		},
	}
	resolver := NewResolver(api)

	got1, err := resolver.Resolve(context.Background(), "ver", "latest")
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	got2, err := resolver.Resolve(context.Background(), "VER", "latest")
	if err != nil {
		t.Fatalf("Resolve returned error on second call: %v", err)
	}

	if got1 != "1" || got2 != "1" {
		t.Fatalf("expected both resolutions to be 1, got %s and %s", got1, got2)
	}
	if api.calls != 1 {
		t.Fatalf("expected 1 API call due to cache, got %d", api.calls)
	}
}

func TestResolveAcronymNotFound(t *testing.T) {
	t.Parallel()

	api := &mockQueryAPI{response: []map[string]any{}}
	resolver := NewResolver(api)

	_, err := resolver.Resolve(context.Background(), "XYZ", "latest")
	if err == nil {
		t.Fatal("expected error for missing acronym, got nil")
	}
}

func TestResolveAPIError(t *testing.T) {
	t.Parallel()

	api := &mockQueryAPI{err: errors.New("boom")}
	resolver := NewResolver(api)

	_, err := resolver.Resolve(context.Background(), "NOR", "latest")
	if err == nil {
		t.Fatal("expected API error, got nil")
	}
}
