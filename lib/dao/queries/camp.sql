--@table
camp

--@fields
id, title, description, ST_AsGeoJSON(point), num_tent_places

--@list
SELECT ___fields___ FROM ___table___

--@find
SELECT ___fields___ FROM ___table___ WHERE id=$1

--@insert
INSERT INTO ___table___(title, description, point, num_tent_places) VALUES ($1,$2,ST_GeomFromGeoJSON($3),$4) RETURNING id

--@update
UPDATE ___table___ SET title=$2, description=$3, point=ST_GeomFromGeoJSON($4), num_tent_places=$5 WHERE id=$1

--@remove
DELETE FROM ___table___ WHERE id=$1
