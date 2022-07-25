#!/bin/sh

URL="${BASE_URL}/v1/tokenpki/${DEP_NAME}?cn=$1&validity_days=$2"

curl \
	$CURL_OPTS \
	-u "depserver:$APIKEY" \
	"$URL"
