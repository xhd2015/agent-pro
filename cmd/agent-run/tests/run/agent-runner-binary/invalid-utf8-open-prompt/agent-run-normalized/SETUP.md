# Scenario

**Leaf B**: `agent-run` wraps `llm-mock-run-grok` (normalize fix)

```
agent-run run --detach --agent-runner grok-tty \
  --agent-runner-binary <llm-mock-run-grok> \
  --agent-runner-config-home <cfg> \
  --session-id utf8-agent-run-ok \
  <same invalid-utf8 PROMPT>
  -> agent-run normalizes PROMPT
  -> llm-mock-run-grok → real grok does NOT panic
  -> native Go wall-clock 3s (Request.WallClockLimit; no shell harness)
  -> if still alive at 3s: stop waiting; snapshot message; no crash
```

## Steps

1. Keep real `agent-run` binary.
2. Args = detach + session + binary + config-home + incident PROMPT.
3. `WallClockLimit = 3s` so Run uses `execCmdWallClock` (native Go).

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	configHome := filepath.Join(req.TempDir, "agent-runner-config-home")
	if err := os.MkdirAll(configHome, 0755); err != nil {
		return err
	}
	// Parent set: Args = ["run", "--agent-runner", "grok-tty"], AgentRun = real binary.
	req.Args = append(req.Args,
		"--detach",
		"--session-id", invalidUTF8AgentSession,
		"--agent-runner-binary", req.LLMMockRunGrok,
		"--agent-runner-config-home", configHome,
		req.Prompt,
	)
	// Native Go: start agent-run, wait 3s, kill waiter if still running (see execCmdWallClock).
	req.WallClockLimit = invalidUTF8Budget
	return nil
}
```
