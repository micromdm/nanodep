package godep

import (
	"context"
	"net/http"
)

// AccountDetail uses the Apple "Obtain the details for your account" API
// endpoint to get the details about the DEP account and MDM server.
// See https://developer.apple.com/documentation/devicemanagement/get_account_detail
func (c *Client) AccountDetail(ctx context.Context, name string) (*AccountDetail, error) {
	resp := new(AccountDetail)
	return resp, c.Do(ctx, name, http.MethodGet, "/account", nil, resp)
}

// AssignAccountDrivenEnrollmentProfile uses the Apple "Assign Account-Driven
// Enrollment Service Discovery" API endpoint to assign the account-driven
// enrollment profile for service discovery.
// See https://developer.apple.com/documentation/devicemanagement/assign-account-driven-enrollment-profile
func (c *Client) AssignAccountDrivenEnrollmentProfile(ctx context.Context, name string, mdmServiceDiscoveryURL string) error {
	req := &AccountDrivenEnrollmentProfileRequest{
		MdmServiceDiscoveryURL: mdmServiceDiscoveryURL,
	}
	return c.Do(ctx, name, http.MethodPost, "/account-driven-enrollment/profile", req, nil)
}

// GetAccountDrivenEnrollmentProfile uses the Apple "Fetch Account-Driven
// Enrollment Service Discovery" API endpoint to retrieve the account-driven
// enrollment profile for service discovery.
// See https://developer.apple.com/documentation/devicemanagement/fetch-account-driven-enrollment-profile
//
// Introduced in Device Assignment Services 5.1.
func (c *Client) GetAccountDrivenEnrollmentProfile(ctx context.Context, name string) (*GetAccountDrivenEnrollmentProfileResponse, error) {
	resp := new(GetAccountDrivenEnrollmentProfileResponse)
	return resp, c.Do(ctx, name, http.MethodGet, "/account-driven-enrollment/profile", nil, resp)
}

// RemoveAccountDrivenEnrollmentProfile uses the Apple "Remove Account-Driven
// Enrollment Profile" API endpoint to remove the account-driven enrollment
// profile for service discovery.
// See https://developer.apple.com/documentation/devicemanagement/remove-account-driven-enrollment-profile
//
// Introduced in Device Assignment Services 5.1.
func (c *Client) RemoveAccountDrivenEnrollmentProfile(ctx context.Context, name string) error {
	return c.Do(ctx, name, http.MethodDelete, "/account-driven-enrollment/profile", nil, nil)
}

// IsOrgNotSupported returns true if err is a DEP "org not supported" error.
// Indicates that the associated organization isn’t supported.
func IsOrgNotSupported(err error) bool {
	return httpErrorContains(err, http.StatusBadRequest, "ORG_NOT_SUPPORTED")
}
