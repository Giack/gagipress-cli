// dashboard/js/api.js
// Fetch wrapper for Supabase PostgREST queries.
// Always resolves — errors returned as { data: null, error: string }, never thrown.
// NOTE: All imports use .js extensions — no bundler, native browser ES modules.
import { supabase } from '../config.js';

/**
 * Fetch all rows from a Supabase table.
 * @param {string} table — table name matching PostgREST endpoint
 * @param {Object} [options]
 * @param {string} [options.order] — column name to order by ascending (optional)
 * @returns {Promise<{ data: Array|null, error: string|null }>}
 *   On success: { data: Array (may be empty), error: null }
 *   On failure: { data: null, error: 'message string' }
 */
export async function fetchTable(table, options = {}) {
  let query = supabase.from(table).select('*');
  if (options.order) {
    query = query.order(options.order, { ascending: true });
  }
  const { data, error } = await query;
  if (error) {
    return { data: null, error: error.message ?? 'Unknown fetch error' };
  }
  return { data: data ?? [], error: null };
}
