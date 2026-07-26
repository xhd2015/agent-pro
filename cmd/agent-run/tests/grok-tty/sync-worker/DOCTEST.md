# Grok-TTY Sync Worker Integration Tests

Integration doc-style tests for the grok sync worker fix at the **agent-run web**
layer. Verifies `events.jsonl` (not SSE/Playwright) has exactly one user event per
prompt when rapid follow-ups arrive within the overlapping tail window.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run web** — `POST /api/agent-run/sessions` creates grok-tty session;
  `POST .../messages` enqueues follow-ups; must call `agentsync.EnsureGrokSync`
  (not `startGrokFollowUpEventTail` per message).
- **Grok sync worker** (`pkgs/agentsync`) — single persistent tail per session;
  discovery bootstrap when `runner_session_id` unset; appends to
  `AGENT_RUN_HOME/sessions/grok-tty/<id>/events.jsonl`.
- **llm-mock-run-grok** — chrome hook holds PTY alive (`--keep-tty` semantics).
- **Fake grok session** — synthetic `updates.jsonl` under temp `GROK_HOME`;
  delayed scheduler appends turn completions after web POST.
- **Send queue** — web follow-up enqueues via `agentsend`; drainer injects prompt.
- **Test harness** — session-scoped binary build; isolated `AGENT_RUN_HOME`;
  HTTP probes; `events.jsonl` duplicate counting (file-level assertion).

**Behaviors**

- **I1** — POST follow-up A (`hello?`), wait `done`, POST follow-up B
  (`what did I say?`) within overlap window; `events.jsonl` has exactly one user
  line per prompt (reproduces session `web_d6b4b203cc9ff71a` duplicate bug).
- **I2** — POST create with prompt only (no pre-set grok session id); delayed grok
  mock materializes `updates.jsonl`; `events.jsonl` receives user + assistant
  (reproduces `web_a5939cab5f4c7bfe` empty-chat bug).
- **I3** — Seed `finished` session with empty `events.jsonl`, `initial_prompt` in
  meta, grok `updates.jsonl` on disk; **only** `GET` session detail triggers sync;
  `events.jsonl` gains user + assistant; `GrokSyncWorkerActive` true.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/grok-tty/sync-worker/
├── DOCTEST.md
├── SETUP.md
└── web/
    ├── SETUP.md
    ├── rapid-followups-no-duplicate-events/   # I1: file-level events.jsonl dedupe
    ├── create-session-gets-events/            # I2: delayed grok → events.jsonl
    └── open-session-starts-sync/              # I3: GET session detail → sync
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| I1 | `web/rapid-followups-no-duplicate-events` | Web POST A → done → POST B within overlap; one user event per prompt in `events.jsonl` |
| I2 | `web/create-session-gets-events` | POST create with prompt; delayed grok mock; `events.jsonl` has user + assistant |
| I3 | `web/open-session-starts-sync` | Seed finished + empty events; GET session detail only; sync populates `events.jsonl` |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./cmd/agent-run/tests/grok-tty/sync-worker                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./cmd/agent-run/tests/grok-tty/sync-worker
doctest test --label-all ./cmd/agent-run/tests/grok-tty/sync-worker

doctest vet ./cmd/agent-run/tests/grok-tty/sync-worker
doctest test ./cmd/agent-run/tests/grok-tty/sync-worker   # RED before implement

doctest test -v ./cmd/agent-run/tests/grok-tty/sync-worker/web/rapid-followups-no-duplicate-events
doctest test -v ./cmd/agent-run/tests/grok-tty/sync-worker/web/create-session-gets-events
doctest test -v ./cmd/agent-run/tests/grok-tty/sync-worker/web/open-session-starts-sync
```

Regression:

```sh
doctest test ./cmd/agent-run/tests/grok-tty/run/streams-second-turn-after-completed
doctest test ./pkgs/agenttty/tests/grok-updates-tail
```

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

type GrokUpdatesSchedule struct {
	Delay       time.Duration
	Lines       []string
	UpdatesPath string
	OnFire      func()
}

type Request struct {
	RepoRoot string
	TempDir  string
	Home     string
	AgentRun string
	LLMMockRunGrok string
	GrokHome       string
	GrokSessionUUID string
	GrokUpdatesPath string
	GrokUpdatesSchedules []GrokUpdatesSchedule
	Env            []string

	Runner    string
	SessionID string
	PromptA   string
	PromptB   string
	ReplyA    string
	ReplyB    string

	WebToken   string
	WebBaseURL string
	WebCmd     *exec.Cmd
	webStderr  *bytes.Buffer

	Mode              string // web-rapid-followups | web-create-session-events | web-open-session-starts-sync
	FollowUpGap         time.Duration
	ChromeHoldSeconds   int
	CompletionDelayTurn1 time.Duration
	CompletionDelayTurn2 time.Duration
	ProbeTimeout        time.Duration
}

type Response struct {
	EventsFilePath  string
	EventsFileLines []string
	EventsParsed    []map[string]any
	UserCountA      int
	UserCountB      int
	UserCount       int
	AssistantFound  bool
	RunnerSessionID string
	WorkerActive    bool
	GetSessionStatus int
	DoneCount       int
	FollowUpBStatus int
	FollowUpBBody   string
	Stdout          string
	Stderr          string
	ExitCode        int
	Err             error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	switch req.Mode {
	case "web-rapid-followups":
		return runWebRapidFollowups(t, req)
	case "web-create-session-events":
		return runWebCreateSessionEvents(t, req)
	case "web-open-session-starts-sync":
		return runWebOpenSessionStartsSync(t, req)
	default:
		return nil, fmt.Errorf("unknown mode %q", req.Mode)
	}
}
```