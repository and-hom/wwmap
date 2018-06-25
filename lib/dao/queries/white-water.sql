--@by-box
SELECT 
    white_water_rapid.id AS id,
    white_water_rapid.title AS title,
    ST_AsGeoJSON(point) as point,
    category,
    short_description,
    link,
    river_id,
    river.title as river_title
FROM white_water_rapid  LEFT OUTER JOIN river ON white_water_rapid.river_id=river.id 
WHERE point && ST_MakeEnvelope($1,$2,$3,$4)

--@by-river
SELECT 
    white_water_rapid.id AS id,
    white_water_rapid.title AS title,
    ST_AsGeoJSON(point) as point,
    category,
    short_description,
    link,
    river_id,
    river.title as river_title
FROM white_water_rapid  LEFT OUTER JOIN river ON white_water_rapid.river_id=river.id 
WHERE river_id=$1

--@by-river-and-title
SELECT 
    white_water_rapid.id AS id,
    white_water_rapid.title AS title,
    ST_AsGeoJSON(point) as point,
    category,
    short_description,
    link,
    river_id,
    river.title as river_title
FROM white_water_rapid  LEFT OUTER JOIN river ON white_water_rapid.river_id=river.id 
WHERE river_id=$1 AND title=$2

--@with-path
SELECT 
    white_water_rapid.id AS id,
    white_water_rapid.title AS title,
    ST_AsGeoJSON(point) as point,
    category,
    short_description,
    link,
    river_id,
    river.title as river_title,
    CASE WHEN region.fake THEN NULL ELSE region.title END AS region_title, country.title as country_title
FROM white_water_rapid 
INNER JOIN river ON white_water_rapid.river_id=river.id
INNER JOIN region ON river.region_id=region.id
INNER JOIN country ON region.country_id=country.id

--@insert
INSERT INTO white_water_rapid(title,category,point,short_description, link, river_id)
		VALUES ($2, $3, ST_GeomFromGeoJSON($4), $5, $6, $7)

--@update
UPDATE white_water_rapid SET title=$2,category=$3, point=ST_GeomFromGeoJSON($4), short_description=$5, link=$6, river_id=$7
    WHERE id=$1

--@by-id
SELECT
    white_water_rapid.id AS id,
    white_water_rapid.title AS title,
    ST_AsGeoJSON(point) as point,
    category,
    short_description,
    link,
    river_id,
    river.title as river_title
FROM white_water_rapid
INNER JOIN river ON white_water_rapid.river_id=river.id
    WHERE white_water_rapid.id=$1