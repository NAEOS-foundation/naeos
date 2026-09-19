// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

type request struct {
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}

func main() {
	var req request
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		fmt.Fprintf(os.Stdout, `{"ok":false,"error":%q}`+"\n", "invalid request: "+err.Error())
		os.Exit(1)
	}

	p := New()
	if err := p.Initialize(&pluginhost.PluginContext{}); err != nil {
		fmt.Fprintf(os.Stdout, `{"ok":false,"error":%q}`+"\n", "initialize: "+err.Error())
		os.Exit(1)
	}
	defer func() { _ = p.Shutdown() }()

	result, err := p.Execute(req.Method, req.Params)
	if err != nil {
		fmt.Fprintf(os.Stdout, `{"ok":false,"error":%q}`+"\n", err.Error())
		os.Exit(1)
	}

	out, _ := json.Marshal(map[string]any{"ok": true, "result": result})
	fmt.Println(string(out))
}
