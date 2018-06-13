UPDATE white_water_rapid SET short_description=SUBSTRING(short_description, 1, 512);
ALTER TABLE white_water_rapid ALTER COLUMN short_description TYPE CHARACTER VARYING(512);