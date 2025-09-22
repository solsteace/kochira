-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE "register_outbox"(
    "id" SERIAL PRIMARY KEY,
    "user_id" INTEGER UNIQUE NOT NULL,
    "is_done" BOOLEAN DEFAULT false);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE "register_outbox";