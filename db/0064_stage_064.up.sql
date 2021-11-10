ALTER TABLE waterway
    ADD COLUMN heights INTEGER[],
    ADD COLUMN dists DOUBLE PRECISION[];

CREATE OR REPLACE FUNCTION set_waterway_path_simplified()
    RETURNS trigger AS
$BODY$
BEGIN
    NEW.path_simplified = ST_Simplify(NEW.path, 0.0005, FALSE);
    NEW.heights = NULL;
    NEW.dists = NULL;
    RETURN NEW;
END
$BODY$ LANGUAGE 'plpgsql';

CREATE TABLE srtm
(
    lat      INTEGER,
    lon      INTEGER,
    rast     RASTER,
    filename TEXT,
    updated TIMESTAMP WITH TIME ZONE DEFAULT now(),
    PRIMARY KEY (lat, lon)
);

CREATE OR REPLACE FUNCTION srtm_insert()
    RETURNS trigger AS
$BODY$
BEGIN
    NEW.lat = (regexp_matches(NEW.filename, '^[NS](\d+)[WE](\d+)\..*'))[1]::int;
    NEW.lon = (regexp_matches(NEW.filename, '^[NS](\d+)[WE](\d+)\..*'))[2]::int;
    RETURN NEW;
END
$BODY$ LANGUAGE 'plpgsql';

CREATE OR REPLACE FUNCTION srtm_update()
    RETURNS trigger AS
$BODY$
BEGIN
    RAISE EXCEPTION 'Cant update srtm - should remove and insert';
END
$BODY$ LANGUAGE 'plpgsql';

CREATE TRIGGER srtm_insert_trigger
    BEFORE INSERT
    ON srtm
    FOR EACH ROW
EXECUTE PROCEDURE srtm_insert();

CREATE TRIGGER srtm_update_trigger
    BEFORE UPDATE
    ON srtm
    FOR EACH ROW
EXECUTE PROCEDURE srtm_update();

CREATE INDEX srtm_lat_lon_idx ON srtm(lat, lon);
