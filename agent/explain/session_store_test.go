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

	result, err := findMatchingSession("anything")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil, got %+v", result)
	}
}

func TestFindMatchingSession_StrictPrefixMatch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-run-de-guo-qu", SessionData{
		AgentRunner: "opencode",
		Model:       "deepseek",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_1"}),
		},
		Messages: []Message{
			{Role: "user", Message: "run的过去式"},
			{Role: "assistant", Message: "ran"},
		},
	})

	result, err := findMatchingSession("run的过去式和其他")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected match, got nil")
	}
	if len(allUserMessages(result.Data)) != 1 || allUserMessages(result.Data)[0] != "run的过去式" {
		t.Fatalf("unexpected messages: %v", result.Data.Messages)
	}
}

func TestFindMatchingSession_ExactMatchNoMatch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-run-de-guo-qu", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_1"}),
		},
		Messages: []Message{
			{Role: "user", Message: "run的过去式"},
			{Role: "assistant", Message: "ran"},
		},
	})

	result, err := findMatchingSession("run的过去式")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil (exact match not strict), got %+v", result)
	}
}

func TestFindMatchingSession_LongestPrefix(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-short-match", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_short"}),
		},
		Messages: []Message{
			{Role: "user", Message: "run的"},
			{Role: "assistant", Message: "short answer"},
		},
	})

	createSessionDir(t, home, "2026-06-05-14-30-10-long-match", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_long"}),
		},
		Messages: []Message{
			{Role: "user", Message: "run的过去式"},
			{Role: "assistant", Message: "long answer"},
		},
	})

	result, err := findMatchingSession("run的过去式是")
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
	if opencodeMeta.SessionID != "sess_long" {
		t.Fatalf("expected sess_long (longer prefix), got %s", opencodeMeta.SessionID)
	}
}

func TestFindMatchingSession_SameLengthNewer(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-older", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_old"}),
		},
		Messages: []Message{
			{Role: "user", Message: "run的"},
			{Role: "assistant", Message: "older"},
		},
	})

	createSessionDir(t, home, "2026-06-05-14-30-10-newer", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_new"}),
		},
		Messages: []Message{
			{Role: "user", Message: "run的"},
			{Role: "assistant", Message: "newer"},
		},
	})

	result, err := findMatchingSession("run的过")
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

func TestFindMatchingSession_UserRoleOnly(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-test", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_1"}),
		},
		Messages: []Message{
			{Role: "user", Message: "hello"},
			{Role: "assistant", Message: "and more text that happens to contain run的过去式"},
		},
	})

	result, err := findMatchingSession("run的过去式")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil (assistant messages should not match), got %+v", result)
	}
}

func TestFindMatchingSession_NoMatch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-test", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_1"}),
		},
		Messages: []Message{
			{Role: "user", Message: "run的过去式"},
			{Role: "assistant", Message: "ran"},
		},
	})

	result, err := findMatchingSession("python的用法")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil, got %+v", result)
	}
}

func TestFindMatchingSession_MultipleUserMessages(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	createSessionDir(t, home, "2026-06-05-14-30-00-test", SessionData{
		AgentRunner: "opencode",
		AgentRunnersMeta: RunnerMeta{
			"opencode": mustMarshalJSON(map[string]string{"session_id": "sess_1"}),
		},
		Messages: []Message{
			{Role: "user", Message: "run的过去式"},
			{Role: "assistant", Message: "ran"},
			{Role: "user", Message: "run的过去式和各种形态"},
			{Role: "assistant", Message: "ran, running, runs"},
		},
	})

	result, err := findMatchingSession("run的过去式和各种形态")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected strict prefix match via first user msg, got nil")
	}

	userMsgs := allUserMessages(result.Data)
	if len(userMsgs) < 2 {
		t.Fatalf("expected at least 2 user messages, got %d", len(userMsgs))
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
	if readData.Messages[0].Role != "user" || readData.Messages[0].Message != "run的过去式" {
		t.Fatalf("unexpected first message: %+v", readData.Messages[0])
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
		t.Fatalf("expected 2 messages after update, got %d", len(readAgain.Messages))
	}
	if readAgain.Messages[1].Role != "assistant" || readAgain.Messages[1].Message != "world" {
		t.Fatalf("unexpected second message: %+v", readAgain.Messages[1])
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

func TestMakeSessionName_Uniqueness(t *testing.T) {
	n1 := makeSessionName("run的各种形态")
	n2 := makeSessionName("run的过去式")

	if n1 == n2 {
		t.Fatalf("expected different names for different prompts, both got %q", n1)
	}
	if n1[:19] != n2[:19] {
		t.Fatalf("expected same timestamp prefix, got %q and %q", n1[:19], n2[:19])
	}
}

func createSessionDir(t *testing.T, home, dirName string, data SessionData) {
	t.Helper()
	baseDir := filepath.Join(home, sessionsBaseDir)
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
