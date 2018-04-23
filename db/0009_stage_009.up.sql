CREATE TABLE river (
  id            BIGINT PRIMARY KEY                DEFAULT nextval('id_gen'),
  title         CHARACTER VARYING(512) NOT NULL,
  popularity    SMALLINT NOT NULL DEFAULT 0
);

ALTER TABLE waterway DROP COLUMN parent_id;
ALTER TABLE waterway ADD COLUMN river_id BIGINT REFERENCES river(id);
CREATE INDEX waterway_river_id
  ON waterway (river_id) WHERE river_id IS NOT NULL;

INSERT INTO river(title,popularity)
    (SELECT title, max(popularity) FROM waterway WHERE verified GROUP BY title);

UPDATE waterway SET
    river_id=(SELECT id FROM river WHERE river.title=waterway.title)
    WHERE verified;

ALTER TABLE waterway DROP COLUMN verified;
ALTER TABLE waterway DROP COLUMN popularity;

ALTER TABLE white_water_rapid ADD COLUMN river_id BIGINT REFERENCES river(id);
CREATE INDEX white_water_rapid_river_id
  ON white_water_rapid (river_id) WHERE river_id IS NOT NULL;
UPDATE white_water_rapid SET river_id=(SELECT river_id FROM waterway WHERE waterway.id=white_water_rapid.water_way_id);
ALTER TABLE white_water_rapid DROP COLUMN water_way_id;
