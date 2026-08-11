-- +goose Up
CREATE TABLE order_items (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK(quantity > 0),
    price BIGINT NOT NULL CHECK(price >= 0) -- в копейках
);

-- +goose Down
DROP TABLE IF EXISTS order_items;
