import {
  MCPAppsClient,
  Profile,
  ProfileStatusResult,
  Repository,
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
  repoFilterEl = getRequiredElement<HTMLInputElement>('repo-filter');
  statusFilterEl = getRequiredElement<HTMLSelectElement>('status-filter');

  // Set up event listeners
  refreshBtn.addEventListener('click', loadDashboard);
  profileFilterEl.addEventListener('input', renderProfiles);
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

  // Initialize MCP client and load data
  mcpClient = new MCPAppsClient();
  loadDashboard();
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
    // Connect to the MCP host
    await client.connect();

    // Show loading state
    profilesListEl.innerHTML = `
      <div class="loading">
        <div class="spinner"></div>
        Loading profiles...
      </div>
    `;

    // Load profiles
    try {
      const profilesResult = await client.listProfiles();
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
        const status = await client.getProfileStatus({ name: profile.name });
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

    // Load repositories
    try {
      const reposResult = await client.listRepositories();
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

  const filteredProfiles = profiles.filter((p) =>
    p.name.toLowerCase().includes(filter)
  );

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
          return `
          <div class="rule-item">
            <span class="rule-name">${escapeHtml(rule.rule_name || rule.rule_type || 'Unknown rule')}</span>
            <span class="status-badge ${safeStatus}">
              <span class="status-dot ${safeStatus}"></span>
              ${escapeHtml(rule.status || 'pending')}
            </span>
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
 * Render the repositories list.
 */
function renderRepositories(): void {
  const filter = repoFilterEl.value.toLowerCase();
  const statusFilter = statusFilterEl.value;

  const filteredRepos = repositories.filter((r) => {
    const name = `${r.owner}/${r.name}`.toLowerCase();
    return name.includes(filter);
  });

  // Note: statusFilter is available for future use when we have per-repo compliance data
  void statusFilter; // Explicitly mark as intentionally unused for now

  if (filteredRepos.length === 0) {
    repositoriesListEl.innerHTML = `
      <div class="empty-state">
        ${repositories.length === 0 ? 'No repositories found' : 'No repositories match the filter'}
      </div>
    `;
    return;
  }

  repositoriesListEl.innerHTML = filteredRepos
    .map(
      (repo) => `
      <div class="list-item">
        <div class="list-item-header">
          <div>
            <div class="list-item-title">
              <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                <path d="M2 2.5A2.5 2.5 0 0 1 4.5 0h8.75a.75.75 0 0 1 .75.75v12.5a.75.75 0 0 1-.75.75h-2.5a.75.75 0 0 1 0-1.5h1.75v-2h-8a1 1 0 0 0-.714 1.7.75.75 0 1 1-1.072 1.05A2.495 2.495 0 0 1 2 11.5v-9zm10.5-1h-8a1 1 0 0 0-1 1v6.708A2.486 2.486 0 0 1 4.5 9h8.5V1.5zm-8 11h8v1h-8a1 1 0 0 1 0-2z"/>
              </svg>
              ${escapeHtml(repo.owner)}/${escapeHtml(repo.name)}
            </div>
            <div class="list-item-meta">
              Provider: ${escapeHtml(repo.provider || 'unknown')}
            </div>
          </div>
        </div>
      </div>
    `
    )
    .join('');
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
