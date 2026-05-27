package trace

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAgentTraceSessionStoresPromptAndParsesMessages(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dataDir := t.TempDir()
	session, err := StartAgentTraceSession(dataDir, AgentTraceMetadata{
		Command:     "find-prds",
		CommandArgs: []string{"knowledge-portal", "find-prds", "spl/biz-scene/cashout"},
		ProviderID:  "codex",
		Model:       "gpt-5.4",
	}, "test prompt")
	if err != nil {
		t.Fatalf("start trace session: %v", err)
	}
	_ = home
	if !strings.HasPrefix(session.Dir(), filepath.Join(dataDir, "agent-traces")) {
		t.Fatalf("trace dir = %q, want under data dir agent-traces", session.Dir())
	}
	_, err = session.Writer().Write([]byte(`{"type":"item.started","item":{"id":"item_1","type":"command_execution","command":"echo ok","status":"in_progress"}}` + "\n"))
	if err != nil {
		t.Fatalf("write started event: %v", err)
	}
	_, err = session.Writer().Write([]byte(`{"type":"item.completed","item":{"id":"item_1","type":"command_execution","command":"echo ok","exit_code":0,"aggregated_output":"ok\n"}}` + "\n"))
	if err != nil {
		t.Fatalf("write completed event: %v", err)
	}
	_, err = session.Writer().Write([]byte(`{"type":"item.completed","item":{"id":"item_2","type":"agent_message","text":"done"}}` + "\n"))
	if err != nil {
		t.Fatalf("write message event: %v", err)
	}
	session.Finish(nil)

	detail, err := loadAgentTraceDetail(dataDir, session.ID())
	if err != nil {
		t.Fatalf("load detail: %v", err)
	}
	if detail.Metadata.Status != "completed" {
		t.Fatalf("status = %q, want completed", detail.Metadata.Status)
	}
	if detail.Prompt != "test prompt" {
		t.Fatalf("prompt = %q", detail.Prompt)
	}
	if len(detail.RawLines) != 3 {
		t.Fatalf("raw lines = %d, want 3", len(detail.RawLines))
	}
	if len(detail.Messages) != 2 {
		t.Fatalf("messages = %d, want 2", len(detail.Messages))
	}
	if detail.Messages[0].ToolCall == nil || detail.Messages[0].ToolCall.ToolName != "Shell" {
		t.Fatalf("first message did not parse shell tool call: %#v", detail.Messages[0])
	}
	if !strings.Contains(detail.Messages[0].ToolCall.Summary, "ok") {
		t.Fatalf("tool summary missing output: %q", detail.Messages[0].ToolCall.Summary)
	}
	if detail.Messages[1].Role != "assistant" || detail.Messages[1].Content != "done" {
		t.Fatalf("assistant message mismatch: %#v", detail.Messages[1])
	}
}

func TestAgentTraceDetailSerializesEmptyRawLinesAsArray(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dataDir := t.TempDir()
	session, err := StartAgentTraceSession(dataDir, AgentTraceMetadata{
		Command: "empty-trace",
	}, "test prompt")
	if err != nil {
		t.Fatalf("start trace session: %v", err)
	}
	session.Finish(nil)

	detail, err := loadAgentTraceDetail(dataDir, session.ID())
	if err != nil {
		t.Fatalf("load detail: %v", err)
	}
	data, err := json.Marshal(detail)
	if err != nil {
		t.Fatalf("marshal detail: %v", err)
	}
	out := string(data)
	for _, bad := range []string{
		`"messages":null`,
		`"raw_lines":null`,
	} {
		if strings.Contains(out, bad) {
			t.Fatalf("detail JSON contains %s: %s", bad, out)
		}
	}
}

