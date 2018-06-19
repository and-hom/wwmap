CREATE TABLE "user" (
  id                BIGINT PRIMARY KEY                DEFAULT nextval('id_gen'),
  yandex_id         BIGINT NOT NULL UNIQUE,
  role              character varying(16) NOT NULL,
  info              JSONB
);
CREATE INDEX user_yandex_id ON "user"(yandex_id);