DROP TABLE white_water_rapid;
DROP TABLE waterway;

ALTER TABLE track DROP CONSTRAINT path_is_linestring;
ALTER TABLE point DROP CONSTRAINT point_is_point;

