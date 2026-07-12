# Scenario

**Feature**: B2 — detach before bind completes ⇒ always wait for bind (critical)

```
empty GROK_HOME at t=0 + AGENT_RUN_GROK_TTY_GROK_SESSION_ID=<uuid>
  + AGENT_RUN_OPEN_ATTACH_INSTANT=1 (attach returns immediately)
  + schedule updates.jsonl materialization after ~2s
  -> agent-run run --open "bg bind wait"
  -> attach returns before discovery is possible
  -> open process MUST NOT exit unbound while discovery still pending
  -> waits until delayed session appears
  -> stderr grok session + meta.runner_session_id; exit 0
  -> wall time ≥ materialization delay
```

Edge case rule: **detach before bind completes → always wait for bind**.
Never soft-skip with a short best-effort window after detach when a real open
bind was started (hard path via GROK_HOME + session-id hook).

## Steps

1. Point GROK_HOME at empty dir; set session UUID hook; do not pre-seed files.
2. Set `GrokMaterializeDelay` to 2s so discovery only becomes available after
   instant attach would have returned.
3. Run open; assert wait ≥ delay, session printed, meta persisted.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const bgBindDelayedUUID = "550e8400-e29b-41d4-a716-446655440802"

// delay longer than instant-attach return, short enough for CI.
const bgBindMaterializeDelay = 2 * time.Second

func Setup(t *testing.T, req *Request) error {
	prompt := "bg bind wait"
	req.OpenPrompt = prompt
	req.InitialPrompt = prompt
	req.GrokHome = filepath.Join(req.TempDir, "grok-home-delayed")
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return err
	}
	req.GrokSessionUUID = bgBindDelayedUUID
	// Do not pre-seed; materialize after attach would return.
	req.GrokMaterializeDelay = bgBindMaterializeDelay
	req.GrokTTYCommand = fakeTUIHoldSeconds(8)
	req.OpenInstantAttach = true
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--open", prompt}
	// Allow long discovery budget (~20s+) after implement.
	req.ExecTimeout = 90 * time.Second
	return nil
}
```
