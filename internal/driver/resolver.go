package driver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var numberPattern = regexp.MustCompile(`^\d+$`)

type QueryAPI interface {
	Query(ctx context.Context, endpoint string, params map[string]string, filters []string) ([]map[string]any, error)
}

type Resolver struct {
	api   QueryAPI
	cache map[string]string
	mu    sync.Mutex
}

func NewResolver(api QueryAPI) *Resolver {
	return &Resolver{
		api:   api,
		cache: make(map[string]string),
	}
}

func (r *Resolver) Resolve(ctx context.Context, driverInput, sessionKey string) (string, error) {
	input := strings.TrimSpace(driverInput)
	if input == "" {
		return "", errors.New("driver value is required")
	}

	if numberPattern.MatchString(input) {
		return input, nil
	}

	acronym := strings.ToUpper(input)
	if len(acronym) != 3 {
		return "", fmt.Errorf("driver %q must be a number or 3-letter acronym", driverInput)
	}

	session := strings.TrimSpace(sessionKey)
	if session == "" {
		session = "latest"
	}

	cacheKey := session + "|" + acronym

	r.mu.Lock()
	if resolved, ok := r.cache[cacheKey]; ok {
		r.mu.Unlock()
		return resolved, nil
	}
	r.mu.Unlock()

	params := map[string]string{
		"session_key":  session,
		"name_acronym": acronym,
		"limit":        "1",
	}

	records, err := r.api.Query(ctx, "drivers", params, nil)
	if err != nil {
		return "", fmt.Errorf("resolve driver %q: %w", acronym, err)
	}
	if len(records) == 0 {
		return "", fmt.Errorf("driver acronym %q not found", acronym)
	}

	driverNumber, ok := toDriverNumber(records[0]["driver_number"])
	if !ok || driverNumber == "" {
		return "", fmt.Errorf("driver acronym %q returned no driver_number", acronym)
	}

	r.mu.Lock()
	r.cache[cacheKey] = driverNumber
	r.mu.Unlock()

	return driverNumber, nil
}

func toDriverNumber(v any) (string, bool) {
	switch n := v.(type) {
	case string:
		out := strings.TrimSpace(n)
		return out, out != ""
	case float64:
		return strconv.Itoa(int(n)), true
	case float32:
		return strconv.Itoa(int(n)), true
	case int:
		return strconv.Itoa(n), true
	case int64:
		return strconv.FormatInt(n, 10), true
	case int32:
		return strconv.FormatInt(int64(n), 10), true
	case json.Number:
		out := strings.TrimSpace(n.String())
		return out, out != ""
	default:
		return "", false
	}
}
