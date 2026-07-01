# Scenario

**Feature**: `opencode_types.FromOpencode` converts opencode events → AgentEvent, and `ToOpencode` converts back

## Preconditions
- `opencode_types.FromOpencode` converts opencode events → AgentEvent, and `ToOpencode` converts back.
- The round trip preserves each event type identically.

## Steps
1. For each of the 7 opencode event types, create a representative event.
2. Round-trip through FromOpencode → ToOpencode individually.
3. Compare orig JSON vs round-tripped JSON per type.
4. Report [PASS]/[FAIL] per type with mismatch details, then summary.

```go
import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	opencode_types "github.com/xhd2015/agent-pro/agent/event/opencode_types"
)

func Setup(t *testing.T, req *Request) error {
	tests := []struct {
		name  string
		event opencode_types.Event
	}{
		{
			name: "reasoning",
			event: opencode_types.Event{
				Type:      opencode_types.EvtReasoning,
				SessionID: "sess_rt",
				Part: opencode_types.ReasoningPart{
					ID:   "evt_r1",
					Type: opencode_types.PartReasoning,
					Text: "thinking step by step",
				},
			},
		},
		{
			name: "text",
			event: opencode_types.Event{
				Type:      opencode_types.EvtText,
				SessionID: "sess_rt",
				Part: opencode_types.TextPart{
					ID:   "evt_t1",
					Type: opencode_types.PartText,
					Text: "hello world",
				},
			},
		},
		{
			name: "error",
			event: opencode_types.Event{
				Type:      opencode_types.EvtError,
				SessionID: "sess_rt",
				Error: &opencode_types.ErrorDetail{
					Name: "Error",
					Data: &opencode_types.ErrorData{Message: "something broke"},
				},
			},
		},
		{
			name: "done",
			event: opencode_types.Event{
				Type:      opencode_types.EvtDone,
				SessionID: "sess_rt",
				Done:      true,
			},
		},
		{
			name: "step_start",
			event: opencode_types.Event{
				Type:      opencode_types.EvtStepStart,
				SessionID: "sess_rt",
				Timestamp: 1718200000123,
				Part: opencode_types.StepStartPart{
					ID:        "p1",
					SessionID: "sess_rt",
					MessageID: "msg_1",
					Type:      opencode_types.PartStepStart,
					Snapshot:  "snap_abc",
				},
			},
		},
		{
			name: "step_finish",
			event: opencode_types.Event{
				Type:      opencode_types.EvtStepFinish,
				SessionID: "sess_rt",
				Timestamp: 1718200000456,
				Part: opencode_types.StepFinishPart{
					ID:        "p2",
					SessionID: "sess_rt",
					MessageID: "msg_2",
					Type:      opencode_types.PartStepFinish,
					Reason:    "stop",
					Cost:      0.015,
				},
			},
		},
		{
			name: "tool_use",
			event: opencode_types.Event{
				Type:      opencode_types.EvtToolUse,
				SessionID: "sess_rt",
				Part: opencode_types.ToolUsePart{
					ID:     "evt_bash",
					Type:   opencode_types.PartTool,
					CallID: "evt_bash",
					Tool:   "bash",
					State: opencode_types.ToolUseState{
						Input:    map[string]any{"command": "echo hi"},
						Output:   "hi",
						ExitCode: 0,
						Status:   "completed",
					},
				},
			},
		},
	}

	var sb strings.Builder
	mismatches := 0
	for _, tc := range tests {
		agentEvents := opencode_types.FromOpencode([]opencode_types.Event{tc.event}, "sess_rt")
		roundtripped := opencode_types.ToOpencode(agentEvents, "sess_rt")

		origJSON, _ := json.Marshal(tc.event)
		rtJSON, _ := json.Marshal(roundtripped[0])
		match := string(origJSON) == string(rtJSON)

		if match {
			fmt.Fprintf(&sb, "[PASS] %s\n", tc.name)
		} else {
			mismatches++
			fmt.Fprintf(&sb, "[FAIL] %s\n   orig: %s\n     rt: %s\n", tc.name, string(origJSON), string(rtJSON))
		}
	}

	if mismatches == 0 {
		fmt.Fprint(&sb, "ALL MATCH")
	} else {
		fmt.Fprintf(&sb, "MISMATCH: %d of %d types", mismatches, len(tests))
	}

	req.Output = sb.String()
	return nil
}
```
