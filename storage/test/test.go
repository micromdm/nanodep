// Package test offers a battery of tests for storage.AllStorage implementations.
package test

import (
	"bytes"
	"context"
	"errors"
	"math/rand"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/micromdm/nanodep/client"
	"github.com/micromdm/nanodep/cryptoutil"
	"github.com/micromdm/nanodep/godep"
	"github.com/micromdm/nanodep/storage"
	"github.com/micromdm/nanodep/tokenpki"
)

// TestWithStorages runs multiple tests with different storage provided by storageFn.
func TestWithStorages(t *testing.T, ctx context.Context, store storage.AllStorage) {
	depName1, depName2 := genRandName(4), genRandName(4)

	t.Run("empty", func(t *testing.T) {
		TestEmpty(t, ctx, depName1, store)
	})

	t.Run("basic-name1", func(t *testing.T) {
		TestWitName(t, ctx, depName1, store)
	})

	t.Run("basic-name2", func(t *testing.T) {
		TestWitName(t, ctx, depName2, store)
	})

	t.Run("query-dep-names", func(t *testing.T) {
		TestQueryDEPNames(t, ctx, store)
	})

	t.Run("devices", func(t *testing.T) {
		TestDevices(t, ctx, depName1, depName2, store)
	})

}

// TestEmpty tests retrieval methods on an empty/missing name.
func TestEmpty(t *testing.T, ctx context.Context, name string, s storage.AllStorage) {
	if _, _, err := s.RetrieveStagingTokenPKI(ctx, name); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("unexpected error: %s", err)
	}

	if _, err := s.RetrieveAuthTokens(ctx, name); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("unexpected error: %s", err)
	}

	config, err := s.RetrieveConfig(ctx, name)
	checkErr(t, err)
	if config != nil {
		t.Fatalf("expected non-existent config: %+v", config)
	}

	// Profile assigner storing and retrieval.
	profileUUID, modTime, err := s.RetrieveAssignerProfile(ctx, name)
	checkErr(t, err)
	if profileUUID != "" {
		t.Fatal("expected empty profileUUID")
	}
	if !modTime.IsZero() {
		t.Fatal("expected zero modTime")
	}
	cursor, err := s.RetrieveCursor(ctx, name)
	checkErr(t, err)
	if cursor != "" {
		t.Fatal("expected empty cursor")
	}
}

