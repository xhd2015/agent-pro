# Scenario

**Feature**: Grok id already bound in agent-run meta is rejected; message may
hint `resume --grok-session-id`

```
seed Grok session UUID under GROK_HOME
seed AGENT_RUN_HOME meta: runner=grok-tty, runner_session_id=UUID
  -> agent-run run --resume-from-grok-session UUID
  -> exit 1; already mapped (+ optional resume --grok-session-id hint)
```

## Preconditions

- Mapping match is `meta.runner` in {`grok`, `grok-tty`} and
  `meta.runner_session_id` equal to the requested id.
- Grok session still exists so the failure is mapping, not missing Grok.

## Steps

1. Seed Grok session at process workspace.
2. Seed flat meta `mapped-import-s1` with `runner_session_id` = UUID.
3. Run import flag without overriding runner (omitted → grok-tty).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokCWD = absPath(t, req.WorkDir)
	seedGrokSession(t, req.GrokHome, req.GrokCWD, req.GrokSessionID)
	req.MappedSessID = "mapped-import-s1"
	seedMappedMeta(t, req, "grok-tty", req.MappedSessID, req.GrokSessionID)
	req.Args = runArgs(req, req.GrokSessionID)
	return nil
}
```
