#!/usr/bin/sh
set -e

rm -rf ifdb-dl-tool
rm -f ifdb-dl.zip

GOOS=linux GOARCH=arm GOARM=5 go build

# Makes our lives easier


mkdir ifdb-dl-tool 

cp ./ifdb-dl ifdb-dl-tool
cp ./config.xml ifdb-dl-tool
cp ./menu.json ifdb-dl-tool
zip -r ifdb-dl.zip ifdb-dl-tool
rm -r ifdb-dl-tool/
