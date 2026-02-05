// Package minder provides gRPC client utilities for connecting to Minder.
package minder

import "context"

// JWTTokenCredentials implements grpc.PerRPCCredentials for JWT token authentication.
type JWTTokenCredentials struct {
	token            string
	requireTransport bool
}

// NewJWTTokenCredentials creates a new JWTTokenCredentials with the given token.
// By default, transport security is required. Use WithInsecure() for development.
func NewJWTTokenCredentials(token string, insecure bool) *JWTTokenCredentials {
	return &JWTTokenCredentials{
		token:            token,
		requireTransport: !insecure,
	}
}

// GetRequestMetadata returns the authorization header with the JWT token.
func (j *JWTTokenCredentials) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + j.token,
	}, nil
}

// RequireTransportSecurity returns whether transport security (TLS) is required.
// Returns true by default for production safety; only returns false when explicitly
// configured for insecure connections during development.
func (j *JWTTokenCredentials) RequireTransportSecurity() bool {
	return j.requireTransport
}
