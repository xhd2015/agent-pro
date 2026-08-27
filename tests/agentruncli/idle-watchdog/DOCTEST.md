# agentruncli — idle-policy.json + tty/detection idle watchdog

Classic TDD doctests for keep-alive `__serve__` exit-on-idle. Policy travels in
a **session-dir file**. Detection is runner-agnostic via
`pkgs/tty/detection/{changed,occupied,idle}`: resting snapshot unchanged +
space-probe empty for 3 scheduled checks.

**Out of scope:** local-bot ExecLauncher, CLI help / FollowUp emit, real iTerm /
grok / HeadlessRun / e2e, persist on `SessionMeta`, live TTY wiring (covered by
`run-exit-on-idle-*-tty` e2e).

# DSN (Domain Specific Notion)

Serve cannot take idle flags on argv, so the session dir holds the policy.
A watchdog samples resting snapshots + occupy probes on a fake clock and
soft-exits once, then shuts down after grace.

**Participants**

- **`idle-policy.json`** — `$HOME/sessions/<session-id>/idle-policy.json`
- **`IdlePolicyPath` / `WriteIdlePolicy` / `ReadIdlePolicy`** — path + round-trip
- **`pkgs/tty/detection/idle.Watchdog`** — armed when policy found + `exit_on_idle`
- **`changed`** — this resting snap vs last; first sample baselines
- **`occupied`** — space probe; exactly +1 space ⇒ occupied (hold)
- Soft `/exit` then shutdown — three consecutive idle checks → SoftExit;
  after grace → Shutdown. Non-idle (changed / occupied / unknown / queue) resets.

**Behaviors**

- Write then read returns the same `ExitOnIdle` and parsed duration.
- Missing file / `exit_on_idle=false` → Tick never SoftExit/Shutdown.
- Three stable+empty checks → SoftExit; +grace → Shutdown.
- Snapshot change or occupied mid-window → reset; no SoftExit.
- SoftExit is not called twice.

## Version

0.1.0

## Decision Tree

```
tests/agentruncli/idle-watchdog/
├── DOCTEST.md
├── SETUP.md
├── policy/                         # file I/O + policy-gated start
│   ├── write-read-roundtrip/
│   ├── missing-file/
│   ├── disabled-false/
│   └── invalid/
│       ├── json/
│       └── timeout/
└── watchdog/                       # fake-clock Tick (armed policy)
    ├── holds-for-timeout/
    ├── reset-mid-window/           # occupied resets
    ├── changed-resets/             # snapshot change resets
    ├── late-first-idle/
    └── soft-exit-once/
```

Parameter ranking (most → least significant):

1. **Surface** — policy file | watchdog Tick
2. **Policy presence / validity** — written-true | missing | written-false | invalid
3. **Watchdog clock story** — holds | reset (occupy/change) | late start | once

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `policy/write-read-roundtrip` | write `{true,10m}` → read same; compact JSON |
| 2 | `policy/missing-file` | found=false; Tick idle → 0 exits |
| 3 | `policy/disabled-false` | exit_on_idle=false → Tick never exits |
| 4 | `policy/invalid/json` | raw `{` → Read error |
| 5 | `policy/invalid/timeout` | `"idle_timeout":"nope"` → Read error |
| 6 | `watchdog/holds-for-timeout` | 3 stable+empty → SoftExit; +grace → Shutdown |
| 7 | `watchdog/reset-mid-window` | occupy mid-window → SoftExit 0 |
| 8 | `watchdog/changed-resets` | snapshot change mid-window → SoftExit 0 |
| 9 | `watchdog/late-first-idle` | changing then 3 stable → SoftExit on third idle |
| 10 | `watchdog/soft-exit-once` | Tick after SoftExit does not SoftExit twice |

## How to Run

```sh
# From the agent-pro module root:
doctest vet ./tests/agentruncli/idle-watchdog
doctest test ./tests/agentruncli/idle-watchdog

doctest test -v ./tests/agentruncli/idle-watchdog/policy
doctest test -v ./tests/agentruncli/idle-watchdog/watchdog
```

Use `GOWORK=off` when a parent `go.work` would hide the agent-pro module.

### Planned API (detection packages)

```go
// pkgs/tty/detection/occupied
func ExactlyOneMoreSpace(before, after []byte) bool
func Probe(io IO) Status // empty|occupied|unknown

// pkgs/tty/detection/changed
func Changed(before, after []byte) bool
type Tracker struct{ ... }
func (t *Tracker) Note(now string) (changed bool)

// pkgs/tty/detection/idle
type Watchdog struct { Snapshot, ProbeOccupied, QueueLen, SoftExit, Shutdown, ... }
func New(found bool, p Policy, cfg Watchdog) *Watchdog
func (w *Watchdog) Tick()
func Schedule(timeout time.Duration) (first, gap time.Duration)
```

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentruncli"
	"github.com/xhd2015/agent-pro/pkgs/tty/detection/idle"
	"github.com/xhd2015/agent-pro/pkgs/tty/detection/occupied"
	"github.com/xhd2015/doctest/session"
)

