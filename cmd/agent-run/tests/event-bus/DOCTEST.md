# agent-run event bus publish (`agent.tty.started`)

Classic TDD doctests for **P4**: `agent-run` accepts `--event-bus-url` /
optional `--event-bus-token` and publishes **`agent.tty.started`** only when a
**new TTY** is opened successfully (ForceNew / open-profile path). Best-effort:
warn on failure; never fail open.

Prefer **L2 in-process** injectable APIs in `pkgs/agentruncli` (or a small helper
package re-exported from there). No real iTerm. No product implementation in
this tree — leaves stay **RED** until implementer lands APIs + flag wiring.

Shared wire client: `github.com/xhd2015/dot-pkgs/go-pkgs/eventbus` (P1).

**P1 true-TTY library wire + double-fire** lives in sibling tree (keeps this
suite compile-GREEN): `cmd/agent-run/tests/event-bus-on-tty/`. Library hook
policy: `tests/agentrunapi/on-tty-started/`.

## Version

0.0.2

# DSN (Domain Specific Notion)

agent-run publishes a single bus event when a **new** TTY window is opened
successfully; live send never publishes. Publish is best-effort.

**Participants**

- **agent-run CLI (`run`)** — parses optional `--event-bus-url` and
  `--event-bus-token`; documents both on `run -h` / `run --help`
  (L2: `RunHelpText()` returns the same help body Handle prints).
- **EventBusOpts** — URL, Token, optional production `Publisher` (`*eventbus.Publisher`
  or `Publish(ctx, eventbus.Event)`), L2 inject `PublishHook(ctx, type, source, payload)`,
  and `WarnWriter` (`io.Writer`) for failure warnings.
- **NotifyTTYStarted** — builds `type=agent.tty.started`, `source=agent-run`,
  payload `{session_id, runner, workspace}`; best-effort publish; empty URL → no HTTP;
  publish errors → `warning:` line on WarnWriter; never returns error to caller.
- **AppendEventBusFlags** — pure argv helper: when URL non-empty, append
  `--event-bus-url` / value and optional `--event-bus-token` / value; empty URL → no change.
- **NotifyOnOpenPath** — open-path policy dispatch: `kind=new-terminal` (after
  successful ForceNew) → NotifyTTYStarted once; `kind=send` (live) → no-op.
  Production wires the same rule in `openAutoInNewTerminal` vs send/live.
- **ForceNew follow-up** — include event-bus flags via AppendEventBusFlags so the
  child can publish too if needed.
- **Send / live path** — NotifyOnOpenPath("send", …) never publishes (even when URL set).
- **eventbus.Publisher / httptest** — L2 HTTP mock captures POST `/publish` body and
  Authorization.

**Behaviors**

```
# help
run -h / run --help
  -> documents --event-bus-url and --event-bus-token

# NotifyTTYStarted (best-effort)
URL empty -> no HTTP; no warning
URL set + Publisher mock 200
  -> one Event: type agent.tty.started, source agent-run,
     payload session_id / runner / workspace
URL set + token -> Authorization: Bearer <token> (via Publisher or opts)
Publisher non-2xx / transport error
  -> warning: … on WarnWriter; caller sees no error (open still succeeds)

# AppendEventBusFlags
empty URL -> args unchanged
URL set -> append --event-bus-url <url>; if token set also --event-bus-token <token>

# open path (integration L2)
NotifyOnOpenPath("new-terminal", …) + URL set
  -> NotifyTTYStarted once; follow-up argv includes event-bus flags
NotifyOnOpenPath("send", …) + URL set
  -> no publish
```

## Decision Tree

```
cmd/agent-run/tests/event-bus/
├── DOCTEST.md
├── SETUP.md
├── help/                                      # CLI short-path documentation
│   ├── SETUP.md
│   └── run-help-lists-flags/                  # H1: run -h lists both flags
├── notify-tty-started/                        # L2 NotifyTTYStarted
│   ├── SETUP.md
│   ├── empty-url-no-http/                     # N1: empty URL → no HTTP
│   ├── posts-agent-tty-started/               # N2: type/source/payload correct
│   ├── with-bearer-token/                     # N3: Bearer when token set
│   └── publish-failure-warns/                 # N4: non-2xx → warning; no error
├── append-flags/                              # L2 AppendEventBusFlags pure
│   ├── SETUP.md
│   ├── empty-url-unchanged/                   # A1: empty URL → args unchanged
│   └── with-url-and-token/                    # A2: both flags when URL set
└── open-path/                                 # L2 ForceNew vs send wiring
    ├── SETUP.md
    ├── new-terminal-notifies-once/            # O1: open success → one notify + flags in follow-up
    └── send-mode-no-notify/                   # O2: send path → zero publish
```

