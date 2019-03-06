ALTER TABLE white_water_rapid DROP CONSTRAINT point_is_point;
ALTER TABLE white_water_rapid ADD CONSTRAINT point_is_point_or_linestring CHECK (
  GeometryType(point) = 'POINT' OR GeometryType(point) = 'LINESTRING');
