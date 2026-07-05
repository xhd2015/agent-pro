# Grok Session Tests

These doc-style tests verify the `agent/event/grok_session` package: two-way conversion
between grok on-disk session `updates.jsonl` (ACP session updates) and canonical
`types.AgentEvent`.

## Version

0.0.2

# DSN (Domain Specific Notion)

Grok persists conversations under `$GROK_HOME/sessions/<encoded-cwd>/<session-uuid>/updates.jsonl`.
Each line is an ACP **session update** — either a flat object
`{"sessionUpdate":"...", ...}` or a wire envelope
`{"method":"_x.ai/session/update","params":{"sessionId":"...","update":{...}}}`.

Participants and behaviors:

- **Wire line** — one JSONL record. `ParseLine` accepts flat or envelope shapes and
  returns a typed `SessionUpdate`.
- **Converter** — stateful forward adapter. `ProcessLine` ingests one wire line,
  coalesces text chunks (user / thought / assistant), tracks tool-call metadata,
  stamps `extensions.grok_session.turn_index`, and emits canonical `AgentEvent`s.
  `Flush` drains buffered chunks at tail or before turn boundaries.
- **Turn model** — no `turn_started` on wire. A turn begins implicitly at the first
  `user_message_chunk` after the previous `turn_completed`. `turn_completed` flushes
  pending chunks, emits `ActionDone`, and increments the turn counter.
- **FromUpdatesJSONL** — convenience wrapper: walk lines through a converter, flush
  at end, return all canonical events.
- **ToSession** — reverse adapter: canonical `AgentEvent` stream → `[]SessionUpdate`.
  `ActionDone` becomes `turn_completed`; tool calls may split into `tool_call` +
  `tool_call_update` pairs; missing `tool_call_id` is inferred.
- **ToWireLines** — marshal `SessionUpdate`s to JSONL strings (flat or envelope per opts).
- **Semantic roundtrip** — `wire₁ → events₁ → wire₂ → events₂`; assert
  `SemanticEqual(events₁, events₂)` (timestamps and chunk fragmentation may differ).

## Decision Tree

```
Level 1: Operation Mode
├── from_session/     updates.jsonl wire → canonical AgentEvent
├── to_session/       canonical AgentEvent → SessionUpdate / wire lines
└── roundtrip/        wire → events → wire → events (semantic equality)

Level 2 (from_session): Wire shape / update kind / turn behavior
├── flat-user-chunk              user_message_chunk → ActionMessage role=user
├── flat-thought-chunk           agent_thought_chunk → ActionThink
├── flat-assistant-chunk         agent_message_chunk → ActionMessage role=assistant
├── tool-call-pending            tool_call → ActionToolCall status=pending
├── tool-call-update-completed   tool_call_update → Output + status=completed
├── tool-call-update-failed      tool_call_update → status=failed
├── nested-wire-envelope         envelope parse ≡ flat
├── turn-completed-emits-done    turn_completed → ActionDone
├── turn-index-stamped-on-events all events in one turn share turn_index
├── multi-turn-increments-index  turn_index 0 then 1 across turns
└── full-turn-sequence           user + think + tool + assistant + turn_completed

Level 2 (to_session): Canonical event kind
├── user-message                 ActionMessage user → user_message_chunk
├── assistant-message            ActionMessage assistant → agent_message_chunk
├── think                        ActionThink → agent_thought_chunk
├── tool-call-pending            ActionToolCall pending → tool_call wire
├── tool-call-completed          pending + output → tool_call + tool_call_update
├── tool-call-failed             failed status on update wire
├── done-emits-turn-completed    ActionDone → turn_completed
├── infers-tool-call-id          generates tool_call_id when absent
└── multi-turn-done-sequence     two ActionDone → two turn_completed

Level 2 (roundtrip): Scenario
├── single-user-message          user text + turn_index preserved
├── single-assistant-message     assistant text preserved
├── thought-message              think text preserved
├── tool-pair-completed          tool_call_id + status + Output preserved
├── tool-pair-failed             failed status preserved
├── full-turn                    complete single turn roundtrips
├── multi-turn-session           two turns, turn_index 0/1, two ActionDone
├── turn-index-preserved         events₁ and events₂ share turn_index
├── nested-envelope-roundtrip    envelope wire roundtrips
└── chunk-coalescing-roundtrip   multi-chunk coalesce; reverse re-chunks; semantics equal
```

## Test Leaves