func TestWitName(t *testing.T, ctx context.Context, name string, s storage.AllStorage) {
	// PKI storing and retrieval.
	if _, _, err := s.RetrieveStagingTokenPKI(ctx, name); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("unexpected error: %s", err)
	}
	pemCert, pemKey := generatePKI(t, "basicdn", 1)
	err := s.StoreTokenPKI(ctx, name, pemCert, pemKey)
	checkErr(t, err)
	pemCert2, pemKey2, err := s.RetrieveStagingTokenPKI(ctx, name)
	checkErr(t, err)
	if !bytes.Equal(pemCert, pemCert2) {
		t.Fatalf("pem cert mismatch: %s vs. %s", pemCert, pemCert2)
	}
	if !bytes.Equal(pemKey, pemKey2) {
		t.Fatalf("pem key mismatch: %s vs. %s", pemKey, pemKey2)
	}

	err = s.UpstageTokenPKI(ctx, name)
	checkErr(t, err)
	pemCert3, pemKey3, err := s.RetrieveCurrentTokenPKI(ctx, name)
	checkErr(t, err)
	if !bytes.Equal(pemCert, pemCert3) {
		t.Fatalf("pem cert mismatch: %s vs. %s", pemCert, pemCert3)
	}
	if !bytes.Equal(pemKey, pemKey3) {
		t.Fatalf("pem key mismatch: %s vs. %s", pemKey, pemKey3)
	}

	r, err := s.QueryDEPNames(ctx, &storage.DEPNamesQueryRequest{
		Filter: &storage.DEPNamesQueryFilter{DEPNames: []string{name}},
	})
	checkErr(t, err)
	if r == nil {
		t.Fatal("result is nil")
	}
	if have, want := r.DEPNames, []string{name}; !reflect.DeepEqual(have, want) {
		t.Errorf("query DEP names: have: %v, want: %v", have, want)
	}

	// Token storing and retrieval.
	if _, err := s.RetrieveAuthTokens(ctx, name); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("unexpected error: %s", err)
	}
	tokens := &client.OAuth1Tokens{
		ConsumerKey:       "CK_9af2f8218b150c351ad802c6f3d66abe",
		ConsumerSecret:    "CS_9af2f8218b150c351ad802c6f3d66abe",
		AccessToken:       "AT_9af2f8218b150c351ad802c6f3d66abe",
		AccessSecret:      "AS_9af2f8218b150c351ad802c6f3d66abe",
		AccessTokenExpiry: time.Now().UTC(),
	}
	err = s.StoreAuthTokens(ctx, name, tokens)
	checkErr(t, err)
	tokens2, err := s.RetrieveAuthTokens(ctx, name)
	checkErr(t, err)
	checkTokens(t, tokens, tokens2)
	tokens3 := &client.OAuth1Tokens{
		ConsumerKey:       "foo_CK_9af2f8218b150c351ad802c6f3d66abe",
		ConsumerSecret:    "foo_CS_9af2f8218b150c351ad802c6f3d66abe",
		AccessToken:       "foo_AT_9af2f8218b150c351ad802c6f3d66abe",
		AccessSecret:      "foo_AS_9af2f8218b150c351ad802c6f3d66abe",
		AccessTokenExpiry: time.Now().Add(5 * time.Second).UTC(),
	}
	err = s.StoreAuthTokens(ctx, name, tokens3)
	checkErr(t, err)
	tokens4, err := s.RetrieveAuthTokens(ctx, name)
	checkErr(t, err)
	checkTokens(t, tokens3, tokens4)

	// Config storing and retrieval.
	config, err := s.RetrieveConfig(ctx, name)
	checkErr(t, err)
	if config != nil {
		t.Fatalf("expected not-existing config: %+v", config)
	}
	config = &client.Config{
		BaseURL: "https://config.example.com",
	}
	err = s.StoreConfig(ctx, name, config)
	checkErr(t, err)
	config2, err := s.RetrieveConfig(ctx, name)
	checkErr(t, err)
	if *config != *config2 {
		t.Fatalf("config mismatch: %+v vs. %+v", config, config2)
	}
	config2 = &client.Config{
		BaseURL: "https://config2.example.com",
	}
	err = s.StoreConfig(ctx, name, config2)
	checkErr(t, err)
	config3, err := s.RetrieveConfig(ctx, name)
	checkErr(t, err)
	if *config2 != *config3 {
		t.Fatalf("config mismatch: %+v vs. %+v", config2, config3)
	}

	// Profile assigner storing and retrieval.
	profileUUID, modTime, err := s.RetrieveAssignerProfile(ctx, name)
	checkErr(t, err)
	if profileUUID != "" {
		t.Fatal("expected empty profileUUID")
	}
	if !modTime.IsZero() {
		t.Fatal("expected zero modTime")
	}
	profileUUID = "43277A13FBCA0CFC"
	err = s.StoreAssignerProfile(ctx, name, profileUUID)
	checkErr(t, err)
	profileUUID2, modTime, err := s.RetrieveAssignerProfile(ctx, name)
	checkErr(t, err)
	if profileUUID != profileUUID2 {
		t.Fatalf("profileUUID mismatch: %s vs. %s", profileUUID, profileUUID2)
	}
	now := time.Now()
	if modTime.Before(now.Add(-1*time.Minute)) || modTime.After(now.Add(1*time.Minute)) {
		t.Errorf("mismatch modTime, expected: %s (+/- 1m), actual: %s", now, modTime)
	}
	time.Sleep(1 * time.Second)
	profileUUID3 := "foo_43277A13FBCA0CFC"
	err = s.StoreAssignerProfile(ctx, name, profileUUID3)
	checkErr(t, err)
	profileUUID4, modTime2, err := s.RetrieveAssignerProfile(ctx, name)
	checkErr(t, err)
	if profileUUID3 != profileUUID4 {
		t.Fatalf("profileUUID mismatch: %s vs. %s", profileUUID, profileUUID3)
	}
	if modTime2.Equal(modTime) {
		t.Fatalf("expected time update: %s", modTime2)
	}
	now = time.Now()
	if modTime2.Before(now.Add(-1*time.Minute)) || modTime2.After(now.Add(1*time.Minute)) {
		t.Errorf("mismatch modTime, expected: %s (+/- 1m), actual: %s", now, modTime)
	}

	cursor, err := s.RetrieveCursor(ctx, name)
	checkErr(t, err)
	if cursor != "" {
		t.Fatal("expected empty cursor")
	}
	cursor = "MTY1NzI2ODE5Ny0x"
	err = s.StoreCursor(ctx, name, cursor)
	checkErr(t, err)
	cursor2, err := s.RetrieveCursor(ctx, name)
	checkErr(t, err)
	if cursor != cursor2 {
		t.Fatalf("cursor mismatch: %s vs. %s", cursor, cursor2)
	}
	cursor2 = "foo_MTY1NzI2ODE5Ny0x"
	err = s.StoreCursor(ctx, name, cursor2)
	checkErr(t, err)
	cursor3, err := s.RetrieveCursor(ctx, name)
	checkErr(t, err)
	if cursor2 != cursor3 {
		t.Fatalf("cursor mismatch: %s vs. %s", cursor2, cursor3)
	}
}

