--@table
"user"

--@create
INSERT INTO ___table___(ext_id, auth_provider, role, info, session_id) VALUES($1, $2, $3, $4, $5)
ON CONFLICT(auth_provider, ext_id) DO UPDATE SET info=$4, session_id=COALESCE(NULLIF("user".session_id,''), $5)
RETURNING id, role, session_id, xmax=0 as created
--@get-role-by-ext-id
SELECT "role" FROM ___table___ WHERE auth_provider=$1 AND ext_id=$2
--@set-role
UPDATE ___table___ SET "role"=$2 WHERE id=$1 RETURNING "role" AS new_role, (SELECT "role" FROM ___table___ u WHERE u.id=$1) AS old_role

--@user-fields
id, ext_id, auth_provider,"role",info,session_id

--@list
SELECT ___user-fields___ FROM ___table___ ORDER BY id ASC
--@list-by-role
SELECT ___user-fields___ FROM ___table___ WHERE "role"=$1 ORDER BY id ASC
--@find-by-session-id
SELECT ___user-fields___ FROM ___table___ WHERE session_id=$1