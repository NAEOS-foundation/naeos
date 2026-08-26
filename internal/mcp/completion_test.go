package mcp

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/artifacts"
)

func TestInitializeAdvertisesCompletions(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	resp := postMCP(t, s, "initialize", map[string]any{})

	caps, ok := resp["result"].(map[string]any)["capabilities"].(map[string]any)
	if !ok {
		t.Fatal("expected capabilities object")
	}
	if _, ok := caps["completions"]; !ok {
		t.Error("expected completions capability to be advertised")
	}
}

func TestCompletePromptArgumentArchitecture(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  []string
	}{
		{
			name:  "prefix micro",
			value: "micro",
			want:  []string{"microkernel", "microservices"},
		},
		{
			name:  "case-insensitive prefix",
			value: "SERVER",
			want:  []string{"serverless"},
		},
		{
			name:  "empty value returns all",
			value: "",
			want:  []string{"clean", "cqrs", "event-driven", "hexagonal", "layered", "microkernel", "microservices", "monolith", "monolithic", "serverless"},
		},
		{
			name:  "no match",
			value: "quantum",
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer()
			resp := postMCP(t, s, "completion/complete", map[string]any{
				"ref":      map[string]any{"type": "ref/prompt", "name": "explain-architecture"},
				"argument": map[string]any{"name": "architecture", "value": tt.value},
			})

			completion, ok := resp["result"].(map[string]any)["completion"].(map[string]any)
			if !ok {
				t.Fatalf("expected completion object, got %v", resp)
			}
			rawValues, ok := completion["values"].([]any)
			if !ok {
				t.Fatal("expected values array")
			}
			got := make([]string, 0, len(rawValues))
			for _, v := range rawValues {
				got = append(got, v.(string))
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("values = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompletePromptArgumentNoCandidates(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	resp := postMCP(t, s, "completion/complete", map[string]any{
		"ref":      map[string]any{"type": "ref/prompt", "name": "review-spec"},
		"argument": map[string]any{"name": "spec", "value": ""},
	})

	completion, ok := resp["result"].(map[string]any)["completion"].(map[string]any)
	if !ok {
		t.Fatal("expected completion object")
	}
	values, _ := completion["values"].([]any)
	if len(values) != 0 {
		t.Errorf("expected empty values for free-text argument, got %v", values)
	}
}

func TestCompletePromptArgumentErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		params     map[string]any
		errContain string
	}{
		{
			name: "unknown prompt",
			params: map[string]any{
				"ref":      map[string]any{"type": "ref/prompt", "name": "nonexistent"},
				"argument": map[string]any{"name": "spec", "value": ""},
			},
			errContain: "unknown prompt",
		},
		{
			name: "unknown argument",
			params: map[string]any{
				"ref":      map[string]any{"type": "ref/prompt", "name": "review-spec"},
				"argument": map[string]any{"name": "bogus", "value": ""},
			},
			errContain: "has no argument",
		},
		{
			name:       "missing ref type",
			params:     map[string]any{},
			errContain: "invalid params: 'ref.type' is required",
		},
		{
			name: "missing argument name",
			params: map[string]any{
				"ref": map[string]any{"type": "ref/prompt", "name": "review-spec"},
			},
			errContain: "invalid params: 'argument.name' is required",
		},
		{
			name: "unsupported ref type",
			params: map[string]any{
				"ref":      map[string]any{"type": "ref/history"},
				"argument": map[string]any{"name": "x", "value": ""},
			},
			errContain: "unsupported ref type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer()
			resp := postMCP(t, s, "completion/complete", tt.params)

			errObj, ok := resp["error"].(map[string]any)
			if !ok {
				t.Fatalf("expected error object, got %v", resp)
			}
			msg, _ := errObj["message"].(string)
			if !strings.Contains(msg, tt.errContain) {
				t.Errorf("error message = %q, want substring %q", msg, tt.errContain)
			}
		})
	}
}

func TestCompleteResourceURI(t *testing.T) {
	t.Parallel()
	s := newTestServer()
	store := artifacts.NewStore(t.TempDir())
	if _, err := store.Add("main.go", []byte("package main"), artifacts.KindCode); err != nil {
		t.Fatalf("add artifact: %v", err)
	}
	s.SetArtifactStore(store)
	s.TrackPipelineJob(&PipelineJob{ID: "job-1", Status: "running"})

	tests := []struct {
		name  string
		value string
		want  []string
	}{
		{
			name:  "docs concept prefix",
			value: "naeos://docs/mo",
			want:  []string{"naeos://docs/module"},
		},
		{
			name:  "artifact path",
			value: uriSchemeArtifacts + "main",
			want:  []string{uriSchemeArtifacts + "main.go"},
		},
		{
			name:  "job id",
			value: uriSchemeJobs + "job-",
			want:  []string{uriSchemeJobs + "job-1"},
		},
		{
			name:  "no match",
			value: "naeos://nope/",
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := postMCP(t, s, "completion/complete", map[string]any{
				"ref":      map[string]any{"type": "ref/resource", "uri": "naeos://{uri}"},
				"argument": map[string]any{"name": "uri", "value": tt.value},
			})

			completion, ok := resp["result"].(map[string]any)["completion"].(map[string]any)
			if !ok {
				t.Fatalf("expected completion object, got %v", resp)
			}
			rawValues, ok := completion["values"].([]any)
			if !ok {
				t.Fatal("expected values array")
			}
			got := make([]string, 0, len(rawValues))
			for _, v := range rawValues {
				got = append(got, v.(string))
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("values = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterCompletionsCapAndHasMore(t *testing.T) {
	t.Parallel()
	candidates := make([]string, maxCompletionValues+10)
	for i := range candidates {
		candidates[i] = fmt.Sprintf("candidate-%03d", i)
	}

	result := filterCompletions(candidates, "")
	if len(result.Values) != maxCompletionValues {
		t.Errorf("len(values) = %d, want %d", len(result.Values), maxCompletionValues)
	}
	if !result.HasMore {
		t.Error("expected HasMore true when results are capped")
	}
	if result.Total != len(candidates) {
		t.Errorf("total = %d, want %d", result.Total, len(candidates))
	}

	full := filterCompletions(candidates[:3], "")
	if len(full.Values) != 3 || full.HasMore {
		t.Errorf("unexpected uncapped result: %+v", full)
	}
}
