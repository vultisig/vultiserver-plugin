-- +goose Up
-- +goose StatementBegin
CREATE TABLE plugin_rating (
    plugin_id UUID NOT NULL REFERENCES plugins(id) ON DELETE CASCADE,
    rating INT CHECK (rating BETWEEN 1 AND 5),
    count INT DEFAULT 0 CHECK (count >= 0),
    PRIMARY KEY (plugin_id, rating) -- Ensures unique (plugin_id, rating) pair
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS plugin_rating;
-- +goose StatementEnd