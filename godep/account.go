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
