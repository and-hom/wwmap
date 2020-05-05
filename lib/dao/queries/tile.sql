--@inside-bounds
WITH river_ids AS (
    SELECT DISTINCT river_id AS id FROM white_water_rapid WHERE point && ST_MakeEnvelope($3,$4,$5,$6))
    ___select-rivers-internal___

--@spots-by-river-id
WITH river_ids AS (SELECT $3::bigint AS id)
    ___select-rivers-internal___

--@inside-bounds-by-region-id
WITH river_ids AS (
    SELECT DISTINCT river_id AS id FROM white_water_rapid
        INNER JOIN river ON white_water_rapid.river_id = river.id
    WHERE point && ST_MakeEnvelope($3,$4,$5,$6) AND region_id=$7)
    ___select-rivers-internal___

--@inside-bounds-by-country-id
WITH river_ids AS (
    SELECT DISTINCT river_id AS id FROM white_water_rapid
        INNER JOIN river ON white_water_rapid.river_id = river.id
        INNER JOIN region ON river.region_id = region.id
    WHERE point && ST_MakeEnvelope($3,$4,$5,$6) AND country_id=$7)
    ___select-rivers-internal___

--@select-rivers-internal
SELECT (river).id, (river).title, CASE (region).fake WHEN TRUE THEN 0 ELSE (region).id END, (region).country_id,
        (white_water_rapid).id, (white_water_rapid).title, (white_water_rapid).short_description,
        ST_AsGeoJSON((white_water_rapid).point), (white_water_rapid).category, (white_water_rapid).link, (white_water_rapid).props,
        COALESCE((image).id, -1), COALESCE((image).source, ''), COALESCE((image).remote_id, ''), COALESCE((image).url, ''), COALESCE((image).preview_url, ''),
        COALESCE((image).date_published, to_timestamp(0)), COALESCE((image)."type", ''),
       (image).date, COALESCE((image).date_level_updated, to_timestamp(0)), (image).level
FROM (
    SELECT
            region, river, white_water_rapid, image,
            ROW_NUMBER() OVER (PARTITION BY river.id, white_water_rapid.id ORDER BY image.id) AS r
        FROM river
        INNER JOIN region ON river.region_id=region.id
        INNER JOIN white_water_rapid ON river.id=white_water_rapid.river_id
        LEFT OUTER JOIN image ON white_water_rapid.id=image.white_water_rapid_id
    WHERE (river.visible OR $2) AND river.id=ANY(SELECT id FROM river_ids) AND (image.enabled OR image.id IS NULL)
        ORDER BY river.id, white_water_rapid.order_index, white_water_rapid.id, image.main_image DESC, image.date_published DESC
) sq WHERE r<=$1


--@by-id
SELECT (river).id, (river).title, (river).description, (river).props,
        CASE (region).fake WHEN TRUE THEN 0 ELSE (region).id END,
        CASE (region).fake WHEN TRUE THEN '' ELSE (region).title END,
        (region).country_id,
        (white_water_rapid).id, (white_water_rapid).title, (white_water_rapid).short_description,
        ST_AsGeoJSON((white_water_rapid).point), (white_water_rapid).category, (white_water_rapid).link, (white_water_rapid).props,
        COALESCE((image).id, -1), COALESCE((image).source, ''), COALESCE((image).remote_id, ''), COALESCE((image).url, ''), COALESCE((image).preview_url, ''),
        COALESCE((image).date_published, to_timestamp(0)), COALESCE((image)."type", '')
FROM (
    SELECT
            region, river, white_water_rapid, image,
            ROW_NUMBER() OVER (PARTITION BY river.id, white_water_rapid.id, image.type ORDER BY image.id) AS r
        FROM river
        INNER JOIN region ON river.region_id=region.id
        INNER JOIN white_water_rapid ON river.id=white_water_rapid.river_id
        LEFT OUTER JOIN image ON white_water_rapid.id=image.white_water_rapid_id
    WHERE river.id=$1 AND (image.enabled AND image."type" = ANY($2) OR image.id IS NULL)
        ORDER BY river.id, white_water_rapid.order_index, white_water_rapid.id, image.main_image DESC, image.date_published DESC
) sq WHERE r<=$3