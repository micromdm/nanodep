#!/bin/sh

# Delete locally-persisted synced devices (stored by depsyncer) for a DEP name.
# See the "Devices delete" section in docs/operations-guide.md

if [ -z "$DEP_NAME" ]; then
	echo "DEP_NAME environment variable required" >&2
	exit 1
fi

SERIAL_PARAMS=""
for SERIAL in "$@"; do
	SERIAL_PARAMS="${SERIAL_PARAMS}serial=${SERIAL}&"
done

URL="${BASE_URL}/v1/devices?dep_name=${DEP_NAME}&${SERIAL_PARAMS}"

curl \
	$CURL_OPTS \
	-X DELETE \
	-u "depserver:$APIKEY" \
	-A "nanodep-tools/0" \
	"$URL"
