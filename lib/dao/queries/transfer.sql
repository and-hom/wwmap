--@list
SELECT id, title, stations, description FROM transfer ORDER BY ID;

--@list-by-river
SELECT id, title, stations, description FROM transfer t INNER JOIN transfer_river tr on t.id = tr.transfer_id WHERE tr.river_id=$1 ORDER BY ID;

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