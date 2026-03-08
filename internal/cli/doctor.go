package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/barronlroth/f1-cli/internal/output"
	"github.com/spf13/cobra"
)

type doctorReport struct {
	Version       string            `json:"version"`
	APIReachable  bool              `json:"api_reachable"`
	RateLimit     map[string]string `json:"rate_limit,omitempty"`
	LatestSession map[string]any    `json:"latest_session,omitempty"`
}

func (a *app) newDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check API connectivity and report runtime status",
		Example: strings.Join([]string{
			"f1 doctor",
			"f1 doctor --json",
		}, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := a.selectedFormat()
			if err != nil {
				return err
			}
			if format == formatCSV {
				return fmt.Errorf("doctor does not support --csv")
			}

			resp, err := a.client.GetRaw(cmd.Context(), "sessions", map[string]string{
				"session_key": "latest",
				"limit":       "1",
			}, nil)
			if err != nil {
				return fmt.Errorf("api connectivity check failed: %w", err)
			}

			latestSession := map[string]any{}
			var sessions []map[string]any
			if err := json.Unmarshal(resp.Body, &sessions); err == nil && len(sessions) > 0 {
				latestSession = sessions[0]
			}

			report := doctorReport{
				Version:       a.version,
				APIReachable:  true,
				RateLimit:     extractRateLimit(resp.Header),
				LatestSession: latestSession,
			}

			if format == formatJSON {
				raw, err := json.Marshal(report)
				if err != nil {
					return err
				}
				return output.WriteJSON(a.out, raw)
			}

			return writeDoctorText(a, report)
		},
	}
}

func writeDoctorText(a *app, report doctorReport) error {
	_, err := fmt.Fprintf(a.out, "Version: %s\nAPI Reachable: %t\n", report.Version, report.APIReachable)
	if err != nil {
		return err
	}

	if len(report.RateLimit) > 0 {
		for _, key := range []string{"limit", "remaining", "reset", "retry_after"} {
			if value, ok := report.RateLimit[key]; ok {
				if _, err := fmt.Fprintf(a.out, "Rate %s: %s\n", key, value); err != nil {
					return err
				}
			}
		}
	}

	if len(report.LatestSession) > 0 {
		key := valueString(report.LatestSession["session_key"])
		name := valueString(report.LatestSession["session_name"])
		meeting := valueString(report.LatestSession["meeting_name"])
		start := valueString(report.LatestSession["date_start"])

		if _, err := fmt.Fprintf(a.out, "Latest Session: %s %s (%s)\n", key, name, meeting); err != nil {
			return err
		}
		if start != "" {
			if _, err := fmt.Fprintf(a.out, "Latest Session Start: %s\n", start); err != nil {
				return err
			}
		}
	}

	return nil
}

func extractRateLimit(header http.Header) map[string]string {
	out := map[string]string{}

	if v := firstHeader(header, "X-Ratelimit-Limit", "RateLimit-Limit"); v != "" {
		out["limit"] = v
	}
	if v := firstHeader(header, "X-Ratelimit-Remaining", "RateLimit-Remaining"); v != "" {
		out["remaining"] = v
	}
	if v := firstHeader(header, "X-Ratelimit-Reset", "RateLimit-Reset"); v != "" {
		out["reset"] = v
	}
	if v := firstHeader(header, "Retry-After"); v != "" {
		out["retry_after"] = v
	}

	return out
}

func firstHeader(header http.Header, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(header.Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func valueString(v any) string {
	switch typed := v.(type) {
	case string:
		return typed
	case float64:
		return fmt.Sprintf("%.0f", typed)
	default:
		if typed == nil {
			return ""
		}
		return fmt.Sprintf("%v", typed)
	}
}
