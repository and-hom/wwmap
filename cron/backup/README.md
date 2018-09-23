Backup database tables and store at yandex disk.

To restore use commands

```
export DUMP_FILE='your/dump/file/path/here'

tac tables.list | grep -v id_gen | while read -r line; do
    echo "DELETE FROM \"$line\";"
done | sudo -u postgres psql -d wwmap -1  --

pg_restore --data-only --exit-on-error --single-transaction --format=c --dbname wwmap --host 127.0.0.1 -U wwmap -W $DUMP_FILE
```