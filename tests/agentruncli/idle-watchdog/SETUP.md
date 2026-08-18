# Scenario

**Feature**: session idle-policy file + fake-clock watchdog Tick

```
WriteIdlePolicy / ReadIdlePolicy -> sessions/<id>/idle-policy.json
SampleIsIdle(sample) -> idle only when sendable + screen idle + empty box + queue 0
NewIdleWatchdog + Tick(Now, Sample) -> SoftExit once, then Shutdown after grace
```

## Preconditions

- Nested root: `d.DOCTEST_ROOT` is `tests/agentruncli/idle-watchdog` (does not
  inherit parent `tests/agentruncli` Setup/Run).
- Package `github.com/xhd2015/agent-pro/pkgs/agentruncli` exports
  `IdlePolicy`, `IdlePolicyPath`, `WriteIdlePolicy`, `ReadIdlePolicy`,
  `IdleSample`, `SampleIsIdle`, `IdleWatchdog`, `NewIdleWatchdog` (RED until
  implementer).
- Home is a per-leaf dir under `d.DOCTEST_CASE` (parallel-safe; no `t.Setenv`
  / `AGENT_RUN_HOME` / `Chdir` / process stdio).
- No real serve / iTerm / grok. Fake time only.

## Steps

1. Root seeds Home, SessionID, default idle sample fields, fake timeout 10s.
2. Grouping sets `Op`.
3. Leaf sets write/raw file, predicate override, or Tick steps.
4. `Run` dispatches; leaf Assert checks file / Idle / hook counts.

## Context

- Path: `filepath.Join(home, "sessions", sessionID, "idle-policy.json")`.
- Default sample: sendable, screen `idle`, box `empty`, queue 0.
- Fake timeout 10s, default grace 5s (product fills Grace==0).
- Compact policy duration: `10m` (not `10m0s`).

```go
import (
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentruncli"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	if req.Home == "" {
		req.Home = filepath.Join(d.DOCTEST_CASE, "home")
	}
	if req.SessionID == "" {
		req.SessionID = "sess-idle-wd"
	}
	if req.Screen == "" {
		req.Screen = "idle"
	}
	if req.InputBox == "" {
		req.InputBox = "empty"
	}
	if req.WatchdogTimeout == 0 {
		req.WatchdogTimeout = defaultFakeTimeout
	}
	// Default sample is sendable; not-sendable leaf overrides after this.
	req.Sendable = true
	return nil
}

func defaultIdleSample() agentruncli.IdleSample {
	return agentruncli.IdleSample{
		Sendable: true,
		Screen:   "idle",
		InputBox: "empty",
		QueueLen: 0,
	}
}

func wantPolicyPath(home, sessionID string) string {
	return filepath.Join(home, "sessions", sessionID, "idle-policy.json")
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
}

func assertNoAPIError(t *testing.T, resp *Response) {
	t.Helper()
	if resp != nil && resp.ErrString != "" {
		t.Fatalf("unexpected API error: %s", resp.ErrString)
	}
}

func idleAt(at time.Duration) TickStep {
	return TickStep{At: at, Sample: defaultIdleSample()}
}

func sampleAt(at time.Duration, s agentruncli.IdleSample) TickStep {
	return TickStep{At: at, Sample: s}
}

func assertHookAt(t *testing.T, name string, got []time.Duration, want time.Duration) {
	t.Helper()
	if len(got) == 0 {
		t.Fatalf("%s: no calls, want first at %s", name, want)
	}
	if got[0] != want {
		t.Fatalf("%s first at %s, want %s (all %v)", name, got[0], want, got)
	}
}
```
