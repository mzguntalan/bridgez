-- +goose Up
-- +goose StatementBegin
BEGIN TRANSACTION;
CREATE TABLE words (
    word_id serial PRIMARY KEY,
    representation TEXT,
    UNIQUE(representation)
);

CREATE TABLE links (
    link_id serial PRIMARY KEY,
    source_word_id INTEGER REFERENCES words(word_id),
    target_word_id INTEGER REFERENCES words(word_id),
    UNIQUE(source_word_id, target_word_id)
);

COMMIT;

-- +goose StatementEnd

-- +goose Down
BEGIN TRANSACTION;
DROP TABLE links;
DROP TABLE words;
COMMIT;
-- +goose StatementBegin
-- +goose StatementEnd
