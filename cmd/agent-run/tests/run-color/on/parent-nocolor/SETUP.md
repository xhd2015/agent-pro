# Scenario

**Feature**: parent `NO_COLOR=1` is cleared when `--color` is on

```
# agent-run process has NO_COLOR=1; good TERM (no dumb rewrite under test)
run --color --agent-runner-binary env-logger "prompt"
  -> child: no NO_COLOR; FORCE_COLOR=1 CLICOLOR=1 CLICOLOR_FORCE=1
  -> meta does not store color
```

## Steps

1. Set parent `NO_COLOR=1` and `TERM=screen-256color` on agent-run `cmd.Env`.
2. Write env-logging fake runner; stable session id.
3. Run with `--color` (from grouping) + binary + prompt.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	prepareEnvLoggingRun(t, req)
	req.ParentNoColor = true
	req.ParentTERM = "screen-256color"
	applyParentEnvFactors(req)

	req.SessionID = "sess-color-nocolor"
	req.Prompt = "hi"
	req.Args = append(req.Args,
		"--session-id", req.SessionID,
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.Prompt,
	)
	return nil
}
```
