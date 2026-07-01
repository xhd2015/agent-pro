# PiAgent CLI Integration Tests

Tests for `PiAgent` in `agent/cli/pi/pi.go`, which invokes the `pi` CLI binary and
parses its JSON-lines output to implement the `registry.Agent` interface.

These are integration tests that require a real `pi` binary on PATH and valid
API keys configured in `~/.pi/agent/`. Tests are skipped if the binary is not found.

## Version

0.0.2

## DSN (Domain Specific Notion)

Participants:

- **PiAgent** — an in-process adapter that implements `registry.Agent`. It
  resolves the `pi` binary, spawns it, scans its JSON-lines stdout, and
  converts each line into a canonical `AgentEvent` JSONL line on `RawLog`.
- **pi binary** — the CLI subprocess invoked by `PiAgent.Ask`. It emits
  JSON-lines events: `message_start`, `message_update` (assistant content
  deltas), `message_end`, and a terminal `agent_end` carrying the session id.
- **registry.Agent** — the contract `PiAgent` satisfies: `Ask`,
  `FindAgentPath`.

Behaviors:

- `Ask(ctx, question, opts, onDelta)` resolves the binary, spawns the process
  with optional `--session-id`, streams stdout line by line through the event
  writer, accumulates `message_update` content into the answer, captures
  `LastSessionID` from the session header, and returns the answer.
- `FindAgentPath(env)` is `env.LookPath("pi")`, else an error mentioning "pi".
- Session resume: a second `Ask` reuses `LastSessionID` via
  `AskOptions.SessionID` (mapped to pi's `--session-id`).

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

```go
import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/xhd2015/agent-pro/agent/cli/pi"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/exec"
)


type Request struct {
	Prompt       string
	ResumePrompt string
	Model        string
	Env          []string
}

type Response struct {
	Answer    string
	SessionID string
	Events    []json.RawMessage
}

func Run(t *testing.T, req *Request) (*Response, error) {
	paths := &exec.PathsConfig{
		RootDirName: ".pi-agent-test",
		DataDirName: "data",
		BinDirName:  "bin",
	}
	env := exec.NewEnv(paths, "PI_AGENT_TEST_ROOT")

	piPath, err := pi.FindAgentPath(env)
	if err != nil {
		t.Skip("pi not found in PATH; skip integration test")
		return &Response{}, nil
	}

	agent := &pi.PiAgent{
		AgentPath:    piPath,
		SettingsPath: "",
		Workspace:    t.TempDir(),
		Env:          env,
	}

	ctx := t.Context()
	var rawLogBuf bytes.Buffer
	opts := &registry.AskOptions{
		Model:  req.Model,
		RawLog: &rawLogBuf,
	}
	answer, err := agent.Ask(ctx, req.Prompt, opts, nil)
	events := parseRawLog(rawLogBuf)
	if err != nil {
		return &Response{Answer: answer, SessionID: agent.LastSessionID, Events: events}, err
	}

	if req.ResumePrompt != "" {
		sessionID := agent.LastSessionID
		if sessionID == "" {
			t.Fatalf("expected session ID to be captured from first query response:\n%s", answer)
		}
		opts.SessionID = sessionID
		answer2, err := agent.Ask(ctx, req.ResumePrompt, opts, nil)
		events2 := parseRawLog(rawLogBuf)
		return &Response{Answer: answer2, SessionID: agent.LastSessionID, Events: events2}, err
	}

	return &Response{Answer: answer, SessionID: agent.LastSessionID, Events: events}, nil
}
```
