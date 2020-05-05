--@table
country
--@list
SELECT id,title,code FROM country ORDER BY CASE title WHEN '-' THEN NULL ELSE title END ASC;
--@get
SELECT id,title,code FROM country WHERE id=$1;
--@get-by-code
SELECT id,title,code FROM country WHERE code=$1;