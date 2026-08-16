## Expected

- Exactly **one** user `message` with text `hello?`.
- Exactly **one** user `message` with text `what did I say?`.
- Two `done` events (one per turn).
- Worker count for this session is 1 (not overlapping tails).

## Side Effects

- `events.jsonl` under session dir contains both prompts without duplicates.

## Errors

- Must not emit duplicate user messages differing only by `turn_index` replay.

## Exit Code

N/A (direct package call)

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.EnsureErr != nil {
		t.Fatalf("EnsureGrokSync: %v", resp.EnsureErr)
	}

	countA := countUserMessagesByText(resp.Events, syncPromptA)
	countB := countUserMessagesByText(resp.Events, syncPromptB)
	if countA != 1 {
		t.Fatalf("prompt A %q: want 1 user message, got %d; events=%d", syncPromptA, countA, len(resp.Events))
	}
	if countB != 1 {
		t.Fatalf("prompt B %q: want 1 user message, got %d; events=%d", syncPromptB, countB, len(resp.Events))
	}
	if got := countActionDone(resp.Events); got != 2 {
		t.Fatalf("want 2 ActionDone (one per turn), got %d", got)
	}
	if !resp.WorkerActive {
		t.Fatal("expected this session's grok sync worker to be active")
	}
	for _, reply := range []string{syncReplyA, syncReplyB} {
		found := false
		for _, ev := range resp.Events {
			if ev.Type == types.ActionMessage && ev.Role == "assistant" && ev.Text == reply {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing assistant reply %q; events=%d", reply, len(resp.Events))
		}
	}
}
```
