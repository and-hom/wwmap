CREATE TABLE voyage_report(
  id                BIGINT PRIMARY KEY                DEFAULT nextval('id_gen'),
  remote_id         CHARACTER VARYING(32) NOT NULL UNIQUE,
  source            CHARACTER VARYING(16) NOT NULL,
  url               CHARACTER VARYING(512) NOT NULL,
  date_published    timestamp,
  date_modified     timestamp
);

CREATE TABLE voyage_report_river (
    voyage_report_id    BIGINT REFERENCES voyage_report(id),
    river_id            BIGINT REFERENCES river(id),
    PRIMARY KEY(voyage_report_id, river_id)
)