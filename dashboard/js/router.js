// dashboard/js/router.js
// Hash router — maps location.hash to view render functions.
// updateNav() is called on every dispatch (including initial load) to sync active tab state.
import { renderBooks }   from './views/books.js';
import { renderIdeas }   from './views/ideas.js';
import { renderCalendar } from './views/calendar.js';

const routes = {
  '#/books':    renderBooks,
  '#/ideas':    renderIdeas,
  '#/calendar': renderCalendar,
};

/**
 * Highlight the nav link that matches the current hash.
 * Uses [data-nav] selector — requires data-nav attributes on <a> elements in index.html.
 * Active state: text-white + border-b-2 + border-indigo-500
 * Inactive state: text-gray-400
 * @param {string} hash — current location.hash value (e.g. '#/books')
 */
function updateNav(hash) {
  document.querySelectorAll('[data-nav]').forEach(link => {
    const isActive = link.dataset.nav === hash;
    link.classList.toggle('text-white', isActive);
    link.classList.toggle('border-b-2', isActive);
    link.classList.toggle('border-indigo-500', isActive);
    link.classList.toggle('text-gray-400', !isActive);
  });
}

export function dispatch() {
  const hash = location.hash || '#/ideas';
  const view = routes[hash] ?? renderIdeas;
  updateNav(hash);
  view();
}

window.addEventListener('hashchange', dispatch);
window.addEventListener('load', dispatch);
