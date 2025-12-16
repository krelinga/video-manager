-- Remove ingestion_state and ingestion_error columns from media_dvds
-- Ingestion status is now tracked in the tasks table

ALTER TABLE media_dvds DROP COLUMN IF EXISTS ingestion_state;
ALTER TABLE media_dvds DROP COLUMN IF EXISTS ingestion_error;

-- Drop the enum type
DROP TYPE IF EXISTS media_dvd_ingestion_state;

-- Add index for looking up dvd_ingestion tasks by media_id
CREATE INDEX IF NOT EXISTS idx_tasks_dvd_ingestion_media_id
ON tasks (task_type, ((state->>'media_id')::integer))
WHERE task_type = 'dvd_ingestion';
