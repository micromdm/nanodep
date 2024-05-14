package godep

import (
	"context"
	"net/http"
)

// SeedBuildToken corresponds to the Apple DEP API "SeedBuildToken" structure.
// See https://developer.apple.com/documentation/devicemanagement/seedbuildtoken
type SeedBuildToken struct {
	Token string `json:"token"`
	Title string `json:"title"`
	OS    string `json:"os"`
}

// GetSeedBuildTokenResponse corresponds to the Apple DEP API "GetSeedBuildTokenResponse" structure.
// See https://developer.apple.com/documentation/devicemanagement/getseedbuildtokenresponse
type GetSeedBuildTokenResponse struct {
	BetaEnrollmentTokens []SeedBuildToken `json:"betaEnrollmentTokens,omitempty"`
	SeedBuildTokens      []SeedBuildToken `json:"seedBuildTokens,omitempty"`
}

// OSBetaEnrollmentTokens uses the Apple "Get Beta Enrollment Tokens" API endpoint to fetch the
// OS beta enrollment tokens. These are for later use during ADE
// enrollment of devices to force enrollment into beta software enrollment.
// See https://developer.apple.com/documentation/devicemanagement/get_beta_enrollment_tokens
func (c *Client) OSBetaEnrollmentTokens(ctx context.Context, name string) (*OSBetaEnrollmentTokensResponse, error) {
	resp := new(OSBetaEnrollmentTokens)
	return resp, c.do(ctx, name, http.MethodGet, "/os-beta-enrollment/tokens", nil, resp)
}
