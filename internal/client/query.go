package client

import (
	"errors"
	"net/url"
	"path"
	"strings"
)

// BuildURL creates a request URL for an endpoint with standard key/value params
// plus raw OpenF1 filter expressions (for example: speed>=300).
func BuildURL(baseURL, endpoint string, params map[string]string, filters []string) (string, error) {
	if strings.TrimSpace(baseURL) == "" {
		return "", errors.New("base URL is required")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	cleanEndpoint := strings.TrimPrefix(strings.TrimSpace(endpoint), "/")
	if cleanEndpoint != "" {
		u.Path = path.Join(strings.TrimSuffix(u.Path, "/"), cleanEndpoint)
	}

	values := url.Values{}
	for k, v := range params {
		key := strings.TrimSpace(k)
		val := strings.TrimSpace(v)
		if key == "" || val == "" {
			continue
		}
		values.Set(key, val)
	}

	encoded := values.Encode()
	rawFilters := make([]string, 0, len(filters))
	for _, filter := range filters {
		f := strings.TrimSpace(filter)
		if f == "" {
			continue
		}
		rawFilters = append(rawFilters, url.QueryEscape(f))
	}

	switch {
	case encoded == "" && len(rawFilters) == 0:
		u.RawQuery = ""
	case encoded == "":
		u.RawQuery = strings.Join(rawFilters, "&")
	case len(rawFilters) == 0:
		u.RawQuery = encoded
	default:
		u.RawQuery = encoded + "&" + strings.Join(rawFilters, "&")
	}

	return u.String(), nil
}
