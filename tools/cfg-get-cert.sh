#!/bin/sh

CN="${1:-depserver}"
VALIDITY_DAYS="${2:-1}"
URL="${BASE_URL}/v1/tokenpki/${DEP_NAME}?cn=$CN&validity_days=$VALIDITY_DAYS"

curl \
	$CURL_OPTS \
	-u "depserver:$APIKEY" \
	"$URL"
