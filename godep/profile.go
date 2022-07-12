package godep

import (
	"context"
	"net/http"
)

// ProfileResponse corresponds to the Apple DEP API "AssignProfileResponse" structure.
// See https://developer.apple.com/documentation/devicemanagement/assignprofileresponse
type ProfileResponse struct {
	ProfileUUID string            `json:"profile_uuid"`
	Devices     map[string]string `json:"devices"`
}

// AssignProfiles uses the Apple "Assign a profile to a list of devices" API
// endpoint to assign a DEP profile UUID to a list of serial numbers.
// The name parameter specifies the configured DEP name to use.
// Note we use HTTP PUT for compatibility despite modern documentation
// listing HTTP POST for this endpoint.
// See https://developer.apple.com/documentation/devicemanagement/assign_a_profile
func (c *Client) AssignProfile(ctx context.Context, name, uuid string, serials ...string) (*ProfileResponse, error) {
	req := &struct {
		ProfileUUID string   `json:"profile_uuid"`
		Devices     []string `json:"devices"`
	}{
		ProfileUUID: uuid,
		Devices:     serials,
	}
	resp := new(ProfileResponse)
	// historically this has been an HTTP PUT and the DEP simulator depsim
	// requires this. however modern Apple documentation says this is a POST
	// now. we still use PUT here for compatibility.
	return resp, c.do(ctx, name, http.MethodPut, "/profile/devices", req, resp)
}
