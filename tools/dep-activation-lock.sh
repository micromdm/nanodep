#!/bin/sh

# See https://developer.apple.com/documentation/devicemanagement/activation-lock-devices

DEP_ENDPOINT=/device/activationlock
URL="${BASE_URL}/proxy/${DEP_NAME}${DEP_ENDPOINT}"

DEVICE="$1"
ESCROW_KEY="$2"
LOST_MESSAGE="$3"

if [ "x$DEVICE" = "x" ]; then
    echo "device parameter missing"
    echo "$0 <device> [escrow-key] [lost-message]"
    exit 1
fi

# JSON object with device,
# conditionally including escrow_key and lost_message if they are non-empty.
JSON='
{ device: $device }
+ if $escrow_key != "" then { escrow_key: $escrow_key } else {} end
+ if $lost_message != "" then { lost_message: $lost_message } else {} end
'

jq -n \
    --arg device "$DEVICE" \
    --arg escrow_key "$ESCROW_KEY" \
    --arg lost_message "$LOST_MESSAGE" \
    "$JSON" \
	| curl \
		$CURL_OPTS \
		-u "depserver:$APIKEY" \
		-X POST \
		-H 'Content-type: application/json;charset=UTF8' \
		--data-binary @- \
		-A "nanodep-tools/0" \
		"$URL"
