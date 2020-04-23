ALTER TABLE cron.execution DROP COLUMN manual;

DELETE from cron.job WHERE command='cron-clean' AND args='120' AND expr='0 8 * * 2';