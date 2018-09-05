--@table
country
--@list
SELECT id,title FROM country ORDER BY CASE title WHEN '-' THEN NULL ELSE title END ASC