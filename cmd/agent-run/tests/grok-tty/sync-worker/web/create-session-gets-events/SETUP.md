# Scenario

**Bug**: I2 — web POST create must populate events.jsonl via discovery bootstrap

```
empty GROK_HOME at start (no AGENT_RUN_GROK_TTY_GROK_SESSION_ID)
  -> POST /sessions with prompt
  -> delayed grok updates.jsonl + turn completion
  -> events.jsonl has user + assistant
```

Mirrors production bug `web_a5939cab5f4c7bfe`: session marked finished with
empty chat because sync gated on pre-set `runner_session_id`.

## Steps

1. Start `agent-run web` with empty grok home (no pre-seeded session).
2. Do **not** set `AGENT_RUN_GROK_TTY_GROK_SESSION_ID`.
3. POST create session with prompt; schedule delayed grok session 2s later.
4. Poll `events.jsonl` for user + assistant.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

const (
	createSessionPrompt = "run ls and pwd for sync test"
	createSessionReply  = "create-session-sync-reply-marker"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "web-create-session-events"
	req.PromptA = createSessionPrompt
	req.ReplyA = createSessionReply
	req.CompletionDelayTurn1 = 2 * time.Second
	req.ChromeHoldSeconds = 60
	req.ProbeTimeout = 25 * time.Second
	return nil
}
```