## Expected

- `events.jsonl` exists under `AGENT_RUN_HOME/sessions/grok-tty/<id>/`.
- Exactly **one** user `message` with create-session prompt text.
- Assistant message contains `create-session-sync-reply-marker`.
- At least one `done` event.
- `meta.json` `runner_session_id` is non-empty after discovery (optional but preferred).

## Errors

- Empty or missing `events.jsonl` indicates discovery bootstrap not wired in web path
  (PRIMARY RED before `pkgs/agentsync` implement).

## Exit Code

N/A (HTTP integration probe)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
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
			createSessionReply, strings.Join(resp.EventsFileLines, "\n"))
	}
	if resp.DoneCount < 1 {
		t.Fatalf("done count: got %d want >= 1", resp.DoneCount)
	}
}
```