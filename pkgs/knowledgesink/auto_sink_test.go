package knowledgesink

import (
	"testing"
	"time"
)

func TestIsAutoSinkable(t *testing.T) {
	t.Parallel()
	if IsAutoSinkable(nil) {
		t.Fatal("nil should not be sinkable")
	}
	cases := []struct {
		v    StatusView
		want bool
	}{
		{StatusView{State: StateReady, Enabled: true}, true},
		{StatusView{State: StateBehind, Enabled: true}, true},
		{StatusView{State: StateFailed, Enabled: true}, true},
		{StatusView{State: StateSunk, Enabled: false}, false},
		{StatusView{State: StateRunning, Enabled: false}, false},
		{StatusView{State: StateUnavailable, Enabled: false}, false},
	}
	for _, tc := range cases {
		v := tc.v
		if got := IsAutoSinkable(&v); got != tc.want {
			t.Fatalf("state=%s enabled=%v: got %v want %v", tc.v.State, tc.v.Enabled, got, tc.want)
		}
	}
}

func TestFilterAndSortAutoSinkable_OldestFirstWindowAndSkip(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
	old := now.Add(-6 * 24 * time.Hour)
	mid := now.Add(-3 * 24 * time.Hour)
	fresh := now.Add(-1 * time.Hour)
	tooOld := now.Add(-8 * 24 * time.Hour)

	sessions := []SessionMeta{
		{ID: "fresh-ready", UpdatedAt: fresh},
		{ID: "mid-behind", UpdatedAt: mid},
		{ID: "old-ready", UpdatedAt: old},
		{ID: "stale", UpdatedAt: tooOld},
		{ID: "archived", UpdatedAt: old, Archived: true},
		{ID: "sunk", UpdatedAt: mid},
		{ID: "zero-time"},
	}
	status := map[string]*StatusView{
		"fresh-ready": {State: StateReady, Enabled: true},
		"mid-behind":  {State: StateBehind, Enabled: true, Help: "New Grok messages since last sink"},
		"old-ready":   {State: StateReady, Enabled: true},
		"sunk":        {State: StateSunk, Enabled: false},
	}
	rows, err := FilterAndSortAutoSinkable(sessions, now, AutoSinkWindow, func(id string) (*StatusView, error) {
		return status[id], nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("rows=%d want 3: %+v", len(rows), rows)
	}
	wantIDs := []string{"old-ready", "mid-behind", "fresh-ready"}
	for i, id := range wantIDs {
		if rows[i].SessionID != id {
			t.Fatalf("row[%d]=%s want %s", i, rows[i].SessionID, id)
		}
	}
	if rows[0].Why != "never sunk" {
		t.Fatalf("why[0]=%q", rows[0].Why)
	}
	if rows[1].State != StateBehind {
		t.Fatalf("state[1]=%s", rows[1].State)
	}

	pick, err := PickOldestAutoSinkable(sessions, now, AutoSinkWindow, func(id string) (*StatusView, error) {
		return status[id], nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if pick == nil || pick.SessionID != "old-ready" {
		t.Fatalf("pick=%v", pick)
	}
}

func TestPickOldestAutoSinkable_None(t *testing.T) {
	t.Parallel()
	now := time.Now()
	pick, err := PickOldestAutoSinkable(nil, now, AutoSinkWindow, func(string) (*StatusView, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if pick != nil {
		t.Fatalf("want nil, got %+v", pick)
	}
}

func TestWithinAutoSinkWindow(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	if WithinAutoSinkWindow(time.Time{}, now, AutoSinkWindow) {
		t.Fatal("zero updated")
	}
	if !WithinAutoSinkWindow(now.Add(-24*time.Hour), now, AutoSinkWindow) {
		t.Fatal("1d should be in window")
	}
	if WithinAutoSinkWindow(now.Add(-8*24*time.Hour), now, AutoSinkWindow) {
		t.Fatal("8d should be out")
	}
}
