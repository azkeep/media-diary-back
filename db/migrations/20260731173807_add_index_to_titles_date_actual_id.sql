-- +goose Up
CREATE INDEX IF NOT EXISTS idx_titles_date_actual_id ON titles (date_actual DESC, id DESC );

-- +goose Down
DROP INDEX IF EXISTS idx_titles_date_actual_id
