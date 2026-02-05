# Minder MCP Server

## Project Overview

MCP server exposing Minder's read-only operations via streaming HTTP transport. This server allows LLM-based tools to interact with Minder for security policy management, repository monitoring, and compliance evaluation.

## Build & Test Commands

- `task build` - Build binary
- `task lint` - Run golangci-lint
- `task test` - Run tests
- `task run` - Run server locally

**IMPORTANT**: You **MUST** use the Taskfile for building and testing this code base.

## Architecture

- `cmd/minder-mcp/` - Entry point
- `internal/config/` - Environment configuration
- `internal/minder/` - gRPC client wrapper
- `internal/middleware/` - Auth token handling
- `internal/tools/` - MCP tool implementations

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
