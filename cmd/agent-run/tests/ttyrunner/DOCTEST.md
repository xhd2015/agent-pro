# ttyrunner Package Tests

Doc-style tests for `github.com/xhd2015/agent-pro/pkgs/ttyrunner` — pluggable TTY
runner provider registry, unified session storage/resolver, `CheckWritable` /
sendable status, test-only `stub-tty` runner, and multi-attach write/observer policy.

# DSN (Domain Specific Notion)

**Participants**

- **Provider registry** (`pkgs/ttyrunner`) — registers TTY runners (`grok-tty`,
  `codex-tty`, test-only `stub-tty` behind `AGENT_RUN_ENABLE_STUB_TTY=1`). Each
  provider exposes `ID`, `RegistryDir` (`<id>-registry`), banner markers,
  `BuildArgv`, `DetectScreenStatus`, and `CheckWritable`.
- **Unified session storage** — canonical agent session under
  `AGENT_RUN_HOME/sessions/<runner>/<agent-session-id>/` with `meta.json`,
  `events.jsonl`, and new `tty.json` (denormalized TTY cross-ref). Live process
  index remains at `AGENT_RUN_HOME/<runner>-registry/<terminal-session-id>.json`.
- **Session resolver** — `ResolveByTerminalID`, `ResolveByAgentSession`, and
  `LookupSession` shim; replaces hardcoded registry candidate lists in CLI/web.
- **agent-run CLI** — `run --agent-runner=<tty>`, `tty status|attach|send`,
  and web terminal resolve through the resolver. `tty status` reports instantaneous
  `sendable` / `sendable_reason` / `sendable_state` from `CheckWritable`.
  `tty send` blocks until writable (10s timeout with provider reason) then uses
  **server-side** `WriteInput` (never client write attach).
- **ptywrap server** — adhoc per-run HTTP+WebSocket listener; multi-attach policy:
  first `attach_mode=interactive` claims unified write for session lifetime;
  later attaches are permanent observers (zero mutation); server controller always
  writes; output multiplexed to writer + observers; snapshot probes are ephemeral.
- **stub-tty runner** — full alt-screen mock TUI driven by
  `AGENT_RUN_STUB_TTY_SCENARIO` JSON (latency, screen frames, mock LLM events).

**Behaviors**

- Builtin providers register at init: `grok-tty`, `codex-tty`; `stub-tty` only
  when `AGENT_RUN_ENABLE_STUB_TTY=1`.
- TTY run dual-writes registry JSON + `sessions/.../tty.json`; updates
  `meta.terminal_session_id`.
- `ResolveByTerminalID` searches registered providers in registration order;
  skips stale (TCP unreachable) entries; enriches from `tty.json` / `meta.json`.
- `CheckWritable(scrollback)` per provider drives `tty send` wait loop and
  `tty status` sendable fields.
- Multi-attach: first interactive attach → writer; second+ → read-only forever;
  writer detach spends token (no promotion); observers cannot resize/delete/inject;
  `tty send` uses server write even when client holds unified write.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/ttyrunner/
├── DOCTEST.md
├── SETUP.md                              # build agent-run, ttyrunner helpers, stub env
├── provider-registry/
│   ├── SETUP.md
│   ├── lists-builtin-providers/          # IDs() → grok-tty, codex-tty (+ stub when enabled)
│   ├── is-tty-runner/                    # IsTTYRunner true/false per runner id
│   └── registry-dir-convention/          # RegistryDir == id + "-registry"
├── storage/
│   ├── SETUP.md                          # session + registry + tty.json helpers
│   ├── dual-write-tty-json/              # run creates sessions/.../tty.json + registry
│   ├── resolve-by-agent-session/         # meta.terminal_session_id → live registry
│   └── resolve-by-terminal-id/           # registry lookup enriches from tty.json
├── status-sendable/
│   ├── SETUP.md                          # fake ptywrap scrollback + tty status CLI
│   ├── idle-screen-sendable-yes/         # › + turn complete → sendable: yes
│   ├── busy-screen-sendable-no/          # codex working → sendable: no + reason
│   └── unreachable-sendable-no/          # TCP down → sendable: false
├── lookup/
│   ├── SETUP.md
│   ├── finds-grok-entry/                 # grok-tty-registry reachable
│   ├── finds-codex-entry/                # codex-tty-registry reachable
│   ├── deterministic-order/              # same id in both registries → first provider wins
│   ├── skips-stale-entry/                # stale grok removed; live codex returned
│   └── session-not-found/                # no registry file → error
├── stub-tty/
│   ├── SETUP.md                          # scenario JSON fixtures, stub-tty env
│   ├── run-creates-registry-and-tty-json/
│   ├── scenario-banner-delay/
│   ├── scenario-mock-llm-events/
│   ├── scenario-mock-screen-frames/
│   ├── scenario-declared-screen-status/
│   ├── attach-interactive/
│   ├── keep-tty-persists/
│   └── multi-attach/
│       ├── SETUP.md
│       ├── first-attach-writes-second-readonly/
│       ├── writer-detach-no-promotion-for-second/
│       ├── writer-plus-observer-both-receive-output/
│       ├── third-attach-after-writer-gone-still-readonly/
│       ├── multiple-observers-all-receive/
│       ├── observer-resize-ignored-pty-unchanged/
│       ├── observer-close-does-not-delete-session/
│       ├── snapshot-does-not-affect-writer/
│       ├── send-uses-server-write-not-client-attach/
│       ├── send-waits-until-writable-then-injects/
│       ├── send-times-out-with-reason-after-10s/
│       └── server-write-while-client-holds-unified-write/
└── integration/
    └── sealed-trees-regression/          # doc pointer; orchestrator runs sibling trees
