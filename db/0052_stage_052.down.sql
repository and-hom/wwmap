ALTER TABLE level DROP CONSTRAINT level_sensor_id_fk;
DROP TABLE level_sensor;

ALTER TABLE image DROP COLUMN date_level_updated;
ALTER TABLE image DROP COLUMN date;
ALTER TABLE image DROP COLUMN level;
ALTER TABLE image ALTER COLUMN date_published TYPE timestamp;

UPDATE river
SET props = props - 'vodinfo_sensors' || jsonb_build_object('vodinfo_sensor', props -> 'vodinfo_sensors' -> 0)
WHERE props ? 'vodinfo_sensors';

--##################################
DROP TABLE camp_river_ref;
DROP TABLE camp_rate;
DROP TABLE rate;
DROP TABLE camp_photo;
DROP TABLE photo;
DROP TABLE camp;