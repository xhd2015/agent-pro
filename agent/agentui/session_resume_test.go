package agentui

import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/session"
)

func TestLoadQuestionsFromFile(t *testing.T) {
	dir := t.TempDir()
	session.AppendLine(dir, "questions.jsonl", `{"type":"question","id":"1","question":"Q1","options":[{"label":"A"},{"label":"B"}]}`)
	session.AppendLine(dir, "questions.jsonl", `{"type":"question","id":"2","question":"Q2"}`)
	session.AppendLine(dir, "questions.jsonl", `{"type":"answer","id":"1","answer":"Answer 1"}`)

	m := &model{
		sessionDir: dir,
		logs:       []string{},
	}
	m.loadQuestionsFromFile()

	if len(m.questions) != 2 {
		t.Fatalf("expected 2 questions loaded, got %d", len(m.questions))
	}
	if m.questions[0].ID != "1" || m.questions[0].Question != "Q1" {
		t.Errorf("Q1 not loaded correctly: %+v", m.questions[0])
	}
	if m.questions[0].Answer != "Answer 1" {
		t.Errorf("expected Q1 answer 'Answer 1', got %q", m.questions[0].Answer)
	}
	if m.questions[1].Answer != "" {
		t.Errorf("expected Q2 answer empty, got %q", m.questions[1].Answer)
	}
	if len(m.questions[0].Options) != 2 {
		t.Errorf("expected 2 options for Q1, got %d", len(m.questions[0].Options))
	}
}

func TestLoadQuestionsFromFileNoFile(t *testing.T) {
	m := &model{
		sessionDir: "/nonexistent",
		logs:       []string{},
	}
	m.loadQuestionsFromFile()
	if len(m.questions) != 0 {
		t.Errorf("expected 0 questions for missing file, got %d", len(m.questions))
	}
}

func TestNewSessionIDFormat(t *testing.T) {
	id := newSessionID("tcd_")
	if !strings.HasPrefix(id, "tcd_") {
		t.Errorf("session ID should start with 'tcd_', got %q", id)
	}
	if len(id) < 20 {
		t.Errorf("session ID too short: %s", id)
	}
	id2 := newSessionID("tcd_")
	if id == id2 {
		t.Error("newSessionID should produce unique IDs")
	}
}

func TestReadSessionFromDir(t *testing.T) {
	dir := t.TempDir()

	meta := sessionMeta{SessionID: "sid_1", Feature: "Test feature", Model: "gpt-4o"}
	session.WriteJSON(dir, "metadata.json", meta)
	session.AppendLine(dir, "events.jsonl", `{"type":"text","timestamp":1,"sessionID":"sid_1","part":{"text":"hello"}}`)
	session.AppendLine(dir, "events.jsonl", `{"type":"tool_use","timestamp":2,"sessionID":"sid_1","part":{"tool":"bash","state":{"status":"completed","title":"ls"}}}`)

	sid, _, feat, model, logs := readSessionFromDir(dir, "sid_1")
	if sid != "sid_1" {
		t.Errorf("expected session ID 'sid_1', got %q", sid)
	}
	if feat != "Test feature" {
		t.Errorf("expected feature 'Test feature', got %q", feat)
	}
	if model != "gpt-4o" {
		t.Errorf("expected model 'gpt-4o', got %q", model)
	}
	if len(logs) < 1 {
		t.Error("expected at least one formatted log entry")
	}
}

func TestReadSessionFromDirEmptyEvents(t *testing.T) {
	dir := t.TempDir()
	meta := sessionMeta{SessionID: "sid_2", Feature: "X", Model: "Y"}
	session.WriteJSON(dir, "metadata.json", meta)

	sid, _, feat, model, logs := readSessionFromDir(dir, "sid_2")
	if sid != "sid_2" || feat != "X" || model != "Y" {
		t.Errorf("unexpected values: sid=%s feat=%s model=%s", sid, feat, model)
	}
	if logs != nil {
		t.Error("expected nil logs for empty events file")
	}
}

func TestReadSessionFromDirNonExistent(t *testing.T) {
	sid, _, feat, model, logs := readSessionFromDir("/nonexistent/path", "")
	if sid != "" || feat != "" || model != "" || logs != nil {
		t.Error("expected empty result for non-existent dir")
	}
}

func TestReadSessionFromDirFallbackSessionID(t *testing.T) {
	dir := t.TempDir()
	meta := sessionMeta{SessionID: "", Feature: "F", Model: "M"}
	session.WriteJSON(dir, "metadata.json", meta)

	sid, _, _, _, _ := readSessionFromDir(dir, "fallback-id")
	if sid != "fallback-id" {
		t.Errorf("expected fallback session ID 'fallback-id', got %q", sid)
	}
}

func TestResolveSessionResumes(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("AGENT_PRO_HOME", tmp)

	resumeID := "tcd_abc123"
	dir, err := session.Dir("test-case-design-expert", resumeID)
	if err != nil {
		t.Fatalf("create session dir: %v", err)
	}
	session.WriteJSON(dir, "metadata.json", sessionMeta{
		SessionID: resumeID,
		Feature:   "Add dark mode",
		Model:     "gpt-4o",
	})
	session.AppendLine(dir, "events.jsonl", `{"type":"text","timestamp":1,"sessionID":"tcd_abc123","part":{"text":"hello"}}`)

	sid, _, sdir, feat, model, logs, err := resolveSession("test-case-design-expert", resumeID)
	if err != nil {
		t.Fatalf("resolveSession: %v", err)
	}
	if sid != resumeID {
		t.Errorf("expected session ID %q, got %q", resumeID, sid)
	}
	if sdir != dir {
		t.Errorf("expected dir %q, got %q", dir, sdir)
	}
	if feat != "Add dark mode" {
		t.Errorf("expected feature 'Add dark mode', got %q", feat)
	}
	if model != "gpt-4o" {
		t.Errorf("expected model 'gpt-4o', got %q", model)
	}
	if len(logs) == 0 {
		t.Error("expected at least one log entry")
	}
}

func TestResolveSessionNotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("AGENT_PRO_HOME", tmp)

	_, _, _, _, _, _, err := resolveSession("test-agent", "nonexistent_id")
	if err == nil {
		t.Error("expected error for non-existent session")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestResolveSessionEmptyID(t *testing.T) {
	sid, _, _, feat, model, _, err := resolveSession("test-agent", "")
	if err != nil {
		t.Fatalf("resolveSession empty: %v", err)
	}
	if sid != "" {
		t.Errorf("expected empty session ID, got %q", sid)
	}
	if feat != "" || model != "" {
		t.Error("expected empty feature and model for empty resume ID")
	}
}

func TestResolveSessionModelOverride(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("AGENT_PRO_HOME", tmp)

	resumeID := "tcd_xyz"
	dir, err := session.Dir("test-case-design-expert", resumeID)
	if err != nil {
		t.Fatalf("create session dir: %v", err)
	}
	session.WriteJSON(dir, "metadata.json", sessionMeta{
		SessionID: resumeID,
		Feature:   "Feature X",
		Model:     "claude-3",
	})

	_, _, _, _, model, _, err := resolveSession("test-case-design-expert", resumeID)
	if err != nil {
		t.Fatalf("resolveSession: %v", err)
	}
	if model != "claude-3" {
		t.Errorf("expected model from metadata 'claude-3', got %q", model)
	}
}
