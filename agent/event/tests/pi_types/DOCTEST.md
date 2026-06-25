# Pi Types Tests

These doc-style tests verify the `agent/event/pi_types` package: wire-format structs,
`ToPi` (canonical AgentEvent → pi Event), and `FromPi` (pi Event → canonical AgentEvent).

## Decision Tree

```
Level 1: Operation Mode
├── wire-format/        Parse JSON → pi_types.Event (struct fields + roundtrip)
├── to_pi/              types.AgentEvent → pi_types.Event (per-ActionType + Phase)
├── from_pi/            pi_types.Event → types.AgentEvent (per-pi-event-type + Phase)
└── roundtrip/          ToPi → FromPi preserves key fields

Level 2 (wire-format): Pi Event Type
├── session             {type:"session", id:"sess_1"}
├── agent_start         {type:"agent_start"}
├── agent_end           {type:"agent_end", messages:[...]}
├── turn_start          {type:"turn_start"}
├── turn_end            {type:"turn_end", message, toolResults}
├── message_start       {type:"message_start", message:{role:"user",...}}
├── message_update_text     {type:"message_update", assistantMessageEvent:{type:"text_delta",...}}
├── message_update_thinking {type:"message_update", assistantMessageEvent:{type:"thinking_delta",...}}
├── message_end         {type:"message_end", message:{role:"assistant",...}}
├── tool_exec_start     {type:"tool_execution_start", toolCallId, toolName, args}
├── tool_exec_end_ok    {type:"tool_execution_end", result:{...}, isError:false}
└── tool_exec_end_err   {type:"tool_execution_end", isError:true}

Level 2 (to_pi): ActionType + Phase
├── think_instant       ActionThink + Phase=""    → msg_start + msg_update(thinking) + msg_end
├── think_update        ActionThink + PhaseUpdate → msg_update(thinking_delta)
├── msg_instant         ActionMessage + Phase=""  → msg_start + msg_update(text) + msg_end
├── msg_start           ActionMessage + PhaseStart → msg_start
├── msg_update          ActionMessage + PhaseUpdate → msg_update(text_delta)
├── msg_end             ActionMessage + PhaseEnd  → msg_end
├── tc_instant          ActionToolCall + Phase=""  → tool_exec_start + tool_exec_end
├── tc_start            ActionToolCall + PhaseStart → tool_exec_start
├── tc_update           ActionToolCall + PhaseUpdate → tool_exec_update
├── tc_end              ActionToolCall + PhaseEnd → tool_exec_end
├── error               ActionError               → msg_start + msg_end
├── done                ActionDone                → agent_end
├── step_start          ActionStepStart           → turn_start
└── step_finish         ActionStepFinish          → turn_end

Level 2 (from_pi): Pi Event → ActionType + Phase
├── msg_update_text     msg_update text_delta     → ActionMessage + PhaseUpdate
├── msg_update_think    msg_update thinking_delta → ActionThink + PhaseUpdate
├── msg_start           msg_start                 → ActionMessage + PhaseStart
├── msg_end             msg_end                   → ActionMessage + PhaseEnd
├── tool_exec_start     tool_exec_start           → ActionToolCall + PhaseStart
├── tool_exec_update    tool_exec_update          → ActionToolCall + PhaseUpdate
├── tool_exec_end       tool_exec_end ok         → ActionToolCall + PhaseEnd
├── tool_exec_end_err   tool_exec_end isError    → ActionToolCall + PhaseEnd + exit code
├── agent_end           agent_end                 → ActionDone + PhaseEnd
├── agent_start         agent_start               → ActionStepStart + PhaseStart
├── turn_start          turn_start                → ActionStepStart + PhaseStart
├── turn_end            turn_end                  → ActionStepFinish + PhaseEnd
├── session             session metadata          → no action (empty result)
├── non_assistant_update user/toolResult role     → no action (skip)

Level 2 (roundtrip): Scenario
├── text_msg            ActionMessage text → ToPi → FromPi preserves text
├── thinking            ActionThink       → ToPi → FromPi preserves thinking
├── tool_call           ActionToolCall    → ToPi → FromPi preserves tool fields
├── step_start_finish   step_start+finish → ToPi → FromPi preserves steps
└── error               ActionError       → ToPi → FromPi preserves error
```

## Test Leaves

