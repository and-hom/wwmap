--@table
meteo
--@insert
INSERT INTO ___table___(point_id, "date", daytime, temp, rain) VALUES ($1,$2,$3,$4,$5) RETURNING id