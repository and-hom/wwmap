--@table
meteo_point
--@insert
INSERT INTO ___table___(title,point,collect_data) VALUES ($1,ST_GeomFromGeoJSON($2),$3) RETURNING id
--@list
SELECT id, title, ST_AsGeoJSON(point),collect_data FROM ___table___
--@by-id
___list___ WHERE id=$1