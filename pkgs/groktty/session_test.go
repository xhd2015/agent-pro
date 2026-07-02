package groktty

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDiscoverSessionHook(t *testing.T) {
	home := t.TempDir()
	grokHome := filepath.Join(home, "grok-home")
	workspace := home
	uuid := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	prompt := "stream probe"

	encoded, _ := EncodedCwd(workspace)
	dir := filepath.Join(grokHome, "sessions", encoded, uuid)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	summary := map[string]any{
		"info":       map[string]any{"cwd": workspace, "sessionId": uuid},
		"created_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
	b, _ := json.Marshal(summary)
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), b, 0644); err != nil {
		t.Fatal(err)
	}
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "user_message_chunk",
		"content":       map[string]any{"type": "text", "text": prompt},
	})
	if err := os.WriteFile(filepath.Join(dir, "updates.jsonl"), append(append(line, '\n'), '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GROK_HOME", grokHome)
	t.Setenv("AGENT_RUN_GROK_TTY_GROK_SESSION_ID", uuid)

	gotID, gotPath, err := DiscoverSession(context.Background(), grokHome, workspace, prompt, time.Now())
	if err != nil {
		t.Fatalf("DiscoverSession: %v (encoded=%s abs=%s)", err, encoded, mustAbs(workspace))
	}
	if gotID != uuid {
		t.Fatalf("id=%q want %q", gotID, uuid)
	}
	if _, err := os.Stat(gotPath); err != nil {
		t.Fatalf("updates path: %v", err)
	}
}

func mustAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs + " escaped=" + url.PathEscape(abs)
}

