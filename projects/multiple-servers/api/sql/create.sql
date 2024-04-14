CREATE TABLE IF NOT EXISTS images
(
    id serial PRIMARY KEY,
    title text NOT NULL,
    url text NOT NULL,
    alt_text text
);