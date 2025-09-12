package godep

import (
	"context"
	"net/http"
	"net/url"
)

// AssignProfiles uses the Apple "Assign a profile to a list of devices" API
// endpoint to assign a DEP profile UUID to a list of serial numbers.
// The name parameter specifies the configured DEP name to use.
// Note we use HTTP PUT for compatibility despite modern documentation
// listing HTTP POST for this endpoint.
// See https://developer.apple.com/documentation/devicemanagement/assign_a_profile
func (c *Client) AssignProfile(ctx context.Context, name, uuid string, serials ...string) (*AssignProfileResponseJson, error) {
	req := &ProfileServiceRequestJson{ProfileUuid: &uuid, Devices: serials}
	resp := new(AssignProfileResponseJson)
	// historically this has been an HTTP PUT and the DEP simulator depsim
	// requires this. however modern Apple documentation says this is a POST
	// now. we still use PUT here for compatibility.
	return resp, c.Do(ctx, name, http.MethodPut, "/profile/devices", req, resp)
}

// DefineProfile uses the Apple "Define a Profile" command to attempt to create a profile.
// This service defines a profile with Apple's servers that can then be assigned to specific devices.
// This command provides information about the MDM server that is assigned to manage one or more devices,
// information about the host that the managed devices can pair with, and various attributes that control
// the MDM association behavior of the device.
// See https://developer.apple.com/documentation/devicemanagement/define_a_profile
func (c *Client) DefineProfile(ctx context.Context, name string, profile *ProfileJson) (*DefineProfileResponseJson, error) {
	resp := new(DefineProfileResponseJson)
	return resp, c.Do(ctx, name, http.MethodPost, "/profile", profile, resp)
}

// RemoveProfile uses the Apple "Remove a Profile" API endpoint to "unassign"
// any DEP profile UUID from a list of serial numbers.
// A `profile_uuid` API paramater is listed in the documentation but we do not
// support it (nor does it appear to be used on the server-side).
// The name parameter specifies the configured DEP name to use.
// See https://developer.apple.com/documentation/devicemanagement/remove_a_profile-c2c
func (c *Client) RemoveProfile(ctx context.Context, name string, serials ...string) (*ClearProfileResponseJson, error) {
	req := &ClearProfileRequestJson{Devices: serials}
	resp := new(ClearProfileResponseJson)
	return resp, c.Do(ctx, name, http.MethodDelete, "/profile/devices", req, resp)
}

// GetProfile uses the Apple "Get a Profile" API endpoint to return the
// DEP profile named by the given UUID.
// See https://developer.apple.com/documentation/devicemanagement/get_a_profile
func (c *Client) GetProfile(ctx context.Context, name, uuid string) (*ProfileJson, error) {
	v := make(url.Values)
	v.Set("profile_uuid", uuid)
	resp := new(ProfileJson)
	return resp, c.Do(ctx, name, http.MethodGet, "/profile?"+v.Encode(), nil, resp)
}
