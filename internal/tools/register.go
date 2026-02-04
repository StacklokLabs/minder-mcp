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
		mcp.WithDescription("Check Minder server health and connectivity. Returns server status and version information."),
		mcp.WithTitleAnnotation("Check Health"),
		mcp.WithReadOnlyHintAnnotation(true),
	), t.checkHealth)

	// Projects
	s.AddTool(mcp.NewTool("minder_list_projects",
		mcp.WithDescription("List all projects accessible to the current user. Returns project IDs, names, and metadata."),
		mcp.WithTitleAnnotation("List Projects"),
		mcp.WithReadOnlyHintAnnotation(true),
	), t.listProjects)

	s.AddTool(mcp.NewTool("minder_list_child_projects",
		mcp.WithDescription("List child projects nested under a parent project. Returns project hierarchy information."),
		mcp.WithTitleAnnotation("List Child Projects"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("UUID of the parent project to list children for"),
		),
	), t.listChildProjects)

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
	), t.listRepositories)

	s.AddTool(mcp.NewTool("minder_get_repository_by_id",
		mcp.WithDescription("Get detailed information about a specific repository by its UUID. "+
			"Returns full repository details including configuration and status."),
		mcp.WithTitleAnnotation("Get Repository by ID"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("repository_id",
			mcp.Required(),
			mcp.Title("Repository ID"),
			mcp.Description("UUID of the repository to retrieve"),
		),
	), t.getRepositoryByID)

	s.AddTool(mcp.NewTool("minder_get_repository_by_name",
		mcp.WithDescription("Get detailed information about a repository by its owner and name. "+
			"Returns full repository details including configuration and status."),
		mcp.WithTitleAnnotation("Get Repository by Name"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Title("Owner"),
			mcp.Description("Repository owner or organization name"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Title("Name"),
			mcp.Description("Repository name without the owner prefix"),
		),
		mcp.WithString("provider",
			mcp.Title("Provider"),
			mcp.Description("Provider name to scope the lookup (e.g., 'github')"),
		),
	), t.getRepositoryByName)

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
	), t.listProfiles)

	s.AddTool(mcp.NewTool("minder_get_profile_by_id",
		mcp.WithDescription("Get detailed information about a security profile by its UUID. "+
			"Returns full profile configuration including all rule definitions."),
		mcp.WithTitleAnnotation("Get Profile by ID"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("profile_id",
			mcp.Required(),
			mcp.Title("Profile ID"),
			mcp.Description("UUID of the profile to retrieve"),
		),
	), t.getProfileByID)

	s.AddTool(mcp.NewTool("minder_get_profile_by_name",
		mcp.WithDescription("Get detailed information about a security profile by its name. "+
			"Returns full profile configuration including all rule definitions."),
		mcp.WithTitleAnnotation("Get Profile by Name"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Title("Profile Name"),
			mcp.Description("Name of the profile to retrieve"),
		),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Project UUID to scope the lookup. Omit to search all accessible projects"),
		),
	), t.getProfileByName)

	s.AddTool(mcp.NewTool("minder_get_profile_status_by_name",
		mcp.WithDescription("Get the current evaluation status of a profile. "+
			"Returns compliance status across all entities including pass/fail counts."),
		mcp.WithTitleAnnotation("Get Profile Status"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Title("Profile Name"),
			mcp.Description("Name of the profile to check status for"),
		),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Project UUID to scope the lookup. Omit to search all accessible projects"),
		),
	), t.getProfileStatusByName)

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
	), t.listRuleTypes)

	s.AddTool(mcp.NewTool("minder_get_rule_type_by_id",
		mcp.WithDescription("Get detailed information about a rule type by its UUID. "+
			"Returns full rule definition including parameter schema and evaluation logic."),
		mcp.WithTitleAnnotation("Get Rule Type by ID"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("rule_type_id",
			mcp.Required(),
			mcp.Title("Rule Type ID"),
			mcp.Description("UUID of the rule type to retrieve"),
		),
	), t.getRuleTypeByID)

	s.AddTool(mcp.NewTool("minder_get_rule_type_by_name",
		mcp.WithDescription("Get detailed information about a rule type by its name. "+
			"Returns full rule definition including parameter schema and evaluation logic."),
		mcp.WithTitleAnnotation("Get Rule Type by Name"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Title("Rule Type Name"),
			mcp.Description("Name of the rule type to retrieve"),
		),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Project UUID to scope the lookup. Omit to search all accessible projects"),
		),
	), t.getRuleTypeByName)

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
	), t.listDataSources)

	s.AddTool(mcp.NewTool("minder_get_data_source_by_id",
		mcp.WithDescription("Get detailed information about a data source by its UUID. "+
			"Returns full data source configuration and connection details."),
		mcp.WithTitleAnnotation("Get Data Source by ID"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("data_source_id",
			mcp.Required(),
			mcp.Title("Data Source ID"),
			mcp.Description("UUID of the data source to retrieve"),
		),
	), t.getDataSourceByID)

	s.AddTool(mcp.NewTool("minder_get_data_source_by_name",
		mcp.WithDescription("Get detailed information about a data source by its name. "+
			"Returns full data source configuration and connection details."),
		mcp.WithTitleAnnotation("Get Data Source by Name"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Title("Data Source Name"),
			mcp.Description("Name of the data source to retrieve"),
		),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Project UUID to scope the lookup. Omit to search all accessible projects"),
		),
	), t.getDataSourceByName)

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
	), t.listProviders)

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
	), t.getProvider)

	// Users
	s.AddTool(mcp.NewTool("minder_get_user",
		mcp.WithDescription("Get information about the currently authenticated user. "+
			"Returns user ID, identity, and project memberships."),
		mcp.WithTitleAnnotation("Get Current User"),
		mcp.WithReadOnlyHintAnnotation(true),
	), t.getUser)

	s.AddTool(mcp.NewTool("minder_list_invitations",
		mcp.WithDescription("List pending project invitations for the current user. "+
			"Returns invitation details including project and role information."),
		mcp.WithTitleAnnotation("List Invitations"),
		mcp.WithReadOnlyHintAnnotation(true),
	), t.listInvitations)

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
	), t.listArtifacts)

	s.AddTool(mcp.NewTool("minder_get_artifact_by_id",
		mcp.WithDescription("Get detailed information about an artifact by its UUID. "+
			"Returns artifact metadata, versions, and vulnerability information."),
		mcp.WithTitleAnnotation("Get Artifact by ID"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("artifact_id",
			mcp.Required(),
			mcp.Title("Artifact ID"),
			mcp.Description("UUID of the artifact to retrieve"),
		),
	), t.getArtifactByID)

	s.AddTool(mcp.NewTool("minder_get_artifact_by_name",
		mcp.WithDescription("Get detailed information about an artifact by its name. "+
			"Returns artifact metadata, versions, and vulnerability information."),
		mcp.WithTitleAnnotation("Get Artifact by Name"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Title("Artifact Name"),
			mcp.Description("Full artifact name including registry path"),
		),
		mcp.WithString("provider",
			mcp.Title("Provider"),
			mcp.Description("Provider name to scope the lookup (e.g., 'github')"),
		),
	), t.getArtifactByName)

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
	), t.listEvaluationHistory)

	// Permissions
	s.AddTool(mcp.NewTool("minder_list_roles",
		mcp.WithDescription("List available roles that can be assigned to users. Returns role names and permission sets."),
		mcp.WithTitleAnnotation("List Roles"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Filter roles by project UUID. Omit to list from all accessible projects"),
		),
	), t.listRoles)

	s.AddTool(mcp.NewTool("minder_list_role_assignments",
		mcp.WithDescription("List role assignments showing which users have which roles. "+
			"Returns user identities and their assigned roles."),
		mcp.WithTitleAnnotation("List Role Assignments"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id",
			mcp.Title("Project ID"),
			mcp.Description("Filter role assignments by project UUID. Omit to list from all accessible projects"),
		),
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
