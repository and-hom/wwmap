--@table
river
--@find-by-tags
SELECT id,region_id,title,NULL, NULL, '{}' FROM (
		SELECT id,region_id, title, CASE aliases WHEN '[]' THEN NULL ELSE jsonb_array_elements_text(aliases) END AS alias FROM river) sq
WHERE title ilike ANY($1) OR alias ilike ANY($1)
--@nearest
SELECT id,region_id, title, NULL, aliases, river.props FROM (
SELECT ROW_NUMBER() OVER (PARTITION BY id ORDER BY distance ASC) AS r_num, id, title, distance, aliases FROM (
SELECT river.id AS id, river.title AS title, river.aliases AS aliases,
ST_Distance(path,  ST_GeomFromGeoJSON($1)) AS distance FROM river INNER JOIN waterway ON river.id=waterway.river_id) ssq
)sq WHERE r_num<=1 ORDER BY distance ASC LIMIT $2;
--@inside-bounds
SELECT river.id,region_id, river.title, ST_AsGeoJSON(ST_Extent(white_water_rapid.point)), river.aliases, river.props FROM
river INNER JOIN white_water_rapid ON river.id=white_water_rapid.river_id
WHERE exists(SELECT 1 FROM white_water_rapid
    WHERE (river.visible OR $6)
        AND white_water_rapid.river_id=river.id and point && ST_MakeEnvelope($1,$2,$3,$4))
GROUP BY river.id, river.title ORDER BY popularity DESC LIMIT $5

--@by-id
SELECT id,region_id,title,NULL,river.aliases AS aliases, description, visible, props FROM river WHERE id=$1

--@by-region
SELECT river.id, region_id, river.title, NULL, river.aliases, river.props
    FROM river INNER JOIN region ON river.region_id=region.id WHERE region.id=$1
    ORDER BY CASE river.title WHEN '-' THEN NULL ELSE river.title END ASC

--@by-region-full
SELECT river.id, region_id, river.title, NULL, river.aliases, description, visible, river.props
    FROM river INNER JOIN region ON river.region_id=region.id WHERE region.id=$1
    ORDER BY CASE river.title WHEN '-' THEN NULL ELSE river.title END ASC

--@by-country
SELECT river.id as id, region_id, river.title as title, NULL, river.aliases as aliases, river.props
    FROM river INNER JOIN region ON river.region_id=region.id WHERE region.fake AND region.country_id=$1
    ORDER BY CASE river.title WHEN '-' THEN NULL ELSE river.title END ASC

--@by-country-full
SELECT river.id as id, region_id, river.title as title, NULL, river.aliases as aliases, description, visible, river.props
    FROM river INNER JOIN region ON river.region_id=region.id WHERE region.fake AND region.country_id=$1
    ORDER BY CASE river.title WHEN '-' THEN NULL ELSE river.title END ASC

--@by-first-letters
SELECT id, region_id, title, NULL, aliases, props FROM river WHERE title ilike $1||'%' LIMIT $2

--@update
UPDATE river SET region_id=$2, title=$3, aliases=$4, description=$5 WHERE id=$1

--@insert
INSERT INTO river(region_id, title, aliases, description) VALUES($1,$2,$3,$4) RETURNING id

--@delete
DELETE FROM river WHERE id=$1

--@set-visible
UPDATE river SET visible=$2 WHERE id=$1