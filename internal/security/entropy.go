// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package security

import (
	"math"
	"strings"
)

func shannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	freq := make(map[rune]float64)
	for _, c := range s {
		freq[c]++
	}

	length := float64(len([]rune(s)))
	entropy := 0.0
	for _, count := range freq {
		p := count / length
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}

func isHighEntropySecret(s string, threshold float64) bool {
	trimmed := strings.TrimSpace(s)
	if len(trimmed) < 8 {
		return false
	}
	return shannonEntropy(trimmed) >= threshold
}
