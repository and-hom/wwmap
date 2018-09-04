--@by-id
SELECT id, report_id, white_water_rapid_id,source,remote_id,url,preview_url,date_published, enabled, "type", main_image
FROM image WHERE id=$1

--@list
SELECT id, report_id, white_water_rapid_id,source,remote_id,url,preview_url,date_published, enabled, "type", main_image
FROM image WHERE white_water_rapid_id=$1 AND "type"=$2 AND (NOT $3 OR enabled) ORDER BY id DESC LIMIT $4

--@upsert
INSERT INTO image(report_id, white_water_rapid_id,source,remote_id,url,preview_url,date_published, "type")
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT(source, remote_id) DO UPDATE SET date_published=image.date_published RETURNING id

--@insert-local
INSERT INTO image(id,report_id, white_water_rapid_id, "type", source,remote_id,url,preview_url,date_published)
VALUES (nextval('id_gen'), 0, $1, $2, $3, CAST(currval('id_gen') AS CHARACTER VARYING),'','',$4)
RETURNING id, enabled

--@delete
DELETE FROM image WHERE id=$1

--@set-enabled
UPDATE image SET enabled=$1 WHERE id=$2

--@get-main
SELECT id, report_id, white_water_rapid_id,source,remote_id,url,preview_url,date_published, enabled, "type", main_image
FROM image WHERE white_water_rapid_id=$1 AND main_image LIMIT 1

--@set-main
UPDATE image SET
    main_image=CASE id WHEN $2 THEN TRUE ELSE FALSE END
    WHERE white_water_rapid_id=$1

--@drop-main-for-spot
UPDATE image SET main_image=FALSE WHERE white_water_rapid_id=$1