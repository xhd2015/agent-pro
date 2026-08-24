package sessions

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestFind_BasenameMatchSkipsDecoySummaries(t *testing.T) {
	home := t.TempDir()
	wantID := "019f283a-eeee-7eee-eeee-eeeeeeeeee01"
	decoyID := "019f283a-dddd-7ddd-dddd-dddddddddddd"

	// Many wrongly-named dirs whose summary.json claims wantID. Old Find would
	// return the first decoy path; basename Find must ignore them.
	for i := 0; i < 40; i++ {
		n := strconv.Itoa(i)
		writeFindSummary(t, home, "decoy-place-"+n, "wrong-dir-"+n, wantID, "/tmp/decoy")
	}
	// Matching JSON id under a different session dir name must not win.
	writeFindSummary(t, home, "misplaced", decoyID, wantID, "/tmp/misplaced")

	wantCWD := "/tmp/find-target"
	writeFindSummary(t, home, "target-place", wantID, wantID, wantCWD)

	got, err := Find(home, wantID)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if got.ID != wantID {
		t.Fatalf("ID = %q, want %q", got.ID, wantID)
	}
	if got.CWD != wantCWD {
		t.Fatalf("CWD = %q, want %q", got.CWD, wantCWD)
	}
	if !strings.Contains(got.Path, filepath.Join("target-place", wantID, "summary.json")) {
		t.Fatalf("Path = %q, want …/target-place/%s/summary.json", got.Path, wantID)
	}
}

func TestFind_DirNameWithoutMatchingSummaryID(t *testing.T) {
	home := t.TempDir()
	dirID := "019f283a-eeee-7eee-eeee-eeeeeeeeee01"
	// Directory named for the query, but summary.info.id disagrees.
	writeFindSummary(t, home, "place", dirID, "019f283a-ffff-7fff-ffff-ffffffffffff", "/tmp/x")

	_, err := Find(home, dirID)
	if err == nil {
		t.Fatal("expected not found when summary id mismatches directory name")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error = %v, want not found", err)
	}
}

func TestFind_Unknown(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "sessions"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := Find(home, "019f283a-eeee-7eee-eeee-eeeeeeeeee99")
	if err == nil {
		t.Fatal("expected not found")
	}
}

func writeFindSummary(t *testing.T, grokHome, place, dirName, summaryID, cwd string) {
	t.Helper()
	dir := filepath.Join(grokHome, "sessions", url.PathEscape(place), dirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	summary := map[string]any{
		"info": map[string]any{
			"id":  summaryID,
			"cwd": cwd,
		},
		"generated_title":   "find fixture",
		"created_at":        "2026-07-01T10:00:00.000Z",
		"updated_at":        "2026-07-01T11:00:00.000Z",
		"last_active_at":    "2026-07-01T11:00:00.000Z",
		"num_messages":      1,
		"num_chat_messages": 1,
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), append(body, '\n'), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

