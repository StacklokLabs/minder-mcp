// Package middleware provides HTTP middleware for the MCP server.
package middleware

import "context"

// contextKey is an unexported type for context keys to prevent collisions.
// Using a struct with a name field aids debugging.
type contextKey struct{ name string }

// authTokenKey is the unexported context key for the authentication token.
var authTokenKey = &contextKey{"auth_token"}

// ContextWithToken returns a new context with the authentication token set.
func ContextWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, authTokenKey, token)
}

// TokenFromContext extracts the authentication token from the context.
func TokenFromContext(ctx context.Context) string {
	if token, ok := ctx.Value(authTokenKey).(string); ok {
		return token
	}
	return ""
}
