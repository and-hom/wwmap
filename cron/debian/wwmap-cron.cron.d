*/1 * * * * wwmap wwmap-notifier
0 0 * * 7 wwmap wwmap-backup
0 1 * * * wwmap wwmap-catalog-sync -source pdf
0 0 * * * wwmap wwmap-catalog-sync -source tlib
# 0 0 * 5 * wwmap wwmap-catalog-sync -source libru # Site is dead. Nothing will be changed. Call manually.
0 0 * * 6 wwmap wwmap-catalog-sync -source huskytm
0 0 1 * * wwmap wwmap-db-clean
0 23 * * * wwmap wwmap-spot-sort
