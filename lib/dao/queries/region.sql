--@table
region
--@list-real
SELECT id, country_id, title, fake FROM region WHERE country_id=$1 AND not fake
    ORDER BY CASE title WHEN '-' THEN NULL ELSE title END ASC
--@get-by-id
SELECT id, country_id, title, fake FROM region WHERE id=$1
--@list-all-with-country
SELECT region.id AS id, country.id AS country_id, country.title AS country_title, region.title, region.fake
    FROM region INNER JOIN country ON region.country_id=country.id
--@get-fake
SELECT id, country_id, title, fake FROM region WHERE country_id=$1 AND fake LIMIT 1
--@create-fake
INSERT INTO region(country_id, title, fake) VALUES($1,  md5(random()::text), TRUE) RETURNING id
