// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
	pluginapi "github.com/NAEOS-foundation/naeos/pkg/pluginapi/v1"
)

func respond(resp pluginapi.Response) {
	// stdout is reserved for exactly one response object; diagnostics go to stderr.
	out, err := json.Marshal(resp)
	if err != nil {
		resp = pluginapi.Response{
			ProtocolVersion: pluginapi.ProtocolVersion,
			OK:              false,
			Error:           &pluginapi.Error{Code: pluginapi.ErrorCodeInvalidResponse, Message: "encode response"},
		}
		out, _ = json.Marshal(resp)
	}
	fmt.Println(string(out))
}

func fail(req pluginapi.Request, code, message string) {
	respond(pluginapi.Response{
		ProtocolVersion: req.ProtocolVersion,
		RequestID:       req.RequestID,
		OK:              false,
		Error:           &pluginapi.Error{Code: code, Message: message},
	})
	os.Exit(1)
}

// main is the WASM/CLI entry point. It reads one pluginapi v1 request object
// from stdin, executes the plugin, and writes exactly one v1 response object
// on stdout (JSON-over-stdio protocol).
func main() {
	var req pluginapi.Request
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		fail(pluginapi.Request{ProtocolVersion: pluginapi.ProtocolVersion}, pluginapi.ErrorCodeInvalidRequest, "invalid request: "+err.Error())
	}
	if err := req.Validate(); err != nil {
		fail(req, pluginapi.ErrorCodeInvalidRequest, err.Error())
	}

	ctx := &pluginhost.PluginContext{}
	p := New()
	if err := p.Initialize(ctx); err != nil {
		fail(req, pluginapi.ErrorCodePluginError, "initialize: "+err.Error())
	}
	defer func() { _ = p.Shutdown() }()

	params := map[string]any{}
	for key, raw := range req.Params {
		if len(raw) == 0 {
			continue
		}
		var value any
		if err := json.Unmarshal(raw, &value); err != nil {
			fail(req, pluginapi.ErrorCodeInvalidParams, fmt.Sprintf("param %q: %v", key, err))
		}
		params[key] = value
	}

	result, err := p.Execute(req.Action, params)
	if err != nil {
		fail(req, pluginapi.ErrorCodePluginError, err.Error())
	}

	encoded, encErr := json.Marshal(result)
	if encErr != nil {
		fail(req, pluginapi.ErrorCodeInvalidResponse, "encode result: "+encErr.Error())
	}

	respond(pluginapi.Response{
		ProtocolVersion: req.ProtocolVersion,
		RequestID:       req.RequestID,
		OK:              true,
		Result:          encoded,
	})
}
