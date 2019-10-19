--@insert
INSERT INTO waterway_osm_ref(id, ref_id, cross_point) VALUES ($1,$2,ST_GeomFromGeoJSON($3)) ON CONFLICT DO NOTHING