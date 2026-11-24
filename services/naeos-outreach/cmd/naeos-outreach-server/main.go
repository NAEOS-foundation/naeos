// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	outreachconfig "github.com/NAEOS-foundation/naeos/services/naeos-outreach/internal/config"
	outreachhttp "github.com/NAEOS-foundation/naeos/services/naeos-outreach/internal/httpserver"
)

func main() {
	cfg, err := outreachconfig.Load(os.Getenv("NAEOS_OUTREACH_ADDR"))
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           outreachhttp.NewHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("NAEOS Outreach server listening at %s", cfg.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
			_ = server.Close()
		}
		if err := <-serverErrors; err != nil && err != http.ErrServerClosed {
			log.Printf("server shutdown: %v", err)
		}
	}
}
