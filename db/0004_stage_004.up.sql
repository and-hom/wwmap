ALTER TABLE track
  ADD COLUMN "length" FLOAT;
UPDATE track
SET "length" = ST_Length("path" :: GEOGRAPHY);
ALTER TABLE track
  ALTER COLUMN "length" SET NOT NULL;
