package kv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/micromdm/nanodep/godep"
	"github.com/micromdm/nanodep/storage"

	"github.com/micromdm/nanolib/storage/kv"
)

// deviceKey returns the KV key for the compound (depName, serial).
// The length prefix keeps DEP names containing "." unambiguous
// (e.g. "a.b"+"c" vs "a"+"b.c" would otherwise collide on ".").
func deviceKey(depName, serial string) string {
	return fmt.Sprintf("%s%d.%s.%s", keyPfxDevice, len(depName), depName, serial)
}

// coalescePtr returns newV if non-nil, otherwise oldV.
func coalescePtr[T any](newV, oldV *T) *T {
	if newV != nil {
		return newV
	}
	return oldV
}

// mergeDevice overlays incoming onto old, returning the merged result:
// every nil pointer/nil slice field in incoming is filled from old,
// while serial/model/dep_name always come from the incoming side
// (depName supplies the DEP name; within a compound row it always
// matches old). This mirrors the COALESCE policy
// the SQL backends implement in devices.sql.
func mergeDevice(old storage.StoredDevice, incoming godep.Device, depName string) storage.StoredDevice {
	// start from old so that any field we forget to list below
	// conservatively preserves rather than clears stored state
	merged := old
	merged.DEPName = depName
	merged.Device.SerialNumber = incoming.SerialNumber
	merged.Device.Model = incoming.Model
	merged.Device.AssetTag = coalescePtr(incoming.AssetTag, old.Device.AssetTag)
	merged.Device.BluetoothMACAddress = coalescePtr(incoming.BluetoothMACAddress, old.Device.BluetoothMACAddress)
	merged.Device.Color = coalescePtr(incoming.Color, old.Device.Color)
	merged.Device.Description = coalescePtr(incoming.Description, old.Device.Description)
	merged.Device.DeviceAssignedBy = coalescePtr(incoming.DeviceAssignedBy, old.Device.DeviceAssignedBy)
	merged.Device.DeviceAssignedDate = coalescePtr(incoming.DeviceAssignedDate, old.Device.DeviceAssignedDate)
	merged.Device.DeviceFamily = coalescePtr(incoming.DeviceFamily, old.Device.DeviceFamily)
	merged.Device.EID = coalescePtr(incoming.EID, old.Device.EID)
	merged.Device.EthernetMACAddress = coalescePtr(incoming.EthernetMACAddress, old.Device.EthernetMACAddress)
	merged.Device.IsReplacementDevice = coalescePtr(incoming.IsReplacementDevice, old.Device.IsReplacementDevice)
	merged.Device.MdmMigrationDeadline = coalescePtr(incoming.MdmMigrationDeadline, old.Device.MdmMigrationDeadline)
	merged.Device.OpDate = coalescePtr(incoming.OpDate, old.Device.OpDate)
	merged.Device.OpType = coalescePtr(incoming.OpType, old.Device.OpType)
	merged.Device.OS = coalescePtr(incoming.OS, old.Device.OS)
	merged.Device.ProfileAssignTime = coalescePtr(incoming.ProfileAssignTime, old.Device.ProfileAssignTime)
	merged.Device.ProfilePushTime = coalescePtr(incoming.ProfilePushTime, old.Device.ProfilePushTime)
	merged.Device.ProfileStatus = coalescePtr(incoming.ProfileStatus, old.Device.ProfileStatus)
	merged.Device.ProfileUUID = coalescePtr(incoming.ProfileUUID, old.Device.ProfileUUID)
	merged.Device.ReleasedByReplacement = coalescePtr(incoming.ReleasedByReplacement, old.Device.ReleasedByReplacement)
	merged.Device.ResponseStatus = coalescePtr(incoming.ResponseStatus, old.Device.ResponseStatus)
	merged.Device.WiFiMACAddress = coalescePtr(incoming.WiFiMACAddress, old.Device.WiFiMACAddress)
	if incoming.IMEI != nil {
		merged.Device.IMEI = incoming.IMEI
	}
	if incoming.MEID != nil {
		merged.Device.MEID = incoming.MEID
	}
	return merged
}

