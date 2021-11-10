#!/bin/bash
set -ex

for i in *.hgt.zip
do
    f=${i%.SRTMGL1.hgt.zip}
    unzipped1=$f.SRTMGL1.hgt
    unzipped2=$f.hgt

	unzip $i
    if [[ -f "$unzipped1" ]]
	then
	    raster2pgsql -a -F $unzipped1 srtm | sudo -u postgres psql wwmap
        rm $unzipped1
	else
	    raster2pgsql -a -F $unzipped2 srtm | sudo -u postgres psql wwmap
        rm $unzipped2
    fi
done
