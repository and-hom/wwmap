--@fields
img.id,
img.report_id,
img.source,
img.remote_id,
img.url,
img.preview_url,
img.date_published,
img.enabled,
img."type",
img.date,
img.date_level_updated,
img.level,
props

--@by-id
SELECT ___fields___
FROM image img WHERE id=$1;

--@list
SELECT ___fields___
FROM image img INNER JOIN white_water_rapid_image  wwrimg ON img.id = wwrimg.image_id
WHERE wwrimg.white_water_rapid_id=$1 AND img."type"=$2 AND (NOT $3 OR img.enabled) ORDER BY img.id DESC LIMIT $4;

--@list-ext
SELECT ___fields___, vr.url AS report_url, vr.title AS report_title
FROM image img
    INNER JOIN white_water_rapid_image  wwrimg ON img.id = wwrimg.image_id
    LEFT OUTER JOIN voyage_report vr on img.report_id = vr.id
WHERE white_water_rapid_id=$1 AND "type"=$2 AND (NOT $3 OR enabled) ORDER BY id DESC LIMIT $4;

--@list-all-by-spot
SELECT ___fields___
FROM image img
         INNER JOIN white_water_rapid_image wwrimg ON img.id = wwrimg.image_id
WHERE white_water_rapid_id = $1;

--@list-all-by-river
SELECT ___fields___
FROM image img
    INNER JOIN white_water_rapid_image  wwrimg ON img.id = wwrimg.image_id
    INNER JOIN white_water_rapid wwr ON wwrimg.white_water_rapid_id = wwr.id
WHERE wwr.river_id=$1;

--@list-main-by-river
SELECT ___fields___ FROM
(SELECT ROW_NUMBER() OVER (PARTITION BY white_water_rapid_id ORDER BY wwrimg.main DESC) AS r_num, *
 FROM image img INNER JOIN white_water_rapid_image  wwrimg ON img.id = wwrimg.image_id
WHERE white_water_rapid_id IN (SELECT id FROM white_water_rapid WHERE river_id=$1) AND "type"=$2) img WHERE r_num<=1;

--@upsert
INSERT INTO image(report_id,source,remote_id,url,preview_url,date_published, date, "type", props)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, COALESCE($9, '{}'::jsonb))
ON CONFLICT(source, remote_id) DO UPDATE SET date_published=image.date_published RETURNING id;

--@insert-local
INSERT INTO image(id,report_id, "type", source,remote_id,url,preview_url,date_published, date, date_level_updated, level)
VALUES (nextval('id_gen'), 0, $1, $2, CAST(currval('id_gen') AS CHARACTER VARYING),'','',$3, $4, $5, $6)
RETURNING id, enabled;

--@delete-by-spot
DELETE FROM image WHERE id IN (
    SELECT image_id from white_water_rapid_image WHERE white_water_rapid_id=$1
);

--@delete-by-river
DELETE FROM image WHERE id IN (
    SELECT image_id FROM white_water_rapid_image wwrimg
        INNER JOIN white_water_rapid wwr ON wwrimg.white_water_rapid_id=wwr.id
    WHERE river_id=$1
);

--@delete
DELETE FROM image WHERE id=$1;

--@set-enabled
UPDATE image SET enabled=$1 WHERE id=$2;

--@get-main
SELECT ___fields___
FROM image img INNER JOIN white_water_rapid_image  wwrimg ON img.id = wwrimg.image_id
WHERE white_water_rapid_id=$1 AND wwrimg.main LIMIT 1;

--@set-main
UPDATE white_water_rapid_image
SET main = CASE image_id WHEN $2 THEN TRUE ELSE FALSE END
WHERE white_water_rapid_id=$1;

--@drop-main-for-spot
UPDATE white_water_rapid_image SET main=false WHERE white_water_rapid_id = $1;

--@set-level-and-date
UPDATE image SET date=$2, level=$3, date_level_updated=$4 WHERE id=$1;

--@set-manual-level
UPDATE image
SET level=level || jsonb_build_object('0', $2::INT)
WHERE id = $1
RETURNING level::varchar;

--@reset-manual-level
UPDATE image
SET level=level-'0'
WHERE id = $1
RETURNING level::varchar;

--@parent-ids
SELECT image_id, white_water_rapid_id AS spot_id
FROM white_water_rapid_image
WHERE image_id = ANY($1);