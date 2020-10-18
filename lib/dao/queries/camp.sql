--@fields
id, osm_id, title, description, ST_AsGeoJSON(point), num_tent_places

--@list
SELECT ___fields___, array_agg(river_id) filter (where river_id is not null) :: bigint[]
FROM camp LEFT JOIN camp_river_ref on camp.id = camp_river_ref.camp_id
GROUP BY 1,2,3,4,5,6

--@find-witin-bounds
SELECT ___fields___,array[]::bigint[] FROM camp WHERE point && ST_MakeEnvelope($1,$2,$3,$4)

--@find
SELECT ___fields___, array_agg(river_id) filter (where river_id is not null) :: bigint[]
FROM camp LEFT JOIN camp_river_ref on camp.id = camp_river_ref.camp_id WHERE id=$1
GROUP BY 1,2,3,4,5,6

--@insert
INSERT INTO camp(osm_id, title, description, point, num_tent_places) VALUES ($1,$2,$3,ST_GeomFromGeoJSON($4),$5) RETURNING id

--@update
UPDATE camp SET osm_id=$2, title=$3, description=$4, point=ST_GeomFromGeoJSON($5), num_tent_places=$6 WHERE id=$1

--@remove
DELETE FROM camp WHERE id=$1

--@list-refs-by-river
SELECT camp_id FROM camp_river_ref WHERE river_id=$1;

--@insert-refs
INSERT INTO camp_river_ref(camp_id, river_id) VALUES ($1, $2);

--@delete-refs
DELETE FROM camp_river_ref WHERE camp_id=$1;

--@delete-refs-by-river
DELETE FROM camp_river_ref WHERE river_id=$1;

