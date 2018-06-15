--@insert
INSERT INTO report(object_id,comment) VALUES($1,$2) RETURNING id
--@list-unread
SELECT report.id, COALESCE(white_water_rapid.id, -1) as title, COALESCE(white_water_rapid.title, '') as title, COALESCE(river.title, '') as river_title, report.comment, report.created_at
FROM report LEFT OUTER JOIN white_water_rapid ON report.object_id=white_water_rapid.id
LEFT OUTER JOIN river ON white_water_rapid.river_id=river.id
WHERE NOT report.read
ORDER BY created_at ASC LIMIT $1
--@mark-read
UPDATE report SET read=TRUE WHERE id = ANY($1)