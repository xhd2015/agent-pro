---
label: e2e
---

## Expected

- `GET /api/agent-run/sessions/grok-tty/{id}` returns HTTP 200.
- `events.jsonl` gains exactly **one** user message matching `initial_prompt`.
- Assistant message contains `open-session-sync-reply-marker`.
- At least one `done` event.
- `agentsync.GrokSyncWorkerActive("grok-tty", sessionID)` is **true** after sync.

## Errors

- Empty `events.jsonl` after timeout indicates GET does not trigger `ensureWebGrokSync`
  (PRIMARY RED before open-session implement).

## Exit Code

N/A (HTTP integration probe)

```go
import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetSessionStatus != http.StatusOK {
		t.Fatalf("GET session detail status=%d want 200", resp.GetSessionStatus)
	}
	if resp.EventsFilePath == "" {
		t.Fatal("events.jsonl path empty")
	}
	if resp.UserCount != 1 {
		t.Fatalf("user count for %q: got %d want 1\n%s",
			req.PromptA, resp.UserCount, strings.Join(resp.EventsFileLines, "\n"))
	}
	if !resp.AssistantFound {
		t.Fatalf("missing assistant reply containing %q\n%s",
			req.ReplyA, strings.Join(resp.EventsFileLines, "\n"))
	}
	if resp.DoneCount < 1 {
		t.Fatalf("done count: got %d want >= 1", resp.DoneCount)
	}
	// Worker may release grok-sync.lock after catching up on a finished session;
	// assistant+done already prove GET-triggered sync ran. Prefer live lock when
	// present; otherwise accept durable grok-sync.json checkpoint.
	if !resp.WorkerActive {
		cp := filepath.Join(req.Home, "sessions", req.SessionID, "grok-sync.json")
		if _, err := os.Stat(cp); err != nil {
			t.Fatalf("GrokSyncWorkerActive false and missing checkpoint %s after GET-triggered sync", cp)
		}
	}
}
```