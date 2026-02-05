// Package minder provides gRPC client utilities for connecting to Minder.
package minder

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// DefaultClientID is the OAuth2 client ID used for token refresh.
	DefaultClientID = "minder-cli"

	// tokenRefreshBuffer is how far in advance of expiry to refresh.
	// Using 60 seconds to account for network latency, clock skew, and multi-step operations.
	tokenRefreshBuffer = 60 * time.Second

	// offlineTokenType is the Keycloak-specific token type claim for offline/refresh tokens.
	offlineTokenType = "Offline"

	// maxRedirects is the maximum number of HTTP redirects to follow.
	maxRedirects = 3

	// maxWWWAuthHeaderLen is the maximum length of WWW-Authenticate header to parse.
	maxWWWAuthHeaderLen = 2048
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
			CheckRedirect: func(_ *http.Request, via []*http.Request) error {
				if len(via) >= maxRedirects {
					return errors.New("too many redirects")
				}
				return nil
			},
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS13,
				},
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
	// Create a cache key from the refresh token using SHA256 hash
	// This avoids storing the full token and prevents cache collisions
	cacheKey := hashToken(refreshToken)

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

// hashToken creates a SHA256 hash of a token for use as a cache key.
// This avoids storing the full token in memory and prevents cache collisions.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:16]) // Use first 16 bytes (128 bits)
}

// validateRealmURL validates that the discovered realm URL is trusted.
// This prevents SSRF attacks where a malicious server could redirect token requests.
func (*TokenRefresher) validateRealmURL(realmURL string, expectedHost string) error {
	parsedRealm, err := url.Parse(realmURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Validate scheme - require HTTPS except for localhost
	host := parsedRealm.Hostname()
	if parsedRealm.Scheme == "http" {
		if !isLocalhostHost(host) {
			return fmt.Errorf("http scheme only allowed for localhost, got host: %s", host)
		}
	} else if parsedRealm.Scheme != "https" {
		return fmt.Errorf("invalid scheme: %s (https required)", parsedRealm.Scheme)
	}

	// Reject URLs with embedded credentials (userinfo)
	if parsedRealm.User != nil {
		return errors.New("realm URL must not contain credentials")
	}

	// Validate host is present
	if host == "" {
		return errors.New("realm URL must have a valid host")
	}

	// Block private/reserved IP addresses to prevent SSRF
	// Skip this check for localhost (allowed for development)
	if !isLocalhostHost(host) && isPrivateOrReservedHost(host) {
		return fmt.Errorf("realm URL points to private/reserved address: %s", host)
	}

	// Verify the realm URL host is related to the expected server host.
	// This provides defense-in-depth by ensuring the auth server is in a
	// similar domain to the Minder server (e.g., both under stacklok.com).
	if !isRelatedHost(host, expectedHost) {
		return fmt.Errorf("realm host %q not related to expected host %q", host, expectedHost)
	}

	return nil
}

// isLocalhostHost checks if a host is localhost or a loopback address.
func isLocalhostHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// isPrivateOrReservedHost checks if a host resolves to a private or reserved IP address.
func isPrivateOrReservedHost(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		// It's a hostname, not an IP - we allow hostnames as they will be
		// resolved by the HTTP client and validated by TLS certificate verification.
		// Direct IP addresses are more dangerous for SSRF.
		return false
	}

	// Block loopback, private, link-local, and other reserved ranges
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified()
}

// isRelatedHost checks if the realm host is related to the expected server host.
// This provides defense-in-depth by ensuring auth servers are in similar domains.
func isRelatedHost(realmHost, expectedHost string) bool {
	// Normalize hosts to lowercase
	realmHost = strings.ToLower(realmHost)
	expectedHost = strings.ToLower(expectedHost)

	// Extract base domains (last two parts for common TLDs)
	realmBase := getBaseDomain(realmHost)
	expectedBase := getBaseDomain(expectedHost)

	// Allow if base domains match (e.g., auth.stacklok.com and api.stacklok.com)
	if realmBase == expectedBase && realmBase != "" {
		return true
	}

	// Allow localhost for development
	if isLocalhostHost(realmHost) && isLocalhostHost(expectedHost) {
		return true
	}

	// Allow if realm host contains expected host or vice versa
	// This handles cases like expected=api.example.com, realm=auth.example.com
	if strings.HasSuffix(realmHost, "."+expectedBase) || strings.HasSuffix(expectedHost, "."+realmBase) {
		return true
	}

	return false
}

// getBaseDomain extracts the base domain (last two parts) from a hostname.
// For example: "auth.stacklok.com" -> "stacklok.com"
func getBaseDomain(host string) string {
	// Remove port if present
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}

	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return host
	}
	return strings.Join(parts[len(parts)-2:], ".")
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

// discoverRealmURL discovers the Keycloak realm URL from the server's www-authenticate gRPC metadata.
func (*TokenRefresher) discoverRealmURL(ctx context.Context, cfg ServerConfig) (string, error) {
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Set up dial options (TLS or insecure)
	var opts []grpc.DialOption
	if cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			MinVersion: tls.VersionTLS13,
			ServerName: cfg.Host,
		})))
	}

	// Create unauthenticated connection
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// Call GetUser to trigger auth error with realm URL
	client := minderv1.NewUserServiceClient(conn)
	var headers metadata.MD
	_, err = client.GetUser(ctx, &minderv1.GetUserRequest{}, grpc.Header(&headers))

	// We expect Unauthenticated error
	if status.Code(err) != codes.Unauthenticated {
		return "", fmt.Errorf("unexpected response: %w", err)
	}

	// Extract realm from www-authenticate header
	wwwAuth := headers.Get("www-authenticate")
	if len(wwwAuth) == 0 || wwwAuth[0] == "" {
		return "", errors.New("server did not return www-authenticate header")
	}

	realmURL := extractRealmFromWWWAuthenticate(wwwAuth[0])
	if realmURL == "" {
		return "", errors.New("could not parse realm from www-authenticate header")
	}

	return realmURL, nil
}

// extractRealmFromWWWAuthenticate parses the realm URL from a WWW-Authenticate header.
// Example header: Bearer realm="https://auth.stacklok.com/realms/stacklok"
func extractRealmFromWWWAuthenticate(header string) string {
	// Limit header length to prevent DoS
	if len(header) > maxWWWAuthHeaderLen {
		return ""
	}

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
