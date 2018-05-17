## Packages and apps

* **backend** - map backend provides river and whitewater information from database
* **config** - package installs common configuration file of project
* **cron** - crontab file and utilities called periodically:
    * **notifier** - sends reports to email
    * **backup** - performs backups to yandex disk
* **data** - utilities for OSM xml parsing
* **db** - database migrations package
* **rontend** - map frontend: html, css, js, images
* **lib** - not a package - common go sources
* **t-cache** - app for caching of slow tiles

## Packaging

Run ``debuild -us -uc`` in directory containing ``debian`` folder