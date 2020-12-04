1. Fetch rivers from OSM file using command `./ref_fetcher region file.osm`. Region is part of table name: `waterway_${region}`
2. Run SQL to convert osm ref to ref 
    ```sql
    INSERT INTO waterway_ref(id,ref_id,cross_point)
       SELECT w1.id AS id, w2.id AS ref_id, lnk.cross_point AS cross_point FROM waterway w1
           INNER JOIN waterway_osm_ref lnk ON w1.osm_id=lnk.id
           INNER JOIN waterway w2 ON lnk.ref_id=w2.osm_id;
    ```
3. Run SQL to bind tracks to rivers
    ```sql
    WITH
        alias_query AS (SELECT id, jsonb_array_elements_text(aliases) AS alias FROM river),
        river_rects_query AS (SELECT river_id, ST_Envelope(ST_Extent(point)) as bounds FROM white_water_rapid GROUP BY river_id),
        rivers_query AS (SELECT river.id, title, alias_query.alias AS alias, bounds
            FROM river
            LEFT OUTER JOIN alias_query ON river.id = alias_query.id
            INNER JOIN river_rects_query ON river.id = river_rects_query.river_id)
    UPDATE waterway SET river_id=rivers_query.id
        FROM rivers_query
        WHERE  (lower(waterway.title)=lower(rivers_query.title) OR lower(waterway.title)=lower(rivers_query.alias))
          AND ST_Distance(waterway.path::geography, rivers_query.bounds::geography) < 300
    ```    
