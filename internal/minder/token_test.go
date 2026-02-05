package minder

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewTokenRefresher(t *testing.T) {
	t.Parallel()

	refresher := NewTokenRefresher()
	if refresher == nil {
		t.Fatal("NewTokenRefresher() returned nil")
	}
	if refresher.httpClient == nil {
		t.Error("httpClient is nil")
	}
	if refresher.clientID != DefaultClientID {
		t.Errorf("clientID = %q, want %q", refresher.clientID, DefaultClientID)
	}
	if refresher.cache == nil {
		t.Error("cache is nil")
	}
	if refresher.realmURLs == nil {
		t.Error("realmURLs is nil")
	}
}

func TestTokenRefresher_Close(t *testing.T) {
	t.Parallel()

	refresher := NewTokenRefresher()
	// Close should not panic
	refresher.Close()
}

func TestGetValidAccessToken_EmptyToken(t *testing.T) {
	t.Parallel()

	refresher := NewTokenRefresher()
	defer refresher.Close()

	_, err := refresher.GetValidAccessToken(context.Background(), "", ServerConfig{})
	if err == nil {
		t.Error("expected error for empty token")
	}
	if !errors.Is(err, ErrNoToken) {
		t.Errorf("expected ErrNoToken, got: %v", err)
	}
}

func TestGetValidAccessToken_MalformedToken(t *testing.T) {
	t.Parallel()

	refresher := NewTokenRefresher()
	defer refresher.Close()

	_, err := refresher.GetValidAccessToken(context.Background(), "not-a-jwt", ServerConfig{})
	if err == nil {
		t.Error("expected error for malformed token")
	}
	if !errors.Is(err, ErrTokenMalformed) {
		t.Errorf("expected ErrTokenMalformed, got: %v", err)
	}
}

