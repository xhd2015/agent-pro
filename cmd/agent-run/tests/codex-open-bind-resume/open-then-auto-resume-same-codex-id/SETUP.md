# Scenario

**Feature**: open binds Codex `runner_session_id`; auto-send-or-resume reuses it

```
agent-run run --open --session-id codex-open-bind-resume-s1 \
  --agent-runner codex-tty --agent-runner-binary llm-mock-run-codex \
  --agent-runner-config-home <codex-home> --dir <ws> -- "OPEN_MARKER"
  + AGENT_RUN_OPEN_ATTACH_INSTANT=1
  -> meta.runner_session_id bound; one codex rollout

end session (send /exit + kill keep-alive)

agent-run run --auto-send-or-resume --open ... -- "FOLLOW_UP"
  -> same runner_session_id / same codex uuid (no second rollout)
```

Local repro converted to doctest. Expect **FAIL** until Fix A (bind on open).

## Steps

1. Root Setup builds binaries + isolates homes.
2. Run executes open → end → auto-send-or-resume.
3. Assert gates on bind + same codex id.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = "codex-open-bind-resume-s1"
	req.OpenPrompt = "OPEN_MARKER"
	req.FollowupPrompt = "FOLLOW_UP"
	return nil
}
```
