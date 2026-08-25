package knowledgesink

import (
	"path/filepath"
	"testing"
	"time"
)

func TestHasSinkHistory(t *testing.T) {
	if HasSinkHistory(nil) {
		t.Fatal("nil")
	}
	if HasSinkHistory(&Manifest{}) {
		t.Fatal("empty")
	}
	if !HasSinkHistory(&Manifest{LastSinkAt: "2026-08-25 10:00:00 +0800 CST"}) {
		t.Fatal("want true")
	}
}

func TestListSinkedSessions_NewestFirstFiltersIncomplete(t *testing.T) {
	state := t.TempDir()
	newer := time.Date(2026, 8, 25, 10, 39, 0, 0, time.FixedZone("CST", 8*3600))
	older := time.Date(2026, 8, 25, 9, 6, 0, 0, time.FixedZone("CST", 8*3600))

	write := func(id, lastSink, status, mr string) {
		t.Helper()
		dir := SessionDir(state, id)
		if err := WriteManifest(dir, &Manifest{
			MarcusSessionID: id,
			LastSinkAt:      lastSink,
			Status:          status,
			LastMRURL:       mr,
			LastSinkIndex:   -1,
		}); err != nil {
			t.Fatal(err)
		}
	}
	write("sess-new", FormatTime(newer), statusIdle, "https://example/mr/2")
	write("sess-old", FormatTime(older), statusIdle, "https://example/mr/1")
	// Incomplete: no last_sink_at
	if err := WriteManifest(SessionDir(state, "sess-running"), &Manifest{
		MarcusSessionID: "sess-running",
		Status:          statusRunning,
		LastSinkIndex:   -1,
		NextSinkIndex:   1,
	}); err != nil {
		t.Fatal(err)
	}
	if err := WriteManifest(SessionDir(state, "sess-failed"), &Manifest{
		MarcusSessionID: "sess-failed",
		Status:          statusFailed,
		Error:           "boom",
		LastSinkIndex:   -1,
	}); err != nil {
		t.Fatal(err)
	}

	rows, err := ListSinkedSessions(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("len=%d rows=%+v", len(rows), rows)
	}
	if rows[0].SessionID != "sess-new" || rows[1].SessionID != "sess-old" {
		t.Fatalf("order=%v %v", rows[0].SessionID, rows[1].SessionID)
	}
	if rows[0].LastMRURL != "https://example/mr/2" || rows[0].Status != statusIdle {
		t.Fatalf("row0=%+v", rows[0])
	}
	if !rows[0].LastSinkAt.Equal(newer) {
		t.Fatalf("at=%v want %v", rows[0].LastSinkAt, newer)
	}
}

func TestListSinkedSessions_EmptyRoot(t *testing.T) {
	rows, err := ListSinkedSessions(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("rows=%+v", rows)
	}
}
