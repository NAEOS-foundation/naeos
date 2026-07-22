package graphql

import (
	"testing"
)

func TestGraphQLError(t *testing.T) {
	err := &GraphQLError{Message: "test error"}
	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got %q", err.Error())
	}
}

func TestMapResolver(t *testing.T) {
	resolver := MapResolver(map[string]any{"key": "val"})
	result, err := resolver(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]any)
	if m["key"] != "val" {
		t.Error("unexpected value")
	}
}

func TestListResolver(t *testing.T) {
	resolver := ListResolver([]any{"a", "b"})
	result, err := resolver(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	items := result.([]any)
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestExtractName(t *testing.T) {
	if got := extractName("foo(bar: 1)"); got != "foo" {
		t.Errorf("expected 'foo', got %q", got)
	}
	if got := extractName("baz"); got != "baz" {
		t.Errorf("expected 'baz', got %q", got)
	}
}

func TestParseFragment(t *testing.T) {
	frag, err := parseFragment("fragment myFrag on User { name age }")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if frag.Name != "myFrag" {
		t.Errorf("expected 'myFrag', got %q", frag.Name)
	}
	if frag.OnType != "User" {
		t.Errorf("expected 'User', got %q", frag.OnType)
	}
	if len(frag.Selections) != 2 {
		t.Errorf("expected 2 selections, got %d", len(frag.Selections))
	}
}

func TestParseFragmentInvalid(t *testing.T) {
	_, err := parseFragment("invalid")
	if err == nil {
		t.Error("expected error for invalid fragment syntax")
	}
}

func TestParseFragmentMissingOn(t *testing.T) {
	_, err := parseFragment("fragment foo { bar }")
	if err == nil {
		t.Error("expected error for missing 'on' type")
	}
}

func TestToMapStringString(t *testing.T) {
	m, ok := toMap(map[string]string{"key": "val"})
	if !ok {
		t.Fatal("expected ok")
	}
	if m["key"] != "val" {
		t.Errorf("expected 'val', got %v", m["key"])
	}
}

func TestToMapUnsupported(t *testing.T) {
	_, ok := toMap([]string{"a"})
	if ok {
		t.Error("expected false for unsupported type")
	}
}

func TestParseQueryEmpty(t *testing.T) {
	ast, errs := ParseQuery("{}")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if ast == nil {
		t.Error("expected non-nil AST")
	}
}
