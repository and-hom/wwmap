CREATE TABLE white_water_rapid_image
(
    white_water_rapid_id BIGINT REFERENCES white_water_rapid (id),
    image_id             BIGINT REFERENCES image (id) ON DELETE CASCADE,
    main                 bool,
    PRIMARY KEY (white_water_rapid_id, image_id)
);

CREATE TABLE river_image
(
    river_id BIGINT REFERENCES river (id),
    image_id BIGINT REFERENCES image (id) ON DELETE CASCADE,
    main                 bool,
    PRIMARY KEY (river_id, image_id)
);


CREATE UNIQUE INDEX white_water_rapid_image_main ON white_water_rapid_image (white_water_rapid_id, main) WHERE main = true;
CREATE UNIQUE INDEX river_image_main ON river_image (river_id, main) WHERE main = true;


INSERT INTO white_water_rapid_image(white_water_rapid_id, image_id, main)
SELECT white_water_rapid_id, id, main_image
FROM image;

ALTER TABLE image
    DROP COLUMN white_water_rapid_id,
    DROP COLUMN main_image;

CREATE UNIQUE INDEX image_source_remote_id_idx ON image (source, remote_id);