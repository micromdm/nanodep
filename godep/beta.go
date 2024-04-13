package godep

import (
	"context"
	"net/http"
)

// OSToken corresponds to the Apple DEP API "???" structure.
// See https://developer.apple.com/documentation/devicemanagement/???
type OSToken struct {
	Token string `json:"token"`
	Title string `json:"title"`
	OS    string `json:"os"`
}

// OSBetaEnrollmentTokens corresponds to the Apple DEP API "???" structure.
// See https://developer.apple.com/documentation/devicemanagement/???
type OSBetaEnrollmentTokens struct {
	BetaEnrollmentTokens []OSToken `json:"betaEnrollmentTokens,omitempty"`
	SeedBuildTokens      []OSToken `json:"seedBuildTokens,omitempty"`
}

// OSBetaEnrollmentTokens uses the Apple "???" API endpoint to fetch the
// OS beta enrollment tokens. These are for later use during ADE
// enrollment of devices to force enrollment into beta software enrollment.
// See https://developer.apple.com/documentation/devicemanagement/???
func (c *Client) OSBetaEnrollmentTokens(ctx context.Context, name string) (*OSBetaEnrollmentTokens, error) {
	resp := new(OSBetaEnrollmentTokens)
	return resp, c.do(ctx, name, http.MethodGet, "/os-beta-enrollment/tokens", nil, resp)
}
