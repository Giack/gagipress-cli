// NOTE: All imports must use .js extensions — no bundler resolution in browser
export function renderBooks() {
  const app = document.getElementById('app');
  app.innerHTML = `
    <div class="max-w-4xl">
      <h1 class="text-2xl font-bold text-white mb-6">Books</h1>
      <p class="text-gray-400">Books table coming in Phase 3.</p>
    </div>
  `;
}