| Leaf | Description |
|---|---|
| **from_session** | |
| `from_session/flat-user-chunk` | Flat user_message_chunk → ActionMessage role=user, turn_index=0 |
| `from_session/flat-thought-chunk` | agent_thought_chunk → ActionThink, turn_index=0 |
| `from_session/flat-assistant-chunk` | agent_message_chunk → ActionMessage role=assistant |
| `from_session/tool-call-pending` | tool_call → ActionToolCall with tool_call_id, status=pending |
| `from_session/tool-call-update-completed` | tool_call_update → Output set, status=completed |
| `from_session/tool-call-update-failed` | tool_call_update → status=failed |
| `from_session/nested-wire-envelope` | Envelope wire parses same as flat user chunk |
| `from_session/turn-completed-emits-done` | turn_completed → ActionDone with turn_index |
| `from_session/turn-index-stamped-on-events` | All events in a turn share turn_index |
| `from_session/multi-turn-increments-index` | Two turns → turn_index 0 and 1 |
| `from_session/full-turn-sequence` | user + think + tool + assistant + turn_completed |
| **to_session** | |
| `to_session/user-message` | ActionMessage user → user_message_chunk wire |
| `to_session/assistant-message` | ActionMessage assistant → agent_message_chunk wire |
| `to_session/think` | ActionThink → agent_thought_chunk wire |
| `to_session/tool-call-pending` | ActionToolCall pending → tool_call wire |
| `to_session/tool-call-completed` | ActionToolCall with output → tool_call + tool_call_update |
| `to_session/tool-call-failed` | Failed tool → failed status on update |
| `to_session/done-emits-turn-completed` | ActionDone → turn_completed wire |
| `to_session/infers-tool-call-id` | Missing tool_call_id inferred on wire |
| `to_session/multi-turn-done-sequence` | Two ActionDone → two turn_completed lines |
| **roundtrip** | |
| `roundtrip/single-user-message` | User message roundtrips with turn_index |
| `roundtrip/single-assistant-message` | Assistant message roundtrips |
| `roundtrip/thought-message` | Think roundtrips |
| `roundtrip/tool-pair-completed` | Completed tool pair roundtrips |
| `roundtrip/tool-pair-failed` | Failed tool pair roundtrips |
| `roundtrip/full-turn` | Complete single turn roundtrips |
| `roundtrip/multi-turn-session` | Two-turn session roundtrips |
| `roundtrip/turn-index-preserved` | turn_index equal in events₁ vs events₂ |
| `roundtrip/nested-envelope-roundtrip` | Envelope wire roundtrips |
| `roundtrip/chunk-coalescing-roundtrip` | Multi-chunk coalesce roundtrips semantically |

## How to Run

```sh
doctest vet ./agent/event/grok_session/tests
doctest test ./agent/event/grok_session/tests
doctest test ./agent/event/grok_session/tests/roundtrip/...
```

```go
import (
	"encoding/json"
	"strings"
	"testing"

	grok_session "github.com/xhd2015/agent-pro/agent/event/grok_session"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

type Request struct {
	Target    string // "from_session", "to_session", "roundtrip"
	WireLines []string
	Events    []types.AgentEvent
	SessionID string
	ToOpts    grok_session.ToOptions
}

type Response struct {
	Output     string
	Events     []types.AgentEvent
	WireLines  []string
	Events1    []types.AgentEvent
	Events2    []types.AgentEvent
	WireLines1 []string
	WireLines2 []string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	switch req.Target {
	case "from_session":
		events := grok_session.FromUpdatesJSONL(req.WireLines)
		resp.Events = events
		data, _ := json.Marshal(events)
		resp.Output = string(data)
	case "to_session":
		updates := grok_session.ToSession(req.Events, req.ToOpts)
		wire := grok_session.ToWireLines(updates, req.ToOpts)
		resp.WireLines = wire
		resp.Output = strings.Join(wire, "\n")
	case "roundtrip":
		events1 := grok_session.FromUpdatesJSONL(req.WireLines)
		updates2 := grok_session.ToSession(events1, req.ToOpts)
		wire2 := grok_session.ToWireLines(updates2, req.ToOpts)
		events2 := grok_session.FromUpdatesJSONL(wire2)
		resp.WireLines1 = req.WireLines
		resp.WireLines2 = wire2
		resp.Events1 = events1
		resp.Events2 = events2
		payload := map[string]any{
			"events1": events1,
			"events2": events2,
			"wire2":   wire2,
		}
		data, _ := json.Marshal(payload)
		resp.Output = string(data)
	default:
		t.Fatalf("unknown Target %q", req.Target)
	}
	return resp, nil
}
```