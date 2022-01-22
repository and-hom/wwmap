ALTER TABLE image
    ADD COLUMN white_water_rapid_id BIGINT REFERENCES white_water_rapid(id),
    ADD COLUMN main_image BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX image_white_water_rapid_id_idx ON image(white_water_rapid_id);
CREATE INDEX image_main_image ON image (main_image);
CREATE UNIQUE INDEX image_white_source_remote_id_idx
    ON image (source, remote_id, white_water_rapid_id);


UPDATE image img
SET white_water_rapid_id = wwrimg.white_water_rapid_id, main_image = wwrimg.main
FROM white_water_rapid_image wwrimg
WHERE img.id = wwrimg.image_id;

DROP TABLE white_water_rapid_image;
DROP TABLE river_image;
DROP INDEX image_source_remote_id_idx;
