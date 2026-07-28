package testutil

import "strings"

func Contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && strings.Contains(s, substr))
}

func ContainsBytes(haystack []byte, needle string) bool {
	return Contains(string(haystack), needle)
}
