-- Drop all tables and related objects in reverse order of creation.

-- Drop the index for dvd_ingestion task lookups
DROP INDEX IF EXISTS idx_tasks_dvd_ingestion_media_id;

-- Drop index for child tasks
DROP INDEX IF EXISTS idx_tasks_parent_id;

-- Drop trigger and function for tasks
DROP TRIGGER IF EXISTS trg_update_tasks_updated_at ON tasks;
DROP FUNCTION IF EXISTS update_tasks_updated_at();

-- Drop tasks table and indexes
DROP INDEX IF EXISTS idx_tasks_type;
DROP INDEX IF EXISTS idx_tasks_claimable;
DROP TABLE IF EXISTS tasks;

-- Drop task_status enum
DROP TYPE IF EXISTS task_status;

-- Drop media and catalog tables
DROP TABLE IF EXISTS media_sets_x_cards;

DROP TABLE IF EXISTS media_x_cards;

DROP TRIGGER IF EXISTS trg_delete_dvd_media ON media_dvds;
DROP FUNCTION IF EXISTS delete_dvd_media();
DROP TABLE IF EXISTS media_dvds;

DROP TABLE IF EXISTS media;

DROP TABLE IF EXISTS media_sets;

DROP TRIGGER IF EXISTS trg_delete_movie_edition_card ON catalog_movie_editions;
DROP FUNCTION IF EXISTS delete_movie_edition_card();
DROP TABLE IF EXISTS catalog_movie_editions;

DROP INDEX IF EXISTS unique_default_true;
DROP TABLE IF EXISTS catalog_movie_edition_kinds;

DROP TRIGGER IF EXISTS trg_delete_movie_card ON catalog_movies;
DROP FUNCTION IF EXISTS delete_movie_card();
DROP TABLE IF EXISTS catalog_movies;

DROP TABLE IF EXISTS catalog_cards;
