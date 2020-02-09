## Packages and apps

* **api** - js api to add whitewater objects to yandex map (see INTEGRATION_ru.md)
* **backend** - map backend provides river and whitewater information from database
* **config** - package installs common configuration file of project
* **cron** - crontab file and utilities called periodically:
    * **notifier** - sends reports to email
    * **catalog-sync** - synchronize database with remote reports and catalogs
    * **backup** - performs backups to yandex disk
    * **spot-sort** - spot ordering tool. Calculates order index for each spot relying on the rivar track(s)
* **data** - utilities for OSM xml parsing
* **db** - database migrations package
* **frontend** - wwmap site and backoffice frontend
* **lib** - not a package - common go sources
* **t-cache** - app for caching of slow tiles

## Packaging

Run ``debuild -us -uc`` in directory containing ``debian`` folder

## Installation
1. Install postgres and postgis
2. Create database user ``wwmap`` and db ``wwmap`` owned by created user. Enable postgis extension on db ``wwmap``:
```
CREATE ROLE wwmap_group;
CREATE ROLE wwmap LOGIN
  NOSUPERUSER INHERIT NOCREATEDB NOCREATEROLE NOREPLICATION;
CREATE DATABASE wwmap WITH OWNER = wwmap_group ENCODING = 'UTF8';
GRANT wwmap_group TO wwmap;
\password wwmap
\c wwmap
CREATE EXTENSION postgis;
```
3. Install ``wwmap-config`` and change ``WWMAP_POSTGRES_PASSWORD`` with real password of ``wwmap`` user in file ``/etc/wwmap/config.yaml``
4. Install ``wwmap-db`` and run ``wwmap-db-upgrade up``
5. Install ``wwmap-backend`` and run ``service wwmap-backend start``
5. Install ``wwmap-t-cache`` and run ``service wwmap-t-cache start``
6. Install ``nginx`` and remove all from sites-enabled.
7. Install ``wwmap-frontend``, ``wwmap-api`` and run ``service nginx restart``
8. In file ``/etc/wwmap/config.yaml`` set real yandex disk credentials in the ``backup`` section. Install ``pyyaml`` python package:
```
LC_ALL='ru_RU.UTF-8' pip install pyyaml
```
9. Install ``wwmap-cron``

## Integration with any site
See INTEGRATION_ru.md