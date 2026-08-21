package agentrunapi

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func TestEnsureGrokRunnerBound_noopWhenAlreadyBound(t *testing.T) {
	home := t.TempDir()
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	sid := "sess-bound-grok"
	meta := agentstorage.SessionMeta{
		Runner:          "grok-tty",
		SessionID:       sid,
		Status:          "running",
		RunnerSessionID: "01a0230f-b3c3-7b42-9723-b56c7b97bd0a",
	}
	if err := store.CreateSession(sid, meta); err != nil {
		t.Fatal(err)
	}
	got, bound := EnsureGrokRunnerBound(store, meta, GrokRunnerBindOpts{})
	if !bound || got.RunnerSessionID != meta.RunnerSessionID {
		t.Fatalf("bound=%v id=%q", bound, got.RunnerSessionID)
	}
}

func TestEnsureGrokRunnerBound_noopForCodex(t *testing.T) {
	home := t.TempDir()
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	sid := "sess-codex"
	meta := agentstorage.SessionMeta{Runner: "codex-tty", SessionID: sid}
	if err := store.CreateSession(sid, meta); err != nil {
		t.Fatal(err)
	}
	got, bound := EnsureGrokRunnerBound(store, meta, GrokRunnerBindOpts{})
	if bound || got.RunnerSessionID != "" {
		t.Fatalf("codex must not bind via grok path: bound=%v id=%q", bound, got.RunnerSessionID)
	}
}

func TestEnsureGrokRunnerBound_fromBindJSON(t *testing.T) {
	home := t.TempDir()
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		t.Fatal(err)
	}
	sid := "sess-bindjson"
	meta := agentstorage.SessionMeta{
		Runner:    "grok-tty",
		SessionID: sid,
		Status:    "running",
		Workspace: "/tmp/ws",
	}
	if err := store.CreateSession(sid, meta); err != nil {
		t.Fatal(err)
	}
	wantID := "01a0230f-aaaaaaaa-bbbb-cccccccccccc"
	bindPath := filepath.Join(home, "sessions", sid, "bind.json")
	data, _ := json.Marshal(map[string]string{
		"state":             "ok",
		"runner_session_id": wantID,
	})
	if err := os.WriteFile(bindPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	got, bound := EnsureGrokRunnerBound(store, meta, GrokRunnerBindOpts{})
	if !bound || got.RunnerSessionID != wantID {
		t.Fatalf("bound=%v id=%q want %q", bound, got.RunnerSessionID, wantID)
	}
	reloaded, err := store.GetSession(sid)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Meta.RunnerSessionID != wantID {
		t.Fatalf("store not persisted: %q", reloaded.Meta.RunnerSessionID)
	}
}

func TestEnsureGrokRunnerBound_discoverFromGrokHome(t *testing.T) {
	agentHome := t.TempDir()
	grokHome := t.TempDir()
	store, err := agentstorage.NewFileStore(agentHome)
	if err != nil {
		t.Fatal(err)
	}
	ws := filepath.Join(t.TempDir(), "workspace")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatal(err)
	}
	sid := "sess-discover"
	prompt := "unique-resume-bind-prompt-xyz"
	created := time.Now().UTC().Add(-time.Minute)
	meta := agentstorage.SessionMeta{
		Runner:          "grok-tty",
		SessionID:       sid,
		Status:          "running",
		Workspace:       ws,
		InitialPrompt:   prompt,
		CreatedAt:       created.Format(time.RFC3339Nano),
	}
	if err := store.CreateSession(sid, meta); err != nil {
		t.Fatal(err)
	}

	providerID := "01a0230f-dddddddd-eeee-ffffffffffff"
	absWS, err := filepath.Abs(ws)
	if err != nil {
		t.Fatal(err)
	}
	encWS := url.PathEscape(absWS)
	sessDir := filepath.Join(grokHome, "sessions", encWS, providerID)
	if err := os.MkdirAll(sessDir, 0o755); err != nil {
		t.Fatal(err)
	}
	summary := map[string]any{
		"created_at": created.Format(time.RFC3339Nano),
		"info":       map[string]any{"cwd": ws, "id": providerID},
	}
	sumBytes, _ := json.Marshal(summary)
	if err := os.WriteFile(filepath.Join(sessDir, "summary.json"), sumBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	// updates.jsonl must contain a user_message_chunk with the prompt text.
	updateLine := `{"method":"session/update","params":{"sessionId":` + jsonQuote(providerID) + `,"update":{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":` + jsonQuote(prompt) + `}}}}` + "\n"
	if err := os.WriteFile(filepath.Join(sessDir, "updates.jsonl"), []byte(updateLine), 0o644); err != nil {
		t.Fatal(err)
	}

	got, bound := EnsureGrokRunnerBound(store, meta, GrokRunnerBindOpts{
		GrokHome:        grokHome,
		DiscoverTimeout: 2 * time.Second,
		NotBefore:       created.Add(-time.Second),
	})
	if !bound || got.RunnerSessionID != providerID {
		t.Fatalf("bound=%v id=%q want %q", bound, got.RunnerSessionID, providerID)
	}
}

func jsonQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func TestIsGrokRunner(t *testing.T) {
	if !isGrokRunner("grok-tty") || !isGrokRunner("grok") {
		t.Fatal("grok runners")
	}
	if isGrokRunner("codex-tty") {
		t.Fatal("codex is not grok")
	}
}

func TestIsSessionReadyFromStatus_idleBound(t *testing.T) {
	out := "" +
		"screen status: idle\n" +
		"sendable: yes\n" +
		"runner session id: 01a0230f-b3c3-7b42-9723-b56c7b97bd0a\n"
	if !IsSessionReadyFromStatus(out) {
		t.Fatal("idle+sendable+bound should be ready")
	}
}

func TestIsSessionReadyFromStatus_idleUnbound(t *testing.T) {
	out := "" +
		"screen status: idle\n" +
		"sendable: yes\n" +
		"runner session id: (unbound)\n"
	if IsSessionReadyFromStatus(out) {
		t.Fatal("unbound must not be ready")
	}
}

func TestIsSessionReadyFromStatus_legacyNoBindLine(t *testing.T) {
	out := "" +
		"screen status: banner\n" +
		"sendable: yes\n"
	if !IsSessionReadyFromStatus(out) {
		t.Fatal("legacy banner+sendable without bind line should stay ready")
	}
}
