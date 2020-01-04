--@list
SELECT id, l0, l1, l2, l3 FROM level_sensor
--@find
SELECT id, l0, l1, l2, l3 FROM level_sensor WHERE id=$1
--@set-graduation
UPDATE level_sensor SET l0=$2,l1=$3,l2=$4,l3=$5 WHERE id=$1
--@check-and-create-if-missing
INSERT INTO level_sensor(id, l0, l1, l2, l3)
VALUES ($1, 0, 0, 0, 0)
ON CONFLICT DO NOTHING;
