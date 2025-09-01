-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
CREATE TABLE "users"(
    "id" SERIAL PRIMARY KEY,
    "username" VARCHAR(31) UNIQUE NOT NULL,
    "password" VARCHAR(127) NOT NULL,
    "email" VARCHAR(127) UNIQUE NOT NULL);

CREATE TABLE "links"(
    "id" SERIAL PRIMARY KEY,
    "user_id" INTEGER NOT NULL,
    "shortened" VARCHAR(15) UNIQUE NOT NULL,
    "destination" VARCHAR(255) NOT NULL,
    "is_open" BOOLEAN DEFAULT true,
    "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_DATE,
    "expired_at" TIMESTAMP NOT NULL,

    FOREIGN KEY ("user_id")
        REFERENCES "users"("id")
        ON DELETE CASCADE);

CREATE TABLE "subscriptions"(
    "id" SERIAL PRIMARY KEY,
    "user_id" INTEGER NOT NULL,
    "expired_at" TIMESTAMP NOT NULL,

    FOREIGN KEY ("user_id")
        REFERENCES "users"("id")
        ON DELETE CASCADE);

CREATE TABLE "payments"(
    "id" SERIAL PRIMARY KEY,
    "user_id" INTEGER NOT NULL,
    "invoice_number" VARCHAR(63) UNIQUE NOT NULL,
    "quantity" INTEGER NOT NULL,
    "bill" INTEGER NOT NULL,
    "subscription_type_id" SMALLINT NOT NULL,
    "status" VARCHAR(15) NOT NULL,
    "expired_at" TIMESTAMP NOT NULL,

    FOREIGN KEY("user_id")
        REFERENCES "users"("id")
        ON DELETE SET NULL);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE "payments";
DROP TABLE "subscriptions";
DROP TABLE "links";
DROP TABLE "users";
