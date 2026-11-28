// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
)

type policy struct {
	version string
}

func (p policy) Version() string { return p.version }

func (p policy) EvaluateExecutionEvidence(_ context.Context, o controlplane.AuthoritativeExecutionObservation, e controlplane.ExecutionOutcome) error {
	if o.Outcome != e {
		return controlplane.ErrAuthoritativeEvidenceNotVerified
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Printf("verifier stopped: %v", err)
		os.Exit(1)
	}
}

func run() error {
	sourceURL := os.Getenv("NAEOS_EVIDENCE_SOURCE_URL")
	sourceToken := os.Getenv("NAEOS_EVIDENCE_SOURCE_TOKEN")
	serviceToken := os.Getenv("NAEOS_EVIDENCE_VERIFIER_TOKEN")
	version := strings.TrimSpace(os.Getenv("NAEOS_EVIDENCE_POLICY_VERSION"))
	addr := os.Getenv("NAEOS_EVIDENCE_VERIFIER_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8081"
	}
	if sourceURL == "" || sourceToken == "" || serviceToken == "" || version == "" {
		return errors.New("required verifier configuration is missing")
	}
	src, err := controlplane.NewHTTPAuthoritativeEvidenceSource(sourceURL, sourceToken)
	if err != nil {
		return err
	}
	svc, err := controlplane.NewProviderNeutralEvidenceVerificationService(src, policy{version}, serviceToken)
	if err != nil {
		return err
	}
	handler := newHandler(svc)
	server := &http.Server{
		Addr: addr, Handler: handler, ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second,
		IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 * 1024,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errs := make(chan error, 1)
	go func() { errs <- server.ListenAndServe() }()
	select {
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	case err := <-errs:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func newHandler(svc *controlplane.ProviderNeutralEvidenceVerificationService) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/v1/verify", svc.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/configz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("configured"))
	})
	// The source interface has no dependency health contract. Do not claim
	// dependency readiness until a bounded, read-only probe is implemented.
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.Error(w, "authoritative source readiness is not verified", http.StatusServiceUnavailable)
	})
	return mux
}
