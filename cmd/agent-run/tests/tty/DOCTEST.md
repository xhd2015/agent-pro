# agent-run tty Subcommand Tests

Doc-style tests for `agent-run tty` subcommands — status, attach, and send
operations on live TTY sessions, plus the `agent-run attach` and `agent-run send`
shortcut aliases.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — dispatches `tty status`, `tty attach`, `tty send`, `tty kill`,
  and `attach` (alias for `tty attach`) and `send` (alias for `tty send`) against
  TTY runner sessions. Top-level `kill` mirrors `tty kill`.
- **TTY session registry** — JSON files at `AGENT_RUN_HOME/<runner>-registry/<session-id>.json`
  mapping session id to `pid`, `listen_addr`, and `created_at`. Persists while
  the terminal is alive; normally removed on run exit unless `--keep-tty` is set.
- **Session storage** — `AGENT_RUN_HOME/sessions/<runner>/<session-id>/meta.json`
  carries `terminal_session_id` for cross-referencing.
- **ptywrap server** — adhoc HTTP+WebSocket listener bound to `127.0.0.1:<port>`;
  exposes scrollback, input injection, and session management. Tests use a fake
  in-process HTTP server that mimics the ptywrap WebSocket protocol.
- **Screen status detection** — connects to the ptywrap WebSocket, reads scrollback,
  analyzes for banner markers, input-prompt, idle (persistent turn complete), or
  error states using the same detection functions in `groktty/capture.go`.

**Behaviors**

- `agent-run tty status <session-id>` reads the registry entry (deterministic
  search across `grok-tty-registry` and `codex-tty-registry`), reports pid,
  port, tty type, session id, session file path, start time, TCP reachability,
  and screen status. Output is human-readable text; `--json` emits structured JSON.
- `agent-run tty attach <session-id>` looks up the registry entry, connects via
  `ptyclient.Attach` to the hidden ptywrap listener, and pipes interactive I/O.
- `agent-run attach <session-id>` is an alias that delegates to the same logic
  as `tty attach`.
- `agent-run send <session-id> "msg"` is an alias that delegates to the same
  logic as `tty send`.
- `agent-run tty send <session-id> "msg"` resolves the terminal, connects a
  WebSocket, injects the prompt via the same path as `sendPromptToLiveTerminal`,
  captures the assistant response, and appends events to the session file.
- Registry search is deterministic: grok-tty-registry first, then codex-tty-registry.
  Stale entries (TCP unreachable) are cleaned; reachable entries matching the session
  id are returned.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/tty/
├── DOCTEST.md
├── SETUP.md                           # build agent-run, registry + fake ptywrap helpers
├── help/
│   └── lists-subcommands/             # tty --help lists status, attach, send, kill
├── status/                            # agent-run tty status <session-id>
│   ├── SETUP.md                       # writes mock registry entry (no live server needed)
│   ├── missing-session-id/            # no positional arg → error
│   ├── session-not-found/             # session id with no registry file → error
│   ├── registry-entry-valid/          # valid registry, all fields displayed
│   │   ├── human-readable/            # default output: pid, port, tty type, session id, ...
│   │   ├── with-json/                 # --json flag → valid JSON with same fields
│   │   └── tcp-unreachable/           # registry file but port closed → shows unreachable
│   └── screen-status/                 # read scrollback via fake ptywrap WebSocket
│       ├── SETUP.md                   # starts fake ptywrap HTTP+WS server with scrollback
│       ├── banner-detected/           # scrollback has GROK_TTY_BANNER → screen: banner
│       ├── idle-detected/             # scrollback has response text after prompt → screen: idle
│       └── no-server-reachable/       # fake listen addr but no server → screen: unknown
├── attach/                            # agent-run tty attach <session-id>
│   ├── SETUP.md                       # writes mock registry entry
│   ├── missing-session-id/            # no arg → error
│   ├── session-not-found/             # registry missing → error
│   └── connects-via-registry/         # valid registry → ptyclient.Attach connects
├── attach-shortcut/                   # agent-run attach <session-id> (alias)
│   ├── delegates-to-tty-attach/       # both attach and tty attach produce same error output
│   └── same-error-for-missing/        # attach bad-id gives same error as tty attach bad-id
├── send-shortcut/                     # agent-run send <session-id> "msg" (alias)
│   ├── delegates-to-tty-send/         # send bad-id gives same error as tty send bad-id
│   ├── sends-to-live-terminal/        # fake ptywrap WS; send injects prompt, captures response
│   └── same-error-for-missing-args/   # send with no args gives same error as tty send
└── send/                              # agent-run tty send <session-id> "msg"
    ├── SETUP.md                       # writes mock registry entry
    ├── missing-args/                  # missing session-id or message → error
    ├── session-not-found/             # registry missing → error
    ├── sends-to-live-terminal/        # fake ptywrap WS; send injects prompt, captures response
    └── terminal-unreachable/          # registry but port closed → error
