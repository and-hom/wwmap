--@table
level
--@insert
INSERT INTO ___table___(sensor_id, "date", hour_of_day, level) VALUES ($1,$2,$3,$4) RETURNING id
--@list-one
SELECT id, sensor_id, "date", hour_of_day, level FROM
  (SELECT id, sensor_id, "date", hour_of_day, level,
          ROW_NUMBER() OVER (PARTITION BY sensor_id, date ORDER BY hour_of_day DESC) AS rn
  FROM ___table___ WHERE level IS NOT NULL AND date>=$1)sq
 WHERE rn=1 ORDER BY sensor_id, date DESC
--@remove-nulls
DELETE FROM ___table___ WHERE "level" IS NULL AND "date"<$1