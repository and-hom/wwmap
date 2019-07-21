INSERT INTO waterway_ref(id,ref_id)
    SELECT w1.id AS id, w2.id AS ref_id FROM waterway w1
        INNER JOIN waterway_osm_ref lnk ON w1.osm_id=lnk.id
        INNER JOIN waterway w2 ON lnk.ref_id=w2.osm_id;