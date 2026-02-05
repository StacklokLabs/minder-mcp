# Minder MCP Server

An MCP (Model Context Protocol) server that exposes Minder's read operations via streaming HTTP transport.

## Features

- Read-only access to Minder resources through MCP tools
- Supports authentication via HTTP header or environment variable
- Streaming HTTP transport with heartbeat support
- **Compliance Dashboard**: Interactive UI served as an MCP resource for MCP Apps-enabled clients

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
| `MINDER_SERVER_HOST` | Minder GRPC host (required) | `` |
| `MINDER_SERVER_PORT` | Minder GRPC port | `443` |
| `MINDER_INSECURE` | Allow insecure connections | `false` |
| `MCP_PORT` | MCP HTTP server port | `8080` |
| `MCP_ENDPOINT_PATH` | MCP endpoint path | `/mcp` |
| `LOG_LEVEL` | logging level | `info` |

## Authentication

The server supports two authentication methods (in priority order):

1. **Authorization Header**: Pass a Bearer token in the HTTP `Authorization` header
2. **Environment Variable**: Set `MINDER_AUTH_TOKEN` as a fallback

## Available Tools

### Projects
- `minder_list_projects` - List projects (all accessible, or children of a specific project)

### Repositories
- `minder_list_repositories` - List repositories registered with Minder
- `minder_get_repository` - Get a repository by ID or owner/name

### Profiles
- `minder_list_profiles` - List all profiles
- `minder_get_profile` - Get a profile by ID or name
- `minder_get_profile_status` - Get profile evaluation status by ID or name

### Rule Types
- `minder_list_rule_types` - List all rule types
- `minder_get_rule_type` - Get a rule type by ID or name

### Data Sources
- `minder_list_data_sources` - List all data sources
- `minder_get_data_source` - Get a data source by ID or name

### Providers
- `minder_list_providers` - List all providers
- `minder_get_provider` - Get a provider by name

### Artifacts
- `minder_list_artifacts` - List artifacts
- `minder_get_artifact` - Get an artifact by ID or name

### Evaluation Results
- `minder_list_evaluation_history` - List evaluation history with optional filters

## Resources

### Compliance Dashboard
- **URI**: `ui://minder/compliance-dashboard`
- **MIME Type**: `text/html`

An interactive compliance dashboard that displays real-time compliance status across repositories with drill-down capabilities. The dashboard is designed for MCP Apps-enabled clients (VS Code, Claude, etc.) and provides:

- Summary cards showing total repositories, passing/failing counts, and compliance rate
- Profile list with expandable rule evaluation details
- Repository list with filtering
- Real-time data from Minder via MCP tools

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

# Build UI only
task build:ui

# Lint UI (TypeScript)
task lint:ui

# Lint everything (Go + UI)
task lint:all
```

## License

Apache 2.0
