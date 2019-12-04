ALTER TABLE image
    ADD COLUMN date_level_updated TIMESTAMP WITHOUT TIME ZONE,
    ADD COLUMN level JSONB NOT NULL DEFAULT '{}'::jsonb;
COMMENT ON COLUMN image.date_level_updated IS
    'Date of computed water level. Normally should be equals to date_published. On re-calculate water level values should set date_level_updated=date_published';