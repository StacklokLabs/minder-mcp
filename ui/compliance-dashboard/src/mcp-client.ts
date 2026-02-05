import {
  App,
  PostMessageTransport,
  type McpUiToolResultNotification,
  type McpUiToolInputNotification,
} from '@modelcontextprotocol/ext-apps';

/**
 * Type for tool result notification params.
 */
export type ToolResultParams = McpUiToolResultNotification['params'];

/**
 * Type for tool input notification params.
 */
export type ToolInputParams = McpUiToolInputNotification['params'];

/**
 * Callback type for receiving tool results via notification.
 */
export type ToolResultCallback = (result: ToolResultParams) => void;

/**
 * Callback type for receiving tool input via notification.
 */
export type ToolInputCallback = (input: ToolInputParams) => void;

/**
 * Container dimensions from the host.
 */
export interface ContainerDimensions {
  width?: number;
  height?: number;
  maxWidth?: number;
  maxHeight?: number;
}

/**
 * Callback type for receiving container dimension changes from host.
 */
export type DimensionsCallback = (dimensions: ContainerDimensions) => void;

/**
 * MCP Apps client for communicating with the MCP server from a UI iframe.
 * Uses the official @modelcontextprotocol/ext-apps SDK for secure communication.
 */
export class MCPAppsClient {
  private app: App;
  private connected = false;
  private onToolResultCallback: ToolResultCallback | null = null;
  private onToolInputCallback: ToolInputCallback | null = null;
  private onDimensionsCallback: DimensionsCallback | null = null;

  constructor() {
    this.app = new App({
      name: 'Minder Compliance Dashboard',
      version: '1.0.0',
    });
  }

  /**
   * Register a callback for tool results pushed by the host.
   * MUST be called before connect() to receive initial data.
   */
  onToolResult(callback: ToolResultCallback): void {
    this.onToolResultCallback = callback;
    this.app.ontoolresult = (result): void => {
      console.log('[MCP] Received tool result notification:', result);
      // Defensive null check - some hosts may pass null/undefined
      if (!result) {
        console.warn('[MCP] Received null/undefined tool result');
        return;
      }
      if (this.onToolResultCallback) {
        this.onToolResultCallback(result);
      }
    };
  }

  /**
   * Register a callback for tool input pushed by the host.
   * MUST be called before connect() to receive initial data.
   */
  onToolInput(callback: ToolInputCallback): void {
    this.onToolInputCallback = callback;
    this.app.ontoolinput = (input): void => {
      console.log('[MCP] Received tool input notification:', input);
      // Defensive null check - some hosts may pass null/undefined
      if (!input) {
        console.warn('[MCP] Received null/undefined tool input');
        return;
      }
      if (this.onToolInputCallback) {
        this.onToolInputCallback(input);
      }
    };
  }

  /**
   * Register a callback for container dimension changes from the host.
   * Called when the host resizes the iframe container.
   */
  onDimensionsChange(callback: DimensionsCallback): void {
    this.onDimensionsCallback = callback;
    this.app.onhostcontextchanged = (context): void => {
      console.log('[MCP] Received host context change:', context);
      // Defensive null check - some hosts (e.g., Goose) may pass null/undefined
      if (!context) {
        console.warn('[MCP] Received null/undefined host context change');
        return;
      }
      if (context.containerDimensions && this.onDimensionsCallback) {
        const dims = context.containerDimensions as ContainerDimensions;
        this.onDimensionsCallback(dims);
      }
    };
  }

  async connect(): Promise<void> {
    if (this.connected) {
      return;
    }

    // Use official PostMessageTransport with explicit source validation
    const transport = new PostMessageTransport(window.parent, window.parent);
    await this.app.connect(transport);
    this.connected = true;
    console.log('[MCP] Connected to host');
  }

  /**
   * Check if the host supports calling server tools directly.
   */
  supportsServerTools(): boolean {
    try {
      const caps = this.app.getHostCapabilities();
      // serverTools is an object when supported, undefined when not
      // Some hosts may return null instead of an object
      if (!caps) {
        console.log('[MCP] Host capabilities is null/undefined');
        return false;
      }
      return caps.serverTools !== undefined;
    } catch (error) {
      console.warn('[MCP] Error checking host capabilities:', error);
      return false;
    }
  }

  async disconnect(): Promise<void> {
    if (!this.connected) {
      return;
    }

    await this.app.close();
    this.connected = false;
  }

  /**
   * Send a message to the chat from the dashboard.
   * This allows the dashboard to communicate back to the conversation.
   *
   * @param text - The message text to send
   * @returns True if the message was accepted, false if rejected by host
   */
  async sendMessage(text: string): Promise<boolean> {
    if (!this.connected) {
      console.warn('[MCP] Cannot send message: not connected');
      return false;
    }

    try {
      const result = await this.app.sendMessage({
        role: 'user',
        content: [{ type: 'text', text }],
      });
      return !result.isError;
    } catch (error) {
      console.error('[MCP] Failed to send message:', error);
      return false;
    }
  }

