package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := WriteJSON(&buf, []byte(`[{"a":1}]`))
	if err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	want := "[\n  {\n    \"a\": 1\n  }\n]\n"
	if buf.String() != want {
		t.Fatalf("unexpected JSON output\nwant:\n%s\ngot:\n%s", want, buf.String())
	}
}

func TestWriteCSVAddsTrailingNewline(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := WriteCSV(&buf, []byte("a,b\n1,2"))
	if err != nil {
		t.Fatalf("WriteCSV returned error: %v", err)
	}

	if got := buf.String(); got != "a,b\n1,2\n" {
		t.Fatalf("unexpected CSV output: %q", got)
	}
}

func TestWriteTable(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	records := []map[string]any{
		{"driver_number": float64(44), "name_acronym": "HAM"},
		{"driver_number": float64(1), "name_acronym": "VER"},
	}

	err := WriteTable(&buf, records)
	if err != nil {
		t.Fatalf("WriteTable returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "DRIVER NUMBER") {
		t.Fatalf("missing DRIVER NUMBER header:\n%s", out)
	}
	if !strings.Contains(out, "NAME ACRONYM") {
		t.Fatalf("missing NAME ACRONYM header:\n%s", out)
	}
	if !strings.Contains(out, "44") || !strings.Contains(out, "HAM") {
		t.Fatalf("missing expected row content:\n%s", out)
	}
}
