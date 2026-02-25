// NOTE: All imports must use .js extensions — no bundler resolution in browser
import { fetchTable } from '../api.js';
import { renderError, renderEmpty } from '../components.js';

function escapeHtml(str) {
  return String(str ?? '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function renderBooksSkeleton() {
  const skeletonRow = `
    <tr class="border-b border-gray-700">
      <td class="py-3 px-4"><div class="h-4 bg-gray-700 rounded animate-pulse w-40"></div></td>
      <td class="py-3 px-4"><div class="h-4 bg-gray-700 rounded animate-pulse w-24"></div></td>
      <td class="py-3 px-4"><div class="h-4 bg-gray-700 rounded animate-pulse w-20"></div></td>
      <td class="py-3 px-4"><div class="h-4 bg-gray-700 rounded animate-pulse w-32"></div></td>
    </tr>`;
  return `
    <div class="max-w-4xl overflow-x-auto">
      <h1 class="text-2xl font-bold text-white mb-6">Books</h1>
      <table class="w-full text-sm text-left">
        <thead class="sticky top-0 bg-gray-800 text-gray-400 uppercase text-xs">
          <tr>
            <th class="py-3 px-4">Title</th>
            <th class="py-3 px-4">ASIN</th>
            <th class="py-3 px-4">Genre</th>
            <th class="py-3 px-4">Target Audience</th>
          </tr>
        </thead>
        <tbody>
          ${skeletonRow.repeat(4)}
        </tbody>
      </table>
    </div>`;
}

function renderBooksTable(books) {
  const rows = books.map(book => `
    <tr class="border-b border-gray-700 hover:bg-gray-800 transition-colors">
      <td class="py-3 px-4 max-w-xs truncate" title="${escapeHtml(book.title)}">${escapeHtml(book.title)}</td>
      <td class="py-3 px-4"><a href="https://www.amazon.com/dp/${escapeHtml(book.kdp_asin ?? '')}" target="_blank" rel="noopener" class="text-indigo-400 hover:underline">${escapeHtml(book.kdp_asin ?? '—')}</a></td>
      <td class="py-3 px-4 text-gray-300">${escapeHtml(book.genre ?? '—')}</td>
      <td class="py-3 px-4 text-gray-300">${escapeHtml(book.target_audience ?? '—')}</td>
    </tr>`).join('');
  return `
    <div class="max-w-4xl overflow-x-auto">
      <h1 class="text-2xl font-bold text-white mb-6">Books</h1>
      <table class="w-full text-sm text-left">
        <thead class="sticky top-0 bg-gray-800 text-gray-400 uppercase text-xs">
          <tr>
            <th class="py-3 px-4">Title</th>
            <th class="py-3 px-4">ASIN</th>
            <th class="py-3 px-4">Genre</th>
            <th class="py-3 px-4">Target Audience</th>
          </tr>
        </thead>
        <tbody class="text-gray-100">
          ${rows}
        </tbody>
      </table>
    </div>`;
}

export async function renderBooks() {
  const app = document.getElementById('app');
  app.innerHTML = renderBooksSkeleton();
  const { data, error } = await fetchTable('books', { order: 'title' });
  if (error) {
    app.innerHTML = renderError(error);
  } else if (data.length === 0) {
    app.innerHTML = renderEmpty('No books in your catalog yet');
  } else {
    app.innerHTML = renderBooksTable(data);
  }
}
