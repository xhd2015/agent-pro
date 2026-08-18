# agentruncli — idle-policy.json + idle watchdog (`Tick`)

Classic TDD doctests for **P2**: keep-alive `__serve__` exits after a continuous
idle window when launch-time policy is on. Policy travels in a **session-dir
file** (not serve argv / `CommandEnv`). `Tick` is the unit under test; tests
drive a fake clock and injectable `Sample` / `SoftExit` / `Shutdown`.

Isolated nested root so parent `tests/agentruncli/` Handle-extract leaves stay
**GREEN** (do not add these APIs to the parent Run). This root is RED /
compile-RED until implementer lands policy I/O + predicate + watchdog.

P1 (`NormalizeIdle` / `DefaultIdleTimeout`, CLI flags, FollowUp emit) is sealed
— reuse the default `10m` compact string; do **not** retest flag parse.

**Out of scope (P2):** local-bot ExecLauncher (P3), CLI help / FollowUp emit,
real iTerm / grok / HeadlessRun / e2e, persist on `SessionMeta`,
`agent-run resume` flags, live TTY / `OnListening` wiring.

# DSN (Domain Specific Notion)

Serve cannot take idle flags on argv, so the session dir holds the policy.
A watchdog samples idle-ness on a fake clock and soft-exits once, then shuts
down after grace.

**Participants**

- **`idle-policy.json`** — `$HOME/sessions/<session-id>/idle-policy.json`
  (not `meta.json`). Wire: `{"exit_on_idle":true,"idle_timeout":"10m"}` with
  compact `time.ParseDuration` strings (`10m`, `2s`).
- **`IdlePolicyPath` / `WriteIdlePolicy` / `ReadIdlePolicy`** — path +
  round-trip. Missing file → found=false, no error.
- **`IdleSample` / `SampleIsIdle`** — one sample is idle only when sendable,
  screen is `idle` (not `starting`|`busy`|`modal`|`unknown`), queue length 0,
  and input box `empty` (not `occupied`|`unknown`).
- **`NewIdleWatchdog` + `IdleWatchdog.Tick`** — armed only when the policy
  file is found and `exit_on_idle` is true. `Now` / `Sample` / `SoftExit` /
  `Shutdown` are injectable. Grace defaults to **5s**.
- **Soft `/exit` then shutdown** — first continuous idle of `idle_timeout`
  calls `SoftExit` once; after grace, if still reachable, `Shutdown` once.
  Clock does not start until the first idle sample.

**Behaviors**

- Write then read returns the same `ExitOnIdle` and parsed duration; file
  sits at `sessions/<id>/idle-policy.json` with compact `idle_timeout`.
- Missing file → found=false, no error; watchdog never starts (Tick never exits).
- `exit_on_idle=false` → watchdog never starts / Tick never SoftExit/Shutdown.
- `SampleIsIdle` true only when all four idle conditions hold; any fail → false.
- Continuous idle from t=0 for timeout → SoftExit once; +grace → Shutdown once.
- Busy every tick past timeout → no SoftExit.
- Idle almost-timeout, then occupied, then idle almost-timeout → no SoftExit
  (clock resets).
- First idle late → no exit at wall-clock timeout; exit at first-idle+timeout.
- SoftExit is not called twice if Tick continues after timeout.

## Version

0.0.2

## Decision Tree

