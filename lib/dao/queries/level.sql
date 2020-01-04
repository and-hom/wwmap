--@insert
INSERT INTO level(sensor_id, "date", hour_of_day, level) VALUES ($1,$2,$3,$4) RETURNING id
--@list-one
SELECT id, sensor_id, "date", hour_of_day, level FROM
  (SELECT id, sensor_id, "date", hour_of_day, level,
          ROW_NUMBER() OVER (PARTITION BY sensor_id, date ORDER BY hour_of_day DESC) AS rn
  FROM level WHERE level IS NOT NULL AND date>=$1 AND date<=$2)sq
 WHERE rn=1 ORDER BY sensor_id, date DESC
--@remove-nulls
DELETE FROM level WHERE "level" IS NULL AND "date"<$1

--@latest-not-null-for-date
SELECT id, sensor_id, "date", hour_of_day, level FROM level WHERE id = (
    SELECT max(id) FROM level WHERE sensor_id=$1 AND date>=$2 AND date<=$3)

--@list-by-sensor
SELECT id, sensor_id, "date", hour_of_day, level FROM
    (SELECT id, sensor_id, "date", hour_of_day, level,
            ROW_NUMBER() OVER (PARTITION BY sensor_id, date ORDER BY hour_of_day DESC) AS rn
     FROM level WHERE level IS NOT NULL AND sensor_id=$1)sq
WHERE rn=1 ORDER BY sensor_id, date DESC