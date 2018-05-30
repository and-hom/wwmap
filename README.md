## Packages and apps

* **backend** - map backend provides river and whitewater information from database
* **config** - package installs common configuration file of project
* **cron** - crontab file and utilities called periodically:
    * **notifier** - sends reports to email
    * **backup** - performs backups to yandex disk
* **data** - utilities for OSM xml parsing
* **db** - database migrations package
* **frontend** - map frontend: html, css, js, images
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
7. Install ``wwmap-frontend`` and run ``service nginx restart``
8. In file ``/etc/wwmap/config.yaml`` set real yandex disk credentials in the ``backup`` section. Install ``pyyaml`` python package:
```
LC_ALL='ru_RU.UTF-8' pip install pyyaml
```
9. Install ``wwmap-cron``