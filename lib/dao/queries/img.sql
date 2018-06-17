--@list
SELECT id, report_id, white_water_rapid_id,source,remote_id,url,preview_url,date_published
FROM image WHERE white_water_rapid_id=$1 LIMIT $2

--@upsert
INSERT INTO image(report_id, white_water_rapid_id,source,remote_id,url,preview_url,date_published)
VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING RETURNING id