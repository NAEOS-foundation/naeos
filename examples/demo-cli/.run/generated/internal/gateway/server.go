// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"fmt"
	"net/http"
)

func Run(port int) error {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"service\":\"%s\",\"status\":\"ok\"}", "gateway")
	})
	addr := fmt.Sprintf("%d", port)
	fmt.Printf("%s listening on %s\n", "gateway", addr)
	return http.ListenAndServe(addr, handler)
}
