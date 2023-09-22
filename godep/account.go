package godep

import (
	"context"
	"net/http"
)

// Limit corresponds to the Apple DEP API "Limit" structure.
// See https://developer.apple.com/documentation/devicemanagement/limit
type Limit struct {
	Default int `json:"default"`
	Maximum int `json:"maximum"`
}

// URL corresponds to the Apple DEP API "Url" structure.
// See https://developer.apple.com/documentation/devicemanagement/url
type URL struct {
	HTTPMethod []string `json:"http_method"`
	Limit      Limit    `json:"limit"`
	URI        string   `json:"uri"`
}

// AccountDetail corresponds to the Apple DEP API "AccountDetail" structure.
// See https://developer.apple.com/documentation/devicemanagement/accountdetail
type AccountDetail struct {
	AdminID       string `json:"admin_id"`
	FacilitatorID string `json:"facilitator_id,omitempty"`
	OrgAddress    string `json:"org_address"`
	OrgEMail      string `json:"org_email"`
	OrgID         string `json:"org_id"`
	OrgIDHash     string `json:"org_id_hash"`
	OrgName       string `json:"org_name"`
	OrgPhone      string `json:"org_phone"`
	OrgType       string `json:"org_type"`
	OrgVersion    string `json:"org_version"`
	ServerName    string `json:"server_name"`
	ServerUUID    string `json:"server_uuid"`
	URLs          []URL  `json:"urls"`
}

// AccountDetail uses the Apple "Obtain the details for your account" API
// endpoint to get the details about the DEP account and MDM server.
// See https://developer.apple.com/documentation/devicemanagement/get_account_detail
func (c *Client) AccountDetail(ctx context.Context, name string) (*AccountDetail, error) {
	resp := new(AccountDetail)
	return resp, c.do(ctx, name, http.MethodGet, "/account", nil, resp)
}
