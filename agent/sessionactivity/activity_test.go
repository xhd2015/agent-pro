package sessionactivity

import (
	"encoding/json"
	"strings"
	"testing"

	dshsessions "github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

func TestClassifyChat(t *testing.T) {
	cases := []struct {
		kind, tool string
		live       bool
		want       Kind
		detail     string
	}{
		{kind: "thinking", live: true, want: Thinking},
		{kind: "thinking", live: false, want: ErrorKind, detail: "thinking but session is not live"},
		{kind: "tool", tool: "bash", live: true, want: ToolCalling, detail: "bash"},
		{kind: "tool", live: false, want: ErrorKind, detail: "tool call but session is not live"},
		{kind: "response", live: true, want: Idle},
		{kind: "response", live: false, want: Idle},
		{kind: "user", live: true, want: Thinking},
		{kind: "user", live: false, want: Idle},
		{kind: "", live: true, want: Idle},
	}
	for _, tc := range cases {
		got, detail := ClassifyChat(tc.kind, tc.tool, tc.live)
		if got != tc.want || detail != tc.detail {
			t.Fatalf("%s live=%v: got %s %q want %s %q", tc.kind, tc.live, got, detail, tc.want, tc.detail)
		}
	}
}

func TestProbeReadUnknownRunner(t *testing.T) {
	_, err := (Probe{}).Read("clippy", "s")
	if err == nil || !strings.Contains(err.Error(), "unknown runner") {
		t.Fatalf("err=%v", err)
	}
}

func TestProbeReadDSH(t *testing.T) {
	inspection := &dshsessions.Inspection{
		Summary: dshsessions.Summary{ID: "s1", Cwd: "/ws", Outcome: "incomplete", UpdatedAt: 9},
		Events: []dshsessions.Event{
			{Type: "turn/start", Data: json.RawMessage(`{"turn":1}`)},
			{Type: "tool/call", Data: json.RawMessage(`{"name":"bash"}`)},
		},
	}
	p := Probe{
		DSH: func(id string) (Snapshot, error) {
			return snapshotFromDSH(inspection, true, true), nil
		},
	}
	got, err := p.Read("dsh", "s1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != ToolCalling || got.Detail != "bash" || !got.Live || got.Workspace != "/ws" || got.Runner != "dsh" || got.SessionID != "s1" {
		t.Fatalf("%+v", got)
	}
}

func TestProbeReadRequiresIDs(t *testing.T) {
	if _, err := (Probe{}).Read("", "s"); err == nil {
		t.Fatal("empty runner accepted")
	}
	if _, err := (Probe{}).Read("dsh", "  "); err == nil {
		t.Fatal("empty session accepted")
	}
}
