--@fields
id, report_id, white_water_rapid_id,source,remote_id,url,preview_url,date_published, enabled, "type", main_image,
date_level_updated, level

--@by-id
SELECT ___fields___
FROM image WHERE id=$1

--@list
SELECT ___fields___
FROM image WHERE white_water_rapid_id=$1 AND "type"=$2 AND (NOT $3 OR enabled) ORDER BY id DESC LIMIT $4

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
INSERT INTO image(id,report_id, white_water_rapid_id, "type", source,remote_id,url,preview_url,date_published)
VALUES (nextval('id_gen'), 0, $1, $2, $3, CAST(currval('id_gen') AS CHARACTER VARYING),'','',$4)
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

--@parent-ids
SELECT id AS image_id, white_water_rapid_id AS spot_id FROM image WHERE image.id = ANY($1)