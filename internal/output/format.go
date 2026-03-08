package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

func WriteJSON(w io.Writer, body []byte) error {
	var out bytes.Buffer
	if err := json.Indent(&out, body, "", "  "); err != nil {
		return fmt.Errorf("format JSON output: %w", err)
	}
	out.WriteByte('\n')
	_, err := w.Write(out.Bytes())
	return err
}

func WriteCSV(w io.Writer, body []byte) error {
	if len(body) == 0 {
		return nil
	}

	if _, err := w.Write(body); err != nil {
		return err
	}

	if body[len(body)-1] != '\n' {
		_, err := io.WriteString(w, "\n")
		return err
	}
	return nil
}

func WriteTable(w io.Writer, records []map[string]any) error {
	if len(records) == 0 {
		_, err := io.WriteString(w, "No results.\n")
		return err
	}

	headers := collectHeaders(records)
	if len(headers) == 0 {
		_, err := io.WriteString(w, "No results.\n")
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for i, header := range headers {
		if i > 0 {
			_, _ = io.WriteString(tw, "\t")
		}
		_, _ = io.WriteString(tw, prettyHeader(header))
	}
	_, _ = io.WriteString(tw, "\n")

	for _, record := range records {
		for i, header := range headers {
			if i > 0 {
				_, _ = io.WriteString(tw, "\t")
			}
			_, _ = io.WriteString(tw, formatValue(record[header]))
		}
		_, _ = io.WriteString(tw, "\n")
	}

	return tw.Flush()
}

func collectHeaders(records []map[string]any) []string {
	seen := make(map[string]struct{})
	for _, record := range records {
		for key := range record {
			seen[key] = struct{}{}
		}
	}

	headers := make([]string, 0, len(seen))
	for key := range seen {
		headers = append(headers, key)
	}
	sort.Strings(headers)
	return headers
}

func prettyHeader(header string) string {
	header = strings.ReplaceAll(header, "_", " ")
	return strings.ToUpper(header)
}

func formatValue(v any) string {
	switch typed := v.(type) {
	case nil:
		return ""
	case string:
		return typed
	case bool:
		return strconv.FormatBool(typed)
	case float64:
		return formatFloat(typed)
	case float32:
		return formatFloat(float64(typed))
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case int32:
		return strconv.FormatInt(int64(typed), 10)
	default:
		raw, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprintf("%v", typed)
		}
		return string(raw)
	}
}

func formatFloat(f float64) string {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return ""
	}
	if math.Mod(f, 1) == 0 {
		return strconv.FormatInt(int64(f), 10)
	}
	s := strconv.FormatFloat(f, 'f', 3, 64)
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}
