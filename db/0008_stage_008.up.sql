CREATE TABLE report(
  id        BIGINT PRIMARY KEY                DEFAULT nextval('id_gen'),
  object_id BIGINT,
  comment   CHARACTER VARYING(4096) NOT NULL
);