--@table
changes_log

--@values
___table___.object_type,
___table___.object_id,
___table___.auth_provider,
___table___.ext_id,
___table___.login,
___table___."type",
___table___.description,
___table___."time"

--@insert
INSERT INTO ___table___(___values___) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id
--@list
SELECT id, ___values___ FROM ___table___ WHERE object_type=$1 AND object_id=$2 ORDER BY "time" ASC LIMIT $3
--@list-all
SELECT id, ___values___ FROM ___table___ ORDER BY "time" DESC LIMIT $1
--@list-time-range
SELECT id, ___values___ FROM ___table___ WHERE time>=$1 AND time<$2 ORDER BY "time" DESC LIMIT $3
--@list-all-with-user-info
SELECT changes_log.id, ___values___, COALESCE(info, '{}'::jsonb)
FROM changes_log
         LEFT OUTER JOIN "user" ON changes_log.auth_provider = "user".auth_provider AND changes_log.ext_id = "user".ext_id
ORDER BY "time" DESC
LIMIT $1
--@list-with-user-info
SELECT changes_log.id, ___values___, COALESCE(info, '{}'::jsonb)
FROM changes_log
         LEFT OUTER JOIN "user" ON changes_log.auth_provider = "user".auth_provider AND changes_log.ext_id = "user".ext_id
WHERE object_type=$1 AND object_id=$2
ORDER BY "time" ASC
LIMIT $3