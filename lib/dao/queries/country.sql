--@table
country
--@list
SELECT id,title,code FROM country ORDER BY CASE title WHEN '-' THEN NULL ELSE title END ASC;
--@get
SELECT id,title,code FROM country WHERE id=$1;
--@get-by-code
SELECT id,title,code FROM country WHERE code=$1;
--@insert
INSERT INTO country(title, code) VALUES($1, $2) RETURNING id;
--@update
UPDATE country SET title=$1, code=$2 WHERE id=$3;
--@delete
DELETE FROM country WHERE id=$1
