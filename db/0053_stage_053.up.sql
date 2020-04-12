CREATE SCHEMA cron;

CREATE TABLE cron.job
(
    id      BIGINT PRIMARY KEY    DEFAULT nextval('id_gen'),
    title   VARCHAR(512) NOT NULL,
    expr    VARCHAR(128),
    enabled BOOLEAN      NOT NULL DEFAULT FALSE,
    command CHARACTER VARYING(64),
    args TEXT
);
CREATE TABLE cron.execution
(
    id     BIGINT PRIMARY KEY                       DEFAULT nextval('id_gen'),
    job_id BIGINT REFERENCES cron.job (id) NOT NULL,
    start  TIMESTAMP WITH TIME ZONE     NOT NULL DEFAULT now(),
    "end"  TIMESTAMP WITH TIME ZONE CHECK ( "end" IS NULL OR "end" >= start ),
    status VARCHAR(8)                      NOT NULL DEFAULT 'NEW'
);