func TestGetValidAccessToken_ValidAccessToken(t *testing.T) {
	t.Parallel()

	refresher := NewTokenRefresher()
	defer refresher.Close()

	// Create a valid JWT with future expiration
	token := createTestJWT(t, map[string]interface{}{
		"typ": "Bearer",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	result, err := refresher.GetValidAccessToken(context.Background(), token, ServerConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != token {
		t.Errorf("token was modified; got different token back")
	}
}

func TestGetValidAccessToken_ExpiredAccessToken(t *testing.T) {
	t.Parallel()

	refresher := NewTokenRefresher()
	defer refresher.Close()

	// Create a JWT that's already expired
	token := createTestJWT(t, map[string]interface{}{
		"typ": "Bearer",
		"exp": time.Now().Add(-time.Hour).Unix(),
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	})

	_, err := refresher.GetValidAccessToken(context.Background(), token, ServerConfig{})
	if err == nil {
		t.Error("expected error for expired token")
	}
	if !errors.Is(err, ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired, got: %v", err)
	}
}

func TestExtractRealmFromWWWAuthenticate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "standard header",
			header: `Bearer realm="https://auth.example.com/realms/test"`,
			want:   "https://auth.example.com/realms/test",
		},
		{
			name:   "header with additional params",
			header: `Bearer realm="https://auth.example.com/realms/test", error="invalid_token"`,
			want:   "https://auth.example.com/realms/test",
		},
		{
			name:   "header without quotes",
			header: `Bearer realm=https://auth.example.com/realms/test`,
			want:   "https://auth.example.com/realms/test",
		},
		{
			name:   "not bearer",
			header: `Basic realm="test"`,
			want:   "",
		},
		{
			name:   "no realm",
			header: `Bearer error="invalid_token"`,
			want:   "",
		},
		{
			name:   "empty header",
			header: "",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractRealmFromWWWAuthenticate(tt.header)
			if got != tt.want {
				t.Errorf("extractRealmFromWWWAuthenticate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDiscoverRealmURL(t *testing.T) {
	t.Parallel()

	// Create a test server that returns WWW-Authenticate header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("WWW-Authenticate", `Bearer realm="https://auth.example.com/realms/test"`)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	// Parse server URL to get host and port
	host := strings.TrimPrefix(server.URL, "http://")
	parts := strings.Split(host, ":")
	if len(parts) != 2 {
		t.Fatalf("unexpected server URL format: %s", server.URL)
	}

	var port int
	if _, err := fmt.Sscanf(parts[1], "%d", &port); err != nil {
		t.Fatalf("failed to parse port: %v", err)
	}

	refresher := NewTokenRefresher()
	defer refresher.Close()

	cfg := ServerConfig{
		Host:     parts[0],
		Port:     port,
		Insecure: true,
	}

	realmURL, err := refresher.discoverRealmURL(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if realmURL != "https://auth.example.com/realms/test" {
		t.Errorf("realmURL = %q, want %q", realmURL, "https://auth.example.com/realms/test")
	}
}

func TestDiscoverRealmURL_NoHeader(t *testing.T) {
	t.Parallel()

	// Create a test server that doesn't return WWW-Authenticate header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	host := strings.TrimPrefix(server.URL, "http://")
	parts := strings.Split(host, ":")
	var port int
	if _, err := fmt.Sscanf(parts[1], "%d", &port); err != nil {
		t.Fatalf("failed to parse port: %v", err)
	}

	refresher := NewTokenRefresher()
	defer refresher.Close()

	cfg := ServerConfig{
		Host:     parts[0],
		Port:     port,
		Insecure: true,
	}

	_, err := refresher.discoverRealmURL(context.Background(), cfg)
	if err == nil {
		t.Error("expected error when no WWW-Authenticate header")
	}
	// Check the error message contains useful information
	if !strings.Contains(err.Error(), "authentication realm") {
		t.Errorf("error should mention authentication realm, got: %v", err)
	}
}

func TestValidateRealmURL(t *testing.T) {
	t.Parallel()

	refresher := NewTokenRefresher()
	defer refresher.Close()

	tests := []struct {
		name      string
		realmURL  string
		wantError bool
	}{
		{
			name:      "valid https URL",
			realmURL:  "https://auth.example.com/realms/test",
			wantError: false,
		},
		{
			name:      "valid http URL",
			realmURL:  "http://localhost:8080/realms/test",
			wantError: false,
		},
		{
			name:      "invalid scheme",
			realmURL:  "ftp://auth.example.com/realms/test",
			wantError: true,
		},
		{
			name:      "invalid URL",
			realmURL:  "not a url",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := refresher.validateRealmURL(tt.realmURL, "example.com")
			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestWrapOAuthError(t *testing.T) {
	t.Parallel()

	refresher := NewTokenRefresher()
	defer refresher.Close()

	tests := []struct {
		name        string
		err         error
		wantContain string
	}{
		{
			name:        "nil error",
			err:         nil,
			wantContain: "",
		},
		{
			name:        "invalid_grant error",
			err:         fmt.Errorf("oauth2: cannot fetch token: invalid_grant"),
			wantContain: "expired or revoked",
		},
		{
			name:        "invalid_client error",
			err:         fmt.Errorf("oauth2: invalid_client"),
			wantContain: "client authentication failed",
		},
		{
			name:        "unauthorized_client error",
			err:         fmt.Errorf("oauth2: unauthorized_client"),
			wantContain: "not authorized",
		},
		{
			name:        "generic error",
			err:         fmt.Errorf("network timeout"),
			wantContain: "network timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wrapped := refresher.wrapOAuthError(tt.err)
			if tt.err == nil {
				if wrapped != nil {
					t.Errorf("expected nil, got: %v", wrapped)
				}
				return
			}
			if !errors.Is(wrapped, ErrRefreshFailed) {
				t.Errorf("expected ErrRefreshFailed, got: %v", wrapped)
			}
			if !strings.Contains(wrapped.Error(), tt.wantContain) {
				t.Errorf("error should contain %q, got: %v", tt.wantContain, wrapped)
			}
		})
	}
}

// createTestJWT creates a simple JWT for testing (unsigned).
func createTestJWT(t *testing.T, claims map[string]interface{}) string {
	t.Helper()

	header := map[string]interface{}{
		"alg": "none",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("failed to marshal header: %v", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("failed to marshal claims: %v", err)
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	return fmt.Sprintf("%s.%s.", headerB64, claimsB64)
}
