# Event Package Tests

These doc-style tests verify the `agent/event/types`, `agent/event/codex_types`, `agent/event/opencode_types`, and `agent/event/crush_types` packages.

## Decision Tree: Crush Event Types

The crush runner uses SSE payloads with `{\"type\": \"<payload_type>\", \"payload\": {...}}`.
Conversion functions `FromCrush` (crush→canonical) and `ToCrush` (canonical→crush) are tested.

```
crush_types/
├── to_crush/                        (Target="crush", canonical→crush)
│   ├── think/                       ActionThink → message with reasoning part
│   ├── message/                     ActionMessage → message with text part
│   ├── tool_call/                   ActionToolCall → message with tool_call part
│   ├── error/                       ActionError → agent_event with type=error
│   └── done/                        ActionDone → run_complete
│
└── from_crush/                      (Target="from_crush", crush→canonical)
    ├── reasoning/                    reasoning part → ActionThink
    ├── text/                         text part → ActionMessage
    ├── tool_call/                    tool_call part → ActionToolCall
    ├── tool_result/                  tool_result part → skipped (no action)
    ├── finish_error/                 finish reason=error → ActionError
    ├── finish_no_error/              finish reason=end_turn → skipped
    ├── mixed_parts/                  text + tool_call → multiple actions
    ├── non_assistant/                user/system role → skipped
    ├── agent_event_error/            agent_event type=error → ActionError
    ├── agent_event_response/         agent_event type=response → no ActionError
    ├── run_complete/                 run_complete success → ActionDone
    ├── run_complete_error/           run_complete with error → ActionDone
    └── run_complete_cancelled/       run_complete cancelled → ActionDone
```

## Test Leaves

| Leaf | Description |
|---|---|
| `crush_types-to-crush-think` | ToCrush: ActionThink → crush reasoning message |
| `crush_types-to-crush-message` | ToCrush: ActionMessage → crush text message |
| `crush_types-to-crush-tool-call` | ToCrush: ActionToolCall → crush tool_call message |
| `crush_types-to-crush-error` | ToCrush: ActionError → crush agent_event error |
| `crush_types-to-crush-done` | ToCrush: ActionDone → crush run_complete |
| `crush_types-from-crush-reasoning` | FromCrush: reasoning part → ActionThink |
| `crush_types-from-crush-text` | FromCrush: text part → ActionMessage |
| `crush_types-from-crush-tool-call` | FromCrush: tool_call part → ActionToolCall |
| `crush_types-from-crush-tool-result` | FromCrush: tool_result part → skipped |
| `crush_types-from-crush-finish-error` | FromCrush: finish error → ActionError |
| `crush_types-from-crush-finish-no-error` | FromCrush: finish end_turn → no action |
| `crush_types-from-crush-mixed-parts` | FromCrush: text + tool_call in one message |
| `crush_types-from-crush-non-assistant` | FromCrush: user/system message → skipped |
| `crush_types-from-crush-agent-event-error` | FromCrush: agent_event error → ActionError |
| `crush_types-from-crush-agent-event-response` | FromCrush: agent_event response → no error |
| `crush_types-from-crush-run-complete` | FromCrush: run_complete success → ActionDone |
| `crush_types-from-crush-run-complete-error` | FromCrush: run_complete with error text |
| `crush_types-from-crush-run-complete-cancelled` | FromCrush: run_complete cancelled=true |

## Decision Tree: Crush Server-Mode SSE Integration

The `Target="crush_server"` mode starts a live crush server, sends prompts
via HTTP, captures SSE events, and runs them through `FromCrush`.

```
crush_integration/
├── SETUP.md (grouping: Target="crush_server", ModelName="deepseek-v4-pro")
│
├── crush-integration-real/            ✓ Happy path: start server, prompt,
│   │                                       capture SSE, verify "paris"
│   ├── SETUP.md: Prompt="one word of French capital", HostPort=0
│   └── ASSERT.md: Output contains ActionMessage with "paris" (case-insensitive)
│
└── crush-integration-binary-not-found/  ✗ Explicit CrushPath not on disk
    ├── SETUP.md: CrushPath="/nonexistent/path/to/crush"
    └── ASSERT.md: non-nil error referencing missing binary
```

### New Test Leaves

| Leaf | Description |
|---|---|
| `crush-integration-real` | End-to-end: start crush server, prompt, SSE capture, FromCrush, verify "paris" |
| `crush-integration-binary-not-found` | Explicit CrushPath points to nonexistent file → error or skip |

## How to Run

```sh
doctest test ./agent/event/tests
```
