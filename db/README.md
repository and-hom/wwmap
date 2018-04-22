### SQL migrations

Migration files for https://github.com/mattes/migrate

#### Before first usage
1. Perform installation instructions https://github.com/mattes/migrate/blob/master/cli/README.md
2. Install postgres and postgis
3. Create database wwmap owned to user wwmap
4. Add postgis extension to database: ``CREATE EXTENSION postgis;``

#### Upgrade db:
```
migrate -database 'postgres://wwmap:<your secret password>@localhost:5432/wwmap' -source file://. up
```

#### Revert latest
```
migrate -database 'postgres://wwmap:<your secret password>@localhost:5432/wwmap' -source file://. down 1
```

### Simple migrate data to prod:

```
sudo -u postgres pg_dump --data-only --table=waterway wwmap > waterway.pg
scp waterway.pg my-server:
ssh my-server
sudo -u postgres psql wwmap < waterway.pg
```
