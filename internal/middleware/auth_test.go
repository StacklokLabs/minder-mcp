package middleware

import (
	"context"
	"testing"
)

func TestContextWithToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "sets non-empty token",
			token: "my-jwt-token",
		},
		{
			name:  "sets empty token",
			token: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := ContextWithToken(context.Background(), tt.token)
			got := TokenFromContext(ctx)
			if got != tt.token {
				t.Errorf("TokenFromContext() after ContextWithToken() = %q, want %q", got, tt.token)
			}
		})
	}
}

func TestTokenFromContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "returns token when set via ContextWithToken",
			ctx:  ContextWithToken(context.Background(), "test-token"),
			want: "test-token",
		},
		{
			name: "returns empty string when token not in context",
			ctx:  context.Background(),
			want: "",
		},
		{
			name: "returns empty string when wrong type stored at key",
			ctx:  context.WithValue(context.Background(), authTokenKey, 12345),
			want: "",
		},
		{
			name: "returns empty string for nil context value",
			ctx:  context.WithValue(context.Background(), authTokenKey, nil),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := TokenFromContext(tt.ctx)
			if got != tt.want {
				t.Errorf("TokenFromContext() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContextKeyCollision(t *testing.T) {
	t.Parallel()

	// Verify that using a different key type with same string value doesn't collide
	type otherKey struct{ name string }
	otherAuthKey := &otherKey{"auth_token"}

	ctx := context.WithValue(context.Background(), otherAuthKey, "other-token")
	ctx = ContextWithToken(ctx, "real-token")

	// Our TokenFromContext should get our token, not the other one
	got := TokenFromContext(ctx)
	if got != "real-token" {
		t.Errorf("TokenFromContext() = %q, want %q", got, "real-token")
	}

	// The other key's value should still be accessible separately
	if other, ok := ctx.Value(otherAuthKey).(string); !ok || other != "other-token" {
		t.Errorf("other key value = %q, want %q", other, "other-token")
	}
}
