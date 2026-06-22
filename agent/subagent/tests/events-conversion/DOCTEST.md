# Events Conversion Tests

Verify the events.jsonl pipeline conversion from raw runner-native JSONL to canonical
`AgentEvent` JSONL, and the downstream trace/status display using the new typed events.

Focus areas: converter router (`agent/event/convert`), `formatEventLine` rendering,
`traceSession` (end-to-end trace), and `showStatus` (end-to-end status).

## Decision Tree

```
events-conversion/
├── DOCTEST.md                            # This file
├── SETUP.md                              # Root: Request/Response, Run dispatches by Operation
│
├── converter-router/                     # === ConvertRawLine(raw, agentRunner) → AgentEvent ===
│   ├── SETUP.md                          # Operation="convert_raw", sets up runner-specific paths
│   ├── opencode/raw-to-agent-event/      # Feed opencode native JSON → verify AgentEvent fields
│   ├── pi/raw-to-agent-event/            # Feed pi native JSON array → verify AgentEvent fields (backward compat)
│   ├── pi/single-event-not-array/        # Feed pi single JSON object → verify accepted (Bug A fix)
│   ├── codex/raw-to-agent-event/         # Feed codex native JSON → verify AgentEvent fields
│   ├── crush/raw-to-agent-event/         # Feed crush native JSON → verify AgentEvent fields
│   └── unknown-runner/                   # Unknown runner name → error
│
├── format-event-line/                    # === formatEventLine(AgentEvent JSON) → string ===
│   ├── SETUP.md                          # Operation="format_event"
│   ├── tool-call/                        # Type=tool_call → formatted with tool name and output
│   ├── message/                          # Type=message → formatted with text content
│   ├── think/                            # Type=think → formatted with reasoning text
│   ├── error/                            # Type=error → formatted with error text
│   ├── step-start-finish/                # Type=step_start/step_finish → formatted
│   └── with-timestamp/                   # Event with Timestamp → timestamp in output
│
├── trace-session/                        # === traceSession() e2e ===
│   ├── SETUP.md                          # Operation="trace"
│   ├── no-events/                        # No events.jsonl → "(no events yet)"
│   ├── with-agent-events/               # events.jsonl AgentEvent lines → formatted trace
│   └── with-streaming-grok-thought-events/  # Grok per-word think deltas → coalesced trace (RED)
│
└── show-status/                          # === showStatus() e2e ===
    ├── SETUP.md                          # Operation="status"
    ├── event-count/                      # N events → shows correct count
    ├── last-event-time/                  # Event with Timestamp → relative time
    └── event-details/                    # Last 3 events → formatted summaries
```

## Test Index

### converter-router — 6 leaves
| Leaf | Description |
|------|-------------|
| `converter-router/opencode/raw-to-agent-event` | OpenCode native events → correct AgentEvent Type/Text/Tool |
| `converter-router/pi/raw-to-agent-event` | Pi native events (array) → correct AgentEvent fields (backward compat) |
| `converter-router/pi/single-event-not-array` | Pi single object (not array) → accepted + converted (Bug A fix) |
| `converter-router/codex/raw-to-agent-event` | Codex native events → correct AgentEvent fields (tests FromCodex) |
| `converter-router/crush/raw-to-agent-event` | Crush native events → correct AgentEvent fields |
| `converter-router/unknown-runner` | Unknown runner name → non-nil error |

### format-event-line — 6 leaves
| Leaf | Description |
|------|-------------|
| `format-event-line/tool-call` | `AgentEvent{Type:tool_call}` with tool/input/output → formatted string |
| `format-event-line/message` | `AgentEvent{Type:message}` with text → formatted string |
| `format-event-line/think` | `AgentEvent{Type:think}` with text → formatted string |
| `format-event-line/error` | `AgentEvent{Type:error}` with text → formatted string |
| `format-event-line/step-start-finish` | `step_start`/`step_finish` events → formatted string |
| `format-event-line/with-timestamp` | Event with Timestamp field → timestamp in output |

### trace-session — 4 leaves
| Leaf | Description |
|------|-------------|
| `trace-session/no-events` | No events.jsonl → "(no events yet)", Done footer |
| `trace-session/with-agent-events` | events.jsonl with AgentEvent lines → formatted lines in order |
| `trace-session/with-streaming-pi-events` | Pi per-token message deltas → coalesced ASSISTANT block |
| `trace-session/with-streaming-grok-thought-events` | Grok per-word think deltas → coalesced think block (RED) |

### show-status — 3 leaves
| Leaf | Description |
|------|-------------|
| `show-status/event-count` | events.jsonl with N AgentEvent lines → shows correct count |
| `show-status/last-event-time` | Event with Timestamp → correct relative time ("Xs ago") |
| `show-status/event-details` | Last 3 events → each shown with formatted summary |

Total: **19 leaves** across **4 feature areas**.

## How to Run

```sh
doctest test -v ./agent/subagent/tests/events-conversion/
```
