package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

// main is the WASM/CLI entry point. It reads the action from argv[1] and an
// optional JSON-encoded params object from argv[2], executes the plugin, and
// prints the result as JSON on stdout (JSON-over-stdio protocol).
func main() {
	if len(os.Args) < 2 {
		fmt.Println(`{"error":"missing action","ok":false}`)
		os.Exit(1)
	}

	action := os.Args[1]
	var params map[string]any
	if len(os.Args) > 2 {
		_ = json.Unmarshal([]byte(os.Args[2]), &params)
	}

	ctx := &pluginhost.PluginContext{}
	p := New()
	if err := p.Initialize(ctx); err != nil {
		fmt.Printf(`{"error":"initialize: %s","ok":false}`+"\n", err)
		os.Exit(1)
	}
	defer func() { _ = p.Shutdown() }()

	result, err := p.Execute(action, params)
	if err != nil {
		fmt.Printf(`{"error":"%s","ok":false}`+"\n", err)
		os.Exit(1)
	}

	out, _ := json.Marshal(map[string]any{"ok": true, "result": result})
	fmt.Println(string(out))
}
