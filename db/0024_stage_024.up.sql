CREATE TABLE referer(
    id          BIGINT PRIMARY KEY DEFAULT nextval('id_gen'),
    host        CHARACTER VARYING(128) NOT NULL,
    schema      CHARACTER VARYING(8) NOT NULL,
    base_url         CHARACTER VARYING(512) NOT NULL,
    page_url         CHARACTER VARYING(512) NOT NULL,
    last_access TIMESTAMP  NOT NULL DEFAUlT now()
);
CREATE UNIQUE INDEX ON referer(host);