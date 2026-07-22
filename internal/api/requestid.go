package api

import (
	"context"
	"crypto/rand"
	"fmt"
)

type contextKey string

// RequestIDKey is the context key used to store the request ID.
const RequestIDKey contextKey = "request_id"

// UserContextKey is the context key used to store the authenticated user ID.
const UserContextKey contextKey = "user_id"

// TenantContextKey is the context key used to store the tenant ID.
const TenantContextKey contextKey = "tenant_id"

// GenerateRequestID returns a new random UUID v4 string.
func GenerateRequestID() string {
	var buf [16]byte
	_, _ = rand.Read(buf[:])
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16])
}

// ContextWithRequestID stores the request ID in the context.
func ContextWithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, RequestIDKey, id)
}

// RequestIDFromContext retrieves the request ID from the context.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}
