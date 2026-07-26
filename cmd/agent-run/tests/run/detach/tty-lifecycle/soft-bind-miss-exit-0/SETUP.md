# Scenario

**Feature**: soft grok bind miss still exits 0 under `--detach` (not hard-fail)

```
# isolate default grok home via HOME; do NOT set GROK_HOME
HOME=<temp>/fake-home  (empty sessions)
  + agent-run run --agent-runner grok-tty --detach "soft miss"
  -> exit 0; both ids printed
  -> must NOT hard-fail with "grok session id not resolved"
  -> runner_session_id may be empty (soft unbound)
```

## Preconditions

- Harness `NoGrokHomeEnv` strips ambient `GROK_HOME` / config-home session hooks.
- Fake TUI hold for keep-alive.
- Soft budget may be up to 1 minute in product; this leaf only asserts soft miss
  still succeeds (no hard require). Prefer fast fail-when-no-home in impl.

## Steps

1. Point `HOME` at temp fake-home; strip grok env.
2. Run detach with non-empty prompt; assert exit 0 + both ids.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	prompt := "soft miss detach"
	req.Prompt = prompt

	fakeHome := filepath.Join(req.TempDir, "fake-home")
	if err := os.MkdirAll(filepath.Join(fakeHome, ".grok"), 0755); err != nil {
		return err
	}
	req.NoGrokHomeEnv = true
	setEnvKV(req, "HOME", fakeHome)

	setGrokTTYCommand(req, fakeTUIHoldSeconds(30))
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--detach", prompt}
	req.Mode = "read-meta+registry"
	// Allow product soft budget (~1m) plus overhead if discovery waits full window.
	req.ExecTimeout = 90 * time.Second
	return nil
}
```
