package storage

import (
	"context"

	"github.com/micromdm/nanodep/godep"
)

// StoredDevice is a godep.Device plus the DEP name it synced on.
// Rows are keyed by the compound key (dep_name, serial_number): the
// same serial may exist under multiple DEP names independently.
type StoredDevice struct {
	DEPName string `json:"dep_name"`
	godep.Device
}

// DevicesQueryFilter is the filter parameters for querying devices.
type DevicesQueryFilter struct {
	// DEPNames specifies which DEP names to query for.
	DEPNames []string `json:"dep_names,omitempty"`

	// SerialNumbers specifies which device serial numbers to query for.
	SerialNumbers []string `json:"serial_numbers,omitempty"`
}

// DevicesQueryRequest is the parameters for querying devices.
type DevicesQueryRequest struct {
	Filter     *DevicesQueryFilter `json:"filter,omitempty"`
	Pagination *Pagination         `json:"pagination,omitempty"`
}

// DevicesQueryResult is the resulting paginated list of devices.
type DevicesQueryResult struct {
	Devices []StoredDevice `json:"devices"`

	PaginationNextCursor
}

// DevicesQuery queries and returns stored devices.
type DevicesQuery interface {
	// QueryDevices queries and returns devices matching req.
	// An empty or nil filter matches all devices.
	QueryDevices(ctx context.Context, req *DevicesQueryRequest) (*DevicesQueryResult, error)
}

// DeviceStorer persists synced devices.
type DeviceStorer interface {
	// StoreDevices upserts devices for depName preserving slice order.
	// A nil field in an incoming device means "unknown" and preserves
	// the stored value; it never clears a column.
	StoreDevices(ctx context.Context, depName string, devices []godep.Device) error
}

// DeviceDeleter deletes stored devices.
type DeviceDeleter interface {
	// DeleteDevices deletes the devices with the given serial numbers
	// for the single DEP name depName. An empty or nil serials slice
	// is a no-op returning nil. Deleting a non-existent device is
	// not an error.
	DeleteDevices(ctx context.Context, depName string, serials []string) error
}
