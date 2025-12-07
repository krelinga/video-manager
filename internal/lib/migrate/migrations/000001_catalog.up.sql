-- Create catalog_cards table
CREATE TABLE IF NOT EXISTS catalog_cards (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL CHECK (name <> ''),
    note TEXT
);

-- Create catalog_movies table
-- movies are a kind of catalog_card, and so they share a primary key.
CREATE TABLE IF NOT EXISTS catalog_movies (
    card_id INTEGER PRIMARY KEY,
    release_year INTEGER,
    tmdb_id INTEGER,
    fanart_id TEXT,
    CONSTRAINT fk_catalog_movies_card_id 
        FOREIGN KEY (card_id) REFERENCES catalog_cards(id) ON DELETE CASCADE
);

-- A trigger to ensure that the catalog_card for a movie is deleted when the corresponding catalog_movie is deleted.
CREATE OR REPLACE FUNCTION delete_movie_card()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM catalog_cards WHERE id = OLD.card_id;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_delete_movie_card ON catalog_movies;

CREATE TRIGGER trg_delete_movie_card
AFTER DELETE ON catalog_movies
FOR EACH ROW
EXECUTE FUNCTION delete_movie_card();

-- Create catalog_movie_edition_kinds table
CREATE TABLE IF NOT EXISTS catalog_movie_edition_kinds (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE
);

-- Enforce at most one TRUE value for is_default
CREATE UNIQUE INDEX unique_default_true
ON catalog_movie_edition_kinds (is_default)
WHERE is_default;

-- Create catalog_movie_editions table
-- movie_editions are a kind of catalog_card, and so they cshare a primary key.
-- The movie_id field references the parent movie's card_id, and must be set.
CREATE TABLE IF NOT EXISTS catalog_movie_editions (
    card_id INTEGER PRIMARY KEY,
    kind_id INTEGER NOT NULL,
    movie_card_id INTEGER NOT NULL,
    CONSTRAINT fk_catalog_movie_editions_card_id 
        FOREIGN KEY (card_id) REFERENCES catalog_cards(id) ON DELETE CASCADE,
    CONSTRAINT fk_catalog_movie_editions_kind_id 
        FOREIGN KEY (kind_id) REFERENCES catalog_movie_edition_kinds(id),
    CONSTRAINT fk_catalog_movie_editions_movie_card_id 
        FOREIGN KEY (movie_card_id) REFERENCES catalog_movies(card_id) ON DELETE CASCADE
);

-- A trigger to ensure that the catalog_card for a movie edition is deleted when the corresponding catalog_movie_edition is deleted.
CREATE OR REPLACE FUNCTION delete_movie_edition_card()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM catalog_cards WHERE id = OLD.card_id;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_delete_movie_edition_card ON catalog_movie_editions;

CREATE TRIGGER trg_delete_movie_edition_card
AFTER DELETE ON catalog_movie_editions
FOR EACH ROW
EXECUTE FUNCTION delete_movie_edition_card();

-- Create media_sets table
CREATE TABLE IF NOT EXISTS media_sets (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL CHECK (name <> ''),
    note TEXT
);

-- Create media table
CREATE TABLE IF NOT EXISTS media (
    id SERIAL PRIMARY KEY,
    media_set_id INTEGER,
    note TEXT,
    CONSTRAINT fk_media_media_set_id 
        FOREIGN KEY (media_set_id) REFERENCES media_sets(id)
);

-- Create media_dvd_ingestion_state enum type (if it doesn't already exist)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'media_dvd_ingestion_state') THEN
    CREATE TYPE media_dvd_ingestion_state AS ENUM ('pending', 'done', 'error');
  END IF;
END
$$;

-- Create media_dvds table
CREATE TABLE IF NOT EXISTS media_dvds (
    media_id INTEGER PRIMARY KEY,
    path TEXT NOT NULL CHECK (path <> ''),
    ingestion_state media_dvd_ingestion_state NOT NULL DEFAULT 'pending',
    ingestion_error TEXT,
    CONSTRAINT fk_media_dvds_media_id 
        FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE CASCADE
);

-- A trigger to ensure that the media record is deleted when the corresponding media_dvd is deleted.
CREATE OR REPLACE FUNCTION delete_dvd_media()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM media WHERE id = OLD.media_id;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_delete_dvd_media ON media_dvds;

CREATE TRIGGER trg_delete_dvd_media
AFTER DELETE ON media_dvds
FOR EACH ROW
EXECUTE FUNCTION delete_dvd_media();

-- Create media_x_cards table
CREATE TABLE IF NOT EXISTS media_x_cards (
    media_id INTEGER NOT NULL,
    card_id INTEGER NOT NULL,
    PRIMARY KEY (media_id, card_id),
    CONSTRAINT fk_media_cards_media_id 
        FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE CASCADE,
    CONSTRAINT fk_media_cards_card_id 
        FOREIGN KEY (card_id) REFERENCES catalog_cards(id) ON DELETE CASCADE
);

-- Create media_sets_x_cards table
CREATE TABLE IF NOT EXISTS media_sets_x_cards (
    media_set_id INTEGER NOT NULL,
    card_id INTEGER NOT NULL,
    PRIMARY KEY (media_set_id, card_id),
    CONSTRAINT fk_media_sets_cards_media_set_id 
        FOREIGN KEY (media_set_id) REFERENCES media_sets(id) ON DELETE CASCADE,
    CONSTRAINT fk_media_sets_cards_card_id 
        FOREIGN KEY (card_id) REFERENCES catalog_cards(id) ON DELETE CASCADE
);
