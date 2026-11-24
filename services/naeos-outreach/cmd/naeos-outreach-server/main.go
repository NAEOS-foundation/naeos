// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	outreachhttp "github.com/NAEOS-foundation/naeos/services/naeos-outreach/internal/httpserver"
)

func main() {
	addr := os.Getenv("NAEOS_OUTREACH_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8088"
	}
	handler := outreachhttp.NewHandler()
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	// Validate bind address before starting. Loopback is the default; callers
	// explicitly opt into a non-loopback listener by setting NAEOS_OUTREACH_ADDR.
	if _, _, err := net.SplitHostPort(addr); err != nil {
		log.Fatalf("invalid NAEOS_OUTREACH_ADDR %q: %v", addr, err)
	}
	log.Printf("NAEOS Outreach server listening at %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
