// Package test offers a battery of tests for storage.AllStorage implementations.
package test

import (
	"bytes"
	"context"
	"errors"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/micromdm/nanodep/client"
	"github.com/micromdm/nanodep/cryptoutil"
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
		t.Fatalf("mismatch modTime, expected: %s (+/- 1m), actual: %s", now, modTime)
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
		t.Fatalf("mismatch modTime, expected: %s (+/- 1m), actual: %s", now, modTime)
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
