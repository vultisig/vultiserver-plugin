-- +goose Up
-- +goose StatementBegin
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    color VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE plugin_tags (
    plugin_id UUID NOT NULL REFERENCES plugins(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (plugin_id, tag_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS plugin_tags;

DROP TABLE IF EXISTS tags;
-- +goose StatementEnd
