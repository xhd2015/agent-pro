# GrokAgent CLI Integration Tests

Tests for `GrokAgent` in `agent/cli/grok/grok.go`, which invokes the `grok` CLI
binary and parses its streaming JSON output to implement the `registry.Agent`
interface.

These are integration tests that require a real `grok` binary on PATH
(install with `curl -fsSL https://x.ai/cli/install.sh | bash`).
Tests are skipped if the binary is not found.

## Decision Tree

```
operation mode?
├── Ask() ──► ask/                        (grouping: Ask() operation)
│   │
│   └── session management?
│       ├── fresh ──► ask/fresh/           (grouping: fresh session)
│       │   │
│       │   └── model specified?
│       │       ├── no  ──► basic-query/   (default model)
│       │       │   Ask("what is the capital of France? answer in one word")
│       │       │   ├── Answer contains "paris"
│       │       │   ├── SessionID non-empty, persisted in LastSessionID
│       │       │   ├── Events non-empty, all valid JSON
│       │       │   ├── Has "text" event
│       │       │   └── Has "end" event with sessionId
│       │       │
│       │       └── yes ──► model-override/
│       │           Ask("what is the capital of France? answer in one word", Model="grok-build")
│       │           ├── Answer non-empty
│       │           └── No error
│       │
│       └── resume ──► session-resume/     (session resume)
│           Ask("what is the capital of France? answer in one word")
│           → capture LastSessionID
│           Ask("what did I ask?", SessionID=lastID)
│           ├── Answer references "french" or "capital"
│           ├── SessionID non-empty
│           └── Session persisted across turns
│
└── ListModels() ──► list-models/
    Models()
    ├── Result non-empty (has models)
    ├── Contains "grok-composer-2.5-fast"
    └── No error
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `ask/fresh/basic-query` | Fresh `Ask()` for "one word of French capital" — validates answer, session ID, and JSON event stream (text, end events) |
| 2 | `ask/fresh/model-override` | Fresh `Ask()` with `Model="grok-build"` — validates answer non-empty, no error |
| 3 | `ask/session-resume` | Multi-turn: initial query establishes session, resume verifies context retention — validates answer references prior question, session persisted |
| 4 | `list-models` | `ListModels()` returns model list containing `grok-composer-2.5-fast` |

## How to Run

```sh
# Vet the tree structure (no compilation):
doctest vet ./agent/cli/grok/tests

# Build and run (requires grok binary on PATH):
doctest test -v ./agent/cli/grok/tests
```
