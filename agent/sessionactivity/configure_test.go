package sessionactivity

import (
	"os"
	"path/filepath"
	"testing"
)

// Configure is what lets a pinned backend read its own home through the activity
// folds: the package builds its own store, so without it Read/Follow would answer
// about ~/.dsh while the rest of the process reads the pinned home.
func TestConfigurePinsTheStoreHome(t *testing.T) {
	Configure(Options{})
	t.Cleanup(func() { Configure(Options{}) })

	home := t.TempDir()
	const sessionID = "session-configured"
	dir := filepath.Join(home, "sessions", "project", sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"type":"session","version":4,"id":"` + sessionID + `","createdAt":10,"cwd":"/workspace","isSeeded":false,"delegationDepth":0,"agentPreset":"std-no-subagent"}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, "session.v4.jsonl"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	Configure(Options{Home: home})
	if got := storeOptions().Home; got != home {
		t.Fatalf("store home = %q, want %q", got, home)
	}
	if _, err := dshHeader(sessionID); err != nil {
		t.Fatalf("header under the configured home: %v", err)
	}

	// A tail cursor describes one home's file; a changed home must not keep it.
	dshSessions.Store(sessionID, &dshSession{})
	Configure(Options{Home: t.TempDir()})
	if _, ok := dshSessions.Load(sessionID); ok {
		t.Fatal("the tail cache survived a home change")
	}
	if _, err := dshHeader(sessionID); err == nil {
		t.Fatal("the header resolved in a home that does not hold the session")
	}
}
