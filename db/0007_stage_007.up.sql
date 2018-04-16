CREATE TABLE waterway_tmp (
  id        BIGINT PRIMARY KEY                DEFAULT nextval('id_gen'),
  title     CHARACTER VARYING(512) NOT NULL,
  type      CHARACTER VARYING(64),
  parent_id BIGINT REFERENCES waterway_tmp (id),
  comment   TEXT
);
CREATE INDEX river_parent_tmp
  ON waterway_tmp (parent_id);


CREATE TABLE point_ref_tmp (
  id            BIGINT PRIMARY KEY DEFAULT nextval('id_gen'),
  parent_id     BIGINT REFERENCES waterway_tmp (id),
  idx           INTEGER
);
CREATE INDEX point_ref_parent_tmp
  ON point_ref_tmp (parent_id);

