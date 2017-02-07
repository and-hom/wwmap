CREATE TABLE route (
  id       BIGINT PRIMARY KEY                DEFAULT nextval('id_gen'),
  title    VARCHAR(512)             NOT NULL,
  category VARCHAR(4),
  created  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);
-- generate routes by tracks
INSERT INTO route (title, category, created) (SELECT
                                                title,
                                                category,
                                                created
                                              FROM track);

ALTER TABLE track
  ADD COLUMN route_id BIGINT REFERENCES route (id);

-- add foreign key to tracks
UPDATE track
SET route_id = (SELECT id
                FROM route
                WHERE route.title = track.title AND
                      route.created = track.created AND
                      (route.category = track.category OR route.category IS NULL)
                LIMIT 1);
CREATE INDEX ON track (route_id);

-- bind event points to route, not tracks
ALTER TABLE point
  ADD COLUMN route_id BIGINT REFERENCES route (id);
UPDATE point
SET route_id = (SELECT route_id
                FROM track
                WHERE track.id = point.track_id);
ALTER TABLE point
  ALTER COLUMN route_id SET NOT NULL;
ALTER TABLE point
  DROP COLUMN track_id;
CREATE INDEX ON point (route_id);
