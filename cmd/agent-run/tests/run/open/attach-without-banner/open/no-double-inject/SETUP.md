# Scenario

**Feature**: new-session `--open` prompt is on argv only — no PTY re-inject

```
agent-run run --agent-runner grok-tty --open "once-only"
  + fake TUI probe: records ARGV/PROMPT_ARG and any stdin line
  + no banner/OpenReady markers; no AGENT_RUN_OPEN_ATTACH_INSTANT
  -> exit 0; no banner error
  -> probe PROMPT_ARG=once-only
  -> probe STDIN_COUNT=0 (must not re-inject same prompt into PTY)
```

## Preconditions

- New session (`ResumeSessionID` empty): initial prompt already on runner argv.
- Re-injecting the same prompt would double-submit; attach-first open must skip
  inject for this case.
- Probe fake still never paints ready markers (attach-first, not inject-ready).

## Steps

1. Write probe TUI under temp dir; set as `AGENT_RUN_GROK_TTY_COMMAND`.
2. Run open with prompt `once-only`.
3. Assert probe shows argv prompt and no stdin inject.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	clearOpenInstantAttach(req)
	req.Prompt = "once-only"
	probePath := filepath.Join(req.TempDir, "tui-probe.txt")
	// Stash probe path for Assert via env (Request has no ProbePath field).
	setEnvKV(req, "DOCTEST_TUI_PROBE_PATH", probePath)
	script := writeFakeTUINoBannerArgvStdinProbe(t, req.TempDir, probePath, 5, 3)
	setGrokTTYCommand(req, script)
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--open", req.Prompt}
	req.ExecTimeout = 55 * time.Second
	return nil
}
```
