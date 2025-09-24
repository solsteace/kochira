-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE "link_shortened_outbox"(
    "id" SERIAL PRIMARY KEY,
    "user_id" INTEGER NOT NULL,
    "link_id" INTEGER UNIQUE NOT NULL,
    "is_done" BOOLEAN DEFAULT false);

CREATE TABLE "short_configured_outbox"(
    "id" SERIAL PRIMARY KEY,
    "user_id" INTEGER NOT NULL,
    "link_id" INTEGER NOT NULL,
    "destination" VARCHAR(255) NOT NULL,
    "shortened" VARCHAR(15) NOT NULL, 
    "is_done" BOOLEAN DEFAULT false);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE "link_shortened_outbox";
DROP TABLE "short_configured_outbox";