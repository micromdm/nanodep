package pgsql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/micromdm/nanodep/godep"
	"github.com/micromdm/nanodep/storage"
	"github.com/micromdm/nanodep/storage/pgsql/sqlc"
)

// StoreDevices upserts devices for depName in slice order within a single
// transaction. NULL incoming fields preserve stored values via COALESCE
// in the upsert (see devices.sql); they never clear a column.
func (s *PSQLStorage) StoreDevices(ctx context.Context, depName string, devices []godep.Device) error {
	if len(devices) < 1 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	q := s.q.WithTx(tx)
	now := time.Now().UTC().UnixMicro()
	for _, device := range devices {
		imei, err := toRawJSON(device.IMEI)
		if err != nil {
			return fmt.Errorf("marshal imei: %w", err)
		}
		meid, err := toRawJSON(device.MEID)
		if err != nil {
			return fmt.Errorf("marshal meid: %w", err)
		}
		err = q.UpsertDevice(ctx, sqlc.UpsertDeviceParams{
			SerialNumber:          device.SerialNumber,
			DepName:               depName,
			Model:                 device.Model,
			AssetTag:              toNullString(device.AssetTag),
			BluetoothMacAddress:   toNullString(device.BluetoothMACAddress),
			Color:                 toNullString(device.Color),
			Description:           toNullString(device.Description),
			DeviceAssignedBy:      toNullString(device.DeviceAssignedBy),
			DeviceAssignedDate:    toNullMicro(device.DeviceAssignedDate),
			DeviceFamily:          toNullString(device.DeviceFamily),
			Eid:                   toNullString(device.EID),
			EthernetMacAddress:    toNullString(device.EthernetMACAddress),
			Imei:                  imei,
			IsReplacementDevice:   toNullBool(device.IsReplacementDevice),
			MdmMigrationDeadline:  toNullMicro(device.MdmMigrationDeadline),
			Meid:                  meid,
			OpDate:                toNullMicro(device.OpDate),
			OpType:                toNullString(device.OpType),
			Os:                    toNullString(device.OS),
			ProfileAssignTime:     toNullMicro(device.ProfileAssignTime),
			ProfilePushTime:       toNullMicro(device.ProfilePushTime),
			ProfileStatus:         toNullString(device.ProfileStatus),
			ProfileUuid:           toNullString(device.ProfileUUID),
			ReleasedByReplacement: toNullBool(device.ReleasedByReplacement),
			ResponseStatus:        toNullString(device.ResponseStatus),
			WifiMacAddress:        toNullString(device.WiFiMACAddress),
			CreatedAt:             now,
			UpdatedAt:             now,
		})
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// QueryDevices queries and returns stored devices.
func (s *PSQLStorage) QueryDevices(ctx context.Context, req *storage.DevicesQueryRequest) (*storage.DevicesQueryResult, error) {
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

	var depNames, serials []string
	if req != nil && req.Filter != nil {
		depNames = req.Filter.DEPNames
		serials = req.Filter.SerialNumbers
	}

	var rows []sqlc.DepDevice
	switch {
	case len(depNames) > 0 && len(serials) > 0:
		rows, err = s.q.ListDevicesByDEPAndSerial(ctx, sqlc.ListDevicesByDEPAndSerialParams{
			DepNames: depNames,
			Serials:  serials,
			Limit:    int32(limit),
			Offset:   int32(offset),
		})
	case len(depNames) > 0:
		rows, err = s.q.ListDevicesByDEP(ctx, sqlc.ListDevicesByDEPParams{
			DepNames: depNames,
			Limit:    int32(limit),
			Offset:   int32(offset),
		})
	case len(serials) > 0:
		rows, err = s.q.ListDevicesBySerial(ctx, sqlc.ListDevicesBySerialParams{
			Serials: serials,
			Limit:   int32(limit),
			Offset:  int32(offset),
		})
	default:
		rows, err = s.q.ListDevices(ctx, sqlc.ListDevicesParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
	}
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}

	ret := &storage.DevicesQueryResult{
		Devices: make([]storage.StoredDevice, 0, len(rows)),
	}
	for _, row := range rows {
		stored, err := depDeviceToStored(row)
		if err != nil {
			return nil, err
		}
		ret.Devices = append(ret.Devices, stored)
	}
	return ret, nil
}

// DeleteDevices deletes the devices with the given serial numbers for
// the single DEP name depName. An empty serials slice is a no-op.
func (s *PSQLStorage) DeleteDevices(ctx context.Context, depName string, serials []string) error {
	if len(serials) < 1 {
		return nil
	}
	return s.q.DeleteDevicesByDEPAndSerials(ctx, sqlc.DeleteDevicesByDEPAndSerialsParams{
		DepName: depName,
		Serials: serials,
	})
}

// depDeviceToStored maps a sqlc row to a storage.StoredDevice.
func depDeviceToStored(row sqlc.DepDevice) (storage.StoredDevice, error) {
	imei, err := fromRawJSON(row.Imei)
	if err != nil {
		return storage.StoredDevice{}, fmt.Errorf("unmarshal imei: %w", err)
	}
	meid, err := fromRawJSON(row.Meid)
	if err != nil {
		return storage.StoredDevice{}, fmt.Errorf("unmarshal meid: %w", err)
	}
	return storage.StoredDevice{
		DEPName: row.DepName,
		Device: godep.Device{
			SerialNumber:          row.SerialNumber,
			Model:                 row.Model,
			AssetTag:              fromNullString(row.AssetTag),
			BluetoothMACAddress:   fromNullString(row.BluetoothMacAddress),
			Color:                 fromNullString(row.Color),
			Description:           fromNullString(row.Description),
			DeviceAssignedBy:      fromNullString(row.DeviceAssignedBy),
			DeviceAssignedDate:    fromNullMicro(row.DeviceAssignedDate),
			DeviceFamily:          fromNullStringTyped[godep.DeviceDeviceFamily](row.DeviceFamily),
			EID:                   fromNullString(row.Eid),
			EthernetMACAddress:    fromNullString(row.EthernetMacAddress),
			IMEI:                  imei,
			IsReplacementDevice:   fromNullBool(row.IsReplacementDevice),
			MdmMigrationDeadline:  fromNullMicro(row.MdmMigrationDeadline),
			MEID:                  meid,
			OpDate:                fromNullMicro(row.OpDate),
			OpType:                fromNullStringTyped[godep.DeviceOpType](row.OpType),
			OS:                    fromNullStringTyped[godep.DeviceOS](row.Os),
			ProfileAssignTime:     fromNullMicro(row.ProfileAssignTime),
			ProfilePushTime:       fromNullMicro(row.ProfilePushTime),
			ProfileStatus:         fromNullStringTyped[godep.DeviceProfileStatus](row.ProfileStatus),
			ProfileUUID:           fromNullString(row.ProfileUuid),
			ReleasedByReplacement: fromNullBool(row.ReleasedByReplacement),
			ResponseStatus:        fromNullString(row.ResponseStatus),
			WiFiMACAddress:        fromNullString(row.WifiMacAddress),
		},
	}, nil
}

// toNullString converts a *string (or ~string) to sql.NullString.
func toNullString[T ~string](p *T) sql.NullString {
	if p == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: string(*p), Valid: true}
}

// fromNullString converts sql.NullString to *string.
func fromNullString(n sql.NullString) *string {
	if !n.Valid {
		return nil
	}
	s := n.String
	return &s
}

// fromNullStringTyped converts sql.NullString to a *T for Apple string enums.
func fromNullStringTyped[T ~string](n sql.NullString) *T {
	if !n.Valid {
		return nil
	}
	v := T(n.String)
	return &v
}

// toNullMicro converts *time.Time to Unix microseconds, invalid when nil/zero.
func toNullMicro(t *time.Time) sql.NullInt64 {
	if t == nil || t.IsZero() {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: t.UnixMicro(), Valid: true}
}

// fromNullMicro converts Unix microseconds to *time.Time (UTC).
func fromNullMicro(n sql.NullInt64) *time.Time {
	if !n.Valid {
		return nil
	}
	t := time.UnixMicro(n.Int64).UTC()
	return &t
}

// toNullBool converts *bool to sql.NullBool.
func toNullBool(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{}
	}
	return sql.NullBool{Bool: *b, Valid: true}
}

// fromNullBool converts sql.NullBool to *bool.
func fromNullBool(n sql.NullBool) *bool {
	if !n.Valid {
		return nil
	}
	b := n.Bool
	return &b
}

// toRawJSON marshals a string slice to JSON, nil when the slice is nil.
// Per project convention an empty-non-nil slice is normalized to NULL;
// document any change to this choice in a code comment here.
func toRawJSON(v []string) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}

// fromRawJSON unmarshals a JSON array column, nil when NULL/empty.
func fromRawJSON(raw []byte) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var out []string
	return out, json.Unmarshal(raw, &out)
}