```
tests/agentruncli/idle-watchdog/
├── DOCTEST.md
├── SETUP.md
├── policy/                         # file I/O + policy-gated start
│   ├── write-read-roundtrip/       # true + 10m
│   ├── missing-file/               # found=false; Tick never exits
│   ├── disabled-false/             # exit_on_idle=false; Tick never exits
│   └── invalid/
│       ├── json/                   # raw `{` → Read error
│       └── timeout/                # idle_timeout=nope → Read error
├── predicate/                      # SampleIsIdle
│   ├── idle/                       # all four hold
│   └── not-idle/
│       ├── occupied-box/
│       ├── unknown-box/
│       ├── queue-nonzero/
│       ├── not-sendable/
│       └── screen/
│           ├── busy/
│           ├── unknown/
│           ├── starting/
│           └── modal/
└── watchdog/                       # fake-clock Tick (armed policy)
    ├── holds-for-timeout/          # SoftExit + Shutdown
    ├── never-idle-busy/            # no SoftExit
    ├── reset-mid-window/           # interrupt resets clock
    ├── late-first-idle/            # clock starts at first idle
    └── soft-exit-once/             # no second /exit
```

Parameter ranking (most → least significant):

1. **Surface** — policy file | idle predicate | watchdog Tick
2. **Policy presence / validity** — written-true | missing | written-false | invalid
3. **Predicate outcome** — all-hold vs which factor fails
4. **Watchdog clock story** — holds | never idle | reset | late start | once

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `policy/write-read-roundtrip` | write `{true,10m}` → read same; compact JSON at session path |
| 2 | `policy/missing-file` | Read found=false, no error; Tick idle past timeout → 0 exits |
| 3 | `policy/disabled-false` | write `exit_on_idle=false` → Tick never SoftExit/Shutdown |
| 4 | `policy/invalid/json` | raw `{` → Read error |
| 5 | `policy/invalid/timeout` | `"idle_timeout":"nope"` → Read error |
| 6 | `predicate/idle` | sendable + screen idle + empty box + queue 0 → idle |
| 7 | `predicate/not-idle/occupied-box` | occupied box → not idle (even if sendable) |
| 8 | `predicate/not-idle/unknown-box` | unknown box → not idle |
| 9 | `predicate/not-idle/queue-nonzero` | queue 1 → not idle |
| 10 | `predicate/not-idle/not-sendable` | Sendable=false → not idle |
| 11 | `predicate/not-idle/screen/busy` | screen busy → not idle |
| 12 | `predicate/not-idle/screen/unknown` | screen unknown → not idle |
| 13 | `predicate/not-idle/screen/starting` | screen starting → not idle |
| 14 | `predicate/not-idle/screen/modal` | screen modal → not idle |
| 15 | `watchdog/holds-for-timeout` | idle t=0..timeout → SoftExit 1; +grace → Shutdown 1 |
| 16 | `watchdog/never-idle-busy` | busy every tick > timeout → SoftExit 0 |
| 17 | `watchdog/reset-mid-window` | idle 9s, occupied, idle 9s (timeout 10s) → SoftExit 0 |
| 18 | `watchdog/late-first-idle` | first idle at 8s; no exit at 10s; SoftExit at 18s |
| 19 | `watchdog/soft-exit-once` | Tick after timeout does not SoftExit twice |

## How to Run

```sh
# From the agent-pro module root:
doctest vet ./tests/agentruncli/idle-watchdog
doctest test ./tests/agentruncli/idle-watchdog

doctest test -v ./tests/agentruncli/idle-watchdog/policy
doctest test -v ./tests/agentruncli/idle-watchdog/predicate
doctest test -v ./tests/agentruncli/idle-watchdog/watchdog
```

Expect **RED** (compile or assert) until policy I/O, `SampleIsIdle`, and
`IdleWatchdog.Tick` land. No `label: e2e`. Fake time only — never sleep 10m.

Use `GOWORK=off` when a parent `go.work` would hide the agent-pro module.

### Planned API addition

```go
// package agentruncli  (or a small helper next to serve, re-exported here)

type IdlePolicy struct {
    ExitOnIdle  bool          `json:"exit_on_idle"`
    IdleTimeout time.Duration `json:"-"` // parsed; file field is compact string
}

func IdlePolicyPath(home, sessionID string) string
func WriteIdlePolicy(home, sessionID string, p IdlePolicy) error
func ReadIdlePolicy(home, sessionID string) (p IdlePolicy, found bool, err error)

type IdleSample struct {
    Sendable bool
    Screen   string // idle|busy|starting|modal|unknown
    InputBox string // empty|occupied|unknown
    QueueLen int
}

func SampleIsIdle(s IdleSample) bool

type IdleWatchdog struct {
    Timeout  time.Duration
    Grace    time.Duration // 0 → 5s
    Now      func() time.Time
    Sample   func() IdleSample
    SoftExit func()
    Shutdown func()
}

// NewIdleWatchdog copies cfg. Tick is a no-op when !found or !p.ExitOnIdle.
// Timeout comes from p.IdleTimeout when cfg.Timeout == 0.
func NewIdleWatchdog(found bool, p IdlePolicy, cfg IdleWatchdog) *IdleWatchdog
func (w *IdleWatchdog) Tick()
```

Default grace: **5s**. Test timeouts use **10s of fake time** (not wall 10m).
Reuse `agentrunapi.DefaultIdleTimeout` (`10m`) only as the policy-file default
string. SoftExit models `trySoftExit` / `InjectMessage(..., "/exit", true)`
once; Shutdown models the post-grace serve kill if still reachable.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentruncli"
	"github.com/xhd2015/doctest/session"
)

const (
	opPolicy    = "policy"
	opPredicate = "predicate"
	opWatchdog  = "watchdog"
)

const defaultFakeTimeout = 10 * time.Second
const defaultFakeGrace = 5 * time.Second

// TickStep is one fake-clock observation: advance Now to t0+At, then Tick.
type TickStep struct {
	At     time.Duration
	Sample agentruncli.IdleSample
}

// Request drives policy I/O, SampleIsIdle, or IdleWatchdog.Tick.
type Request struct {
	Op string

	Home      string
	SessionID string

	WritePolicy bool
	RawFile     []byte

	ExitOnIdle  bool
	IdleTimeout time.Duration

	Sendable bool
	Screen   string
	InputBox string
	QueueLen int

	WatchdogTimeout time.Duration
	WatchdogGrace   time.Duration
	TickAfterPolicy bool
	Steps           []TickStep
}

// Response is the harness observation.
type Response struct {
	Path         string
	Found        bool
	PolicyOn     bool
	PolicyTimeout time.Duration
	FileBody     string
	FileExists   bool
	ErrString    string

	Idle       bool
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
	case opPredicate:
		idle := agentruncli.SampleIsIdle(agentruncli.IdleSample{
			Sendable: req.Sendable,
			Screen:   req.Screen,
			InputBox: req.InputBox,
			QueueLen: req.QueueLen,
		})
		return &Response{Idle: idle}, nil
	case opWatchdog:
		return runWatchdog(t, req, true, agentruncli.IdlePolicy{
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
		wd, err := runWatchdog(t, req, found, p)
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

func runWatchdog(t *testing.T, req *Request, found bool, p agentruncli.IdlePolicy) (*Response, error) {
	t.Helper()
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := t0
	sample := defaultIdleSample()
	var soft, shut int
	var softAt, shutAt []time.Duration
	cfg := agentruncli.IdleWatchdog{
		Timeout: req.WatchdogTimeout,
		Grace:   req.WatchdogGrace,
		Now:     func() time.Time { return now },
		Sample:  func() agentruncli.IdleSample { return sample },
		SoftExit: func() {
			soft++
			softAt = append(softAt, now.Sub(t0))
		},
		Shutdown: func() {
			shut++
			shutAt = append(shutAt, now.Sub(t0))
		},
	}
	w := agentruncli.NewIdleWatchdog(found, p, cfg)
	for _, step := range req.Steps {
		now = t0.Add(step.At)
		sample = step.Sample
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
