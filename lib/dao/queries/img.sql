--@fields
image.id, image.report_id, image.white_water_rapid_id,image.source,image.remote_id,image.url,
image.preview_url,image.date_published, image.enabled, image."type", image.main_image,
image.date, image.date_level_updated, image.level

--@by-id
SELECT ___fields___
FROM image WHERE id=$1

--@list
SELECT ___fields___
FROM image WHERE white_water_rapid_id=$1 AND "type"=$2 AND (NOT $3 OR enabled) ORDER BY id DESC LIMIT $4

--@list-ext
SELECT ___fields___, vr.url AS report_url, vr.title AS report_title
FROM image
LEFT OUTER JOIN voyage_report vr on image.report_id = vr.id
WHERE white_water_rapid_id=$1 AND "type"=$2 AND (NOT $3 OR enabled) ORDER BY id DESC LIMIT $4

--@list-all-by-spot
SELECT ___fields___
FROM image WHERE white_water_rapid_id=$1

--@list-all-by-river
SELECT ___fields___
FROM image WHERE white_water_rapid_id IN (SELECT id FROM white_water_rapid WHERE river_id=$1)

--@list-main-by-river
SELECT ___fields___ FROM
(SELECT ROW_NUMBER() OVER (PARTITION BY white_water_rapid_id ORDER BY main_image DESC) AS r_num, *
FROM image WHERE white_water_rapid_id IN (SELECT id FROM white_water_rapid WHERE river_id=$1) AND "type"=$2) sq WHERE r_num<=1

--@upsert
INSERT INTO image(report_id, white_water_rapid_id,source,remote_id,url,preview_url,date_published, "type")
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT(source, remote_id) DO UPDATE SET date_published=image.date_published RETURNING id

--@insert-local
INSERT INTO image(id,report_id, white_water_rapid_id, "type", source,remote_id,url,preview_url,date_published, date)
VALUES (nextval('id_gen'), 0, $1, $2, $3, CAST(currval('id_gen') AS CHARACTER VARYING),'','',$4, $5)
RETURNING id, enabled

--@delete-by-spot
DELETE FROM image WHERE white_water_rapid_id=$1

--@delete-by-river
DELETE FROM image WHERE white_water_rapid_id IN (SELECT id FROM white_water_rapid WHERE river_id=$1)

--@delete
DELETE FROM image WHERE id=$1

--@set-enabled
UPDATE image SET enabled=$1 WHERE id=$2

--@get-main
SELECT ___fields___
FROM image WHERE white_water_rapid_id=$1 AND main_image LIMIT 1

--@set-main
UPDATE image SET
    main_image=CASE id WHEN $2 THEN TRUE ELSE FALSE END
    WHERE white_water_rapid_id=$1

--@drop-main-for-spot
UPDATE image SET main_image=FALSE WHERE white_water_rapid_id=$1

--@set-level-and-date
UPDATE image SET date=$2, level=$3, date_level_updated=$2 WHERE id=$1

--@set-manual-level
UPDATE image
SET level=level || jsonb_build_object('0', $2::INT)
WHERE id = $1
RETURNING level::varchar

    --@reset-manual-level
UPDATE image
SET level=level-'0'
WHERE id = $1
RETURNING level::varchar

--@parent-ids
SELECT id AS image_id, white_water_rapid_id AS spot_id FROM image WHERE image.id = ANY($1)