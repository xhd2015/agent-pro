# Scenario

**Feature**: `--open` with argv prompt succeeds without banner/OpenReady markers

```
agent-run run --agent-runner grok-tty --open "open-no-banner"
  + fake TUI: "booting" only, sleep 12, exit
  + no AGENT_RUN_OPEN_ATTACH_INSTANT
  -> exit 0
  -> no "banner not detected"
  -> stderr once "grok-tty: <id>"
```

## Preconditions

- New session: prompt is on runner argv; open must not hard-wait inject-ready
  solely to attach.
- No INSTANT env so the leaf stays RED until production open path is attach-first.

## Steps

1. Set prompt positional + no-banner hold fake TUI.
2. Assert exit 0, session id, no banner error.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	clearOpenInstantAttach(req)
	req.Prompt = "open-no-banner"
	hold := writeFakeTUINoBannerHold(t, req.TempDir, 12)
	setGrokTTYCommand(req, hold)
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--open", req.Prompt}
	return nil
}
```
