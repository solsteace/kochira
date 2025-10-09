-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

ALTER TABLE "subscriptions"
    ADD CONSTRAINT "subscriptions_user_id_key" UNIQUE("user_id");

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

ALTER TABLE "subscriptions"
    DROP CONSTRAINT "subscriptions_user_id_key";