# Grok Event Types Tests

These doc-style tests verify the `agent/event/grok_types` package: native grok streaming-json structs,
`ToGrok` (canonical AgentEvent → grok Event), and `FromGrok` (grok Event → canonical AgentEvent).

## Grok Streaming-JSON Format

The grok CLI emits lines of JSON with a `type` field:

| Type | Fields | Description |
|---|---|---|
| `text` | `data` (string) | Text response chunk |
| `thought` | `data` (string) | Reasoning / thinking text |
| `end` | `sessionId` (string), `stopReason` (string), `requestId` (string) | Session end marker |
| `tool_started` | `tool_name` (string) | Tool invocation started |
| `tool_completed` | `tool_name`, `outcome`, `duration_ms` | Tool finished (paired with `tool_started`) |

## Decision Tree

```
Level 1: Operation Mode
├── to_grok/              ToGrok: canonical AgentEvent → grok Event
├── from_grok/            FromGrok: grok Event → canonical AgentEvent
└── roundtrip/            ToGrok → FromGrok roundtrip preserves key fields

Level 2 (to_grok): Canonical ActionType
├── message               ActionMessage → text event {data}
├── think                 ActionThink → thought event {data}
├── done                  ActionDone → end event {sessionId}
├── error                 ActionError → text event {data} with error message
├── tool-call             ActionToolCall → skipped (no output)
├── step-start            ActionStepStart → skipped (no output)
└── step-finish           ActionStepFinish → skipped (no output)

Level 2 (from_grok): Grok Event Type
├── text                  text event → ActionMessage
├── thought               thought event → ActionThink
├── end                   end event → ActionDone (sessionId captured in ToolInput)
├── unknown               unknown type → skipped (no output)
├── empty-data            text event with empty data → ActionMessage with empty text
└── tool-started-read     tool_started Read → ActionToolCall (RED)

Level 2 (roundtrip): Scenario
├── text-msg              ActionMessage → ToGrok → FromGrok preserves text
├── thinking              ActionThink → ToGrok → FromGrok preserves thinking text
├── end                   ActionDone → ToGrok → FromGrok preserves session ID
└── error                 ActionError → ToGrok → FromGrok preserves error message
```

## Test Leaves

| Leaf | Description |
|---|---|
| **to_grok** | |
| `to_grok/message` | ActionMessage → grok `text` event with Data field |
| `to_grok/think` | ActionThink → grok `thought` event with Data field |
| `to_grok/done` | ActionDone → grok `end` event with SessionID |
| `to_grok/error` | ActionError → grok `text` event with error message in Data |
| `to_grok/tool-call` | ActionToolCall → no grok events emitted |
| `to_grok/step-start` | ActionStepStart → no grok events emitted |
| `to_grok/step-finish` | ActionStepFinish → no grok events emitted |
| **from_grok** | |
| `from_grok/text` | grok `text` event → ActionMessage |
| `from_grok/thought` | grok `thought` event → ActionThink |
| `from_grok/end` | grok `end` event → ActionDone (sessionId in ToolInput) |
| `from_grok/unknown` | grok unknown type → no agent events emitted |
| `from_grok/empty-data` | grok `text` event with empty data → ActionMessage with empty Text |
| `from_grok/tool-started-read` | grok `tool_started` Read → ActionToolCall with tool read (RED) |
| **roundtrip** | |
| `roundtrip/text-msg` | ActionMessage → ToGrok → FromGrok preserves text content |
| `roundtrip/thinking` | ActionThink → ToGrok → FromGrok preserves thinking text |
| `roundtrip/end` | ActionDone → ToGrok → FromGrok preserves session ID |
| `roundtrip/error` | ActionError → ToGrok → FromGrok preserves error message |

## How to Run

```sh
doctest test ./agent/event/grok_types/tests
```
