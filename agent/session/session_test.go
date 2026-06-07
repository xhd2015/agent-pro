package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDirCreatesAndReturnsPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(homeEnv, tmp)

	dir, err := Dir("my-agent", "session-1")
	if err != nil {
		t.Fatalf("Dir: %v", err)
	}

	if !strings.HasSuffix(dir, filepath.Join("my-agent", "sessions", "session-1")) {
		t.Errorf("unexpected dir path: %s", dir)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if !info.IsDir() {
		t.Error("Dir should create a directory")
	}
}

func TestDirHomeFallback(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	os.Unsetenv(homeEnv)

	dir, err := Dir("test-agent", "abc")
	if err != nil {
		t.Fatalf("Dir: %v", err)
	}

	if !strings.Contains(dir, filepath.Join(".agent-pro", "dedicated-agents", "test-agent", "sessions", "abc")) {
		t.Errorf("unexpected dir path: %s", dir)
	}
}

func TestWriteReadJSON(t *testing.T) {
	dir := t.TempDir()

	type meta struct {
		SessionID string `json:"session_id"`
		Feature   string `json:"feature"`
		Model     string `json:"model"`
	}
	original := meta{
		SessionID: "ses_123",
		Feature:   "Test feature",
		Model:     "gpt-4o",
	}

	if err := WriteJSON(dir, "metadata.json", original); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var loaded meta
	if err := ReadJSON(dir, "metadata.json", &loaded); err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}

	if loaded.SessionID != "ses_123" || loaded.Feature != "Test feature" || loaded.Model != "gpt-4o" {
		t.Errorf("round-trip mismatch: %+v", loaded)
	}
}

func TestAppendReadLines(t *testing.T) {
	dir := t.TempDir()

	if err := AppendLine(dir, "events.jsonl", `{"type":"text","text":"hello"}`); err != nil {
		t.Fatalf("AppendLine 1: %v", err)
	}
	if err := AppendLine(dir, "events.jsonl", `{"type":"text","text":"world"}`); err != nil {
		t.Fatalf("AppendLine 2: %v", err)
	}

	lines, err := ReadLines(dir, "events.jsonl")
	if err != nil {
		t.Fatalf("ReadLines: %v", err)
	}

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != `{"type":"text","text":"hello"}` {
		t.Errorf("unexpected line 0: %s", lines[0])
	}
	if lines[1] != `{"type":"text","text":"world"}` {
		t.Errorf("unexpected line 1: %s", lines[1])
	}
}

func TestReadLinesNonExistent(t *testing.T) {
	dir := t.TempDir()
	lines, err := ReadLines(dir, "nonexistent.jsonl")
	if err != nil {
		t.Fatalf("ReadLines: %v", err)
	}
	if lines != nil {
		t.Errorf("expected nil for non-existent file, got %v", lines)
	}
}

func TestReadJSONNonExistent(t *testing.T) {
	dir := t.TempDir()
	var v struct{ X string }
	err := ReadJSON(dir, "nonexistent.json", &v)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestDirMultipleAgents(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(homeEnv, tmp)

	dir1, _ := Dir("agent-a", "s1")
	dir2, _ := Dir("agent-b", "s2")

	if dir1 == dir2 {
		t.Error("different agents should have different session dirs")
	}
	if !strings.Contains(dir1, "agent-a") || !strings.Contains(dir2, "agent-b") {
		t.Error("agent names should appear in paths")
	}
}
