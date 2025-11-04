#!/usr/bin/sh
rm -r ifdb-dl-tool
rm ifdb-dl.zip

set -e
GOOS=linux GOARCH=arm GOARM=5 go build

# Makes our lives easier


mkdir ifdb-dl-tool 

cp ./ifdb-dl ifdb-dl-tool
cp ./config.xml ifdb-dl-tool
cp ./menu.json ifdb-dl-tool
zip -r ifdb-dl.zip ifdb-dl-tool
rm -r ifdb-dl-tool/
