--@table
camp

--@fields
id, osm_id, title, description, ST_AsGeoJSON(point), num_tent_places

--@list
SELECT ___fields___ FROM ___table___

--@find-witin-bounds
SELECT ___fields___ FROM ___table___ WHERE point && ST_MakeEnvelope($1,$2,$3,$4)

--@find
SELECT ___fields___ FROM ___table___ WHERE id=$1

--@insert
INSERT INTO ___table___(osm_id, title, description, point, num_tent_places) VALUES ($1,$2,$3,ST_GeomFromGeoJSON($4),$5) RETURNING id

--@update
UPDATE ___table___ SET osm_id=$1, title=$3, description=$4, point=ST_GeomFromGeoJSON($5), num_tent_places=$6 WHERE id=$1

--@remove
DELETE FROM ___table___ WHERE id=$1
