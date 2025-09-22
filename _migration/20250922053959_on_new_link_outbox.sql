-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE "new_link_outbox"(
    "id" SERIAL PRIMARY KEY,
    "user_id" INTEGER NOT NULL,
    "link_id" INTEGER UNIQUE NOT NULL,
    "is_done" BOOLEAN DEFAULT false);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE "new_link_outbox";