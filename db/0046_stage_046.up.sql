CREATE TABLE level
(
  id        BIGINT PRIMARY KEY DEFAULT nextval('id_gen'),
  sensor_id CHARACTER VARYING(16) NOT NULL,
  date      DATE                  NOT NULL,
  level     INTEGER NULL
);

CREATE INDEX level_sensor_id_idx ON level(sensor_id);