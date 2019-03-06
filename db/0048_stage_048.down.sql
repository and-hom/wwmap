ALTER TABLE white_water_rapid DROP CONSTRAINT point_is_point_or_linestring;
UPDATE white_water_rapid SET point = (SELECT ST_DumpPoints(point) LIMIT 1).geom WHERE GeometryType(point) = 'LINESTRING';
ALTER TABLE white_water_rapid ADD CONSTRAINT point_is_point CHECK (GeometryType(point) = 'POINT');
