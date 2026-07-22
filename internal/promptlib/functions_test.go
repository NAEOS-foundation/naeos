package promptlib

import (
	"testing"
)

func TestToJSONFunc(t *testing.T) {
	s, err := toJSONFunc(map[string]any{"a": 1})
	if err != nil {
		t.Fatalf("toJSONFunc: %v", err)
	}
	if s == "" {
		t.Error("expected non-empty json")
	}
}

func TestToJSONFuncError(t *testing.T) {
	_, err := toJSONFunc(make(chan int))
	if err == nil {
		t.Error("expected error for unmarshalable type")
	}
}

func TestToYAMLFunc(t *testing.T) {
	s, err := toYAMLFunc(map[string]any{"a": 1})
	if err != nil {
		t.Fatalf("toYAMLFunc: %v", err)
	}
	if s == "" {
		t.Error("expected non-empty yaml")
	}
}

func TestDefaultFuncNil(t *testing.T) {
	got := defaultFunc("default", nil)
	if got != "default" {
		t.Errorf("expected default, got %v", got)
	}
}

func TestDefaultFuncEmptyString(t *testing.T) {
	got := defaultFunc("default", "")
	if got != "default" {
		t.Errorf("expected default, got %v", got)
	}
}

func TestDefaultFuncNonEmptyString(t *testing.T) {
	got := defaultFunc("default", "val")
	if got != "val" {
		t.Errorf("expected val, got %v", got)
	}
}

func TestDefaultFuncEmptyAnySlice(t *testing.T) {
	got := defaultFunc("default", []any{})
	if got != "default" {
		t.Errorf("expected default, got %v", got)
	}
}

func TestDefaultFuncNonEmptyAnySlice(t *testing.T) {
	got := defaultFunc("default", []any{1})
	if got.([]any)[0] != 1 {
		t.Errorf("expected [1], got %v", got)
	}
}

func TestDefaultFuncEmptyStringSlice(t *testing.T) {
	got := defaultFunc("default", []string{})
	if got != "default" {
		t.Errorf("expected default, got %v", got)
	}
}

func TestDefaultFuncNonEmptyStringSlice(t *testing.T) {
	got := defaultFunc("default", []string{"a"})
	if got.([]string)[0] != "a" {
		t.Errorf("expected [a], got %v", got)
	}
}

func TestDefaultFuncInt(t *testing.T) {
	got := defaultFunc(0, 42)
	if got != 42 {
		t.Errorf("expected 42, got %v", got)
	}
}

func TestLenFuncString(t *testing.T) {
	if lenFunc("hello") != 5 {
		t.Errorf("expected 5, got %d", lenFunc("hello"))
	}
}

func TestLenFuncAnySlice(t *testing.T) {
	if lenFunc([]any{1, 2, 3}) != 3 {
		t.Errorf("expected 3, got %d", lenFunc([]any{1, 2, 3}))
	}
}

func TestLenFuncStringSlice(t *testing.T) {
	if lenFunc([]string{"a", "b"}) != 2 {
		t.Errorf("expected 2, got %d", lenFunc([]string{"a", "b"}))
	}
}

func TestLenFuncMap(t *testing.T) {
	if lenFunc(map[string]any{"a": 1}) != 1 {
		t.Errorf("expected 1, got %d", lenFunc(map[string]any{"a": 1}))
	}
}

func TestLenFuncDefault(t *testing.T) {
	if lenFunc(42) != 0 {
		t.Errorf("expected 0, got %d", lenFunc(42))
	}
}

func TestRangeSeqFunc(t *testing.T) {
	r := rangeSeqFunc(5)
	if len(r) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(r))
	}
	for i, v := range r {
		if v != i+1 {
			t.Errorf("expected %d at index %d, got %d", i+1, i, v)
		}
	}
}

func TestBacktickFunc(t *testing.T) {
	if backtickFunc() != "`" {
		t.Errorf("expected backtick, got %s", backtickFunc())
	}
}

func TestCodeFunc(t *testing.T) {
	if codeFunc("hello") != "`hello`" {
		t.Errorf("expected `hello`, got %s", codeFunc("hello"))
	}
}

func TestTitleFuncEmpty(t *testing.T) {
	if titleFunc("") != "" {
		t.Errorf("expected empty, got %s", titleFunc(""))
	}
}

func TestTitleFunc(t *testing.T) {
	got := titleFunc("hello world")
	want := "Hello World"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestTitleFuncSingle(t *testing.T) {
	got := titleFunc("hello")
	if got != "Hello" {
		t.Errorf("expected Hello, got %s", got)
	}
}
