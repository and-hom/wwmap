--@table
river

--@spot-counters
(SELECT to_json(spot_counters) FROM (
    SELECT COALESCE(sum(CASE

        WHEN auto_ordering AND last_auto_ordering=(
                select max(last_auto_ordering) FROM white_water_rapid where river_id=6605
            ) AND last_auto_ordering>to_timestamp(0) AND order_index>0 THEN 1
        WHEN order_index>0 THEN 1
        ELSE 0 END),0) as ordered,

        count(1) as total FROM white_water_rapid WHERE river_id=river.id

) spot_counters) AS spot_counters

--@bounds
SELECT ___table___.id id, ST_AsGeoJSON(ST_Extent(white_water_rapid.point)) bounds from
    ___table___ INNER JOIN white_water_rapid ON  ___table___.id=white_water_rapid.river_id
    GROUP BY ___table___.id

--@find-by-tags
WITH
  alias_query AS (SELECT id, jsonb_array_elements_text(aliases) AS alias FROM ___table___),
  rivers_query AS (SELECT ___table___.id, region_id, title, alias_query.alias AS alias, visible
                                  FROM ___table___ LEFT OUTER JOIN alias_query ON ___table___.id=alias_query.id)
SELECT sq.id, region_id, region.country_id, sq.title, region.title AS region_title, fake AS region_fake, NULL, NULL, '{}', visible FROM rivers_query sq
		INNER JOIN region ON sq.region_id=region.id
WHERE sq.title ilike ANY($1) OR alias ilike ANY($1)

--@inside-bounds
SELECT river.id, region_id, region.country_id, river.title, region.title AS region_title, fake AS region_fake,
       ST_AsGeoJSON(ST_Extent(white_water_rapid.point)), river.aliases, river.props, visible
    FROM ___table___
        INNER JOIN white_water_rapid ON river.id=white_water_rapid.river_id
        INNER JOIN region ON river.region_id=region.id
WHERE (river.visible OR $6) AND exists
    (SELECT 1 FROM white_water_rapid WHERE white_water_rapid.river_id=river.id AND point && ST_MakeEnvelope($1,$2,$3,$4))
GROUP BY river.id, region_id, region.country_id, region.title, region.fake ORDER BY popularity DESC LIMIT $5

--@by-id
SELECT river.id,region_id, region.country_id,river.title, region.title AS region_title, fake AS region_fake,NULL,
       river.aliases AS aliases, description, visible, river.props, ___spot-counters___
 FROM river INNER JOIN region ON river.region_id=region.id WHERE river.id=$1

--@for-image
SELECT river.id,region_id, region.country_id,river.title, region.title AS region_title, fake AS region_fake,NULL,
       river.aliases AS aliases, description, visible, river.props, ___spot-counters___
 FROM river
     INNER JOIN region ON river.region_id=region.id
     INNER JOIN white_water_rapid wwr ON river.id = wwr.river_id
     INNER JOIN image ON wwr.id = image.white_water_rapid_id
WHERE image.id=$1

--@for-spot
SELECT river.id,region_id, region.country_id,river.title, region.title AS region_title, fake AS region_fake,NULL,
       river.aliases AS aliases, description, visible, river.props, ___spot-counters___
FROM river
         INNER JOIN region ON river.region_id=region.id
         INNER JOIN white_water_rapid wwr ON river.id = wwr.river_id
WHERE wwr.id=$1

--@by-region
SELECT river.id, region_id, region.country_id, river.title, region.title AS region_title, fake AS region_fake, NULL,
       river.aliases, river.props, river.visible
    FROM river INNER JOIN region ON river.region_id=region.id WHERE region.id=$1
    ORDER BY CASE river.title WHEN '-' THEN NULL ELSE river.title END ASC

--@by-region-full
WITH bounds AS (___bounds___)
SELECT river.id, region_id, region.country_id, river.title, region.title AS region_title, fake AS region_fake, b.bounds, river.aliases, description, visible, river.props, ___spot-counters___
    FROM river INNER JOIN region ON river.region_id=region.id
     INNER JOIN (SELECT * FROM bounds) b ON b.id=___table___.id
    WHERE region.id=$1
    ORDER BY CASE river.title WHEN '-' THEN NULL ELSE river.title END ASC

