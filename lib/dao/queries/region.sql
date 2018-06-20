--@list-real
SELECT id, country_id, title FROM region WHERE country_id=$1 AND not fake