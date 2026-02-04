// Package middleware provides HTTP middleware for the MCP server.
package middleware

import "context"

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

// AuthTokenKey is the context key for the authentication token.
const AuthTokenKey ContextKey = "auth_token"

// TokenFromContext extracts the authentication token from the context.
func TokenFromContext(ctx context.Context) string {
	if token, ok := ctx.Value(AuthTokenKey).(string); ok {
		return token
	}
	return ""
}
