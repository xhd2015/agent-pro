package mockpreset

import (
	"bytes"
	"strings"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func TestListContainsAllMVPPresets(t *testing.T) {
	names := make(map[string]struct{})
	for _, entry := range List() {
		names[entry.Name] = struct{}{}
		if entry.Description == "" {
			t.Fatalf("preset %q missing description", entry.Name)
		}
	}
	for _, want := range []string{
		"simple",
		"think-message",
		"multi-think",
		"tool-bash",
		"tool-read",
		"think-tool-message",
	} {
		if _, ok := names[want]; !ok {
			t.Fatalf("List() missing preset %q", want)
		}
	}
}

func TestResolveThinkMessage(t *testing.T) {
	events, err := Resolve("think-message")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("len = %d, want 2", len(events))
	}
	if events[0].Type != types.ActionThink || events[0].Text == "" {
		t.Fatalf("first event = %#v, want non-empty think", events[0])
	}
	if events[1].Type != types.ActionMessage || events[1].Text == "" {
		t.Fatalf("second event = %#v, want non-empty message", events[1])
	}
	if events[0].Text == events[1].Text {
		t.Fatalf("think and message text must differ: %q", events[0].Text)
	}
}

func TestResolveToolBash(t *testing.T) {
	events, err := Resolve("tool-bash")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("len = %d, want 1", len(events))
	}
	if events[0].Type != types.ActionToolCall || events[0].Tool != "bash" {
		t.Fatalf("event = %#v, want bash tool_call", events[0])
	}
}

func TestResolveUnknown(t *testing.T) {
	_, err := Resolve("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown preset")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Fatalf("error = %v, want mention of nonexistent", err)
	}
}

func TestPrintList(t *testing.T) {
	var buf bytes.Buffer
	PrintList(&buf)
	out := buf.String()
	for _, name := range []string{
		"simple",
		"think-message",
		"multi-think",
		"tool-bash",
		"tool-read",
		"think-tool-message",
	} {
		if !strings.Contains(out, name) {
			t.Fatalf("PrintList output missing %q:\n%s", name, out)
		}
	}
}