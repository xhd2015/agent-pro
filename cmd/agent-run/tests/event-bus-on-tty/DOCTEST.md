# agent-run event-bus OnTTYStarted wire (true first live TTY)

Classic TDD **P1** extension: wire library `OnTTYStarted` to HTTP publish via
`--event-bus-url` / `--event-bus-token`, and guard against double-fire when
ForceNew `NotifyOnOpenPath` and the library hook both run for one open.

Sibling of `cmd/agent-run/tests/event-bus/` so that existing help / notify /
append-flags / open-path leaves stay **GREEN** while this tree is **RED** until
implementer lands APIs.

**Out of scope:** ai-critic; real iTerm; new event type strings (use
`agent.tty.started` / `agent-run`).

Library policy alone: `tests/agentrunapi/on-tty-started/`.

## Version

0.0.2

# DSN (Domain Specific Notion)

agent-run wires a neutral library TTY-started hook to best-effort HTTP publish;
ForceNew and library paths share an at-most-once guard.

**Participants**

- **EventBusOpts** — URL, Token, PublishHook (L2 inject), WarnWriter, and
  **`AlreadyNotified *bool`** for at-most-once per open when shared across call sites.
- **NotifyTTYStarted** — existing best-effort publish of `agent.tty.started` /
  `agent-run` with payload `{session_id, runner, workspace}`; empty URL → no HTTP.
  When `AlreadyNotified` non-nil: if `*true` skip; after publishing with non-empty
  URL set `*true` (empty URL does not set).
- **NotifyOnOpenPath** — existing ForceNew (`new-terminal`) vs send policy.
- **WireOnTTYStarted** — returns `func(agentrunapi.TTYStartedInfo)` for
  `Opts.OnTTYStarted`; non-empty URL → NotifyTTYStarted; empty URL → nil/no-op.
  Production `runAutoSendOrResume` installs this when flags set.
- **AutoSendOrResume + OnTTYStarted** — library fires once on first live TTY
  (ModeRun); not on live send (covered in `tests/agentrunapi/on-tty-started`).
- **HTTPCapture / PublishHook** — L2 records type/source/payload without real network.

**Behaviors**

```
# library wire (true TTY, not ForceNew label alone)
WireOnTTYStarted(URL set) + AutoSendOrResume ModeRun
  -> one publish type=agent.tty.started source=agent-run
WireOnTTYStarted(URL empty) + ModeRun
  -> no HTTP; no warning

# double-fire guard
shared AlreadyNotified + URL set
  NotifyOnOpenPath("new-terminal") + WireOnTTYStarted(info)
  -> PublishCount == 1
```

### Planned public API (Classic TDD — locked for implementer)

```go
// pkgs/agentruncli

// EventBusOpts gains:
AlreadyNotified *bool // optional once-guard for NotifyTTYStarted

// WireOnTTYStarted returns OnTTYStarted callback (nil/no-op when URL empty).
func WireOnTTYStarted(opts EventBusOpts) func(agentrunapi.TTYStartedInfo)

// pkgs/agentrunapi — see tests/agentrunapi/on-tty-started
type TTYStartedInfo struct {
	SessionID string
	Runner    string
	Workspace string
}
// Opts.OnTTYStarted func(TTYStartedInfo)
```

## Decision Tree

```
cmd/agent-run/tests/event-bus-on-tty/
├── DOCTEST.md
├── SETUP.md
├── library-wire/                              # true-TTY → HTTP
│   ├── SETUP.md
│   ├── true-tty-publishes-once/               # W1: URL set → 1 publish
│   └── empty-url-no-http/                     # W2: empty URL → 0
└── double-fire-guard/                         # ForceNew + library ≤1
    ├── SETUP.md
    └── force-new-plus-library-at-most-one/    # D1: PublishCount==1
```

Parameter ranking (most → least significant):

1. **Surface** — library-wire (true TTY) | double-fire (ForceNew+library)
2. **URL** — set vs empty (under library-wire)

## Test Index

| # | Leaf | Description | Expect |
|---|------|-------------|--------|
| 1 | `library-wire/true-tty-publishes-once` | WireOnTTYStarted + ModeRun → 1 publish | RED |
| 2 | `library-wire/empty-url-no-http` | empty URL wire + ModeRun → 0 publish | RED |
| 3 | `double-fire-guard/force-new-plus-library-at-most-one` | ForceNew + library shared guard → PublishCount==1 | RED |

## How to Run

```sh
# From agent-pro module root:
doctest vet ./cmd/agent-run/tests/event-bus-on-tty
doctest test ./cmd/agent-run/tests/event-bus-on-tty

# Sibling regression (must stay GREEN):
doctest test ./cmd/agent-run/tests/event-bus

# Library policy:
doctest test ./tests/agentrunapi/on-tty-started
```

