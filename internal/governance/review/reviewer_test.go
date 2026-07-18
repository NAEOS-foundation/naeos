package review

import (
	"testing"
)

func TestNewReviewer(t *testing.T) {
	r := NewReviewer()
	if r == nil {
		t.Fatal("expected non-nil reviewer")
	}
}

func TestReviewNilInput(t *testing.T) {
	r := NewReviewer()
	err := r.Review(nil)
	if err == nil {
		t.Fatal("expected error for nil input")
	}
}

func TestReviewValidInput(t *testing.T) {
	r := NewReviewer()
	err := r.Review("something")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReviewArtifactEmptyName(t *testing.T) {
	r := NewReviewer()
	_, err := r.ReviewArtifact("", "content", nil)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestReviewArtifactEmptyContent(t *testing.T) {
	r := NewReviewer()
	result, err := r.ReviewArtifact("test.go", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusRejected {
		t.Fatalf("expected rejected status for empty content, got %s", result.Status)
	}
}

func TestReviewArtifactNoTODO(t *testing.T) {
	r := NewReviewer()
	result, err := r.ReviewArtifact("test.go", "package main\n\nfunc main() {}", []string{"no-todo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusApproved {
		t.Fatalf("expected approved status, got %s", result.Status)
	}
}

func TestReviewArtifactWithTODO(t *testing.T) {
	r := NewReviewer()
	content := "package main\n\n// TODO: implement this\nfunc main() {}"
	result, err := r.ReviewArtifact("test.go", content, []string{"no-todo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusChanges {
		t.Fatalf("expected changes_requested status, got %s", result.Status)
	}
	if len(result.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(result.Comments))
	}
}

func TestReviewArtifactNoPlaceholder(t *testing.T) {
	r := NewReviewer()
	content := "package main\n\nfunc main() {}"
	result, err := r.ReviewArtifact("test.go", content, []string{"no-placeholder"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusApproved {
		t.Fatalf("expected approved status, got %s", result.Status)
	}
}

func TestReviewArtifactWithPlaceholder(t *testing.T) {
	r := NewReviewer()
	content := "package main\n\n// FIXME: broken\nfunc main() {}"
	result, err := r.ReviewArtifact("test.go", content, []string{"no-placeholder"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusChanges {
		t.Fatalf("expected changes_requested status, got %s", result.Status)
	}
}

func TestReviewArtifactHasPackageDecl(t *testing.T) {
	r := NewReviewer()
	content := "package main\n\nfunc main() {}"
	result, err := r.ReviewArtifact("test.go", content, []string{"has-package-declaration"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusApproved {
		t.Fatalf("expected approved status, got %s", result.Status)
	}
}

func TestReviewArtifactMissingPackageDecl(t *testing.T) {
	r := NewReviewer()
	content := "func main() {}"
	result, err := r.ReviewArtifact("test.go", content, []string{"has-package-declaration"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusRejected {
		t.Fatalf("expected rejected status, got %s", result.Status)
	}
}

func TestReviewArtifactMultipleRules(t *testing.T) {
	r := NewReviewer()
	content := "// TODO: fix\nfunc main() {}"
	result, err := r.ReviewArtifact("test.go", content, []string{"no-todo", "has-package-declaration"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusRejected {
		t.Fatalf("expected rejected status, got %s", result.Status)
	}
	if len(result.Comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(result.Comments))
	}
}

func TestReviewArtifactDetectsLowercaseTodoAndPlaceholder(t *testing.T) {
	r := NewReviewer()
	content := "package main\n\n// todo: implement this\nfunc main() {}"
	result, err := r.ReviewArtifact("test.go", content, []string{"no-todo", "no-placeholder"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusChanges {
		t.Fatalf("expected changes_requested status, got %s", result.Status)
	}
	if len(result.Comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(result.Comments))
	}
}

func TestReviewArtifactNoRules(t *testing.T) {
	r := NewReviewer()
	content := "package main\n\nfunc main() {}"
	result, err := r.ReviewArtifact("test.go", content, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusApproved {
		t.Fatalf("expected approved status, got %s", result.Status)
	}
}

func TestReviewArtifactMissingLicenseHeader(t *testing.T) {
	r := NewReviewer()
	content := "package main\n\nfunc main() {}"
	result, err := r.ReviewArtifact("test.go", content, []string{"has-license-header"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusChanges {
		t.Fatalf("expected changes_requested status, got %s", result.Status)
	}
	if len(result.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(result.Comments))
	}
}

func TestReviewArtifactHasLicenseHeader(t *testing.T) {
	r := NewReviewer()
	content := "// Copyright 2026 NAEOS Foundation\n// Licensed under the Apache License, Version 2.0 (the \"License\");\npackage main\n\nfunc main() {}"
	result, err := r.ReviewArtifact("test.go", content, []string{"has-license-header"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusApproved {
		t.Fatalf("expected approved status, got %s", result.Status)
	}
}
