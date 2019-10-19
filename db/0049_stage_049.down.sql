DROP TABLE waterway_ref;
DROP TABLE waterway_osm_ref;
DROP INDEX waterway_osm_id_idx;
DROP TRIGGER waterway_path_simplified_change_trigger ON waterway;
DROP TRIGGER waterway_path_simplified_trigger ON waterway;
DROP FUNCTION path_simplified_changed();
DROP FUNCTION set_waterway_path_simplified();
DROP INDEX waterway_path_idx;
ALTER TABLE waterway
    DROP COLUMN path_simplified,
    DROP COLUMN path_geogr;
ALTER TABLE "user" DROP COLUMN experimental_features;

DROP INDEX waterway_title_idx;
