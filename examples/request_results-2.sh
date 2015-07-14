#!/bin/bash

dataFile="$(dirname $0)/data_$(date +%s).json"

cat > $dataFile <<EOT
{
	"type": "finished",
	"min":  0,
	"max":  10
}
EOT

curl -i \
	--request POST \
	--data @$dataFile \
	--header "Content-Type:application/json" \
	--url http://localhost:8080/results

rm -f $dataFile
