--@table
country
--@list
SELECT id,title,code FROM country ORDER BY CASE title WHEN '-' THEN NULL ELSE title END ASC