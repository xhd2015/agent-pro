# agent-run send queue Tests

Doc-style tests for the per-session central send queue (`pkgs/agentsend`) wired
through `agent-run send` / `agent-run tty send` and `agent-run msg status|cancel`.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `send` and `tty send` enqueue follow-up messages;
  `msg cancel` removes pending entries; `msg status` reports pending or delivered.
  Successful enqueue prints exactly one
  stdout line: the session-local message id (`msg_1`, `msg_2`, …).
- **agentsend queue** — FIFO JSONL at
  `AGENT_RUN_HOME/send-queue/<runner>/<terminal_session_id>.jsonl` with flock on
  companion `.lock` file. Message ids are monotonic per session queue.
- **Queue drainer** — elected by flock on next send; waits until terminal
  writable, injects head via `ttywatch.SendMessage`, dequeues on success.
- **TTY session registry** — resolves CLI session id to live stub-tty (or other
  runner) process via `agenttty.ResolveByTerminalID`.
- **stub-tty runner** — test-only TTY behind `AGENT_RUN_ENABLE_STUB_TTY=1`;
  scenario JSON controls busy/idle writable state and scrollback timing.
- **ptywrap server** — adhoc listener from stub-tty run; drainer uses server-side
  `WriteInput` for delivery.

**Behaviors**

- Default send: enqueue → print id → poll until this message absent from queue
  (delivered).
- `--no-wait`: enqueue → print id → exit 0 immediately; delivery may happen later.
- `--max-wait DURATION`: enqueue → print id → wall-clock wait from enqueue;
  timeout removes only that message → exit 1.
- `msg status <session-id>/<message-id>`: prints `pending` or `delivered` to stdout.
- `msg cancel <session-id>/<message-id>`: pending → silent exit 0; missing or
  already delivered → exit 1 with stderr error.
- FIFO: head delivered first; later messages wait behind earlier pending entries.
- `agent-run send` delegates to the same queue logic as `agent-run tty send`.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/send-queue/
├── DOCTEST.md
├── SETUP.md                              # build agent-run, stub-tty + queue helpers
├── enqueue/                              # message-id stdout on successful enqueue
│   ├── SETUP.md                          # idle stub-tty background session
│   ├── first-send-prints-msg-1/          # default send → stdout msg_1\n, delivered
│   ├── second-send-prints-msg-2/         # sequential send → stdout msg_2\n
│   ├── no-wait-prints-id-immediately/    # --no-wait on busy → quick msg_N\n
│   └── max-wait-prints-id-before-wait/   # --max-wait prints id before blocking
├── wait/                                 # wait-mode delivery semantics
│   ├── SETUP.md
│   ├── default-waits-until-writable-then-delivers/  # busy→idle; blocks >10s, exit 0
│   ├── max-wait-times-out-removes-message/          # busy + --max-wait 2s → exit 1
│   └── no-wait-returns-before-delivery/             # busy + --no-wait <1s, not injected
├── fifo/
│   ├── SETUP.md
│   └── fifo-two-messages/                # --no-wait A then B → inject order A, B
├── cancel/
│   ├── SETUP.md
│   ├── cancel-pending-message/           # --no-wait then cancel → never injected
│   ├── cancel-unknown-id/              # bogus id → exit 1, stderr not found
│   └── cancel-after-delivered-fails/     # deliver then cancel → exit 1
├── errors/
│   ├── SETUP.md
│   ├── missing-args/                     # no session/message → exit 1, no stdout id
│   └── session-not-found/                # unknown session → exit 1, no stdout id
└── alias/
    ├── SETUP.md
    └── send-shortcut-same-as-tty-send/   # send and tty send share queue + id format
```

Parameter ranking (most → least significant):

1. **Operation** — send enqueue vs wait mode vs cancel vs errors vs alias
2. **Wait mode** — default (indefinite) vs `--no-wait` vs `--max-wait`
3. **Terminal writable state** — idle vs busy vs busy-then-idle
4. **Queue outcome** — delivered vs timed out vs cancelled vs not found
5. **CLI entrypoint** — `send` vs `tty send` (alias only)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `enqueue/first-send-prints-msg-1` | Default send prints `msg_1\n`, exit 0, message injected |
| 2 | `enqueue/second-send-prints-msg-2` | Second send on same session prints `msg_2\n` |
| 3 | `enqueue/no-wait-prints-id-immediately` | `--no-wait` returns quickly with id on stdout |
| 4 | `enqueue/max-wait-prints-id-before-wait` | `--max-wait` prints id before blocking for delivery |
| 5 | `wait/default-waits-until-writable-then-delivers` | Busy stub becomes idle; send blocks >10s then exit 0 |
| 6 | `wait/max-wait-times-out-removes-message` | Permanently busy + `--max-wait 2s` → exit 1, queue lacks id |
| 7 | `wait/no-wait-returns-before-delivery` | `--no-wait` on busy session returns in <1s with id |
| 8 | `fifo/fifo-two-messages` | Two `--no-wait` enqueues delivered FIFO (A then B) |
| 9 | `cancel/cancel-pending-message` | Cancel pending `--no-wait` message → exit 0, never injected |
| 10 | `cancel/cancel-unknown-id` | Cancel bogus id → exit 1, stderr mentions not found |
| 11 | `cancel/cancel-after-delivered-fails` | Cancel after delivery → exit 1 |
| 12 | `errors/missing-args` | Missing args → exit 1, no stdout id line |
| 13 | `errors/session-not-found` | Unknown session → exit 1, no stdout id line |
| 14 | `alias/send-shortcut-same-as-tty-send` | `send` and `tty send` share queue id semantics |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/send-queue
doctest test ./cmd/agent-run/tests/send-queue
doctest test -v ./cmd/agent-run/tests/send-queue/enqueue/first-send-prints-msg-1
doctest test -v ./cmd/agent-run/tests/send-queue/wait/max-wait-times-out-removes-message
```

```go
import (
	"os/exec"
	"testing"
	"time"
)

type Request struct {
	RepoRoot string
	TempDir  string
	Home     string
	AgentRun string
	Env      []string

	Operation string // enqueue | wait | fifo | cancel | errors | alias
	Action    string // leaf slug

	// stub-tty session
	EnableStubTTY     bool
	StubScenarioJSON  string
	StubScenarioPath  string
	TerminalSessionID string
	RunnerID          string
	BackgroundCmd     *exec.Cmd

	// send invocation
	SendArgs     []string
	SendMessage  string
	UseTTYSubcmd bool
	ExecTimeout  time.Duration
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error

	MsgID            string
	SecondStdout     string
	SecondMsgID      string
	CancelStdout       string
	CancelStderr       string
	CancelExitCode     int
	StatusBeforeStdout string
	StatusAfterStdout  string
	SendDuration     time.Duration
	IdLineLatency    time.Duration
	InjectedMessages []string
	QueueFilePath    string
	QueueHasMsgID    bool
	ShortcutStdout   string
	ShortcutMsgID    string
	TTYSubcmdStdout  string
	TTYSubcmdMsgID   string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return runSendQueueOp(t, req)
}
```