func TestEncodedCwdMatchesPathEscape(t *testing.T) {
	home := t.TempDir()
	got, err := EncodedCwd(home)
	if err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs(home)
	want := url.PathEscape(abs)
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFirstUserMessageChunkNestedWire(t *testing.T) {
	home := t.TempDir()
	updatesPath := filepath.Join(home, "updates.jsonl")
	line := `{"method":"session/update","params":{"sessionId":"550e8400-e29b-41d4-a716-446655440000","update":{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"run ls"}}}}`
	if err := os.WriteFile(updatesPath, []byte(line+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got, ok := firstUserMessageChunk(updatesPath)
	if !ok {
		t.Fatal("firstUserMessageChunk returned false for nested wire format")
	}
	if got != "run ls" {
		t.Fatalf("got %q want %q", got, "run ls")
	}
}

func TestDiscoverSessionNestedWireFormat(t *testing.T) {
	home := t.TempDir()
	grokHome := filepath.Join(home, "grok-home")
	workspace := home
	uuid := "33333333-3333-3333-3333-333333333333"
	prompt := "run ls"

	encoded, _ := EncodedCwd(workspace)
	dir := filepath.Join(grokHome, "sessions", encoded, uuid)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs(workspace)
	summary := map[string]any{
		"info":       map[string]any{"cwd": abs, "sessionId": uuid},
		"created_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
	b, _ := json.Marshal(summary)
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), b, 0644); err != nil {
		t.Fatal(err)
	}
	line := `{"timestamp":1782714542,"method":"session/update","params":{"sessionId":"` + uuid + `","update":{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"run ls"}}}}`
	if err := os.WriteFile(filepath.Join(dir, "updates.jsonl"), []byte(line+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	gotID, gotPath, err := DiscoverSession(context.Background(), grokHome, workspace, prompt, time.Now().Add(-time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if gotID != uuid {
		t.Fatalf("id=%q want %q", gotID, uuid)
	}
	if _, err := os.Stat(gotPath); err != nil {
		t.Fatalf("updates path: %v", err)
	}
}

func TestDiscoverFromActiveSessionsSnakeCase(t *testing.T) {
	home := t.TempDir()
	grokHome := filepath.Join(home, "grok-home")
	workspace := home
	uuid := "44444444-4444-4444-4444-444444444444"
	runStart := time.Now().Add(-time.Minute)

	encoded, _ := EncodedCwd(workspace)
	dir := filepath.Join(grokHome, "sessions", encoded, uuid)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs(workspace)
	summary := map[string]any{
		"info":       map[string]any{"cwd": abs, "sessionId": uuid},
		"created_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
	b, _ := json.Marshal(summary)
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), b, 0644); err != nil {
		t.Fatal(err)
	}
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "user_message_chunk",
		"content":       map[string]any{"type": "text", "text": "run ls"},
	})
	if err := os.WriteFile(filepath.Join(dir, "updates.jsonl"), append(append(line, '\n'), '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	active, _ := json.Marshal([]map[string]any{{
		"session_id": uuid,
		"cwd":        abs,
		"opened_at":  time.Now().UTC().Format(time.RFC3339Nano),
	}})
	if err := os.MkdirAll(grokHome, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(grokHome, "active_sessions.json"), active, 0644); err != nil {
		t.Fatal(err)
	}

	gotID, gotPath, ok := discoverFromActiveSessions(grokHome, workspace, "run ls", runStart)
	if !ok {
		t.Fatal("discoverFromActiveSessions returned false for snake_case active_sessions.json")
	}
	if gotID != uuid {
		t.Fatalf("id=%q want %q", gotID, uuid)
	}
	if _, err := os.Stat(gotPath); err != nil {
		t.Fatalf("updates path: %v", err)
	}
}

func TestDiscoverSessionRejectsPriorSessionSamePrompt(t *testing.T) {
	home := t.TempDir()
	grokHome := filepath.Join(home, "grok-home")
	workspace := home
	oldUUID := "11111111-1111-1111-1111-111111111111"
	newUUID := "22222222-2222-2222-2222-222222222222"
	runStart := time.Now()

	writeSession := func(uuid, prompt string, created time.Time) {
		encoded, _ := EncodedCwd(workspace)
		dir := filepath.Join(grokHome, "sessions", encoded, uuid)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		abs, _ := filepath.Abs(workspace)
		summary := map[string]any{
			"info":       map[string]any{"cwd": abs, "sessionId": uuid},
			"created_at": created.UTC().Format(time.RFC3339Nano),
		}
		b, _ := json.Marshal(summary)
		if err := os.WriteFile(filepath.Join(dir, "summary.json"), b, 0644); err != nil {
			t.Fatal(err)
		}
		line, _ := json.Marshal(map[string]any{
			"sessionUpdate": "user_message_chunk",
			"content":       map[string]any{"type": "text", "text": prompt},
		})
		if err := os.WriteFile(filepath.Join(dir, "updates.jsonl"), append(append(line, '\n'), '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}

	writeSession(oldUUID, "run ls", runStart.Add(-2*time.Hour))
	writeSession(newUUID, "run ls", runStart.Add(200*time.Millisecond))

	gotID, _, err := DiscoverSession(context.Background(), grokHome, workspace, "run ls", runStart)
	if err != nil {
		t.Fatal(err)
	}
	if gotID != newUUID {
		t.Fatalf("id=%q want %q (must not match prior session with same prompt)", gotID, newUUID)
	}
}

func TestDiscoverSessionByPrompt(t *testing.T) {
	home := t.TempDir()
	grokHome := filepath.Join(home, "grok-home")
	workspace := home
	wrongUUID := "11111111-1111-1111-1111-111111111111"
	rightUUID := "22222222-2222-2222-2222-222222222222"

	writeSession := func(uuid, prompt string) {
		encoded, _ := EncodedCwd(workspace)
		dir := filepath.Join(grokHome, "sessions", encoded, uuid)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		abs, _ := filepath.Abs(workspace)
		summary := map[string]any{
			"info":       map[string]any{"cwd": abs, "sessionId": uuid},
			"created_at": time.Now().UTC().Format(time.RFC3339Nano),
		}
		b, _ := json.Marshal(summary)
		_ = os.WriteFile(filepath.Join(dir, "summary.json"), b, 0644)
		line, _ := json.Marshal(map[string]any{
			"sessionUpdate": "user_message_chunk",
			"content":       map[string]any{"type": "text", "text": prompt},
		})
		_ = os.WriteFile(filepath.Join(dir, "updates.jsonl"), append(append(line, '\n'), '\n'), 0644)
	}

	writeSession(wrongUUID, "say wrong")
	writeSession(rightUUID, "run ls")

	gotID, _, err := DiscoverSession(context.Background(), grokHome, workspace, "run ls", time.Now().Add(-time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if gotID != rightUUID {
		t.Fatalf("id=%q want %q", gotID, rightUUID)
	}
}