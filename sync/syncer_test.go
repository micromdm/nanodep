package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/micromdm/nanodep/client"
	"github.com/micromdm/nanodep/godep"
)

// testStore is a minimal godep.ClientStorage + CursorStorage for tests.
type testStore struct {
	baseURL string
	cursors map[string]string
}

func (s *testStore) RetrieveAuthTokens(context.Context, string) (*client.OAuth1Tokens, error) {
	return &client.OAuth1Tokens{
		ConsumerKey:    "test",
		ConsumerSecret: "test",
		AccessToken:    "test",
		AccessSecret:   "test",
	}, nil
}

func (s *testStore) RetrieveConfig(context.Context, string) (*client.Config, error) {
	return &client.Config{BaseURL: s.baseURL}, nil
}

func (s *testStore) RetrieveCursor(_ context.Context, name string) (string, error) {
	return s.cursors[name], nil
}

func (s *testStore) StoreCursor(_ context.Context, name string, cursor string) error {
	s.cursors[name] = cursor
	return nil
}

// serveSession handles the DEP /session auth endpoint with a dummy token.
// The test server ignores the OAuth signature.
func serveSession(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json;charset=UTF8")
	_, _ = w.Write([]byte(`{"auth_session_token":"test-session"}`))
}

// TestSyncerSameCursorMoreToFollow is a regression test for
// https://github.com/micromdm/nanodep/issues/73: Apple sometimes echoes
// the request cursor back with more_to_follow set. The syncer must not
// spin re-requesting the same cursor; it should exit the cycle early.
func TestSyncerSameCursorMoreToFollow(t *testing.T) {
	const depName = "test-dep"
	const seedCursor = "cursor-A"

	var requests atomic.Int32
	// maxRequests is a test-only short-circuit: without the fix the
	// syncer would loop forever re-requesting the same cursor, so the
	// fake server stops asserting more-to-follow after this many
	// requests to let the buggy loop terminate.
	const maxRequests = 25

	store := &testStore{cursors: map[string]string{depName: seedCursor}}
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/session" {
			serveSession(w, r)
			return
		}
		if r.URL.Path != "/server/devices" {
			http.NotFound(w, r)
			return
		}
		n := requests.Add(1)
		var req godep.FetchDeviceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var cursor string
		if req.Cursor != nil {
			cursor = *req.Cursor
		}
		more := true
		if n >= maxRequests {
			more = false
		}
		w.Header().Set("Content-Type", "application/json;charset=UTF8")
		_ = json.NewEncoder(w).Encode(godep.FetchDeviceResponse{
			Cursor:       cursor, // echo the request cursor (the Apple bug)
			MoreToFollow: more,
		})
	}))
	defer srv.Close()
	store.baseURL = srv.URL

	depClient := godep.NewClient(store)
	syncer := NewSyncer(depClient, depName, store) // run-once mode (no duration)

	if err := syncer.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if n := requests.Load(); n != 1 {
		t.Fatalf("expected 1 fetch request (exit early on same cursor), got %d", n)
	}
	if cursor := store.cursors[depName]; cursor != seedCursor {
		t.Fatalf("expected stored cursor to remain %q, got %q", seedCursor, cursor)
	}
}

// TestSyncerAdvancingCursorStillPages ensures the same-cursor guard does
// not break normal pagination: while the server advances the cursor the
// syncer must keep following more-to-follow.
func TestSyncerAdvancingCursorStillPages(t *testing.T) {
	const depName = "test-dep"

	var requests atomic.Int32
	cursors := []string{"cursor-1", "cursor-2", "cursor-2"}

	store := &testStore{cursors: map[string]string{depName: ""}}
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/session" {
			serveSession(w, r)
			return
		}
		if r.URL.Path != "/server/devices" {
			http.NotFound(w, r)
			return
		}
		n := requests.Add(1)
		more := n < 3
		w.Header().Set("Content-Type", "application/json;charset=UTF8")
		_ = json.NewEncoder(w).Encode(godep.FetchDeviceResponse{
			Cursor:       cursors[n-1],
			MoreToFollow: more,
		})
	}))
	defer srv.Close()
	store.baseURL = srv.URL

	depClient := godep.NewClient(store)
	syncer := NewSyncer(depClient, depName, store) // run-once mode (no duration)

	if err := syncer.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if n := requests.Load(); n != 3 {
		t.Fatalf("expected 3 fetch requests while cursor advances, got %d", n)
	}
	if cursor := store.cursors[depName]; cursor != "cursor-2" {
		t.Fatalf("expected stored cursor %q, got %q", "cursor-2", cursor)
	}
}
