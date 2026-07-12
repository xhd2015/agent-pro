# Scenario

**Feature**: `--open` attaches without hard-failing on missing banner/OpenReady;
non-open headless still waits for inject-ready banner

```
# open attach-first (production path; no AGENT_RUN_OPEN_ATTACH_INSTANT)
agent-run run --agent-runner grok-tty --open ["prompt"]
  + fake TUI never paints GROK_TTY_BANNER / Grok › / modern OpenReady chrome
  -> exit 0; no "banner not detected"; session id after attach

# non-open compat: hard-wait inject-ready still works
agent-run run --agent-runner grok-tty "hi"
  + delayed GROK_TTY_BANNER
  -> inject after banner; exit 0
```

## Preconditions

- Child `open/` leaves must **not** set `AGENT_RUN_OPEN_ATTACH_INSTANT`.
- Fake no-marker TUI holds then exits so `AttachWriter` can return without a
  controlling TTY (ExitOnTerminalExit), while still exercising production
  readiness (not the INSTANT soft-banner branch).
- Existing `tty-lifecycle/*` INSTANT leaves remain the CI-fast lifecycle suite.

## Steps

1. Grouping documents banner-wait policy class.
2. `open/` vs `non-open/` split on whether `--open` is set.
3. Leaves choose empty/with-prompt/probe or delayed-banner fixtures.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Ensure base run args; children set open vs non-open and runner.
	if len(req.Args) == 0 || req.Args[0] != "run" {
		req.Args = []string{"run"}
	}
	req.Runner = "grok-tty"
	// Default: production attach-first path (no instant attach env).
	clearOpenInstantAttach(req)
	req.ExecTimeout = 55 * time.Second
	return nil
}
```
