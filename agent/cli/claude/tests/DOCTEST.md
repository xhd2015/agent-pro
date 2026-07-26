# ClaudeAgent CLI Integration Tests

Tests for `ClaudeAgent` in `agent/cli/claude/claude.go`, which invokes the
`claude` CLI binary in headless mode (`claude -p <question> --output-format
stream-json --verbose`) and parses its streaming NDJSON output via the
`agent/event/claude_types` package to implement the `registry.Agent`
interface.

These are integration tests that require a real `claude` binary on PATH.
Tests are skipped if the binary is not found, or if `CLAUDE_SKIP_INTEGRATION=1`.

## Version

0.0.2

## DSN (Domain Specific Notion)

Participants:

- **ClaudeAgent** — an in-process adapter that implements `registry.Agent`. It
  resolves the `claude` binary, spawns it headless, scans its NDJSON stdout,
  and converts each line into a canonical `AgentEvent` JSONL line on `RawLog`
  via a `ClaudeEventWriter`.
- **claude binary** — the headless CLI subprocess invoked as
  `claude -p <question> --output-format stream-json --verbose` (the
  `--verbose` flag is required). It emits one JSON object per line: `system`
  init, `assistant` content blocks (text / thinking / tool_use), `user`
  tool_result echoes, and a terminal `result` (success or error) carrying the
  `session_id`.
- **claude_types.FromClaude** — the mapping layer that turns native
  `StreamEvent`s into canonical `AgentEvent`s (`step_start`, `message`,
  `think`, `tool_call`, `done`, `error`). `user` tool_result lines are
  skipped (they fold into the preceding tool call).
- **ClaudeEventWriter** — coalesces each native NDJSON line into one or more
  `AgentEvent` JSONL lines on a `RawLog` writer.
- **registry.Agent** — the contract `ClaudeAgent` satisfies: `Ask`,
  `ListModels`, `FindAgentPath`. `ListModels` returns `nil, nil` because
  `claude` has no model-listing command.

Behaviors:

- `Ask(ctx, question, opts, onDelta)` resolves the binary, builds the args
  (optional `--model`, `--resume`), spawns the process, streams stdout line by
  line through the event writer, accumulates assistant `text` blocks into the
  answer, captures `LastSessionID` from any event carrying `session_id`, and
  returns the answer (falling back to `result.result`).
- `ListModels(ctx)` returns `nil, nil` (unsupported).
- `FindAgentPath(env)` is `env.LookPath("claude")`, else an error mentioning
  "claude".

## Decision Tree

```
operation mode?
├── Ask() ──► ask/                              (grouping: Ask() operation)
│   │
│   └── session management?
│       ├── fresh ──► ask/fresh/                 (grouping: fresh session)
│       │   │
│       │   └── model specified?
│       │       ├── no  ──► basic-query/         [slow, heavy] default model
│       │       │   Ask("Reply with exactly the word: pong")
│       │       │   ├── Answer contains "pong"
│       │       │   ├── SessionID non-empty (captured from system/result)
│       │       │   ├── Events non-empty, all valid JSON
│       │       │   ├── ≥1 AgentEvent type "message"
│       │       │   └── No error
│       │       │
│       │       └── yes ──► model-override/      [slow, heavy] Model="haiku"
│       │           Ask(..., Model="haiku")
│       │           ├── Answer non-empty
│       │           └── No error
│       │
│       └── resume ──► session-resume/           [slow, heavy] two-turn resume
│           Ask(...) → capture LastSessionID
│           Ask(resume, SessionID=lastID)
│           ├── Answer non-empty
│           ├── SessionID non-empty
│           └── No error
│
├── ListModels() ──► list-models/                (nil models, unsupported)
│   ListModels()
│   ├── Models is nil
│   └── No error
│
└── binary-not-found/                            (path resolution error)
    ClaudeAgent{AgentPath:"/nonexistent/claude"}.Ask
    └── Error non-nil, mentions "claude"
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `ask/fresh/basic-query` | Fresh `Ask()` for a deterministic word — validates answer, session ID, and JSON event stream (≥1 `message` event) |
| 2 | `ask/fresh/model-override` | Fresh `Ask()` with `Model="haiku"` — validates answer non-empty, no error |
| 3 | `ask/session-resume` | Multi-turn: initial query establishes session, resume reuses `LastSessionID` — validates answer non-empty, session persisted |
| 4 | `list-models` | `ListModels()` returns nil models, nil error (claude has no model-listing command) |
| 5 | `binary-not-found` | `ClaudeAgent{AgentPath:"/nonexistent/claude"}.Ask` returns a non-nil error mentioning "claude" |

## How to Run

```sh
# Vet the tree structure (no compilation):
doctest vet ./agent/cli/claude/tests

# Build and run (default: skips slow,heavy leaves; skips when claude absent):
doctest test ./agent/cli/claude/tests

# Run the live LLM-calling leaves (requires claude binary on PATH):
doctest test --label 'slow && heavy' ./agent/cli/claude/tests
```

```go
import (

	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/cli/claude"
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

	// AgentPath overrides the resolved binary path. When non-empty, Run does
	// NOT skip on a missing binary (used by binary-not-found).
	AgentPath string
}

type Response struct {
	Answer    string
	SessionID string
	Events    []json.RawMessage
	Models    []string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.Operation == "" {
		req.Operation = OpAsk
	}

	// binary-not-found: caller supplies an explicit (bogus) AgentPath; do not
	// resolve or skip — let Ask surface the error.
	if req.AgentPath == "" {
		if os.Getenv("CLAUDE_SKIP_INTEGRATION") == "1" {
			t.Skip("CLAUDE_SKIP_INTEGRATION=1; skip claude integration test")
			return &Response{}, nil
		}
		paths := &exec.PathsConfig{
			RootDirName: ".claude-agent-test",
			DataDirName: "data",
			BinDirName:  "bin",
		}
		env := exec.NewEnv(paths, "CLAUDE_AGENT_TEST_ROOT")
		claudePath, err := claude.FindAgentPath(env)
		if err != nil {
			t.Skip("claude not found in PATH; skip integration test")
			return &Response{}, nil
		}
		req.AgentPath = claudePath
	}

	paths := &exec.PathsConfig{
		RootDirName: ".claude-agent-test",
		DataDirName: "data",
		BinDirName:  "bin",
	}
	env := exec.NewEnv(paths, "CLAUDE_AGENT_TEST_ROOT")

	agent := &claude.ClaudeAgent{
		AgentPath:    req.AgentPath,
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
