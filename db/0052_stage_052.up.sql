ALTER TABLE image
    ADD COLUMN date_level_updated DATE,
    ADD COLUMN date DATE,
    ADD COLUMN level JSONB NOT NULL DEFAULT '{}'::jsonb,
    ALTER COLUMN date_published TYPE timestamp with time zone;
COMMENT ON COLUMN image.date_level_updated IS
    'Date of computed water level. Normally should be equals to date_published. On re-calculate water level values should set date_level_updated=date_published';

UPDATE river
SET props = props - 'vodinfo_sensor' ||
            jsonb_build_object('vodinfo_sensors', jsonb_build_array(props -> 'vodinfo_sensor'))
WHERE props ? 'vodinfo_sensor';

CREATE TABLE level_sensor (
    id CHARACTER VARYING(16) PRIMARY KEY,
    l0 INT NOT NULL,
    l1 INT NOT NULL CHECK ( l1>=l0 ),
    l2 INT NOT NULL CHECK ( l2>=l1 ),
    l3 INT NOT NULL CHECK ( l3>=l2 )
);
INSERT INTO level_sensor(id,l0,l1,l2,l3) (SELECT DISTINCT sensor_id, 0, 0, 0, 0 FROM level);
ALTER TABLE level ADD CONSTRAINT level_sensor_id_fk FOREIGN KEY (sensor_id) REFERENCES level_sensor(id);

-- #######################################################

CREATE TABLE camp
(
    id              BIGINT PRIMARY KEY DEFAULT nextval('id_gen'),
    osm_id          BIGINT UNIQUE,
    title           VARCHAR(512) NOT NULL,
    description     TEXT         NOT NULL,
    point           GEOMETRY     NOT NULL UNIQUE,
    num_tent_places SMALLINT,

    CONSTRAINT point_is_point CHECK (GeometryType(point) = 'POINT'),
    CONSTRAINT num_tent_places_positive CHECK ( num_tent_places > 0 )
);

CREATE TABLE rate
(
    ref_id        BIGINT       NOT NULL,
    rate          SMALLINT     NOT NULL,
    auth_provider VARCHAR(16)  NOT NULL,
    ext_id        VARCHAR(128) NOT NULL,
    PRIMARY KEY (ref_id, auth_provider, ext_id)
);

CREATE TABLE camp_rate
(
    FOREIGN KEY (ref_id) REFERENCES camp (id)
) INHERITS (rate);

CREATE TABLE photo
(
    id          BIGINT PRIMARY KEY DEFAULT nextval('id_gen'),
    ref_id      BIGINT  NOT NULL,
    url         VARCHAR(512),
    preview_url VARCHAR(512),
    main_image  boolean NOT NULL   DEFAULT false
);

CREATE TABLE camp_photo
(
    FOREIGN KEY (ref_id) REFERENCES camp (id)
) INHERITS (photo);

CREATE TABLE camp_river_ref
(
    camp_id  BIGINT NOT NULL REFERENCES camp (id),
    river_id BIGINT NOT NULL REFERENCES river (id),
    PRIMARY KEY (camp_id, river_id)
);