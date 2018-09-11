UPDATE waterway SET "path"=ST_FlipCoordinates("path");
UPDATE white_water_rapid SET "point"=ST_FlipCoordinates("point");