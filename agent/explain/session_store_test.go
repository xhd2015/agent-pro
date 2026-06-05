package explain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindMatchingSession_EmptyStore(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	result, err := findMatchingSession([]string{"anything"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil, got %+v", result)
	}
}

func TestFindMatchingSession_SingleArgNeverMatches(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-a1b2c3d4", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_1"}),
		},
		Messages: []Message{
			{Role: "user", Message: "hello"},
			{Role: "assistant", Message: "world"},
		},
	})

	result, err := findMatchingSession([]string{"hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil (1 arg never matches), got %+v", result)
	}
}

func TestFindMatchingSession_TwoArgsExactMatch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-a1b2c3d4", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_1"}),
		},
		Messages: []Message{
			{Role: "user", Message: "A F"},
			{Role: "assistant", Message: "answer"},
		},
	})

	result, err := findMatchingSession([]string{"A F", "B"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected match, got nil")
	}
	if result.MatchedCount != 1 {
		t.Fatalf("expected MatchedCount=1, got %d", result.MatchedCount)
	}
}

func TestFindMatchingSession_ElementMismatch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-a1b2c3d4", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_1"}),
		},
		Messages: []Message{
			{Role: "user", Message: "A"},
			{Role: "assistant", Message: "answer"},
		},
	})

	result, err := findMatchingSession([]string{"A F", "B"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil (\"A\" != \"A F\"), got %+v", result)
	}
}

func TestFindMatchingSession_ThreeArgsTwoMatch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-a1b2c3d4", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_1"}),
		},
		Messages: []Message{
			{Role: "user", Message: "A"},
			{Role: "assistant", Message: "ans1"},
			{Role: "user", Message: "B"},
			{Role: "assistant", Message: "ans2"},
		},
	})

	result, err := findMatchingSession([]string{"A", "B", "C"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected match, got nil")
	}
	if result.MatchedCount != 2 {
		t.Fatalf("expected MatchedCount=2, got %d", result.MatchedCount)
	}
}

func TestFindMatchingSession_NewerWinsOnTie(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-older", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_old"}),
		},
		Messages: []Message{
			{Role: "user", Message: "A"},
			{Role: "assistant", Message: "old"},
		},
	})

	createSessionDir(t, home, "2026-06-05-14-30-10-newer", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_new"}),
		},
		Messages: []Message{
			{Role: "user", Message: "A"},
			{Role: "assistant", Message: "new"},
		},
	})

	result, err := findMatchingSession([]string{"A", "X"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected match, got nil")
	}

	var opencodeMeta struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(result.Data.AgentRunnersMeta["opencode"], &opencodeMeta); err != nil {
		t.Fatalf("unmarshal opencode meta: %v", err)
	}
	if opencodeMeta.SessionID != "sess_new" {
		t.Fatalf("expected sess_new (newer), got %s", opencodeMeta.SessionID)
	}
}

func TestFindMatchingSession_LongerPrefixWins(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-short", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_short"}),
		},
		Messages: []Message{
			{Role: "user", Message: "A"},
			{Role: "assistant", Message: "short"},
		},
	})

	createSessionDir(t, home, "2026-06-05-14-30-10-long", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_long"}),
		},
		Messages: []Message{
			{Role: "user", Message: "A"},
			{Role: "assistant", Message: "ans"},
			{Role: "user", Message: "B"},
			{Role: "assistant", Message: "ans"},
		},
	})

	result, err := findMatchingSession([]string{"A", "B", "C"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected match, got nil")
	}
	if result.MatchedCount != 2 {
		t.Fatalf("expected MatchedCount=2 (longer prefix wins), got %d", result.MatchedCount)
	}

	var opencodeMeta struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(result.Data.AgentRunnersMeta["opencode"], &opencodeMeta); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if opencodeMeta.SessionID != "sess_long" {
		t.Fatalf("expected sess_long, got %s", opencodeMeta.SessionID)
	}
}

func TestFindMatchingSession_NoMatchUnrelated(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-test", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_1"}),
		},
		Messages: []Message{
			{Role: "user", Message: "X"},
			{Role: "assistant", Message: "ans"},
		},
	})

	result, err := findMatchingSession([]string{"A", "B"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil, got %+v", result)
	}
}

