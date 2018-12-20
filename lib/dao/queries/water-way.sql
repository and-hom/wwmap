--@table
waterway
--@insert
INSERT INTO ___table___(osm_id, title, type, comment, path) VALUES ($1, $2, $3, $4, ST_GeomFromGeoJSON($5))
--@update
UPDATE ___table___ SET path=ST_GeomFromGeoJSON($1) WHERE osm_id=$2
--@select-fields
id, osm_id, ___table___.river_id, title, type, comment, ST_AsGeoJSON(path)
--@list
SELECT ___select-fields___ FROM ___table___
--@unlink-river
UPDATE ___table___ SET river_id=NULL WHERE river_id=$1

--@detect-for-river
WITH
	wwpts AS (SELECT point::geography as point FROM white_water_rapid WHERE river_id=$1 AND auto_ordering),
	ways AS (SELECT id,title,path::geography as path FROM waterway),
	way_ids AS (SELECT id FROM wwpts INNER JOIN ways ON ST_Distance(path,point)<$2 GROUP BY id HAVING count(*)>=least(2, (SELECT count(1)FROM wwpts)))
SELECT ___select-fields___ FROM ___table___ WHERE id=ANY(SELECT id FROM way_ids)
