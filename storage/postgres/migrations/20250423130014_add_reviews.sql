-- +goose Up
-- +goose StatementBegin
CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    address TEXT NOT NULL,
    rating INT CHECK (rating BETWEEN 1 AND 5),
    comment TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    plugin_id UUID NOT NULL,
    CONSTRAINT fk_plugin FOREIGN KEY (plugin_id) REFERENCES plugins(id) ON DELETE CASCADE

);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reviews;
-- +goose StatementEnd