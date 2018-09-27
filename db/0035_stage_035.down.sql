ALTER TABLE "user" DROP COLUMN auth_provider;
ALTER TABLE "user" RENAME COLUMN ext_id TO yandex_id;

DROP INDEX user_yandex_id;
CREATE INDEX user_yandex_id ON "user"(yandex_id);