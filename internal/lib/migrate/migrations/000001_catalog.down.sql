-- Drop all tables and related objects in reverse order of creation.
-- Group tables together with their immeidate dependencies (functions and triggers) for clarity.

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

DROP TYPE IF EXISTS media_dvd_ingestion_state;
