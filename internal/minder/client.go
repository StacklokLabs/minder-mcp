package minder

import (
	"crypto/tls"
	"fmt"

	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps a gRPC connection and provides access to Minder service clients.
type Client struct {
	conn *grpc.ClientConn
}

// ClientConfig holds configuration for creating a Minder client.
type ClientConfig struct {
	Host     string
	Port     int
	Insecure bool
	Token    string
}

// NewClient creates a new Minder gRPC client.
func NewClient(cfg ClientConfig) (*Client, error) {
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	opts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(NewJWTTokenCredentials(cfg.Token)),
	}

	// Add transport credentials
	if cfg.Insecure || cfg.Host == "localhost" || cfg.Host == "127.0.0.1" || cfg.Host == "::1" {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			MinVersion: tls.VersionTLS13,
			ServerName: cfg.Host,
		})))
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Minder: %w", err)
	}

	return &Client{conn: conn}, nil
}

// Close closes the gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Health returns the HealthServiceClient.
func (c *Client) Health() minderv1.HealthServiceClient {
	return minderv1.NewHealthServiceClient(c.conn)
}

// Repositories returns the RepositoryServiceClient.
func (c *Client) Repositories() minderv1.RepositoryServiceClient {
	return minderv1.NewRepositoryServiceClient(c.conn)
}

// Profiles returns the ProfileServiceClient.
func (c *Client) Profiles() minderv1.ProfileServiceClient {
	return minderv1.NewProfileServiceClient(c.conn)
}

// RuleTypes returns the RuleTypeServiceClient.
func (c *Client) RuleTypes() minderv1.RuleTypeServiceClient {
	return minderv1.NewRuleTypeServiceClient(c.conn)
}

// DataSources returns the DataSourceServiceClient.
func (c *Client) DataSources() minderv1.DataSourceServiceClient {
	return minderv1.NewDataSourceServiceClient(c.conn)
}

// Providers returns the ProvidersServiceClient.
func (c *Client) Providers() minderv1.ProvidersServiceClient {
	return minderv1.NewProvidersServiceClient(c.conn)
}

// Projects returns the ProjectsServiceClient.
func (c *Client) Projects() minderv1.ProjectsServiceClient {
	return minderv1.NewProjectsServiceClient(c.conn)
}

// Users returns the UserServiceClient.
func (c *Client) Users() minderv1.UserServiceClient {
	return minderv1.NewUserServiceClient(c.conn)
}

// Artifacts returns the ArtifactServiceClient.
func (c *Client) Artifacts() minderv1.ArtifactServiceClient {
	return minderv1.NewArtifactServiceClient(c.conn)
}

// EvalResults returns the EvalResultsServiceClient.
func (c *Client) EvalResults() minderv1.EvalResultsServiceClient {
	return minderv1.NewEvalResultsServiceClient(c.conn)
}

// Permissions returns the PermissionsServiceClient.
func (c *Client) Permissions() minderv1.PermissionsServiceClient {
	return minderv1.NewPermissionsServiceClient(c.conn)
}

// Invites returns the InviteServiceClient.
func (c *Client) Invites() minderv1.InviteServiceClient {
	return minderv1.NewInviteServiceClient(c.conn)
}
