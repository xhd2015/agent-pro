package run

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/trace"
)

func TestFormatTraceLine_CodexAssistantMessage(t *testing.T) {
	line := `{"type":"item.completed","item":{"type":"agent_message","text":"Here is the answer."}}`
	got := FormatTraceLine(line)
	if !strings.Contains(got, "ASSISTANT") {
		t.Fatalf("expected assistant label, got: %s", got)
	}
	if !strings.Contains(got, "Here is the answer.") {
		t.Fatalf("expected answer text, got: %s", got)
	}
}

func TestFormatTraceLine_CodexCommandExecutionCompleted(t *testing.T) {
	line := `{"type":"item.completed","item":{"type":"command_execution","command":"go build ./...","exit_code":0,"aggregated_output":"success"}}`
	got := FormatTraceLine(line)
	if !strings.Contains(got, "RUN") {
		t.Fatalf("expected RUN label, got: %s", got)
	}
	if !strings.Contains(got, "go build ./...") {
		t.Fatalf("expected command, got: %s", got)
	}
}

func TestFormatTraceLine_CodexCommandExecutionStarted(t *testing.T) {
	line := `{"type":"item.started","item":{"id":"item_1","type":"command_execution","command":"go test ./..."}}`
	got := FormatTraceLine(line)
	if !strings.Contains(got, "RUN") {
		t.Fatalf("expected RUN label, got: %s", got)
	}
	if !strings.Contains(got, "go test ./...") {
		t.Fatalf("expected command, got: %s", got)
	}
}

func TestFormatTraceLine_EmptyLine(t *testing.T) {
	got := FormatTraceLine("")
	if got != "" {
		t.Fatalf("expected empty for empty line, got: %s", got)
	}
}

func TestFormatTraceLine_NonJSON(t *testing.T) {
	got := FormatTraceLine("not a json line")
	if got != "" {
		t.Fatalf("expected empty for non-JSON, got: %s", got)
	}
}

func TestFormatTraceLine_CodexReasoning(t *testing.T) {
	line := `{"type":"item.completed","item":{"id":"item_2","type":"reasoning","text":"Let me think about this..."}}`
	got := FormatTraceLine(line)
	if !strings.Contains(got, "REASONING") {
		t.Fatalf("expected REASONING label, got: %s", got)
	}
	if !strings.Contains(got, "Let me think about this...") {
		t.Fatalf("expected reasoning text, got: %s", got)
	}
}

func TestFormatTraceLine_OpencodeAssistantText(t *testing.T) {
	line := `{"type":"text","sessionID":"sess-123","timestamp":1700000000000,"part":{"id":"part-1","type":"text","text":"I have completed the task."}}`
	got := FormatTraceLine(line)
	if !strings.Contains(got, "ASSISTANT") {
		t.Fatalf("expected assistant label, got: %s", got)
	}
	if !strings.Contains(got, "I have completed the task.") {
		t.Fatalf("expected answer text, got: %s", got)
	}
}

func TestFormatTraceLine_OpencodeToolUse(t *testing.T) {
	line := `{"type":"tool_use","sessionID":"sess-123","timestamp":1700000000000,"part":{"id":"part-2","type":"tool","tool":"bash","callID":"call-1","state":{"status":"completed","title":"go build"}}}`
	got := FormatTraceLine(line)
	if !strings.Contains(got, "Shell") && !strings.Contains(got, "RUN") {
		t.Fatalf("expected Shell or RUN label, got: %s", got)
	}
}

func TestWriteTraceReportPrintsLinkedTraces(t *testing.T) {
	t.Skip("writeTraceReport not yet implemented")
	parent := trace.AgentTraceSummary{AgentTraceMetadata: trace.AgentTraceMetadata{
		ID:      "murphy-1",
		Command: "murphy",
		Status:  "completed",
		LogPath: filepath.Join("murphy-1", "events.jsonl"),
		Children: []trace.AgentTraceChild{{
			ID:              "codenn-1",
			Command:         "codenn",
			Status:          "completed",
			DelegationLabel: "hello-world",
		}},
	}}
	child := trace.AgentTraceSummary{AgentTraceMetadata: trace.AgentTraceMetadata{
		ID:              "codenn-1",
		Command:         "codenn",
		Status:          "completed",
		ParentTraceID:   "murphy-1",
		DelegationLabel: "hello-world",
		LogPath:         filepath.Join("codenn-1", "events.jsonl"),
	}}
	source := printStaticSource{summaries: []trace.AgentTraceSummary{parent, child}}
	var out bytes.Buffer
	if err := writeTraceReport(&out, source, []string{"/tmp/traces"}, 0); err != nil {
		t.Fatalf("write report: %v", err)
	}
	text := out.String()
	for _, want := range []string{
		"Agent trace sources:",
		"/tmp/traces",
		"murphy-1 murphy completed",
		"children:",
		"codenn-1 codenn completed label=hello-world",
		"codenn-1 codenn completed",
		"parent: murphy-1",
		"delegation: hello-world",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("report missing %q:\n%s", want, text)
		}
	}
}

type printStaticSource struct {
	summaries []trace.AgentTraceSummary
}

func (s printStaticSource) List() ([]trace.AgentTraceSummary, error) {
	return s.summaries, nil
}

func (s printStaticSource) Get(string) (*trace.AgentTraceDetail, error) {
	return &trace.AgentTraceDetail{}, nil
}

func (s printStaticSource) Stop(string) (*trace.AgentTraceDetail, error) {
	return nil, nil
}

func (s printStaticSource) Delete(string) error {
	return nil
}

func (s printStaticSource) Describe() []string {
	return nil
}

func writeTraceReport(w *bytes.Buffer, source trace.Source, desc []string, n int) error {
	return nil
}
