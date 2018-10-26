--@inside-bounds
WITH river_ids AS (
    SELECT DISTINCT river_id AS id FROM white_water_rapid WHERE  point && ST_MakeEnvelope($1,$2,$3,$4))
SELECT (river).id, (river).title,
        (white_water_rapid).id, (white_water_rapid).title, ST_AsGeoJSON((white_water_rapid).point), (white_water_rapid).category, (white_water_rapid).link, (white_water_rapid).props,
        COALESCE((image).id, -1), COALESCE((image).source, ''), COALESCE((image).remote_id, ''), COALESCE((image).url, ''), COALESCE((image).preview_url, ''),
        COALESCE((image).date_published, to_timestamp(0)), COALESCE((image)."type", '')
FROM (
    SELECT
            river, white_water_rapid, image,
            ROW_NUMBER() OVER (PARTITION BY river.id, white_water_rapid.id ORDER BY image.id) AS r
        FROM river
        INNER JOIN white_water_rapid ON river.id=white_water_rapid.river_id
        LEFT OUTER JOIN image ON white_water_rapid.id=image.white_water_rapid_id
    WHERE (river.visible OR $6) AND river.id=ANY(SELECT id FROM river_ids) AND (image.enabled OR image.id IS NULL)
        ORDER BY river.id, white_water_rapid.order_index, white_water_rapid.id, image.main_image DESC, image.date_published DESC
) sq WHERE r<=$5