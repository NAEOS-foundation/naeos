// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package interop

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/gowebpki/jcs"
)

const candidateDigestProfileV1 = "naeos-candidate-jcs-sha256-v1"
const candidateDigestDomainV1 = "NAEOS-CANDIDATE-DIGEST-V1\n"

// CandidateDigestJSON validates and hashes a raw candidate envelope using the
// NAEOS JCS/SHA-256 v1 profile. The input must be exactly {"tool": string,
// "input": JSON value}; unknown envelope fields and duplicate object keys fail.
func CandidateDigestJSON(raw []byte) (string, []byte, error) {
	if !utf8.Valid(raw) {
		return "", nil, errors.New("candidate JSON is not valid UTF-8")
	}
	if err := rejectDuplicateKeysAndTrailingJSON(raw); err != nil {
		return "", nil, err
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return "", nil, fmt.Errorf("invalid candidate JSON: %w", err)
	}
	if len(envelope) != 2 {
		return "", nil, errors.New("candidate envelope must contain exactly tool and input")
	}
	toolRaw, ok := envelope["tool"]
	if !ok {
		return "", nil, errors.New("tool is required")
	}
	inputRaw, ok := envelope["input"]
	if !ok {
		return "", nil, errors.New("input is required")
	}
	var tool string
	if err := json.Unmarshal(toolRaw, &tool); err != nil || strings.TrimSpace(tool) == "" {
		return "", nil, errors.New("tool must be a non-empty string")
	}
	if len(inputRaw) == 0 {
		return "", nil, errors.New("input is required")
	}

	canonical, err := jcs.Transform(raw)
	if err != nil {
		return "", nil, fmt.Errorf("JCS canonicalization failed: %w", err)
	}
	hashInput := make([]byte, 0, len(candidateDigestDomainV1)+len(canonical))
	hashInput = append(hashInput, []byte(candidateDigestDomainV1)...)
	hashInput = append(hashInput, canonical...)
	sum := sha256.Sum256(hashInput)
	return hex.EncodeToString(sum[:]), canonical, nil
}

// rejectDuplicateKeysAndTrailingJSON walks the token stream so duplicate keys
// are detected at every nesting depth before any map-based decoding can erase them.
func rejectDuplicateKeysAndTrailingJSON(raw []byte) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := readJSONValue(dec); err != nil {
		return fmt.Errorf("invalid candidate JSON: %w", err)
	}
	if _, err := dec.Token(); err != io.EOF {
		if err == nil {
			return errors.New("candidate contains trailing JSON")
		}
		return fmt.Errorf("invalid trailing JSON: %w", err)
	}
	return nil
}

func readJSONValue(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	delim, isDelim := tok.(json.Delim)
	if !isDelim {
		return nil
	}
	switch delim {
	case '{':
		seen := make(map[string]struct{})
		for dec.More() {
			keyToken, err := dec.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("object key is not a string")
			}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("duplicate object key %q", key)
			}
			seen[key] = struct{}{}
			if err := readJSONValue(dec); err != nil {
				return err
			}
		}
		end, err := dec.Token()
		if err != nil {
			return err
		}
		if end != json.Delim('}') {
			return errors.New("unterminated object")
		}
	case '[':
		for dec.More() {
			if err := readJSONValue(dec); err != nil {
				return err
			}
		}
		end, err := dec.Token()
		if err != nil {
			return err
		}
		if end != json.Delim(']') {
			return errors.New("unterminated array")
		}
	default:
		return fmt.Errorf("unexpected delimiter %q", delim)
	}
	return nil
}

// CandidateDigestProfileV1 returns the stable profile identifier to carry with
// a digest when communicating across runtimes.
func CandidateDigestProfileV1() string { return candidateDigestProfileV1 }

// Keep the helper referenced in this package so callers cannot accidentally
// accept non-finite float values when constructing a typed Candidate.
func validateJCSValue(v any) error {
	switch x := v.(type) {
	case string:
		if !utf8.ValidString(x) {
			return errors.New("string is not valid UTF-8")
		}
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return errors.New("non-finite number is not valid JSON")
		}
	case float32:
		if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) {
			return errors.New("non-finite number is not valid JSON")
		}
	case map[string]any:
		for k, value := range x {
			if err := validateJCSValue(k); err != nil { return err }
			if err := validateJCSValue(value); err != nil { return err }
		}
	case []any:
		for _, value := range x {
			if err := validateJCSValue(value); err != nil { return err }
		}
	}
	return nil
}
