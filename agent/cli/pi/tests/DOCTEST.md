# PiAgent CLI Integration Tests

Tests for `PiAgent` in `agent/cli/pi/pi.go`, which invokes the `pi` CLI binary and
parses its JSON-lines output to implement the `registry.Agent` interface.

These are integration tests that require a real `pi` binary on PATH and valid
API keys configured in `~/.pi/agent/`. Tests are skipped if the binary is not found.

## Decision Tree

```
pi-binary?
├── NOT FOUND ──► SKIP ALL (t.Skip)
└── FOUND
    └── PiAgent.Ask() operation
        ├── basic-query/       (fresh session)
        │   Ask("one word of French capital")
        │   ├── Answer contains "paris"
        │   ├── SessionID non-empty
        │   ├── Events non-empty, all valid JSON
        │   ├── Has "message_update" event
        │   └── Has "agent_end" event
        │
        └── session-resume/    (resume prior session)
            Ask("one word of French capital") → get SessionID
            Ask("what did I ask?", SessionID)
            ├── Answer references "french"/"capital"
            ├── SessionID non-empty
            ├── Events non-empty, all valid JSON
            ├── Has "message_start" event
            ├── Has "message_update" event
            ├── Has "message_end" event
            └── Has "agent_end" event
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `basic-query` | Fresh `Ask()` for "one word of French capital" — validates answer, session ID, and JSON event stream (message_update, agent_end) |
| 2 | `session-resume` | Multi-turn: initial query establishes session, resume verifies context retention — validates answer, session ID, and JSON event stream (message_start, message_update, message_end, agent_end) |

## How to Run

```sh
# Vet the tree structure (no compilation):
doctest vet ./agent/cli/pi/tests

# Build and run (requires pi binary on PATH and API keys configured):
doctest test -v ./agent/cli/pi/tests
```
