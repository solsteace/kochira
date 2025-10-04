-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- TODO: convert to Go file to make column backup creating/loading easier

ALTER TABLE "short_configured_outbox" 
    ADD COLUMN "is_open" BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE "short_configured_outbox"
    RENAME COLUMN "shortened" TO "alias";

ALTER TABLE "links"
    ADD COLUMN "alias" VARCHAR(32) NOT NULL DEFAULT '';
UPDATE "links"
    SET "alias" = "shortened";
ALTER TABLE "links" 
    ADD CONSTRAINT "links_alias_key" UNIQUE("alias"); -- following how postgres gives default contraint names

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

-- DROP TABLE "link_notifications";
ALTER TABLE "short_configured_outbox" DROP COLUMN "is_open";
ALTER TABLE "short_configured_outbox" 
    RENAME COLUMN "alias" TO "shortened";

ALTER TABLE "links" DROP COLUMN "alias";