// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"unicode/utf8"

	"github.com/gowebpki/jcs"
)

const CandidateDigestProfileV1 = "naeos-candidate-jcs-sha256-v1"
const candidateDigestPrefixV1 = "NAEOS-CANDIDATE-DIGEST-V1\n"

const (
	candidateDigestContextKeyV1  = "naeos_candidate_digest_v1"
	candidateProfileContextKeyV1 = "naeos_candidate_digest_profile"
)

// CandidateDigestV1 binds the exact normalized tool/action arguments using
// RFC 8785 JCS and SHA-256. Request and invocation IDs are evidence identities,
// not candidate arguments; authorization metadata is checked separately.
// cloneCandidateMap makes an isolated, type-preserving snapshot of JSON-like input.
// It recursively copies maps, slices, arrays, pointers, and exported struct fields.
// The depth limit also rejects cyclic values before canonicalization can recurse forever.
func cloneCandidateMap(input map[string]any) (map[string]any, error) {
	if input == nil {
		return nil, nil
	}
	cloned, err := cloneCandidateValue(reflect.ValueOf(input), 0)
	if err != nil {
		return nil, err
	}
	result, ok := cloned.Interface().(map[string]any)
	if !ok {
		return nil, fmt.Errorf("candidate snapshot has unexpected type %T", cloned.Interface())
	}
	return result, nil
}

func cloneCandidateValue(value reflect.Value, depth int) (reflect.Value, error) {
	if depth > 128 {
		return reflect.Value{}, fmt.Errorf("candidate value exceeds snapshot depth limit or contains a cycle")
	}
	if !value.IsValid() {
		return value, nil
	}
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return reflect.Zero(value.Type()), nil
		}
		cloned, err := cloneCandidateValue(value.Elem(), depth+1)
		if err != nil {
			return reflect.Value{}, err
		}
		out := reflect.New(value.Type()).Elem()
		out.Set(cloned)
		return out, nil
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type()), nil
		}
		cloned, err := cloneCandidateValue(value.Elem(), depth+1)
		if err != nil {
			return reflect.Value{}, err
		}
		out := reflect.New(value.Type().Elem())
		out.Elem().Set(cloned)
		return out, nil
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type()), nil
		}
		if value.Type().Key().Kind() != reflect.String {
			return reflect.Value{}, fmt.Errorf("candidate snapshot requires string-keyed maps")
		}
		out := reflect.MakeMapWithSize(value.Type(), value.Len())
		iter := value.MapRange()
		for iter.Next() {
			cloned, err := cloneCandidateValue(iter.Value(), depth+1)
			if err != nil {
				return reflect.Value{}, err
			}
			out.SetMapIndex(iter.Key(), cloned)
		}
		return out, nil
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type()), nil
		}
		out := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for i := 0; i < value.Len(); i++ {
			cloned, err := cloneCandidateValue(value.Index(i), depth+1)
			if err != nil {
				return reflect.Value{}, err
			}
			out.Index(i).Set(cloned)
		}
		return out, nil
	case reflect.Array:
		out := reflect.New(value.Type()).Elem()
		for i := 0; i < value.Len(); i++ {
			cloned, err := cloneCandidateValue(value.Index(i), depth+1)
			if err != nil {
				return reflect.Value{}, err
			}
			out.Index(i).Set(cloned)
		}
		return out, nil
	case reflect.Struct:
		out := reflect.New(value.Type()).Elem()
		out.Set(value)
		for i := 0; i < value.NumField(); i++ {
			if value.Type().Field(i).PkgPath != "" {
				continue
			}
			cloned, err := cloneCandidateValue(value.Field(i), depth+1)
			if err != nil {
				return reflect.Value{}, err
			}
			out.Field(i).Set(cloned)
		}
		return out, nil
	default:
		return value, nil
	}
}

func CandidateDigestV1(req ToolRequest) (string, error) {
	if req.Tool == "" {
		return "", fmt.Errorf("candidate tool is required")
	}
	if !utf8.ValidString(req.Tool) {
		return "", fmt.Errorf("candidate tool is not valid UTF-8")
	}
	input := map[string]any{
		"action":  req.Action,
		"payload": req.Payload,
		"context": req.Context,
	}
	if err := validateJSONValue(input); err != nil {
		return "", fmt.Errorf("candidate input is not valid interoperable JSON: %w", err)
	}
	raw, err := json.Marshal(map[string]any{"tool": req.Tool, "input": input})
	if err != nil {
		return "", fmt.Errorf("encode candidate: %w", err)
	}
	canonical, err := jcs.Transform(raw)
	if err != nil {
		return "", fmt.Errorf("canonicalize candidate with RFC 8785 JCS: %w", err)
	}
	preimage := append([]byte(candidateDigestPrefixV1), canonical...)
	sum := sha256.Sum256(preimage)
	return hex.EncodeToString(sum[:]), nil
}

func candidateAuthorizationContextV1(context map[string]any, digest string) map[string]any {
	out := cloneMap(context)
	if out == nil {
		out = make(map[string]any)
	}
	// These reserved values are always overwritten by the gateway, never
	// accepted from the adapter or agent as authoritative authorization data.
	out[candidateDigestContextKeyV1] = digest
	out[candidateProfileContextKeyV1] = CandidateDigestProfileV1
	return out
}

func validateJSONValue(value any) error {
	return validateJSONReflect(reflect.ValueOf(value), 0)
}

func validateJSONReflect(value reflect.Value, depth int) error {
	// CandidateDigestV1 is also callable directly, outside the gateway's
	// snapshot path. Bound recursive validation so cyclic or adversarially
	// deep values fail closed instead of exhausting the goroutine stack.
	if depth > 128 {
		return fmt.Errorf("candidate input exceeds validation depth limit or contains a cycle")
	}
	if !value.IsValid() {
		return nil
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		if value.IsNil() {
			return nil
		}
		return validateJSONReflect(value.Elem(), depth+1)
	case reflect.String:
		if !utf8.ValidString(value.String()) {
			return fmt.Errorf("invalid UTF-8 string")
		}
	case reflect.Float32, reflect.Float64:
		if math.IsNaN(value.Float()) || math.IsInf(value.Float(), 0) {
			return fmt.Errorf("non-finite number")
		}
	case reflect.Map:
		if value.IsNil() {
			return nil
		}
		if value.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("JSON object keys must be strings")
		}
		iter := value.MapRange()
		for iter.Next() {
			if !utf8.ValidString(iter.Key().String()) {
				return fmt.Errorf("invalid UTF-8 object key")
			}
			if err := validateJSONReflect(iter.Value(), depth+1); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		if value.Kind() == reflect.Slice && value.IsNil() {
			return nil
		}
		for i := 0; i < value.Len(); i++ {
			if err := validateJSONReflect(value.Index(i), depth+1); err != nil {
				return err
			}
		}
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Type().Field(i)
			if field.PkgPath == "" { // exported fields only
				if err := validateJSONReflect(value.Field(i), depth+1); err != nil {
					return err
				}
			}
		}
	}
	// encoding/json rejects unsupported values; JCS validates the encoded
	// representation and its interoperable numeric constraints.
	if value.CanInterface() {
		if _, err := json.Marshal(value.Interface()); err != nil {
			return err
		}
	}
	return nil
}
