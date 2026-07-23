package testutil

import "testing"

func TestContains(t *testing.T) {
	if !Contains("hello world", "world") {
		t.Error("expected substring match")
	}
	if Contains("hello", "xyz") {
		t.Error("expected no match")
	}
	if !Contains("", "") {
		t.Error("expected empty match")
	}
}

func TestContainsBytes(t *testing.T) {
	if !ContainsBytes([]byte("hello world"), "world") {
		t.Error("expected substring match")
	}
	if ContainsBytes([]byte("hello"), "xyz") {
		t.Error("expected no match")
	}
}
