# Scenario

**Feature**: non-open run still waits for delayed banner then injects (compat)

```
fake TUI sleep 0.3 → GROK_TTY_BANNER → read → probe STDIN=<prompt>
agent-run run --agent-runner grok-tty "hi"
  -> probe records STDIN=hi (inject after banner)
  -> no "banner not detected"
```

## Preconditions

- If inject ran before banner, `read` would miss input and probe would lack `STDIN=hi`.
- Probe proves inject-ready wait without depending on grok session discovery
  or capture streaming (those are covered under `cmd/agent-run/tests/grok-tty`).
- Sibling external suite: `cmd/agent-run/tests/grok-tty/run/waits-for-banner`.

## Steps

1. Write delayed-banner probe fake TUI under temp dir.
2. Run with prompt `hi` (no `--open`).
3. Assert probe has injected stdin; no banner timeout error.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	clearOpenInstantAttach(req)
	req.Prompt = "hi"
	probePath := filepath.Join(req.TempDir, "non-open-inject-probe.txt")
	setEnvKV(req, "DOCTEST_TUI_PROBE_PATH", probePath)
	script := writeFakeTUIDelayedBannerProbe(t, req.TempDir, probePath)
	setGrokTTYCommand(req, script)
	req.Args = []string{"run", "--agent-runner", "grok-tty", req.Prompt}
	// Banner delay is short; allow headless discovery/turn path to finish or fail
	// after inject — probe is the hard inject proof.
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
