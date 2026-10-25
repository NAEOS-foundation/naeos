// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package neir_test

import (
	"encoding/json"
	"reflect"
	"testing"

	neirv1 "github.com/NAEOS-foundation/naeos/pkg/neir/v1"
)

func TestDecodeEncodePreservesUnknownFields(t *testing.T) {
	input := []byte(`{"project":{"name":"demo","future_project":{"enabled":true}},"modules":[{"name":"core","path":"./core","future_module":"kept"}],"services":[{"name":"api","endpoints":[{"path":"/health","future_endpoint":7}]}],"metadata":{"neir_version":"2.0.0","future_metadata":[1,2]},"future_root":{"nested":"kept"}}`)

	document, err := neirv1.Decode(input)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if document.Project == nil || document.Project.Name != "demo" || len(document.Modules) != 1 || document.Modules[0].Name != "core" {
		t.Fatalf("typed fields were not decoded: %#v", document)
	}
	if document.Metadata.NEIRVersion != "2.0.0" {
		t.Fatalf("metadata version = %q", document.Metadata.NEIRVersion)
	}

	output, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var want, got any
	if err := json.Unmarshal(input, &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip changed document\n got: %#v\nwant: %#v", got, want)
	}
}

func TestDecodeRejectsNonObjectDocument(t *testing.T) {
	for _, input := range []string{`null`, `[]`, `"neir"`} {
		if _, err := neirv1.Decode([]byte(input)); err == nil {
			t.Errorf("Decode(%s) unexpectedly succeeded", input)
		}
	}
}

func TestMarshalRejectsInvalidAdditionalPropertyJSON(t *testing.T) {
	document := &neirv1.Document{
		AdditionalProperties: map[string]json.RawMessage{"future": json.RawMessage("{")},
	}
	if _, err := json.Marshal(document); err == nil {
		t.Fatal("expected invalid additional JSON to fail")
	}
}
