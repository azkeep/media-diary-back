-- +goose Up
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS public.titles (
    id BIGINT GENERATED ALWAYS AS IDENTITY
        CONSTRAINT titles_pk
            PRIMARY KEY,
    date_actual   DATE,
    title         VARCHAR(50),
    is_finished   BOOLEAN,
    media_type    VARCHAR(50),
    media_genre   VARCHAR(50),
    media_comment VARCHAR(250),
    is_dropped    BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_media_title_trgm
    ON public.titles USING gin (title public.gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_media_comment_trgm
    ON public.titles USING gin (media_comment public.gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_media_genre_trgm
    ON public.titles USING gin (media_genre public.gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS idx_media_title_trgm;
DROP INDEX IF EXISTS idx_media_comment_trgm;
DROP INDEX IF EXISTS idx_media_type_trgm;
DROP TABLE IF EXISTS public.titles;