```

Parameter ranking (most → least significant):

1. **Subcommand** — status vs attach vs send vs attach/send shortcuts vs --help
2. **Session state** — registry found + reachable vs missing vs expired
3. **Output format** — human-readable vs --json (status only)
4. **Screen state** — banner / input-prompt / idle / error / unknown (status only)
5. **Runner type** — grok-tty vs codex-tty registry (status: detected from registry dir)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `help/lists-subcommands` | `tty --help` lists status, attach, send, kill subcommands |
| 2 | `status/missing-session-id` | `tty status` without session id → error, usage hint |
| 3 | `status/session-not-found` | `tty status bogus-id` → error, session not found |
| 4 | `status/registry-entry-valid/human-readable` | Valid registry → pid, port, tty type, session id, path, start time, tcp reachable |
| 5 | `status/registry-entry-valid/with-json` | `--json` → valid JSON object with same fields |
| 6 | `status/registry-entry-valid/tcp-unreachable` | Registry file with closed port → shows "unreachable" |
| 7 | `status/screen-status/banner-detected` | Fake ptywrap scrollback with banner → screen: banner detected |
| 8 | `status/screen-status/idle-detected` | Fake ptywrap scrollback after prompt+response → screen: idle |
| 9 | `status/screen-status/no-server-reachable` | Listen addr but ptywrap server dead → screen: unknown |
| 10 | `attach/missing-session-id` | `tty attach` without session id → error |
| 11 | `attach/session-not-found` | `tty attach bogus-id` → error, session not found |
| 12 | `attach/connects-via-registry` | Valid registry + live ptywrap → attach probe OK |
| 13 | `attach-shortcut/delegates-to-tty-attach` | `attach` and `tty attach` with bad id produce identical errors |
| 14 | `attach-shortcut/same-error-for-missing` | `attach bogus-id` same error as `tty attach bogus-id` |
| 15 | `send-shortcut/delegates-to-tty-send` | `send` and `tty send` with bad id produce identical errors |
| 16 | `send-shortcut/sends-to-live-terminal` | `send` via fake ptywrap WS → inject prompt, captures response |
| 17 | `send-shortcut/same-error-for-missing-args` | `send` with no args same error as `tty send` with no args |
| 18 | `send/missing-args` | `tty send` without args → error |
| 19 | `send/session-not-found` | `tty send bogus-id "hi"` → error |
| 20 | `send/sends-to-live-terminal` | Fake ptywrap WS → send prompt, captures response, appends events |
| 21 | `send/terminal-unreachable` | Registry with closed port → send fails with error |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./cmd/agent-run/tests/tty                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./cmd/agent-run/tests/tty
doctest test --label-all ./cmd/agent-run/tests/tty

doctest vet ./cmd/agent-run/tests/tty
doctest test ./cmd/agent-run/tests/tty
doctest test -v ./cmd/agent-run/tests/tty/status/registry-entry-valid/human-readable
doctest test -v ./cmd/agent-run/tests/tty/send/sends-to-live-terminal
doctest test -v ./cmd/agent-run/tests/tty/send-shortcut/sends-to-live-terminal
```

```go
import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/doctest/session"
)

type RegistryEntryData struct {
	SessionID  string `json:"session_id"`
	ListenAddr string `json:"listen_addr"`
	PID        int    `json:"pid"`
	CreatedAt  string `json:"created_at"`
}

type Request struct {
	RepoRoot string
	TempDir  string
	Home     string
	AgentRun string
	Args     []string
	Env      []string

	Mode string // "" | "status-json" | "attach-probe" | "send-probe"

	RegistryDir       string
	RegistryEntryJSON string
	RegistrySessionID string

	StartFakePTYWrap   bool
	FakePTYWrapPort    int
	FakePTYWrapScrollback string

	AttachTimeout       time.Duration
	ExecTimeout         time.Duration
	BackgroundCmd       *exec.Cmd
	BackgroundCmdStderr *bytes.Buffer
	BackgroundCmdStdout *bytes.Buffer
	SendExpectResponse  bool
	SendResponseSubstr  string
	FakePTYInputReceived chan string
	FakePTYReady         chan struct{}
}

type Response struct {
	Stdout    string
	Stderr    string
	ExitCode  int
	Err       error
	JSONBody  map[string]any

	AttachProbeOK  bool
	AttachProbeErr string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	switch req.Mode {
	case "status-json":
		return runStatusJSON(t, req)
	case "attach-probe":
		return runAttachProbe(t, req)
	case "send-probe":
		return runSendProbe(t, req)
	default:
		return runAgentRun(t, req, req.Args...)
	}
}
```
