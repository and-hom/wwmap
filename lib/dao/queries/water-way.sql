--@table
waterway
--@insert
INSERT INTO @@table@@(osm_id, title, type, comment, path) VALUES ($1, $2, $3, $4, ST_GeomFromGeoJSON($5))
--@update
UPDATE @@table@@ SET path=ST_GeomFromGeoJSON($1) WHERE osm_id=$2
--@select-fields
id, osm_id, @@table@@.river_id, title, type, comment, ST_AsGeoJSON(path)
--@list
SELECT @@select-fields@@ FROM @@table@@
--@unlink-river
UPDATE @@table@@ SET river_id=NULL WHERE river_id=$1

--@detect-for-river
WITH
	wwpts AS (SELECT point::geography as point FROM white_water_rapid WHERE river_id=$1 AND auto_ordering),
	ways AS (SELECT id,title,path::geography as path FROM waterway),
	way_ids AS (SELECT id FROM wwpts INNER JOIN ways ON ST_Distance(path,point)<$2 GROUP BY id HAVING count(*)>=least(2, (SELECT count(1)FROM wwpts)))
SELECT @@select-fields@@ FROM @@table@@ WHERE id=ANY(SELECT id FROM way_ids)
