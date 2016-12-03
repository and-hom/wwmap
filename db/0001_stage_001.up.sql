CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE route (
  id       BIGSERIAL PRIMARY KEY,
  title    VARCHAR(512) NOT NULL,
  category VARCHAR(4),
  created  TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE track (
  id       BIGSERIAL PRIMARY KEY,
  route_id BIGINT REFERENCES route (id) NOT NULL ,
  path     GEOMETRY NOT NULL
);
CREATE INDEX ON track (route_id);

CREATE TABLE point (
  id       BIGSERIAL PRIMARY KEY,
  track_id BIGINT REFERENCES track (id) NOT NULL ,
  point    GEOMETRY NOT NULL,
  text     TEXT
);
CREATE INDEX ON point (track_id);
