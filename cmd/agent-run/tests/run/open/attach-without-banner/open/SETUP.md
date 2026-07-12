# Scenario

**Feature**: `--open` production path — attach-first; no hard banner/OpenReady wait

```
agent-run run --agent-runner grok-tty --open ["prompt"]
  + fake TUI: prints only "booting" (not banner/OpenReady), holds, exits
  + no AGENT_RUN_OPEN_ATTACH_INSTANT
  -> exit 0
  -> stderr must not contain "banner not detected"
  -> post-attach "grok-tty: <id>" on stderr
```

## Preconditions

- `OpenInstantAttach` cleared; env `AGENT_RUN_OPEN_ATTACH_INSTANT` absent.
- Default fake TUI is no-marker hold (leaves may override with probe script).
- Hold duration outlives soft optional wait but is short enough for CI (~12s).

## Steps

1. Grouping installs grok-tty `--open` args, clears INSTANT, installs no-banner hold.
2. Leaves set empty vs with-prompt vs argv/stdin probe.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	clearOpenInstantAttach(req)
	req.Runner = "grok-tty"
	req.Mode = "open-registry-after"
	hold := writeFakeTUINoBannerHold(t, req.TempDir, 12)
	setGrokTTYCommand(req, hold)
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--open"}
	req.ExecTimeout = 55 * time.Second
	return nil
}
```
