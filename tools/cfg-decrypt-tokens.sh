#!/bin/sh

if [ "x$1" = "x" ]; then
	echo 'filename of token not provided'
	exit 1
fi

URL="${BASE_URL}/v1/tokenpki/${DEP_NAME}?force=$2"

curl \
	$CURL_OPTS \
	-u "depserver:$APIKEY" \
	-T "$1" \
	"$URL"
