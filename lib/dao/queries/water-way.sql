--@insert
INSERT INTO waterway(osm_id, title, type, comment, path) VALUES ($1, $2, $3, $4, ST_GeomFromGeoJSON($5))
--@update
UPDATE waterway SET path=ST_GeomFromGeoJSON($1) WHERE osm_id=$2
--@list
SELECT id, osm_id, river_Id, title, type, comment, ST_AsGeoJSON(path) FROM waterway
--@unlink-river
UPDATE waterway SET river_id=NULL WHERE river_id=$1