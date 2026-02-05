# Testing MCP Apps with basic-host

This guide explains how to test the Minder Compliance Dashboard (an MCP App) locally using the `basic-host` test tool from the MCP Apps SDK.

## Prerequisites

- Node.js 18+
- Go 1.25+
- A valid Minder auth token

## Setup

### 1. Clone the ext-apps repository

```bash
cd ~/Development
git clone --depth 1 https://github.com/modelcontextprotocol/ext-apps.git mcp-ext-apps
cd mcp-ext-apps
npm install
```

### 2. Build basic-host

```bash
cd examples/basic-host
npm install
npm run build
```

### 3. Build minder-mcp

```bash
cd /path/to/minder-mcp
task build
```

## Running

### 1. Start minder-mcp on port 9090

The basic-host uses ports 8080/8081 by default, so run minder-mcp on a different port:

```bash
MCP_PORT=9090 \
MINDER_SERVER_HOST=api.stacklok.com \
MINDER_AUTH_TOKEN=<your-token> \
./bin/minder-mcp
```

### 2. Start basic-host

In a separate terminal:

```bash
cd ~/Development/mcp-ext-apps/examples/basic-host
SERVERS='["http://localhost:9090/mcp"]' npm run serve
```

You should see:
```
Host server:    http://localhost:8080
Sandbox server: http://localhost:8081
```

### 3. Open the test interface

Navigate to **http://localhost:8080** in your browser.

## Testing the Dashboard

1. Select your server from the dropdown (should show "minder-mcp")
2. Select the `minder_show_dashboard` tool
3. Click **Call Tool**
4. The compliance dashboard should render in the iframe below

## How It Works

```
┌─────────────────────────────────────────────────────────┐
│  Browser (localhost:8080)                               │
│  ┌───────────────────────────────────────────────────┐  │
│  │  basic-host UI                                    │  │
│  │  - Connects to MCP server                         │  │
│  │  - Lists tools                                    │  │
│  │  - Renders MCP Apps in sandboxed iframe           │  │
│  │  ┌─────────────────────────────────────────────┐  │  │
│  │  │  Sandboxed iframe (localhost:8081)          │  │  │
│  │  │  - Loads dashboard HTML from ui:// resource │  │  │
│  │  │  - Communicates via postMessage             │  │  │
│  │  │  - Can call server tools through host       │  │  │
│  │  └─────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                           │
                           │ HTTP (MCP protocol)
                           ▼
┌─────────────────────────────────────────────────────────┐
│  minder-mcp server (localhost:9090)                     │
│  - Serves MCP tools (list_profiles, list_repositories)  │
│  - Serves ui:// resource (dashboard HTML)               │
│  - CORS enabled for cross-origin requests               │
│  - Session management via Mcp-Session-Id header         │
└─────────────────────────────────────────────────────────┘
                           │
                           │ gRPC
                           ▼
┌─────────────────────────────────────────────────────────┐
│  Minder API (api.stacklok.com)                          │
└─────────────────────────────────────────────────────────┘
```

## Troubleshooting

### "Loading..." stuck on server dropdown

Check that minder-mcp is running and accessible:
```bash
curl http://localhost:9090/mcp -X POST \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"initialize","id":1,"params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'
```

### "Invalid session ID" error

Ensure the server has CORS properly configured with `Access-Control-Expose-Headers: Mcp-Session-Id`. This allows the browser to read the session ID from responses.

### Dashboard shows 0 repos/profiles

Check browser console for errors. The dashboard needs the MCP client to support `serverTools` capability to actively fetch data. If using passive mode, data only appears when the host pushes tool results.

### Port 8080 already in use

Either stop the process using port 8080, or run minder-mcp on a different port and update the `SERVERS` environment variable for basic-host.

## Key Configuration

### CORS Headers (minder-mcp)

The server must expose the session header for browser clients:
```go
cors.Options{
    AllowedOrigins:   []string{"*"},
    AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"*"},
    ExposedHeaders:   []string{"Mcp-Session-Id"},
    AllowCredentials: true,
}
```

### Tool Definition

The dashboard tool must include `_meta.ui.resourceUri` in its definition for hosts to preload the UI:
```go
dashboardTool.Meta = mcp.NewMetaFromMap(map[string]any{
    "ui": map[string]any{
        "resourceUri": "ui://minder/compliance-dashboard",
    },
})
```
