ALTER TABLE "user" ADD COLUMN auth_provider CHARACTER VARYING(16);
UPDATE "user" SET auth_provider='yandex';
ALTER TABLE "user" ALTER COLUMN auth_provider SET NOT NULL;

ALTER TABLE "user" RENAME COLUMN yandex_id TO ext_id;

