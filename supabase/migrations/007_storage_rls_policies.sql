-- Allow CLI (anon key) to upload images to campaign-media bucket
CREATE POLICY "allow_insert_campaign_media"
ON storage.objects FOR INSERT
WITH CHECK (bucket_id = 'campaign-media');

-- Allow public reads (needed for public URLs via Blotato / Edge Function)
CREATE POLICY "allow_select_campaign_media"
ON storage.objects FOR SELECT
USING (bucket_id = 'campaign-media');
