-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE "subscription_checked_outbox"(
    "id" SERIAL PRIMARY KEY,
    "context_id" INTEGER NOT NULL,
    "usecase" VARCHAR(31) NOT NULL,
    "lifetime" BIGINT NOT NULL, -- Go's time.Duration type is 64-bit unsigned integer
    "limit" INTEGER NOT NULL,
    "allow_short_edit" BOOLEAN NOT NULL,
    "is_done" BOOLEAN DEFAULT false);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE "subscription_checked_outbox";
