# Scenario

**Feature**: `run --new-terminal` with `--auto-send-or-resume` launches iTerm2
ModeForceNew (run/resume) or ignores the flag on live send

```
# gate
run --new-terminal … (no --auto-send-or-resume)
  -> exit 1; requires --auto-send-or-resume
run -h -> documents --new-terminal

# MODE=run | resume + --new-terminal
auto + --new-terminal + session
  -> collect flags; strip --new-terminal; reconstruct argv
  -> iterm2.OpenConfig(ModeForceNew, shell-quoted "… agent-run run …")
  -> KOOL_ITERM2_SCRIPT_OUT written; launcher exit 0; no in-process provider

# MODE=send + --new-terminal
live + auto + --new-terminal + prompt
  -> ignore --new-terminal; enqueue/send msg_N; no iTerm script
```

## Steps

1. Default `ItermScriptOut` under TempDir for leaves that capture AppleScript.
2. Keep short exec timeout for validation leaves; longer for mode leaves.

## Context

- iTerm2 package env hooks: `KOOL_ITERM2_INSTALLED=1`,
  `KOOL_ITERM2_SCRIPT_OUT`, `KOOL_ITERM2_GOOS=darwin` (applied via
  `applyIterm2TestHooks` when `ItermScriptOut` is set).
- Follow-up command in script is shell-quoted full agent-run invocation;
  assert key tokens (`run`, `--auto-send-or-resume`, session-id, prompt
  separator) rather than exact quoting style.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	// New-terminal leaves: default script capture path (leaves may override).
	ensureItermScriptOutPath(req)
	if req.ExecTimeout <= 0 {
		req.ExecTimeout = 45 * time.Second
	}
	// Avoid ambient TTY command hook on argv-sensitive / launcher leaves.
	req.GrokTTYCommand = ""
	req.Env = withoutEnvKey(req.Env, envGrokTTYCommand)
	return nil
}
```