Parameter ranking (most → least significant):

1. **Surface** — help | NotifyTTYStarted | AppendEventBusFlags | open-path wiring
2. **URL present** — empty (no-op) vs set (publish / append)
3. **HTTP outcome / token** — success | failure warn | Bearer header
4. **Open kind** — new-terminal ForceNew success vs send/live

## Test Index

| # | Leaf | Req | Description |
|---|------|-----|-------------|
| 1 | `help/run-help-lists-flags` | H1 | `run -h` documents `--event-bus-url` and `--event-bus-token` |
| 2 | `notify-tty-started/empty-url-no-http` | N1 | empty URL → no HTTP request |
| 3 | `notify-tty-started/posts-agent-tty-started` | N2 | one POST; type/source/payload locked |
| 4 | `notify-tty-started/with-bearer-token` | N3 | token set → `Authorization: Bearer …` |
| 5 | `notify-tty-started/publish-failure-warns` | N4 | non-2xx → `warning:` on WarnWriter; no error to caller |
| 6 | `append-flags/empty-url-unchanged` | A1 | empty URL → input args unchanged |
| 7 | `append-flags/with-url-and-token` | A2 | appends url + token flags when URL set |
| 8 | `open-path/new-terminal-notifies-once` | O1 | ForceNew success → one notify; follow-up has event-bus flags |
| 9 | `open-path/send-mode-no-notify` | O2 | send/live path → zero publishes |

## How to Run

```sh
# From agent-pro module root (external/agent-pro-master-2026-08-10-2):
doctest vet ./cmd/agent-run/tests/event-bus
doctest test ./cmd/agent-run/tests/event-bus

# Per branch
doctest test -v ./cmd/agent-run/tests/event-bus/help
doctest test -v ./cmd/agent-run/tests/event-bus/notify-tty-started
doctest test -v ./cmd/agent-run/tests/event-bus/append-flags
doctest test -v ./cmd/agent-run/tests/event-bus/open-path

# Shared client must resolve via go.mod replace to brought go-pkgs (eventbus P1):
#   replace github.com/xhd2015/dot-pkgs/go-pkgs => ../dot-pkgs-master-2026-08-10-1/go-pkgs
```

