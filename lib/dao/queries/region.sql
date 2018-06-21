--@list-real
SELECT id, country_id, title FROM region WHERE country_id=$1 AND not fake
--@get-by-id
SELECT id, country_id, title FROM region WHERE id=$1
--@list-all-with-country
SELECT region.id AS id, country.id AS country_id, country.title AS country_title, region.title
    FROM region INNER JOIN country ON region.country_id=country.id
