# Scenario

**Bug**: I3 — opening session detail must kick grok sync without POST

```
seed finished session: empty events.jsonl, initial_prompt in meta
  -> grok updates.jsonl on disk with completed turn
  -> agent-run web starts
  -> GET /sessions/grok-tty/{id} only (no POST)
  -> events.jsonl gains user + assistant; GrokSyncWorkerActive true
```

**RED before implement:** GET handlers do not call `ensureWebGrokSync` today.

## Steps

1. Pre-seed `meta.json` (`status=finished`, `initial_prompt`, `runner_session_id`).
2. Pre-seed grok `updates.jsonl` with matching prompt + turn completion.
3. Ensure `events.jsonl` is absent.
4. Start web; issue single GET session detail.
5. Poll `events.jsonl` for user + assistant within timeout.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "web-open-session-starts-sync"
	req.SessionID = "web_open_sync_seed"
	req.PromptA = "open session sync probe prompt"
	req.ReplyA = "open-session-sync-reply-marker"
	req.GrokSessionUUID = "22222222-2222-2222-2222-222222222222"
	req.ProbeTimeout = 15 * time.Second
	return nil
}
```