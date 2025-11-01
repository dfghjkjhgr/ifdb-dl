rm -r ifdb-dl-tool
rm ifdb-dl.zip

set -e
go build

# Makes our lives easier


mkdir ifdb-dl-tool 

cp ./ifdb-dl ifdb-dl-tool
cp ./config.xml ifdb-dl-tool
cp ./menu.json ifdb-dl-tool
echo "
#!/bin/sh

/mnt/us/extensions/kterm/bin/kterm -e \"bash /mnt/us/extensions/ifdb-dl/ifdb-dl\" -k 1 -o U -s 7" > ifdb-dl-tool/run.sh
zip -r ifdb-dl.zip ifdb-dl-tool
