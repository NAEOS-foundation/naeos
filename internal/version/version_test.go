package version

import (
	"testing"
)

func TestString(t *testing.T) {
	v := String()
	if v == "" {
		t.Error("expected non-empty version string")
	}
	if v != "0.7.0" {
		t.Errorf("expected version 0.7.0, got %s", v)
	}
}

func TestFull(t *testing.T) {
	full := Full()
	if full == "" {
		t.Error("expected non-empty full version")
	}
}

func TestFullWithCommit(t *testing.T) {
	saved := GitCommit
	GitCommit = "abc123"
	defer func() { GitCommit = saved }()

	full := Full()
	expected := "0.7.0 (abc123)"
	if full != expected {
		t.Errorf("expected %s, got %s", expected, full)
	}
}
