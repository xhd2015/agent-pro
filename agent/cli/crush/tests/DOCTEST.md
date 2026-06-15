# CrushAgent CLI Integration Tests

Tests for `CrushAgent` in `agent/cli/crush/crush.go`, which invokes the `crush`
CLI binary and parses its SSE-style JSON output to implement the `registry.Agent`
interface.

These are integration tests that require a real `crush` binary on PATH.
Tests are skipped if the binary is not found.

## Decision Tree

```
crush-binary?
├── NOT FOUND ──► SKIP ALL (t.Skip)
└── FOUND
    └── Ask() operation
        ├── basic-query/       (fresh session)
        │   Ask("one word of French capital") → contains "paris"
        │
        └── session-resume/    (resume prior session)
            Ask("one word of French capital") → get SessionID
            Ask("what did I ask?", SessionID) → references "French"/"capital"
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `basic-query` | Fresh Ask() for "one word of French capital" → answer contains "paris" |
| 2 | `session-resume` | Multi-turn: initial query establishes session, resume verifies context retention |

## How to Run

```sh
# Vet the tree structure (no compilation):
doctest vet ./agent/cli/crush/tests

# Build and run (requires crush binary on PATH):
doctest test -v ./agent/cli/crush/tests
```
