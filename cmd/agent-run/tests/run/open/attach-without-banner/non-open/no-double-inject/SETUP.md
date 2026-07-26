# Scenario

**Bug**: non-open new-session `agent-run run --agent-runner grok-tty "<prompt>"`
puts prompt on argv **and** re-injects via PTY → double submit for real Grok
(web path: KeepTerminalAlive, Open=false uses same `RunHeadless` inject block)

```
# agent-run starts grok-tty headless; new-session prompt already on argv
agent-run run --agent-runner grok-tty "once-only"
  + fake TUI: banner → records ARGV/PROMPT_ARG + timed stdin read
  -> exit 0; no "banner not detected"
  -> probe PROMPT_ARG=once-only
  -> probe STDIN_COUNT=0 (must not re-inject same prompt into PTY)
```

## Preconditions

- New session (`ResumeSessionID` empty, `!NoSubmit`): initial prompt already on
  runner argv (Grok auto-submits → turn 1). Re-inject would start turn 2.
- Non-open hard-waits inject-ready banner; probe paints `GROK_TTY_BANNER` so wait
  succeeds, then proves inject policy via timed stdin read.
- Mirrors `attach-without-banner/open/no-double-inject` for the non-open path.

## Steps

1. Write banner+argv+stdin probe TUI under temp dir; set as `AGENT_RUN_GROK_TTY_COMMAND`.
2. Run non-open with prompt `once-only` (no `--open`).
3. Assert probe shows argv prompt and no stdin inject.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	clearOpenInstantAttach(req)
	req.Prompt = "once-only"
	probePath := filepath.Join(req.TempDir, "tui-probe.txt")
	setEnvKV(req, "DOCTEST_TUI_PROBE_PATH", probePath)
	// Short banner delay; timed read proves inject absence without hanging.
	script := writeFakeTUIBannerArgvStdinProbe(t, req.TempDir, probePath, 0.1, 5, 2)
	setGrokTTYCommand(req, script)
	req.Args = []string{"run", "--agent-runner", "grok-tty", req.Prompt}
	req.ExecTimeout = 55 * time.Second
	return nil
}
```
