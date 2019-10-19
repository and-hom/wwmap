CREATE INDEX waterway_title_idx ON waterway(title);

ALTER TABLE waterway
    ADD COLUMN path_simplified geometry,
    ADD CONSTRAINT path_simplified_is_linestring CHECK (GeometryType(path_simplified) = 'LINESTRING');

UPDATE waterway
SET path_simplified = ST_Simplify(path, 0.0005, FALSE);
DELETE
from waterway
where ST_IsEmpty(path);

ALTER TABLE waterway
    ALTER COLUMN path_simplified SET NOT NULL;

CREATE OR REPLACE FUNCTION set_waterway_path_simplified()
    RETURNS trigger AS
$BODY$
BEGIN
    NEW.path_simplified = ST_Simplify(NEW.path, 0.0005, FALSE);
    RETURN NEW;
END
$BODY$ LANGUAGE 'plpgsql';

CREATE OR REPLACE FUNCTION path_simplified_changed()
    RETURNS trigger AS
$BODY$
BEGIN
    RAISE EXCEPTION 'path_simplified modified directly';
END
$BODY$ LANGUAGE 'plpgsql';

CREATE TRIGGER waterway_path_simplified_trigger
    BEFORE UPDATE OF "path" OR INSERT
    ON waterway
    FOR EACH ROW
EXECUTE PROCEDURE set_waterway_path_simplified();

CREATE TRIGGER waterway_path_simplified_change_trigger
    BEFORE UPDATE OF "path_simplified"
    ON waterway
    FOR EACH ROW
EXECUTE PROCEDURE path_simplified_changed();

ALTER TABLE "user"
    ADD COLUMN experimental_features BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX waterway_path_idx ON waterway USING gist (path);

ALTER TABLE waterway
    ADD COLUMN path_geogr geography;
UPDATE waterway
SET path_geogr=path_simplified::geography;

CREATE UNIQUE INDEX waterway_osm_id_idx ON waterway (osm_id);
CREATE TABLE waterway_osm_ref
(
    id          BIGINT REFERENCES waterway (osm_id) NOT NULL,
    ref_id      BIGINT REFERENCES waterway (osm_id) NOT NULL,
    cross_point GEOMETRY                            NOT NULL,
    CONSTRAINT no_self_ref CHECK (id <> ref_id)
);
CREATE INDEX waterway_osm_ref_id_idx ON waterway_osm_ref (id);
CREATE INDEX waterway_osm_ref_ref_id_idx ON waterway_osm_ref (ref_id);

CREATE TABLE waterway_ref
(
    id          BIGINT REFERENCES waterway (id) NOT NULL,
    ref_id      BIGINT REFERENCES waterway (id) NOT NULL,
    cross_point GEOMETRY                        NOT NULL,
    CONSTRAINT no_self_ref CHECK (id <> ref_id)
);
CREATE INDEX waterway_ref_id_idx ON waterway_ref (id);
CREATE INDEX waterway_ref_ref_id_idx ON waterway_ref (ref_id);