```go
import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentruncli"
	"github.com/xhd2015/doctest/session"
)

// Locked wire vocabulary (same strings as go-pkgs/eventbus constants).
// Doctest testcase modules under cmd/ do not always inherit parent go.mod
// replace for go-pkgs, so asserts use these literals instead of importing eventbus.
const (
	wireTypeAgentTTYStarted = "agent.tty.started"
	wireSourceAgentRun      = "agent-run"
)

// Op selects the L2 surface under test.
//   help | notify | append-flags | open-path
const (
	opHelp        = "help"
	opNotify      = "notify"
	opAppendFlags = "append-flags"
	opOpenPath    = "open-path"
)

// OpenPathKind is the production branch that may or may not publish.
//   new-terminal — ForceNew / open-profile success path (must publish once when URL set)
//   send         — live send path (must never publish)
const (
	openKindNewTerminal = "new-terminal"
	openKindSend        = "send"
)

// HTTPCapture records publish attempts observed via PublishHook inject.
type HTTPCapture struct {
	mu       sync.Mutex
	Requests []CapturedRequest
}

// CapturedRequest is one observed publish (JSON Event envelope body).
type CapturedRequest struct {
	// Event envelope fields decoded from product publish body when available.
	Type    string
	Source  string
	Payload json.RawMessage
	// Full JSON body when product marshals a complete Event.
	Body []byte
}

func (c *HTTPCapture) add(r CapturedRequest) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Requests = append(c.Requests, r)
}

func (c *HTTPCapture) Len() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.Requests)
}

func (c *HTTPCapture) Last() (CapturedRequest, bool) {
	if c == nil {
		return CapturedRequest{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.Requests) == 0 {
		return CapturedRequest{}, false
	}
	return c.Requests[len(c.Requests)-1], true
}

// Request drives one L2 scenario against agent-run event-bus surfaces.
type Request struct {
	Op string

	// help
	Args []string

	// Event bus options (product inputs).
	EventBusURL   string
	EventBusToken string

	// Notify / open-path session identity (payload fields).
	SessionID string
	Runner    string
	Workspace string

	// AppendEventBusFlags inputs.
	BaseArgs []string

	// open-path branch: new-terminal | send
	OpenKind string

	// Test harness: injectable PublishHook via Capture.
	Capture           *HTTPCapture
	UseInjectPublisher bool
	InjectPublishFail  bool

	// When true, open-path new-terminal leaf also builds follow-up argv via AppendEventBusFlags.
	BuildFollowUpArgs bool
}

// Response holds observed outputs for Assert.
type Response struct {
	// help
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error

	// notify / open-path
	WarnOutput string // contents of WarnWriter after NotifyTTYStarted

	// append-flags / open-path follow-up
	ResultArgs []string

	// Publish call count (from Capture).
	PublishCount int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	switch req.Op {
	case opHelp:
		return runHelp(t, req)
	case opNotify:
		return runNotify(t, req)
	case opAppendFlags:
		out := agentruncli.AppendEventBusFlags(append([]string(nil), req.BaseArgs...), req.EventBusURL, req.EventBusToken)
		return &Response{ResultArgs: out}, nil
	case opOpenPath:
		return runOpenPath(t, req)
	default:
		t.Fatalf("unknown Op %q", req.Op)
		return nil, nil
	}
}

func runHelp(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	_ = req
	// Pure L2: same help body as `agent-run run -h` / flags.Help(runHelp).
	text := agentruncli.RunHelpText()
	return &Response{
		Stdout:   text,
		ExitCode: 0,
	}, nil
}

func makeOpts(req *Request, warn io.Writer) agentruncli.EventBusOpts {
	opts := agentruncli.EventBusOpts{
		URL:        req.EventBusURL,
		Token:      req.EventBusToken,
		WarnWriter: warn,
	}
	if req.UseInjectPublisher {
		if req.Capture == nil {
			req.Capture = &HTTPCapture{}
		}
		fail := req.InjectPublishFail
		cap := req.Capture
		// PublishHook is the L2 inject seam (avoids testcase importing eventbus).
		// Production NotifyTTYStarted must honor PublishHook when set; otherwise
		// use eventbus.NewPublisher(URL, WithToken(Token)).
		opts.PublishHook = func(ctx context.Context, eventType, source string, payload json.RawMessage) error {
			_ = ctx
			if fail {
				return fmt.Errorf("eventbus: publish status 500")
			}
			// Record a minimal Event-shaped body for asserts.
			body, _ := json.Marshal(map[string]any{
				"type":    eventType,
				"source":  source,
				"payload": json.RawMessage(payload),
			})
			// payload field above may double-encode; also store raw fields.
			cap.add(CapturedRequest{
				Type:    eventType,
				Source:  source,
				Payload: append(json.RawMessage(nil), payload...),
				Body:    body,
			})
			return nil
		}
	}
	return opts
}

func runNotify(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	warnBuf := &stringWriter{}
	opts := makeOpts(req, warnBuf)

	// NotifyTTYStarted is best-effort: no error return to caller.
	agentruncli.NotifyTTYStarted(opts, req.SessionID, req.Runner, req.Workspace)

	resp := &Response{
		WarnOutput:   warnBuf.String(),
		PublishCount: 0,
	}
	if req.Capture != nil {
		resp.PublishCount = req.Capture.Len()
	}
	return resp, nil
}

// stringWriter is a tiny io.Writer that accumulates text for WarnWriter asserts.
type stringWriter struct {
	b []byte
}

func (w *stringWriter) Write(p []byte) (int, error) {
	w.b = append(w.b, p...)
	return len(p), nil
}

func (w *stringWriter) String() string {
	if w == nil {
		return ""
	}
	return string(w.b)
}

func runOpenPath(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	warnBuf := &stringWriter{}
	opts := makeOpts(req, warnBuf)

	resp := &Response{}

	// Follow-up argv flags are independent of notify policy (built for ForceNew child).
	if req.BuildFollowUpArgs {
		base := req.BaseArgs
		if len(base) == 0 {
			base = []string{"run", "--auto-send-or-resume", "--session-id", req.SessionID}
		}
		resp.ResultArgs = agentruncli.AppendEventBusFlags(append([]string(nil), base...), req.EventBusURL, req.EventBusToken)
	}

	// Policy dispatch: new-terminal → notify once; send → no-op.
	switch req.OpenKind {
	case openKindNewTerminal, openKindSend:
		agentruncli.NotifyOnOpenPath(req.OpenKind, opts, req.SessionID, req.Runner, req.Workspace)
	default:
		t.Fatalf("unknown OpenKind %q", req.OpenKind)
	}

	resp.WarnOutput = warnBuf.String()
	if req.Capture != nil {
		resp.PublishCount = req.Capture.Len()
	}
	return resp, nil
}
```
