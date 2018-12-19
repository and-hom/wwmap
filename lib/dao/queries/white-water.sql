--@table
white_water_rapid
--@select-columns
    @@table@@.id AS id,
    @@table@@.title AS title,
    ST_AsGeoJSON(point) as point,
    category,
    short_description,
    link,
    river_id,
    river.title as river_title

--@select-columns-full
    @@table@@.id AS id,
    @@table@@.title AS title,
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

    order_index,
    auto_ordering,
    last_auto_ordering,

    @@table@@.aliases,
    @@table@@.props


--@by-box
SELECT @@select-columns@@
FROM @@table@@  LEFT OUTER JOIN river ON @@table@@.river_id=river.id
WHERE point && ST_MakeEnvelope($1,$2,$3,$4)

--@by-river
SELECT @@select-columns@@
FROM @@table@@  LEFT OUTER JOIN river ON @@table@@.river_id=river.id
WHERE river_id=$1
ORDER BY order_index ASC

--@by-title-part
SELECT @@select-columns@@
FROM @@table@@  LEFT OUTER JOIN river ON @@table@@.river_id=river.id
WHERE @@table@@.id=ANY(
    SELECT DISTINCT id FROM
        (SELECT id, title, CASE aliases WHEN '[]' THEN NULL ELSE jsonb_array_elements_text(aliases) END AS alias FROM @@table@@) sq
      WHERE  title ilike '%'||$1||'%' OR alias ilike '%'||$2||'%'
    )
LIMIT $3 OFFSET $4

--@by-river-full
SELECT @@select-columns-full@@
FROM @@table@@  LEFT OUTER JOIN river ON @@table@@.river_id=river.id
WHERE river_id=$1
ORDER BY order_index ASC

--@by-river-and-title
SELECT @@select-columns@@
FROM @@table@@  LEFT OUTER JOIN river ON @@table@@.river_id=river.id
WHERE river_id=$1 AND title=$2

--@insert
INSERT INTO @@table@@(title,category,point,short_description, link, river_id)
		VALUES ($2, $3, ST_GeomFromGeoJSON($4), $5, $6, $7)

--@update
UPDATE @@table@@ SET title=$2,category=$3, point=ST_GeomFromGeoJSON($4), short_description=$5, link=$6, river_id=$7
    WHERE id=$1

--@by-id
SELECT @@select-columns@@
FROM @@table@@
INNER JOIN river ON @@table@@.river_id=river.id
    WHERE @@table@@.id=$1

--@by-id-full
SELECT @@select-columns-full@@
FROM @@table@@
INNER JOIN river ON @@table@@.river_id=river.id
    WHERE @@table@@.id=$1

--@insert-full
INSERT INTO @@table@@(title,category, point, short_description, link, river_id,
    lw_category, lw_description,
    mw_category, mw_description,
    hw_category, hw_description,
    orient, approach, safety,
    order_index, auto_ordering, aliases, props)
    VALUES ($1,$2,ST_GeomFromGeoJSON($3),$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,
    $16,$17,$18, $19) RETURNING id
    
--@update-full
UPDATE @@table@@ SET title=$2,category=$3, point=ST_GeomFromGeoJSON($4), short_description=$5, link=$6, river_id=$7,
    lw_category=$8, lw_description=$9,
    mw_category=$10, mw_description=$11,
    hw_category=$12, hw_description=$13,
    orient=$14, approach=$15, safety=$16,
    order_index=CASE WHEN $18 THEN order_index ELSE $17 END, auto_ordering=$18,
    last_auto_ordering=CASE ST_GeomFromGeoJSON($4) WHEN point THEN last_auto_ordering ELSE NULL END,
    aliases=$19, props=$20
    WHERE id=$1

--@delete
DELETE FROM @@table@@ WHERE id=$1

--@delete-for-river
DELETE FROM @@table@@ WHERE river_id=$1

--@geom-center-by-river
SELECT center FROM (
    SELECT ST_AsGeoJSON(ST_Centroid(ST_Collect(point))) center FROM @@table@@ WHERE river_id=$1
) sq WHERE center IS NOT NULL


--@auto-ordering-river-ids
SELECT river_id FROM @@table@@
    WHERE auto_ordering and river_id IS NOT NULL
    GROUP BY river_id
    HAVING count(1)>1 AND count(distinct last_auto_ordering) + sum(CASE WHEN last_auto_ordering IS NULL THEN 1 ELSE 0 END) > 1

--@distance-from-beginning
WITH p AS (select ST_GeomFromGeoJSON($2) AS path),
	wwpts AS (SELECT id,point FROM @@table@@ WHERE auto_ordering AND river_id=$1)
SELECT wwpts.id, ST_Length(ST_LineSubstring(
        path,
        ST_LineLocatePoint(path, ST_StartPoint(path)),
        ST_LineLocatePoint(path, point))::geography)::int
        FROM p inner join wwpts on true
        WHERE ST_Distance(path::geography,point::geography)<$3
        ORDER BY 2

--@update-order-idx
UPDATE @@table@@ SET order_index=$2,last_auto_ordering=$3  WHERE id=$1