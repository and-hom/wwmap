--@table
meteo_point
--@insert
INSERT INTO ___table___(title,point) VALUES ($1,ST_GeomFromGeoJSON($2)) RETURNING id
--@list
SELECT id, title, ST_AsGeoJSON(point) FROM ___table___
--@by-id
___list___ WHERE id=$1