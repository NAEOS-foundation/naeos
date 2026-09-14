// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package strutil

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"  spaces  ", "spaces"},
		{"with/slashes", "with-slashes"},
		{"with_underscores", "with-underscores"},
		{"UPPER CASE", "upper-case"},
		{"special!@#$chars", "specialchars"},
		{"multiple---dashes", "multiple-dashes"},
		{"", ""},
		{"--leading-trailing--", "leading-trailing"},
	}

	for _, tt := range tests {
		got := Slugify(tt.input)
		if got != tt.want {
			t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCamelCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "helloWorld"},
		{"Hello World", "helloWorld"},
		{"hello-world", "helloWorld"},
		{"hello_world", "helloWorld"},
		{"helloWorld", "helloWorld"},
		{"HTMLParser", "htmlParser"},
		{"XMLHttpRequest", "xmlHttpRequest"},
		{"", ""},
		{"single", "single"},
	}

	for _, tt := range tests {
		got := CamelCase(tt.input)
		if got != tt.want {
			t.Errorf("CamelCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "HelloWorld"},
		{"hello-world", "HelloWorld"},
		{"hello_world", "HelloWorld"},
		{"", ""},
	}

	for _, tt := range tests {
		got := PascalCase(tt.input)
		if got != tt.want {
			t.Errorf("PascalCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "hello_world"},
		{"Hello World", "hello_world"},
		{"helloWorld", "hello_world"},
		{"hello-world", "hello_world"},
		{"", ""},
	}

	for _, tt := range tests {
		got := SnakeCase(tt.input)
		if got != tt.want {
			t.Errorf("SnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestKebabCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "hello-world"},
		{"Hello World", "hello-world"},
		{"helloWorld", "hello-world"},
		{"hello_world", "hello-world"},
		{"", ""},
	}

	for _, tt := range tests {
		got := KebabCase(tt.input)
		if got != tt.want {
			t.Errorf("KebabCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "he..."},
		{"hi", 5, "hi"},
		{"hello", 3, "hel"},
		{"", 5, ""},
	}

	for _, tt := range tests {
		got := Truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

func TestTruncateRunes(t *testing.T) {
	input := "héllo wörld"
	got := TruncateRunes(input, 6)
	want := "hél..."
	if got != want {
		t.Errorf("TruncateRunes(%q, 6) = %q, want %q", input, got, want)
	}
}

func TestIsValidSlug(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"hello-world", true},
		{"hello", true},
		{"123", true},
		{"hello-world-123", true},
		{"Hello-World", false},
		{"hello world", false},
		{"hello_world", false},
		{"", false},
		{"-hello", false},
		{"hello-", false},
		{"hello--world", false},
	}

	for _, tt := range tests {
		got := IsValidSlug(tt.input)
		if got != tt.want {
			t.Errorf("IsValidSlug(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"hello", true},
		{"_hello", true},
		{"hello123", true},
		{"hello_world", true},
		{"123hello", false},
		{"hello-world", false},
		{"", false},
		{"hello world", false},
	}

	for _, tt := range tests {
		got := IsValidIdentifier(tt.input)
		if got != tt.want {
			t.Errorf("IsValidIdentifier(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestContainsAny(t *testing.T) {
	if !ContainsAny("hello world", "world", "foo") {
		t.Error("expected ContainsAny to find 'world'")
	}
	if ContainsAny("hello", "foo", "bar") {
		t.Error("expected ContainsAny to not find")
	}
}

func TestContainsAll(t *testing.T) {
	if !ContainsAll("hello world", "hello", "world") {
		t.Error("expected ContainsAll to find both")
	}
	if ContainsAll("hello world", "hello", "foo") {
		t.Error("expected ContainsAll to not find 'foo'")
	}
}

func TestCollapseWhitespace(t *testing.T) {
	got := CollapseWhitespace("  hello   world  ")
	if got != "hello world" {
		t.Errorf("CollapseWhitespace = %q, want %q", got, "hello world")
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "olleh"},
		{"", ""},
		{"a", "a"},
		{"ab", "ba"},
		{"héllo", "olléh"},
	}

	for _, tt := range tests {
		got := Reverse(tt.input)
		if got != tt.want {
			t.Errorf("Reverse(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsBlank(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"", true},
		{"  ", true},
		{"\t\n", true},
		{"hello", false},
		{" hello ", false},
	}

	for _, tt := range tests {
		got := IsBlank(tt.input)
		if got != tt.want {
			t.Errorf("IsBlank(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestDefaultIfBlank(t *testing.T) {
	if got := DefaultIfBlank("", "fallback"); got != "fallback" {
		t.Errorf("DefaultIfBlank('') = %q, want %q", got, "fallback")
	}
	if got := DefaultIfBlank("value", "fallback"); got != "value" {
		t.Errorf("DefaultIfBlank('value') = %q, want %q", got, "value")
	}
}

func TestJoinNonEmpty(t *testing.T) {
	got := JoinNonEmpty(", ", "a", "", "b", "", "c")
	if got != "a, b, c" {
		t.Errorf("JoinNonEmpty = %q, want %q", got, "a, b, c")
	}
}

func TestPadLeft(t *testing.T) {
	got := PadLeft("hi", 5, '0')
	if got != "000hi" {
		t.Errorf("PadLeft = %q, want %q", got, "000hi")
	}
	if got := PadLeft("hello", 3, '0'); got != "hello" {
		t.Errorf("PadLeft short = %q, want %q", got, "hello")
	}
}

func TestPadRight(t *testing.T) {
	got := PadRight("hi", 5, '0')
	if got != "hi000" {
		t.Errorf("PadRight = %q, want %q", got, "hi000")
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"cat", "cats"},
		{"dog", "dogs"},
		{"city", "cities"},
		{"day", "days"},
		{"box", "boxes"},
		{"church", "churches"},
		{"wish", "wishes"},
		{"bus", "buses"},
		{"quiz", "quizzes"},
		{"", ""},
	}

	for _, tt := range tests {
		got := Pluralize(tt.input)
		if got != tt.want {
			t.Errorf("Pluralize(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIndent(t *testing.T) {
	got := Indent("hello\nworld", "  ")
	want := "  hello\n  world"
	if got != want {
		t.Errorf("Indent = %q, want %q", got, want)
	}
}

func TestWrap(t *testing.T) {
	got := Wrap("hello world foo bar", 10)
	if len(got) == 0 {
		t.Error("expected non-empty wrap")
	}
}

func TestDirName(t *testing.T) {
	if got := DirName("/foo/bar/baz.txt"); got != "/foo/bar" {
		t.Errorf("DirName = %q, want %q", got, "/foo/bar")
	}
}

func TestBaseName(t *testing.T) {
	if got := BaseName("/foo/bar/baz.txt"); got != "baz.txt" {
		t.Errorf("BaseName = %q, want %q", got, "baz.txt")
	}
}

func TestExt(t *testing.T) {
	if got := Ext("file.txt"); got != ".txt" {
		t.Errorf("Ext = %q, want %q", got, ".txt")
	}
	if got := Ext("file"); got != "" {
		t.Errorf("Ext = %q, want empty", got)
	}
}

func TestReplaceExt(t *testing.T) {
	tests := []struct {
		input string
		ext   string
		want  string
	}{
		{"file.txt", ".go", "file.go"},
		{"file", ".txt", "file.txt"},
		{"file.old.new", ".new", "file.old.new"},
	}

	for _, tt := range tests {
		got := ReplaceExt(tt.input, tt.ext)
		if got != tt.want {
			t.Errorf("ReplaceExt(%q, %q) = %q, want %q", tt.input, tt.ext, got, tt.want)
		}
	}
}

func TestQuote(t *testing.T) {
	if got := Quote("hello"); got != `"hello"` {
		t.Errorf("Quote = %q, want %q", got, `"hello"`)
	}
}

func TestUnquote(t *testing.T) {
	if got := Unquote(`"hello"`); got != "hello" {
		t.Errorf("Unquote = %q, want %q", got, "hello")
	}
	if got := Unquote("hello"); got != "hello" {
		t.Errorf("Unquote = %q, want %q", got, "hello")
	}
}

func TestRepeat(t *testing.T) {
	got := Repeat("ab", 3, ",")
	if got != "ab,ab,ab" {
		t.Errorf("Repeat = %q, want %q", got, "ab,ab,ab")
	}
	if got := Repeat("x", 0, ","); got != "" {
		t.Errorf("Repeat(0) = %q, want empty", got)
	}
}

func TestRandomString(t *testing.T) {
	s1 := RandomString(16)
	s2 := RandomString(16)

	if len(s1) != 16 {
		t.Errorf("expected length 16, got %d", len(s1))
	}
	if s1 == s2 {
		t.Error("expected different random strings")
	}
}
