package queue

import (
	"testing"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func TestDequeueToBreakpoint_sleepThenMessage(t *testing.T) {
	q := []types.AgentEvent{
		{Type: types.ActionSleep, DelayMs: 50},
		{Type: types.ActionMessage, Text: "done"},
	}
	batch, ok := DequeueToBreakpoint(&q)
	if !ok {
		t.Fatal("expected ok")
	}
	if len(batch.PrefixSleep) != 1 || batch.PrefixSleep[0].DelayMs != 50 {
		t.Fatalf("PrefixSleep=%v", batch.PrefixSleep)
	}
	if batch.Breakpoint.Type != types.ActionMessage || batch.Breakpoint.Text != "done" {
		t.Fatalf("Breakpoint=%+v", batch.Breakpoint)
	}
	if len(q) != 0 {
		t.Fatalf("queue leftover %d", len(q))
	}
	evts := ConsumedEvents(batch)
	if DelayFor(evts) != 50*time.Millisecond {
		t.Fatalf("DelayFor=%s want 50ms", DelayFor(evts))
	}
}

func TestDequeueToBreakpoint_thinkToolUnchanged(t *testing.T) {
	q := []types.AgentEvent{
		{Type: types.ActionThink, Text: "t"},
		{Type: types.ActionToolCall, Tool: "bash"},
		{Type: types.ActionMessage, Text: "after"},
	}
	batch, ok := DequeueToBreakpoint(&q)
	if !ok {
		t.Fatal("expected ok")
	}
	if len(batch.PrefixThink) != 1 || batch.Breakpoint.Type != types.ActionToolCall {
		t.Fatalf("batch=%+v", batch)
	}
	if len(q) != 1 || q[0].Type != types.ActionMessage {
		t.Fatalf("leftover=%v", q)
	}
}

func TestDelayFor_messageDelayMs(t *testing.T) {
	d := DelayFor([]types.AgentEvent{{Type: types.ActionMessage, Text: "x", DelayMs: 12}})
	if d != 12*time.Millisecond {
		t.Fatalf("DelayFor=%s", d)
	}
}
