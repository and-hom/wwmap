#!/bin/bash
# set -ex

for lat in {37..99}
do
    for lon in {026..180}
    do
        ZIP_FILE="N${lat}E${lon}.SRTMGL1.hgt.zip"
        FILE="N${lat}E${lon}.hgt"
        echo "Downloading $ZIP_FILE"
        if [[ ! -f "$ZIP_FILE" && ! -f "$FILE" ]]
        then
            echo "File $FILE not exists: downloading!"
            wget --header='Cookie: DATA=YT0a8hCZHmrWDbGRp1qTVQAAAKU' \
                -O "$ZIP_FILE" \
                -q \
                "http://e4ftl01.cr.usgs.gov/MEASURES/SRTMGL1.003/2000.02.11/N${lat}E${lon}.SRTMGL1.hgt.zip"
        else
            echo "Already exists"
        fi
    done
done
