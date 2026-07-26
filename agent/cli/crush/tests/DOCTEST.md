# Crush Agent Tests

These doc-style tests verify the `agent/cli/crush` package — event conversion (`UnwrapEvent`, `ToCrush`/`FromCrush` round-trip), server client integration (`CrushServerClient`), and the `CrushAgent.Ask()` method in both subprocess and server modes.

## Version

0.0.2

## DSN (Domain Specific Notion)

Participants:

- **CrushAgent** — an in-process adapter that implements `registry.Agent`. It
  drives `crush` either as a subprocess (default `Mode=""`) or through a
  long-lived `crush server` daemon via `CrushServerClient` (`Mode="server-ask"`).
- **crush binary / crush server** — the CLI and its daemon. The server exposes
  `/v1/health`, `/v1/workspaces`, sessions, and an SSE events stream; runs
  emit a 3-level SSE envelope (`message` / `agent_event` / `run_complete`).
- **UnwrapEvent** — converts one 3-level crush SSE data line into a canonical
  `crush_types.Event`. Unknown outer types and malformed JSON are dropped
  (return nil).
- **ToCrush / FromCrush** — the `crush_types` conversion pair that round-trips
  canonical `AgentEvent`s through crush's native event model and back
  (`Mode="convert-roundtrip"`). Known gaps: `ActionDone` loses its text, and
  empty IDs get a synthetic `"evt_1"` (prefixed `"crush:"` after `FromCrush`).
- **CrushServerClient** — manages the daemon lifecycle: `EnsureServer`
  (auto-start + health probe), `CreateWorkspace`, `CreateSession`,
  `SubscribeEvents`, `SendMessage`, and `KillExistingServer`. Multiple clients
  reuse one OS process.

Behaviors:

- `Ask(ctx, question, opts, onDelta)` in subprocess mode shells out to `crush`;
  in server mode it creates a workspace/session, sends the prompt, and drains
  SSE until `run_complete`.
- Session resume: a second `Ask` reuses `LastSessionID` via
  `AskOptions.SessionID`.
- `FindAgentPath(env)` is `env.LookPath("crush")`, else an error mentioning
  "crush".

## Decision Tree

```
tests/
├── basic-query/                        (Mode="",  subprocess Ask)
│   └── ASSERT: answer contains "paris"
├── session-resume/                     (Mode="",  subprocess Ask + resume)
│   └── ASSERT: answer references "french" or "capital"
│
└── server/                             (grouping: server-mode tests)
    │
    ├── convert/                        (grouping: Mode="convert")
    │   ├── ...                         (UnwrapEvent leaves — unchanged)
    │   │
    │   └── roundtrip/                  (grouping: Mode="convert-roundtrip")
    │       ├── think/                  ActionThink → EventMessage(PartReasoning) → ActionThink
    │       │   └── ASSERT: type, text preserved
    │       ├── message/                ActionMessage → EventMessage(PartText) → ActionMessage
    │       │   └── ASSERT: type, text preserved
    │       ├── tool-call/              ActionToolCall → EventMessage(PartToolCall) → ActionToolCall
    │       │   └── ASSERT: tool, input preserved
    │       ├── error/                  ActionError → EventAgentEvent → ActionError
    │       │   └── ASSERT: type, text preserved
    │       ├── done/                   ActionDone → EventRunComplete → ActionDone
    │       │   └── ASSERT: type preserved, text lost (known limitation)
    │       ├── empty-id/               ActionMessage with empty ID → synthetic ID assigned
    │       │   └── ASSERT: ID is "crush:evt_1"
    │       ├── multi-event/            4 events of different types → round-trip all
    │       │   └── ASSERT: count=4, types/order preserved
    │       └── tool-call-empty-input/  ActionToolCall with nil ToolInput → no panic
    │           └── ASSERT: empty map, no panic
    │
    ├── server-client/                  (grouping: Mode="server-client")
    │   ├── ...                         (existing server-client leaves — unchanged)
    │   │
    │   ├── lifecycle/                  (grouping: server-lifecycle)
    │   │   └── server-lifecycle/       daemon start → health → kill → confirm stopped
    │   │       └── ASSERT: process counts, health status at each stage
    │   └── reuse/                      (grouping: server-reuse)
    │       └── server-reuse/           two clients share one daemon
    │           └── ASSERT: exactly 1 process, both healthy
    │
    └── session-persist/                (Mode="server-ask", resume session)
        └── ASSERT: second answer references first query context
```

