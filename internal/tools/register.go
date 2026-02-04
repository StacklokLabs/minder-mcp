package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/stacklok/minder-mcp/internal/config"
	"github.com/stacklok/minder-mcp/internal/middleware"
	"github.com/stacklok/minder-mcp/internal/minder"
)

// Tools holds the tool handlers and configuration.
type Tools struct {
	cfg *config.Config
}

// New creates a new Tools instance.
func New(cfg *config.Config) *Tools {
	return &Tools{cfg: cfg}
}

// Register registers all MCP tools with the server.
func (t *Tools) Register(s *server.MCPServer) {
	// Health
	s.AddTool(mcp.NewTool("minder_check_health",
		mcp.WithDescription("Check Minder server health status"),
		mcp.WithTitleAnnotation("Check Health"),
		mcp.WithReadOnlyHintAnnotation(true),
	), t.checkHealth)

	// Projects
	s.AddTool(mcp.NewTool("minder_list_projects",
		mcp.WithDescription("List all projects accessible to the current user"),
		mcp.WithTitleAnnotation("List Projects"),
		mcp.WithReadOnlyHintAnnotation(true),
	), t.listProjects)

	s.AddTool(mcp.NewTool("minder_list_child_projects",
		mcp.WithDescription("List child projects of a given project"),
		mcp.WithTitleAnnotation("List Child Projects"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id", mcp.Description("Parent project ID")),
	), t.listChildProjects)

	// Repositories
	s.AddTool(mcp.NewTool("minder_list_repositories",
		mcp.WithDescription("List repositories registered with Minder"),
		mcp.WithTitleAnnotation("List Repositories"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id", mcp.Description("Project ID (optional)")),
		mcp.WithString("provider", mcp.Description("Provider name filter (optional)")),
	), t.listRepositories)

	s.AddTool(mcp.NewTool("minder_get_repository_by_id",
		mcp.WithDescription("Get a repository by its ID"),
		mcp.WithTitleAnnotation("Get Repository by ID"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("repository_id", mcp.Required(), mcp.Description("Repository ID")),
	), t.getRepositoryByID)

	s.AddTool(mcp.NewTool("minder_get_repository_by_name",
		mcp.WithDescription("Get a repository by owner and name"),
		mcp.WithTitleAnnotation("Get Repository by Name"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("owner", mcp.Required(), mcp.Description("Repository owner")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithString("provider", mcp.Description("Provider name (optional)")),
	), t.getRepositoryByName)

	// Profiles
	s.AddTool(mcp.NewTool("minder_list_profiles",
		mcp.WithDescription("List all profiles"),
		mcp.WithTitleAnnotation("List Profiles"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id", mcp.Description("Project ID (optional)")),
		mcp.WithString("label_filter", mcp.Description("Label filter (optional)")),
	), t.listProfiles)

	s.AddTool(mcp.NewTool("minder_get_profile_by_id",
		mcp.WithDescription("Get a profile by its ID"),
		mcp.WithTitleAnnotation("Get Profile by ID"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("profile_id", mcp.Required(), mcp.Description("Profile ID")),
	), t.getProfileByID)

	s.AddTool(mcp.NewTool("minder_get_profile_by_name",
		mcp.WithDescription("Get a profile by name"),
		mcp.WithTitleAnnotation("Get Profile by Name"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("name", mcp.Required(), mcp.Description("Profile name")),
		mcp.WithString("project_id", mcp.Description("Project ID (optional)")),
	), t.getProfileByName)

	s.AddTool(mcp.NewTool("minder_get_profile_status_by_name",
		mcp.WithDescription("Get profile evaluation status by name"),
		mcp.WithTitleAnnotation("Get Profile Status"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("name", mcp.Required(), mcp.Description("Profile name")),
		mcp.WithString("project_id", mcp.Description("Project ID (optional)")),
	), t.getProfileStatusByName)

	// Rule Types
	s.AddTool(mcp.NewTool("minder_list_rule_types",
		mcp.WithDescription("List all rule types"),
		mcp.WithTitleAnnotation("List Rule Types"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id", mcp.Description("Project ID (optional)")),
	), t.listRuleTypes)

	s.AddTool(mcp.NewTool("minder_get_rule_type_by_id",
		mcp.WithDescription("Get a rule type by its ID"),
		mcp.WithTitleAnnotation("Get Rule Type by ID"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("rule_type_id", mcp.Required(), mcp.Description("Rule type ID")),
	), t.getRuleTypeByID)

	s.AddTool(mcp.NewTool("minder_get_rule_type_by_name",
		mcp.WithDescription("Get a rule type by name"),
		mcp.WithTitleAnnotation("Get Rule Type by Name"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("name", mcp.Required(), mcp.Description("Rule type name")),
		mcp.WithString("project_id", mcp.Description("Project ID (optional)")),
	), t.getRuleTypeByName)

	// Data Sources
	s.AddTool(mcp.NewTool("minder_list_data_sources",
		mcp.WithDescription("List all data sources"),
		mcp.WithTitleAnnotation("List Data Sources"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id", mcp.Description("Project ID (optional)")),
	), t.listDataSources)

	s.AddTool(mcp.NewTool("minder_get_data_source_by_id",
		mcp.WithDescription("Get a data source by its ID"),
		mcp.WithTitleAnnotation("Get Data Source by ID"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("data_source_id", mcp.Required(), mcp.Description("Data source ID")),
	), t.getDataSourceByID)

	s.AddTool(mcp.NewTool("minder_get_data_source_by_name",
		mcp.WithDescription("Get a data source by name"),
		mcp.WithTitleAnnotation("Get Data Source by Name"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("name", mcp.Required(), mcp.Description("Data source name")),
		mcp.WithString("project_id", mcp.Description("Project ID (optional)")),
	), t.getDataSourceByName)

	// Providers
	s.AddTool(mcp.NewTool("minder_list_providers",
		mcp.WithDescription("List all providers"),
		mcp.WithTitleAnnotation("List Providers"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id", mcp.Description("Project ID (optional)")),
	), t.listProviders)

	s.AddTool(mcp.NewTool("minder_get_provider",
		mcp.WithDescription("Get a provider by name"),
		mcp.WithTitleAnnotation("Get Provider"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("name", mcp.Required(), mcp.Description("Provider name")),
		mcp.WithString("project_id", mcp.Description("Project ID (optional)")),
	), t.getProvider)

	// Users
	s.AddTool(mcp.NewTool("minder_get_user",
		mcp.WithDescription("Get current user information"),
		mcp.WithTitleAnnotation("Get Current User"),
		mcp.WithReadOnlyHintAnnotation(true),
	), t.getUser)

	s.AddTool(mcp.NewTool("minder_list_invitations",
		mcp.WithDescription("List pending invitations for the current user"),
		mcp.WithTitleAnnotation("List Invitations"),
		mcp.WithReadOnlyHintAnnotation(true),
	), t.listInvitations)

	// Artifacts
	s.AddTool(mcp.NewTool("minder_list_artifacts",
		mcp.WithDescription("List artifacts"),
		mcp.WithTitleAnnotation("List Artifacts"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id", mcp.Description("Project ID (optional)")),
		mcp.WithString("provider", mcp.Description("Provider name filter (optional)")),
	), t.listArtifacts)

	s.AddTool(mcp.NewTool("minder_get_artifact_by_id",
		mcp.WithDescription("Get an artifact by its ID"),
		mcp.WithTitleAnnotation("Get Artifact by ID"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("artifact_id", mcp.Required(), mcp.Description("Artifact ID")),
	), t.getArtifactByID)

	s.AddTool(mcp.NewTool("minder_get_artifact_by_name",
		mcp.WithDescription("Get an artifact by name"),
		mcp.WithTitleAnnotation("Get Artifact by Name"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("name", mcp.Required(), mcp.Description("Artifact name")),
		mcp.WithString("provider", mcp.Description("Provider name (optional)")),
	), t.getArtifactByName)

	// Evaluation Results
	s.AddTool(mcp.NewTool("minder_list_evaluation_history",
		mcp.WithDescription("List evaluation history"),
		mcp.WithTitleAnnotation("List Evaluation History"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id", mcp.Description("Project ID (optional)")),
		mcp.WithString("profile_name", mcp.Description("Profile name filter (optional)")),
		mcp.WithString("entity_type", mcp.Description("Entity type filter (optional): repository, artifact, pull_request")),
		mcp.WithString("entity_name", mcp.Description("Entity name filter (optional)")),
		mcp.WithString("evaluation_status", mcp.Description("Status filter (optional): success, failure, error, skipped, pending")),
		mcp.WithString("remediation_status", mcp.Description("Remediation status filter (optional)")),
		mcp.WithString("alert_status", mcp.Description("Alert status filter (optional)")),
		mcp.WithString("from", mcp.Description("Start time filter (RFC3339 format, optional)")),
		mcp.WithString("to", mcp.Description("End time filter (RFC3339 format, optional)")),
	), t.listEvaluationHistory)

	// Permissions
	s.AddTool(mcp.NewTool("minder_list_roles",
		mcp.WithDescription("List available roles"),
		mcp.WithTitleAnnotation("List Roles"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id", mcp.Description("Project ID (optional)")),
	), t.listRoles)

	s.AddTool(mcp.NewTool("minder_list_role_assignments",
		mcp.WithDescription("List role assignments"),
		mcp.WithTitleAnnotation("List Role Assignments"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id", mcp.Description("Project ID (optional)")),
	), t.listRoleAssignments)
}

// getClient creates a new Minder client using the token from context.
func (t *Tools) getClient(ctx context.Context) (*minder.Client, error) {
	token := middleware.TokenFromContext(ctx)

	return minder.NewClient(minder.ClientConfig{
		Host:     t.cfg.Minder.Host,
		Port:     t.cfg.Minder.Port,
		Insecure: t.cfg.Minder.Insecure,
		Token:    token,
	})
}
