CREATE TABLE transfer
(
    id          BIGINT PRIMARY KEY              DEFAULT nextval('id_gen'),
    title       CHARACTER VARYING(255) NOT NULL CHECK ( trim(title) <> '' ),
    stations    JSONB                  NOT NULL DEFAULT '[]',
    description TEXT
);

CREATE TABLE transfer_river
(
    transfer_id BIGINT REFERENCES transfer (id),
    river_id    BIGINT REFERENCES river (id),
    PRIMARY KEY (transfer_id, river_id)
);