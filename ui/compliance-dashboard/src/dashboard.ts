import {
  MCPAppsClient,
  Profile,
  ProfileStatusResult,
  ProfileStatusApiResponse,
  Repository,
  RepositoriesApiResponse,
  RuleEvaluationStatus,
  type ToolResultParams,
  type ToolInputParams,
} from './mcp-client.js';

/**
 * Get a required DOM element by ID with type safety.
 */
function getRequiredElement<T extends HTMLElement>(id: string): T {
  const element = document.getElementById(id);
  if (!element) {
    throw new Error(`Required element #${id} not found`);
  }
  return element as T;
}

// Dashboard state
let profiles: Profile[] = [];
let profileStatuses: Map<string, ProfileStatusResult> = new Map();
let repositories: Repository[] = [];
let isLoading = false;
let mcpClient: MCPAppsClient | null = null;
let resizeObserver: ResizeObserver | null = null;
let resizeTimeout: ReturnType<typeof setTimeout> | null = null;

// Project context - extracted from tool inputs
let currentProjectId: string | null = null;

/**
 * Notify the host about content size changes.
 * Required for Goose compatibility - the host uses this to size the iframe.
 * Debounced to prevent excessive postMessage calls during rapid resize events.
 */
function notifySizeChange(): void {
  if (resizeTimeout) {
    clearTimeout(resizeTimeout);
  }

  resizeTimeout = setTimeout(() => {
    const height = document.body.scrollHeight;
    window.parent.postMessage(
      {
        type: 'ui-size-change',
        payload: { height },
      },
      '*'
    );
  }, 16); // ~1 frame at 60fps
}

/**
 * Set up ResizeObserver to automatically notify host of size changes.
 */
function setupResizeObserver(): void {
  if (resizeObserver) {
    return; // Already set up
  }

  resizeObserver = new ResizeObserver(() => {
    notifySizeChange();
  });

  // Observe the body for any size changes
  resizeObserver.observe(document.body);

  // Also send initial size
  notifySizeChange();
}

// DOM element references
let refreshBtn: HTMLButtonElement;
let errorContainer: HTMLDivElement;
let totalReposEl: HTMLElement;
let passingCountEl: HTMLElement;
let failingCountEl: HTMLElement;
let complianceRateEl: HTMLElement;
let profilesListEl: HTMLDivElement;
let repositoriesListEl: HTMLDivElement;
let profileFilterEl: HTMLInputElement;
let profileStatusFilterEl: HTMLSelectElement;
let repoFilterEl: HTMLInputElement;
let statusFilterEl: HTMLSelectElement;

/**
 * Initialize the dashboard when the DOM is loaded.
 */
export function initDashboard(): void {
  // Get DOM references with type safety
  refreshBtn = getRequiredElement<HTMLButtonElement>('refresh-btn');
  errorContainer = getRequiredElement<HTMLDivElement>('error-container');
  totalReposEl = getRequiredElement<HTMLElement>('total-repos');
  passingCountEl = getRequiredElement<HTMLElement>('passing-count');
  failingCountEl = getRequiredElement<HTMLElement>('failing-count');
  complianceRateEl = getRequiredElement<HTMLElement>('compliance-rate');
  profilesListEl = getRequiredElement<HTMLDivElement>('profiles-list');
  repositoriesListEl = getRequiredElement<HTMLDivElement>('repositories-list');
  profileFilterEl = getRequiredElement<HTMLInputElement>('profile-filter');
  profileStatusFilterEl = getRequiredElement<HTMLSelectElement>('profile-status-filter');
  repoFilterEl = getRequiredElement<HTMLInputElement>('repo-filter');
  statusFilterEl = getRequiredElement<HTMLSelectElement>('status-filter');

  // Set up event listeners
  refreshBtn.addEventListener('click', loadDashboard);
  profileFilterEl.addEventListener('input', renderProfiles);
  profileStatusFilterEl.addEventListener('change', renderProfiles);
  repoFilterEl.addEventListener('input', renderRepositories);
  statusFilterEl.addEventListener('change', renderRepositories);

  // Set up tab switching
  document.querySelectorAll('.tab').forEach((tab) => {
    tab.addEventListener('click', (e) => {
      const target = e.currentTarget as HTMLElement;
      const tabName = target.dataset.tab;
      if (tabName) {
        switchTab(tabName);
      }
    });
  });

  // Set up ResizeObserver for Goose compatibility
  setupResizeObserver();

  // Initialize MCP client
  mcpClient = new MCPAppsClient();

  // Register handlers BEFORE connecting to receive initial data push
  mcpClient.onToolResult((result: ToolResultParams) => {
    console.log('[Dashboard] Received tool result:', result);
    handleToolResult(result);
  });

  mcpClient.onToolInput((input: ToolInputParams) => {
    console.log('[Dashboard] Received tool input:', input);
    handleToolInput(input);
  });

  // Now connect and load data
  loadDashboard();
}

