# Scenario

**Feature**: good parent TERM is left alone under `--color`

```
# agent-run process has TERM=screen-256color (not empty/dumb)
run --color --agent-runner-binary env-logger "prompt"
  -> child TERM=screen-256color (not rewritten to xterm-256color)
  -> force color keys still applied
```

## Steps

1. Set parent `TERM=screen-256color` on agent-run `cmd.Env`.
2. Env-logger + `--color` run.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	prepareEnvLoggingRun(t, req)
	req.ParentTERM = "screen-256color"
	applyParentEnvFactors(req)

	req.SessionID = "sess-color-term-good"
	req.Prompt = "hi"
	req.Args = append(req.Args,
		"--session-id", req.SessionID,
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.Prompt,
	)
	return nil
}
```
