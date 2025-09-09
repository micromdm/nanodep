package godep

import (
	"context"
	"net/http"
)

// OSBetaEnrollmentTokens uses the Apple "Get Beta Enrollment Tokens" API endpoint to fetch the
// OS beta enrollment tokens. These are for later use during ADE
// enrollment of devices to force enrollment into beta software enrollment.
// See https://developer.apple.com/documentation/devicemanagement/get_beta_enrollment_tokens
func (c *Client) OSBetaEnrollmentTokens(ctx context.Context, name string) (*GetSeedBuildTokenResponseJson, error) {
	resp := new(GetSeedBuildTokenResponseJson)
	return resp, c.Do(ctx, name, http.MethodGet, "/os-beta-enrollment/tokens", nil, resp)
}

// IsAppleSeedForITTurnedOff returns true if err indicates your organization doesn't allow beta access.
func IsAppleSeedForITTurnedOff(err error) bool {
	return httpErrorContains(err, http.StatusForbidden, "APPLE_SEED_FOR_IT_TURNED_OFF")
}
