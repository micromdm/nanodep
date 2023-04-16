#!/bin/sh

# See https://developer.apple.com/documentation/devicemanagement/assign_a_profile

DEP_ENDPOINT=/profile/devices
URL="${BASE_URL}/proxy/${DEP_NAME}${DEP_ENDPOINT}"

PROFILE_UUID="$1"
shift

jq -n --arg profile_uuid "$PROFILE_UUID" --arg devices "$*" '{profile_uuid: $profile_uuid, devices: ($devices|split(" "))}' \
	| curl \
		$CURL_OPTS \
		-u "depserver:$APIKEY" \
		-X POST \
		-H 'Content-type: application/json;charset=UTF8' \
		--data-binary @- \
		-A "nanodep-tools/0" \
		"$URL"
