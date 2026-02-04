# Minder MCP Server

An MCP (Model Context Protocol) server that exposes Minder's read operations via streaming HTTP transport.

## Features

- Read-only access to Minder resources through MCP tools
- Supports authentication via HTTP header or environment variable
- Streaming HTTP transport with heartbeat support

## Installation

```bash
go install github.com/stacklok/minder-mcp/cmd/minder-mcp@latest
```

Or build from source:

```bash
task build
```

## Configuration

The server is configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `MINDER_AUTH_TOKEN` | Static auth token (fallback) | - |
| `MINDER_SERVER_HOST` | Minder GRPC host | `api.stacklok.com` |
| `MINDER_SERVER_PORT` | Minder GRPC port | `443` |
| `MINDER_INSECURE` | Allow insecure connections | `false` |
| `MCP_PORT` | MCP HTTP server port | `8080` |
| `MCP_ENDPOINT_PATH` | MCP endpoint path | `/mcp` |

## Authentication

The server supports two authentication methods (in priority order):

1. **Authorization Header**: Pass a Bearer token in the HTTP `Authorization` header
2. **Environment Variable**: Set `MINDER_AUTH_TOKEN` as a fallback

## Available Tools

### Health
- `minder_check_health` - Check Minder server health status

### Projects
- `minder_list_projects` - List all projects accessible to the current user
- `minder_list_child_projects` - List child projects of a given project

### Repositories
- `minder_list_repositories` - List repositories registered with Minder
- `minder_get_repository_by_id` - Get a repository by its ID
- `minder_get_repository_by_name` - Get a repository by owner and name

### Profiles
- `minder_list_profiles` - List all profiles
- `minder_get_profile_by_id` - Get a profile by its ID
- `minder_get_profile_by_name` - Get a profile by name
- `minder_get_profile_status_by_name` - Get profile evaluation status

### Rule Types
- `minder_list_rule_types` - List all rule types
- `minder_get_rule_type_by_id` - Get a rule type by its ID
- `minder_get_rule_type_by_name` - Get a rule type by name

### Data Sources
- `minder_list_data_sources` - List all data sources
- `minder_get_data_source_by_id` - Get a data source by its ID
- `minder_get_data_source_by_name` - Get a data source by name

### Providers
- `minder_list_providers` - List all providers
- `minder_get_provider` - Get a provider by name

### Users
- `minder_get_user` - Get current user information
- `minder_list_invitations` - List pending invitations

### Artifacts
- `minder_list_artifacts` - List artifacts
- `minder_get_artifact_by_id` - Get an artifact by its ID
- `minder_get_artifact_by_name` - Get an artifact by name

### Evaluation Results
- `minder_list_evaluation_history` - List evaluation history with optional filters

### Permissions
- `minder_list_roles` - List available roles
- `minder_list_role_assignments` - List role assignments

## Usage

### Running the Server

```bash
# Using environment variable for auth
export MINDER_AUTH_TOKEN="your-token-here"
task run

# Or run the binary directly
./bin/minder-mcp
```

### Development

```bash
# Run linter
task lint

# Run tests
task test

# Run tests with coverage
task test:coverage

# Format code
task fmt

# Build, lint, and test
task all
```

## License

Apache 2.0
