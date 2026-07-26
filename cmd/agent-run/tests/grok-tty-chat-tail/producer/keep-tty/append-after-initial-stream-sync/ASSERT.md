---
label: e2e
---

## Expected

- Initial pending `tool_call` appears in `events.jsonl` **before** assistant marker (proves early sync).
- Delayed completion still produces assistant `message` with `CHAT_TAIL_ASSISTANT_MARKER`.
- Completed tool and `done` after assistant — same end state as P1 despite streamed race.

## Side Effects

- Simulates real `run ls and pwd` timing: first batch synced, tool finishes later.

## Exit Code

0 (run may be killed after events satisfy probe)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.EventsFileLines) == 0 {
		t.Fatalf("events.jsonl empty; stderr:\n%s", resp.Stderr)
	}
	if !resp.HasPendingToolFirst {
		t.Fatalf("expected pending tool_call in events before assistant (initial sync); events=%v", resp.EventsParsed)
	}
	if !resp.HasAssistantMarker {
		t.Fatalf("events.jsonl missing assistant marker %q after streamed race; events=%v stderr:\n%s",
			chatTailAssistantMarker, resp.EventsParsed, resp.Stderr)
	}
	if !resp.HasCompletedTool {
		t.Fatalf("events.jsonl missing completed tool_call after delayed append; events=%v", resp.EventsParsed)
	}
	if !resp.DoneAfterAssistant {
		t.Fatalf("expected done after assistant marker; events=%v", resp.EventsParsed)
	}
}
```