/**
 * Handle tool results pushed by the MCP host.
 * This is called when the LLM invokes a Minder tool and the host pushes the result.
 */
function handleToolResult(result: ToolResultParams): void {
  // Extract the tool name and content from the result
  // The result may contain structuredContent or content array
  let data: unknown = null;

  // Check for structuredContent first (preferred for rich data)
  const structuredContent = result.structuredContent as
    | Record<string, unknown>
    | undefined;
  if (structuredContent) {
    data = structuredContent;
  } else if (
    result.content &&
    Array.isArray(result.content) &&
    result.content.length > 0
  ) {
    const firstContent = result.content[0];
    if (
      firstContent.type === 'text' &&
      'text' in firstContent &&
      typeof firstContent.text === 'string'
    ) {
      try {
        data = JSON.parse(firstContent.text);
      } catch {
        console.warn(
          '[Dashboard] Failed to parse tool result as JSON:',
          firstContent.text
        );
        return;
      }
    }
  }

  if (!data) {
    console.warn('[Dashboard] No data in tool result');
    return;
  }

  // Try to identify the data type and update accordingly
  // API returns profiles as array directly
  if (isProfilesArray(data)) {
    console.log(
      '[Dashboard] Received profiles data:',
      data.length,
      'profiles'
    );
    profiles = data;
    updateSummaryCards();
    renderProfiles();
  // API returns profile status with nested profile_status object
  } else if (isProfileStatusApiResponse(data)) {
    const status: ProfileStatusResult = {
      profile_id: data.profile_status.profile_id,
      profile_name: data.profile_status.profile_name,
      profile_status: data.profile_status.profile_status,
      rule_evaluation_status: data.rule_evaluation_status,
    };
    console.log('[Dashboard] Received profile status:', status.profile_name);
    profileStatuses.set(status.profile_id, status);
    updateSummaryCards();
    renderProfiles();
  // API returns repositories as { results: [...] }
  } else if (isRepositoriesApiResponse(data)) {
    console.log(
      '[Dashboard] Received repositories data:',
      data.results.length,
      'repos'
    );
    repositories = data.results || [];
    updateSummaryCards();
    renderRepositories();
  } else {
    console.log('[Dashboard] Received unknown data type:', data);
  }
}

/**
 * Type guard for ProfilesResult - API returns array directly
 */
function isProfilesArray(data: unknown): data is Profile[] {
  return Array.isArray(data) && data.every(item =>
    typeof item === 'object' && item !== null && 'name' in item
  );
}

/**
 * Type guard for ProfileStatusApiResponse - API returns nested structure
 */
function isProfileStatusApiResponse(data: unknown): data is ProfileStatusApiResponse {
  return (
    typeof data === 'object' &&
    data !== null &&
    'profile_status' in data &&
    typeof (data as ProfileStatusApiResponse).profile_status === 'object'
  );
}

