-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

ALTER TABLE "subscriptions"
    ADD COLUMN "checked_at" TIMESTAMP DEFAULT NULL;

CREATE TABLE "subscription_expired_outbox"(
    "id" SERIAL PRIMARY KEY,
    "user_id" INTEGER NOT NULL,
    "is_done" BOOLEAN DEFAULT false);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

ALTER TABLE "subscriptions" DROP COLUMN "checked_at";

DROP TABLE "subscription_expired_outbox";