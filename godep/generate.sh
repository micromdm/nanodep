#!/bin/sh

# using fork github.com/jessepeterson/go-jsonschema (micromdm branch)

go-jsonschema \
	-p godep \
	--tags json \
	--disable-omitzero \
	--only-models \
	--capitalization ID,EID,MAC,MEID,IMEI,WiFi,OS,UUID,URL,URLs,HTTP \
	-t \
	--output schema.go \
	apple-device-services/dep/schemas/*.json

