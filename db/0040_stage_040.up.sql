-- preserve old val
ALTER TABLE "user" ADD COLUMN old_ext_id BIGINT;
UPDATE "user" SET old_ext_id=ext_id;
-- create column of new type
ALTER TABLE "user" DROP COLUMN ext_id;
ALTER TABLE "user" ADD COLUMN ext_id CHARACTER VARYING(128);
-- restore old val
UPDATE "user" SET ext_id=(old_ext_id::CHARACTER VARYING);
ALTER TABLE "user" DROP COLUMN old_ext_id;
-- set not null
ALTER TABLE "user" ALTER COLUMN ext_id SET NOT NULL;
-- create indexes and constraints
CREATE INDEX user_yandex_id ON "user"(auth_provider, ext_id);
ALTER TABLE "user" ADD CONSTRAINT user_yandex_id_key UNIQUE(auth_provider, ext_id);