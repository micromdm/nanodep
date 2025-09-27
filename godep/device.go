package godep

import (
	"context"
	"net/http"
)

type syncCfg struct {
	cursor string
	limit  int
}

type DeviceRequestOption func(*syncCfg)

// WithCursor includes a cursor in the fetch or sync request. The initial
// fetch request should omit this option.
func WithCursor(cursor string) DeviceRequestOption {
	return func(d *syncCfg) {
		d.cursor = cursor
	}
}

// WithCursor includes a device limit in the fetch or sync request.
// Per Apple the API has a default of 100 and a maximum of 1000.
func WithLimit(limit int) DeviceRequestOption {
	return func(d *syncCfg) {
		d.limit = limit
	}
}

// FetchDevices uses the Apple "Get a List of Devices" API endpoint to retrieve
// a list of all devices corresponding to this configured DEP server (DEP name).
// The name parameter specifies the configured DEP name to use.
// You should provide a cursor that is returned from previous FetchDevices
// call responses on any subsequent calls.
// See https://developer.apple.com/documentation/devicemanagement/get_a_list_of_devices
func (c *Client) FetchDevices(ctx context.Context, name string, opts ...DeviceRequestOption) (*FetchDeviceResponseJson, error) {
	req := new(FetchDeviceRequestJson)
	cfg := new(syncCfg)
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.limit > 0 {
		req.Limit = cfg.limit
	}
	if cfg.cursor != "" {
		req.Cursor = &cfg.cursor
	}
	resp := new(FetchDeviceResponseJson)
	return resp, c.Do(ctx, name, http.MethodPost, "/server/devices", req, resp)
}

// SyncDevices uses the Apple "Sync the List of Devices" API endpoint to get
// updates about the list of devices corresponding to this configured DEP
// server (DEP name).
// The name parameter specifies the configured DEP name to use.
// You should provide a cursor that is returned from previous FetchDevices or
// SyncDevices call responses.
// See https://developer.apple.com/documentation/devicemanagement/sync_the_list_of_devices
func (c *Client) SyncDevices(ctx context.Context, name string, opts ...DeviceRequestOption) (*FetchDeviceResponseJson, error) {
	req := new(SyncDeviceRequestJson)
	cfg := new(syncCfg)
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.limit > 0 {
		req.Limit = cfg.limit
	}
	if cfg.cursor != "" {
		req.Cursor = cfg.cursor
	}
	resp := new(FetchDeviceResponseJson)
	return resp, c.Do(ctx, name, http.MethodPost, "/devices/sync", req, resp)
}

// IsCursorExhausted returns true if err is a DEP "exhausted cursor" error.
func IsCursorExhausted(err error) bool {
	return httpErrorContains(err, http.StatusBadRequest, "EXHAUSTED_CURSOR")
}

// IsCursorInvalid returns true if err is a DEP "invalid cursor" error.
func IsCursorInvalid(err error) bool {
	return httpErrorContains(err, http.StatusBadRequest, "INVALID_CURSOR")
}

// IsCursorExpired returns true if err is a DEP "expired cursor" error.
// Per Apple this indicates the cursor is older than 7 days.
func IsCursorExpired(err error) bool {
	return httpErrorContains(err, http.StatusBadRequest, "EXPIRED_CURSOR")
}

// DeviceDetails uses the Apple "Get Device Details" API endpoint to get the
// details on a set of devices.
// See https://developer.apple.com/documentation/devicemanagement/get_device_details
func (c *Client) DeviceDetails(ctx context.Context, name string, serials ...string) (*DeviceListResponseJson, error) {
	req := &DeviceListRequestJson{Devices: serials}
	resp := new(DeviceListResponseJson)
	return resp, c.Do(ctx, name, http.MethodPost, "/devices", req, resp)
}

// DisownDevices uses the Apple "Disown Devices" API endpoint to disclaim
// ownership of device serial numbers.
// WARNING: This will permanantly remove devices from the ABM/ASM/ABE instance.
// Use with caution.
// See https://developer.apple.com/documentation/devicemanagement/disown_devices
func (c *Client) DisownDevices(ctx context.Context, name string, serials ...string) (*DeviceStatusResponseJson, error) {
	req := &DeviceListRequestJson{Devices: serials}
	resp := new(DeviceStatusResponseJson)
	return resp, c.Do(ctx, name, http.MethodPost, "/devices/disown", req, resp)
}
