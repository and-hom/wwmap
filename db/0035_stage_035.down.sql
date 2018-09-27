ALTER TABLE "user" DROP COLUMN auth_provider;
ALTER TABLE "user" RENAME COLUMN ext_id TO yandex_id;