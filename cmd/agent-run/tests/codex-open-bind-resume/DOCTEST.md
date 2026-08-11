# codex-open-bind-resume

Regression tree for **codex-tty `--open` bind + auto-send-or-resume same Codex id**.

Converted from the local repro: open with `llm-mock-run-codex` never writes
`meta.runner_session_id`, so a later `--auto-send-or-resume --open` starts a
**second** Codex conversation under the same agent-run session id.

# DSN

**Participants**

- **agent-run** — `run --open`, `send`, `status`, `run --auto-send-or-resume --open`
- **llm-mock-run-codex** + sibling **llm-mock** — real Codex TUI, mock Responses API
- **Isolated homes** — `AGENT_RUN_HOME`, `CODEX_HOME` / `LLM_MOCK_CODEX_HOME` /
  `--agent-runner-config-home` (same dir)
- **Codex rollouts** — `CODEX_HOME/sessions/*/*/*/rollout-*-<uuid>.jsonl`

**Desired behavior (product)**

```
run --open --session-id S --agent-runner codex-tty \
  --agent-runner-binary llm-mock-run-codex \
  --agent-runner-config-home <codex-home> --dir <ws> -- "OPEN_MARKER"
  -> meta.runner_session_id = <codex-uuid-1>   # BIND ON OPEN (Fix A)
  -> one rollout for uuid-1

end session (send /exit and/or kill keep-alive)

run --auto-send-or-resume --open --session-id S ... -- "FOLLOW_UP"
  -> ModeResume: same meta.runner_session_id
  -> still one codex conversation (uuid-1), not a new uuid-2
```

**Today (bug)**

```
open  -> meta has NO runner_session_id; status runner unbound
resume -> second rollout uuid-2; still unbound
```

## Version

0.0.1

## Decision Tree

```
cmd/agent-run/tests/codex-open-bind-resume/
├── DOCTEST.md
├── SETUP.md
└── open-then-auto-resume-same-codex-id/   # primary repro → red until Fix A
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `open-then-auto-resume-same-codex-id` | open binds; end; auto-send-or-resume keeps same codex id |

## How to Run

```sh
# skip if codex not on PATH (leaf t.Skip)
# labeled e2e+codex — not in default unlabeled discovery

doctest test --label e2e --label codex \
  ./cmd/agent-run/tests/codex-open-bind-resume

doctest test -v --label e2e --label codex \
  ./cmd/agent-run/tests/codex-open-bind-resume/open-then-auto-resume-same-codex-id
```

Expect **FAIL** until open binds `runner_session_id` (Fix A).

## Types

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

type Request struct {
	RepoRoot string
	TempDir  string
	Home     string // AGENT_RUN_HOME
	CodexHome string
	Workspace string

	AgentRun        string
	LLMMock         string
	LLMMockRunCodex string
	MockConfigFile  string

	SessionID      string
	OpenPrompt     string
	FollowupPrompt string
	ExecTimeout    time.Duration

	Env []string
}

type CmdResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

type Response struct {
	SessionID string

	Open     CmdResult
	SendExit CmdResult
	Resume   CmdResult

	StatusAfterOpen   string
	StatusAfterEnd    string
	StatusAfterResume string

	MetaAfterOpen   map[string]any
	MetaAfterResume map[string]any

	RunnerSessionIDAfterOpen   string
	RunnerSessionIDAfterResume string

	CodexIDsAfterOpen   []string
	CodexIDsAfterResume []string

	KilledServe bool
	Elapsed     time.Duration
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	return runOpenThenAutoResume(t, req)
}
```
