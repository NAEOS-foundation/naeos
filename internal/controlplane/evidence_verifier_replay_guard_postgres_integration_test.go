// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0
package controlplane

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"sync"
	"testing"
)

func TestPostgresEvidenceVerifierReplayGuardIntegration(t *testing.T) {
	url := os.Getenv("NAEOS_EVIDENCE_VERIFIER_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set NAEOS_EVIDENCE_VERIFIER_TEST_DATABASE_URL to an isolated verifier test DB with migration applied")
	}
	ctx := context.Background()
	g, err := NewPostgresEvidenceVerifierReplayGuard(ctx, url)
	if err != nil {
		t.Fatalf("open replay guard: %v", err)
	}
	defer g.Close()
	idb := make([]byte, 16)
	if _, err := rand.Read(idb); err != nil {
		t.Fatal(err)
	}
	id := hex.EncodeToString(idb)
	scope := "integration-" + id
	const contenders = 24
	var wg sync.WaitGroup
	wins := 0
	var mu sync.Mutex
	for i := 0; i < contenders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, err := g.Claim(ctx, scope, id)
			if err != nil {
				t.Errorf("claim failed: %v", err)
				return
			}
			if ok {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if wins != 1 {
		t.Fatalf("concurrent winners=%d want 1", wins)
	}
	g2, err := NewPostgresEvidenceVerifierReplayGuard(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer g2.Close()
	again, err := g2.Claim(ctx, scope, id)
	if err != nil {
		t.Fatal(err)
	}
	if again {
		t.Fatal("duplicate claim succeeded after separate pool/reopen")
	}
	other, err := g2.Claim(ctx, scope+"-other", id)
	if err != nil {
		t.Fatal(err)
	}
	if !other {
		t.Fatal("same request ID must be claimable in a different verifier scope")
	}
}
