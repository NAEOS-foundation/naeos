// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package pluginapi_test

import (
	"encoding/json"
	"testing"

	pluginapi "github.com/NAEOS-foundation/naeos/pkg/pluginapi/v1"
)

func TestProtocolEnvelopeRoundTripAndValidation(t *testing.T) {
	request := pluginapi.Request{
		ProtocolVersion: pluginapi.ProtocolVersion,
		RequestID:       "request-1",
		Method:          "execute",
		Action:          "ping",
		Params:          map[string]json.RawMessage{"count": json.RawMessage(`3`)},
	}
	if err := request.Validate(); err != nil {
		t.Fatalf("request validation: %v", err)
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	var decoded pluginapi.Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if err := decoded.Validate(); err != nil {
		t.Fatalf("decoded request validation: %v", err)
	}

	response := pluginapi.Response{
		ProtocolVersion: pluginapi.ProtocolVersion,
		RequestID:       request.RequestID,
		OK:              true,
		Result:          json.RawMessage(`{"status":"ok"}`),
	}
	if err := response.Validate(request); err != nil {
		t.Fatalf("response validation: %v", err)
	}
}

func TestProtocolRejectsUnsupportedOrUncorrelatedMessages(t *testing.T) {
	request := pluginapi.Request{
		ProtocolVersion: "2.0", RequestID: "request-1", Method: "execute", Action: "ping",
	}
	if err := request.Validate(); err == nil {
		t.Fatal("expected unsupported protocol version to fail")
	}
	request.ProtocolVersion = pluginapi.ProtocolVersion
	response := pluginapi.Response{
		ProtocolVersion: pluginapi.ProtocolVersion, RequestID: "other", OK: true,
	}
	if err := response.Validate(request); err == nil {
		t.Fatal("expected mismatched response request ID to fail")
	}
}

func TestManifestProtocolSupportUsesExactVersion(t *testing.T) {
	manifest := pluginapi.Manifest{ProtocolVersions: []string{"1.0", "1.1"}}
	if !manifest.SupportsProtocol(pluginapi.ProtocolVersion) || manifest.SupportsProtocol("1") {
		t.Fatal("manifest protocol support did not use exact matching")
	}
}
