package agentstorage

import (
	"path/filepath"
	"testing"
)

func TestIdleLogPath(t *testing.T) {
	got := IdleLogPath("/home/x/.agent-run", "sess-abc")
	want := filepath.Join("/home/x/.agent-run", "sessions", "sess-abc", "idle.jsonl")
	if got != want {
		t.Fatalf("IdleLogPath=%q want %q", got, want)
	}
}
