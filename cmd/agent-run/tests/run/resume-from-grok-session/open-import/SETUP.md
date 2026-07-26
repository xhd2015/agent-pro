# Scenario

**Feature**: import + `--open` with instant-attach hook completes without waiting
for hold sleep; meta is pre-bound to the Grok UUID

```
seed Grok UUID; hold fake binary (sleep >> timeout)
AGENT_RUN_OPEN_ATTACH_INSTANT=1
  -> agent-run run --open --session-id FIXED
       --agent-runner-binary HOLD --resume-from-grok-session UUID "hi"
  -> exit 0 before hold ends
  -> meta.runner_session_id = UUID
```

## Preconditions

- Same hold+timeout pattern as detach: proves `Open` is wired (otherwise parent
  blocks on hold and the test times out → RED).
- Instant attach hook matches `cmd/agent-run/tests/run/open/` suite.

## Steps

1. Seed Grok session; install long-hold binary.
2. Enable `AGENT_RUN_OPEN_ATTACH_INSTANT=1`.
3. Run import with `--open`, fixed session id, short prompt; timeout 45s.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "import-open-s1"
	req.OpenFlag = true
	req.OpenInstantAttach = true
	req.FollowupPrompt = "open import hi"
	setupDetachOrOpenImport(t, req, 120, 45*time.Second)
	req.Args = runArgs(req, req.GrokSessionID, req.FollowupPrompt)
	return nil
}
```
