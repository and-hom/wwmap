--@upsert
INSERT INTO voyage_report(title, remote_id,source,url,date_published,date_modified,date_of_trip, tags, author)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (source, remote_id) DO UPDATE SET title=$1, url=$4, date_modified=$6, date_of_trip=$7, tags=$8, author=$9
RETURNING id;

--@insert
INSERT INTO voyage_report(title, remote_id,source,url,date_published,date_modified,date_of_trip, tags, author)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id;

--@update
UPDATE voyage_report
SET title = $2,
    remote_id=$3,
    source=$4,
    url=$5,
    date_published=$6,
    date_modified=$7,
    date_of_trip=$8,
    tags=$9,
    author=$10
WHERE id = $1;

--@get-last-id
SELECT COALESCE(max(date_modified), to_timestamp(-65000000000)) FROM voyage_report WHERE source=$1

--@find
SELECT id,title,remote_id,source,url,date_published,date_modified,date_of_trip, tags, author,
       array_agg(river_id) filter (where river_id is not null) :: bigint[]
FROM voyage_report
    LEFT OUTER JOIN voyage_report_river ON voyage_report.id = voyage_report_river.voyage_report_id
WHERE id=$1
GROUP BY 1,2,3,4,5,6,7,8,9,10;

--@list
SELECT id,title,remote_id,source,url,date_published,date_modified,date_of_trip, tags, author,
       array_agg(river_id) filter (where river_id is not null) :: bigint[]
FROM voyage_report
    LEFT OUTER JOIN voyage_report_river ON voyage_report.id = voyage_report_river.voyage_report_id
GROUP BY 1,2,3,4,5,6,7,8,9,10
ORDER BY 6 DESC, 7 DESC;

--@list-by-river
SELECT id,title,remote_id,source,url,date_published,date_modified,date_of_trip, tags, author,'[]'::json FROM (
SELECT ROW_NUMBER() OVER (PARTITION BY source ORDER BY date_of_trip DESC, date_published DESC) AS r_num,
*
FROM voyage_report INNER JOIN voyage_report_river ON voyage_report.id = voyage_report_river.voyage_report_id
WHERE voyage_report_river.river_id = $1) sq WHERE r_num<=$2 ORDER BY source, date_of_trip DESC, date_published DESC

--@list-all
SELECT id,title,remote_id,source,url,date_published,date_modified,date_of_trip, tags,
       author, array_agg(river_id) filter (where river_id is not null) :: bigint[]
FROM voyage_report LEFT OUTER JOIN voyage_report_river on voyage_report.id = voyage_report_river.voyage_report_id
WHERE source=$1
GROUP BY 1,2,3,4,5,6,7,8,9,10;

--@upsert-river-link
INSERT INTO voyage_report_river(voyage_report_id, river_id) VALUES($1,$2) ON CONFLICT DO NOTHING
--@delete-river-link
DELETE FROM voyage_report_river WHERE river_id=$1

--@remove
DELETE FROM voyage_report WHERE id=$1

--@list-refs-by-river
SELECT voyage_report_id FROM voyage_report_river WHERE river_id=$1;

--@count-refs-by-river
SELECT count(1) FROM voyage_report_river WHERE river_id=$1;

--@insert-refs
INSERT INTO voyage_report_river(voyage_report_id, river_id) VALUES ($1, $2);

--@delete-refs
DELETE FROM voyage_report_river WHERE voyage_report_id=$1;

--@delete-refs-by-river
DELETE FROM voyage_report_river WHERE river_id=$1;