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
ORDER BY order_index ASC

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
ORDER BY order_index ASC

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

--@by-id-full
SELECT
    white_water_rapid.id AS id,
    white_water_rapid.title AS title,
    ST_AsGeoJSON(point) as point,
    category,
    short_description,
    link,
    river_id,
    river.title as river_title,

    lw_category,
    lw_description,
    mw_category,
    mw_description,
    hw_category,
    hw_description,

    orient,
    approach,
    safety,

    preview,
    order_index,
    auto_ordering,
    last_auto_ordering
FROM white_water_rapid
INNER JOIN river ON white_water_rapid.river_id=river.id
    WHERE white_water_rapid.id=$1

--@insert-full
INSERT INTO white_water_rapid(title,category, point, short_description, link, river_id,
    lw_category, lw_description,
    mw_category, mw_description,
    hw_category, hw_description,
    orient, approach, safety,
    preview,
    order_index, auto_ordering)
    VALUES ($1,$2,ST_GeomFromGeoJSON($3),$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18) RETURNING id
    
--@update-full
UPDATE white_water_rapid SET title=$2,category=$3, point=ST_GeomFromGeoJSON($4), short_description=$5, link=$6, river_id=$7,
    lw_category=$8, lw_description=$9,
    mw_category=$10, mw_description=$11,
    hw_category=$12, hw_description=$13,
    orient=$14, approach=$15, safety=$16,
    preview=$17,
    order_index=CASE WHEN $19 THEN order_index ELSE $18 END, auto_ordering=$19
    WHERE id=$1

--@delete
DELETE FROM white_water_rapid WHERE id=$1