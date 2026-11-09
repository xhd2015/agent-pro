package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// statusServer answers session/status and records what each call asked for.
func statusServer(t *testing.T, running map[string]bool) (*httptest.Server, *[]map[string]any, *int32) {
	t.Helper()
	var requests []map[string]any
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		if r.Method != "POST" || r.URL.Path != "/api/session/status" || r.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			return
		}
		var req struct {
			Type    string
			RPCID   string `json:"rpcId"`
			Payload struct {
				Args struct {
					Request map[string]any `json:"request"`
				} `json:"args"`
			} `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		if req.Type != "client-request" {
			t.Errorf("envelope type %q", req.Type)
		}
		requests = append(requests, req.Payload.Args.Request)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":  "server-response",
			"rpcId": req.RPCID,
			"result": map[string]any{
				"ok":    true,
				"value": map[string]any{"running": running},
			},
		})
	}))
	t.Cleanup(srv.Close)
	return srv, &requests, &calls
}

func statusClient(srv *httptest.Server) Client {
	return Client{ReadListen: func() (Listen, error) { return Listen{Origin: srv.URL, Token: "secret"}, nil }}
}

func TestStatusRequestShape(t *testing.T) {
	srv, requests, _ := statusServer(t, map[string]bool{"a": true})
	c := statusClient(srv)

	all, err := c.Status(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !all["a"] {
		t.Fatalf("running=%v", all)
	}
	if _, present := (*requests)[0]["ids"]; present {
		t.Fatalf("nil ids must ask for every loaded Session: %v", (*requests)[0])
	}
	none, err := c.Status(context.Background(), []string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(none) != 1 {
		t.Fatalf("empty ids should still return the daemon's answer: %v", none)
	}
	ids, ok := (*requests)[1]["ids"].([]any)
	if !ok || len(ids) != 0 {
		t.Fatalf("empty ids not sent as an empty list: %v", (*requests)[1])
	}
	if _, err := c.Status(context.Background(), []string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	ids, ok = (*requests)[2]["ids"].([]any)
	if !ok || len(ids) != 2 || ids[0] != "a" || ids[1] != "b" {
		t.Fatalf("ids not forwarded: %v", (*requests)[2])
	}
}

func TestRunStatusCachesOneBatchPerTTL(t *testing.T) {
	srv, _, calls := statusServer(t, map[string]bool{"running-session": true, "idle-session": false})
	now := time.Unix(1000, 0)
	rs := NewRunStatus(statusClient(srv))
	rs.Now = func() time.Time { return now }

	if live, known := rs.Running("running-session"); !live || !known {
		t.Fatalf("running: live=%v known=%v", live, known)
	}
	if live, known := rs.Running("idle-session"); live || !known {
		t.Fatalf("idle: live=%v known=%v", live, known)
	}
	if got := atomic.LoadInt32(calls); got != 1 {
		t.Fatalf("expected one RPC for two lookups in the window, got %d", got)
	}
	if live, present, answered := rs.Lookup("unloaded-session"); live || present || !answered {
		t.Fatalf("absent id: live=%v present=%v answered=%v", live, present, answered)
	}
	if _, known := rs.Running("unloaded-session"); known {
		t.Fatal("Running treats an unnamed Session as not known")
	}
	now = now.Add(DefaultRunStatusTTL)
	if live, known := rs.Running("running-session"); !live || !known {
		t.Fatalf("after ttl: live=%v known=%v", live, known)
	}
	if got := atomic.LoadInt32(calls); got != 2 {
		t.Fatalf("expected a refresh after the ttl, got %d calls", got)
	}
}

func TestRunStatusFailureReportsUnknown(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	now := time.Unix(1000, 0)
	rs := NewRunStatus(statusClient(srv))
	rs.Now = func() time.Time { return now }
	if live, known := rs.Running("running-session"); live || known {
		t.Fatalf("failed refresh: live=%v known=%v", live, known)
	}
	// A failure is never replayed as a stale "running", and it is not retried
	// once per Session lookup either.
	if live, known := rs.Running("other-session"); live || known {
		t.Fatalf("failure not remembered: live=%v known=%v", live, known)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected the failure to be cached for the window, got %d calls", got)
	}
	now = now.Add(DefaultRunStatusTTL)
	if _, known := rs.Running("running-session"); known {
		t.Fatal("a dead daemon reported a known answer")
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected one retry after the window, got %d calls", got)
	}
}

// A stalled daemon must not be mistaken for a broken one by the batch clock:
// the default window outlasts a queued reply, and a caller that cannot wait
// shortens it explicitly.
func TestRunStatusTimeoutOutlastsAStall(t *testing.T) {
	if runStatusTimeout < time.Minute {
		t.Fatalf("default status timeout=%s, want at least a minute", runStatusTimeout)
	}
	var calls int32
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		// Only the client's own deadline can end this handler before release:
		// an override that never fired would stall the test for 5s.
		select {
		case <-release:
		case <-time.After(5 * time.Second):
		}
	}))
	defer srv.Close()
	defer close(release)

	rs := NewRunStatus(statusClient(srv))
	rs.Timeout = 50 * time.Millisecond
	started := time.Now()
	if live, known := rs.Running("running-session"); live || known {
		t.Fatalf("stalled refresh: live=%v known=%v", live, known)
	}
	if elapsed := time.Since(started); elapsed < 40*time.Millisecond || elapsed > 2*time.Second {
		t.Fatalf("override timeout took %s, want about 50ms", elapsed)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected one RPC, got %d", got)
	}
}
