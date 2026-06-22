# FormatState Streaming Tests

These doc-style tests verify `print.FormatState.FormatLine` streaming coalescing
for consecutive `AgentEvent` deltas. Used by `traceSession` when rendering
events.jsonl lines.

## Decision Tree

```
format-state-streaming/
├── DOCTEST.md
├── SETUP.md
├── grok-thought-deltas/        ActionThink per-word deltas → 1 think block (RED)
└── message-deltas-coalesced/   ActionMessage deltas → 1 ASSISTANT block
```

## How to Run

```sh
doctest test ./agent/event/print/tests/format-state-streaming/...
```