#!/bin/bash
set -e

TABLES="river white_water_rapid report"

function config() {
    python -c 'import yaml; print(yaml.load(open("/etc/wwmap/config.yaml","r"))'$1')'
}

CONN_STR=`config '["db-connection-string"]'`
YA_EMAIL=`config '["backup"]["email"]'`
YA_PASSWORD=`config '["backup"]["password"]'`

# number of backups to be saved
KEEP=15

# dir to backup
DIR=/var/lib/wwmap/backup

NOW=$(date +"%Y-%m-%d")
# DBS="$(psql -U $USER -lt |awk '{ print $1}' |grep -vE '^-|^List|^Name|template[0|1]')"

BACKUPS=`find $DIR -name "wwmap.*.gz" | wc -l | sed 's/\ //g'`
while [ $BACKUPS -ge $KEEP ]
do
  ls -tr1 $DIR/wwmap.*.gz | head -n 1 | xargs rm -f
  BACKUPS=`expr $BACKUPS - 1`
done
FILE=$DIR/wwmap.$NOW-$(date +"%T").sql.gz

pg_dump $CONN_STR `for t in $TABLES; do echo -n ' -t '$t; done` | gzip -c > $FILE
curl -f --user $YA_EMAIL:$YA_PASSWORD -T "{$FILE}" https://webdav.yandex.ru/backup/

exit 0