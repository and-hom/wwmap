--@list
SELECT id, title, stations, description FROM transfer ORDER BY ID;

--@list-by-river
SELECT id, title, stations, description FROM transfer t INNER JOIN transfer_river tr on t.id = tr.transfer_id WHERE tr.river_id=$1 ORDER BY ID;

--@list-full
WITH
     river_bounds AS (
        SELECT river.id AS id, ST_AsGeoJSON(ST_Extent(white_water_rapid.point)) AS bounds from
        river INNER JOIN white_water_rapid ON  river.id=white_water_rapid.river_id  GROUP BY river.id ),
     river_data AS (
        SELECT r.id, r.title, reg.id AS region_id, reg.country_id, b.bounds FROM river r
            INNER JOIN region reg ON r.region_id = reg.id
            LEFT OUTER JOIN river_bounds b ON b.id = r.id
        )
SELECT t.id, t.title, t.stations, t.description, r.id, r.region_id, r.country_id, r.title, r.bounds FROM transfer t
LEFT OUTER JOIN transfer_river tr ON t.id = tr.transfer_id
LEFT OUTER JOIN river_data r ON r.id = tr.river_id;

--@insert
INSERT INTO transfer(title, stations, description) VALUES ($1, $2, $3) RETURNING id;

--@update
UPDATE transfer SET title=$2, stations=$3, description=$4 WHERE id=$1;

--@delete
DELETE FROM transfer WHERE id=$1;

--@list-refs-by-river
SELECT transfer_id FROM transfer_river WHERE river_id=$1;

--@insert-refs
INSERT INTO transfer_river(transfer_id, river_id) VALUES ($1, $2);

--@delete-refs
DELETE FROM transfer_river WHERE transfer_id=$1;

--@delete-refs-by-river
DELETE FROM transfer_river WHERE river_id=$1;