```

Parameter ranking (most → least significant):

1. **Subsystem** — registry vs storage/resolver vs sendable vs lookup vs stub-tty vs integration
2. **Lookup outcome** — found + reachable vs stale skip vs not found
3. **Provider** — grok-tty vs codex-tty vs stub-tty
4. **Screen/writable state** — idle/sendable vs busy/not sendable vs unreachable
5. **Attach role** — unified writer vs observer vs snapshot vs server send
6. **Scenario timing** — immediate vs delayed banner/frames/events

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `provider-registry/lists-builtin-providers` | `IDs()` lists grok-tty and codex-tty; stub-tty only when env enabled |
| 2 | `provider-registry/is-tty-runner` | `IsTTYRunner` true for TTY ids, false for non-TTY runners |
| 3 | `provider-registry/registry-dir-convention` | Each provider `RegistryDir` is `<id>-registry` |
| 4 | `storage/dual-write-tty-json` | stub-tty run writes registry + `sessions/.../tty.json` |
| 5 | `storage/resolve-by-agent-session` | `ResolveByAgentSession` links agent id → terminal registry |
| 6 | `storage/resolve-by-terminal-id` | `ResolveByTerminalID` enriches from `tty.json` |
| 7 | `status-sendable/idle-screen-sendable-yes` | Idle scrollback → `sendable: yes` / JSON `sendable: true` |
| 8 | `status-sendable/busy-screen-sendable-no` | Codex working scrollback → `sendable: no` + reason |
| 9 | `status-sendable/unreachable-sendable-no` | Closed port → `sendable: false`, unreachable reason |
| 10 | `lookup/finds-grok-entry` | Resolve finds grok-tty-registry entry |
| 11 | `lookup/finds-codex-entry` | Resolve finds codex-tty-registry entry |
| 12 | `lookup/deterministic-order` | Both registries have same id → registration order wins |
| 13 | `lookup/skips-stale-entry` | Unreachable grok entry removed; codex entry returned |
| 14 | `lookup/session-not-found` | Unknown terminal id → not found error |
| 15 | `stub-tty/run-creates-registry-and-tty-json` | stub-tty run creates registry + tty.json |
| 16 | `stub-tty/scenario-banner-delay` | `banner_delay_ms` delays banner before writable |
| 17 | `stub-tty/scenario-mock-llm-events` | `llm_events` appended to `events.jsonl` |
| 18 | `stub-tty/scenario-mock-screen-frames` | `screen_frames` change scrollback over time |
| 19 | `stub-tty/scenario-declared-screen-status` | `screen_status: idle` detected by tty status |
| 20 | `stub-tty/attach-interactive` | Interactive attach to live stub session |
| 21 | `stub-tty/keep-tty-persists` | `--keep-tty` keeps registry + tty.json alive |
| 22 | `stub-tty/multi-attach/first-attach-writes-second-readonly` | Second attach cannot write |
| 23 | `stub-tty/multi-attach/writer-detach-no-promotion-for-second` | Writer detach; second stays observer |
| 24 | `stub-tty/multi-attach/writer-plus-observer-both-receive-output` | Output fan-out to writer + observer |
| 25 | `stub-tty/multi-attach/third-attach-after-writer-gone-still-readonly` | Third attach after writer gone is observer |
| 26 | `stub-tty/multi-attach/multiple-observers-all-receive` | N observers all receive PTY output |
| 27 | `stub-tty/multi-attach/observer-resize-ignored-pty-unchanged` | Observer resize dropped |
| 28 | `stub-tty/multi-attach/observer-close-does-not-delete-session` | Observer disconnect inert |
| 29 | `stub-tty/multi-attach/snapshot-does-not-affect-writer` | Snapshot probe does not claim write |
| 30 | `stub-tty/multi-attach/send-uses-server-write-not-client-attach` | tty send via server WriteInput |
| 31 | `stub-tty/multi-attach/send-waits-until-writable-then-injects` | send blocks until idle prompt |
| 32 | `stub-tty/multi-attach/send-times-out-with-reason-after-10s` | send timeout includes provider reason |
| 33 | `stub-tty/multi-attach/server-write-while-client-holds-unified-write` | Server send while writer attached |
| 34 | `integration/sealed-trees-regression` | Doc pointer to sealed tty/grok-tty/codex-tty trees |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./cmd/agent-run/tests/ttyrunner                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./cmd/agent-run/tests/ttyrunner
doctest test --label-all ./cmd/agent-run/tests/ttyrunner

doctest vet ./cmd/agent-run/tests/ttyrunner
doctest test ./cmd/agent-run/tests/ttyrunner
doctest test -v ./cmd/agent-run/tests/ttyrunner/provider-registry/lists-builtin-providers
doctest test -v ./cmd/agent-run/tests/ttyrunner/stub-tty/multi-attach/first-attach-writes-second-readonly

# Regression guard (sealed trees — must pass unchanged)
doctest test ./cmd/agent-run/tests/tty/...
doctest test ./cmd/agent-run/tests/grok-tty/...
doctest test ./cmd/agent-run/tests/codex-tty/...
```

