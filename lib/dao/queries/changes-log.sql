--@table
changes_log

--@values
object_type,
object_id,
auth_provider,
ext_id,
login,
"type",
description,
"time"

--@insert
INSERT INTO ___table___(___values___) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id
--@list
SELECT id, ___values___ FROM ___table___ WHERE object_type=$1 AND object_id=$2 ORDER BY "time" ASC LIMIT $3
--@list-all
SELECT id, ___values___ FROM ___table___ ORDER BY "time" DESC LIMIT $1