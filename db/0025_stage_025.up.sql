UPDATE white_water_rapid SET
    lw_category = COALESCE(lw_category, ''),
    lw_description = COALESCE(lw_description, ''),
    mw_category = COALESCE(mw_category, ''),
    mw_description = COALESCE(mw_description, ''),
    hw_category = COALESCE(hw_category, ''),
    hw_description = COALESCE(hw_description, ''),
    orient = COALESCE(orient, ''),
    approach = COALESCE(approach, ''),
    safety = COALESCE(safety, ''),
    preview = COALESCE(preview, '');

ALTER TABLE white_water_rapid ALTER COLUMN lw_category SET NOT NULL;
ALTER TABLE white_water_rapid ALTER COLUMN lw_description SET NOT NULL;
ALTER TABLE white_water_rapid ALTER COLUMN mw_category SET NOT NULL;
ALTER TABLE white_water_rapid ALTER COLUMN mw_description SET NOT NULL;
ALTER TABLE white_water_rapid ALTER COLUMN hw_category SET NOT NULL;
ALTER TABLE white_water_rapid ALTER COLUMN hw_description SET NOT NULL;
ALTER TABLE white_water_rapid ALTER COLUMN orient SET NOT NULL;
ALTER TABLE white_water_rapid ALTER COLUMN approach SET NOT NULL;
ALTER TABLE white_water_rapid ALTER COLUMN safety SET NOT NULL;
ALTER TABLE white_water_rapid ALTER COLUMN preview SET NOT NULL;