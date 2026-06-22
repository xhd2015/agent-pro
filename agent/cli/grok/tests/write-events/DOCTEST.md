# Grok writeAgentEventsFromGrokLine Tests

Unit tests for `writeAgentEventsFromGrokLine` in `agent/cli/grok/grok.go`.
No grok CLI binary required.

## Decision Tree

```
write-events/
├── DOCTEST.md
├── SETUP.md
└── thought-streaming-deltas/   Per-word thought lines → 1 coalesced think event (RED)
```

## How to Run

```sh
doctest test ./agent/cli/grok/tests/write-events/...
```