func TestResumeAgentTraceSessionReopensExistingTraceAsRunning(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dataDir := t.TempDir()
	session, err := StartAgentTraceSession(dataDir, AgentTraceMetadata{
		Command:       "design-td",
		CommandArgs:   []string{"design-td", "initial"},
		Workspace:     "/tmp/design-session",
		ResumeCommand: "design-td --resume session 'your follow-up or clarification'",
		ProviderID:    "codex",
	}, "test prompt")
	if err != nil {
		t.Fatalf("start trace session: %v", err)
	}
	if _, err := session.Writer().Write([]byte(`{"type":"item.completed","item":{"id":"item_1","type":"agent_message","text":"initial"}}` + "\n")); err != nil {
		t.Fatalf("write initial event: %v", err)
	}
	session.Finish(nil)

	resumed, err := ResumeAgentTraceSession(session.Dir(), AgentTraceMetadata{
		Command:       "design-td",
		CommandArgs:   []string{"design-td", "--resume", "session", "option 1"},
		Workspace:     "/tmp/design-session",
		ResumeCommand: "design-td --resume session 'your follow-up or clarification'",
		ProviderID:    "codex",
		Model:         "gpt-5",
	})
	if err != nil {
		t.Fatalf("resume trace session: %v", err)
	}
	detail, err := loadAgentTraceDetail(dataDir, session.ID())
	if err != nil {
		t.Fatalf("load running detail: %v", err)
	}
	if detail.Metadata.Status != "running" {
		t.Fatalf("status = %q, want running", detail.Metadata.Status)
	}
	if detail.Metadata.CommandLine != "design-td --resume session option 1" {
		t.Fatalf("command line = %q", detail.Metadata.CommandLine)
	}
	if detail.Metadata.ResumeCommand != "design-td --resume session 'your follow-up or clarification'" {
		t.Fatalf("resume command = %q", detail.Metadata.ResumeCommand)
	}

	if _, err := resumed.Writer().Write([]byte(`{"type":"item.completed","item":{"id":"item_2","type":"agent_message","text":"resumed"}}` + "\n")); err != nil {
		t.Fatalf("write resumed event: %v", err)
	}
	resumed.Finish(nil)
	detail, err = loadAgentTraceDetail(dataDir, session.ID())
	if err != nil {
		t.Fatalf("load completed detail: %v", err)
	}
	if detail.Metadata.Status != "completed" {
		t.Fatalf("status = %q, want completed", detail.Metadata.Status)
	}
	if len(detail.RawLines) != 2 {
		t.Fatalf("raw lines = %d, want 2", len(detail.RawLines))
	}
	if len(detail.Messages) != 2 || detail.Messages[1].Content != "resumed" {
		t.Fatalf("messages = %#v, want appended resumed message", detail.Messages)
	}
}

