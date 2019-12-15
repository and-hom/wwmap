ALTER TABLE image DROP COLUMN date_level_updated;
ALTER TABLE image DROP COLUMN date;
ALTER TABLE image DROP COLUMN level;

UPDATE river
SET props = props - 'vodinfo_sensors' || jsonb_build_object('vodinfo_sensor', props -> 'vodinfo_sensors' -> 0)
WHERE props ? 'vodinfo_sensors';

