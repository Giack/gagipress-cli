import "jsr:@supabase/functions-js/edge-runtime.d.ts";
import { createClient } from "jsr:@supabase/supabase-js@2";

const BLOTATO_BASE_URL = "https://backend.blotato.com/v2";
const MEDIA_TIMEOUT_MS = 2 * 60 * 1000; // 2 minutes

// ─── Types ────────────────────────────────────────────────────────────────────

interface CalendarEntry {
  id: string;
  script_id: string | null;
  platform: string;
  scheduled_for: string;
  generate_media: boolean;
  media_url: string | null;
}

interface ContentScript {
  id: string;
  hook: string;
  full_script: string;
  cta: string;
  hashtags: string[];
}

// ─── Blotato helpers (mirrors internal/social/blotato.go) ─────────────────────

async function getBlotatoAccountId(
  apiKey: string,
  platform: string,
): Promise<string> {
  const res = await fetch(
    `${BLOTATO_BASE_URL}/users/me/accounts?platform=${platform}`,
    { headers: { "blotato-api-key": apiKey, "Content-Type": "application/json" } },
  );
  if (!res.ok) throw new Error(`Blotato accounts error: ${await res.text()}`);
  const data = await res.json();
  if (!data.items?.length) throw new Error(`No Blotato account for platform: ${platform}`);
  return data.items[0].id;
}

async function generateVisual(
  apiKey: string,
  templateId: string,
  prompt: string,
): Promise<string> {
  const res = await fetch(`${BLOTATO_BASE_URL}/videos/from-templates`, {
    method: "POST",
    headers: { "blotato-api-key": apiKey, "Content-Type": "application/json" },
    body: JSON.stringify({ templateId, inputs: {}, prompt, render: true }),
  });
  if (!res.ok) throw new Error(`Blotato visual create error: ${await res.text()}`);
  const data = await res.json();
  return data.item.id;
}

async function waitForVisual(
  apiKey: string,
  creationId: string,
): Promise<string | null> {
  const deadline = Date.now() + MEDIA_TIMEOUT_MS;
  const TERMINAL_FAILURE = "creation-from-template-failed";
  const DONE = "done";

  while (Date.now() < deadline) {
    const res = await fetch(`${BLOTATO_BASE_URL}/videos/creations/${creationId}`, {
      headers: { "blotato-api-key": apiKey },
    });
    if (!res.ok) throw new Error(`Blotato polling error: ${await res.text()}`);
    const data = await res.json();
    const status = data.item?.status;

    if (status === DONE) {
      return data.item.mediaUrl || data.item.imageUrls?.[0] || null;
    }
    if (status === TERMINAL_FAILURE) throw new Error("Blotato visual generation failed");

    await new Promise((r) => setTimeout(r, 5000));
  }
  return null; // timeout → caller uses text-only fallback
}

