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
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
	if !resp.WorkerActive {
		t.Fatal("GrokSyncWorkerActive must be true after GET-triggered sync")
	}
}
```