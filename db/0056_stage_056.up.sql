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
    FOREIGN KEY (ref_id) REFERENCES camp (id) ON DELETE CASCADE
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
    FOREIGN KEY (ref_id) REFERENCES camp (id) ON DELETE CASCADE
) INHERITS (photo);

CREATE TABLE camp_river_ref
(
    camp_id  BIGINT NOT NULL REFERENCES camp (id) ON DELETE CASCADE,
    river_id BIGINT NOT NULL REFERENCES river (id) ON DELETE CASCADE,
    PRIMARY KEY (camp_id, river_id)
);