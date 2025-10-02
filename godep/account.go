package godep

import (
	"context"
	"net/http"
)

// AccountDetail uses the Apple "Obtain the details for your account" API
// endpoint to get the details about the DEP account and MDM server.
// See https://developer.apple.com/documentation/devicemanagement/get_account_detail
func (c *Client) AccountDetail(ctx context.Context, name string) (*AccountDetailJson, error) {
	resp := new(AccountDetailJson)
	return resp, c.Do(ctx, name, http.MethodGet, "/account", nil, resp)
}

// AssignAccountDrivenEnrollmentProfile uses the Apple "Assign Account-Driven
// Enrollment Service Discovery" API endpoint to assign the account-driven
// enrollment profile for service discovery.
// See https://developer.apple.com/documentation/devicemanagement/assign-account-driven-enrollment-profile
func (c *Client) AssignAccountDrivenEnrollmentProfile(ctx context.Context, name string, mdmServiceDiscoveryURL string) error {
	req := &AccountDrivenEnrollmentProfileRequestJson{
		MdmServiceDiscoveryUrl: mdmServiceDiscoveryURL,
	}
	return c.Do(ctx, name, http.MethodPost, "/account-driven-enrollment/profile", req, nil)
}
