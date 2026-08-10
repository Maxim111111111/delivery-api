-- +goose Up
CREATE TABLE orders (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    address TEXT NOT NULL,
    price BIGINT NOT NULL --в копейках
);

-- +goose Down
DROP TABLE IF EXISTS orders;
