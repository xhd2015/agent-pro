# Scenario

**Feature**: `run --open` starts grok session bind in the background and always
waits for the bind worker after detach

```
agent-run run --agent-runner grok-tty --open [prompt]
  + AGENT_RUN_OPEN_ATTACH_INSTANT=1
  + GROK_HOME (+ optional delayed updates.jsonl materialization)
  -> attach returns quickly (does not wait for bind)
  -> after detach: join bind worker
  -> success: stderr grok session + meta.runner_session_id; exit 0
  -> hard fail: "grok session id not resolved"; exit ≠ 0
  -> concurrent status may observe runner binding|bound mid-open
```

## Preconditions

- Instant attach hook required for CI (no interactive TTY).
- Prefer fake TUI via `AGENT_RUN_GROK_TTY_COMMAND`.
- Hard-require path: non-empty open prompt (O1) and/or `GROK_HOME` /
  `AGENT_RUN_GROK_TTY_GROK_SESSION_ID` so discovery is required after wait
  (not soft 750ms skip).
- Delayed materialization uses root `GrokMaterializeDelay` + empty session dir
  until the scheduled write of `summary.json` / `updates.jsonl`.
- O1 uses `NoGrokHomeEnv` + isolated `HOME` so hard-require is proven via prompt alone.
- O3 uses `GrokSessionCwd` to seed under a cwd ≠ agent-run workspace (prompt fallback).

## Steps

1. Leaf configures open flags, discovery fixtures, optional delay, instant attach.
2. `Run` schedules delayed GROK materialization when configured, executes open
   (or open + mid-flight status probe).
3. Assert wall time (B2/O1), stderr/meta (B1/B2/O1/O3), hard error (B3), or mid-open
   runner status (B4).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.OpenInstantAttach = true
	req.Runner = "grok-tty"
	if req.GrokTTYCommand == "" {
		req.GrokTTYCommand = fakeTUIRespondHi()
	}
	return nil
}
```