// TestQueryDEPNames generates names and queries the DEP query name endpoints.
func TestQueryDEPNames(t *testing.T, ctx context.Context, s storage.AllStorage) {
	var depNames []string
	for i := 0; i < 4; i++ {
		depNames = append(depNames, genRandName(4))
	}

	for _, name := range depNames {
		// PKI not yet stored, should be empty query
		resp, err := s.QueryDEPNames(ctx, &storage.DEPNamesQueryRequest{Filter: &storage.DEPNamesQueryFilter{DEPNames: []string{name}}})
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatal("result empty")
		}

		if len(resp.DEPNames) != 0 {
			t.Errorf("should be zero-length result n=%s: %v", name, resp.DEPNames)
		}

		pemCert, pemKey := generatePKI(t, "basicdn", 1)
		err = s.StoreTokenPKI(ctx, name, pemCert, pemKey)
		if err != nil {
			t.Fatal(err)
		}

		// PKI now stored (but not upstaged), should be exactly 1 result
		resp, err = s.QueryDEPNames(ctx, &storage.DEPNamesQueryRequest{Filter: &storage.DEPNamesQueryFilter{DEPNames: []string{name}}})
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatal("result empty")
		}

		if len(resp.DEPNames) != 1 {
			t.Errorf("should be exactly one result for name=%s, but got %d: %v", name, len(resp.DEPNames), resp.DEPNames)
		}
	}

	// now that we have some stored, let's test the pagination
	lim := 1
	ofs := 1
	resp, err := s.QueryDEPNames(
		ctx,
		&storage.DEPNamesQueryRequest{Pagination: &storage.Pagination{
			Limit:  &lim,
			Offset: &ofs,
		}},
	)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil {
		t.Fatal("result empty")
	}

	if len(resp.DEPNames) != 1 {
		t.Errorf("should be exactly %d result(s) but got %d: %v", lim, len(resp.DEPNames), resp.DEPNames)
	}

	// test that all queried results exist
	resp, err = s.QueryDEPNames(
		ctx,
		&storage.DEPNamesQueryRequest{
			Filter: &storage.DEPNamesQueryFilter{
				DEPNames: depNames,
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil {
		t.Fatal("result empty")
	}

	slices.Sort(resp.DEPNames)
	slices.Sort(depNames)

	if have, want := resp.DEPNames, depNames; !reflect.DeepEqual(have, want) {
		t.Errorf("slices not equal; have: %v, want %v", have, want)
	}
}

// TestDevices exercises device persistence across storage.AllStorage
// implementations: full round-trip, sparse-update merging, cross-DEP
// isolation, deleted retention, filtering, deletion, and pagination.
func TestDevices(t *testing.T, ctx context.Context, depName1, depName2 string, s storage.AllStorage) {
	serial1, serial2, serial3 := genRandSerial(), genRandSerial(), genRandSerial()

	// empty query returns zero devices
	resp, err := s.QueryDevices(ctx, &storage.DevicesQueryRequest{})
	checkErr(t, err)
	if resp == nil {
		t.Fatal("result is nil")
	}
	if len(resp.Devices) != 0 {
		t.Fatalf("expected zero devices, got %d", len(resp.Devices))
	}

	// store two fully-populated devices on depName1
	assignedDate := time.Now().UTC().Truncate(time.Microsecond)
	opDate := assignedDate.Add(time.Hour)
	full1 := fullTestDevice(serial1, assignedDate, opDate)
	full2 := fullTestDevice(serial2, assignedDate, opDate)
	checkErr(t, s.StoreDevices(ctx, depName1, []godep.Device{full1, full2}))

	resp, err = s.QueryDevices(ctx, &storage.DevicesQueryRequest{})
	checkErr(t, err)
	if len(resp.Devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(resp.Devices))
	}
	bySerial := indexBySerial(resp.Devices)
	checkStoredDevice(t, bySerial[serial1], depName1, full1)
	checkStoredDevice(t, bySerial[serial2], depName1, full2)

	// query by serial subset
	resp, err = s.QueryDevices(ctx, &storage.DevicesQueryRequest{
		Filter: &storage.DevicesQueryFilter{SerialNumbers: []string{serial2}},
	})
	checkErr(t, err)
	if len(resp.Devices) != 1 || resp.Devices[0].Device.SerialNumber != serial2 {
		t.Fatalf("unexpected serial filter result: %+v", resp.Devices)
	}

	// query by unknown DEP name returns zero
	resp, err = s.QueryDevices(ctx, &storage.DevicesQueryRequest{
		Filter: &storage.DevicesQueryFilter{DEPNames: []string{"no_such_dep_name"}},
	})
	checkErr(t, err)
	if len(resp.Devices) != 0 {
		t.Fatalf("expected zero devices for unknown DEP, got %d", len(resp.Devices))
	}

	// sparse update: only serial+model+op_type must preserve everything else
	added := godep.DeviceOpTypeAdded
	checkErr(t, s.StoreDevices(ctx, depName1, []godep.Device{{
		SerialNumber: serial1,
		Model:        full1.Model,
		OpType:       &added,
	}}))
	resp, err = s.QueryDevices(ctx, &storage.DevicesQueryRequest{
		Filter: &storage.DevicesQueryFilter{SerialNumbers: []string{serial1}},
	})
	checkErr(t, err)
	if len(resp.Devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(resp.Devices))
	}
	got := resp.Devices[0]
	if string(*got.Device.OpType) != string(added) {
		t.Errorf("op_type not updated: %v", got.Device.OpType)
	}
	// previously stored fields must be preserved, not wiped to nil
	want := full1
	want.OpType = &added
	checkStoredDevice(t, got, depName1, want)

	// cross-DEP isolation: same serial on depName2 is an independent
	// row; depName1's row is untouched. Compound key (dep_name, serial).
	checkErr(t, s.StoreDevices(ctx, depName2, []godep.Device{{
		SerialNumber: serial1,
		Model:        full1.Model,
	}}))
	resp, err = s.QueryDevices(ctx, &storage.DevicesQueryRequest{
		Filter: &storage.DevicesQueryFilter{SerialNumbers: []string{serial1}},
	})
	checkErr(t, err)
	if len(resp.Devices) != 2 {
		t.Fatalf("expected 2 rows for serial across DEPs, got %d", len(resp.Devices))
	}
	// depName1 row retains its merged state
	resp, err = s.QueryDevices(ctx, &storage.DevicesQueryRequest{
		Filter: &storage.DevicesQueryFilter{
			DEPNames:      []string{depName1},
			SerialNumbers: []string{serial1},
		},
	})
	checkErr(t, err)
	if len(resp.Devices) != 1 {
		t.Fatalf("expected 1 device for depName1, got %d", len(resp.Devices))
	}
	checkStoredDevice(t, resp.Devices[0], depName1, want)
	// depName2 row is a fresh sparse insert (only identity fields set)
	resp, err = s.QueryDevices(ctx, &storage.DevicesQueryRequest{
		Filter: &storage.DevicesQueryFilter{
			DEPNames:      []string{depName2},
			SerialNumbers: []string{serial1},
		},
	})
	checkErr(t, err)
	if len(resp.Devices) != 1 {
		t.Fatalf("expected 1 device for depName2, got %d", len(resp.Devices))
	}
	got2 := resp.Devices[0]
	if got2.DEPName != depName2 || got2.Device.SerialNumber != serial1 || got2.Device.Model != full1.Model {
		t.Errorf("unexpected depName2 row: %+v", got2)
	}
	if got2.Device.AssetTag != nil || got2.Device.OpType != nil {
		t.Errorf("depName2 sparse insert should not inherit depName1 fields: %+v", got2.Device)
	}

	// deleted op_type is retained with other columns intact (same DEP)
	deleted := godep.DeviceOpTypeDeleted
	checkErr(t, s.StoreDevices(ctx, depName1, []godep.Device{{
		SerialNumber: serial2,
		Model:        full2.Model,
		OpType:       &deleted,
	}}))
	resp, err = s.QueryDevices(ctx, &storage.DevicesQueryRequest{
		Filter: &storage.DevicesQueryFilter{
			DEPNames:      []string{depName1},
			SerialNumbers: []string{serial2},
		},
	})
	checkErr(t, err)
	if len(resp.Devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(resp.Devices))
	}
	want2 := full2
	want2.OpType = &deleted
	checkStoredDevice(t, resp.Devices[0], depName1, want2)

	// pagination: store a third device, then page through
	checkErr(t, s.StoreDevices(ctx, depName1, []godep.Device{fullTestDevice(serial3, assignedDate, opDate)}))
	lim, ofs := 1, 1
	resp, err = s.QueryDevices(ctx, &storage.DevicesQueryRequest{
		Pagination: &storage.Pagination{Limit: &lim, Offset: &ofs},
	})
	checkErr(t, err)
	if len(resp.Devices) != 1 {
		t.Fatalf("expected exactly 1 paged device, got %d", len(resp.Devices))
	}

	// delete: remove serial1 from depName1 only; depName2 row survives
	checkErr(t, s.DeleteDevices(ctx, depName1, []string{serial1}))
	resp, err = s.QueryDevices(ctx, &storage.DevicesQueryRequest{
		Filter: &storage.DevicesQueryFilter{
			DEPNames:      []string{depName1},
			SerialNumbers: []string{serial1},
		},
	})
	checkErr(t, err)
	if len(resp.Devices) != 0 {
		t.Fatalf("expected 0 devices after delete, got %d", len(resp.Devices))
	}
	resp, err = s.QueryDevices(ctx, &storage.DevicesQueryRequest{
		Filter: &storage.DevicesQueryFilter{
			DEPNames:      []string{depName2},
			SerialNumbers: []string{serial1},
		},
	})
	checkErr(t, err)
	if len(resp.Devices) != 1 {
		t.Fatalf("expected depName2 row to survive delete, got %d", len(resp.Devices))
	}

	// delete with empty slice is a no-op
	checkErr(t, s.DeleteDevices(ctx, depName1, nil))
	resp, err = s.QueryDevices(ctx, &storage.DevicesQueryRequest{
		Filter: &storage.DevicesQueryFilter{DEPNames: []string{depName1}},
	})
	checkErr(t, err)
	if len(resp.Devices) != 2 {
		t.Fatalf("expected 2 devices after empty delete, got %d", len(resp.Devices))
	}

	// delete of a non-existent serial is not an error
	checkErr(t, s.DeleteDevices(ctx, depName1, []string{"no_such_serial"}))

	// delete the rest; store is empty afterwards
	checkErr(t, s.DeleteDevices(ctx, depName1, []string{serial2, serial3}))
	checkErr(t, s.DeleteDevices(ctx, depName2, []string{serial1}))
	resp, err = s.QueryDevices(ctx, &storage.DevicesQueryRequest{})
	checkErr(t, err)
	if len(resp.Devices) != 0 {
		t.Fatalf("expected zero devices after full delete, got %d", len(resp.Devices))
	}

	// cursor pagination is unsupported
	cursor := "some_cursor"
	_, err = s.QueryDevices(ctx, &storage.DevicesQueryRequest{
		Pagination: &storage.Pagination{Cursor: &cursor},
	})
	if !errors.Is(err, storage.ErrOnlyOffset) {
		t.Fatalf("expected ErrOnlyOffset, got %v", err)
	}
}

