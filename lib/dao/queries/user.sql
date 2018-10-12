--@create
INSERT INTO "user"(ext_id, auth_provider, role, info) VALUES($1, $2, $3, $4)
ON CONFLICT(ext_id) DO UPDATE SET info=$4
RETURNING id, role, xmax=0 as created
--@get-role-by-ext-id
SELECT "role" FROM "user" WHERE auth_provider=$1 AND ext_id=$2
--@set-role
UPDATE "user" SET "role"=$2 WHERE id=$1
--@list
SELECT id, ext_id, auth_provider,"role",info FROM "user" ORDER BY id ASC