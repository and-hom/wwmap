--@list
SELECT id, report_id, white_water_rapid_id,source,remote_id,url,preview_url,date_published
FROM image WHERE white_water_rapid_id=$1 LIMIT $2

--@upsert
INSERT INTO image(report_id, white_water_rapid_id,source,remote_id,url,preview_url,date_published)
VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT(source, remote_id) DO UPDATE SET date_published=image.date_published RETURNING id

--@insert-local
INSERT INTO image(id,report_id, white_water_rapid_id,source,remote_id,url,preview_url,date_published)
VALUES (nextval('id_gen'), 0, $1, $2, CAST(currval('id_gen') AS CHARACTER VARYING),'','',$5)
RETURNING id

--@delete
DELETE FROM image WHERE id=$1