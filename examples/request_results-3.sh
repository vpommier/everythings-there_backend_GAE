#!/bin/bash

dataFile="$(dirname $0)/data_$(date +%s).json"

cat > $dataFile <<EOT
{
	"jeu":{
		"plaques":[1,2,5,6,7,100],
		"total":100
	},
	"type":"finished"
}
EOT

curl -i \
	--request POST \
	--data @$dataFile \
	--header "Content-Type:application/json" \
	--url http://localhost:8080/results

rm -f $dataFile
