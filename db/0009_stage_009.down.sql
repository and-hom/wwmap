ALTER table waterway ADD COLUMN verified BOOLEAN NOT NULL DEFAULT FALSE;
ALTER table waterway ADD COLUMN popularity SMALLINT NOT NULL DEFAULT 0;
ALTER TABLE waterway ADD COLUMN parent_id BIGINT REFERENCES waterway(id);
CREATE INDEX river_parent
  ON waterway (parent_id);
ALTER TABLE white_water_rapid ADD COLUMN water_way_id BIGINT REFERENCES waterway(id);
CREATE INDEX white_water_rapid_river
  ON white_water_rapid (water_way_id);

UPDATE waterway SET verified=TRUE WHERE river_id IS NOT NULL;
UPDATE waterway SET popularity=COALESCE((SELECT popularity FROM river WHERE river.id=waterway.river_id), 0);

UPDATE white_water_rapid SET water_way_id=(SELECT id FROM waterway WHERE waterway.river_id=white_water_rapid.river_id LIMIT 1);

ALTER TABLE waterway DROP COLUMN river_id;
ALTER TABLE white_water_rapid DROP COLUMN river_id;

DROP TABLE RIVER;