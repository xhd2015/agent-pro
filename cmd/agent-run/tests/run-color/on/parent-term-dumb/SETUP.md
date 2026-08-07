# Scenario

**Feature**: parent `TERM=dumb` is rewritten under `--color`

```
# agent-run process has TERM=dumb
run --color --agent-runner-binary env-logger "prompt"
  -> child TERM=xterm-256color
  -> force color keys still applied
```

## Steps

1. Set parent `TERM=dumb` on agent-run `cmd.Env` (no parent NO_COLOR).
2. Env-logger + `--color` run.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	prepareEnvLoggingRun(t, req)
	req.ParentTERM = "dumb"
	req.ParentNoColor = false
	applyParentEnvFactors(req)

	req.SessionID = "sess-color-term-dumb"
	req.Prompt = "hi"
	req.Args = append(req.Args,
		"--session-id", req.SessionID,
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.Prompt,
	)
	return nil
}
```
