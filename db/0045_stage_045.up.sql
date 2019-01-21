CREATE TABLE meteo_point
(
  id           BIGINT PRIMARY KEY              DEFAULT nextval('id_gen'),
  title        CHARACTER VARYING(128) NOT NULL,
  point        GEOMETRY               NOT NULL,
  collect_data BOOL                   NOT NULL DEFAULT FALSE
);

CREATE TABLE meteo
(
  id       BIGINT PRIMARY KEY DEFAULT nextval('id_gen'),
  point_id BIGINT       NOT NULL REFERENCES meteo_point (id),
  date     DATE         NOT NULL,
  daytime  CHARACTER(1) NOT NULL,
  temp     INTEGER      NOT NULL,
  rain     INTEGER      NOT NULL
);

CREATE INDEX meteo_point_id_fk ON meteo(point_id);