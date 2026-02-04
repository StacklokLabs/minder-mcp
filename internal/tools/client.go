package tools

import (
	"context"

	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"

	"github.com/stacklok/minder-mcp/internal/minder"
)

// MinderClient defines the interface for Minder client operations used by tools.
type MinderClient interface {
	Close() error
	Health() minderv1.HealthServiceClient
	Profiles() minderv1.ProfileServiceClient
	Repositories() minderv1.RepositoryServiceClient
	RuleTypes() minderv1.RuleTypeServiceClient
	DataSources() minderv1.DataSourceServiceClient
	Providers() minderv1.ProvidersServiceClient
	Projects() minderv1.ProjectsServiceClient
	Artifacts() minderv1.ArtifactServiceClient
	EvalResults() minderv1.EvalResultsServiceClient
}

// ClientFactory creates MinderClient instances.
type ClientFactory func(ctx context.Context) (MinderClient, error)

// Verify that minder.Client implements MinderClient at compile time.
var _ MinderClient = (*minder.Client)(nil)
