// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package authboundary

import (
	"context"
	"net/http"
)

type Principal struct {
	Subject string
}

type principalContextKey struct{}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	if !ok || principal.Subject == "" {
		return Principal{}, false
	}
	return principal, true
}

// RequirePrincipal is fail-closed. Identity must be supplied by a trusted upstream
// NAEOS identity adapter; this package intentionally does not parse credentials.
func RequirePrincipal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := PrincipalFromContext(r.Context()); !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
