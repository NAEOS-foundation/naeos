// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package investordemo

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestControlPlaneAuthorizeFailsClosedWithoutToken(t *testing.T) {
	security := newControlPlaneSecurity("", "")
	request := httptest.NewRequest(http.MethodPost, "/api/control-plane/decision", nil)
	request.Header.Set("Authorization", "Bearer plausible-token")

	if security.authorize(request) {
		t.Fatal("expected authorization to fail when server-side token is unset")
	}
}

func TestControlPlaneAuthorizeRequiresExactBearerToken(t *testing.T) {
	security := newControlPlaneSecurity("", "expected-token")

	for _, authorization := range []string{"", "Bearer wrong-token", "expected-token"} {
		request := httptest.NewRequest(http.MethodPost, "/api/control-plane/decision", nil)
		if authorization != "" {
			request.Header.Set("Authorization", authorization)
		}
		if security.authorize(request) {
			t.Fatalf("expected authorization to fail for header %q", authorization)
		}
	}

	request := httptest.NewRequest(http.MethodPost, "/api/control-plane/decision", nil)
	request.Header.Set("Authorization", "Bearer expected-token")
	if !security.authorize(request) {
		t.Fatal("expected exact configured bearer token to authorize")
	}
}
