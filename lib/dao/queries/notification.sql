--@table
"notification"

--@insert
INSERT INTO ___table___(title, object_id, object_title, comment, provider, recipient, classifier, send_before) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id

--@unread-provider-recipient-classifier
SELECT provider, recipient, classifier
    FROM ___table___
    WHERE NOT read
    GROUP BY provider, recipient, classifier
    HAVING min(send_before)<$1

--@list-unread
SELECT id, title, object_id, object_title, comment, created_at, provider, recipient, classifier, send_before
FROM ___table___
WHERE NOT read AND provider=$2 AND recipient=$3 AND classifier=$4
ORDER BY created_at ASC LIMIT $1

--@mark-read
UPDATE ___table___ SET read=TRUE WHERE id = ANY($1)