-- +goose Up
ALTER TABLE orders
    ADD COLUMN status TEXT NOT NULL DEFAULT 'new',
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- +goose Down
ALTER TABLE orders
    DROP COLUMN status,
    DROP COLUMN created_at;
