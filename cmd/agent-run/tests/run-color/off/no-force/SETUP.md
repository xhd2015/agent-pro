# Scenario

**Feature**: without `--color`, agent-run does not force color keys

```
# agent-run process: NO_COLOR=1, FORCE_COLOR=0, CLICOLOR=0, CLICOLOR_FORCE=0
run (no --color) --agent-runner-binary env-logger "prompt"
  -> child: NO_COLOR still set (not cleared by this feature)
  -> FORCE_COLOR is not forced to 1 (stays 0)
  -> CLICOLOR / CLICOLOR_FORCE not forced to 1
```

Note: TTY/PTY layers may set `TERM` independently of the color feature; this
leaf does **not** assert TERM (see `on/parent-term-*` for color-ON TERM policy).

## Steps

1. Hostile parent color env on agent-run `cmd.Env` (not suite Setenv).
2. Env-logger TTY run **without** `--color`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	prepareEnvLoggingRun(t, req)
	req.ParentNoColor = true
	// Good TERM so this leaf isolates force-key policy (not dumb-TERM / PTY noise).
	req.ParentTERM = "screen-256color"
	applyParentEnvFactors(req)
	// Explicit zeros so we can prove agent-run does not rewrite them to 1.
	req.Env = withoutEnvKey(req.Env, "FORCE_COLOR")
	req.Env = withoutEnvKey(req.Env, "CLICOLOR")
	req.Env = withoutEnvKey(req.Env, "CLICOLOR_FORCE")
	req.Env = append(req.Env,
		"FORCE_COLOR=0",
		"CLICOLOR=0",
		"CLICOLOR_FORCE=0",
	)

	req.SessionID = "sess-color-off"
	req.Prompt = "hi"
	req.Args = append(req.Args,
		"--session-id", req.SessionID,
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.Prompt,
	)
	return nil
}
```
