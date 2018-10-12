--@table
"notification"

--@insert
INSERT INTO @@table@@(object_id,comment) VALUES($1,$2) RETURNING id

--@list-unread
SELECT @@table@@.id, COALESCE(white_water_rapid.id, -1) as title, COALESCE(white_water_rapid.title, '') as title, COALESCE(river.title, '') as river_title, @@table@@.comment, @@table@@.created_at
FROM @@table@@ LEFT OUTER JOIN white_water_rapid ON @@table@@.object_id=white_water_rapid.id
LEFT OUTER JOIN river ON white_water_rapid.river_id=river.id
WHERE NOT @@table@@.read
ORDER BY created_at ASC LIMIT $1

--@mark-read
UPDATE @@table@@ SET read=TRUE WHERE id = ANY($1)