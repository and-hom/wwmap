--@list
SELECT id, title, stations, description, array_agg(river_id) filter (where river_id is not null) :: bigint[]
FROM transfer LEFT OUTER JOIN transfer_river ON transfer.id = transfer_river.transfer_id
GROUP BY 1,2,3,4
ORDER BY 1;

--@find
SELECT id, title, stations, description, array_agg(river_id) filter (where river_id is not null) :: bigint[]
FROM transfer LEFT OUTER JOIN transfer_river ON transfer.id = transfer_river.transfer_id
WHERE transfer.id=$1
GROUP BY 1,2,3,4
ORDER BY 1;

--@list-by-river
SELECT id, title, stations, description, array_agg(river_id) filter (where river_id is not null) :: bigint[]
FROM transfer t INNER JOIN transfer_river tr on t.id = tr.transfer_id
WHERE tr.river_id=$1
GROUP BY 1,2,3,4
ORDER BY 1;

--@insert
INSERT INTO transfer(title, stations, description) VALUES ($1, $2, $3) RETURNING id;

--@update
UPDATE transfer SET title=$2, stations=$3, description=$4 WHERE id=$1;

--@delete
DELETE FROM transfer WHERE id=$1;

--@list-refs-by-river
SELECT transfer_id FROM transfer_river WHERE river_id=$1;

--@count-refs-by-river
SELECT count(1) FROM transfer_river WHERE river_id=$1;

--@insert-refs
INSERT INTO transfer_river(transfer_id, river_id) VALUES ($1, $2);

--@delete-refs
DELETE FROM transfer_river WHERE transfer_id=$1;

--@delete-refs-by-river
DELETE FROM transfer_river WHERE river_id=$1;