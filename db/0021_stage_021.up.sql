ALTER TABLE white_water_rapid ADD COLUMN lw_category CHARACTER VARYING(4);
ALTER TABLE white_water_rapid ADD COLUMN lw_description CHARACTER VARYING(4096);
ALTER TABLE white_water_rapid ADD COLUMN mw_category CHARACTER VARYING(4);
ALTER TABLE white_water_rapid ADD COLUMN mw_description CHARACTER VARYING(4096);
ALTER TABLE white_water_rapid ADD COLUMN hw_category CHARACTER VARYING(4);
ALTER TABLE white_water_rapid ADD COLUMN hw_description CHARACTER VARYING(4096);

ALTER TABLE white_water_rapid ADD COLUMN orient CHARACTER VARYING(4096);
ALTER TABLE white_water_rapid ADD COLUMN approach CHARACTER VARYING(4096);
ALTER TABLE white_water_rapid ADD COLUMN safety CHARACTER VARYING(4096);

ALTER TABLE white_water_rapid ADD COLUMN preview CHARACTER VARYING(512);

