--@values
changes_log.object_type,
changes_log.object_id,
changes_log.auth_provider,
changes_log.ext_id,
changes_log.login,
changes_log."type",
changes_log.description,
changes_log."time"

--@insert
INSERT INTO changes_log(
    object_type,
    object_id,
    auth_provider,
    ext_id,
    login,
    "type",
    description,
    "time"
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id;
--@list
SELECT id, ___values___ FROM changes_log WHERE object_type=$1 AND object_id=$2 ORDER BY "time" ASC LIMIT $3;
--@list-all
SELECT id, ___values___ FROM changes_log ORDER BY "time" DESC LIMIT $1;
--@list-time-range
SELECT id, ___values___ FROM changes_log WHERE time>=$1 AND time<$2 ORDER BY "time" DESC LIMIT $3;
--@list-all-with-user-info
SELECT changes_log.id, ___values___, COALESCE(info, '{}'::jsonb)
FROM changes_log
         LEFT OUTER JOIN "user" ON changes_log.auth_provider = "user".auth_provider AND changes_log.ext_id = "user".ext_id
ORDER BY "time" DESC
LIMIT $1;
--@list-with-user-info
SELECT
       changes_log.id,
       changes_log.object_type,
       changes_log.object_id,
       changes_log.auth_provider,
       changes_log.ext_id,
       changes_log.login,
       changes_log."type",
       changes_log.description,
       changes_log."time",
       COALESCE("user".info, '{}'::jsonb)
FROM changes_log
         LEFT OUTER JOIN "user" ON changes_log.auth_provider = "user".auth_provider AND changes_log.ext_id = "user".ext_id
WHERE object_type=$1 AND object_id=$2
ORDER BY "time" ASC
LIMIT $3;
