*/1 * * * * wwmap wwmap-notifier >> /var/log/wwmap/cron/notifier.log 2>&1
0 0 * * 7 wwmap wwmap-backup >> /var/log/wwmap/cron/backup.log 2>&1
0 1 * * * wwmap wwmap-catalog-sync -source pdf >> /var/log/wwmap/cron/cat-sync-pdf.log 2>&1
0 0 * * * wwmap wwmap-catalog-sync -source tlib >> /var/log/wwmap/cron/cat-sync-tlib.log 2>&1
# 0 0 * 5 * wwmap wwmap-catalog-sync -source libru # Site is dead. Nothing will be changed. Call manually.
0 0 * * 5 wwmap wwmap-catalog-sync -source skitalets >> /var/log/wwmap/cron/cat-sync-skitalets.log 2>&1
0 0 * * 6 wwmap wwmap-catalog-sync -source huskytm  >> /var/log/wwmap/cron/cat-sync-huskytm.log 2>&1
0 0 1 * * wwmap wwmap-db-clean >> /var/log/wwmap/cron/db-clean.log 2>&1
0 23 * * * wwmap wwmap-spot-sort >> /var/log/wwmap/cron/spot-sort.log 2>&1
0 */6 * * * wwmap wwmap-meteo >> /var/log/wwmap/cron/meteo.log 2>&1
0 6-22/4 * * * wwmap wwmap-vodinfo-eye >> /var/log/wwmap/cron/vodinfo-eye.log 2>&1
0 23 * * * wwmap wwmap-river-tracks-bind >> /var/log/wwmap/cron/river-tracks-bind.log 2>&1
