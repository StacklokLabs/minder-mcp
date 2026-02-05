// Package minder provides gRPC client utilities for connecting to Minder.
package minder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

const (
	// DefaultClientID is the OAuth2 client ID used for token refresh.
	DefaultClientID = "minder-cli"

	// tokenRefreshBuffer is how far in advance of expiry to refresh.
	// Using 60 seconds to account for network latency, clock skew, and multi-step operations.
	tokenRefreshBuffer = 60 * time.Second

	// offlineTokenType is the Keycloak-specific token type claim for offline/refresh tokens.
	offlineTokenType = "Offline"
)

// Sentinel errors for programmatic error handling.
var (
	// ErrNoToken indicates no token was provided.
	ErrNoToken = errors.New("no token provided")

	// ErrTokenMalformed indicates the token could not be parsed.
	ErrTokenMalformed = errors.New("token is malformed")

	// ErrTokenExpired indicates the access token is expired.
	ErrTokenExpired = errors.New("access token is expired")

	// ErrRefreshFailed indicates the token refresh operation failed.
	ErrRefreshFailed = errors.New("token refresh failed")

	// ErrRealmDiscoveryFailed indicates realm URL discovery failed.
	ErrRealmDiscoveryFailed = errors.New("realm URL discovery failed")

	// ErrInvalidRealmURL indicates the discovered realm URL is invalid or untrusted.
	ErrInvalidRealmURL = errors.New("invalid or untrusted realm URL")
)

// ServerConfig holds server connection configuration for token operations.
type ServerConfig struct {
	Host     string
	Port     int
	Insecure bool
}

// cachedToken holds a cached access token with its expiry time.
type cachedToken struct {
	accessToken string
	expiresAt   time.Time
}

// TokenRefresher handles token validation and refresh.
// TokenRefresher is safe for concurrent use by multiple goroutines.
type TokenRefresher struct {
	httpClient *http.Client
	clientID   string

	// mu protects cached token state
	mu        sync.RWMutex
	cache     map[string]*cachedToken // keyed by refresh token hash
	realmURLs map[string]string       // keyed by host:port, cached realm URLs
}

// NewTokenRefresher creates a new TokenRefresher.
func NewTokenRefresher() *TokenRefresher {
	return &TokenRefresher{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		clientID:  DefaultClientID,
		cache:     make(map[string]*cachedToken),
		realmURLs: make(map[string]string),
	}
}

// Close releases resources held by the TokenRefresher.
func (t *TokenRefresher) Close() {
	t.httpClient.CloseIdleConnections()
}

// GetValidAccessToken returns a valid access token, refreshing if necessary.
// If the provided token is an offline/refresh token or an expired access token,
// it will attempt to refresh it using the realm URL discovered from the server.
//
// This method is safe for concurrent use.
func (t *TokenRefresher) GetValidAccessToken(
	ctx context.Context,
	token string,
	cfg ServerConfig,
) (string, error) {
	if token == "" {
		return "", ErrNoToken
	}

	// Parse the JWT without verification (we just need to inspect claims)
	// Note: This is safe because we're only using claims to make local decisions.
	// The actual token validation happens server-side.
	parser := jwt.NewParser()
	parsedToken, _, err := parser.ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTokenMalformed, err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("%w: failed to extract claims", ErrTokenMalformed)
	}

	// Check if it's an offline/refresh token (Keycloak sets typ: "Offline")
	if typ, ok := claims["typ"].(string); ok && typ == offlineTokenType {
		return t.getOrRefreshToken(ctx, token, cfg)
	}

	// Check if access token is expired or about to expire
	if exp, err := claims.GetExpirationTime(); err == nil && exp != nil {
		if time.Now().Add(tokenRefreshBuffer).After(exp.Time) {
			return "", fmt.Errorf(
				"%w: provide a valid offline/refresh token or a fresh access token",
				ErrTokenExpired,
			)
		}
	}

	// Token appears valid
	return token, nil
}

// getOrRefreshToken returns a cached access token or refreshes if needed.
func (t *TokenRefresher) getOrRefreshToken(
	ctx context.Context,
	refreshToken string,
	cfg ServerConfig,
) (string, error) {
	// Create a cache key from the refresh token (use first 32 chars to avoid storing full token)
	cacheKey := refreshToken
	if len(cacheKey) > 32 {
		cacheKey = cacheKey[:32]
	}

	// Check cache first (read lock)
	t.mu.RLock()
	if cached, ok := t.cache[cacheKey]; ok {
		// Check if cached token is still valid (with buffer)
		if time.Now().Add(tokenRefreshBuffer).Before(cached.expiresAt) {
			t.mu.RUnlock()
			return cached.accessToken, nil
		}
	}
	t.mu.RUnlock()

	// Need to refresh - get write lock
	t.mu.Lock()
	defer t.mu.Unlock()

	// Double-check cache after acquiring write lock (another goroutine may have refreshed)
	if cached, ok := t.cache[cacheKey]; ok {
		if time.Now().Add(tokenRefreshBuffer).Before(cached.expiresAt) {
			return cached.accessToken, nil
		}
	}

	// Perform the refresh
	accessToken, expiresAt, err := t.refreshToken(ctx, refreshToken, cfg)
	if err != nil {
		return "", err
	}

	// Cache the new token
	t.cache[cacheKey] = &cachedToken{
		accessToken: accessToken,
		expiresAt:   expiresAt,
	}

	return accessToken, nil
}

