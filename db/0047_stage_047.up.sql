ALTER TABLE level ADD COLUMN hour_of_day SMALLINT;
UPDATE level SET hour_of_day=0;
ALTER TABLE level ALTER COLUMN hour_of_day SET NOT NULL;