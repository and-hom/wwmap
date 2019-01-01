DELETE FROM "user" WHERE NOT ext_id ~ '^[0-9]+$' OR LENGTH(ext_id)>18;
ALTER TABLE "user" ADD COLUMN old_ext_id CHARACTER VARYING(128);
UPDATE "user" SET old_ext_id=ext_id;

ALTER TABLE "user" DROP COLUMN ext_id;
ALTER TABLE "user" ADD COLUMN ext_id BIGINT;

UPDATE "user" SET ext_id=(old_ext_id::BIGINT);
ALTER TABLE "user" DROP COLUMN old_ext_id;

ALTER TABLE "user" ALTER COLUMN ext_id SET NOT NULL;

CREATE INDEX user_yandex_id ON "user"(auth_provider, ext_id);
ALTER TABLE "user" ADD CONSTRAINT user_yandex_id_key UNIQUE(auth_provider, ext_id);