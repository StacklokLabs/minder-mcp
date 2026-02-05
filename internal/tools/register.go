package tools

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/stacklok/minder-mcp/internal/config"
	"github.com/stacklok/minder-mcp/internal/middleware"
	"github.com/stacklok/minder-mcp/internal/minder"
	"github.com/stacklok/minder-mcp/internal/resources"
)

// Tools holds the tool handlers and configuration.
type Tools struct {
	cfg            *config.Config
	clientFactory  ClientFactory
	logger         *slog.Logger
	tokenRefresher *minder.TokenRefresher
}

// New creates a new Tools instance with the default client factory.
func New(cfg *config.Config, logger *slog.Logger) *Tools {
	t := &Tools{
		cfg:            cfg,
		logger:         logger,
		tokenRefresher: minder.NewTokenRefresher(),
	}
	t.clientFactory = t.defaultClientFactory
	return t
}

// NewWithClientFactory creates a new Tools instance with a custom client factory.
// This is useful for testing with mock clients.
func NewWithClientFactory(cfg *config.Config, logger *slog.Logger, factory ClientFactory) *Tools {
	return &Tools{
		cfg:           cfg,
		clientFactory: factory,
		logger:        logger,
		// tokenRefresher not needed when using custom factory (e.g., for tests)
	}
}

// Close releases resources held by the Tools instance.
func (t *Tools) Close() {
	if t.tokenRefresher != nil {
		t.tokenRefresher.Close()
	}
}

// wrapHandler wraps a tool handler with debug logging.
func (t *Tools) wrapHandler(name string, handler server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		start := time.Now()
		t.logger.DebugContext(ctx, "tool invoked", "tool", name, "params", req.Params.Arguments)
		result, err := handler(ctx, req)
		hasError := err != nil
		t.logger.DebugContext(ctx, "tool completed", "tool", name, "duration", time.Since(start), "error", hasError)
		return result, err
	}
}

