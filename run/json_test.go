package run

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/trace"
)

func TestPrintTraceJSONReportFiltersByTraceID(t *testing.T) {
	parent := trace.AgentTraceSummary{AgentTraceMetadata: trace.AgentTraceMetadata{
		ID:      "s5:murphy-1",
		Command: "murphy",
		Status:  "completed",
		LogPath: filepath.Join("murphy-1", "events.jsonl"),
		Children: []trace.AgentTraceChild{{
			ID:              "codenn-2",
			Command:         "codenn",
			Status:          "completed",
			DelegationLabel: "api",
		}},
	}}
	child := trace.AgentTraceSummary{AgentTraceMetadata: trace.AgentTraceMetadata{
		ID:              "codenn-2",
		Command:         "codenn",
		Status:          "completed",
		ParentTraceID:   "murphy-1",
		DelegationLabel: "api",
		LogPath:         filepath.Join("codenn-2", "events.jsonl"),
	}}
	source := printStaticSource{summaries: []trace.AgentTraceSummary{parent, child}}

	var out bytes.Buffer
	if err := printTraceJSONReport(&out, source, []string{"/tmp/traces"}, traceJSONFilters{TraceID: "murphy-1"}); err != nil {
		t.Fatalf("json report: %v", err)
	}
	var report traceJSONReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("decode json: %v\n%s", err, out.String())
	}
	if len(report.Sources) != 1 || report.Sources[0] != "/tmp/traces" {
		t.Fatalf("sources = %#v", report.Sources)
	}
	if len(report.Traces) != 1 {
		t.Fatalf("traces = %#v, want one parent trace", report.Traces)
	}
	if report.Traces[0].ID != "s5:murphy-1" {
		t.Fatalf("trace ID = %q, want s5:murphy-1", report.Traces[0].ID)
	}
	if len(report.Traces[0].Children) != 1 || report.Traces[0].Children[0].DelegationLabel != "api" {
		t.Fatalf("children = %#v, want grouped api child", report.Traces[0].Children)
	}
}

func TestFilterTraceJSONSummariesByDelegationLabelMatchesChildren(t *testing.T) {
	parent := trace.AgentTraceSummary{AgentTraceMetadata: trace.AgentTraceMetadata{
		ID:      "murphy-1",
		Command: "murphy",
		Children: []trace.AgentTraceChild{{
			ID:              "codenn-1",
			Command:         "codenn",
			DelegationLabel: "api",
		}},
	}}
	other := trace.AgentTraceSummary{AgentTraceMetadata: trace.AgentTraceMetadata{
		ID:      "murphy-2",
		Command: "murphy",
	}}
	got := filterTraceJSONSummaries([]trace.AgentTraceSummary{parent, other}, traceJSONFilters{DelegationLabel: "api"})
	if len(got) != 1 || got[0].ID != "murphy-1" {
		t.Fatalf("filtered summaries = %#v, want murphy-1", got)
	}
}
