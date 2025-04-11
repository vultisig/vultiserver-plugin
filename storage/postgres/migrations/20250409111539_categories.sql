-- +goose Up
-- +goose StatementBegin
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL
);

ALTER TABLE plugins
    ADD COLUMN category_id UUID NOT NULL,
    ADD CONSTRAINT fk_plugins_category FOREIGN KEY (category_id) REFERENCES categories(id);

INSERT INTO categories (name)
    VALUES ('AI Agent'), ('Plugin');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE plugins
    DROP CONSTRAINT fk_plugins_category,
    DROP COLUMN category_id;

DROP TABLE IF EXISTS categories;
-- +goose StatementEnd
