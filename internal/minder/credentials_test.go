package minder

import (
	"context"
	"testing"

	"google.golang.org/grpc/credentials"
)

// Compile-time interface compliance check
var _ credentials.PerRPCCredentials = (*JWTTokenCredentials)(nil)

func TestNewJWTTokenCredentials(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "creates credentials with token",
			token: "my-jwt-token",
		},
		{
			name:  "creates credentials with empty token",
			token: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			creds := NewJWTTokenCredentials(tt.token)
			if creds == nil {
				t.Fatal("NewJWTTokenCredentials() returned nil")
			}
			if creds.token != tt.token {
				t.Errorf("token = %q, want %q", creds.token, tt.token)
			}
		})
	}
}

func TestJWTTokenCredentials_GetRequestMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		token    string
		wantAuth string
	}{
		{
			name:     "returns bearer token",
			token:    "my-jwt-token",
			wantAuth: "Bearer my-jwt-token",
		},
		{
			name:     "returns bearer with empty token",
			token:    "",
			wantAuth: "Bearer ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			creds := NewJWTTokenCredentials(tt.token)
			metadata, err := creds.GetRequestMetadata(context.Background())
			if err != nil {
				t.Fatalf("GetRequestMetadata() error = %v", err)
			}
			if metadata == nil {
				t.Fatal("GetRequestMetadata() returned nil metadata")
			}
			if auth, ok := metadata["authorization"]; !ok {
				t.Error("metadata missing 'authorization' key")
			} else if auth != tt.wantAuth {
				t.Errorf("authorization = %q, want %q", auth, tt.wantAuth)
			}
		})
	}
}

func TestJWTTokenCredentials_GetRequestMetadata_IgnoresURI(t *testing.T) {
	t.Parallel()

	creds := NewJWTTokenCredentials("test-token")

	// Call with various URI arguments - should all return same result
	uriCases := [][]string{
		{},
		{"https://example.com"},
		{"https://example.com", "https://other.com"},
	}

	for _, uris := range uriCases {
		metadata, err := creds.GetRequestMetadata(context.Background(), uris...)
		if err != nil {
			t.Fatalf("GetRequestMetadata(%v) error = %v", uris, err)
		}
		if metadata["authorization"] != "Bearer test-token" {
			t.Errorf("GetRequestMetadata(%v) authorization = %q, want %q",
				uris, metadata["authorization"], "Bearer test-token")
		}
	}
}

func TestJWTTokenCredentials_RequireTransportSecurity(t *testing.T) {
	t.Parallel()

	creds := NewJWTTokenCredentials("any-token")
	if creds.RequireTransportSecurity() {
		t.Error("RequireTransportSecurity() = true, want false")
	}
}
