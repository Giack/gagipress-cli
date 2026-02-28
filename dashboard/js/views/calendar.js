// NOTE: All imports must use .js extensions — no bundler resolution in browser
import { fetchTable } from '../api.js';
import { renderError, renderEmpty } from '../components.js';

function escapeHtml(str) {
  return String(str ?? '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

const COLUMNS = [
  { label: 'scheduled',   dbStatus: 'pending_approval', labelClass: 'text-blue-400'   },
  { label: 'approved',    dbStatus: 'approved',          labelClass: 'text-green-400'  },
  { label: 'publishing',  dbStatus: 'publishing',        labelClass: 'text-yellow-400' },
  { label: 'published',   dbStatus: 'published',         labelClass: 'text-indigo-400' },
  { label: 'failed',      dbStatus: 'failed',            labelClass: 'text-red-400'    },
];

const PLATFORM_BADGE_CLASSES = {
  tiktok:    'text-pink-400',
  instagram: 'text-purple-400',
};

function groupByDbStatus(entries) {
  const groups = {};
  for (const col of COLUMNS) groups[col.dbStatus] = [];
  for (const entry of entries) {
    if (groups[entry.status] !== undefined) groups[entry.status].push(entry);
  }
  return groups;
}

function renderCalendarSkeleton() {
  const skeletonCard = `
    <div class="bg-gray-800 rounded-lg p-3">
      <div class="h-4 bg-gray-700 rounded animate-pulse w-full mb-2"></div>
      <div class="h-3 bg-gray-700 rounded animate-pulse w-16"></div>
    </div>`;
  const skeletonColumn = `
    <div class="flex-shrink-0 w-56">
      <div class="h-4 bg-gray-700 rounded animate-pulse w-24 mb-3"></div>
      <div class="flex flex-col gap-2">
        ${skeletonCard.repeat(3)}
      </div>
    </div>`;
  return `
    <h1 class="text-2xl font-bold text-white mb-6">Calendar</h1>
    <div class="flex gap-4 overflow-x-auto">
      ${skeletonColumn.repeat(5)}
    </div>`;
}

function renderCard(entry, ideaTitle) {
  const date = entry.scheduled_for
    ? new Date(entry.scheduled_for).toLocaleDateString()
    : '—';
  const badgeClass = PLATFORM_BADGE_CLASSES[entry.platform] ?? 'text-gray-400';
  return `
    <div class="bg-gray-800 rounded-lg p-3">
      <p class="text-white text-sm font-medium truncate">${escapeHtml(ideaTitle)}</p>
      <span class="text-xs ${badgeClass} mt-1 block">${escapeHtml(entry.platform)}</span>
      <span class="text-xs text-gray-400 mt-1 block">${escapeHtml(date)}</span>
    </div>`;
}

function renderCalendarKanban(entries, scriptsMap, ideasMap) {
  const grouped = groupByDbStatus(entries);
  const columns = COLUMNS.map(col => {
    const colEntries = grouped[col.dbStatus];
    const cardsHtml = colEntries.length === 0
      ? renderEmpty('Nothing here yet')
      : `<div class="flex flex-col gap-2">${colEntries.map(e => {
          const ideaId = scriptsMap.get(e.script_id);
          const title = ideaId ? (ideasMap.get(ideaId) ?? '(no title)') : '(no title)';
          return renderCard(e, title);
        }).join('')}</div>`;
    return `
      <div class="flex-shrink-0 w-56">
        <h2 class="text-sm font-semibold uppercase tracking-wide mb-3 ${col.labelClass}">${escapeHtml(col.label)}</h2>
        ${cardsHtml}
      </div>`;
  }).join('');
  return `
    <h1 class="text-2xl font-bold text-white mb-6">Calendar</h1>
    <div class="flex gap-4 overflow-x-auto pb-4">
      ${columns}
    </div>`;
}

export async function renderCalendar() {
  const app = document.getElementById('app');
  app.innerHTML = renderCalendarSkeleton();  // line 1 — sync, no await

  const [calResult, scriptsResult, ideasResult] = await Promise.all([
    fetchTable('content_calendar', { order: 'scheduled_for' }),
    fetchTable('content_scripts'),
    fetchTable('content_ideas'),
  ]);

  if (calResult.error) {
    app.innerHTML = renderError(calResult.error);
    return;
  }

  const entries = calResult.data;

  if (entries.length === 0) {
    app.innerHTML = renderEmpty('No calendar entries yet — plan some with the CLI');
    return;
  }

  // scriptsMap: script.id → script.idea_id
  const scriptsMap = new Map();
  if (scriptsResult.data) {
    for (const s of scriptsResult.data) {
      scriptsMap.set(s.id, s.idea_id);
    }
  }

  // ideasMap: idea.id → idea.brief_description
  const ideasMap = new Map();
  if (ideasResult.data) {
    for (const idea of ideasResult.data) {
      ideasMap.set(idea.id, idea.brief_description);
    }
  }

  app.innerHTML = renderCalendarKanban(entries, scriptsMap, ideasMap);
}
