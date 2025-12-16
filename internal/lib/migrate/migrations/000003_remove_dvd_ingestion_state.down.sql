-- Drop the index for dvd_ingestion task lookups
DROP INDEX IF EXISTS idx_tasks_dvd_ingestion_media_id;

-- Recreate the enum type
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'media_dvd_ingestion_state') THEN
    CREATE TYPE media_dvd_ingestion_state AS ENUM ('pending', 'done', 'error');
  END IF;
END
$$;

-- Re-add ingestion_state and ingestion_error columns to media_dvds
ALTER TABLE media_dvds ADD COLUMN IF NOT EXISTS ingestion_state media_dvd_ingestion_state NOT NULL DEFAULT 'pending';
ALTER TABLE media_dvds ADD COLUMN IF NOT EXISTS ingestion_error TEXT;

-- Re-add the check constraint
ALTER TABLE media_dvds ADD CONSTRAINT chk_ingestion_error CHECK (
    (ingestion_state = 'error' AND ingestion_error IS NOT NULL) OR
    (ingestion_state <> 'error' AND ingestion_error IS NULL)
);
