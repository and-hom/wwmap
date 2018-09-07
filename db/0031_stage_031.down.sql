CREATE TABLE route (
  id       BIGINT PRIMARY KEY                DEFAULT nextval('id_gen'),
  title    VARCHAR(512)             NOT NULL,
  category VARCHAR(4),
  created  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE point
(
  id       BIGINT PRIMARY KEY                DEFAULT nextval('id_gen'),
  type character varying(32),
  title text,
  point geometry NOT NULL,
  content text,
  "time" timestamp with time zone,
  route_id bigint NOT NULL REFERENCES route (id),
  CONSTRAINT point_is_point CHECK (geometrytype(point) = 'POINT'::text)
);
CREATE INDEX point_route_id_idx ON point USING btree(route_id);

-- Table: track

-- DROP TABLE track;

CREATE TABLE track
(
  id       BIGINT PRIMARY KEY                DEFAULT nextval('id_gen'),
  title character varying(512) NOT NULL,
  category character varying(4),
  path geometry NOT NULL,
  created timestamp with time zone NOT NULL DEFAULT now(),
  type character varying(4),
  route_id bigint REFERENCES route (id),
  length double precision NOT NULL,
  start_time timestamp without time zone NOT NULL DEFAULT now(),
  end_time timestamp without time zone NOT NULL DEFAULT now(),
  CONSTRAINT path_is_linestring CHECK (geometrytype(path) = 'LINESTRING'::text)
);
CREATE INDEX track_route_id_idx  ON track  USING btree(route_id);