func stringPtr(s string) *string { return &s }

func fullTestDevice(serial string, assignedDate, opDate time.Time) godep.Device {
	family := godep.DeviceDeviceFamilyIPhone
	os := godep.DeviceOSIOS
	opType := godep.DeviceOpTypeAdded
	status := godep.DeviceProfileStatusAssigned
	return godep.Device{
		SerialNumber:          serial,
		Model:                 "iPhone16,2",
		AssetTag:              stringPtr("asset-" + serial),
		BluetoothMACAddress:   stringPtr("aa:bb:cc:dd:ee:ff"),
		Color:                 stringPtr("black"),
		Description:           stringPtr("iPhone"),
		DeviceAssignedBy:      stringPtr("admin@example.com"),
		DeviceAssignedDate:    &assignedDate,
		DeviceFamily:          &family,
		EID:                   stringPtr("eid-" + serial),
		EthernetMACAddress:    stringPtr("11:22:33:44:55:66"),
		IMEI:                  []string{"imei1-" + serial, "imei2-" + serial},
		IsReplacementDevice:   boolPtr(false),
		MdmMigrationDeadline:  &opDate,
		MEID:                  []string{"meid-" + serial},
		OpDate:                &opDate,
		OpType:                &opType,
		OS:                    &os,
		ProfileAssignTime:     &assignedDate,
		ProfilePushTime:       &opDate,
		ProfileStatus:         &status,
		ProfileUUID:           stringPtr("43277A13FBCA0CFC"),
		ReleasedByReplacement: boolPtr(false),
		ResponseStatus:        stringPtr("SUCCESS"),
		WiFiMACAddress:        stringPtr("ff:ee:dd:cc:bb:aa"),
	}
}

