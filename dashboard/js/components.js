// dashboard/js/components.js
// Pure UI state renderer functions — return HTML strings, no DOM side effects.
// Called by views before fetch (loading) and after fetch resolves (error/empty/content).
// Dark palette: matches bg-gray-900 body established in index.html.

/**
 * Spinner shown while data is loading.
 * Uses Tailwind animate-spin — class must be injected into DOM to trigger JIT compilation.
 * @returns {string} HTML string
 */
export function renderLoading() {
  return `
    <div class="flex items-center gap-3 text-gray-400 py-12">
      <svg class="animate-spin h-5 w-5 text-indigo-400"
           xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10"
                stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
      </svg>
      <span>Loading...</span>
    </div>`;
}

/**
 * Error banner shown when a fetch fails.
 * @param {string} message — error message (HTML-escaped before injection)
 * @returns {string} HTML string
 */
export function renderError(message) {
  return `
    <div class="bg-red-900/40 border border-red-700 text-red-300 rounded-lg px-4 py-3 mt-4">
      <strong class="font-semibold">Error:</strong> ${escapeHtml(message)}
    </div>`;
}

/**
 * Empty state shown when fetch succeeds but returns no rows.
 * @param {string} [message] — custom message (defaults to generic text)
 * @returns {string} HTML string
 */
export function renderEmpty(message = 'No items found.') {
  return `
    <div class="text-gray-500 italic py-8 text-center">${escapeHtml(message)}</div>`;
}

/**
 * Escape HTML special characters to prevent broken UI when error messages
 * contain <, >, &, or " characters (e.g., from Supabase error messages).
 * Not a security concern (no untrusted users) — prevents broken HTML structure.
 * @param {string} str
 * @returns {string}
 */
function escapeHtml(str) {
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}