/**
 * Type guard for RepositoriesApiResponse - API returns { results: [...] }
 */
function isRepositoriesApiResponse(data: unknown): data is RepositoriesApiResponse {
  return (
    typeof data === 'object' &&
    data !== null &&
    'results' in data &&
    Array.isArray((data as RepositoriesApiResponse).results)
  );
}

/**
 * Handle tool input pushed by the MCP host.
 * Extracts project_id from arguments for use in subsequent calls.
 */
function handleToolInput(input: ToolInputParams): void {
  // Extract project_id from the tool arguments if present
  const args = input.arguments as Record<string, unknown> | undefined;
  if (args && typeof args.project_id === 'string' && args.project_id) {
    const newProjectId = args.project_id;
    if (newProjectId !== currentProjectId) {
      console.log('[Dashboard] Project context set to:', newProjectId);
      currentProjectId = newProjectId;
    }
  }
}

/**
 * Load dashboard data from Minder via MCP.
 */
async function loadDashboard(): Promise<void> {
  if (isLoading || !mcpClient) {
    return;
  }

  // Capture client reference for use in closures (TypeScript narrowing)
  const client = mcpClient;

  isLoading = true;
  refreshBtn.disabled = true;
  showError(null);

  try {
    // Connect to the MCP host (handlers are already registered)
    await client.connect();

    // Check if the host supports calling server tools
    const supportsServerTools = client.supportsServerTools();
    console.log('[Dashboard] Host supports serverTools:', supportsServerTools);

    if (!supportsServerTools) {
      // Host doesn't support active tool calls - show waiting state
      // Data will come through ontoolresult notifications when LLM calls tools
      console.log('[Dashboard] Waiting for tool results via notifications...');
      // Hide refresh button since we can't actively fetch data
      refreshBtn.style.display = 'none';
      showWaitingForData();
      return;
    }

    // Host supports serverTools - actively fetch data
    // Show loading state
    profilesListEl.innerHTML = `
      <div class="loading">
        <div class="spinner"></div>
        Loading profiles...
      </div>
    `;

    // Load profiles (use project context if available)
    try {
      const profilesResult = await client.listProfiles(
        currentProjectId ?? undefined
      );
      profiles = profilesResult.profiles || [];
    } catch (e) {
      console.error('Failed to load profiles:', e);
      showInfoMessage();
      return;
    }

    // Load status for each profile in parallel
    profileStatuses = new Map();
    const statusPromises = profiles.map(async (profile) => {
      try {
        // Use project from profile context if available, or fall back to current context
        const profileProjectId =
          profile.context?.project ?? currentProjectId ?? undefined;
        const status = await client.getProfileStatus({
          name: profile.name,
          projectId: profileProjectId,
        });
        return { profile, status };
      } catch (e) {
        console.error(`Failed to get status for profile ${profile.name}:`, e);
        return null;
      }
    });
    const statusResults = await Promise.all(statusPromises);
    for (const result of statusResults) {
      if (result) {
        profileStatuses.set(result.profile.id, result.status);
      }
    }

    // Load repositories (use project context if available)
    try {
      const reposResult = await client.listRepositories({
        projectId: currentProjectId ?? undefined,
      });
      repositories = reposResult.repositories || [];
    } catch (e) {
      console.error('Failed to load repositories:', e);
    }

    // Update UI
    updateSummaryCards();
    renderProfiles();
    renderRepositories();
  } catch (error) {
    console.error('Dashboard load error:', error);
    showInfoMessage();
  } finally {
    isLoading = false;
    refreshBtn.disabled = false;
  }
}

/**
 * Show a state indicating we're waiting for data via notifications.
 */