// Register registers all MCP tools with the server.
func (t *Tools) Register(s *server.MCPServer) {
	// Projects
	s.AddTool(mcp.NewTool("minder_list_projects",
		mcp.WithDescription("List projects accessible to the current user. "+
			"If project_id is provided, lists child projects of that project. "+
			"Otherwise, lists all top-level accessible projects."),
		mcp.WithTitleAnnotation("List Projects"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("UUID of a parent project to list children for. Omit to list all accessible projects"),
		),
	), t.wrapHandler("minder_list_projects", t.listProjects))

	// Repositories
	s.AddTool(mcp.NewTool("minder_list_repositories",
		mcp.WithDescription("List repositories registered with Minder. "+
			"Returns repository details including ID, name, owner, provider, and registration status. "+
			"Supports cursor-based pagination."),
		mcp.WithTitleAnnotation("List Repositories"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Filter repositories by project UUID. Omit to list from all accessible projects"),
		),
		mcp.WithString("provider",
			mcp.Title("Provider"),
			mcp.Description("Filter repositories by provider name (e.g., 'github')"),
		),
		mcp.WithString("cursor",
			mcp.Title("Pagination Cursor"),
			mcp.Description("Cursor from previous response for pagination. Omit for first page"),
		),
		mcp.WithNumber("limit",
			mcp.Title("Page Size"),
			mcp.Description("Maximum number of results per page (1-100)"),
			mcp.Min(1),
			mcp.Max(100),
		),
	), t.wrapHandler("minder_list_repositories", t.listRepositories))

	s.AddTool(mcp.NewTool("minder_get_repository",
		mcp.WithDescription("Get a repository by ID or owner/name. "+
			"Use repository_id for UUID lookup, or provide both owner and name for name lookup."),
		mcp.WithTitleAnnotation("Get Repository"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("repository_id",
			mcp.Title("Repository ID"),
			mcp.Description("UUID of the repository. Mutually exclusive with owner/name"),
		),
		mcp.WithString("owner",
			mcp.Title("Owner"),
			mcp.Description("Repository owner or organization. Required with name for name lookup"),
		),
		mcp.WithString("name",
			mcp.Title("Name"),
			mcp.Description("Repository name without owner prefix. Required with owner for name lookup"),
		),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Project scope. Only valid with owner/name lookup"),
		),
		mcp.WithString("provider",
			mcp.Title("Provider"),
			mcp.Description("Provider filter. Only valid with owner/name lookup"),
		),
	), t.wrapHandler("minder_get_repository", t.getRepository))

	// Profiles
	s.AddTool(mcp.NewTool("minder_list_profiles",
		mcp.WithDescription("List security profiles configured in Minder. "+
			"Returns profile names, IDs, and associated rule configurations."),
		mcp.WithTitleAnnotation("List Profiles"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Filter profiles by project UUID. Omit to list from all accessible projects"),
		),
		mcp.WithString("label_filter",
			mcp.Title("Label Filter"),
			mcp.Description("Filter profiles by label selector expression"),
		),
	), t.wrapHandler("minder_list_profiles", t.listProfiles))

	s.AddTool(mcp.NewTool("minder_get_profile",
		mcp.WithDescription("Get a security profile by ID or name. "+
			"Use profile_id for UUID lookup, or name for name lookup."),
		mcp.WithTitleAnnotation("Get Profile"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("profile_id",
			mcp.Title("Profile ID"),
			mcp.Description("UUID of the profile. Mutually exclusive with name"),
		),
		mcp.WithString("name",
			mcp.Title("Profile Name"),
			mcp.Description("Name of the profile. Mutually exclusive with profile_id"),
		),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Project scope. Only valid with name lookup"),
		),
	), t.wrapHandler("minder_get_profile", t.getProfile))

	s.AddTool(mcp.NewTool("minder_get_profile_status",
		mcp.WithDescription("Get the current evaluation status of a profile by ID or name. "+
			"Use profile_id for UUID lookup, or name for name lookup. "+
			"Returns compliance status and detailed per-rule evaluation results for all entities."),
		mcp.WithTitleAnnotation("Get Profile Status"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("profile_id",
			mcp.Title("Profile ID"),
			mcp.Description("UUID of the profile. Mutually exclusive with name"),
		),
		mcp.WithString("name",
			mcp.Title("Profile Name"),
			mcp.Description("Name of the profile. Mutually exclusive with profile_id"),
		),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Project scope. Only valid with name lookup"),
		),
	), t.wrapHandler("minder_get_profile_status", t.getProfileStatus))

	// Rule Types
	s.AddTool(mcp.NewTool("minder_list_rule_types",
		mcp.WithDescription("List available rule types that can be used in profiles. "+
			"Returns rule type names, descriptions, and parameter schemas."),
		mcp.WithTitleAnnotation("List Rule Types"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Filter rule types by project UUID. Omit to list from all accessible projects"),
		),
	), t.wrapHandler("minder_list_rule_types", t.listRuleTypes))

	s.AddTool(mcp.NewTool("minder_get_rule_type",
		mcp.WithDescription("Get a rule type by ID or name. "+
			"Use rule_type_id for UUID lookup, or name for name lookup."),
		mcp.WithTitleAnnotation("Get Rule Type"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("rule_type_id",
			mcp.Title("Rule Type ID"),
			mcp.Description("UUID of the rule type. Mutually exclusive with name"),
		),
		mcp.WithString("name",
			mcp.Title("Rule Type Name"),
			mcp.Description("Name of the rule type. Mutually exclusive with rule_type_id"),
		),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Project scope. Only valid with name lookup"),
		),
	), t.wrapHandler("minder_get_rule_type", t.getRuleType))

	// Data Sources
	s.AddTool(mcp.NewTool("minder_list_data_sources",
		mcp.WithDescription("List data sources available for rule evaluations. "+
			"Returns data source names, types, and configuration details."),
		mcp.WithTitleAnnotation("List Data Sources"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Filter data sources by project UUID. Omit to list from all accessible projects"),
		),
	), t.wrapHandler("minder_list_data_sources", t.listDataSources))

	s.AddTool(mcp.NewTool("minder_get_data_source",
		mcp.WithDescription("Get a data source by ID or name. "+
			"Use data_source_id for UUID lookup, or name for name lookup."),
		mcp.WithTitleAnnotation("Get Data Source"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("data_source_id",
			mcp.Title("Data Source ID"),
			mcp.Description("UUID of the data source. Mutually exclusive with name"),
		),
		mcp.WithString("name",
			mcp.Title("Data Source Name"),
			mcp.Description("Name of the data source. Mutually exclusive with data_source_id"),
		),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Project scope. Only valid with name lookup"),
		),
	), t.wrapHandler("minder_get_data_source", t.getDataSource))

	// Providers
	s.AddTool(mcp.NewTool("minder_list_providers",
		mcp.WithDescription("List configured providers (e.g., GitHub, GitLab). "+
			"Returns provider names, types, and connection status. Supports cursor-based pagination."),
		mcp.WithTitleAnnotation("List Providers"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Filter providers by project UUID. Omit to list from all accessible projects"),
		),
		mcp.WithString("cursor",
			mcp.Title("Pagination Cursor"),
			mcp.Description("Cursor from previous response for pagination. Omit for first page"),
		),
		mcp.WithNumber("limit",
			mcp.Title("Page Size"),
			mcp.Description("Maximum number of results per page (1-100)"),
			mcp.Min(1),
			mcp.Max(100),
		),
	), t.wrapHandler("minder_list_providers", t.listProviders))

	s.AddTool(mcp.NewTool("minder_get_provider",
		mcp.WithDescription("Get detailed information about a provider by its name. Returns provider configuration and capabilities."),
		mcp.WithTitleAnnotation("Get Provider"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Title("Provider Name"),
			mcp.Description("Name of the provider to retrieve (e.g., 'github')"),
		),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Project UUID to scope the lookup. Omit to search all accessible projects"),
		),
	), t.wrapHandler("minder_get_provider", t.getProvider))

	// Artifacts
	s.AddTool(mcp.NewTool("minder_list_artifacts",
		mcp.WithDescription("List artifacts (container images, packages) tracked by Minder. "+
			"Returns artifact names, versions, and associated repositories."),
		mcp.WithTitleAnnotation("List Artifacts"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Filter artifacts by project UUID. Omit to list from all accessible projects"),
		),
		mcp.WithString("provider",
			mcp.Title("Provider"),
			mcp.Description("Filter artifacts by provider name (e.g., 'github')"),
		),
	), t.wrapHandler("minder_list_artifacts", t.listArtifacts))

	s.AddTool(mcp.NewTool("minder_get_artifact",
		mcp.WithDescription("Get an artifact by ID or name. "+
			"Use artifact_id for UUID lookup, or name for name lookup."),
		mcp.WithTitleAnnotation("Get Artifact"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("artifact_id",
			mcp.Title("Artifact ID"),
			mcp.Description("UUID of the artifact. Mutually exclusive with name"),
		),
		mcp.WithString("name",
			mcp.Title("Artifact Name"),
			mcp.Description("Full artifact name including registry path. Mutually exclusive with artifact_id"),
		),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Project scope. Only valid with name lookup"),
		),
		mcp.WithString("provider",
			mcp.Title("Provider"),
			mcp.Description("Provider filter. Only valid with name lookup"),
		),
	), t.wrapHandler("minder_get_artifact", t.getArtifact))

	// Evaluation Results
	s.AddTool(mcp.NewTool("minder_list_evaluation_history",
		mcp.WithDescription("List historical evaluation results for profile rules. "+
			"Returns evaluation timestamps, statuses, and entity details with filtering support. "+
			"Supports cursor-based pagination with total record count."),
		mcp.WithTitleAnnotation("List Evaluation History"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Filter by project UUID. Omit to list from all accessible projects"),
		),
		mcp.WithString("profile_name",
			mcp.Title("Profile Name"),
			mcp.Description("Filter evaluations by profile name"),
		),
		mcp.WithString("entity_type",
			mcp.Title("Entity Type"),
			mcp.Description("Filter by the type of entity that was evaluated"),
			mcp.Enum("repository", "artifact", "pull_request"),
		),
		mcp.WithString("entity_name",
			mcp.Title("Entity Name"),
			mcp.Description("Filter evaluations by entity name"),
		),
		mcp.WithString("evaluation_status",
			mcp.Title("Evaluation Status"),
			mcp.Description("Filter by evaluation result status"),
			mcp.Enum("success", "failure", "error", "skipped", "pending"),
		),
		mcp.WithString("remediation_status",
			mcp.Title("Remediation Status"),
			mcp.Description("Filter by auto-remediation status"),
			mcp.Enum("success", "failure", "error", "skipped", "not_available", "pending"),
		),
		mcp.WithString("alert_status",
			mcp.Title("Alert Status"),
			mcp.Description("Filter by alert notification status"),
			mcp.Enum("on", "off", "error", "skipped", "not_available"),
		),
		mcp.WithString("from",
			mcp.Title("From Time"),
			mcp.Description("Start of time range filter in RFC3339 format (e.g., 2024-01-15T09:00:00Z)"),
		),
		mcp.WithString("to",
			mcp.Title("To Time"),
			mcp.Description("End of time range filter in RFC3339 format (e.g., 2024-01-15T17:00:00Z)"),
		),
		mcp.WithString("cursor",
			mcp.Title("Pagination Cursor"),
			mcp.Description("Cursor from previous response for pagination. Omit for first page"),
		),
		mcp.WithNumber("page_size",
			mcp.Title("Page Size"),
			mcp.Description("Number of results per page (1-100)"),
			mcp.Min(1),
			mcp.Max(100),
		),
		mcp.WithString("label_filter",
			mcp.Title("Label Filter"),
			mcp.Description("Filter by profile labels. '*' includes all (default), "+
				"empty for unlabeled only. Prefix with '!' to exclude (e.g., '!system')."),
		),
	), t.wrapHandler("minder_list_evaluation_history", t.listEvaluationHistory))

	// Dashboard - includes _meta.ui.resourceUri for MCP Apps support
	dashboardTool := mcp.NewTool("minder_show_dashboard",
		mcp.WithDescription("Display the Minder Compliance Dashboard - an interactive visual interface "+
			"showing repository security posture, profile compliance status, and evaluation history "+
			"across all monitored repositories"),
		mcp.WithTitleAnnotation("Show Compliance Dashboard"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	dashboardTool.Meta = mcp.NewMetaFromMap(map[string]any{
		"ui": map[string]any{
			"resourceUri": resources.DashboardURI,
		},
	})
	s.AddTool(dashboardTool, t.wrapHandler("minder_show_dashboard", t.showComplianceDashboard))
}

// getClient returns a MinderClient using the configured factory.
func (t *Tools) getClient(ctx context.Context) (MinderClient, error) {
	return t.clientFactory(ctx)
}

// defaultClientFactory creates a real Minder client using the token from context.
// If the token is an offline/refresh token or expired, it will be refreshed automatically.
func (t *Tools) defaultClientFactory(ctx context.Context) (MinderClient, error) {
	token := middleware.TokenFromContext(ctx)

	// Log token status for debugging
	if token == "" {
		t.logger.WarnContext(ctx, "no authentication token provided",
			"hint", "set MINDER_AUTH_TOKEN or pass Authorization header")
		return nil, fmt.Errorf("no authentication token: set MINDER_AUTH_TOKEN environment variable or pass Authorization header")
	}

	serverCfg := minder.ServerConfig{
		Host:     t.cfg.Minder.Host,
		Port:     t.cfg.Minder.Port,
		Insecure: t.cfg.Minder.Insecure,
	}

	// Validate and potentially refresh the token
	validToken, err := t.tokenRefresher.GetValidAccessToken(ctx, token, serverCfg)
	if err != nil {
		t.logger.ErrorContext(ctx, "token validation failed",
			"error", err,
			"server_host", t.cfg.Minder.Host,
			"server_port", t.cfg.Minder.Port,
			"insecure", t.cfg.Minder.Insecure,
		)
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	t.logger.DebugContext(ctx, "token validated successfully")

	return minder.NewClient(minder.ClientConfig{
		Host:     t.cfg.Minder.Host,
		Port:     t.cfg.Minder.Port,
		Insecure: t.cfg.Minder.Insecure,
		Token:    validToken,
	})
}
