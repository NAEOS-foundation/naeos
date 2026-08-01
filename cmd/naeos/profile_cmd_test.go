package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProfileList(t *testing.T) {
	output := captureOutput(t, func() {
		if err := run([]string{"profile", "list"}); err != nil {
			t.Fatalf("profile list failed: %v", err)
		}
	})
	if !strings.Contains(output, "saas") || !strings.Contains(output, "ID") {
		t.Fatalf("expected profile entries, got %q", output)
	}
}

func TestProfileListByIndustry(t *testing.T) {
	output := captureOutput(t, func() {
		if err := run([]string{"profile", "list", "--industry", "technology"}); err != nil {
			t.Fatalf("profile list --industry failed: %v", err)
		}
	})
	if !strings.Contains(output, "saas") {
		t.Fatalf("expected technology profiles, got %q", output)
	}
}

func TestProfileListUnknownIndustry(t *testing.T) {
	output := captureOutput(t, func() {
		if err := run([]string{"profile", "list", "--industry", "nonsense"}); err != nil {
			t.Fatalf("profile list with unknown industry failed: %v", err)
		}
	})
	if !strings.Contains(output, "No profiles found.") {
		t.Fatalf("expected empty message, got %q", output)
	}
}

func TestProfileShow(t *testing.T) {
	output := captureOutput(t, func() {
		if err := run([]string{"profile", "show", "saas"}); err != nil {
			t.Fatalf("profile show failed: %v", err)
		}
	})
	if !strings.Contains(output, "Profile:") || !strings.Contains(output, "ID: saas") {
		t.Fatalf("expected profile details, got %q", output)
	}
	if !strings.Contains(output, "Architecture:") || !strings.Contains(output, "Security:") {
		t.Fatalf("expected architecture/security sections, got %q", output)
	}
}

func TestProfileShowNotFound(t *testing.T) {
	err := run([]string{"profile", "show", "ghost"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestProfileSearch(t *testing.T) {
	output := captureOutput(t, func() {
		if err := run([]string{"profile", "search", "saas"}); err != nil {
			t.Fatalf("profile search failed: %v", err)
		}
	})
	if !strings.Contains(output, "saas") {
		t.Fatalf("expected search result, got %q", output)
	}
}

func TestProfileSearchNoMatch(t *testing.T) {
	output := captureOutput(t, func() {
		if err := run([]string{"profile", "search", "zzzz-nomatch"}); err != nil {
			t.Fatalf("profile search no match failed: %v", err)
		}
	})
	if !strings.Contains(output, "No profiles match your search.") {
		t.Fatalf("expected no-match message, got %q", output)
	}
}

func TestProfileApply(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "spec.yaml")
	output := captureOutput(t, func() {
		if err := run([]string{"profile", "apply", "saas", "--output", outFile}); err != nil {
			t.Fatalf("profile apply failed: %v", err)
		}
	})
	if !strings.Contains(output, "applied to") {
		t.Fatalf("expected apply message, got %q", output)
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read applied spec: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected spec file to contain content")
	}
}

func TestProfileApplyNotFound(t *testing.T) {
	err := run([]string{"profile", "apply", "ghost", "--output", filepath.Join(t.TempDir(), "spec.yaml")})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestProfileCompare(t *testing.T) {
	output := captureOutput(t, func() {
		if err := run([]string{"profile", "compare", "saas", "fintech"}); err != nil {
			t.Fatalf("profile compare failed: %v", err)
		}
	})
	if !strings.Contains(output, "Comparing:") || !strings.Contains(output, "Industry") {
		t.Fatalf("expected comparison output, got %q", output)
	}
}

func TestProfileCompareFirstNotFound(t *testing.T) {
	err := run([]string{"profile", "compare", "ghost", "fintech"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestProfileCompareSecondNotFound(t *testing.T) {
	err := run([]string{"profile", "compare", "saas", "ghost"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestProfileCategories(t *testing.T) {
	output := captureOutput(t, func() {
		if err := run([]string{"profile", "categories"}); err != nil {
			t.Fatalf("profile categories failed: %v", err)
		}
	})
	if !strings.Contains(output, "Profile Categories:") || !strings.Contains(output, "profiles") {
		t.Fatalf("expected categories output, got %q", output)
	}
}
