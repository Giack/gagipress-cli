-- Allow DELETE and UPDATE on campaign-media bucket (needed for regenerating images)
CREATE POLICY allow_delete_campaign_media ON storage.objects
  FOR DELETE TO anon
  USING (bucket_id = 'campaign-media');

CREATE POLICY allow_update_campaign_media ON storage.objects
  FOR UPDATE TO anon
  USING (bucket_id = 'campaign-media');
