// Copy this file to config.js and fill in your Supabase credentials
// dashboard/config.js is gitignored — never commit actual credentials

import { createClient } from 'https://cdn.jsdelivr.net/npm/@supabase/supabase-js@2.97.0/+esm';

export const SUPABASE_URL = "https://YOUR_PROJECT_REF.supabase.co";
export const SUPABASE_ANON_KEY = "your-anon-key-here";

export const supabase = createClient(SUPABASE_URL, SUPABASE_ANON_KEY);