func boolPtr(b bool) *bool { return &b }

func indexBySerial(devices []storage.StoredDevice) map[string]storage.StoredDevice {
	out := make(map[string]storage.StoredDevice, len(devices))
	for _, d := range devices {
		out[d.Device.SerialNumber] = d
	}
	return out
}

// checkStoredDevice compares got against want field-by-field, comparing
// times at microsecond precision (storage granularity).
func checkStoredDevice(t *testing.T, got storage.StoredDevice, wantDEP string, want godep.Device) {
	t.Helper()
	if got.DEPName != wantDEP {
		t.Errorf("dep name: have %s, want %s", got.DEPName, wantDEP)
	}
	g, w := got.Device, want
	if g.SerialNumber != w.SerialNumber || g.Model != w.Model {
		t.Errorf("identity: have %s/%s, want %s/%s", g.SerialNumber, g.Model, w.SerialNumber, w.Model)
	}
	checkStrPtr(t, "asset_tag", g.AssetTag, w.AssetTag)
	checkStrPtr(t, "color", g.Color, w.Color)
	checkStrPtr(t, "description", g.Description, w.Description)
	checkStrPtr(t, "device_assigned_by", g.DeviceAssignedBy, w.DeviceAssignedBy)
	checkStrPtr(t, "op_type", strOf(g.OpType), strOf(w.OpType))
	checkStrPtr(t, "profile_status", strOf(g.ProfileStatus), strOf(w.ProfileStatus))
	checkStrPtr(t, "profile_uuid", g.ProfileUUID, w.ProfileUUID)
	checkStrPtr(t, "os", strOf(g.OS), strOf(w.OS))
	checkStrPtr(t, "device_family", strOf(g.DeviceFamily), strOf(w.DeviceFamily))
	checkTimePtr(t, "device_assigned_date", g.DeviceAssignedDate, w.DeviceAssignedDate)
	checkTimePtr(t, "op_date", g.OpDate, w.OpDate)
	checkTimePtr(t, "profile_assign_time", g.ProfileAssignTime, w.ProfileAssignTime)
	checkTimePtr(t, "profile_push_time", g.ProfilePushTime, w.ProfilePushTime)
	checkTimePtr(t, "mdm_migration_deadline", g.MdmMigrationDeadline, w.MdmMigrationDeadline)
	if !reflect.DeepEqual(g.IMEI, w.IMEI) {
		t.Errorf("imei: have %v, want %v", g.IMEI, w.IMEI)
	}
	if !reflect.DeepEqual(g.MEID, w.MEID) {
		t.Errorf("meid: have %v, want %v", g.MEID, w.MEID)
	}
	if !reflect.DeepEqual(g.IsReplacementDevice, w.IsReplacementDevice) {
		t.Errorf("is_replacement_device: have %v, want %v", g.IsReplacementDevice, w.IsReplacementDevice)
	}
}

