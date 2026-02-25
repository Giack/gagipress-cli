import { renderBooks }   from './views/books.js';
import { renderIdeas }   from './views/ideas.js';
import { renderCalendar } from './views/calendar.js';

const routes = {
  '#/books':    renderBooks,
  '#/ideas':    renderIdeas,
  '#/calendar': renderCalendar,
};

export function dispatch() {
  const hash = location.hash || '#/ideas';
  const view = routes[hash] ?? renderIdeas;
  view();
}

window.addEventListener('hashchange', dispatch);
window.addEventListener('load', dispatch);
