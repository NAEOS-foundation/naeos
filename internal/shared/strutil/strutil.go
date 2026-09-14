// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package strutil

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path"
	"regexp"
	"strings"
	"unicode"
)

func Slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, "_", "-")
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	value = reg.ReplaceAllString(value, "")
	reg2 := regexp.MustCompile(`-{2,}`)
	value = reg2.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}

func CamelCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	words := splitWords(s)
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(strings.ToLower(words[0]))
	for _, w := range words[1:] {
		if len(w) > 0 {
			b.WriteString(strings.ToUpper(w[:1]))
			b.WriteString(strings.ToLower(w[1:]))
		}
	}
	return b.String()
}

func PascalCase(s string) string {
	s = CamelCase(s)
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func SnakeCase(s string) string {
	words := splitWords(s)
	return strings.Join(words, "_")
}

func KebabCase(s string) string {
	words := splitWords(s)
	return strings.Join(words, "-")
}

func splitWords(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	var words []string
	var current strings.Builder

	for i, r := range s {
		if r == '_' || r == '-' || r == ' ' || r == '.' {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
			continue
		}

		if unicode.IsUpper(r) && current.Len() > 0 {
			prev := rune(s[i-1])
			if unicode.IsLower(prev) || (unicode.IsDigit(prev) && !unicode.IsDigit(r)) {
				words = append(words, current.String())
				current.Reset()
			} else if unicode.IsUpper(prev) && i+1 < len(s) {
				next := rune(s[i+1])
				if unicode.IsLower(next) {
					words = append(words, current.String())
					current.Reset()
				}
			}
		}
		current.WriteRune(r)
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}

	var result []string
	for _, w := range words {
		lower := strings.ToLower(w)
		if lower != "" {
			result = append(result, lower)
		}
	}
	return result
}

func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func TruncateRunes(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}

func IsValidSlug(s string) bool {
	if s == "" {
		return false
	}
	reg := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	return reg.MatchString(s)
}

func IsValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	reg := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return reg.MatchString(s)
}

func ContainsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func ContainsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}

func HasPrefixes(s string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func HasSuffixes(s string, suffixes ...string) bool {
	for _, s2 := range suffixes {
		if strings.HasSuffix(s, s2) {
			return true
		}
	}
	return false
}

func CollapseWhitespace(s string) string {
	reg := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(reg.ReplaceAllString(s, " "))
}

func RemovePrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

func RemoveSuffix(s, suffix string) string {
	return strings.TrimSuffix(s, suffix)
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func DefaultIfBlank(s, def string) string {
	if IsBlank(s) {
		return def
	}
	return s
}

func JoinNonEmpty(sep string, parts ...string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, sep)
}

func PadLeft(s string, length int, pad rune) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(string(pad), length-len(s)) + s
}

func PadRight(s string, length int, pad rune) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(string(pad), length-len(s))
}

func RandomString(length int) string {
	b := make([]byte, length/2+1)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}

func Pluralize(s string) string {
	if s == "" {
		return s
	}
	lower := strings.ToLower(s)
	switch {
	case strings.HasSuffix(lower, "quiz"):
		return s + "zes"
	case strings.HasSuffix(lower, "y") && !containsVowelBefore(s, len(s)-1):
		return s[:len(s)-1] + "ies"
	case strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") || strings.HasSuffix(lower, "z") ||
		strings.HasSuffix(lower, "ch") || strings.HasSuffix(lower, "sh"):
		return s + "es"
	default:
		return s + "s"
	}
}

func containsVowelBefore(s string, idx int) bool {
	if idx <= 0 {
		return false
	}
	r := unicode.ToLower(rune(s[idx-1]))
	return r == 'a' || r == 'e' || r == 'i' || r == 'o' || r == 'u'
}

func Indent(s string, prefix string) string {
	lines := strings.Split(s, "\n")
	var b strings.Builder
	for i, line := range lines {
		if i > 0 {
			b.WriteString("\n")
		}
		if strings.TrimSpace(line) != "" {
			b.WriteString(prefix)
		}
		b.WriteString(line)
	}
	return b.String()
}

func Wrap(s string, width int) string {
	if width <= 0 || len(s) <= width {
		return s
	}

	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}

	var lines []string
	var current strings.Builder

	for _, word := range words {
		if current.Len()+len(word)+1 > width && current.Len() > 0 {
			lines = append(lines, current.String())
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(word)
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	return strings.Join(lines, "\n")
}

func DirName(p string) string {
	return path.Dir(p)
}

func BaseName(p string) string {
	return path.Base(p)
}

func Ext(p string) string {
	return path.Ext(p)
}

func ReplaceExt(p, newExt string) string {
	ext := path.Ext(p)
	if ext == "" {
		return p + newExt
	}
	return p[:len(p)-len(ext)] + newExt
}

func Quote(s string) string {
	return fmt.Sprintf("%q", s)
}

func Unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func Repeat(s string, n int, sep string) string {
	if n <= 0 {
		return ""
	}
	parts := make([]string, n)
	for i := range parts {
		parts[i] = s
	}
	return strings.Join(parts, sep)
}

func IndexOf(s, substr string) int {
	return strings.Index(s, substr)
}

func Count(s, substr string) int {
	return strings.Count(s, substr)
}

func ReplaceN(s, old, new string, n int) string {
	return strings.Replace(s, old, new, n)
}
