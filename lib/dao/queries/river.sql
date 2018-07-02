--@find-by-tags
SELECT id,region_id,title,NULL, NULL FROM (
		SELECT id, title, CASE aliases WHEN '[]' THEN NULL ELSE jsonb_array_elements_text(aliases) END AS alias FROM river) sq
WHERE title ilike ANY($1) OR alias ilike ANY($1)
--@nearest
SELECT id,region_id, title, NULL, aliases FROM (
SELECT ROW_NUMBER() OVER (PARTITION BY id ORDER BY distance ASC) AS r_num, id, title, distance, aliases FROM (
SELECT river.id AS id, river.title AS title, river.aliases AS aliases,
ST_Distance(path,  ST_GeomFromGeoJSON($1)) AS distance FROM river INNER JOIN waterway ON river.id=waterway.river_id) ssq
)sq WHERE r_num<=1 ORDER BY distance ASC LIMIT $2;
--@inside-bounds
SELECT river.id,region_id, river.title, ST_AsGeoJSON(ST_Extent(white_water_rapid.point)), river.aliases FROM
river INNER JOIN white_water_rapid ON river.id=white_water_rapid.river_id
WHERE exists(SELECT 1 FROM white_water_rapid WHERE white_water_rapid.river_id=river.id and point && ST_MakeEnvelope($1,$2,$3,$4))
GROUP BY river.id, river.title ORDER BY popularity DESC LIMIT $5

--@by-id
SELECT id,region_id,title,NULL,river.aliases AS aliases FROM river WHERE id=$1

--@by-region
SELECT river.id, region_id, river.title, NULL, river.aliases
    FROM river INNER JOIN region ON river.region_id=region.id WHERE region.id=$1

--@by-country
SELECT river.id as id, region_id, river.title as title, NULL, river.aliases as aliases
    FROM river INNER JOIN region ON river.region_id=region.id WHERE region.fake AND region.country_id=$1

--@by-first-letters
SELECT id, region_id, title, NULL, aliases FROM river WHERE title ilike $1||'%' LIMIT $2

--@update
UPDATE river SET region_id=$2, title=$3, aliases=$4 WHERE id=$1

--@insert
INSERT INTO river(region_id, title, aliases) VALUES($1,$2,$3) RETURNING id

--@fix-linked-waterways
UPDATE waterway SET river_id=NULL WHERE river_id=$1

--@delete-linked-wwpts
DELETE FROM white_water_rapid WHERE river_id=$1

--@delete-linked-reports
DELETE FROM voyage_report_river WHERE river_id=$1

--@delete
DELETE FROM river WHERE id=$1