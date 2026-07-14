package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestFormatOutputJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "test", "value": "ok"}

	err := FormatOutput(&buf, data, "json")
	if err != nil {
		t.Fatalf("FormatOutput json returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"name": "test"`) {
		t.Fatalf("expected JSON output, got %q", out)
	}
	if !strings.Contains(out, `"value": "ok"`) {
		t.Fatalf("expected JSON output with value, got %q", out)
	}
}

func TestFormatOutputYAML(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "test", "value": "ok"}

	err := FormatOutput(&buf, data, "yaml")
	if err != nil {
		t.Fatalf("FormatOutput yaml returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "name: test") {
		t.Fatalf("expected YAML output, got %q", out)
	}
	if !strings.Contains(out, "value: ok") {
		t.Fatalf("expected YAML output with value, got %q", out)
	}
}

func TestFormatOutputTableWithTableFormatter(t *testing.T) {
	var buf bytes.Buffer
	data := &tableFormatterStub{
		headers: []string{"Name", "Status"},
		rows:    [][]string{{"svc-a", "running"}, {"svc-b", "stopped"}},
	}

	err := FormatOutput(&buf, data, "table")
	if err != nil {
		t.Fatalf("FormatOutput table returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Name") || !strings.Contains(out, "Status") {
		t.Fatalf("expected table headers, got %q", out)
	}
	if !strings.Contains(out, "svc-a") || !strings.Contains(out, "running") {
		t.Fatalf("expected table row data, got %q", out)
	}
}

func TestFormatOutputTableFallback(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "val"}

	err := FormatOutput(&buf, data, "table")
	if err != nil {
		t.Fatalf("FormatOutput table fallback returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"key": "val"`) {
		t.Fatalf("expected JSON fallback for table format, got %q", out)
	}
}

func TestFormatOutputUnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "val"}

	err := FormatOutput(&buf, data, "xml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported output format") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestFormatTableBasic(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"Col1", "Col2"}
	rows := [][]string{{"a", "b"}, {"c", "d"}}

	err := FormatTable(&buf, headers, rows)
	if err != nil {
		t.Fatalf("FormatTable returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Col1") || !strings.Contains(out, "Col2") {
		t.Fatalf("expected headers in output, got %q", out)
	}
	if !strings.Contains(out, "| a    |") || !strings.Contains(out, "| c    |") {
		t.Fatalf("expected padded row data in output, got %q", out)
	}
}

func TestFormatTableEmptyHeaders(t *testing.T) {
	var buf bytes.Buffer
	err := FormatTable(&buf, nil, nil)
	if err != nil {
		t.Fatalf("FormatTable with empty headers returned error: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected empty output, got %q", buf.String())
	}
}

func TestFormatTableRowFewerColumns(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"A", "B", "C"}
	rows := [][]string{{"1", "2"}}

	err := FormatTable(&buf, headers, rows)
	if err != nil {
		t.Fatalf("FormatTable returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "A") || !strings.Contains(out, "B") || !strings.Contains(out, "C") {
		t.Fatalf("expected all headers, got %q", out)
	}
}

type tableFormatterStub struct {
	headers []string
	rows    [][]string
}

func (t *tableFormatterStub) FormatTable(w io.Writer) error {
	return FormatTable(w, t.headers, t.rows)
}
