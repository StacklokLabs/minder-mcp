# Minder MCP Server

## Project Overview

MCP server exposing Minder's read-only operations via streaming HTTP transport. This server allows LLM-based tools to interact with Minder for security policy management, repository monitoring, and compliance evaluation.

## Build & Test Commands

- `task build` - Build binary (includes UI build)
- `task build:ui` - Build TypeScript dashboard only
- `task lint` - Run golangci-lint
- `task lint:ui` - Lint TypeScript (ESLint + typecheck)
- `task lint:all` - Run all linters (Go + UI)
- `task test` - Run tests
- `task run` - Run server locally

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
- Run `task lint:ui` before committing
- The `task build` command automatically builds UI first
- HTML output goes to `internal/resources/dist/index.html`

## Git Workflow

- Never use `git add -A`
- Add specific files when staging
