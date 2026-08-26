package mcp

import (
	"fmt"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/artifacts"
)

func TestPing(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	resp := postMCP(t, s, "ping", nil)

	if _, ok := resp["result"].(map[string]any); !ok {
		t.Fatalf("expected empty result object, got %v", resp)
	}
	if _, ok := resp["error"]; ok {
		t.Errorf("unexpected error: %v", resp["error"])
	}
}

func TestPaginationWalkAllResources(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	store := artifacts.NewStore(t.TempDir())
	const extraArtifacts = defaultPageSize + 5
	for i := 0; i < extraArtifacts; i++ {
		name := fmt.Sprintf("file%03d.txt", i)
		if _, err := store.Add(name, []byte("content-"+name), artifacts.KindDocs); err != nil {
			t.Fatalf("add artifact %s: %v", name, err)
		}
	}
	s.SetArtifactStore(store)

	seen := make(map[string]bool)
	cursor := ""
	pages := 0
	for {
		params := map[string]any{}
		if cursor != "" {
			params["cursor"] = cursor
		}
		resp := postMCP(t, s, "resources/list", params)
		result, ok := resp["result"].(map[string]any)
		if !ok {
			t.Fatalf("expected result object, got %v (page %d)", resp, pages)
		}
		rawList, ok := result["resources"].([]any)
		if !ok {
			t.Fatalf("expected resources array on page %d", pages)
		}
		for _, raw := range rawList {
			r := raw.(map[string]any)
			uri, _ := r["uri"].(string)
			if seen[uri] {
				t.Errorf("duplicate resource %q across pages", uri)
			}
			seen[uri] = true
		}
		if len(rawList) > defaultPageSize {
			t.Errorf("page %d has %d items, want <= %d", pages, len(rawList), defaultPageSize)
		}
		next, _ := result["nextCursor"].(string)
		pages++
		if next == "" {
			break
		}
		if pages > 10 {
			t.Fatal("pagination did not terminate")
		}
		cursor = next
	}

	wantTotal := len(conceptNames()) + extraArtifacts
	if len(seen) != wantTotal {
		t.Errorf("collected %d unique resources across %d pages, want %d", len(seen), pages, wantTotal)
	}
}

func TestPaginationSmallListsSinglePage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		key    string
		want   int
	}{
		{name: "tools", method: "tools/list", key: "tools", want: len(newTestServer().listTools())},
		{name: "prompts", method: "prompts/list", key: "prompts", want: len(builtinPrompts())},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer()
			resp := postMCP(t, s, tt.method, nil)

			result, ok := resp["result"].(map[string]any)
			if !ok {
				t.Fatalf("expected result object, got %v", resp)
			}
			rawList, ok := result[tt.key].([]any)
			if !ok {
				t.Fatalf("expected %q array", tt.key)
			}
			if len(rawList) != tt.want {
				t.Errorf("len(%s) = %d, want %d", tt.key, len(rawList), tt.want)
			}
			if _, hasMore := result["nextCursor"]; hasMore {
				t.Error("unexpected nextCursor for list smaller than page size")
			}
		})
	}
}

func TestPaginationInvalidCursor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
	}{
		{name: "tools", method: "tools/list"},
		{name: "resources", method: "resources/list"},
		{name: "prompts", method: "prompts/list"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer()
			resp := postMCP(t, s, tt.method, map[string]any{"cursor": "@@not-a-cursor@@"})

			errObj, ok := resp["error"].(map[string]any)
			if !ok {
				t.Fatalf("expected error object, got %v", resp)
			}
			msg, _ := errObj["message"].(string)
			if !strings.Contains(msg, "invalid params") {
				t.Errorf("error message = %q, want invalid params", msg)
			}
		})
	}
}

func TestEncodeDecodeCursorRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		offset int
	}{
		{name: "zero", offset: 0},
		{name: "positive", offset: 42},
		{name: "large", offset: 1 << 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := decodeCursor(encodeCursor(tt.offset))
			if err != nil {
				t.Fatalf("decodeCursor(encodeCursor(%d)): %v", tt.offset, err)
			}
			if got != tt.offset {
				t.Errorf("round trip = %d, want %d", got, tt.offset)
			}
		})
	}

	if _, err := decodeCursor("###"); err == nil {
		t.Error("expected error for malformed cursor")
	}
	if _, err := decodeCursor("-"); err == nil {
		t.Error("expected error for negative offset cursor")
	}
}