const (
	opPolicy   = "policy"
	opWatchdog = "watchdog"
)

const defaultFakeTimeout = 10 * time.Second
const defaultFakeGrace = 5 * time.Second
const defaultSnap = "stable chrome"

// TickStep is one fake-clock observation: advance Now to t0+At, then Tick
// with the given resting snapshot + occupy status.
type TickStep struct {
	At       time.Duration
	Snapshot string
	Occupy   occupied.Status
	QueueLen int
}

// Request drives policy I/O or idle.Watchdog.Tick.
type Request struct {
	Op string

	Home      string
	SessionID string

	WritePolicy bool
	RawFile     []byte

	ExitOnIdle  bool
	IdleTimeout time.Duration

	WatchdogTimeout time.Duration
	WatchdogGrace   time.Duration
	TickAfterPolicy bool
	Steps           []TickStep
}

// Response is the harness observation.
type Response struct {
	Path          string
	Found         bool
	PolicyOn      bool
	PolicyTimeout time.Duration
	FileBody      string
	FileExists    bool
	ErrString     string

	SoftExitN  int
	ShutdownN  int
	SoftExitAt []time.Duration
	ShutdownAt []time.Duration
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	switch req.Op {
	case opPolicy:
		return runPolicy(t, req)
	case opWatchdog:
		return runWatchdog(t, req, true, idle.Policy{
			ExitOnIdle:  true,
			IdleTimeout: req.WatchdogTimeout,
		})
	default:
		t.Fatalf("unknown Op %q", req.Op)
		return nil, nil
	}
}

func runPolicy(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	path := agentruncli.IdlePolicyPath(req.Home, req.SessionID)
	resp := &Response{Path: path}

	if len(req.RawFile) > 0 {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, req.RawFile, 0o644); err != nil {
			return nil, err
		}
	} else if req.WritePolicy {
		err := agentruncli.WriteIdlePolicy(req.Home, req.SessionID, agentruncli.IdlePolicy{
			ExitOnIdle:  req.ExitOnIdle,
			IdleTimeout: req.IdleTimeout,
		})
		if err != nil {
			resp.ErrString = err.Error()
			return resp, nil
		}
	}

	p, found, err := agentruncli.ReadIdlePolicy(req.Home, req.SessionID)
	resp.Found = found
	resp.PolicyOn = p.ExitOnIdle
	resp.PolicyTimeout = p.IdleTimeout
	if err != nil {
		resp.ErrString = err.Error()
		return resp, nil
	}
	if b, err := os.ReadFile(path); err == nil {
		resp.FileBody = string(b)
		resp.FileExists = true
	}

	if req.TickAfterPolicy {
		wd, err := runWatchdog(t, req, found, idle.Policy{
			ExitOnIdle:  p.ExitOnIdle,
			IdleTimeout: p.IdleTimeout,
		})
		if err != nil {
			return resp, err
		}
		resp.SoftExitN = wd.SoftExitN
		resp.ShutdownN = wd.ShutdownN
		resp.SoftExitAt = wd.SoftExitAt
		resp.ShutdownAt = wd.ShutdownAt
	}
	return resp, nil
}

func runWatchdog(t *testing.T, req *Request, found bool, p idle.Policy) (*Response, error) {
	t.Helper()
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := t0
	snap := defaultSnap
	occupy := occupied.Empty
	queue := 0
	var soft, shut int
	var softAt, shutAt []time.Duration
	cfg := idle.Watchdog{
		Timeout: req.WatchdogTimeout,
		Grace:   req.WatchdogGrace,
		Now:     func() time.Time { return now },
		Snapshot: func() (string, error) {
			return snap, nil
		},
		ProbeOccupied: func() occupied.Status { return occupy },
		QueueLen:      func() int { return queue },
		SoftExit: func() {
			soft++
			softAt = append(softAt, now.Sub(t0))
		},
		Shutdown: func() {
			shut++
			shutAt = append(shutAt, now.Sub(t0))
		},
	}
	w := idle.New(found, p, cfg)
	for _, step := range req.Steps {
		now = t0.Add(step.At)
		if step.Snapshot != "" {
			snap = step.Snapshot
		} else {
			snap = defaultSnap
		}
		if step.Occupy != "" {
			occupy = step.Occupy
		} else {
			occupy = occupied.Empty
		}
		queue = step.QueueLen
		w.Tick()
	}
	return &Response{
		Found:      found,
		PolicyOn:   p.ExitOnIdle,
		SoftExitN:  soft,
		ShutdownN:  shut,
		SoftExitAt: softAt,
		ShutdownAt: shutAt,
	}, nil
}
```
