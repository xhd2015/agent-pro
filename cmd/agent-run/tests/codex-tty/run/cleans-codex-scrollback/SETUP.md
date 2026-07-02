# Scenario

**Bug**: Codex scrollback fallback emits useful result text without raw TUI/control transcript

```
fake Codex TUI scrollback
  -> terminal controls, box chrome, startup status, prompt history
  -> useful final output: ls output, AGENTS.md, cmd, pkgs

capture sidecar -> cleaned stdout/events
```

## Preconditions

- The fake TUI emits realistic Codex screen noise and then exits.
- There is no structured Codex sidecar stream; output comes from scrollback fallback.

## Steps

1. Run `agent-run run --agent-runner codex-tty "run ls"` with the noisy fake Codex TUI.
2. Capture stdout and `events.jsonl`.
3. Assert useful result lines are preserved and raw control/UI transcript is absent.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CodexTTYCommand = fakeTUIRawCodexScrollback()
	req.Args = append(req.Args, "run ls")
	return nil
}
```