--@all
SELECT river.id as id, region_id, region.country_id, river.title, region.title AS region_title, fake AS region_fake, NULL,
       river.aliases as aliases, river.props, river.visible
    FROM river INNER JOIN region ON river.region_id=region.id

--@by-country
SELECT river.id as id, region_id, region.country_id, river.title, region.title AS region_title, fake AS region_fake, NULL,
       river.aliases as aliases, river.props, river.visible
    FROM river INNER JOIN region ON river.region_id=region.id WHERE region.fake AND region.country_id=$1
    ORDER BY CASE river.title WHEN '-' THEN NULL ELSE river.title END ASC

--@by-country-full
WITH bounds AS (___bounds___)
SELECT river.id as id, region_id, region.country_id, river.title, region.title AS region_title, fake AS region_fake,
    b.bounds, river.aliases as aliases, description, visible, river.props, ___spot-counters___
    FROM river INNER JOIN region ON river.region_id=region.id
     INNER JOIN (SELECT * FROM bounds) b ON b.id=___table___.id
    WHERE region.fake AND region.country_id=$1
    ORDER BY CASE river.title WHEN '-' THEN NULL ELSE river.title END ASC

--@by-first-letters
SELECT river.id, region_id, country_id, river.title, '', fake, NULL, river.aliases, river.props, river.visible
  FROM river LEFT OUTER JOIN region ON river.region_id=region.id
  WHERE river.title ilike $1||'%' LIMIT $2

--@update-full
UPDATE river SET region_id=$2, title=$3, aliases=$4, description=$5, props=$6 WHERE id=$1

--@update
UPDATE river SET region_id=$2, title=$3, aliases=$4, props=$5 WHERE id=$1

--@insert
INSERT INTO river(region_id, title, aliases, description,props) VALUES($1,$2,$3,$4,$5) RETURNING id

--@delete
DELETE FROM river WHERE id=$1

--@set-visible
UPDATE river SET visible=$2 WHERE id=$1

--@by-title-part
WITH
  alias_query AS (
      SELECT riv.id, jsonb_array_elements_text(riv.aliases) AS alias FROM river riv
        INNER JOIN region reg ON riv.region_id = reg.id
      WHERE ($2=0 OR reg.id = $2) AND ($3=0 OR reg.country_id = $3)
      ),
  visible_rivers_query AS (SELECT river.id, title, alias_query.alias AS alias FROM river
                                LEFT OUTER JOIN alias_query ON river.id=alias_query.id
                                WHERE visible=TRUE OR $6),
  rank_query AS (SELECT id,
                        wwmap_search(visible_rivers_query.title, $1) AS title_rank ,
                        wwmap_search(visible_rivers_query.alias, $1) AS alias_rank
                    FROM visible_rivers_query),
  final_rank_query AS (SELECT id, max(title_rank) + sum(alias_rank) AS own_rank FROM rank_query GROUP BY id),
  river_data_query AS (SELECT river.id, region_id, region.country_id, river.title, region.title AS region_title, fake AS region_fake,
                              ST_AsGeoJSON(ST_Extent(white_water_rapid.point)), river.aliases, river.props, visible
                       FROM river
                                INNER JOIN final_rank_query ON river.id=final_rank_query.id AND own_rank>0
                                INNER JOIN white_water_rapid ON river.id=white_water_rapid.river_id
                                INNER JOIN region ON river.region_id=region.id
                       GROUP BY river.id, region_id, region.country_id, region.title, region.fake)
SELECT river_data_query.*
    FROM river_data_query INNER JOIN final_rank_query ON river_data_query.id=final_rank_query.id
    ORDER BY final_rank_query.own_rank DESC
LIMIT $4 OFFSET $5

--@parent-ids
SELECT ___table___.id AS river_id, CASE WHEN region.fake THEN 0 ELSE region.id END AS region_id, region.country_id AS country_id,
       ___table___.title AS river_title
    FROM ___table___
    INNER JOIN region ON ___table___.region_id = region.id
WHERE ___table___.id = ANY($1)

--@count-by-region
SELECT count(1) FROM river WHERE region_id=$1
