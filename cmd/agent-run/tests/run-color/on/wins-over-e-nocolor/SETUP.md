# Scenario

**Feature**: color policy wins over user `-e NO_COLOR=1`

```
run --color -e NO_COLOR=1 --agent-runner-binary env-logger "prompt"
  -> child still has NO_COLOR unset (color applied last)
  -> FORCE_COLOR=1 CLICOLOR=1 CLICOLOR_FORCE=1
```

## Steps

1. Env-logger; pass `--color` (grouping) and `-e NO_COLOR=1` after it.
2. Parent TERM good so this leaf isolates the `-e` precedence, not TERM rewrite.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	prepareEnvLoggingRun(t, req)
	req.ParentTERM = "xterm-256color"
	applyParentEnvFactors(req)

	req.SessionID = "sess-color-wins-e"
	req.Prompt = "hi"
	req.Args = append(req.Args,
		"--session-id", req.SessionID,
		"-e", "NO_COLOR=1",
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.Prompt,
	)
	return nil
}
```