function showWaitingForData(): void {
  const message = `
    <div class="empty-state">
      <div class="spinner" style="margin-bottom: 16px;"></div>
      <h3 style="margin-bottom: 12px;">Waiting for Compliance Data</h3>
      <p style="margin-bottom: 16px;">
        Ask your AI assistant to fetch Minder data using these tools:
      </p>
      <ul style="text-align: left; max-width: 400px; margin: 0 auto 16px;">
        <li><code>minder_list_profiles</code> - List all profiles</li>
        <li><code>minder_get_profile_status</code> - Get compliance status</li>
        <li><code>minder_list_repositories</code> - List repositories</li>
      </ul>
      <p style="color: var(--text-muted); font-size: 13px;">
        Data will appear automatically when received.
      </p>
    </div>
  `;
  profilesListEl.innerHTML = message;
  repositoriesListEl.innerHTML = message;
  isLoading = false;
  refreshBtn.disabled = false;
}

/**
 * Show informational message when MCP Apps is not supported.
 */
function showInfoMessage(): void {
  const message = `
    <div class="empty-state">
      <h3 style="margin-bottom: 12px;">MCP Apps Dashboard</h3>
      <p style="margin-bottom: 16px;">
        This compliance dashboard is designed to work with MCP Apps-enabled clients.
      </p>
      <p style="margin-bottom: 16px;">
        To view compliance data, ask your AI assistant to:
      </p>
      <ul style="text-align: left; max-width: 400px; margin: 0 auto 16px;">
        <li>List profiles: <code>minder_list_profiles</code></li>
        <li>Get profile status: <code>minder_get_profile_status</code></li>
        <li>List repositories: <code>minder_list_repositories</code></li>
        <li>View evaluation history: <code>minder_list_evaluation_history</code></li>
      </ul>
      <p style="color: var(--text-muted); font-size: 13px;">
        The AI will aggregate the data and provide compliance insights.
      </p>
    </div>
  `;
  profilesListEl.innerHTML = message;
  repositoriesListEl.innerHTML = message;
}

/**
 * Show or clear error message.
 */
function showError(message: string | null): void {
  if (message) {
    errorContainer.innerHTML = `<div class="error-message">${escapeHtml(message)}</div>`;
  } else {
    errorContainer.innerHTML = '';
  }
}

/**
 * Escape HTML to prevent XSS.
 */
