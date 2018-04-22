ALTER table waterway ADD COLUMN verified BOOLEAN NOT NULL DEFAULT FALSE;
ALTER table waterway ADD COLUMN popularity SMALLINT NOT NULL DEFAULT 0;
ALTER table waterway ADD COLUMN osm_id BIGINT;

CREATE INDEX waterway_verified
  ON waterway (verified);

