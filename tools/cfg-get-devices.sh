#!/bin/sh

# Query locally-persisted synced devices (stored by depsyncer).
# See the "Devices query" section in docs/operations-guide.md

DEP_NAME_PARAM=""
if [ -n "$DEP_NAME" ]; then
	DEP_NAME_PARAM="dep_name=${DEP_NAME}&"
fi

SERIAL_PARAMS=""
for SERIAL in "$@"; do
	SERIAL_PARAMS="${SERIAL_PARAMS}serial=${SERIAL}&"
done

URL="${BASE_URL}/v1/devices?${DEP_NAME_PARAM}${SERIAL_PARAMS}"

curl \
	$CURL_OPTS \
	-u "depserver:$APIKEY" \
	-A "nanodep-tools/0" \
	"$URL"
