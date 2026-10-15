// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/monitoring"
)

func (s *Server) Handler() http.Handler {
	mw := monitoring.MetricsMiddleware(s.metrics)
	v2Handler := s.v2IdempotencyMiddleware(s.Router)
	return mw(s.loggingMiddleware(s.handlerWithMiddleware(v2Handler.ServeHTTP)))
}

// Start begins listening for HTTP requests and handles graceful shutdown.
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:         s.Addr,
		Handler:      s.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		slog.Warn("shutting down server", "component", "api-server")
		if s.wsServer != nil {
			s.wsServer.Stop()
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = s.server.Shutdown(ctx)
	}()

	slog.Info("starting NAEOS API server", "addr", s.Addr, "component", "api-server")
	return s.server.ListenAndServe()
}

// Stop gracefully shuts down the server and any attached WebSocket server.
func (s *Server) Stop() error {
	if s.wsServer != nil {
		s.wsServer.Stop()
	}
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}
