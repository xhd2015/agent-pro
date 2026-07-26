package agentruncli

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func TestResumeFromGrokSession_AlreadyMappedWithoutFork(t *testing.T) {
	home, grokHome, _, parentID := seedForkFixture(t)
	t.Setenv("AGENT_RUN_HOME", home)
	t.Setenv("GROK_HOME", grokHome)

	err := runResumeFromGrokSession(resumeFromGrokOpts{
		grokSessionID: parentID,
		sessionID:     "import-again",
		prompt:        "hi",
		fork:          false,
	})
	if err == nil {
		t.Fatal("expected already mapped error")
	}
	if !strings.Contains(err.Error(), "already mapped") {
		t.Fatalf("got %v", err)
	}
	if !strings.Contains(err.Error(), "--fork") {
		t.Fatalf("error should hint --fork: %v", err)
	}
}

func seedForkFixture(t *testing.T) (home, grokHome, cwd, parentID string) {
	t.Helper()
	root := t.TempDir()
	home = filepath.Join(root, ".agent-run")
	grokHome = filepath.Join(root, "grok-home")
	cwd = filepath.Join(root, "ws")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	parentID = "550e8400-e29b-41d4-a716-446655440a01"
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		t.Fatal(err)
	}
	sessDir := filepath.Join(grokHome, "sessions", url.PathEscape(absCwd), parentID)
	if err := os.MkdirAll(sessDir, 0o755); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	summary := map[string]any{
		"info": map[string]any{
			"id":  parentID,
			"cwd": absCwd,
		},
		"generated_title": "fork fixture",
		"created_at":      now,
		"updated_at":      now,
		"last_active_at":  now,
	}
	b, _ := json.MarshalIndent(summary, "", "  ")
	if err := os.WriteFile(filepath.Join(sessDir, "summary.json"), append(b, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateSession("mapped-parent", agentstorage.SessionMeta{
		Runner:          "grok-tty",
		SessionID:       "mapped-parent",
		Status:          "finished",
		RunnerSessionID: parentID,
		Workspace:       absCwd,
	}); err != nil {
		t.Fatal(err)
	}
	return home, grokHome, absCwd, parentID
}
