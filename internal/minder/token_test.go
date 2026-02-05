package minder

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// mockUserService is a mock gRPC server for testing realm URL discovery.
type mockUserService struct {
	minderv1.UnimplementedUserServiceServer
	realmURL string
}

func (m *mockUserService) GetUser(ctx context.Context, _ *minderv1.GetUserRequest) (*minderv1.GetUserResponse, error) {
	if m.realmURL != "" {
		_ = grpc.SendHeader(ctx, metadata.New(map[string]string{
			"www-authenticate": fmt.Sprintf(`Bearer realm="%s"`, m.realmURL),
		}))
	}
	return nil, status.Error(codes.Unauthenticated, "unauthenticated")
}

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

	// Start a mock gRPC server that returns www-authenticate header
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	srv := grpc.NewServer()
	minderv1.RegisterUserServiceServer(srv, &mockUserService{
		realmURL: "https://auth.example.com/realms/test",
	})
	go func() {
		_ = srv.Serve(lis)
	}()
	defer srv.Stop()

	// Parse listener address to get host and port
	addr := lis.Addr().String()
	parts := strings.Split(addr, ":")
	require.Len(t, parts, 2, "unexpected listener address format: %s", addr)

	var port int
	_, err = fmt.Sscanf(parts[1], "%d", &port)
	require.NoError(t, err)

	refresher := NewTokenRefresher()
	defer refresher.Close()

	cfg := ServerConfig{
		Host:     parts[0],
		Port:     port,
		Insecure: true,
	}

	realmURL, err := refresher.discoverRealmURL(context.Background(), cfg)
	require.NoError(t, err)
	require.Equal(t, "https://auth.example.com/realms/test", realmURL)
}

func TestDiscoverRealmURL_NoHeader(t *testing.T) {
	t.Parallel()

	// Start a mock gRPC server that doesn't return www-authenticate header
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	srv := grpc.NewServer()
	minderv1.RegisterUserServiceServer(srv, &mockUserService{
		realmURL: "", // Empty realm URL means no header will be sent
	})
	go func() {
		_ = srv.Serve(lis)
	}()
	defer srv.Stop()

	addr := lis.Addr().String()
	parts := strings.Split(addr, ":")
	require.Len(t, parts, 2)

	var port int
	_, err = fmt.Sscanf(parts[1], "%d", &port)
	require.NoError(t, err)

	refresher := NewTokenRefresher()
	defer refresher.Close()

	cfg := ServerConfig{
		Host:     parts[0],
		Port:     port,
		Insecure: true,
	}

	_, err = refresher.discoverRealmURL(context.Background(), cfg)
	require.Error(t, err)
	// Check the error message contains useful information
	require.Contains(t, err.Error(), "www-authenticate")
}

func TestValidateRealmURL(t *testing.T) {
	t.Parallel()

	refresher := NewTokenRefresher()
	defer refresher.Close()

	tests := []struct {
		name         string
		realmURL     string
		expectedHost string
		wantError    bool
	}{
		{
			name:         "valid https URL with related host",
			realmURL:     "https://auth.example.com/realms/test",
			expectedHost: "api.example.com",
			wantError:    false,
		},
		{
			name:         "valid http URL for localhost",
			realmURL:     "http://localhost:8080/realms/test",
			expectedHost: "localhost",
			wantError:    false,
		},
		{
			name:         "valid http URL for 127.0.0.1",
			realmURL:     "http://127.0.0.1:8080/realms/test",
			expectedHost: "127.0.0.1",
			wantError:    false,
		},
		{
			name:         "http not allowed for non-localhost",
			realmURL:     "http://auth.example.com/realms/test",
			expectedHost: "api.example.com",
			wantError:    true,
		},
		{
			name:         "invalid scheme",
			realmURL:     "ftp://auth.example.com/realms/test",
			expectedHost: "api.example.com",
			wantError:    true,
		},
		{
			name:         "invalid URL",
			realmURL:     "not a url",
			expectedHost: "example.com",
			wantError:    true,
		},
		{
			name:         "URL with embedded credentials rejected",
			realmURL:     "https://user:pass@auth.example.com/realms/test",
			expectedHost: "api.example.com",
			wantError:    true,
		},
		{
			name:         "private IP address rejected",
			realmURL:     "https://192.168.1.1/realms/test",
			expectedHost: "api.example.com",
			wantError:    true,
		},
		{
			name:         "link-local IP rejected",
			realmURL:     "https://169.254.1.1/realms/test",
			expectedHost: "api.example.com",
			wantError:    true,
		},
		{
			name:         "unrelated host rejected",
			realmURL:     "https://auth.attacker.com/realms/test",
			expectedHost: "api.example.com",
			wantError:    true,
		},
		{
			name:         "same base domain allowed",
			realmURL:     "https://auth.stacklok.com/realms/test",
			expectedHost: "api.stacklok.com",
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := refresher.validateRealmURL(tt.realmURL, tt.expectedHost)
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
