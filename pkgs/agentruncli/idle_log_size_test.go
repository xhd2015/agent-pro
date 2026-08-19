package agentruncli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveCodexRolloutPath_fromMeta(t *testing.T) {
	home := t.TempDir()
	codexHome := t.TempDir()
	sessionID := "seatalk-local-bot-test"
	runnerID := "01a018f9-a81b-76e1-84df-9c7ae9a054ef"
	day := filepath.Join(codexHome, "sessions", "2026", "08", "19")
	if err := os.MkdirAll(day, 0755); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(day, "rollout-2026-08-19T15-43-29-"+runnerID+".jsonl")
	body := []byte("{\"type\":\"session_meta\"}\n")
	if err := os.WriteFile(want, body, 0644); err != nil {
		t.Fatal(err)
	}
	metaDir := filepath.Join(home, "sessions", sessionID)
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		t.Fatal(err)
	}
	meta, err := json.Marshal(map[string]string{
		"runner":                   "codex-tty",
		"session_id":               sessionID,
		"runner_session_id":        runnerID,
		"agent_runner_config_home": codexHome,
		"status":                   "running",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(metaDir, "meta.json"), meta, 0644); err != nil {
		t.Fatal(err)
	}

	got := resolveCodexRolloutPath(home, sessionID)
	if got != want {
		t.Fatalf("path=%q want %q", got, want)
	}

	var cache idleLogSizeCache
	sample := IdleSample{}
	cache.fill(&sample, home, sessionID)
	if !sample.LogFound || sample.LogBytes != int64(len(body)) {
		t.Fatalf("fill: found=%v bytes=%d want %d", sample.LogFound, sample.LogBytes, len(body))
	}
}

func TestResolveCodexRolloutPath_skipWhenMissing(t *testing.T) {
	home := t.TempDir()
	if got := resolveCodexRolloutPath(home, "no-such"); got != "" {
		t.Fatalf("missing meta: %q", got)
	}
	sess := filepath.Join(home, "sessions", "s1")
	if err := os.MkdirAll(sess, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sess, "meta.json"), []byte(`{"runner":"codex-tty","session_id":"s1","status":"running"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if got := resolveCodexRolloutPath(home, "s1"); got != "" {
		t.Fatalf("empty runner_session_id: %q", got)
	}

	sample := IdleSample{Sendable: true, Screen: "idle", InputBox: "empty"}
	var cache idleLogSizeCache
	cache.fill(&sample, home, "s1")
	if sample.LogFound {
		t.Fatal("missing jsonl must skip LogFound")
	}
}

func TestNoteLogSize_baselineThenGrowth(t *testing.T) {
	w := NewIdleWatchdog(true, IdlePolicy{ExitOnIdle: true}, IdleWatchdog{})
	if w.noteLogSize(IdleSample{LogFound: true, LogBytes: 100}) {
		t.Fatal("first size is baseline, not growth")
	}
	if !w.noteLogSize(IdleSample{LogFound: true, LogBytes: 200}) {
		t.Fatal("size change must count as growth")
	}
	if w.noteLogSize(IdleSample{LogFound: true, LogBytes: 200}) {
		t.Fatal("unchanged size is not growth")
	}
	if w.noteLogSize(IdleSample{}) {
		t.Fatal("missing file skips gate")
	}
}
