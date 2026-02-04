// Package minder provides GRPC client utilities for connecting to Minder.
package minder

import "context"

// JWTTokenCredentials implements grpc.PerRPCCredentials for JWT token authentication.
type JWTTokenCredentials struct {
	token string
}

// NewJWTTokenCredentials creates a new JWTTokenCredentials with the given token.
func NewJWTTokenCredentials(token string) *JWTTokenCredentials {
	return &JWTTokenCredentials{token: token}
}

// GetRequestMetadata returns the authorization header with the JWT token.
func (j *JWTTokenCredentials) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + j.token,
	}, nil
}

// RequireTransportSecurity returns false to allow insecure connections for local development.
func (*JWTTokenCredentials) RequireTransportSecurity() bool {
	return false
}
