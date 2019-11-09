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

--@bind-to-river
WITH
	wwpts AS (SELECT point::geography as point FROM white_water_rapid WHERE river_id=$1),
	ways AS (SELECT id,path::geography as path FROM waterway WHERE lower(title) = ANY($2)),
	way_ids AS (SELECT distinct id FROM wwpts INNER JOIN ways ON ST_Distance(path,point)<$3)
UPDATE ___table___ SET river_id=$1 WHERE id IN (SELECT id FROM way_ids) RETURNING id

--@list-by-river-ids
SELECT ___select-fields___ FROM ___table___ WHERE river_id=ANY($1)

--@list-by-river-id-4-router
SELECT id, ST_AsGeoJSON(path_simplified), '[]' FROM ___table___ WHERE river_id=$1

--@list-by-bbox-4-router
WITH
    rivers_in_area AS (select * FROM waterway WHERE path && ST_MakeEnvelope($1,$2,$3,$4)),
    river_refs AS (SELECT waterway.id, json_agg(
            json_build_object('id',ref.ref_id,'cross_point',ST_AsGeoJSON(ref.cross_point)::json)) as refs
                   FROM rivers_in_area waterway
                            INNER JOIN waterway_ref ref ON waterway.id=ref.id
                   GROUP BY waterway.id)
SELECT rivers_in_area.id, ST_AsGeoJSON(path_simplified), COALESCE(refs,'[]')
    FROM rivers_in_area
    LEFT OUTER JOIN river_refs
    ON river_refs.id=rivers_in_area.id

--@list-by-bbox
SELECT ___select-fields___ FROM ___table___ WHERE path && ST_MakeEnvelope($1,$2,$3,$4)