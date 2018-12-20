--@table
river

--@spot-counters
(SELECT to_json(spot_counters) FROM (
    SELECT COALESCE(sum(CASE

        WHEN auto_ordering AND last_auto_ordering=(
                select max(last_auto_ordering) FROM white_water_rapid where river_id=6605
            ) AND last_auto_ordering>to_timestamp(0) AND order_index>0 THEN 1
        WHEN order_index>0 THEN 1
        ELSE 0 END),0) as ordered,

        count(1) as total FROM white_water_rapid WHERE river_id=river.id

) spot_counters) AS spot_counters

--@bounds
SELECT ___table___.id id, ST_AsGeoJSON(ST_Extent(white_water_rapid.point)) bounds from
    ___table___ INNER JOIN white_water_rapid ON  ___table___.id=white_water_rapid.river_id
    GROUP BY ___table___.id

--@find-by-tags
SELECT sq.id, region_id, region.country_id, title, region.title AS region_title, fake AS region_fake, NULL, NULL, '{}' FROM (
		SELECT id,region_id, title, CASE aliases WHEN '[]' THEN NULL ELSE jsonb_array_elements_text(aliases) END AS alias FROM ___table___) sq
		INNER JOIN region ON sq.region_id=region.id
WHERE title ilike ANY($1) OR alias ilike ANY($1)

--@inside-bounds
SELECT river.id, region_id, region.country_id, river.title, region.title AS region_title, fake AS region_fake, ST_AsGeoJSON(ST_Extent(white_water_rapid.point)), river.aliases, river.props
    FROM ___table___
        INNER JOIN white_water_rapid ON river.id=white_water_rapid.river_id
        INNER JOIN region ON river.region_id=region.id
WHERE (river.visible OR $6) AND exists
    (SELECT 1 FROM white_water_rapid WHERE white_water_rapid.river_id=river.id AND point && ST_MakeEnvelope($1,$2,$3,$4))
GROUP BY river.id, region_id, region.country_id, region.title, region.fake ORDER BY popularity DESC LIMIT $5

--@by-id
SELECT river.id,region_id, region.country_id,river.title, region.title AS region_title, fake AS region_fake,NULL,river.aliases AS aliases, description, visible, river.props, ___spot-counters___
 FROM river INNER JOIN region ON river.region_id=region.id WHERE river.id=$1

--@by-region
SELECT river.id, region_id, region.country_id, river.title, region.title AS region_title, fake AS region_fake, NULL, river.aliases, river.props
    FROM river INNER JOIN region ON river.region_id=region.id WHERE region.id=$1
    ORDER BY CASE river.title WHEN '-' THEN NULL ELSE river.title END ASC

--@by-region-full
WITH bounds AS (___bounds___)
SELECT river.id, region_id, region.country_id, river.title, region.title AS region_title, fake AS region_fake, b.bounds, river.aliases, description, visible, river.props, ___spot-counters___
    FROM river INNER JOIN region ON river.region_id=region.id
     INNER JOIN (SELECT * FROM bounds) b ON b.id=___table___.id
    WHERE region.id=$1
    ORDER BY CASE river.title WHEN '-' THEN NULL ELSE river.title END ASC

--@by-country
SELECT river.id as id, region_id, region.country_id, river.title, region.title AS region_title, fake AS region_fake, NULL, river.aliases as aliases, river.props
    FROM river INNER JOIN region ON river.region_id=region.id WHERE region.fake AND region.country_id=$1
    ORDER BY CASE river.title WHEN '-' THEN NULL ELSE river.title END ASC

--@by-country-full
WITH bounds AS (___bounds___)
SELECT river.id as id, region_id, region.country_id, river.title, region.title AS region_title, fake AS region_fake,
    b.bounds, river.aliases as aliases, description, visible, river.props, ___spot-counters___
    FROM river INNER JOIN region ON river.region_id=region.id
     INNER JOIN (SELECT * FROM bounds) b ON b.id=___table___.id
    WHERE region.fake AND region.country_id=$1
    ORDER BY CASE river.title WHEN '-' THEN NULL ELSE river.title END ASC

--@by-first-letters
SELECT id, region_id, 0, title, "", NULL, aliases, props FROM river WHERE title ilike $1||'%' LIMIT $2

--@update
UPDATE river SET region_id=$2, title=$3, aliases=$4, description=$5, props=$6 WHERE id=$1

--@insert
INSERT INTO river(region_id, title, aliases, description,props) VALUES($1,$2,$3,$4,$5) RETURNING id

--@delete
DELETE FROM river WHERE id=$1

--@set-visible
UPDATE river SET visible=$2 WHERE id=$1

--@by-title-part
SELECT river.id, region_id, region.country_id, river.title, region.title AS region_title, fake AS region_fake, ST_AsGeoJSON(ST_Extent(white_water_rapid.point)), river.aliases, river.props
    FROM ___table___
        INNER JOIN white_water_rapid ON river.id=white_water_rapid.river_id
        INNER JOIN region ON river.region_id=region.id
WHERE ___table___.id=ANY(
    SELECT DISTINCT id FROM
    (SELECT id, title, CASE aliases WHEN '[]' THEN NULL ELSE jsonb_array_elements_text(aliases) END AS alias FROM ___table___) sq
    WHERE  title ilike '%'||$1||'%' OR alias ilike '%'||$2||'%'
    )
GROUP BY river.id, region_id, region.country_id, region.title, region.fake
LIMIT $3 OFFSET $4