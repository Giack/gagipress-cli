// NOTE: All imports must use .js extensions — no bundler resolution in browser
import { fetchTable } from '../api.js';
import { renderError, renderEmpty } from '../components.js';

function escapeHtml(str) {
  return String(str ?? '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

const COLUMNS = ['pending', 'approved', 'rejected', 'scripted'];
const COLUMN_LABEL_CLASSES = {
  pending:  'text-yellow-400',
  approved: 'text-green-400',
  rejected: 'text-red-400',
  scripted: 'text-indigo-400',
};

function groupByStatus(ideas) {
  const groups = { pending: [], approved: [], rejected: [], scripted: [] };
  for (const idea of ideas) {
    if (groups[idea.status]) groups[idea.status].push(idea);
  }
  return groups;
}

function renderIdeasSkeleton() {
  const skeletonCard = `
    <div class="bg-gray-800 rounded-lg p-3">
      <div class="h-4 bg-gray-700 rounded animate-pulse w-full mb-2"></div>
      <div class="h-3 bg-gray-700 rounded animate-pulse w-16"></div>
    </div>`;
  const skeletonColumn = `
    <div class="flex-shrink-0 w-64">
      <div class="h-4 bg-gray-700 rounded animate-pulse w-24 mb-3"></div>
      <div class="flex flex-col gap-2">
        ${skeletonCard.repeat(3)}
      </div>
    </div>`;
  return `
    <h1 class="text-2xl font-bold text-white mb-6">Ideas</h1>
    <div class="flex gap-4 overflow-x-auto">
      ${skeletonColumn.repeat(4)}
    </div>`;
}

function renderCard(idea) {
  return `
    <div class="bg-gray-800 rounded-lg p-3 cursor-pointer hover:bg-gray-700 transition-colors" data-idea-id="${escapeHtml(idea.id)}">
      <p class="text-white text-sm font-medium truncate">${escapeHtml(idea.brief_description)}</p>
      <span class="text-xs text-gray-400 mt-1 block">${escapeHtml(idea.type)}</span>
      <div class="hidden mt-2 text-xs text-gray-300 whitespace-pre-wrap border-t border-gray-700 pt-2" data-script-preview></div>
    </div>`;
}

function renderIdeasKanban(ideas) {
  const grouped = groupByStatus(ideas);
  const columns = COLUMNS.map(status => {
    const columnIdeas = grouped[status];
    const cardsHtml = columnIdeas.length === 0
      ? renderEmpty('No ideas yet')
      : `<div class="flex flex-col gap-2">${columnIdeas.map(renderCard).join('')}</div>`;
    return `
      <div class="flex-shrink-0 w-64">
        <h2 class="text-sm font-semibold uppercase tracking-wide mb-3 ${COLUMN_LABEL_CLASSES[status]}">${status}</h2>
        ${cardsHtml}
      </div>`;
  }).join('');
  return `
    <h1 class="text-2xl font-bold text-white mb-6">Ideas</h1>
    <div class="flex gap-4 overflow-x-auto pb-4">
      ${columns}
    </div>`;
}

export async function renderIdeas() {
  const app = document.getElementById('app');
  app.innerHTML = renderIdeasSkeleton();  // line 1 — sync, no await

  // Fetch ideas and scripts in parallel
  const [ideasResult, scriptsResult] = await Promise.all([
    fetchTable('content_ideas', { order: 'generated_at' }),
    fetchTable('content_scripts'),
  ]);

  if (ideasResult.error) {
    app.innerHTML = renderError(ideasResult.error);
    return;
  }

  const ideas = ideasResult.data;

  if (ideas.length === 0) {
    app.innerHTML = renderEmpty('No ideas yet — generate some with the CLI');
    return;
  }

  // Build lookup map: idea_id -> script row
  const scriptsMap = new Map();
  if (scriptsResult.data) {
    for (const s of scriptsResult.data) {
      scriptsMap.set(s.idea_id, s);
    }
  }

  app.innerHTML = renderIdeasKanban(ideas);

  // Event delegation — one listener on #app, not per card
  app.addEventListener('click', (e) => {
    const card = e.target.closest('[data-idea-id]');
    if (!card) return;
    const ideaId = card.dataset.ideaId;
    const previewEl = card.querySelector('[data-script-preview]');
    if (!previewEl) return;
    // Populate on first expand (lazy but synchronous — data already in memory)
    if (!previewEl.dataset.loaded) {
      const script = scriptsMap.get(ideaId);
      previewEl.textContent = script ? script.full_script : 'Script not found.';
      previewEl.dataset.loaded = 'true';
    }
    previewEl.classList.toggle('hidden');
  });
}
