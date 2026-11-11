package web

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// The instant an answer carries is the batch's instant, not the lookup's: a
// cache hit must report the same At, or a reader could mistake an old answer for
// a fresh one and call a live turn dead.
func TestRunStatusReportsTheBatchInstant(t *testing.T) {
	srv, _, _ := statusServer(t, map[string]bool{"running-session": true})
	now := time.Unix(1000, 0)
	rs := NewRunStatus(statusClient(srv))
	rs.Now = func() time.Time { return now }

	first := rs.Lookup("running-session")
	if !first.Running || !first.Present || !first.Answered || !first.At.Equal(now) {
		t.Fatalf("%+v want running at %v", first, now)
	}
	now = now.Add(DefaultRunStatusTTL / 2)
	second := rs.Lookup("running-session")
	if !second.At.Equal(first.At) {
		t.Fatalf("at=%v want %v, a cache hit keeps the batch's instant", second.At, first.At)
	}

	now = first.At.Add(DefaultRunStatusTTL)
	third := rs.Lookup("running-session")
	if !third.At.Equal(now) {
		t.Fatalf("at=%v want %v, the refresh samples a new instant", third.At, now)
	}
}

// An absent id and a failed RPC both carry no testimony: At may be set for the
// batch, but Answered and Present tell the reader not to use it.
func TestRunStatusUnknownAnswersCarryNoTestimony(t *testing.T) {
	srv, _, _ := statusServer(t, map[string]bool{})
	now := time.Unix(1000, 0)
	rs := NewRunStatus(statusClient(srv))
	rs.Now = func() time.Time { return now }

	absent := rs.Lookup("never-loaded")
	if absent.Present || !absent.Answered || !absent.At.Equal(now) {
		t.Fatalf("%+v want an answered map without the id", absent)
	}
	if running, known := rs.Running("never-loaded"); running || known {
		t.Fatalf("running=%v known=%v", running, known)
	}
}

// A failed refresh reports unknown for every id while still naming when it
// tried, so the caller can tell "not asked" from "asked and refused".
func TestRunStatusFailureKeepsTheAttemptInstant(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	now := time.Unix(1000, 0)
	rs := NewRunStatus(statusClient(srv))
	rs.Now = func() time.Time { return now }

	answer := rs.Lookup("anything")
	if answer.Answered || answer.Present || answer.Running {
		t.Fatalf("%+v want unknown", answer)
	}
	if !answer.At.Equal(now) {
		t.Fatalf("at=%v want %v", answer.At, now)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("calls=%d want 1", got)
	}
}
