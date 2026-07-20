package view

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeTestGrokSession(t *testing.T, grokHome, id, cwd string, updates string) {
	t.Helper()
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	dir := filepath.Join(grokHome, "sessions", url.PathEscape(absCwd), id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	summary := map[string]any{
		"info": map[string]any{
			"id":  id,
			"cwd": absCwd,
		},
		"generated_title":   "Test session",
		"created_at":        "2026-07-03T10:00:00.000Z",
		"updated_at":        "2026-07-03T12:00:00.000Z",
		"last_active_at":    "2026-07-03T12:00:00.000Z",
		"num_messages":      2,
		"num_chat_messages": 2,
		"current_model_id":  "grok-test",
	}
	b, _ := json.Marshal(summary)
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), b, 0644); err != nil {
		t.Fatalf("write summary: %v", err)
	}
	if updates != "" {
		if err := os.WriteFile(filepath.Join(dir, "updates.jsonl"), []byte(updates), 0644); err != nil {
			t.Fatalf("write updates: %v", err)
		}
	}
}

func acpUser(text string) string {
	b, _ := json.Marshal(map[string]any{
		"sessionUpdate": "user_message_chunk",
		"content":       map[string]any{"type": "text", "text": text},
	})
	return string(b)
}

func acpAssistant(text string) string {
	b, _ := json.Marshal(map[string]any{
		"sessionUpdate": "agent_message_chunk",
		"content":       map[string]any{"type": "text", "text": text},
	})
	return string(b)
}

func acpTurnCompleted() string {
	b, _ := json.Marshal(map[string]any{"sessionUpdate": "turn_completed"})
	return string(b)
}

func TestBootstrapConvertsUserAndAssistant(t *testing.T) {
	grokHome := t.TempDir()
	id := "019f-view-test-session-0001"
	cwd := filepath.Join(t.TempDir(), "proj")
	_ = os.MkdirAll(cwd, 0755)
	updates := strings.Join([]string{
		acpUser("VIEW_USER_MARKER"),
		acpAssistant("VIEW_ASSISTANT_MARKER"),
		acpTurnCompleted(),
	}, "\n") + "\n"
	writeTestGrokSession(t, grokHome, id, cwd, updates)

	v, err := Open(grokHome, id)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := v.Bootstrap(); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	events := v.Events()
	if len(events) == 0 {
		t.Fatal("expected events")
	}
	var foundUser, foundAsst bool
	for _, ev := range events {
		if strings.Contains(ev.Text, "VIEW_USER_MARKER") {
			foundUser = true
		}
		if strings.Contains(ev.Text, "VIEW_ASSISTANT_MARKER") {
			foundAsst = true
		}
	}
	if !foundUser || !foundAsst {
		t.Fatalf("missing markers in events: %+v", events)
	}
	if v.Offset() <= 0 {
		t.Fatalf("expected positive virtual offset, got %d", v.Offset())
	}
}

func TestFollowAppendsNewLines(t *testing.T) {
	grokHome := t.TempDir()
	id := "019f-view-test-session-0002"
	cwd := filepath.Join(t.TempDir(), "proj")
	_ = os.MkdirAll(cwd, 0755)
	updatesPath := filepath.Join(grokHome, "sessions", url.PathEscape(mustAbs(t, cwd)), id, "updates.jsonl")
	writeTestGrokSession(t, grokHome, id, cwd, acpUser("FIRST_USER")+"\n"+acpTurnCompleted()+"\n")

	v, err := Open(grokHome, id)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := v.Bootstrap(); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	n0 := len(v.Events())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- v.Follow(ctx) }()

	// Give WatchLine time to attach.
	time.Sleep(100 * time.Millisecond)
	f, err := os.OpenFile(updatesPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("open updates: %v", err)
	}
	_, _ = f.WriteString(acpAssistant("FOLLOW_ASSISTANT_MARKER") + "\n" + acpTurnCompleted() + "\n")
	_ = f.Close()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		for _, ev := range v.Events() {
			if strings.Contains(ev.Text, "FOLLOW_ASSISTANT_MARKER") {
				cancel()
				<-done
				if len(v.Events()) <= n0 {
					t.Fatalf("expected more events after follow")
				}
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	cancel()
	<-done
	t.Fatalf("timeout waiting for follow event; events=%+v", v.Events())
}

func TestWebSSEAndNoAgentRunHome(t *testing.T) {
	grokHome := t.TempDir()
	agentRunHome := t.TempDir()
	t.Setenv("AGENT_RUN_HOME", agentRunHome)
	id := "019f-view-test-session-0003"
	cwd := filepath.Join(t.TempDir(), "proj")
	_ = os.MkdirAll(cwd, 0755)
	updates := acpUser("WEB_USER_MARKER") + "\n" + acpAssistant("WEB_ASSISTANT_MARKER") + "\n" + acpTurnCompleted() + "\n"
	writeTestGrokSession(t, grokHome, id, cwd, updates)

	v, err := Open(grokHome, id)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := v.Bootstrap(); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

	mux := http.NewServeMux()
	registerViewAPI(mux, v)

	srv := &http.Server{Handler: mux}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() { _ = srv.Serve(ln) }()
	defer srv.Close()

	base := "http://" + ln.Addr().String()
	res, err := http.Get(base + "/api/agent-run/sessions/" + id)
	if err != nil {
		t.Fatalf("GET detail: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("status %d", res.StatusCode)
	}
	var detail map[string]any
	if err := json.NewDecoder(res.Body).Decode(&detail); err != nil {
		t.Fatalf("decode: %v", err)
	}
	raw, _ := json.Marshal(detail)
	if !strings.Contains(string(raw), "WEB_ASSISTANT_MARKER") {
		t.Fatalf("detail missing marker: %s", raw)
	}

	// Ensure no sessions were written under AGENT_RUN_HOME.
	entries, err := os.ReadDir(filepath.Join(agentRunHome, "sessions"))
	if err == nil && len(entries) > 0 {
		t.Fatalf("expected no agent-run sessions, found %v", entries)
	}
}

func mustAbs(t *testing.T, p string) string {
	t.Helper()
	a, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return a
}
