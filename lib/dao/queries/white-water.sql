--@table
white_water_rapid
--@select-columns
    ___table___.id AS id,
    ___table___.title AS title,
    ST_AsGeoJSON(point) as point,
    category,
    short_description,
    link,
    river_id,
    river.title as river_title

--@select-columns-full
    ___table___.id AS id,
    ___table___.title AS title,
    ST_AsGeoJSON(point) as point,
    category,
    short_description,
    link,
    river_id,
    river.title as river_title,
    region.id as region_id,
    region.country_id as country_id,
    region.fake as region_fake,

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

    ___table___.aliases,
    ___table___.props


--@by-box
SELECT ___select-columns___
FROM ___table___  LEFT OUTER JOIN river ON ___table___.river_id=river.id
WHERE point && ST_MakeEnvelope($1,$2,$3,$4)

--@by-river
SELECT ___select-columns___
FROM ___table___  LEFT OUTER JOIN river ON ___table___.river_id=river.id
WHERE river_id=$1
ORDER BY order_index ASC

--@by-title-part
WITH
    alias_query AS (SELECT id, jsonb_array_elements_text(aliases) AS alias FROM white_water_rapid),
    spot_with_alias_query AS (SELECT ___table___.id as id, river_id, title, alias_query.alias AS alias FROM ___table___ LEFT OUTER JOIN alias_query ON ___table___.id=alias_query.id),
    rank_query AS (SELECT spot_with_alias_query.id,
                            wwmap_search(spot_with_alias_query.title, $1) AS title_rank,
                            wwmap_search(spot_with_alias_query.alias, $1) AS alias_rank,
                            wwmap_search(river.title, $1) AS river_rank FROM spot_with_alias_query LEFT OUTER JOIN river ON spot_with_alias_query.river_id=river.id WHERE river.visible=TRUE),
    final_rank_query AS (SELECT id, max(title_rank) + sum(alias_rank) AS own_rank, max(river_rank) AS river_rank FROM rank_query GROUP BY id)
SELECT ___select-columns___
FROM ___table___
    INNER JOIN final_rank_query ON ___table___.id=final_rank_query.id AND own_rank>0
    LEFT OUTER JOIN river ON ___table___.river_id=river.id
    INNER JOIN region ON river.region_id = region.id
    WHERE river.visible=TRUE
      AND CASE
              WHEN $2 = 0 THEN $3 = 0 OR region.country_id = $3
              ELSE region.id = $2
        END
    ORDER BY river_rank*10 + own_rank DESC
LIMIT $4 OFFSET $5

--@by-river-full
SELECT ___select-columns-full___
FROM ___table___
LEFT OUTER JOIN river ON ___table___.river_id=river.id
LEFT OUTER JOIN region ON river.region_id=region.id
WHERE river_id=$1
ORDER BY order_index ASC

--@by-river-and-title
SELECT ___select-columns___
FROM ___table___  LEFT OUTER JOIN river ON ___table___.river_id=river.id
WHERE river_id=$1 AND title=$2

--@insert
INSERT INTO ___table___(title,category,point,short_description, link, river_id)
		VALUES ($2, $3, ST_GeomFromGeoJSON($4), $5, $6, $7)

--@update
UPDATE ___table___ SET title=$2,category=$3, point=ST_GeomFromGeoJSON($4), short_description=$5, link=$6, river_id=$7
    WHERE id=$1

--@by-id
SELECT ___select-columns___
FROM ___table___
INNER JOIN river ON ___table___.river_id=river.id
    WHERE ___table___.id=$1

--@by-id-full
SELECT ___select-columns-full___
FROM ___table___
INNER JOIN river ON ___table___.river_id=river.id
LEFT OUTER JOIN region ON river.region_id=region.id
    WHERE ___table___.id=$1

--@insert-full
INSERT INTO ___table___(title,category, point, short_description, link, river_id,
    lw_category, lw_description,
    mw_category, mw_description,
    hw_category, hw_description,
    orient, approach, safety,
    order_index, auto_ordering, aliases, props)
    VALUES ($1,$2,ST_GeomFromGeoJSON($3),$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,
    $16,$17,$18, $19) RETURNING id
    
--@update-full
UPDATE ___table___ SET title=$2,category=$3, point=ST_GeomFromGeoJSON($4), short_description=$5, link=$6, river_id=$7,
    lw_category=$8, lw_description=$9,
    mw_category=$10, mw_description=$11,
    hw_category=$12, hw_description=$13,
    orient=$14, approach=$15, safety=$16,
    order_index=CASE WHEN $18 THEN order_index ELSE $17 END, auto_ordering=$18,
    last_auto_ordering=CASE ST_GeomFromGeoJSON($4) WHEN point THEN last_auto_ordering ELSE NULL END,
    aliases=$19, props=$20
    WHERE id=$1

--@delete
DELETE FROM ___table___ WHERE id=$1

--@delete-for-river
DELETE FROM ___table___ WHERE river_id=$1

--@geom-center-by-river
SELECT center FROM (
    SELECT ST_AsGeoJSON(ST_Centroid(ST_Collect(point))) center FROM ___table___ WHERE river_id=$1
) sq WHERE center IS NOT NULL

--@river-bounds
SELECT bounds FROM (
    SELECT ST_AsGeoJSON(ST_Extent(point)) bounds FROM ___table___ WHERE river_id=$1
) sq WHERE bounds IS NOT NULL


--@auto-ordering-river-ids
SELECT river_id FROM ___table___
    WHERE auto_ordering and river_id IS NOT NULL
    GROUP BY river_id
    HAVING count(1)>1 AND count(distinct last_auto_ordering) + sum(CASE WHEN last_auto_ordering IS NULL THEN 1 ELSE 0 END) > 1

--@distance-from-beginning
WITH p AS (select ST_GeomFromGeoJSON($2) AS path),
	wwpts AS (SELECT id,point FROM ___table___ WHERE auto_ordering AND river_id=$1)
SELECT wwpts.id, ST_Length(ST_LineSubstring(
        path,
        ST_LineLocatePoint(path, ST_StartPoint(path)),
        ST_LineLocatePoint(path, point))::geography)::int
        FROM p inner join wwpts on true
        WHERE ST_Distance(path::geography,point::geography)<$3
        ORDER BY 2

--@update-order-idx
UPDATE ___table___ SET order_index=$2,last_auto_ordering=$3  WHERE id=$1

--@parent-ids
SELECT ___table___.id AS spot_id, river.id AS river_id,
       CASE WHEN region.fake THEN 0 ELSE region.id END AS region_id,
       region.country_id AS country_id,
       ___table___.title AS spot_title,
       river.title AS river_title
FROM ___table___
         INNER JOIN river ON ___table___.river_id = river.id
         INNER JOIN region ON river.region_id = region.id
WHERE ___table___.id = ANY($1)