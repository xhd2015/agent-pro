# codex-concurrent-open-bind

Regression tree for **concurrent codex-tty `--open` on the same workspace**:
each agent-run session must bind a **distinct** Codex `runner_session_id`
matched to its own open prompt (not the newest cwd-matched rollout).

Converted from crime scene:
`~/.sandbox/transcripts/2026-08-12T104200Z-crime-scene-codex-concurrent-bind.md`
(**REPRODUCED**: two concurrent opens shared one codex uuid while two rollouts
existed; wrong bind steals the newer thread).

# DSN (Domain Specific Notion)

**Participants**

- **agent-run** — two concurrent `run --open` (different `--session-id` + prompts)
- **llm-mock-run-codex** + sibling **llm-mock** — real Codex TUI, mock Responses API
- **Shared workspace** — same `--dir` for both opens (eval-like: many questions, one cwd)
- **Shared CODEX_HOME** — same discovery root / rollouts directory
- **Isolated AGENT_RUN_HOME** — two session dirs under one home

**Desired behavior (product)**

```
concurrent:
  agent-run run --open --session-id A --dir WS ... -- "QUESTION_A"
  agent-run run --open --session-id B --dir WS ... -- "QUESTION_B"
  (same CODEX_HOME, same WS)

  -> meta(A).runner_session_id = uuidA  (non-empty)
  -> meta(B).runner_session_id = uuidB  (non-empty)
  -> uuidA != uuidB
  -> first real user prompt in rollout(uuidA) matches QUESTION_A
  -> first real user prompt in rollout(uuidB) matches QUESTION_B
```

**Before fix (bug)**

```
both metas bind the newest cwd-matched rollout uuid (same id)
  -> duplicate answers / wrong resume identity under concurrent eval
```

**Fix:** prompt-matched discovery + multi-candidate fail-closed
(`pkgs/agenttty` `DiscoverCodexSessionID` / `scanActiveCodexTranscripts`).

## Version

0.0.1

## Decision Tree

```
cmd/agent-run/tests/codex-concurrent-open-bind/
├── DOCTEST.md
├── SETUP.md
└── concurrent-same-workspace-distinct-ids/   # primary crime-scene leaf → RED until fix
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `concurrent-same-workspace-distinct-ids` | two concurrent opens, same dir → distinct correct binds |

## How to Run

```sh
# skip if codex not on PATH (Setup t.Skip)
# labeled e2e+codex — not in default unlabeled discovery

doctest test --label e2e --label codex \
  ./cmd/agent-run/tests/codex-concurrent-open-bind

doctest test -v --label e2e --label codex \
  ./cmd/agent-run/tests/codex-concurrent-open-bind/concurrent-same-workspace-distinct-ids
```

Expect **PASS** after prompt-aware / fail-closed open-bind (never “newest cwd only”).

## Types

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

type Request struct {
	RepoRoot  string
	TempDir   string
	Home      string // AGENT_RUN_HOME
	CodexHome string
	Workspace string

	AgentRun        string
	LLMMock         string
	LLMMockRunCodex string
	MockConfigFile  string

	SessionIDA string
	SessionIDB string
	PromptA    string
	PromptB    string

	ExecTimeout time.Duration
	Env         []string
}

type CmdResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

type Response struct {
	OpenA CmdResult
	OpenB CmdResult

	MetaA map[string]any
	MetaB map[string]any

	RunnerSessionIDA string
	RunnerSessionIDB string

	CodexIDs []string

	// First real user prompt text per bound codex id (from rollout jsonl).
	PromptByCodexID map[string]string

	Elapsed time.Duration
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	return runConcurrentSameWorkspace(t, req)
}
```
