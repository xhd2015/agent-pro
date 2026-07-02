# Scenario

**Bug**: real grok run must print discovered session id + updates path on stderr and stream stdout live

```
real grok PTY + live updates.jsonl tail
  -> stderr grok session + updates path
  -> stdout non-empty before timeout
```

## Preconditions

- Real `grok` on PATH (`t.Skip` if absent).
- No fake TUI; no synthetic `GROK_HOME`.

## Steps

1. Run `agent-run run --agent-runner grok-tty "run ls"` with `Mode=stream-probe`.
2. Poll stdout for listing tokens before 60s.
3. Assert stderr contains `grok session` and `grok updates` lines.

```go
import (
	"os/exec"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	if _, err := exec.LookPath("grok"); err != nil {
		t.Skip("grok not found in PATH")
	}
	req.SkipFakeTUI = true
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_GROK_TTY_COMMAND")
	req.Env = withoutEnvKey(req.Env, "GROK_HOME")
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_GROK_TTY_GROK_SESSION_ID")

	req.Args = []string{"run", "--agent-runner", "grok-tty", "run ls"}
	req.Mode = "stream-probe"
	req.StreamProbeSubstring = "agent"
	req.StreamProbeTimeout = 60 * time.Second
	req.ExecTimeout = 120 * time.Second
	return nil
}
```