## Test Leaves

| Leaf | Mode | Description |
|---|---|---|
| `basic-query` | `""` | Subprocess: capital of France → "paris" |
| `session-resume` | `""` | Subprocess: ask, resume, verify context |
| `server/convert/message-event` | `convert` | `UnwrapEvent`: message SSE → EventMessage with correct fields |
| `server/convert/agent-error` | `convert` | `UnwrapEvent`: agent_event SSE → EventAgentEvent with type=error |
| `server/convert/run-complete` | `convert` | `UnwrapEvent`: run_complete SSE → EventRunComplete |
| `server/convert/drop-unknown` | `convert` | `UnwrapEvent`: unknown type → nil (dropped) |
| `server/convert/malformed-json` | `convert` | `UnwrapEvent`: garbage input → nil (dropped) |
| `server/convert/roundtrip/think` | `convert-roundtrip` | `ToCrush`/`FromCrush`: ActionThink round-trips with text preserved |
| `server/convert/roundtrip/message` | `convert-roundtrip` | `ToCrush`/`FromCrush`: ActionMessage round-trips with text preserved |
| `server/convert/roundtrip/tool-call` | `convert-roundtrip` | `ToCrush`/`FromCrush`: ActionToolCall round-trips with tool/input preserved |
| `server/convert/roundtrip/error` | `convert-roundtrip` | `ToCrush`/`FromCrush`: ActionError round-trips with text preserved |
| `server/convert/roundtrip/done` | `convert-roundtrip` | `ToCrush`/`FromCrush`: ActionDone type preserved, text lost |
| `server/convert/roundtrip/empty-id` | `convert-roundtrip` | `ToCrush`/`FromCrush`: empty ID gets synthetic "evt_1" |
| `server/convert/roundtrip/multi-event` | `convert-roundtrip` | `ToCrush`/`FromCrush`: 4 mixed events, count and order preserved |
| `server/convert/roundtrip/tool-call-empty-input` | `convert-roundtrip` | `ToCrush`/`FromCrush`: nil ToolInput handled without panic |
| `server/server-client/health-check` | `server-client` | `ensureServer`: probe /v1/health returns 200 |
| `server/server-client/auto-start` | `server-client` | `ensureServer`: auto-start when server not running |
| `server/server-client/create-workspace` | `server-client` | `createWorkspace`: POST returns valid id |
| `server/server-client/send-and-receive` | `server-client` | Full cycle: subscribe, send, read SSE events |
| `server/server-client/lifecycle/server-lifecycle` | `server-client` | Daemon lifecycle: start → health → kill → confirm stopped |
| `server/server-client/reuse/server-reuse` | `server-client` | Two clients share one daemon, exactly 1 OS process |
| `server/session-persist` | `server-ask` | Server-mode: Ask with session resume |

## How to Run

```sh
doctest test ./agent/cli/crush/tests/...
```

Integration tests (server-client leaves, session-persist) require the `crush` binary in PATH.

```sh
doctest test ./agent/cli/crush/tests/...
```

```go
import (

	"encoding/json"
	"fmt"
	osexec "os/exec"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/cli/crush"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	crush_types "github.com/xhd2015/agent-pro/agent/event/crush_types"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/agent/exec"
	"github.com/xhd2015/doctest/session"
)


type Request struct {
	Prompt       string
	ResumePrompt string
	Model        string
	Env          []string

	Mode            string // "", "convert", "server-client", "server-ask", "convert-roundtrip"
	SSEInput        string // raw SSE data line for convert mode
	ServerOperation string // "health-check", "auto-start", "create-workspace", "send-and-receive", "server-lifecycle", "server-reuse"
	AgentEventsJSON string // JSON array of types.AgentEvent for convert-roundtrip mode
	SessionID       string // session ID for convert-roundtrip mode
}

type Response struct {
	Answer    string // for Ask results
	Output    string // for convert / server-client results (JSON)
	SessionID string // session ID from agent after Ask
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	switch req.Mode {
	case "convert":
		return runConvert(t, req)
	case "server-client":
		return runServerClient(t, req)
	case "server-ask":
		return runServerAsk(t, req)
	case "convert-roundtrip":
		return runConvertRoundtrip(t, req)
	default:
		return runSubprocess(t, req)
	}
}
```