| Leaf | Description |
|---|---|
| **wire-format** | |
| `wire-format/session` | Parse session event with id, verify fields |
| `wire-format/agent-start` | Parse agent_start (empty payload), verify type |
| `wire-format/agent-end` | Parse agent_end with messages array |
| `wire-format/turn-start` | Parse turn_start (empty payload) |
| `wire-format/turn-end` | Parse turn_end with assistant message + tool results |
| `wire-format/message-start` | Parse message_start with user message |
| `wire-format/message-update-text` | Parse message_update with text_delta assistant event |
| `wire-format/message-update-thinking` | Parse message_update with thinking_delta |
| `wire-format/message-end` | Parse message_end with full assistant message |
| `wire-format/tool-exec-start` | Parse tool_execution_start with toolCallId, toolName, args |
| `wire-format/tool-exec-end-ok` | Parse tool_execution_end success (isError:false) |
| `wire-format/tool-exec-end-err` | Parse tool_execution_end error (isError:true) |
| **to_pi** | |
| `to_pi/think-instant` | ActionThink Phase="" → msg_start + msg_update thinking + msg_end |
| `to_pi/think-update` | ActionThink PhaseUpdate → msg_update thinking_delta |
| `to_pi/msg-instant` | ActionMessage Phase="" → msg_start + msg_update text + msg_end |
| `to_pi/msg-start` | ActionMessage PhaseStart → msg_start |
| `to_pi/msg-update` | ActionMessage PhaseUpdate → msg_update text_delta |
| `to_pi/msg-end` | ActionMessage PhaseEnd → msg_end |
| `to_pi/tc-instant` | ActionToolCall Phase="" → tool_exec_start + tool_exec_end |
| `to_pi/tc-start` | ActionToolCall PhaseStart → tool_exec_start |
| `to_pi/tc-update` | ActionToolCall PhaseUpdate → tool_exec_update |
| `to_pi/tc-end` | ActionToolCall PhaseEnd → tool_exec_end |
| `to_pi/error` | ActionError → msg_start + msg_end |
| `to_pi/done` | ActionDone → agent_end |
| `to_pi/step-start` | ActionStepStart → turn_start |
| `to_pi/step-finish` | ActionStepFinish → turn_end |
| **from_pi** | |
| `from_pi/msg-update-text` | msg_update text_delta → ActionMessage PhaseUpdate (Text = Delta, not full Content) |
| `from_pi/msg-update-accumulated-text` | msg_update with large accumulated Content → ActionMessage PhaseUpdate (Text = Delta " feature." only) |
| `from_pi/msg-update-think` | msg_update thinking_delta → ActionThink PhaseUpdate (Text = Delta, not Content[0].Thinking) |
| `from_pi/msg-start` | msg_start → ActionMessage PhaseStart |
| `from_pi/msg-end` | msg_end → ActionMessage PhaseEnd |
| `from_pi/tool-exec-start` | tool_exec_start → ActionToolCall PhaseStart |
| `from_pi/tool-exec-update` | tool_exec_update → ActionToolCall PhaseUpdate |
| `from_pi/tool-exec-end` | tool_exec_end ok → ActionToolCall PhaseEnd |
| `from_pi/tool-exec-end-err` | tool_exec_end isError → ActionToolCall PhaseEnd + exit code |
| `from_pi/agent-end` | agent_end → ActionDone PhaseEnd |
| `from_pi/agent-start` | agent_start → ActionStepStart PhaseStart |
| `from_pi/turn-start` | turn_start → ActionStepStart PhaseStart |
| `from_pi/turn-end` | turn_end → ActionStepFinish PhaseEnd |
| `from_pi/session` | session → empty result (no action) |
| `from_pi/non-assistant-update` | user/toolResult role message → skip (no action) |
| **roundtrip** | |
| `roundtrip/text-msg` | ActionMessage → ToPi → FromPi preserves text content |
| `roundtrip/thinking` | ActionThink → ToPi → FromPi preserves thinking text |
| `roundtrip/tool-call` | ActionToolCall → ToPi → FromPi preserves tool fields |
| `roundtrip/step-start-finish` | step_start + step_finish → ToPi → FromPi preserves steps |
| `roundtrip/error` | ActionError → ToPi → FromPi preserves error message |

## How to Run

```sh
doctest test ./agent/event/tests/pi_types
```

```go
import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	pi_types "github.com/xhd2015/agent-pro/agent/event/pi_types"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)


type Request struct {
	Target    string           // "wire", "to_pi", "from_pi"
	JSONInput string           // raw JSON for wire format parsing
	Events    []types.AgentEvent // input for ToPi
	PiEvents  []pi_types.Event   // input for FromPi
	Output    string           // passthrough
}

type Response struct {
	Output string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	var output string
	switch req.Target {
	case "wire":
		var evt pi_types.Event
		if err := json.Unmarshal([]byte(req.JSONInput), &evt); err != nil {
			return nil, fmt.Errorf("unmarshal error: %w", err)
		}
		data, _ := json.Marshal(evt)
		output = string(data)
	case "to_pi":
		piEvts := pi_types.ToPi(req.Events)
		data, _ := json.Marshal(piEvts)
		output = string(data)
	case "from_pi":
		agentEvts := pi_types.FromPi(req.PiEvents)
		data, _ := json.Marshal(agentEvts)
		output = string(data)
	case "roundtrip":
		piEvts := pi_types.ToPi(req.Events)
		agentEvts := pi_types.FromPi(piEvts)
		data, _ := json.Marshal(agentEvts)
		output = string(data)
	default:
		output = req.Output
	}
	return &Response{Output: output}, nil
}
```
