-- Create catalog_cards table
CREATE TABLE IF NOT EXISTS catalog_cards (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL CHECK (name <> ''),
    year INTEGER NOT NULL
);

-- Create catalog_movies table
CREATE TABLE IF NOT EXISTS catalog_movies (
    id SERIAL PRIMARY KEY,
    tmdb_id INTEGER,
    fanart_id TEXT,
    card_id INTEGER NOT NULL,
    CONSTRAINT fk_catalog_movies_card_id 
        FOREIGN KEY (card_id) REFERENCES catalog_cards(id)
);

-- Create catalog_movie_edition_kinds table
CREATE TABLE IF NOT EXISTS catalog_movie_edition_kinds (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- Create catalog_movie_editions table
CREATE TABLE IF NOT EXISTS catalog_movie_editions (
    id SERIAL PRIMARY KEY,
    kind_id INTEGER NOT NULL,
    movie_id INTEGER NOT NULL,
    CONSTRAINT fk_catalog_movie_editions_kind_id 
        FOREIGN KEY (kind_id) REFERENCES catalog_movie_edition_kinds(id),
    CONSTRAINT fk_catalog_movie_editions_movie_id 
        FOREIGN KEY (movie_id) REFERENCES catalog_movies(id)
);