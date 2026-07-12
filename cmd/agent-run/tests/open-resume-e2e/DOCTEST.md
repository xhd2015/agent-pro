# agent-run open → Paris → /exit → resume --open E2E (mock-grok)

Stable, non-flaky integration tree for the full open/resume lifecycle with
**no real grok**. The PTY child is `llm-mock-run-grok` driven by a deterministic
`LLM_MOCK_RUN_GROK_COMMAND` multi-turn hook.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — dispatches `run --open`, `snapshot`, `send`, `status`, and
  `resume --open` under an isolated `AGENT_RUN_HOME`.
- **llm-mock-run-grok** — built from `./agent/llm/llm-mock/llm-mock-run-grok` and
  passed as `--agent-runner-binary`. Isolates `GROK_HOME` via
  `--agent-runner-config-home` (and announces it on stderr).
- **Mock hook** — `LLM_MOCK_RUN_GROK_COMMAND` shell orchestrator that:
  1. seeds `$GROK_HOME/sessions/<encoded-cwd>/<uuid>/updates.jsonl` with
     deterministic assistant text (**Paris** then **hello** markers),
  2. prints the same text to PTY scrollback for `snapshot`,
  3. on `/exit`, prints `grok --resume <uuid>` + `[Terminal exited]` then exits
     so status reports `runner.exited=true` while keep-alive may still hold the
     registry (zombie reclaim path).
- **Workspace dir** — project-like `--dir` under the leaf temp tree (not bare
  `/tmp`) so discovery cwd encoding matches real open flows.
- **Open attach test hook** — `AGENT_RUN_OPEN_ATTACH_INSTANT=1` makes auto-attach
  return without a controlling TTY (CI-safe).
- **TTY registry / keep-alive serve** — after `run --open`, serve stays for
  snapshot/send; after `/exit`, reclaim must free the id so `resume --open`
  does not fail with `already in use`.

**Behaviors**

```
# primary integration (mock-grok)
agent-run run --open --agent-runner grok-tty \
  --agent-runner-binary llm-mock-run-grok \
  --agent-runner-config-home <grok-home> \
  --session-id <id> --dir <workspace> \
  "one word of France capital"
  + AGENT_RUN_OPEN_ATTACH_INSTANT=1
  + LLM_MOCK_RUN_GROK_COMMAND multi-turn
  -> bind runner_session_id; snapshot/events show Paris
  -> agent-run send <id> /exit
  -> agent-run status: exited true; resume.ready
  -> agent-run resume --open <id> "hello"
     (same binary + config home)
  -> NOT already in use
  -> snapshot/events show proper followup text
```

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/open-resume-e2e/
├── DOCTEST.md
├── SETUP.md
└── mock-grok/                                 # llm-mock-run-grok (no real grok)
    ├── SETUP.md
    ├── open-paris-exit-resume-hello/          # M1 PRIMARY full flow
    ├── resume-no-followup/                    # M2 reopen only after exit
    ├── resume-open-no-submit-hello/           # M3 --no-submit with --open
    ├── live-send-followup-no-exit/            # M4 send hello while still live
    ├── double-resume/                         # M5 exit→resume→exit→resume
    ├── resume-while-live-denied/              # M6 resume before /exit → error
    ├── open-bind-prints-session/              # M7 open stderr grok session lines
    ├── send-two-followups-while-live/         # M8 two live sends
    ├── status-ready-after-exit/               # M9 resume.ready after exit
    └── resume-keep-tty-followup/              # M10 resume without --open + followup
```

Parameter ranking (most → least significant):

1. **Runner backend** — mock-grok only (real grok out of scope for this tree).
2. **Lifecycle scenario** — open / exit / resume / live-send variants.
3. **Isolation** — per-leaf `AGENT_RUN_HOME`, `GROK_HOME`, workspace, session id.

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `mock-grok/open-paris-exit-resume-hello` | M1: open → Paris → /exit → exited → resume --open hello |
| 2 | `mock-grok/resume-no-followup` | M2: after exit, resume without followup (reopen only) |
| 3 | `mock-grok/resume-open-no-submit-hello` | M3: resume --open --no-submit hello |
| 4 | `mock-grok/live-send-followup-no-exit` | M4: open → Paris → send hello (no /exit) |
| 5 | `mock-grok/double-resume` | M5: exit → resume → exit → resume again |
| 6 | `mock-grok/resume-while-live-denied` | M6: resume while still live → denied (use send) |
| 7 | `mock-grok/open-bind-prints-session` | M7: open stderr prints grok-tty + grok session |
| 8 | `mock-grok/send-two-followups-while-live` | M8: two live followups after open |
| 9 | `mock-grok/status-ready-after-exit` | M9: after /exit, resume.ready=yes + exited true |
| 10 | `mock-grok/resume-keep-tty-followup` | M10: resume keep-tty followup without --open |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/open-resume-e2e
doctest test ./cmd/agent-run/tests/open-resume-e2e
doctest test -v ./cmd/agent-run/tests/open-resume-e2e/mock-grok/open-paris-exit-resume-hello
```

```go
import (
	"os/exec"
	"testing"
	"time"
)

type Request struct {
	RepoRoot       string
	TempDir        string
	Home           string // AGENT_RUN_HOME
	Workspace      string // --dir project-like workspace
	GrokHome       string // --agent-runner-config-home
	AgentRun       string
	LLMMockRunGrok string
	Env            []string

	// Scenario selects Run branch.
	Scenario string

	SessionID       string
	GrokSessionUUID string
	OpenPrompt      string // "one word of France capital"
	FollowupPrompt  string // "hello"
	WantParis       string // "Paris"
	HelloMarker     string // deterministic resume assistant text
	SecondFollowup  string // optional second live send
	SecondMarker    string // optional second mock response marker

	// Flow toggles (leaf Setup).
	ResumeOpen   bool // include --open on resume (default true for most)
	ResumeNoOpen bool // force resume without --open
	NoSubmit     bool // resume --no-submit (requires --open)
	SkipExit     bool // do not send /exit
	SkipResume   bool // stop after open/exit/status
	ResumeTwice  bool // second exit+resume cycle
	ExpectResumeDenied bool // resume should fail with still-active / use send

	ParisWait   time.Duration
	ExitWait    time.Duration
	ResumeWait  time.Duration
	ExecTimeout time.Duration
}

type CmdResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

type Response struct {
	Open                CmdResult
	SendExit            CmdResult
	SendFollowup        CmdResult
	SendSecond          CmdResult
	StatusAfterExit     CmdResult
	StatusJSONAfterExit string
	StatusAfterOpen     CmdResult
	Resume              CmdResult
	Resume2             CmdResult
	ParisSnapshot       string
	ResumeSnapshot      string
	EventsFilePath      string
	EventsBlob          string
	HasParis            bool
	ExitedTrue          bool
	ResumeReady         bool
	HasHello            bool
	HasSecond           bool
	AlreadyInUse        bool
	ResumeDenied        bool
	BoundOnOpen         bool
	Elapsed             time.Duration
	Err                 error
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return runScenario(t, req)
}
```
