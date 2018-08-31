--@put
INSERT INTO referer(host, schema, base_url, page_url) VALUES($1, $2, $3, $4)
ON CONFLICT(host) DO UPDATE SET last_access=now(),
    schema   =  CASE $2 WHEN 'https' THEN $2 ELSE referer.schema END,
    base_url =  CASE $2 WHEN 'https' THEN $3 ELSE referer.base_url END,
    page_url =  CASE $2 WHEN 'https' THEN $4 ELSE referer.page_url END
    WHERE referer.host=$1
--@list
SELECT host, schema, base_url, page_url FROM referer WHERE (last_access+$1) >= now()
--@remove
DELETE FROM referer WHERE (last_access+$1) < now()