async function publishPost(
  apiKey: string,
  accountId: string,
  platform: string,
  text: string,
  mediaUrls: string[],
  scheduledTime: string | null,
): Promise<string> {
  const body: Record<string, unknown> = {
    post: {
      accountId,
      content: { text, mediaUrls, platform },
      target: { targetType: platform },
    },
  };
  if (scheduledTime) body.scheduledTime = scheduledTime;

  const res = await fetch(`${BLOTATO_BASE_URL}/posts`, {
    method: "POST",
    headers: { "blotato-api-key": apiKey, "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`Blotato publish error: ${await res.text()}`);
  const data = await res.json();
  return data.postSubmissionId;
}

// ─── Main handler ─────────────────────────────────────────────────────────────

Deno.serve(async (_req) => {
  const supabaseUrl = Deno.env.get("SUPABASE_URL")!;
  const serviceRoleKey = Deno.env.get("SUPABASE_SERVICE_ROLE_KEY")!;
  const blotatoApiKey = Deno.env.get("BLOTATO_API_KEY") ?? "";
  const blotatoTemplateId = Deno.env.get("BLOTATO_TEMPLATE_ID") ?? "";

  if (!blotatoApiKey) {
    return new Response(
      JSON.stringify({ error: "BLOTATO_API_KEY secret not set" }),
      { status: 500, headers: { "Content-Type": "application/json" } },
    );
  }

  const supabase = createClient(supabaseUrl, serviceRoleKey);

  // ── Step 1: Rollback stale locks (entries stuck in 'publishing' > 10 min) ──
  const { error: rollbackError } = await supabase
    .from("content_calendar")
    .update({ status: "approved" })
    .eq("status", "publishing")
    .lt("updated_at", new Date(Date.now() - 10 * 60 * 1000).toISOString());

  if (rollbackError) {
    console.error("Rollback stale locks error:", rollbackError.message);
  }

  // ── Step 2: Atomic lock — grab entries that are due now ───────────────────
  const now = new Date().toISOString();
  const { data: entries, error: lockError } = await supabase
    .from("content_calendar")
    .update({ status: "publishing" })
    .eq("status", "approved")
    .lte("scheduled_for", now)
    .is("published_at", null)
    .select("id, script_id, platform, scheduled_for, generate_media, media_url")
    .returns<CalendarEntry[]>();

  if (lockError) {
    return new Response(
      JSON.stringify({ error: `Lock query failed: ${lockError.message}` }),
      { status: 500, headers: { "Content-Type": "application/json" } },
    );
  }

  if (!entries || entries.length === 0) {
    return new Response(
      JSON.stringify({ processed: 0, published: 0, failed: 0 }),
      { headers: { "Content-Type": "application/json" } },
    );
  }

  console.log(`Processing ${entries.length} entries`);

  // ── Step 3: Publish each locked entry ────────────────────────────────────
  let published = 0;
  let failed = 0;
  const accountIdCache: Record<string, string> = {};

  for (const entry of entries) {
    try {
      if (!entry.script_id) throw new Error("No script attached to calendar entry");

      // Fetch script
      const { data: scripts, error: scriptError } = await supabase
        .from("content_scripts")
        .select("id, hook, full_script, cta, hashtags")
        .eq("id", entry.script_id)
        .returns<ContentScript[]>();

      if (scriptError || !scripts?.length) {
        throw new Error(`Script fetch failed: ${scriptError?.message ?? "not found"}`);
      }
      const script = scripts[0];

      // Build post text (mirrors publish.go runBatchPublish)
      const hashtagLine = script.hashtags?.length
        ? "\n\n" + script.hashtags.join(" ")
        : "";
      const postText =
        `${script.hook}\n\n${script.full_script}\n\n${script.cta}${hashtagLine}`;

      // Get Blotato account (cached per platform)
      if (!accountIdCache[entry.platform]) {
        accountIdCache[entry.platform] = await getBlotatoAccountId(
          blotatoApiKey,
          entry.platform,
        );
      }
      const accountId = accountIdCache[entry.platform];

      // Optional media: use pre-generated Supabase Storage image if available,
      // otherwise fall back to Blotato template generation.
      const mediaUrls: string[] = [];
      if (entry.media_url) {
        // Pre-generated by `gagipress calendar generate-media`
        mediaUrls.push(entry.media_url);
      } else if (entry.generate_media && blotatoTemplateId) {
        try {
          const prompt = `Create a promotional visual for a book post.\nHook: ${script.hook}\nMain topic: ${script.full_script}`;
          const creationId = await generateVisual(blotatoApiKey, blotatoTemplateId, prompt);
          const mediaUrl = await waitForVisual(blotatoApiKey, creationId);
          if (mediaUrl) mediaUrls.push(mediaUrl);
          else console.warn(`Media timeout for entry ${entry.id}, publishing text-only`);
        } catch (mediaErr) {
          console.warn(`Media generation failed for ${entry.id}, falling back to text-only:`, mediaErr);
        }
      }

      // Publish
      const scheduledTime = entry.scheduled_for
        ? new Date(entry.scheduled_for).toISOString()
        : null;
      await publishPost(blotatoApiKey, accountId, entry.platform, postText, mediaUrls, scheduledTime);

      // Mark published
      await supabase
        .from("content_calendar")
        .update({ status: "published", published_at: new Date().toISOString() })
        .eq("id", entry.id);

      published++;
      console.log(`Published entry ${entry.id}`);
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      console.error(`Failed entry ${entry.id}:`, msg);

      await supabase
        .from("content_calendar")
        .update({ status: "failed", publish_errors: msg })
        .eq("id", entry.id);

      failed++;
    }
  }

  const result = { processed: entries.length, published, failed };
  console.log("Run complete:", result);
  return new Response(JSON.stringify(result), {
    headers: { "Content-Type": "application/json" },
  });
});
