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
    alias_query AS (SELECT id, jsonb_array_elements_text(aliases) AS alias FROM river),
    river_rects_query AS (SELECT river_id, ST_Envelope(ST_Extent(point)) as bounds FROM white_water_rapid GROUP BY river_id),
    rivers_query AS (SELECT river.id, title, alias_query.alias AS alias, bounds
        FROM river
        LEFT OUTER JOIN alias_query ON river.id = alias_query.id
        INNER JOIN river_rects_query ON river.id = river_rects_query.river_id)
UPDATE ___table___ SET river_id=rivers_query.id
    FROM rivers_query
    WHERE  (lower(___table___.title)=lower(rivers_query.title) OR lower(___table___.title)=lower(rivers_query.alias))
      AND ST_Distance(___table___.path::geography, rivers_query.bounds::geography) < $1

--@list-by-river-ids
SELECT ___select-fields___ FROM ___table___ WHERE river_id=ANY($1)

--@list-with-river-without-heights
SELECT id, ST_AsGeoJSON(path_simplified),'[]'::json
FROM ___table___ WHERE river_id IS NOT NULL AND heights IS NULL

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

--@list-by-bbox-with-heights
SELECT id, ST_AsGeoJSON(path_simplified), heights, ST_Length("path_simplified" :: GEOGRAPHY)
FROM ___table___ WHERE river_id IS NOT NULL AND heights IS NOT NULL AND path && ST_MakeEnvelope($1,$2,$3,$4)

--@list-by-bbox
SELECT ___select-fields___ FROM ___table___ WHERE path && ST_MakeEnvelope($1,$2,$3,$4)

--@list-4-correction
SELECT id, ST_AsGeoJSON(path), ST_AsGeoJSON(ST_Simplify(path, 0.0005, FALSE)) FROM waterway LIMIT $1 OFFSET $2

--@get-ref-points
SELECT id, ST_AsGeoJSON(cross_point) AS point
FROM (
         SELECT DISTINCT id, cross_point FROM waterway_ref WHERE id = ANY ($1) OR ref_id = ANY ($1)) sq

--@update-path-simplified
UPDATE ___table___ SET path_simplified=ST_GeomFromGeoJSON($2) WHERE id=$1

--@update-path-height-and-dists
UPDATE ___table___ SET heights=$2, dists=$3 WHERE id=$1