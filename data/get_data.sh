#!/bin/bash

output_file="data.csv"

curl -o "$output_file" "https://data.cityofnewyork.us/api/views/25th-nujf/rows.csv?accessType=DOWNLOAD"

if [ $? -eq 0 ]; then
    echo "download success, save as $output_file."
else
    echo "download failed."
fi
