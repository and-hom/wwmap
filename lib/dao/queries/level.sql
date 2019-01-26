--@table
level
--@insert
INSERT INTO ___table___(sensor_id, "date", hour_of_day, level) VALUES ($1,$2,$3,$4) RETURNING id