CREATE SEQUENCE id_gen;

CREATE TABLE track (
  id       BIGINT PRIMARY KEY DEFAULT nextval('id_gen'),
  title    VARCHAR(512) NOT NULL,
  category VARCHAR(4),
  path     GEOMETRY NOT NULL,
  created  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE point (
  id       BIGINT PRIMARY KEY DEFAULT nextval('id_gen'),
  track_id BIGINT REFERENCES track (id) NOT NULL ,
  point    GEOMETRY NOT NULL,
  title     TEXT,
  text     TEXT,
  time  TIMESTAMP WITH TIME ZONE
);
CREATE INDEX ON point (track_id);