Expect **RED** (compile or assert) until WireOnTTYStarted, AlreadyNotified,
TTYStartedInfo, OnTTYStarted, and production wire land.

```go
import (
	"context"
	"encoding/json"
	"io"
	"path/filepath"
	"sync"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/agent-pro/pkgs/agentruncli"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/doctest/session"
)

// Locked wire vocabulary (same strings as go-pkgs/eventbus).
const (
	wireTypeAgentTTYStarted = "agent.tty.started"
	wireSourceAgentRun      = "agent-run"
)

const (
	opLibraryWire = "library-wire"
	opDoubleFire  = "double-fire"
)

const openKindNewTerminal = "new-terminal"

// HTTPCapture records publish attempts via PublishHook inject.
type HTTPCapture struct {
	mu       sync.Mutex
	Requests []CapturedRequest
}

// CapturedRequest is one observed publish.
type CapturedRequest struct {
	Type    string
	Source  string
	Payload json.RawMessage
	Body    []byte
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

// Request drives one P1 wire / double-fire scenario.
type Request struct {
	Op string

	EventBusURL   string
	EventBusToken string

	SessionID string
	Runner    string
	Workspace string

	Capture            *HTTPCapture
	UseInjectPublisher bool
	UseAlreadyNotified bool
}

// Response holds harness observations.
type Response struct {
	WarnOutput   string
	PublishCount int
	APIErrString string
	RunCalls     int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	switch req.Op {
	case opLibraryWire:
		return runLibraryWire(t, req)
	case opDoubleFire:
		return runDoubleFire(t, req)
	default:
		t.Fatalf("unknown Op %q", req.Op)
		return nil, nil
	}
}

func makeOpts(req *Request, warn io.Writer) agentruncli.EventBusOpts {
	opts := agentruncli.EventBusOpts{
		URL:        req.EventBusURL,
		Token:      req.EventBusToken,
		WarnWriter: warn,
	}
	if req.UseAlreadyNotified {
		var notified bool
		opts.AlreadyNotified = &notified
	}
	if req.UseInjectPublisher {
		if req.Capture == nil {
			req.Capture = &HTTPCapture{}
		}
		cap := req.Capture
		opts.PublishHook = func(ctx context.Context, eventType, source string, payload json.RawMessage) error {
			_ = ctx
			body, _ := json.Marshal(map[string]any{
				"type":    eventType,
				"source":  source,
				"payload": json.RawMessage(payload),
			})
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

// stringWriter accumulates WarnWriter output.
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

// runLibraryWire: WireOnTTYStarted + AutoSendOrResume ModeRun (true-TTY path).
func runLibraryWire(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	warnBuf := &stringWriter{}
	busOpts := makeOpts(req, warnBuf)

	home := filepath.Join(t.TempDir(), ".agent-run")
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		return nil, err
	}

	resp := &Response{}
	apiOpts := agentrunapi.Opts{
		SessionID:    req.SessionID,
		Prompt:       "true-tty wire establish",
		WorkspaceDir: req.Workspace,
		AgentRunner:  req.Runner,
		Store:        store,
		NewTerminal:  false,
		Probe:        agentrunapi.EmptyProbe,
		OnTTYStarted: agentruncli.WireOnTTYStarted(busOpts),
		RunSession: func(ctx context.Context, o agentrunapi.Opts, meta agentstorage.SessionMeta, found bool) error {
			_ = ctx
			_ = o
			_ = meta
			_ = found
			resp.RunCalls++
			return nil
		},
	}
	if err := agentrunapi.AutoSendOrResume(context.Background(), apiOpts); err != nil {
		resp.APIErrString = err.Error()
	}
	resp.WarnOutput = warnBuf.String()
	if req.Capture != nil {
		resp.PublishCount = req.Capture.Len()
	}
	return resp, nil
}

// runDoubleFire: ForceNew NotifyOnOpenPath + library WireOnTTYStarted with shared guard.
func runDoubleFire(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	warnBuf := &stringWriter{}
	req.UseAlreadyNotified = true
	opts := makeOpts(req, warnBuf)

	agentruncli.NotifyOnOpenPath(openKindNewTerminal, opts, req.SessionID, req.Runner, req.Workspace)

	hook := agentruncli.WireOnTTYStarted(opts)
	if hook != nil {
		hook(agentrunapi.TTYStartedInfo{
			SessionID: req.SessionID,
			Runner:    req.Runner,
			Workspace: req.Workspace,
		})
	}

	resp := &Response{WarnOutput: warnBuf.String()}
	if req.Capture != nil {
		resp.PublishCount = req.Capture.Len()
	}
	return resp, nil
}
```
