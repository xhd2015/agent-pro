package sessions

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewestTimestamp_MaxLastLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rollout-test.jsonl")
	body := "" +
		`{"timestamp":"2026-08-25T02:49:03.544Z","type":"session_meta"}` + "\n" +
		`{"timestamp":"2026-08-25T02:50:00.000Z","type":"event_msg"}` + "\n" +
		`{"timestamp":"2026-08-25T02:52:54.542Z","type":"response_item"}` + "\n" +
		`{"no_ts":true}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	tip, err := NewestTimestamp(path)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := time.Parse(time.RFC3339Nano, "2026-08-25T02:52:54.542Z")
	if !tip.Equal(want) {
		t.Fatalf("tip=%v want %v", tip, want)
	}
}

func TestNewestTimestamp_MissingEmpty(t *testing.T) {
	tip, err := NewestTimestamp(filepath.Join(t.TempDir(), "missing.jsonl"))
	if err != nil || !tip.IsZero() {
		t.Fatalf("tip=%v err=%v", tip, err)
	}
	tip, err = NewestTimestamp("")
	if err != nil || !tip.IsZero() {
		t.Fatalf("empty path tip=%v err=%v", tip, err)
	}
}

func TestTipForSession_ViaFind(t *testing.T) {
	home := t.TempDir()
	sid := "01a036d2-3e21-7381-a2d9-f7392d0efc29"
	dir := filepath.Join(home, "sessions", "2026", "08", "25")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "rollout-2026-08-25T10-49-03-"+sid+".jsonl")
	line := `{"timestamp":"2026-08-25T03:00:00.000Z","type":"event_msg"}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
	tip, err := TipForSession(home, sid)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := time.Parse(time.RFC3339Nano, "2026-08-25T03:00:00.000Z")
	if !tip.Equal(want) {
		t.Fatalf("tip=%v want %v", tip, want)
	}
}
