CREATE TABLE voyage_report(
  id                BIGINT PRIMARY KEY                DEFAULT nextval('id_gen'),
  title             CHARACTER VARYING(1024) NOT NULL UNIQUE,
  source            CHARACTER VARYING(16) NOT NULL,
  remote_id         CHARACTER VARYING(32) NOT NULL,
  url               CHARACTER VARYING(512) NOT NULL,
  date_published    timestamp,
  date_modified     timestamp
);

CREATE UNIQUE INDEX voyage_report_remote_unique ON voyage_report(source, remote_id);

CREATE TABLE voyage_report_river (
    voyage_report_id    BIGINT NOT NULL REFERENCES voyage_report(id),
    river_id            BIGINT NOT NULL REFERENCES river(id),
    PRIMARY KEY(voyage_report_id, river_id)
);

CREATE TABLE image(
  id                    BIGINT PRIMARY KEY                DEFAULT nextval('id_gen'),
  white_water_rapid_id  BIGINT NOT NULL REFERENCES white_water_rapid(id),
  source                CHARACTER VARYING(16) NOT NULL,
  remote_id             CHARACTER VARYING(32) NOT NULL,
  url                   CHARACTER VARYING(512) NOT NULL,
  preview_url           CHARACTER VARYING(512) NOT NULL,
  date_published        timestamp
);

CREATE INDEX image_white_water_rapid_id_idx ON image(white_water_rapid_id);