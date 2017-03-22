ALTER TABLE waterway
  ADD COLUMN epsilon_area GEOMETRY;
-- create epsilon-area around river
UPDATE waterway
        -- set epsilon 50 meters. all data stored as geometry, but we need 50 meters, not 50 degrees
        -- so cast to geography, perform operation, then to geography back
SET epsilon_area = (ST_Buffer(path::geography, 50))::geometry;