// StoreDevices upserts devices for depName preserving slice order.
// A nil field in an incoming device preserves the stored value
// (via mergeDevice); it never clears a column.
func (s *KV) StoreDevices(ctx context.Context, depName string, devices []godep.Device) error {
	if len(devices) < 1 {
		return nil
	}
	return kv.PerformCRUDBucketTxn(ctx, s.b, func(ctx context.Context, txn kv.CRUDBucket) error {
		merged := make(map[string][]byte, len(devices))
		for _, device := range devices {
			key := deviceKey(depName, device.SerialNumber)
			stored := storage.StoredDevice{DEPName: depName, Device: device}
			if existing, err := txn.Get(ctx, key); err == nil {
				var old storage.StoredDevice
				if err := json.Unmarshal(existing, &old); err != nil {
					return err
				}
				stored = mergeDevice(old, device, depName)
			} else if !errors.Is(err, kv.ErrKeyNotFound) {
				return err
			}
			mergedJSON, err := json.Marshal(stored)
			if err != nil {
				return err
			}
			merged[key] = mergedJSON
		}
		return kv.SetMap(ctx, txn, merged)
	})
}

// QueryDevices queries and returns stored devices.
// [ErrOnlyOffset] is returned if cursor pagination is attempted.
// A default limit of 100 results is returned, ordered by DEP name
// then serial number.
func (s *KV) QueryDevices(ctx context.Context, req *storage.DevicesQueryRequest) (*storage.DevicesQueryResult, error) {
	var offset, limit int
	var err error
	if req != nil {
		if req.Pagination != nil && req.Pagination.Cursor != nil {
			// cursor method not supported for this backend
			return nil, storage.ErrOnlyOffset
		}
		_, offset, limit, err = req.Pagination.ValidateDefaultOffsetLimit(100)
		if err != nil {
			return nil, err
		}
	}

	var depFilter, serialFilter map[string]struct{}
	if req != nil && req.Filter != nil {
		if len(req.Filter.DEPNames) > 0 {
			depFilter = make(map[string]struct{}, len(req.Filter.DEPNames))
			for _, name := range req.Filter.DEPNames {
				depFilter[name] = struct{}{}
			}
		}
		if len(req.Filter.SerialNumbers) > 0 {
			serialFilter = make(map[string]struct{}, len(req.Filter.SerialNumbers))
			for _, serial := range req.Filter.SerialNumbers {
				serialFilter[serial] = struct{}{}
			}
		}
	}

	var all []storage.StoredDevice
	cancel := make(chan struct{})
	for key := range s.b.KeysPrefix(ctx, keyPfxDevice, cancel) {
		value, err := s.b.Get(ctx, key)
		if err != nil {
			close(cancel)
			return nil, err
		}
		var stored storage.StoredDevice
		if err := json.Unmarshal(value, &stored); err != nil {
			close(cancel)
			return nil, err
		}
		if depFilter != nil {
			if _, ok := depFilter[stored.DEPName]; !ok {
				continue
			}
		}
		if serialFilter != nil {
			if _, ok := serialFilter[stored.Device.SerialNumber]; !ok {
				continue
			}
		}
		all = append(all, stored)
	}

	// sort by DEP name then serial number for stable pagination
	sort.Slice(all, func(i, j int) bool {
		if all[i].DEPName != all[j].DEPName {
			return all[i].DEPName < all[j].DEPName
		}
		return all[i].Device.SerialNumber < all[j].Device.SerialNumber
	})

	ret := &storage.DevicesQueryResult{}
	for i := offset; i < len(all) && len(ret.Devices) < limit; i++ {
		ret.Devices = append(ret.Devices, all[i])
	}

	return ret, nil
}

// DeleteDevices deletes the devices with the given serial numbers for
// the single DEP name depName. An empty serials slice is a no-op.
// Deleting a non-existent device is not an error.
func (s *KV) DeleteDevices(ctx context.Context, depName string, serials []string) error {
	if len(serials) < 1 {
		return nil
	}
	keys := make([]string, 0, len(serials))
	for _, serial := range serials {
		keys = append(keys, deviceKey(depName, serial))
	}
	return kv.PerformCRUDBucketTxn(ctx, s.b, func(ctx context.Context, txn kv.CRUDBucket) error {
		return kv.DeleteSlice(ctx, txn, keys)
	})
}
