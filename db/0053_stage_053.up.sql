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

INSERT INTO cron.job (title, expr, enabled, command, args) VALUES ('Import TLib', '0 0 * * *', true, 'catalog-sync', '-source tlib -stage sync-reports');
INSERT INTO cron.job (title, expr, enabled, command, args) VALUES ('Import Risk.ru', '0 2 * * 5', true, 'catalog-sync', '-source riskru -stage sync-reports');
INSERT INTO cron.job (title, expr, enabled, command, args) VALUES ('Import Lib.ru', '0 0 * 5 *', false, 'catalog-sync', '-source libru -stage sync-reports');
INSERT INTO cron.job (title, expr, enabled, command, args) VALUES ('Import Skitalets', '0 0 * * 5', true, 'catalog-sync', '-source skitalets -stage sync-reports');
INSERT INTO cron.job (title, expr, enabled, command, args) VALUES ('Import Huskytm', '0 0 * * 6', true, 'catalog-sync', '-source huskytm -stage sync-reports');
INSERT INTO cron.job (title, expr, enabled, command, args) VALUES ('DB Clean', '0 0 1 * *', true, 'db-clean', '');
INSERT INTO cron.job (title, expr, enabled, command, args) VALUES ('Meteo', '0 */6 * * *', true, 'meteo', '');
INSERT INTO cron.job (title, expr, enabled, command, args) VALUES ('Vodinfo', '0 6-22/4 * * *', true, 'vodinfo-eye', '');
INSERT INTO cron.job (title, expr, enabled, command, args) VALUES ('PDF generate', '0 1 * * *', true, 'catalog-sync', '-source pdf');
INSERT INTO cron.job (title, expr, enabled, command, args) VALUES ('Backup', '0 0 * * 0', true, 'backup', '');
INSERT INTO cron.job (title, expr, enabled, command, args) VALUES ('Data changed notifications', '0 23 * * *', true, 'log-notifications', '');
INSERT INTO cron.job (title, expr, enabled, command, args) VALUES ('Notifier', '*/1 * * * *', true, 'notifier', '');
