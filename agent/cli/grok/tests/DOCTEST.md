# GrokAgent CLI Integration Tests

Tests for `GrokAgent` in `agent/cli/grok/grok.go`, which invokes the `grok` CLI
binary and parses its streaming JSON output to implement the `registry.Agent`
interface.

These are integration tests that require a real `grok` binary on PATH
(install with `curl -fsSL https://x.ai/cli/install.sh | bash`).
Tests are skipped if the binary is not found.

## Version

0.0.2

## DSN (Domain Specific Notion)

Participants:

- **GrokAgent** — an in-process adapter that implements `registry.Agent`. It
  resolves the `grok` binary, spawns it, scans its streaming-JSON stdout, and
  converts each line into a canonical `AgentEvent` JSONL line on `RawLog` via a
  `GrokEventWriter`.
- **grok binary** — the CLI subprocess invoked by `GrokAgent.Ask`. It emits one
  JSON object per line: `thought` (per-word reasoning deltas), `text`
  (assistant content), tool `tool_started`/`tool_completed` pairs, and a
  terminal `end` event carrying the `sessionId`.
- **GrokEventWriter / writeAgentEventsFromGrokLine** — the mapping layer that
  turns native grok streaming lines into canonical `AgentEvent`s (`think`,
  `message`, `tool_call`, `done`). Per-word `thought` deltas coalesce into one
  think event; `end` carries the captured `LastSessionID`.
- **registry.Agent** — the contract `GrokAgent` satisfies: `Ask`,
  `ListModels`, `FindAgentPath`.

Behaviors:

- `Ask(ctx, question, opts, onDelta)` resolves the binary, builds the args
  (optional `--model`, `--resume`), spawns the process, streams stdout line by
  line through the event writer, accumulates assistant `text` into the answer,
  captures `LastSessionID` from the `end` event, and returns the answer.
- `ListModels(ctx)` queries the grok CLI for its available models.
- `FindAgentPath(env)` is `env.LookPath("grok")`, else an error mentioning
  "grok".
- Session resume: a second `Ask` reuses `LastSessionID` via
  `AskOptions.SessionID` (mapped to grok's `--resume`).

## Decision Tree

```
write-events/                            writeAgentEventsFromGrokLine (no CLI required)
└── thought-streaming-deltas             Per-word thought lines → 1 coalesced think event (RED)

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
| 5 | `write-events/thought-streaming-deltas` | Per-word grok `thought` lines → one coalesced `ActionThink` in events.jsonl (RED) |

## How to Run

```sh
# Vet the tree structure (no compilation):
doctest vet ./agent/cli/grok/tests

# Build and run (requires grok binary on PATH):
doctest test -v ./agent/cli/grok/tests
```

```go
import (

	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/cli/grok"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/exec"
	"github.com/xhd2015/doctest/session"
)


type Operation string

const (
	OpAsk        Operation = "ask"
	OpListModels Operation = "list-models"
)

type Request struct {
	Prompt       string
	ResumePrompt string
	Model        string
	Operation    Operation
}

type Response struct {
	Answer    string
	SessionID string
	Events    []json.RawMessage
	Models    []string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	paths := &exec.PathsConfig{
		RootDirName: ".grok-agent-test",
		DataDirName: "data",
		BinDirName:  "bin",
	}
	env := exec.NewEnv(paths, "GROK_AGENT_TEST_ROOT")

	grokPath, err := grok.FindAgentPath(env)
	if err != nil {
		t.Skip("grok not found in PATH; skip integration test")
		return &Response{}, nil
	}

	agent := &grok.GrokAgent{
		AgentPath:    grokPath,
		SettingsPath: "",
		Workspace:    t.TempDir(),
		Env:          env,
	}

	switch req.Operation {
	case OpListModels:
		return runListModels(t, agent)
	default:
		return runAsk(t, agent, req)
	}
}
```
