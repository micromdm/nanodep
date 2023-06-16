#!/bin/sh

URL="${BASE_URL}/v1/maidjwt/${DEP_NAME}"

curl \
	$CURL_OPTS \
	-u "depserver:$APIKEY" \
	"$URL"