  async callTool<T = unknown>(
    name: string,
    args: Record<string, unknown> = {}
  ): Promise<T> {
    if (!this.connected) {
      throw new Error('Client not connected');
    }

    const result = await this.app.callServerTool({
      name,
      arguments: args ?? {},
    });

    // Defensive null check for result
    if (!result) {
      console.warn('[MCP] callServerTool returned null/undefined');
      throw new Error('Tool call returned no result');
    }

    // Extract the content from the tool result
    if (
      result.content &&
      Array.isArray(result.content) &&
      result.content.length > 0
    ) {
      const firstContent = result.content[0];
      if (
        firstContent &&
        firstContent.type === 'text' &&
        typeof firstContent.text === 'string'
      ) {
        try {
          return JSON.parse(firstContent.text) as T;
        } catch {
          return firstContent.text as T;
        }
      }
    }

    return result as T;
  }

  async listProfiles(projectId?: string): Promise<ProfilesResult> {
    const args: Record<string, unknown> = {};
    if (projectId) {
      args.project_id = projectId;
    }
    // API returns array directly, wrap it for dashboard
    const profiles = await this.callTool<Profile[]>('minder_list_profiles', args);
    return { profiles: Array.isArray(profiles) ? profiles : [] };
  }

  async getProfileStatus(options: {
    profileId?: string;
    name?: string;
    projectId?: string;
  }): Promise<ProfileStatusResult> {
    const args: Record<string, unknown> = {};
    if (options.profileId) {
      args.profile_id = options.profileId;
    }
    if (options.name) {
      args.name = options.name;
    }
    if (options.projectId) {
      args.project_id = options.projectId;
    }
    // API returns nested structure, flatten it for dashboard
    const response = await this.callTool<ProfileStatusApiResponse>(
      'minder_get_profile_status',
      args
    );
    return {
      profile_id: response.profile_status?.profile_id ?? '',
      profile_name: response.profile_status?.profile_name ?? '',
      profile_status: response.profile_status?.profile_status ?? '',
      rule_evaluation_status: response.rule_evaluation_status,
    };
  }

  async listRepositories(options?: {
    projectId?: string;
    provider?: string;
    cursor?: string;
    limit?: number;
  }): Promise<RepositoriesResult> {
    const args: Record<string, unknown> = {};
    if (options?.projectId) {
      args.project_id = options.projectId;
    }
    if (options?.provider) {
      args.provider = options.provider;
    }
    if (options?.cursor) {
      args.cursor = options.cursor;
    }
    if (options?.limit) {
      args.limit = options.limit;
    }
    // API returns { results: [...], has_more, next_cursor }
    const response = await this.callTool<RepositoriesApiResponse>(
      'minder_list_repositories',
      args
    );
    return {
      repositories: response.results ?? [],
      cursor: response.next_cursor,
    };
  }
}

// Type definitions for Minder data - matching actual API responses

export interface Profile {
  id: string;
  name: string;
  labels?: string[];
  context?: {
    project?: string;
  };
}

// API returns array directly, this wraps it for dashboard use
export interface ProfilesResult {
  profiles: Profile[];
}

export interface RuleEvaluationStatus {
  rule_name?: string;
  rule_type_name?: string;
  status: 'success' | 'failure' | 'error' | 'pending' | 'skipped';
  entity_info?: {
    name?: string;
    entity_type?: string;
  };
  remediation_status?: string;
  alert?: {
    status?: string;
  };
}

// API returns nested structure with profile_status object
export interface ProfileStatusApiResponse {
  profile_status: {
    profile_id: string;
    profile_name: string;
    profile_status: string;
  };
  rule_evaluation_status?: RuleEvaluationStatus[];
}

// Flattened version for dashboard use
export interface ProfileStatusResult {
  profile_id: string;
  profile_name: string;
  profile_status: string;
  rule_evaluation_status?: RuleEvaluationStatus[];
}

export interface Repository {
  id?: string;
  name: string;
  owner: string;
  provider?: string;
  repo_id?: number;
  is_private?: boolean;
  is_fork?: boolean;
  hook_url?: string;
  deploy_url?: string;
  clone_url?: string;
  default_branch?: string;
}

// API returns { results: [...], has_more: bool }
export interface RepositoriesApiResponse {
  results: Repository[];
  has_more: boolean;
  next_cursor?: string;
}

// Dashboard-friendly version
export interface RepositoriesResult {
  repositories: Repository[];
  cursor?: string;
}
