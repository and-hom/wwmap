CREATE TABLE waterway (
  id        BIGINT PRIMARY KEY                DEFAULT nextval('id_gen'),
  title     CHARACTER VARYING(512) NOT NULL,
  type      CHARACTER VARYING(64),
  path      GEOMETRY               NOT NULL,
  parent_id BIGINT REFERENCES waterway (id),
  comment   TEXT,
  CONSTRAINT path_is_linestring CHECK (GeometryType(path) = 'LINESTRING')
);
CREATE INDEX river_parent
  ON waterway (parent_id);

CREATE TABLE white_water_rapid (
  id            BIGINT PRIMARY KEY DEFAULT nextval('id_gen'),
  osm_id        BIGINT,
  water_way_id BIGINT REFERENCES waterway (id),
  title         VARCHAR(512),
  type          VARCHAR(64),
  category      VARCHAR(4),
  point         GEOMETRY NOT NULL,
  comment       TEXT,
  CONSTRAINT point_is_point CHECK (GeometryType(point) = 'POINT')
);
CREATE INDEX white_water_rapid_river
  ON white_water_rapid (whater_way_id);
CREATE INDEX white_water_rapid_osm_id
  ON white_water_rapid (osm_id);

ALTER TABLE track
  ADD CONSTRAINT path_is_linestring CHECK (GeometryType(path) = 'LINESTRING');
ALTER TABLE point
  ADD CONSTRAINT point_is_point CHECK (GeometryType(point) = 'POINT');