func strOf[T ~string](p *T) *string {
	if p == nil {
		return nil
	}
	s := string(*p)
	return &s
}

func checkStrPtr(t *testing.T, name string, have, want *string) {
	t.Helper()
	if !reflect.DeepEqual(have, want) {
		t.Errorf("%s: have %v, want %v", name, strVal(have), strVal(want))
	}
}

func strVal(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func checkTimePtr(t *testing.T, name string, have, want *time.Time) {
	t.Helper()
	if have == nil || want == nil {
		if have != want {
			// one nil and the other not (both nil caught by != on pointers
			// only when exactly one is nil)
			if (have == nil) != (want == nil) {
				t.Errorf("%s: have %v, want %v", name, have, want)
			}
		}
		return
	}
	if have.UnixMicro() != want.UnixMicro() {
		t.Errorf("%s: have %v, want %v", name, have, want)
	}
}

func checkTokens(t *testing.T, t1 *client.OAuth1Tokens, t2 *client.OAuth1Tokens) {
	if t1 == nil || t2 == nil {
		t.Fatalf("check tokens nil")
		return
	}
	if t1.ConsumerKey != t2.ConsumerKey {
		t.Fatalf("tokens consumer_key mismatch: %s vs. %s", t1.ConsumerKey, t2.ConsumerKey)
	}
	if t1.ConsumerSecret != t2.ConsumerSecret {
		t.Fatalf("tokens consumer_secret mismatch: %s vs. %s", t1.ConsumerSecret, t2.ConsumerSecret)
	}
	if t1.AccessToken != t2.AccessToken {
		t.Fatalf("tokens access_token mismatch: %s vs. %s", t1.AccessToken, t2.AccessToken)
	}
	if t1.AccessSecret != t2.AccessSecret {
		t.Fatalf("tokens access_secret mismatch: %s vs. %s", t1.AccessSecret, t2.AccessSecret)
	}
	diff := t1.AccessTokenExpiry.Sub(t2.AccessTokenExpiry)
	if diff > 1*time.Second || diff < -1*time.Second {
		t.Fatalf("tokens expiry mismatch: %s vs. %s", t1.AccessTokenExpiry, t2.AccessTokenExpiry)
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}

func generatePKI(t *testing.T, cn string, days int64) (pemCert []byte, pemKey []byte) {
	key, cert, err := tokenpki.SelfSignedRSAKeypair(cn, days)
	if err != nil {
		t.Fatal(err)
	}
	pemCert = cryptoutil.PEMCertificate(cert.Raw)
	pemKey = cryptoutil.PEMRSAPrivateKey(key)
	return pemCert, pemKey
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func genRandName(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = byte(rand.Intn(26) + 'a')
	}
	return "go_test_dep_name." + string(result)
}

func genRandSerial() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 12)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return "T" + string(result)
}
