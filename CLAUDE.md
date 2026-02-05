# Minder MCP Server

## Project Overview

MCP server exposing Minder's read-only operations via streaming HTTP transport. This server allows LLM-based tools to interact with Minder for security policy management, repository monitoring, and compliance evaluation.

## Build & Test Commands

**Primary commands (operate on entire project):**

- `task build` - Build everything (Go binary + UI)
- `task lint` - Lint everything (Go + UI)
- `task fmt` - Format everything (Go + UI)
- `task test` - Run all tests
- `task check` - Run lint + test + build
- `task run` - Run server locally

**Specific targets:**

- `task build:ui` - Build TypeScript dashboard only
- `task lint:go` / `task lint:ui` - Lint specific codebase
- `task fmt:go` / `task fmt:ui` - Format specific codebase
- `task lint:go:fix` - Auto-fix Go lint issues
- `task dev` - Run server with auto-rebuild (requires `air`)
- `task dev:ui` - Run UI dev server with hot reload

**IMPORTANT**: You **MUST** use the Taskfile for building and testing this code base.

## Architecture

- `cmd/minder-mcp/` - Entry point
- `internal/config/` - Environment configuration
- `internal/minder/` - gRPC client wrapper
- `internal/middleware/` - Auth token handling
- `internal/tools/` - MCP tool implementations
- `internal/resources/` - MCP resource handlers (compliance dashboard)
- `ui/compliance-dashboard/` - TypeScript frontend for MCP Apps dashboard

## MCP Tool Conventions

- Names: `minder_<action>_<resource>` (snake_case)
- All tools are read-only
- Use `mcp.WithTitleAnnotation()` for display titles
- Use `mcp.WithReadOnlyHintAnnotation(true)` for all tools
- Use `mcp.Enum()` for constrained values
- Use `mcp.Title()` for parameter display names

## Compliance Dashboard (MCP Apps)

The `ui/compliance-dashboard/` directory contains a TypeScript frontend served as an MCP resource:

- Uses `@modelcontextprotocol/ext-apps` SDK for iframe ↔ host communication
- Built with Vite into a single HTML file embedded in the Go binary via `go:embed`
- Resource URI: `ui://minder/compliance-dashboard`

When modifying the dashboard:
- Run `task lint:ui` and `task fmt:ui` before committing
- Use `task dev:ui` for hot-reload development
- The `task build` command automatically builds UI first
- HTML output goes to `internal/resources/dist/index.html` (not committed to git)

## Git Workflow

- Never use `git add -A`
- Add specific files when staging