```go
import (
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
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/ttyrunner"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	RepoRoot string
	TempDir  string
	Home     string
	AgentRun string
	Env      []string

	Operation string // registry | storage | lookup | status-sendable | stub-tty | multi-attach | integration
	Action    string // leaf-specific action within Operation

	// registry
	EnableStubTTY bool
	RunnerID      string

	// storage / lookup
	RegistryDir         string
	RegistrySessionID   string
	AgentSessionID      string
	Runner              string
	ListenAddr          string
	RegistryEntryJSON   string
	TTYJSON             string
	MetaTerminalSession string

	// status-sendable
	StartFakePTYWrap      bool
	FakePTYWrapPort       int
	FakePTYWrapScrollback string
	StatusArgs            []string

	// stub-tty / multi-attach
	StubScenarioJSON string
	StubScenarioPath string
	StubPrompt       string
	KeepTTY          bool
	StubExitAfterTurn bool

	// multi-attach probes
	AttachModes       []string
	ExpectWriterIndex int
	ExpectReadOnly    bool
	SendMessage       string
	SendTimeout       time.Duration
	ExpectSendTimeout bool
	ExpectSendReason  string

	ExecTimeout       time.Duration
	BackgroundCmd     *exec.Cmd
	BackgroundStderr    *bytes.Buffer
	BackgroundStdout  *bytes.Buffer
	TerminalSessionID string
}

type Response struct {
	ProviderIDs       []string
	IsTTYRunner       bool
	Provider          ttyrunner.Provider
	RegistryDir       string
	TTYSession        *ttyrunner.TTYSession
	LookupEntry       *ttyrunner.RegistryEntry
	LookupRunnerID    string
	WritableStatus    ttyrunner.WritableStatus
	Err               error

	Stdout   string
	Stderr   string
	ExitCode int

	JSONBody map[string]any

	RegistryPath string
	TTYJSONPath  string
	TTYJSONData  map[string]any
	MetaPath     string

	EventsFilePath  string
	EventsFileLines []string

	MultiAttachProbe *MultiAttachProbeResult

	SealedTreesDoc string
}

type MultiAttachProbeResult struct {
	WriterCanWrite      bool
	ObserverCanWrite    bool
	WriterReceived      string
	ObserverReceived    string
	Observer2Received   string
	ResizeAccepted      bool
	PTYColsAfterResize  int
	SessionStillAlive   bool
	SendInjected        bool
	SendTimedOut        bool
	SendTimeoutReason   string
	ServerSendWhileWriter bool
	SnapshotClaimedWrite bool
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	// No parent applyEnv: in-process APIs take req.Home explicitly; CLI subprocesses
	// set cmd.Env via execAgentRun / startStubTTYBackground (req.Env).
	switch req.Operation {
	case "registry":
		return runRegistryOp(t, req)
	case "storage":
		return runStorageOp(t, req)
	case "lookup":
		return runLookupOp(t, req)
	case "status-sendable":
		return runStatusSendableOp(t, req)
	case "stub-tty":
		return runStubTTYOp(t, req)
	case "multi-attach":
		return runMultiAttachOp(t, req)
	case "integration":
		return runIntegrationOp(t, req)
	default:
		return nil, fmt.Errorf("unknown operation %q", req.Operation)
	}
}
```