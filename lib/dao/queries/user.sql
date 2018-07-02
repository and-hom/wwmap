--@create
INSERT INTO "user"(yandex_id, role, info) VALUES($1, $2, $3)
ON CONFLICT(yandex_id) DO UPDATE SET info=$3
RETURNING id
--@get-role-by-yandex-id
SELECT "role" FROM "user" WHERE yandex_id=$1