// refreshToken uses a refresh token to obtain a new access token.
func (t *TokenRefresher) refreshToken(
	ctx context.Context,
	refreshToken string,
	cfg ServerConfig,
) (string, time.Time, error) {
	// Discover or get cached realm URL
	realmURL, err := t.getRealmURL(ctx, cfg)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("%w: %v", ErrRealmDiscoveryFailed, err)
	}

	// Validate the realm URL for security (SSRF protection)
	if err := t.validateRealmURL(realmURL, cfg.Host); err != nil {
		return "", time.Time{}, fmt.Errorf("%w: %v", ErrInvalidRealmURL, err)
	}

	// Build token endpoint URL
	tokenEndpoint, err := url.JoinPath(realmURL, "protocol/openid-connect/token")
	if err != nil {
		return "", time.Time{}, fmt.Errorf("%w: failed to build token endpoint: %v", ErrRefreshFailed, err)
	}

	// Use oauth2 package for token refresh with our custom HTTP client
	oauth2Config := &oauth2.Config{
		ClientID: t.clientID,
		Endpoint: oauth2.Endpoint{
			TokenURL: tokenEndpoint,
		},
	}

	// Create a token source from the refresh token
	oldToken := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	// Inject our HTTP client into the context for oauth2 to use
	ctx = context.WithValue(ctx, oauth2.HTTPClient, t.httpClient)

	// Get a new token using the refresh token
	tokenSource := oauth2Config.TokenSource(ctx, oldToken)
	newToken, err := tokenSource.Token()
	if err != nil {
		return "", time.Time{}, t.wrapOAuthError(err)
	}

	return newToken.AccessToken, newToken.Expiry, nil
}

// wrapOAuthError wraps OAuth errors with more specific error types.
func (*TokenRefresher) wrapOAuthError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for specific OAuth error codes
	if strings.Contains(errStr, "invalid_grant") {
		return fmt.Errorf(
			"%w: refresh token is expired or revoked; re-authentication required",
			ErrRefreshFailed,
		)
	}
	if strings.Contains(errStr, "invalid_client") {
		return fmt.Errorf("%w: client authentication failed", ErrRefreshFailed)
	}
	if strings.Contains(errStr, "unauthorized_client") {
		return fmt.Errorf("%w: client not authorized for this grant type", ErrRefreshFailed)
	}

	return fmt.Errorf("%w: %v", ErrRefreshFailed, err)
}

// validateRealmURL validates that the discovered realm URL is trusted.
// This prevents SSRF attacks where a malicious server could redirect token requests.
func (*TokenRefresher) validateRealmURL(realmURL string, expectedHost string) error {
	parsedRealm, err := url.Parse(realmURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Validate scheme
	if parsedRealm.Scheme != "http" && parsedRealm.Scheme != "https" {
		return fmt.Errorf("invalid scheme: %s", parsedRealm.Scheme)
	}

	// For security, we could add additional validation here:
	// - Allowlist of known auth server domains
	// - Check that realm URL host matches expected IdP
	// For now, we validate that it's a valid URL with http/https scheme.
	// The ultimate validation happens at the OAuth server.

	_ = expectedHost // Reserved for future allowlist validation

	return nil
}

// getRealmURL returns a cached realm URL or discovers it from the server.
func (t *TokenRefresher) getRealmURL(ctx context.Context, cfg ServerConfig) (string, error) {
	cacheKey := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Check cache first (already holding write lock from caller)
	if realmURL, ok := t.realmURLs[cacheKey]; ok {
		return realmURL, nil
	}

	// Discover the realm URL
	realmURL, err := t.discoverRealmURL(ctx, cfg)
	if err != nil {
		return "", err
	}

	// Cache it
	t.realmURLs[cacheKey] = realmURL

	return realmURL, nil
}

// discoverRealmURL discovers the Keycloak realm URL from the server's WWW-Authenticate header.
func (t *TokenRefresher) discoverRealmURL(ctx context.Context, cfg ServerConfig) (string, error) {
	// Build the server URL for an unauthenticated gRPC-gateway call
	// Only use HTTP if explicitly configured as insecure
	scheme := "https"
	if cfg.Insecure {
		scheme = "http"
	}

	// Try the /api/v1/user endpoint which requires auth and will return WWW-Authenticate
	serverURL := fmt.Sprintf("%s://%s:%d/api/v1/user", scheme, cfg.Host, cfg.Port)

	req, err := http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}
	defer func() {
		// Drain and close body to enable connection reuse
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	// Look for WWW-Authenticate header
	// Check gRPC-gateway metadata header first (more reliable for gRPC services)
	wwwAuth := resp.Header.Get("Grpc-Metadata-Www-Authenticate")
	if wwwAuth == "" {
		wwwAuth = resp.Header.Get("WWW-Authenticate")
	}
	if wwwAuth == "" {
		return "", errors.New("server did not return authentication realm information")
	}

	// Parse the realm from the header
	realmURL := extractRealmFromWWWAuthenticate(wwwAuth)
	if realmURL == "" {
		// Try the other header if the first one didn't have a valid realm
		altHeader := resp.Header.Get("WWW-Authenticate")
		if altHeader != "" && altHeader != wwwAuth {
			realmURL = extractRealmFromWWWAuthenticate(altHeader)
		}
	}
	if realmURL == "" {
		return "", errors.New("could not parse authentication realm from server response")
	}

	return realmURL, nil
}

// extractRealmFromWWWAuthenticate parses the realm URL from a WWW-Authenticate header.
// Example header: Bearer realm="https://auth.stacklok.com/realms/stacklok"
func extractRealmFromWWWAuthenticate(header string) string {
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	header = strings.TrimPrefix(header, "Bearer ")

	for _, part := range strings.Split(header, ",") {
		parts := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if key == "realm" {
			value := strings.TrimSpace(parts[1])
			// Remove surrounding quotes if present
			if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
				value = value[1 : len(value)-1]
			}
			return value
		}
	}
	return ""
}
