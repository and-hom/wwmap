ALTER TABLE cron.execution ADD COLUMN manual BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE cron.job ADD COLUMN critical BOOLEAN NOT NULL DEFAULT false;

INSERT INTO cron.job (title, expr, enabled, command, args) VALUES ('Cron cleaner', '0 8 * * 2', true, 'cron-clean', '120');