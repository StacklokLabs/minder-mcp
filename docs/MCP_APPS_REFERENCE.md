# MCP Apps Reference Documentation

A comprehensive guide to MCP Apps - interactive user interfaces within AI chat clients.

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Technical Specification](#technical-specification)
4. [SDK Reference](#sdk-reference)
5. [Goose Integration](#goose-integration)
6. [Best Practices](#best-practices)
7. [Minder MCP Implementation](#minder-mcp-implementation)
8. [Resources](#resources)

---

## Overview

### What are MCP Apps?

MCP Apps is an extension to the Model Context Protocol (MCP) that enables MCP servers to serve interactive user interfaces directly within AI chat clients. Introduced as [SEP-1865](https://github.com/modelcontextprotocol/modelcontextprotocol/pull/1865), it was standardized in November 2025 through collaboration between Anthropic, OpenAI, and other MCP stakeholders.

### The Problem They Solve

Traditional MCP tools return text, images, or structured data displayed as static content. This creates limitations:

- Users cannot interact directly with data visualizations
- Forms require back-and-forth conversation prompts
- Real-time monitoring requires repeated queries
- Multi-step workflows become cumbersome conversations

### The Solution

MCP Apps allow tools to declare references to interactive HTML interfaces rendered inline within conversations:

- **Context Preservation**: Apps live inside the conversation without tab switching
- **Bidirectional Data Flow**: Apps call MCP tools directly; hosts push fresh data to apps
- **Host Integration**: Apps delegate actions to host-connected capabilities
- **Security Guarantees**: All apps run in sandboxed iframes with restricted permissions

### Client Support

MCP Apps are supported by:
- Claude and Claude Desktop
- Visual Studio Code (Insiders)
- Goose (1.19.0+)
- Postman
- MCPJam

---

## Architecture

### Core Primitives

MCP Apps relies on two key MCP primitives:

1. **Tools with UI Metadata**: Tools include `_meta.ui.resourceUri` pointing to a UI resource
2. **UI Resources**: Server-side resources via the `ui://` scheme containing bundled HTML/JavaScript

### Communication Flow

```
┌─────────┐         ┌─────────┐         ┌─────────┐
│  User   │         │  Agent  │         │ Server  │
└────┬────┘         └────┬────┘         └────┬────┘
     │                   │                   │
     │ "show analytics"  │                   │
     │──────────────────>│                   │
     │                   │    tools/call     │
     │                   │──────────────────>│
     │                   │                   │
     │                   │<──────────────────│
     │                   │   tool result     │
     │                   │                   │
     │         ┌─────────┴─────────┐         │
     │         │   App (iframe)    │         │
     │         └─────────┬─────────┘         │
     │                   │                   │
     │   user interacts  │                   │
     │──────────────────>│                   │
     │                   │   tools/call      │
     │                   │──────────────────>│
     │                   │<──────────────────│
     │                   │   fresh data      │
     │<──────────────────│                   │
     │   updated view    │                   │
```

### Workflow Sequence

1. **UI Preloading**: Host reads `_meta.ui.resourceUri` from tool description
2. **Resource Fetch**: Host fetches UI resource (HTML with bundled JS/CSS)
3. **Sandboxed Rendering**: HTML rendered in sandboxed iframe within conversation
4. **Bidirectional Communication**: App and host communicate via JSON-RPC over postMessage

---

## Technical Specification

### URI Scheme

The `ui://` scheme signals to hosts that this is an MCP App resource:

```
ui://my-tool/widget-01
ui://minder/compliance-dashboard
ui://get-time/mcp-app.html
```

### MIME Type

```
text/html;profile=mcp-app
```

This MIME type tells clients the resource should be:
1. Rendered as HTML
2. Treated as an MCP App with full capabilities
3. Run in a sandboxed iframe

### Tool-to-Resource Linking

Tools declare associated UI resources via metadata:

```typescript
{
  name: "visualize_data",
  description: "Visualize data as an interactive chart",
  inputSchema: { /* ... */ },
  _meta: {
    ui: {
      resourceUri: "ui://charts/interactive"
    }
  }
}
```

### JSON-RPC Protocol

Communication follows JSON-RPC 2.0 with `ui/` method prefix:

| Method | Direction | Purpose |
|--------|-----------|---------|
| `ui/initialize` | App → Host | Initialize connection, receive host context |
| `ui/notifications/initialized` | App → Host | Confirm initialization complete |
| `ui/notifications/tool-input` | Host → App | Push tool input parameters to app |
| `ui/notifications/tool-result` | Host → App | Push tool execution result to app |
| `ui/notifications/size-changed` | App → Host | App requests size change |
| `ui/notifications/host-context-changed` | Host → App | Theme/context changes |
| `tools/call` | App → Host | App requests a tool call |

### Host Context

When a View initializes, the Host provides:
- Container dimensions
- Theme (light/dark mode)
- Available capabilities
- User preferences

### Content Security Policy

Apps can specify additional resource loading via `_meta.ui.csp`:

```typescript
_meta: {
  ui: {
    resourceUri: "ui://my-app/widget",
    csp: "https://cdn.example.com",
    permissions: ["microphone", "camera"]
  }
}
```

---

## SDK Reference

### Installation

```bash
npm install @modelcontextprotocol/ext-apps
```

### Sub-packages

| Package | Purpose |
|---------|---------|
| `@modelcontextprotocol/ext-apps` | Core App class for UI side |
| `@modelcontextprotocol/ext-apps/server` | Server helpers (registerAppTool, registerAppResource) |
| `@modelcontextprotocol/ext-apps/react` | React hooks (useApp) |
| `@modelcontextprotocol/ext-apps/app-bridge` | For building MCP hosts |

### UI Side - App Class

```typescript
import { App } from '@modelcontextprotocol/ext-apps';

const app = new App({ name: 'My App', version: '1.0.0' });

// Register callbacks BEFORE connecting
app.ontoolresult = (result) => {
  const data = result.content?.find(c => c.type === 'text')?.text;
  // Process data...
};

app.ontoolinput = (input) => {
  // Process input parameters...
};

// Connect to host
await app.connect();

// Check host capabilities
const caps = app.getHostCapabilities();
const supportsServerTools = caps?.serverTools !== undefined;

// Call server tools (if supported)
if (supportsServerTools) {
  const result = await app.callServerTool({
    name: 'get-data',
    arguments: { id: '123' }
  });
}

// Close connection
await app.close();
```

### Server Side - Tool Registration

```typescript
import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import {
  registerAppTool,
  registerAppResource,
  RESOURCE_MIME_TYPE,
} from '@modelcontextprotocol/ext-apps/server';

const server = new McpServer({ name: 'My Server', version: '1.0.0' });
const resourceUri = 'ui://my-tool/app.html';

// Register tool with UI metadata
registerAppTool(
  server,
  'my-tool',
  {
    title: 'My Tool',
    description: 'Does something useful',
    inputSchema: {},
    _meta: { ui: { resourceUri } },
  },
  async () => ({
    content: [{ type: 'text', text: JSON.stringify(data) }],
  })
);

// Register UI resource
registerAppResource(
  server,
  resourceUri,
  resourceUri,
  { mimeType: RESOURCE_MIME_TYPE },
  async () => ({
    contents: [
      { uri: resourceUri, mimeType: RESOURCE_MIME_TYPE, text: htmlContent },
    ],
  })
);
```

### React Integration

```typescript
import { useApp } from '@modelcontextprotocol/ext-apps/react';

function MyComponent() {
  const [data, setData] = useState(null);

  useApp({
    appInfo: { name: 'My App', version: '1.0.0' },
    capabilities: {},
    onAppCreated: (app) => {
      app.ontoolresult = (result) => {
        const parsed = extractData(result);
        setData(parsed);
      };
    },
  });

  return <div>{/* render data */}</div>;
}
```

---

## Goose Integration

### Overview

Goose (Block's AI coding assistant) shipped MCP Apps support in **version 1.19.0**. The implementation includes:

- **Resource-based architecture**: MCP servers return `ui://` resource URIs rendered in sandboxed iframes
- **postMessage communication**: Standard postMessage API for theme sync, chat messaging, and resizing
- **Experimental status**: Based on the draft specification (SEP-1865)

### Configuration

**Via Goose Desktop UI:**
1. Navigate to Settings > Extensions
2. Click "Add custom extension"
3. Configure the MCP server providing MCP Apps resources

**Via Configuration File** (`~/.config/goose/config.yaml`):

```yaml
extensions:
  my_mcp_app:
    type: sse
    url: "https://your-mcp-server.com/mcp"
    headers:
      Authorization: "Bearer ${API_TOKEN}"
    timeout: 300
```

### Goose-Specific Patterns

**Resource Structure:**

```javascript
{
  uri: "ui://your-app/main",
  mimeType: "text/html",
  _meta: {
    ui: {
      resourceUri: "ui://your-app/main",
      csp: {
        // Custom CSP rules for external resources
      }
    }
  }
}
```

**ResizeObserver Pattern:**

Goose requires apps to send size change messages:

```javascript
window.parent.postMessage({
  type: "ui-size-change",
  payload: { height: document.body.scrollHeight }
}, "*");
```

**App Class Usage:**

```javascript
import { App } from "@modelcontextprotocol/ext-apps";

const app = new App();

// Get theme information
const theme = await app.getTheme();

// Send a message to the chat
await app.sendMessage("User clicked the button");

// Handle resize
app.onResize((dimensions) => {
  // Respond to container resize
});
```

### Goose Documentation

- [Building MCP Apps Tutorial](https://block.github.io/goose/docs/tutorials/building-mcp-apps/)
- [Rich Interactive Chat with MCP Apps](https://block.github.io/goose/docs/guides/interactive-chat/)
- [goose Lands MCP Apps Blog](https://block.github.io/goose/blog/2026/01/06/mcp-apps/)
- [From MCP-UI to MCP Apps Migration](https://block.github.io/goose/blog/2026/01/22/mcp-ui-to-mcp-apps/)
- [5 Tips for Building MCP Apps](https://block.github.io/goose/blog/2026/01/30/5-tips-building-mcp-apps/)

---

## Best Practices

### Security

1. **Trust the Sandbox**: All UI runs in sandboxed iframes with restricted permissions
2. **Use Pre-declared Templates**: Hosts can review HTML before rendering
3. **Auditable Messages**: All communication goes through loggable JSON-RPC
4. **User Consent**: Hosts can require explicit approval for UI-initiated tool calls

### Development

1. **Bundle Everything**: Use Vite with `vite-plugin-singlefile` for single-file output
2. **Register Callbacks Before Connect**: Set up handlers before `app.connect()`
3. **Handle Latency**: Design UI to handle round-trip delays gracefully
4. **Check Capabilities**: Always verify host support before using features

```typescript
if (app.getHostCapabilities()?.serverTools) {
  await app.callServerTool({ name: 'my-tool', arguments: {} });
} else {
  showFallbackUI();
}
```

### UI Design

1. **Content Security Policy**: Define appropriate CSP in HTML
2. **Responsive Design**: Container dimensions vary by host
3. **Theme Support**: Listen for `host-context-changed` for dark/light modes
4. **Error Handling**: Handle tool call failures gracefully
5. **Dual-mode Operation**: Support both active (calling tools) and passive (receiving results) modes

### Project Structure

```
my-mcp-app/
├── package.json
├── tsconfig.json
├── vite.config.ts
├── server.ts          # MCP server with tool + resource
├── mcp-app.html       # UI entry point
└── src/
    └── mcp-app.ts     # UI logic
```

---

## Minder MCP Implementation

### Overview

The minder-mcp project implements a compliance dashboard as an MCP App, demonstrating production patterns for MCP Apps.

### Architecture

```
minder-mcp/
├── cmd/minder-mcp/          # Entry point
├── internal/
│   ├── tools/               # MCP tool implementations (27 tools)
│   │   └── dashboard.go     # minder_show_dashboard tool
│   └── resources/           # MCP resource handlers
│       ├── dashboard_html.go
│       └── dist/index.html  # Embedded HTML (generated)
└── ui/compliance-dashboard/ # TypeScript frontend
    ├── src/
    │   ├── main.ts          # Entry point
    │   ├── mcp-client.ts    # MCP Apps SDK wrapper
    │   ├── dashboard.ts     # Dashboard logic (650 lines)
    │   └── styles.css       # Styling
    └── index.html           # HTML template
```

### Resource Configuration

**Go Constants:**

```go
const (
    DashboardURI = "ui://minder/compliance-dashboard"
    DashboardMIMEType = "text/html;profile=mcp-app"
)
```

**HTML Embedding:**

```go
//go:embed dist/index.html
var dashboardHTML string
```

### Dashboard Tool

The `minder_show_dashboard` tool triggers dashboard display:

```go
func (t *Tools) ShowDashboard(ctx context.Context, params ShowDashboardParams) (*mcp.CallToolResult, error) {
    return &mcp.CallToolResult{
        Content: []mcp.Content{
            mcp.TextContent{
                Type: "text",
                Text: "Dashboard displayed",
            },
        },
        Meta: map[string]interface{}{
            "ui": map[string]interface{}{
                "resourceUri": DashboardURI,
            },
        },
    }, nil
}
```

### MCP Client Wrapper

The dashboard wraps the SDK for Minder-specific operations:

```typescript
// ui/compliance-dashboard/src/mcp-client.ts
export class MCPAppsClient {
  private app: App;

  async connect(): Promise<void> {
    this.app = new App();
    this.app.ontoolresult = this.handleToolResult.bind(this);
    this.app.ontoolinput = this.handleToolInput.bind(this);
    await this.app.connect();
  }

  async callTool<T>(name: string, args: Record<string, unknown>): Promise<T> {
    const result = await this.app.callServerTool({ name, arguments: args });
    return this.extractData(result);
  }

  // Minder-specific helpers
  async listProfiles(projectId?: string): Promise<ProfilesResult> {
    return this.callTool('minder_list_profiles', { project_id: projectId });
  }

  async getProfileStatus(profileId: string): Promise<ProfileStatusResult> {
    return this.callTool('minder_get_profile_status', { profile_id: profileId });
  }

  async listRepositories(projectId?: string): Promise<RepositoriesResult> {
    return this.callTool('minder_list_repositories', { project_id: projectId });
  }
}
```

### Dual-Mode Operation

The dashboard supports two modes:

**Active Mode** (host supports `serverTools`):
```typescript
if (this.client.supportsServerTools()) {
  const profiles = await this.client.listProfiles();
  this.renderProfiles(profiles);
}
```

**Passive Mode** (waiting for tool results):
```typescript
this.client.onToolResult((result) => {
  if (this.isProfilesResult(result)) {
    this.renderProfiles(result);
  }
});
```

### Build Integration

**Taskfile.yml:**

```yaml
tasks:
  build:
    deps: [build:ui]
    cmds:
      - go build -o bin/minder-mcp ./cmd/minder-mcp

  build:ui:
    dir: ui/compliance-dashboard
    cmds:
      - npm install
      - npm run build
    generates:
      - ../../internal/resources/dist/index.html
```

**Vite Configuration:**

```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import { viteSingleFile } from 'vite-plugin-singlefile';

export default defineConfig({
  plugins: [viteSingleFile()],
  build: {
    outDir: '../../internal/resources/dist',
    emptyOutDir: true,
  },
});
```

---

## Resources

### Official Documentation

- [MCP Apps - Model Context Protocol](https://modelcontextprotocol.io/docs/extensions/apps)
- [MCP Apps Blog Post](http://blog.modelcontextprotocol.io/posts/2025-11-21-mcp-apps/)
- [@modelcontextprotocol/ext-apps API](https://modelcontextprotocol.github.io/ext-apps/api/documents/Quickstart.html)

### GitHub Repositories

- [modelcontextprotocol/ext-apps](https://github.com/modelcontextprotocol/ext-apps) - Official SDK
- [block/goose](https://github.com/block/goose) - Goose AI assistant
- [MCP Apps Examples](https://github.com/modelcontextprotocol/ext-apps/tree/main/examples)

### NPM Packages

- [@modelcontextprotocol/ext-apps](https://www.npmjs.com/package/@modelcontextprotocol/ext-apps)

### Example Projects

| Example | Description |
|---------|-------------|
| `map-server` | CesiumJS globe visualization |
| `threejs-server` | Three.js 3D scenes |
| `pdf-server` | PDF viewer with pan/zoom |
| `system-monitor-server` | Real-time system metrics |
| `sheet-music-server` | Music notation display |
| `cohort-heatmap-server` | Heatmap visualizations |

### Use Cases

1. **Data Exploration**: Interactive maps, charts, drill-down dashboards
2. **Configuration Forms**: Multi-option setup with validation
3. **Rich Media Viewing**: PDFs, 3D models, image galleries
4. **Real-time Monitoring**: Live metrics, logs, status dashboards
5. **Multi-step Workflows**: Approval flows, code review, triage systems

---

*Document generated from research on MCP Apps specification, Goose implementation, and minder-mcp codebase analysis.*