function escapeHtml(text: string): string {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

/**
 * Escape string for use in HTML attributes.
 */
function escapeAttr(text: string): string {
  return String(text)
    .replace(/&/g, '&amp;')
    .replace(/'/g, '&#39;')
    .replace(/"/g, '&quot;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}

/**
 * Update the summary cards with current data.
 */
function updateSummaryCards(): void {
  const totalRepos = repositories.length;
  let passing = 0;
  let failing = 0;

  // Count from profile statuses
  for (const status of profileStatuses.values()) {
    if (status.rule_evaluation_status) {
      for (const rule of status.rule_evaluation_status) {
        if (rule.status === 'success') {
          passing++;
        } else if (rule.status === 'failure' || rule.status === 'error') {
          failing++;
        }
      }
    }
  }

  totalReposEl.textContent = String(totalRepos);
  passingCountEl.textContent = String(passing);
  failingCountEl.textContent = String(failing);

  const total = passing + failing;
  const rate = total > 0 ? Math.round((passing / total) * 100) : 0;
  complianceRateEl.textContent = String(rate);
  complianceRateEl.className =
    'percentage ' +
    (rate >= 80 ? 'success' : rate >= 50 ? 'warning' : 'failure');
}

/**
 * Get overall status from profile status.
 */
function getOverallStatus(
  status: ProfileStatusResult | undefined
): 'success' | 'failure' | 'pending' | 'skipped' {
  if (!status?.rule_evaluation_status) {
    return 'pending';
  }

  const statuses = status.rule_evaluation_status.map((r) => r.status);
  if (statuses.some((s) => s === 'failure' || s === 'error')) {
    return 'failure';
  }
  if (statuses.every((s) => s === 'success')) {
    return 'success';
  }
  if (statuses.some((s) => s === 'pending')) {
    return 'pending';
  }
  return 'skipped';
}

/**
 * Render the profiles list.
 */
function renderProfiles(): void {
  const filter = profileFilterEl.value.toLowerCase();
  const statusFilter = profileStatusFilterEl.value;

  const filteredProfiles = profiles.filter((p) => {
    const matchesName = p.name.toLowerCase().includes(filter);

    // Apply status filter if set
    if (statusFilter) {
      const status = profileStatuses.get(p.id);
      const overallStatus = getOverallStatus(status);
      if (statusFilter === 'success' && overallStatus !== 'success') return false;
      if (statusFilter === 'failure' && overallStatus !== 'failure') return false;
      if (statusFilter === 'pending' && overallStatus !== 'pending') return false;
    }

    return matchesName;
  });

  if (filteredProfiles.length === 0) {
    profilesListEl.innerHTML = `
      <div class="empty-state">
        ${profiles.length === 0 ? 'No profiles found' : 'No profiles match the filter'}
      </div>
    `;
    return;
  }

  profilesListEl.innerHTML = filteredProfiles
    .map((profile) => {
      const status = profileStatuses.get(profile.id);
      const overallStatus = getOverallStatus(status);
      const safeId = escapeAttr(profile.id);
      const safeStatus = escapeAttr(overallStatus);

      return `
        <div class="list-item" data-profile-id="${safeId}">
          <div class="list-item-header">
            <div>
              <div class="list-item-title">
                <svg class="chevron" data-chevron-id="${safeId}" width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                  <path d="M6.22 3.22a.75.75 0 0 1 1.06 0l4.25 4.25a.75.75 0 0 1 0 1.06l-4.25 4.25a.75.75 0 0 1-1.06-1.06L9.94 8 6.22 4.28a.75.75 0 0 1 0-1.06z"/>
                </svg>
                ${escapeHtml(profile.name)}
              </div>
              <div class="list-item-meta">
                ${profile.labels ? escapeHtml(profile.labels.join(', ')) : 'No labels'}
              </div>
            </div>
            <span class="status-badge ${safeStatus}">
              <span class="status-dot ${safeStatus}"></span>
              ${escapeHtml(overallStatus.charAt(0).toUpperCase() + overallStatus.slice(1))}
            </span>
          </div>
        </div>
        <div class="expandable-content" data-content-id="${safeId}">
          ${renderProfileRules(status)}
        </div>
      `;
    })
    .join('');

  // Add click handlers for profile expansion using event delegation
  profilesListEl.querySelectorAll('.list-item').forEach((item) => {
    item.addEventListener('click', () => {
      const profileId = (item as HTMLElement).dataset.profileId;
      if (profileId) {
        toggleProfileExpand(profileId);
      }
    });
  });
}

/**
 * Render profile rules for expanded view.
 */
function renderProfileRules(status: ProfileStatusResult | undefined): string {
  if (
    !status?.rule_evaluation_status ||
    status.rule_evaluation_status.length === 0
  ) {
    return '<div class="empty-state">No rule evaluations</div>';
  }

  return `
    <div class="rule-list">
      ${status.rule_evaluation_status
        .map((rule) => {
          const safeStatus = escapeAttr(rule.status || 'pending');
          const entityName = rule.entity_info?.name || 'Unknown entity';
          const entityType = rule.entity_info?.entity_type || 'entity';

          return `
          <div class="rule-item-detailed">
            <div class="rule-item-header">
              <span class="rule-name">${escapeHtml(rule.rule_name || rule.rule_type_name || 'Unknown rule')}</span>
              <span class="status-badge ${safeStatus}">
                <span class="status-dot ${safeStatus}"></span>
                ${escapeHtml(rule.status || 'pending')}
              </span>
            </div>
            <div class="rule-item-details">
              <span class="rule-entity" title="${escapeAttr(entityType)}">
                <svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor" style="vertical-align: -2px; margin-right: 4px;">
                  <path d="M2 2.5A2.5 2.5 0 0 1 4.5 0h8.75a.75.75 0 0 1 .75.75v12.5a.75.75 0 0 1-.75.75h-2.5a.75.75 0 0 1 0-1.5h1.75v-2h-8a1 1 0 0 0-.714 1.7.75.75 0 1 1-1.072 1.05A2.495 2.495 0 0 1 2 11.5v-9zm10.5-1h-8a1 1 0 0 0-1 1v6.708A2.486 2.486 0 0 1 4.5 9h8.5V1.5zm-8 11h8v1h-8a1 1 0 0 1 0-2z"/>
                </svg>
                ${escapeHtml(entityName)}
              </span>
              <span class="rule-entity-type">${escapeHtml(entityType)}</span>
            </div>
          </div>
        `;
        })
        .join('')}
    </div>
  `;
}

/**
 * Toggle profile expansion.
 */
function toggleProfileExpand(profileId: string): void {
  const content = document.querySelector(
    `[data-content-id="${CSS.escape(profileId)}"]`
  );
  const chevron = document.querySelector(
    `[data-chevron-id="${CSS.escape(profileId)}"]`
  );

  if (content && chevron) {
    content.classList.toggle('expanded');
    chevron.classList.toggle('expanded');
  }
}

/**
 * Build a map of repository name to its rule evaluations.
 */
function getRepoRuleEvaluations(): Map<string, RuleEvaluationStatus[]> {
  const repoRules = new Map<string, RuleEvaluationStatus[]>();

  for (const status of profileStatuses.values()) {
    if (status.rule_evaluation_status) {
      for (const rule of status.rule_evaluation_status) {
        const entityName = rule.entity_info?.name;
        if (entityName) {
          const existing = repoRules.get(entityName) || [];
          existing.push(rule);
          repoRules.set(entityName, existing);
        }
      }
    }
  }

  return repoRules;
}

/**
 * Get overall status for a repository based on its rule evaluations.
 */
function getRepoStatus(
  rules: RuleEvaluationStatus[] | undefined
): 'success' | 'failure' | 'pending' | 'skipped' {
  if (!rules || rules.length === 0) {
    return 'pending';
  }

  const statuses = rules.map((r) => r.status);
  if (statuses.some((s) => s === 'failure' || s === 'error')) {
    return 'failure';
  }
  if (statuses.every((s) => s === 'success')) {
    return 'success';
  }
  if (statuses.some((s) => s === 'pending')) {
    return 'pending';
  }
  return 'skipped';
}

/**
 * Render the repositories list.
 */
function renderRepositories(): void {
  const filter = repoFilterEl.value.toLowerCase();
  const statusFilter = statusFilterEl.value;

  // Build repo -> rules mapping
  const repoRules = getRepoRuleEvaluations();

  const filteredRepos = repositories.filter((r) => {
    const name = `${r.owner}/${r.name}`.toLowerCase();
    const matchesFilter = name.includes(filter);

    // Apply status filter if set
    if (statusFilter && statusFilter !== 'all') {
      const repoName = `${r.owner}/${r.name}`;
      const rules = repoRules.get(repoName);
      const status = getRepoStatus(rules);
      if (statusFilter === 'passing' && status !== 'success') return false;
      if (statusFilter === 'failing' && status !== 'failure') return false;
    }

    return matchesFilter;
  });

  if (filteredRepos.length === 0) {
    repositoriesListEl.innerHTML = `
      <div class="empty-state">
        ${repositories.length === 0 ? 'No repositories found' : 'No repositories match the filter'}
      </div>
    `;
    return;
  }

  repositoriesListEl.innerHTML = filteredRepos
    .map((repo) => {
      const repoName = `${repo.owner}/${repo.name}`;
      const rules = repoRules.get(repoName);
      const status = getRepoStatus(rules);
      const safeRepoId = escapeAttr(repoName);
      const safeStatus = escapeAttr(status);
      const ruleCount = rules?.length || 0;
      const passingCount = rules?.filter((r) => r.status === 'success').length || 0;
      const failingCount = rules?.filter((r) => r.status === 'failure' || r.status === 'error').length || 0;

      return `
        <div class="list-item" data-repo-id="${safeRepoId}">
          <div class="list-item-header">
            <div>
              <div class="list-item-title">
                <svg class="chevron" data-repo-chevron="${safeRepoId}" width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                  <path d="M6.22 3.22a.75.75 0 0 1 1.06 0l4.25 4.25a.75.75 0 0 1 0 1.06l-4.25 4.25a.75.75 0 0 1-1.06-1.06L9.94 8 6.22 4.28a.75.75 0 0 1 0-1.06z"/>
                </svg>
                ${escapeHtml(repoName)}
              </div>
              <div class="list-item-meta">
                ${ruleCount > 0
                  ? `${passingCount} passing, ${failingCount} failing`
                  : 'No rule evaluations'}
              </div>
            </div>
            <span class="status-badge ${safeStatus}">
              <span class="status-dot ${safeStatus}"></span>
              ${escapeHtml(status.charAt(0).toUpperCase() + status.slice(1))}
            </span>
          </div>
        </div>
        <div class="expandable-content" data-repo-content="${safeRepoId}">
          ${renderRepoRules(rules)}
        </div>
      `;
    })
    .join('');

  // Add click handlers for repo expansion
  repositoriesListEl.querySelectorAll('.list-item[data-repo-id]').forEach((item) => {
    item.addEventListener('click', () => {
      const repoId = (item as HTMLElement).dataset.repoId;
      if (repoId) {
        toggleRepoExpand(repoId);
      }
    });
  });
}

/**
 * Render rules for a repository's expanded view.
 */
function renderRepoRules(rules: RuleEvaluationStatus[] | undefined): string {
  if (!rules || rules.length === 0) {
    return '<div class="empty-state">No rule evaluations for this repository</div>';
  }

  return `
    <div class="rule-list">
      ${rules
        .map((rule) => {
          const safeStatus = escapeAttr(rule.status || 'pending');
          return `
          <div class="rule-item-detailed">
            <div class="rule-item-header">
              <span class="rule-name">${escapeHtml(rule.rule_name || rule.rule_type_name || 'Unknown rule')}</span>
              <span class="status-badge ${safeStatus}">
                <span class="status-dot ${safeStatus}"></span>
                ${escapeHtml(rule.status || 'pending')}
              </span>
            </div>
          </div>
        `;
        })
        .join('')}
    </div>
  `;
}

/**
 * Toggle repository expansion.
 */
function toggleRepoExpand(repoId: string): void {
  const content = document.querySelector(
    `[data-repo-content="${CSS.escape(repoId)}"]`
  );
  const chevron = document.querySelector(
    `[data-repo-chevron="${CSS.escape(repoId)}"]`
  );

  if (content && chevron) {
    content.classList.toggle('expanded');
    chevron.classList.toggle('expanded');
  }
}

/**
 * Switch between tabs.
 */
function switchTab(tab: string): void {
  document
    .querySelectorAll('.tab')
    .forEach((t) => t.classList.remove('active'));
  document
    .querySelector(`[data-tab="${CSS.escape(tab)}"]`)
    ?.classList.add('active');

  const profilesSection = document.getElementById('profiles-section');
  const repositoriesSection = document.getElementById('repositories-section');

  if (profilesSection && repositoriesSection) {
    profilesSection.style.display = tab === 'profiles' ? 'block' : 'none';
    repositoriesSection.style.display =
      tab === 'repositories' ? 'block' : 'none';
  }
}

// Export for global access
declare global {
  interface Window {
    initDashboard: typeof initDashboard;
  }
}
window.initDashboard = initDashboard;
