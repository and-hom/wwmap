--@by-box
SELECT 
white_water_rapid.id AS id, 
osm_id, 
type, 
white_water_rapid.title AS title, 
comment, 
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
osm_id, 
type, 
white_water_rapid.title AS title, 
comment, 
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
osm_id, 
type, 
white_water_rapid.title AS title, 
comment, 
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
osm_id, 
type, 
white_water_rapid.title AS title, 
comment, 
ST_AsGeoJSON(point) as point, 
category, 
short_description, 
link, 
river_id, 
river.title as river_title, CASE WHEN region.fake THEN NULL ELSE region.title END AS region_title, country.title as country_title
FROM white_water_rapid 
INNER JOIN river ON white_water_rapid.river_id=river.id
INNER JOIN region ON river.region_id=region.id
INNER JOIN country ON region.country_id=country.id

--@insert
INSERT INTO white_water_rapid(osm_id, title,type,category,comment,point,short_description, link, river_id)
		VALUES ($1, $2, $3, $4, $5, ST_GeomFromGeoJSON($6), $7, $8, $9)