func TestCountPrefixMatch(t *testing.T) {
	tests := []struct {
		stored   []string
		input    []string
		expected int
	}{
		{[]string{"A"}, []string{"A"}, 0},
		{[]string{"A"}, []string{"A", "B"}, 1},
		{[]string{"A", "B"}, []string{"A", "B", "C"}, 2},
		{[]string{"A"}, []string{"A F", "B"}, 0},
		{[]string{"A F"}, []string{"A F", "B"}, 1},
		{[]string{"A", "B"}, []string{"A", "X"}, 1},
		{[]string{"A", "B"}, []string{"A", "C"}, 1},
		{[]string{}, []string{"A"}, 0},
		{[]string{"A", "B"}, []string{"A", "B"}, 0},
	}

	for _, tt := range tests {
		got := countPrefixMatch(tt.stored, tt.input)
		if got != tt.expected {
			t.Fatalf("countPrefixMatch(%v, %v) = %d, want %d", tt.stored, tt.input, got, tt.expected)
		}
	}
}

func TestSaveSession(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	data := SessionData{
		AgentRunner: "opencode",
		Model:       "deepseek",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_1"}),
		},
		Messages: []Message{
			{Role: "user", Message: "run的过去式"},
			{Role: "assistant", Message: "ran"},
		},
	}

	dir, err := saveSession("run的过去式", data)
	if err != nil {
		t.Fatalf("saveSession failed: %v", err)
	}
	if dir == "" {
		t.Fatal("expected non-empty dir")
	}

	readData, err := readSession(dir)
	if err != nil {
		t.Fatalf("readSession failed: %v", err)
	}
	if readData.AgentRunner != "opencode" {
		t.Fatalf("expected opencode, got %s", readData.AgentRunner)
	}
	if readData.Model != "deepseek" {
		t.Fatalf("expected deepseek, got %s", readData.Model)
	}
	if len(readData.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(readData.Messages))
	}
}

func TestUpdateSession(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	data := SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_1"}),
		},
		Messages: []Message{
			{Role: "user", Message: "hello"},
		},
	}

	dir, err := saveSession("hello", data)
	if err != nil {
		t.Fatalf("saveSession failed: %v", err)
	}

	read, err := readSession(dir)
	if err != nil {
		t.Fatalf("readSession failed: %v", err)
	}
	read.Messages = append(read.Messages, Message{Role: "assistant", Message: "world"})
	if err := updateSession(dir, read); err != nil {
		t.Fatalf("updateSession failed: %v", err)
	}

	readAgain, err := readSession(dir)
	if err != nil {
		t.Fatalf("readSession failed: %v", err)
	}
	if len(readAgain.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(readAgain.Messages))
	}
}

func TestDebugConfigHomeEnv(t *testing.T) {
	debugHome := t.TempDir()
	t.Setenv(debugConfigHomeEnv, debugHome)
	t.Setenv("HOME", "/nonexistent")

	dir, err := sessionsDir()
	if err != nil {
		t.Fatalf("sessionsDir failed: %v", err)
	}
	expected := filepath.Join(debugHome, "sessions")
	if dir != expected {
		t.Fatalf("expected %s, got %s", expected, dir)
	}
}

func TestMakeSessionName_HashSuffix(t *testing.T) {
	n1 := makeSessionName("run的各种形态")
	n2 := makeSessionName("run的过去式")

	if n1 == n2 {
		t.Fatalf("expected different names, both got %q", n1)
	}
	if n1[:19] != n2[:19] {
		t.Fatalf("expected same timestamp prefix, got %q and %q", n1[:19], n2[:19])
	}
}

func TestSlugFromPrompt(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "hello-world"},
		{"run的过去式和各种形态", "run"},
		{"Python 3.12 新特性", "python-3-12"},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := slugFromPrompt(tt.input)
			if got != tt.expected {
				t.Fatalf("slugFromPrompt(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMakeSessionName(t *testing.T) {
	name := makeSessionName("hello")
	if len(name) < 20 {
		t.Fatalf("name too short: %s", name)
	}
	ts := time.Now().Format("2006-01-02-15-04-05")
	if name[:19] != ts {
		t.Fatalf("expected timestamp prefix %s, got %s", ts, name[:19])
	}
}

func TestUserMessageSlice(t *testing.T) {
	data := SessionData{
		Messages: []Message{
			{Role: "user", Message: "A"},
			{Role: "assistant", Message: "ans"},
			{Role: "user", Message: "B"},
			{Role: "assistant", Message: "ans"},
			{Role: "system", Message: "sys"},
		},
	}
	msgs := userMessageSlice(data)
	if len(msgs) != 2 {
		t.Fatalf("expected 2 user messages, got %d", len(msgs))
	}
	if msgs[0] != "A" || msgs[1] != "B" {
		t.Fatalf("unexpected messages: %v", msgs)
	}
}

func createSessionDir(t *testing.T, home, dirName string, data SessionData) {
	t.Helper()
	baseDir := filepath.Join(home, defaultSessionsBaseDir, "sessions")
	dir := filepath.Join(baseDir, dirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("create test dir: %v", err)
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, fileName), bytes, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
