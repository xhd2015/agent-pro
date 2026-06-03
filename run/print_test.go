package run

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/trace"
)

func TestWriteTraceReportPrintsLinkedTraces(t *testing.T) {
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
