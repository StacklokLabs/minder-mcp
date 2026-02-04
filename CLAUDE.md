# Minder MCP Server

## Project Overview

MCP server exposing Minder's read-only operations via streaming HTTP transport. This server allows LLM-based tools to interact with Minder for security policy management, repository monitoring, and compliance evaluation.

## Build & Test Commands

- `task build` - Build binary
- `task lint` - Run golangci-lint
- `task test` - Run tests
- `task run` - Run server locally

## Architecture

- `cmd/minder-mcp/` - Entry point
- `internal/config/` - Environment configuration
- `internal/minder/` - gRPC client wrapper
- `internal/middleware/` - Auth token handling
- `internal/tools/` - MCP tool implementations

## Code Patterns

### Parameter Extraction

Tools use `req.GetString()` and `req.GetInt()` for parameter extraction with default values:

```go
projectID := req.GetString("project_id", "")
limit := req.GetInt("limit", 20)
```

### Error Handling

Errors use `mcp.NewToolResultError()` with `MapGRPCError()` for gRPC errors:

```go
if err != nil {
    return mcp.NewToolResultError(MapGRPCError(err)), nil
}
```

### Responses

All responses are JSON via `mcp.NewToolResultText()`:

```go
data, err := json.MarshalIndent(resp, "", "  ")
return mcp.NewToolResultText(string(data)), nil
```

## MCP Tool Conventions

- Names: `minder_<action>_<resource>` (snake_case)
- All tools are read-only
- Use `mcp.WithTitleAnnotation()` for display titles
- Use `mcp.WithReadOnlyHintAnnotation(true)` for all tools
- Use `mcp.Enum()` for constrained values
- Use `mcp.Title()` for parameter display names

## Git Workflow

- Never use `git add -A`
- Add specific files when staging
