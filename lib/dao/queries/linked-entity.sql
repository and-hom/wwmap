--@list-rivers
SELECT r.id,
       r.region_id,
       reg.country_id,
       r.title,
       ST_AsGeoJSON(ST_Extent(white_water_rapid.point)) AS bounds
FROM river r
         INNER JOIN region reg ON r.region_id = reg.id
         INNER JOIN white_water_rapid ON r.id = white_water_rapid.river_id
WHERE r.id = ANY($1)
GROUP BY 1, 2, 3, 4;