func TestAgentTraceStreamSendsDetailAndDone(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dataDir := t.TempDir()
	session, err := StartAgentTraceSession(dataDir, AgentTraceMetadata{
		Command: "find-prds",
	}, "test prompt")
	if err != nil {
		t.Fatalf("start trace session: %v", err)
	}
	if _, err := session.Writer().Write([]byte(`{"type":"item.completed","item":{"id":"item_1","type":"agent_message","text":"done"}}` + "\n")); err != nil {
		t.Fatalf("write event: %v", err)
	}
	session.Finish(nil)

	store := NewStore(dataDir)
	req := httptest.NewRequest(http.MethodGet, "/api/knowledge/agent-traces/"+session.ID()+"/stream", nil)
	rr := httptest.NewRecorder()
	store.HandleRoutes(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, want := range []string{
		"event: detail",
		"event: done",
		`"status":"completed"`,
		`"content":"done"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("stream body missing %q: %s", want, body)
		}
	}
}

func TestAgentTraceRunningIdleShowsNotRespondingTag(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dataDir := t.TempDir()
	session, err := StartAgentTraceSession(dataDir, AgentTraceMetadata{
		Command: "idle-agent",
	}, "test prompt")
	if err != nil {
		t.Fatalf("start trace session: %v", err)
	}
	if _, err := session.Writer().Write([]byte(`{"type":"item.completed","item":{"id":"item_1","type":"agent_message","text":"still working"}}` + "\n")); err != nil {
		t.Fatalf("write event: %v", err)
	}
	old := time.Now().Add(-traceNotRespondingAfter - time.Minute)
	if err := os.Chtimes(session.meta.LogPath, old, old); err != nil {
		t.Fatalf("age trace log: %v", err)
	}

	detail, err := loadAgentTraceDetail(dataDir, session.ID())
	if err != nil {
		t.Fatalf("load detail: %v", err)
	}
	if !hasTraceTag(detail.Metadata.Tags, traceNotRespondingTag) {
		t.Fatalf("detail tags = %#v, want %q", detail.Metadata.Tags, traceNotRespondingTag)
	}
	summaries, err := loadAgentTraceSummaries(dataDir)
	if err != nil {
		t.Fatalf("load summaries: %v", err)
	}
	if len(summaries) != 1 || !hasTraceTag(summaries[0].Tags, traceNotRespondingTag) {
		t.Fatalf("summary tags = %#v, want %q", summaries, traceNotRespondingTag)
	}
}

func TestAgentTraceStopRouteMarksRunningSessionStopped(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dataDir := t.TempDir()
	session, err := StartAgentTraceSession(dataDir, AgentTraceMetadata{
		Command: "stale-agent",
	}, "test prompt")
	if err != nil {
		t.Fatalf("start trace session: %v", err)
	}
	if _, err := session.Writer().Write([]byte(`{"type":"item.completed","item":{"id":"item_1","type":"agent_message","text":"started"}}` + "\n")); err != nil {
		t.Fatalf("write event: %v", err)
	}

	store := NewStore(dataDir)
	req := httptest.NewRequest(http.MethodPost, "/api/knowledge/agent-traces/"+session.ID()+"/stop", nil)
	rr := httptest.NewRecorder()
	store.HandleRoutes(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	var detail AgentTraceDetail
	if err := json.Unmarshal(rr.Body.Bytes(), &detail); err != nil {
		t.Fatalf("decode detail: %v", err)
	}
	if detail.Metadata.Status != traceStatusStopped {
		t.Fatalf("status = %q, want %q", detail.Metadata.Status, traceStatusStopped)
	}
	if hasTraceTag(detail.Metadata.Tags, traceNotRespondingTag) {
		t.Fatalf("stopped detail tags = %#v, did not want %q", detail.Metadata.Tags, traceNotRespondingTag)
	}
	meta, err := readAgentTraceMetadata(session.Dir())
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	if meta.Status != traceStatusStopped {
		t.Fatalf("stored status = %q, want %q", meta.Status, traceStatusStopped)
	}
}

func TestAgentTraceDeleteRouteRemovesSession(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dataDir := t.TempDir()
	session, err := StartAgentTraceSession(dataDir, AgentTraceMetadata{
		Command: "delete-agent",
	}, "test prompt")
	if err != nil {
		t.Fatalf("start trace session: %v", err)
	}
	if _, err := session.Writer().Write([]byte(`{"type":"item.completed","item":{"id":"item_1","type":"agent_message","text":"temporary"}}` + "\n")); err != nil {
		t.Fatalf("write event: %v", err)
	}
	session.Finish(nil)

	store := NewStore(dataDir)
	req := httptest.NewRequest(http.MethodDelete, "/api/knowledge/agent-traces/"+session.ID(), nil)
	rr := httptest.NewRecorder()
	store.HandleRoutes(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	if _, err := os.Stat(session.Dir()); !os.IsNotExist(err) {
		t.Fatalf("session dir still exists or stat failed unexpectedly: %v", err)
	}
	summaries, err := loadAgentTraceSummaries(dataDir)
	if err != nil {
		t.Fatalf("load summaries: %v", err)
	}
	if len(summaries) != 0 {
		t.Fatalf("summaries = %#v, want empty after delete", summaries)
	}
}

func hasTraceTag(tags []string, want string) bool {
	for _, tag := range tags {
		if tag == want {
			return true
		}
	}
	return false
}
